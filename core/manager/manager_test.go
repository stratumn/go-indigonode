// Copyright © 2017  Stratumn SAS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager/mockmanager"
)

func mockService(ctrl *gomock.Controller, id string) Service {
	serv := mockmanager.NewMockService(ctrl)
	serv.EXPECT().ID().Return(id).AnyTimes()
	serv.EXPECT().Name().Return(id).AnyTimes()
	serv.EXPECT().Desc().Return(id).AnyTimes()

	return serv
}

func mockNeedy(ctrl *gomock.Controller, needs map[string]struct{}) Needy {
	needy := mockmanager.NewMockNeedy(ctrl)
	needy.EXPECT().Needs().Return(needs).AnyTimes()

	return needy
}

func mockFriendly(ctrl *gomock.Controller, likes map[string]struct{}) *mockmanager.MockFriendly {
	friendly := mockmanager.NewMockFriendly(ctrl)
	friendly.EXPECT().Likes().Return(likes).AnyTimes()

	return friendly
}

func mockPluggable(ctrl *gomock.Controller, needs map[string]struct{}) *mockmanager.MockPluggable {
	pluggable := mockmanager.NewMockPluggable(ctrl)
	pluggable.EXPECT().Needs().Return(needs).AnyTimes()

	return pluggable
}

func mockExposer(ctrl *gomock.Controller, exposed interface{}) Exposer {
	exposer := mockmanager.NewMockExposer(ctrl)
	exposer.EXPECT().Expose().Return(exposed).AnyTimes()

	return exposer
}

type mockRunnerFn func(context.Context, func(), func()) error

type testRunner mockRunnerFn

func (r testRunner) Run(ctx context.Context, running, stopping func()) error {
	return r(ctx, running, stopping)
}

func mockRunner(run mockRunnerFn) Runner {
	return testRunner(run)
}

func mockNeedyService(ctrl *gomock.Controller, id string, needs map[string]struct{}) Service {
	return struct {
		Service
		Needy
	}{
		mockService(ctrl, id),
		mockNeedy(ctrl, needs),
	}
}

func mockExposerService(ctrl *gomock.Controller, id string, expose interface{}) Service {
	return struct {
		Service
		Exposer
	}{
		mockService(ctrl, id),
		mockExposer(ctrl, expose),
	}
}

func mockRunnerService(ctrl *gomock.Controller, id string, run mockRunnerFn) Service {
	return struct {
		Service
		Runner
	}{
		mockService(ctrl, id),
		mockRunner(run),
	}
}

var errMockCrash = errors.New("crashed")

func mockCrashStart(ctrl *gomock.Controller, id string) Service {
	return mockRunnerService(ctrl, id, func(context.Context, func(), func()) error {
		return nil
	})
}

func mockCrashStartErr(ctrl *gomock.Controller, id string) Service {
	return mockRunnerService(ctrl, id, func(context.Context, func(), func()) error {
		return errMockCrash
	})
}

func mockCrashStartStatus(ctrl *gomock.Controller, id string) Service {
	return mockRunnerService(ctrl, id, func(ctx context.Context, running, stopping func()) error {
		stopping()
		return nil
	})
}

func mockCrashStop(ctrl *gomock.Controller, id string) Service {
	return mockRunnerService(ctrl, id, func(ctx context.Context, running, stopping func()) error {
		running()
		<-ctx.Done()
		stopping()
		return errMockCrash
	})
}

func mockPluggableErr(ctrl *gomock.Controller, id string, needs map[string]struct{}) Service {
	pluggable := mockmanager.NewMockPluggable(ctrl)
	pluggable.EXPECT().Needs().Return(needs).AnyTimes()
	pluggable.EXPECT().Plug(gomock.Any()).Return(errMockCrash).AnyTimes()

	return struct {
		Service
		Pluggable
	}{
		mockService(ctrl, id),
		pluggable,
	}
}

