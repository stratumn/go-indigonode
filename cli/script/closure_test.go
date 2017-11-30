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

package script

import (
	"testing"

	"github.com/pkg/errors"
)

func sym(s string) *SExp {
	return &SExp{Type: TypeSym, Str: s}
}

func TestNewClosure_env(t *testing.T) {
	c := NewClosure(OptEnv("$", []string{"a=one", "b=two"}))

	got, ok := c.Get("$a")
	if !ok {
		t.Errorf(`c.Get("$a"): ok = %v want true`, ok)
	} else if got.String() != "(\"one\")" {
		t.Errorf(`c.Get("$a") = %v want ("one")`, got.String())
	}

	got, ok = c.Get("$b")
	if !ok {
		t.Errorf(`c.Get("$b"): ok = %v want true`, ok)
	} else if got.String() != "(\"two\")" {
		t.Errorf(`c.Get("$b") = %v want ("two")`, got.String())
	}
}

func TestClosure_Set(t *testing.T) {
	c := NewClosure()

	c.Set("a", sym("one"))
	got, ok := c.Get("a")
	if !ok {
		t.Errorf(`c.Get("a"): ok = %v want true`, ok)
	} else if got.String() != "(one)" {
		t.Errorf(`c.Get("a") = %v want (one)`, got.String())
	}
}

func TestClosure_Set_overwriteParent(t *testing.T) {
	c1 := NewClosure()
	c1.Set("a", sym("one"))

	c2 := NewClosure(OptParent(c1))
	c2.Set("a", sym("two"))

	got, ok := c1.Get("a")
	if !ok {
		t.Errorf(`c1.Get("a"): ok = %v want true`, ok)
	} else if got.String() != "(two)" {
		t.Errorf(`c1.Get("a") = %v want (two)`, got.String())
	}

	got, ok = c2.Get("a")
	if !ok {
		t.Errorf(`c2.Get("a"): ok = %v want true`, ok)
	} else if got.String() != "(two)" {
		t.Errorf(`c2.Get("a") = %v want (two)`, got.String())
	}
}

func TestClosure_Set_new(t *testing.T) {
	c1 := NewClosure()

	c2 := NewClosure(OptParent(c1))
	c2.Set("a", sym("one"))

	_, ok := c1.Get("a")
	if ok {
		t.Errorf(`c1.Get("a"): ok = %v want false`, ok)
	}

	got, ok := c2.Get("a")
	if !ok {
		t.Errorf(`c2.Get("a"): ok = %v want true`, ok)
	} else if got.String() != "(one)" {
		t.Errorf(`c2.Get("a") = %v want (one)`, got.String())
	}
}

func TestClosure_SetLocal(t *testing.T) {
	c1 := NewClosure()
	c1.Set("a", sym("one"))

	c2 := NewClosure(OptParent(c1))
	c2.SetLocal("a", sym("two"))

	got, ok := c1.Get("a")
	if !ok {
		t.Errorf(`c1.Get("a"): ok = %v want true`, ok)
	} else if got.String() != "(one)" {
		t.Errorf(`c1.Get("a") = %v want (one)`, got.String())
	}

	got, ok = c2.Get("a")
	if !ok {
		t.Errorf(`c2.Get("a"): ok = %v want true`, ok)
	} else if got.String() != "(two)" {
		t.Errorf(`c2.Get("a") = %v want (two)`, got.String())
	}
}

func TestClosure_Resolve(t *testing.T) {
	c := NewClosure()

	c.Set("a", sym("one"))
	got, err := c.Resolve(sym("a"))
	if err != nil {
		t.Error(`c.Resolve(sym("a")): error: `, err)
	} else if got.String() != "(one)" {
		t.Errorf(`c.Resolve(sym("a")) = %v want (one)`, got.String())
	}

	_, err = c.Resolve(sym("b"))
	if got, want := errors.Cause(err), ErrSymNotFound; got != want {
		t.Errorf(`c.Resolve(sym("b")): error = %v want %v`, got, want)
	}
}

func TestClosure_Resolve_resolver(t *testing.T) {
	c := NewClosure(OptResolver(ResolveName))

	c.Set("a", sym("one"))
	got, err := c.Resolve(sym("a"))
	if err != nil {
		t.Error(`c.Resolve(sym("a")): error: `, err)
	} else if got.String() != "(one)" {
		t.Errorf(`c.Resolve(sym("a")) = %v want (one)`, got.String())
	}

	got, err = c.Resolve(sym("b"))
	if err != nil {
		t.Error(`c.Resolve(sym("b")): error: `, err)
	} else if got.String() != `("b")` {
		t.Errorf(`c.Resolve(sym("b")) = %v want ("b")`, got.String())
	}
}
