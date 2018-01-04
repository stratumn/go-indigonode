// Copyright © 2017-2018 Stratumn SAS
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
	"github.com/stretchr/testify/assert"
)

func sym(s string) SExp {
	return Symbol(s, Meta{})
}

func TestNewClosure_env(t *testing.T) {
	c := NewClosure(ClosureOptEnv([]string{"a=one", "b=two"}))

	got, ok := c.Get("a")
	assert.True(t, ok, `c.Get("a")`)
	assert.Equal(t, "\"one\"", got.String(), `c.Get("a")`)

	got, ok = c.Get("b")
	assert.True(t, ok, `c.Get("b")`)
	assert.Equal(t, "\"two\"", got.String(), `c.Get("b")`)
}

func TestClosure_Set(t *testing.T) {
	c := NewClosure()

	c.Set("a", sym("one"))
	got, ok := c.Get("a")
	assert.True(t, ok, `c.Get("a")`)
	assert.Equal(t, "one", got.String(), `c.Get("a")`)
}

func TestClosure_Set_overwriteParent(t *testing.T) {
	c1 := NewClosure()
	c1.Set("a", sym("one"))

	c2 := NewClosure(ClosureOptParent(c1))
	c2.Set("a", sym("two"))

	got, ok := c1.Get("a")
	assert.True(t, ok, `c1.Get("a")`)
	assert.Equal(t, "two", got.String(), `c1.Get("a")`)

	got, ok = c2.Get("a")
	assert.True(t, ok, `c2.Get("a")`)
	assert.Equal(t, "two", got.String(), `c2.Get("a")`)
}

func TestClosure_Set_new(t *testing.T) {
	c1 := NewClosure()

	c2 := NewClosure(ClosureOptParent(c1))
	c2.Set("a", sym("one"))

	_, ok := c1.Get("a")
	assert.False(t, ok, `c1.Get("a")`)

	got, ok := c2.Get("a")
	assert.True(t, ok, `c2.Get("a")`)
	assert.Equal(t, "one", got.String(), `c2.Get("a")`)
}

func TestClosure_SetLocal(t *testing.T) {
	c1 := NewClosure()
	c1.Set("a", sym("one"))

	c2 := NewClosure(ClosureOptParent(c1))
	c2.SetLocal("a", sym("two"))

	got, ok := c1.Get("a")
	assert.True(t, ok, `c1.Get("a")`)
	assert.Equal(t, "one", got.String(), `c1.Get("a")`)

	got, ok = c2.Get("a")
	assert.True(t, ok, `c2.Get("a")`)
	assert.Equal(t, "two", got.String(), `c2.Get("a")`)
}

func TestClosure_Local(t *testing.T) {
	c1 := NewClosure()
	c1.Set("a", sym("one"))

	c2 := NewClosure(ClosureOptParent(c1))

	_, ok := c2.Local("a")
	assert.False(t, ok, `c2.Local("a")`)
}

func TestClosure_Resolve(t *testing.T) {
	c := NewClosure()

	c.Set("a", sym("one"))
	got, err := c.Resolve(sym("a"))
	assert.NoError(t, err, `c.Resolve(sym("a"))`)
	assert.Equal(t, "one", got.String(), `c.Resolve(sym("a"))`)

	_, err = c.Resolve(sym("b"))
	assert.Equal(t, ErrSymNotFound, errors.Cause(err))
}

func TestClosure_Resolve_resolver(t *testing.T) {
	c := NewClosure(ClosureOptResolver(ResolveName))

	c.Set("a", sym("one"))
	got, err := c.Resolve(sym("a"))
	assert.NoError(t, err, `c.Resolve(sym("a"))`)
	assert.Equal(t, "\"a\"", got.String(), `c.Resolve(sym("a"))`)

	got, err = c.Resolve(sym("b"))
	assert.NoError(t, err, `c.Resolve(sym("b"))`)
	assert.Equal(t, "\"b\"", got.String(), `c.Resolve(sym("b"))`)
}