func createTestMgr(ctx context.Context, t testing.TB, ctrl *gomock.Controller) *Manager {
	mgr := New()

	go func() {
		err := mgr.Work(ctx)
		if err != nil && errors.Cause(err) != context.Canceled {
			t.Errorf(`manager: Work(ctx): error: %s`, err)
		}
	}()

	mgr.RegisterService()
	mgr.Register(mockService(ctrl, "net"))
	mgr.Register(mockService(ctrl, "fs"))
	mgr.Register(mockNeedyService(ctrl, "crypto", map[string]struct{}{
		"fs": struct{}{},
	}))
	mgr.Register(mockNeedyService(ctrl, "apps", map[string]struct{}{
		"net":    struct{}{},
		"crypto": struct{}{},
		"fs":     struct{}{},
	}))
	mgr.Register(mockExposerService(ctrl, "api", "api"))
	mgr.Register(mockCrashStart(ctrl, "crash-start"))
	mgr.Register(mockCrashStartErr(ctrl, "crash-start-err"))
	mgr.Register(mockCrashStartStatus(ctrl, "crash-start-status"))
	mgr.Register(mockCrashStop(ctrl, "crash-stop"))
	mgr.Register(mockPluggableErr(ctrl, "crash-plug", map[string]struct{}{
		"fs": struct{}{},
	}))

	return mgr
}

type mgrTest struct {
	name   string
	do     func(*Manager) error
	err    error
	status map[string]StatusCode
}

var mgrTests = []mgrTest{{
	"Start_deps",
	func(mgr *Manager) error {
		return mgr.Start("apps")
	},
	nil,
	map[string]StatusCode{
		"manager":            Stopped,
		"net":                Running,
		"fs":                 Running,
		"crypto":             Running,
		"apps":               Running,
		"api":                Stopped,
		"crash-start":        Stopped,
		"crash-start-err":    Stopped,
		"crash-start-status": Stopped,
		"crash-stop":         Stopped,
		"crash-plug":         Stopped,
	},
}, {
	"Start_inexistent",
	func(mgr *Manager) error {
		return mgr.Start("http")
	},
	ErrNotFound,
	nil,
}, {
	"Start_manager",
	func(mgr *Manager) error {
		return mgr.Start("manager")
	},
	nil,
	map[string]StatusCode{
		"manager": Running,
	},
}, {
	"Start_crash",
	func(mgr *Manager) error {
		return mgr.Start("crash-start")
	},
	ErrInvalidStatus,
	map[string]StatusCode{
		"crash-start": Errored,
	},
}, {
	"Start_crash_err",
	func(mgr *Manager) error {
		return mgr.Start("crash-start-err")
	},
	errMockCrash,
	map[string]StatusCode{
		"crash-start-err": Errored,
	},
}, {
	"Start_crash_status",
	func(mgr *Manager) error {
		return mgr.Start("crash-start-status")
	},
	ErrInvalidStatus,
	map[string]StatusCode{
		"crash-start-status": Errored,
	},
}, {
	"Start_crash_plug",
	func(mgr *Manager) error {
		return mgr.Start("crash-plug")
	},
	errMockCrash,
	map[string]StatusCode{
		"crash-plug": Errored,
	},
}, {
	"Stop",
	func(mgr *Manager) error {
		if err := mgr.Start("apps"); err != nil {
			return err
		}
		return mgr.Stop("apps")
	},
	nil,
	map[string]StatusCode{
		"manager":            Stopped,
		"net":                Running,
		"fs":                 Running,
		"crypto":             Running,
		"apps":               Stopped,
		"api":                Stopped,
		"crash-start":        Stopped,
		"crash-start-err":    Stopped,
		"crash-start-status": Stopped,
		"crash-stop":         Stopped,
		"crash-plug":         Stopped,
	},
}, {
	"Stop_needed",
	func(mgr *Manager) error {
		if err := mgr.Start("apps"); err != nil {
			return err
		}
		return mgr.Stop("crypto")
	},
	ErrNeeded,
	map[string]StatusCode{
		"net":    Running,
		"fs":     Running,
		"crypto": Running,
	},
}, {
	"Stop_inexistent",
	func(mgr *Manager) error {
		return mgr.Stop("http")
	},
	ErrNotFound,
	nil,
}, {
	"Stop_crash",
	func(mgr *Manager) error {
		if err := mgr.Start("crash-stop"); err != nil {
			return err
		}
		return mgr.Stop("crash-stop")
	},
	errMockCrash,
	map[string]StatusCode{
		"crash-stop": Errored,
	},
}, {
	"StopAll",
	func(mgr *Manager) error {
		if err := mgr.Start("apps"); err != nil {
			return err
		}
		if err := mgr.Start("fs"); err != nil {
			return err
		}
		mgr.StopAll()
		return nil
	},
	nil,
	map[string]StatusCode{
		"manager": Stopped,
		"net":     Stopped,
		"fs":      Stopped,
		"crypto":  Stopped,
		"apps":    Stopped,
		"api":     Stopped,
	},
}, {
	"Prune",
	func(mgr *Manager) error {
		if err := mgr.Start("apps"); err != nil {
			return err
		}
		if err := mgr.Start("fs"); err != nil {
			return err
		}
		if err := mgr.Start("api"); err != nil {
			return err
		}
		if err := mgr.Stop("apps"); err != nil {
			return err
		}
		mgr.Prune()
		return nil
	},
	nil,
	map[string]StatusCode{
		"manager": Stopped,
		"net":     Stopped,
		"fs":      Running,
		"crypto":  Stopped,
		"apps":    Stopped,
		"api":     Running,
	},
}, {
	"Group",
	func(mgr *Manager) error {
		mgr.Register(&ServiceGroup{
			GroupID: "group",
			Services: map[string]struct{}{
				"net": struct{}{},
				"api": struct{}{},
			},
		})
		return mgr.Start("group")
	},
	nil,
	map[string]StatusCode{
		"net":    Running,
		"fs":     Stopped,
		"crypto": Stopped,
		"apps":   Stopped,
		"api":    Running,
	},
}}

