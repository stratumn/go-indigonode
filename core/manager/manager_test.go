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
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

type testService struct {
	id    string
	needs map[string]struct{}
}

func (s testService) ID() string {
	return s.id
}

func (testService) Name() string {
	return ""
}

func (testService) Desc() string {
	return ""
}

func (s testService) Needs() map[string]struct{} {
	return s.needs
}

func (s testService) Expose() interface{} {
	return nil
}

func (testService) Plug(map[string]interface{}) error {
	return nil
}

func (testService) Run(ctx context.Context, running chan struct{}, stopping chan struct{}) error {
	running <- struct{}{}
	<-ctx.Done()
	stopping <- struct{}{}

	return errors.WithStack(ctx.Err())
}

func TestManagerDeps(t *testing.T) {
	tests := []struct {
		name     string
		services []testService
		sid      string
		err      error
		want     []string
	}{
		{
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
		},
	}

	for _, test := range tests {
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

func TestManager(t *testing.T) {
	mgr := New()
	defer mgr.StopAll()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := mgr.Work(ctx)
		if err != nil && errors.Cause(err) != context.Canceled {
			t.Errorf(`manager: Work(ctx): error: %s`, err)
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

	err := mgr.Start("sleep")
	if got, want := errors.Cause(err), errors.Cause(ErrNotFound); got != want {
		t.Fatalf(`start unexisting: Start("sleep"): error = %s want %s`, got, want)
	}

	if err := mgr.Start("apps"); err != nil {
		t.Fatalf(`start existing: Start("apps"): error: %s`, err)
	}

	got, err := mgr.Status("apps")
	if err != nil {
		t.Fatalf(`start status: Status("apps"): error: %s`, err)
	}
	if want := Running; got != want {
		t.Fatalf(`start status: Status("apps") = %q want %q`, got, want)
	}

	got, err = mgr.Status("fs")
	if err != nil {
		t.Fatalf(`dep status: Status("fs"): error: %s`, err)
	}
	if want := Running; got != want {
		t.Fatalf(`dep status: Status("fs") = %q want %q`, got, want)
	}

	err = mgr.Stop("fs")
	if got, want := errors.Cause(err), errors.Cause(ErrNeeded); got != want {
		t.Fatalf(`stop with refs: Stop("fs"): error = %s want %s`, got, want)
	}

	if err := mgr.Stop("apps"); err != nil {
		t.Fatalf(`stop without refs: Stop("apps"): error: %s`, err)
	}

	got, err = mgr.Status("apps")
	if err != nil {
		t.Fatalf(`stopped status: Status("apps"): error: %s`, err)
	}
	if want := Stopped; got != want {
		t.Fatalf(`stopped status: Status("apps") = %q want %q`, got, want)
	}

	if err := mgr.Stop("fs"); err != nil {
		t.Fatalf(`stop no more refs: Stop("fs"): error: %s`, err)
	}

	if err := mgr.Start("apps"); err != nil {
		t.Fatalf(`start existing: Start("apps"): error: %s`, err)
	}

	got, err = mgr.Status("apps")
	if err != nil {
		t.Fatalf(`restart status: Status("apps"): error: %s`, err)
	}
	if want := Running; got != want {
		t.Fatalf(`restart status: Status("apps") = %q want %q`, got, want)
	}

	if err := mgr.Start("crypto"); err != nil {
		t.Fatalf(`make non-prunable: Start("crypto"): error: %s`, err)
	}

	if err := mgr.Stop("apps"); err != nil {
		t.Fatalf(`stop before prune: Stop("apps"): error: %s`, err)
	}

	mgr.Prune()

	got, err = mgr.Status("fs")
	if err != nil {
		t.Fatalf(`pruned status: Status("fs"): error: %s`, err)
	}
	if want := Stopped; got != want {
		t.Fatalf(`pruned status: Status("fs") = %q want %q`, got, want)
	}

	got, err = mgr.Status("net")
	if err != nil {
		t.Fatalf(`pruned status: Status("net"): error: %s`, err)
	}
	if want := Stopped; got != want {
		t.Fatalf(`pruned status: Status("net") = %q want %q`, got, want)
	}

	got, err = mgr.Status("crypto")
	if err != nil {
		t.Fatalf(`pruned status: Status("crypto"): error: %s`, err)
	}
	if want := Running; got != want {
		t.Fatalf(`pruned status: Status("crypto") = %q want %q`, got, want)
	}

	mgr.StopAll()

	got, err = mgr.Status("crypto")
	if err != nil {
		t.Fatalf(`stop all status: Status("crypto"): error: %s`, err)
	}
	if want := Stopped; got != want {
		t.Fatalf(`stop all status: Status("crypto") = %q want %q`, got, want)
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
