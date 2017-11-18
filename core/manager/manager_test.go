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
	"github.com/stratumn/alice/core/manager/mock_manager"
)

func mockService(ctrl *gomock.Controller, id string) Service {
	serv := mock_manager.NewMockService(ctrl)
	serv.EXPECT().ID().Return(id).AnyTimes()
	serv.EXPECT().Name().Return(id).AnyTimes()
	serv.EXPECT().Desc().Return(id).AnyTimes()

	return serv
}

func mockNeedy(ctrl *gomock.Controller, needs map[string]struct{}) Needy {
	needy := mock_manager.NewMockNeedy(ctrl)
	needy.EXPECT().Needs().Return(needs).AnyTimes()

	return needy
}

func mockFriendly(ctrl *gomock.Controller, likes map[string]struct{}) *mock_manager.MockFriendly {
	friendly := mock_manager.NewMockFriendly(ctrl)
	friendly.EXPECT().Likes().Return(likes).AnyTimes()

	return friendly
}

func mockPluggable(ctrl *gomock.Controller, needs map[string]struct{}) *mock_manager.MockPluggable {
	pluggable := mock_manager.NewMockPluggable(ctrl)
	pluggable.EXPECT().Needs().Return(needs).AnyTimes()

	return pluggable
}

func mockExposer(ctrl *gomock.Controller, exposed interface{}) Exposer {
	exposer := mock_manager.NewMockExposer(ctrl)
	exposer.EXPECT().Expose().Return(exposed).AnyTimes()

	return exposer
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

func createTestMgr(ctx context.Context, t testing.TB, ctrl *gomock.Controller) *Manager {
	mgr := New()

	go func() {
		err := mgr.Work(ctx)
		if err != nil && errors.Cause(err) != context.Canceled {
			t.Errorf(`manager: Work(ctx): error: %s`, err)
		}
	}()

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

	return mgr
}

var mgrTT = []struct {
	name   string
	do     func(*Manager) error
	err    error
	status map[string]StatusCode
}{{
	"Start_deps",
	func(mgr *Manager) error {
		return mgr.Start("apps")
	},
	nil,
	map[string]StatusCode{
		"net":    Running,
		"fs":     Running,
		"crypto": Running,
		"apps":   Running,
		"api":    Stopped,
	},
}, {
	"Start_inexistent",
	func(mgr *Manager) error {
		return mgr.Start("http")
	},
	ErrNotFound,
	nil,
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
		"net":    Running,
		"fs":     Running,
		"crypto": Running,
		"apps":   Stopped,
		"api":    Stopped,
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
		"net":    Stopped,
		"fs":     Stopped,
		"crypto": Stopped,
		"apps":   Stopped,
		"api":    Stopped,
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
		"net":    Stopped,
		"fs":     Running,
		"crypto": Stopped,
		"apps":   Stopped,
		"api":    Running,
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

func TestManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, test := range mgrTT {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		mgr := createTestMgr(ctx, t, ctrl)

		got, want := errors.Cause(test.do(mgr)), errors.Cause(test.err)
		if got != want {
			t.Errorf("%s: do(): err = %q want %q", test.name, got, want)
		}

		for servID, want := range test.status {
			got, err := mgr.Status(servID)
			if err != nil {
				t.Errorf(
					"%s: mgr.Status(%q): error: %s",
					test.name, servID, err,
				)
			}

			if got != want {
				t.Errorf(
					"%s: mgr.Status(%q) = %s want %s",
					test.name, servID, got, want,
				)
			}
		}

		mgr.StopAll()
		cancel()
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

func TestFriendly(t *testing.T) {
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

	if err := mgr.Start("friendly"); err != nil {
		t.Errorf(`mgr.Start("friendly"): error: %s`, err)
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

var depsTT = []struct {
	name     string
	services []testService
	sid      string
	err      error
	want     []string
}{{
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
},
	{
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
	},
	{
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
	},
	{
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

func TestManager_Deps(t *testing.T) {
	for _, test := range depsTT {
		mgr := New()

		for _, serv := range test.services {
			mgr.Register(serv)
		}

		got, err := mgr.Deps(test.sid)
		if err != nil {
			if got, want := errors.Cause(err), errors.Cause(test.err); got != want {
				t.Errorf(
					"%s: mgr.Deps(%q): error = %q want %q",
					test.name, test.sid, got, want,
				)
			}
			continue
		}

		gots, wants := fmt.Sprintf("%q", got), fmt.Sprintf("%q", test.want)

		if gots != wants {
			t.Errorf("%s: mgr.Deps(%q) = %s want %s", test.name, test.sid, gots, wants)
		}
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