func testMgr(t *testing.T, ctrl *gomock.Controller, test mgrTest) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	got, want := errors.Cause(test.do(mgr)), errors.Cause(test.err)
	if got != want {
		t.Errorf("do(): err = %v want %v", got, want)
	}

	for servID, want := range test.status {
		got, err := mgr.Status(servID)
		if err != nil {
			t.Errorf("mgr.Status(%v): error: %s", servID, err)
		}

		if got != want {
			t.Errorf("mgr.Status(%v) = %v want %v", servID, got, want)
		}
	}
}

func TestManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range mgrTests {
		t.Run(tt.name, func(t *testing.T) {
			testMgr(t, ctrl, tt)
		})
	}
}

func TestPluggable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	pluggable := mockPluggable(ctrl, map[string]struct{}{
		"api": struct{}{},
	})

	mgr.Register(struct {
		Service
		Pluggable
	}{
		mockNeedyService(ctrl, "pluggable", map[string]struct{}{
			"api": struct{}{},
		}),
		pluggable,
	})

	pluggable.EXPECT().Plug(map[string]interface{}{"api": "api"}).Times(1)

	if err := mgr.Start("pluggable"); err != nil {
		t.Errorf(`mgr.Start("pluggable"): error: %s`, err)
	}
}

func TestPluggable_nil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	pluggable := mockPluggable(ctrl, map[string]struct{}{
		"fs": struct{}{},
	})

	mgr.Register(struct {
		Service
		Pluggable
	}{
		mockNeedyService(ctrl, "pluggable", map[string]struct{}{
			"fs": struct{}{},
		}),
		pluggable,
	})

	pluggable.EXPECT().Plug(map[string]interface{}{"fs": nil}).Times(1)

	if err := mgr.Start("pluggable"); err != nil {
		t.Errorf(`mgr.Start("pluggable"): error: %s`, err)
	}
}

