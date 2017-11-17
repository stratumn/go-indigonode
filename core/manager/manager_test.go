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

	"github.com/pkg/errors"
)

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

type testExposer struct {
	testService
}

func (s testExposer) Expose() interface{} {
	return s.id
}

type testPluggable struct {
	testService
	plugCh chan map[string]interface{}
}

func (s testPluggable) Plug(exposed map[string]interface{}) error {
	s.plugCh <- exposed
	return nil
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

func createTestMgr(ctx context.Context, t *testing.T) *Manager {
	mgr := New()

	go func() {
		err := mgr.Work(ctx)
		if err != nil && errors.Cause(err) != context.Canceled {
			t.Errorf(`manager: Work(ctx): error: %s`, err)
		}
	}()

	mgr.Register(testService{id: "net"})
	mgr.Register(testService{id: "fs"})
	mgr.Register(testService{
		id: "crypto",
		needs: map[string]struct{}{
			"fs": struct{}{},
		},
	})
	mgr.Register(testService{
		id: "apps",
		needs: map[string]struct{}{
			"net":    struct{}{},
			"crypto": struct{}{},
			"fs":     struct{}{},
		},
	})
	mgr.Register(testExposer{testService{id: "api"}})

	return mgr
}

var mgrTT = []struct {
	name   string
	do     func(*Manager) error
	err    error
	status map[string]StatusCode
}{{
	"start with deps",
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
}}

func TestManager(t *testing.T) {
	for _, test := range mgrTT {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mgr := createTestMgr(ctx, t)

		got, want := errors.Cause(test.do(mgr)), errors.Cause(test.err)
		if got != want {
			t.Errorf("%s: do(): err = %q want %q", test.name, got, want)
		}

		for servID, want := range test.status {
			got, err := mgr.Status(servID)
			if err != nil {
				t.Errorf("%s: mgr.Status(%q): error: %s", test.name, servID, err)
			}

			if got != want {
				t.Errorf("%s: mgr.Status(%q) = %s want %s", test.name, servID, got, want)
			}
		}

		mgr.StopAll()
		cancel()
	}
}

func TestPluggable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr := createTestMgr(ctx, t)
	defer mgr.StopAll()

	plugCh := make(chan map[string]interface{}, 1)

	mgr.Register(testPluggable{
		testService{
			id:    "pluggable",
			needs: map[string]struct{}{"api": struct{}{}},
		},
		plugCh,
	})

	if err := mgr.Start("pluggable"); err != nil {
		t.Errorf(`mgr.Start("pluggable"): error: %s`, err)
	}

	select {
	case <-time.After(time.Second):
		t.Errorf("plugCh didn't receive anything")
	case exposed := <-plugCh:
		exp, ok := exposed["api"]
		if !ok {
			t.Errorf(`exposed["api"] = %v want %q`, nil, "api")
		}
		got, ok := exp.(string)
		if !ok {
			t.Errorf(`exposed["api"] = %q want %q`, exp, "api")
		}
		if got != "api" {
			t.Errorf(`exposed["api"] = %q want %q`, got, "api")
		}
	}
}

func TestManager_FGraph(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr := createTestMgr(ctx, t)
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

func BenchmarkManager(b *testing.B) {
	mgr := New()
	defer mgr.StopAll()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := mgr.Work(ctx)
		if err != nil && errors.Cause(err) != context.Canceled {
			b.Errorf(`manager: mgr.Work(ctx): error: %s`, err)
		}
	}()

	mgr.Register(testService{id: "net"})
	mgr.Register(testService{id: "fs"})
	mgr.Register(testService{id: "crypto"})
	mgr.Register(testService{
		id: "apps",
		needs: map[string]struct{}{
			"net":    struct{}{},
			"fs":     struct{}{},
			"crypto": struct{}{},
		},
	})
	mgr.Register(testService{id: "api"})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := mgr.Start("apps"); err != nil {
			b.Fatalf(`mgr.Start("apps"): error: %s`, err)
		}
		if err := mgr.Stop("apps"); err != nil {
			b.Fatalf(`mgr.Stop("apps"): error: %s`, err)
		}
	}
}