func TestFriendly_likes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)

	friendly := mockFriendly(ctrl, map[string]struct{}{
		"api":    struct{}{},
		"crypto": struct{}{},
	})

	mgr.Register(struct {
		Service
		Friendly
	}{
		mockService(ctrl, "friendly"),
		friendly,
	})

	friendly.EXPECT().Befriend("crypto", nil).Times(1)

	if err := mgr.Start("crypto"); err != nil {
		t.Errorf(`mgr.Start("crypto"): error: %s`, err)
	}

	friendly.EXPECT().Befriend("api", "api").Times(1)

	if err := mgr.Start("api"); err != nil {
		t.Errorf(`mgr.Start("api"): error: %s`, err)
	}

	friendly.EXPECT().Befriend("api", nil).Times(1)
	friendly.EXPECT().Befriend("crypto", nil).Times(1)

	stoppedCh := make(chan struct{})
	go func() {
		mgr.StopAll()
		close(stoppedCh)
	}()

	select {
	case <-time.After(time.Second):
		t.Errorf("stoppedCh didn't close")
	case <-stoppedCh:
	}
}

func TestFriendly_nil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)

	friendly := mockFriendly(ctrl, nil)

	mgr.Register(struct {
		Service
		Friendly
	}{
		mockService(ctrl, "friendly"),
		friendly,
	})

	stoppedCh := make(chan struct{})
	go func() {
		mgr.StopAll()
		close(stoppedCh)
	}()

	select {
	case <-time.After(time.Second):
		t.Errorf("stoppedCh didn't close")
	case <-stoppedCh:
	}
}

func TestFriendly_liked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)

	friendly := mockFriendly(ctrl, map[string]struct{}{
		"http": struct{}{},
	})

	mgr.Register(struct {
		Service
		Friendly
	}{
		mockService(ctrl, "friendly"),
		friendly,
	})

	mgr.Register(mockExposerService(ctrl, "http", "http"))

	friendly.EXPECT().Befriend("http", "http").Times(1)

	if err := mgr.Start("http"); err != nil {
		t.Errorf(`mgr.Start("http"): error: %s`, err)
	}

	friendly.EXPECT().Befriend("http", nil).Times(1)

	stoppedCh := make(chan struct{})
	go func() {
		mgr.StopAll()
		close(stoppedCh)
	}()

	select {
	case <-time.After(time.Second):
		t.Errorf("stoppedCh didn't close")
	case <-stoppedCh:
	}
}

func TestManager_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	got := fmt.Sprint(mgr.List())
	want := "[api apps crash-plug crash-start crash-start-err crash-start-status crash-stop crypto fs manager net]"

	if got != want {
		t.Errorf("mgr.List() = %v want %v", got, want)
	}
}

func TestManager_Find(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	serv, err := mgr.Find("api")
	if err != nil {
		t.Errorf(`mgr.Find("api"): error: %s`, err)
	}

	if got, want := serv.ID(), "api"; got != want {
		t.Errorf("serv.ID() = %v want %v", got, want)
	}

	_, err = mgr.Find("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Proto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	proto, err := mgr.Proto("api")
	if err != nil {
		t.Errorf(`mgr.Proto("api"): error: %s`, err)
	}

	if got, want := proto.Id, "api"; got != want {
		t.Errorf("proto.Id = %v want %v", got, want)
	}

	_, err = mgr.Proto("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Status(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	status, err := mgr.Status("api")
	if err != nil {
		t.Errorf(`mgr.Status("api"): error: %s`, err)
	}

	if got, want := status, Stopped; got != want {
		t.Errorf("status = %v want %v", got, want)
	}

	_, err = mgr.Status("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Stoppable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	if err := mgr.Start("apps"); err != nil {
		t.Errorf(`mgr.Start("apps"): error: %s`, err)
	}

	stoppable, err := mgr.Stoppable("apps")
	if err != nil {
		t.Errorf(`mgr.Stoppable("apps"): error: %s`, err)
	}

	if got, want := stoppable, true; got != want {
		t.Errorf("stoppable = %v want %v", got, want)
	}

	stoppable, err = mgr.Stoppable("crypto")
	if err != nil {
		t.Errorf(`mgr.Stoppable("crypto"): error: %s`, err)
	}

	if got, want := stoppable, false; got != want {
		t.Errorf("stoppable = %v want %v", got, want)
	}

	stoppable, err = mgr.Stoppable("api")
	if err != nil {
		t.Errorf(`mgr.Stoppable("api"): error: %s`, err)
	}

	if got, want := stoppable, true; got != want {
		t.Errorf("stoppable = %v want %v", got, want)
	}

	_, err = mgr.Stoppable("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Prunable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	if err := mgr.Start("apps"); err != nil {
		t.Errorf(`mgr.Start("apps"): error: %s`, err)
	}

	prunable, err := mgr.Prunable("apps")
	if err != nil {
		t.Errorf(`mgr.Prunable("apps"): error: %s`, err)
	}

	if got, want := prunable, false; got != want {
		t.Errorf("prunable = %v want %v", got, want)
	}

	prunable, err = mgr.Prunable("crypto")
	if err != nil {
		t.Errorf(`mgr.Prunable("crypto"): error: %s`, err)
	}

	if got, want := prunable, true; got != want {
		t.Errorf("prunable = %v want %v", got, want)
	}

	prunable, err = mgr.Prunable("api")
	if err != nil {
		t.Errorf(`mgr.Prunable("api"): error: %s`, err)
	}

	if got, want := prunable, false; got != want {
		t.Errorf("prunable = %v want %v", got, want)
	}

	_, err = mgr.Prunable("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Expose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	if err := mgr.Start("api"); err != nil {
		t.Errorf(`mgr.Start("api"): error: %s`, err)
	}

	exposed, err := mgr.Expose("apps")
	if err != nil {
		t.Errorf(`mgr.Expose("apps"): error: %s`, err)
	}

	if got, want := exposed, interface{}(nil); got != want {
		t.Errorf("exposed = %v want %v", got, want)
	}

	exposed, err = mgr.Expose("api")
	if err != nil {
		t.Errorf(`mgr.Expose("api"): error: %s`, err)
	}

	if got, want := exposed, "api"; got != want {
		t.Errorf("exposed = %v want %v", got, want)
	}

	_, err = mgr.Expose("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func assertChan(t *testing.T, ch <-chan struct{}) {
	select {
	case <-time.After(time.Second):
		t.Errorf("channel didn't receive")
	case <-ch:
	}
}

func TestManager_Starting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	ch, err := mgr.Starting("api")
	if err != nil {
		t.Errorf(`mgr.Starting("api"): error: %s`, err)
	}

	if err := mgr.Start("api"); err != nil {
		t.Errorf(`mgr.Start("api"): error: %s`, err)
	}

	assertChan(t, ch)

	_, err = mgr.Starting("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Running(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	ch, err := mgr.Running("api")
	if err != nil {
		t.Errorf(`mgr.Running("api"): error: %s`, err)
	}

	if err := mgr.Start("api"); err != nil {
		t.Errorf(`mgr.Start("api"): error: %s`, err)
	}

	assertChan(t, ch)

	_, err = mgr.Running("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Stopping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	ch, err := mgr.Stopping("api")
	if err != nil {
		t.Errorf(`mgr.Stopping("api"): error: %s`, err)
	}

	if err := mgr.Start("api"); err != nil {
		t.Errorf(`mgr.Start("api"): error: %s`, err)
	}

	if err := mgr.Stop("api"); err != nil {
		t.Errorf(`mgr.Stop("api"): error: %s`, err)
	}

	assertChan(t, ch)

	_, err = mgr.Stopping("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func TestManager_Stopped(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	ch, err := mgr.Stopped("api")
	if err != nil {
		t.Errorf(`mgr.Stopped("api"): error: %s`, err)
	}

	if err := mgr.Start("api"); err != nil {
		t.Errorf(`mgr.Start("api"): error: %s`, err)
	}

	if err := mgr.Stop("api"); err != nil {
		t.Errorf(`mgr.Stop("api"): error: %s`, err)
	}

	assertChan(t, ch)

	_, err = mgr.Stopped("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

func assertChanErr(t *testing.T, ch <-chan error) {
	select {
	case <-time.After(time.Second):
		t.Errorf("channel didn't receive")
	case err := <-ch:
		if err == nil {
			t.Errorf("channel received nil")
		}
	}
}

func TestManager_Errored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	ch, err := mgr.Errored("crash-stop")
	if err != nil {
		t.Errorf(`mgr.Errored("crash-stop"): error: %s`, err)
	}

	if err := mgr.Start("crash-stop"); err != nil {
		t.Errorf(`mgr.Start("crash-stop"): error: %s`, err)
	}

	if got, want := errors.Cause(mgr.Stop("crash-stop")), errMockCrash; got != want {
		t.Errorf(`mgr.Stop("crash-stop"): error = %v want %v`, got, want)
	}

	assertChanErr(t, ch)

	_, err = mgr.Errored("http")
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf(`err = %v want %v`, got, want)
	}
}

type testService struct {
	id    string
	needs map[string]struct{}
}

func (s testService) ID() string {
	return s.id
}

func (s testService) Name() string {
	return s.id
}

func (s testService) Desc() string {
	return s.id
}

func (s testService) Needs() map[string]struct{} {
	return s.needs
}

type mgrDepsTest struct {
	name     string
	services []testService
	sid      string
	err      error
	want     []string
}

var mgrDepsTests = []mgrDepsTest{{
	"valid dependencies",
	[]testService{
		{id: "salt"},
		{id: "pepper"},
		{id: "tomatoes"},
		{id: "strawberries"},
		{id: "cheese"},
		{id: "flour"},
		{id: "icecream"},
		{id: "milk"},
		{id: "yeast"},
		{id: "water"},
		{id: "sauce", needs: map[string]struct{}{
			"tomatoes": struct{}{},
			"salt":     struct{}{},
			"pepper":   struct{}{},
		}},
		{id: "dough", needs: map[string]struct{}{
			"flour": struct{}{},
			"yeast": struct{}{},
			"salt":  struct{}{}}},
		{id: "pizza", needs: map[string]struct{}{
			"dough":  struct{}{},
			"sauce":  struct{}{},
			"cheese": struct{}{}}},
		{id: "milkshake", needs: map[string]struct{}{
			"icecream":     struct{}{},
			"milk":         struct{}{},
			"strawberries": struct{}{}}},
	},
	"pizza",
	nil,
	[]string{
		"cheese",
		"flour",
		"salt",
		"yeast",
		"dough",
		"pepper",
		"tomatoes",
		"sauce",
		"pizza"},
}, {
	"unknown service",
	[]testService{
		{id: "salt"},
		{id: "pepper"},
		{id: "tomatoes"},
		{id: "strawberries"},
		{id: "cheese"},
		{id: "flour"},
		{id: "icecream"},
		{id: "milk"},
		{id: "yeast"},
		{id: "water"},
		{id: "sauce", needs: map[string]struct{}{
			"tomatoes": struct{}{},
			"salt":     struct{}{},
			"pepper":   struct{}{},
			"garlic":   struct{}{}}},
		{id: "dough", needs: map[string]struct{}{
			"flour": struct{}{},
			"yeast": struct{}{},
			"salt":  struct{}{}}},
		{id: "pizza", needs: map[string]struct{}{
			"dough":  struct{}{},
			"sauce":  struct{}{},
			"cheese": struct{}{}}},
		{id: "milkshake", needs: map[string]struct{}{
			"icecream":     struct{}{},
			"milk":         struct{}{},
			"strawberries": struct{}{}}},
	},
	"pizza",
	ErrNotFound,
	nil,
}, {
	"cyclic dependencies",
	[]testService{
		{id: "salt"},
		{id: "pepper"},
		{id: "tomatoes"},
		{id: "strawberries"},
		{id: "cheese"},
		{id: "flour"},
		{id: "icecream"},
		{id: "milk"},
		{id: "yeast"},
		{id: "water"},
		{id: "sauce", needs: map[string]struct{}{
			"tomatoes": struct{}{},
			"salt":     struct{}{},
			"pepper":   struct{}{}}},
		{id: "dough", needs: map[string]struct{}{
			"flour": struct{}{},
			"yeast": struct{}{},
			"pizza": struct{}{},
			"salt":  struct{}{}}},
		{id: "pizza", needs: map[string]struct{}{
			"dough":  struct{}{},
			"sauce":  struct{}{},
			"cheese": struct{}{}}},
		{id: "milkshake", needs: map[string]struct{}{
			"icecream":     struct{}{},
			"milk":         struct{}{},
			"strawberries": struct{}{}}},
	},
	"pizza",
	ErrCyclic,
	nil,
}, {
	"self dependency",
	[]testService{
		{id: "salt"},
		{id: "pepper"},
		{id: "tomatoes"},
		{id: "strawberries"},
		{id: "cheese"},
		{id: "flour"},
		{id: "icecream"},
		{id: "milk"},
		{id: "yeast"},
		{id: "water"},
		{id: "sauce", needs: map[string]struct{}{
			"tomatoes": struct{}{},
			"salt":     struct{}{},
			"pepper":   struct{}{}}},
		{id: "dough", needs: map[string]struct{}{
			"flour": struct{}{},
			"yeast": struct{}{},
			"dough": struct{}{},
			"salt":  struct{}{}}},
		{id: "pizza", needs: map[string]struct{}{
			"dough":  struct{}{},
			"sauce":  struct{}{},
			"cheese": struct{}{}}},
		{id: "milkshake", needs: map[string]struct{}{
			"icecream":     struct{}{},
			"milk":         struct{}{},
			"strawberries": struct{}{}}},
	},
	"pizza",
	ErrCyclic,
	nil,
}}

func testMgrDeps(t *testing.T, test mgrDepsTest) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	for _, serv := range test.services {
		mgr.Register(serv)
	}

	got, err := mgr.Deps(test.sid)
	if err != nil {
		if got, want := errors.Cause(err), test.err; got != want {
			t.Errorf(
				"mgr.Deps(%v): error = %v want %v",
				test.sid, got, want,
			)
		}

		return
	}

	gots, wants := fmt.Sprintf("%v", got), fmt.Sprintf("%v", test.want)

	if gots != wants {
		t.Errorf("mgr.Deps(%v) = %v want %v", test.sid, gots, wants)
	}
}

func TestManager_Deps(t *testing.T) {
	for _, tt := range mgrDepsTests {
		t.Run(tt.name, func(t *testing.T) {
			testMgrDeps(t, tt)
		})
	}
}

func TestManager_FGraph(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr := createTestMgr(ctx, t, ctrl)
	defer mgr.StopAll()

	w := bytes.NewBuffer(nil)
	if err := mgr.Fgraph(w, "apps", ""); err != nil {
		t.Errorf(`mgr.Fgraph(w, "apps", ""): error: %s`, err)
	}

	got := w.String()
	want := `apps┬crypto─fs
    │
    ├fs
    │
    └net
`

	if got != want {

		t.Errorf(`mgr.Fgraph(w, "apps", "") =
%s want
%s`, got, want)
	}
}
