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
	"strings"
	"sync"
)

// ClosureOpt is a closure options.
type ClosureOpt func(*Closure)

// ClosureOptParent sets the parent closure.
func ClosureOptParent(parent *Closure) ClosureOpt {
	return func(c *Closure) {
		c.parent = parent

		if parent != nil && c.resolve == nil {
			c.resolve = parent.resolve
		}
	}
}

// ClosureOptEnv sets values from environment variables of the form "key=value".
//
// The given string will be prefixed to variable names.
func ClosureOptEnv(prefix string, env []string) ClosureOpt {
	return func(c *Closure) {
		for _, e := range env {
			parts := strings.Split(e, "=")
			c.values[prefix+parts[0]] = String(parts[1], Meta{})
		}
	}
}

// ClosureOptResolver sets the resolver which is used if a symbol is not found.
func ClosureOptResolver(resolve ResolveHandler) ClosureOpt {
	return func(c *Closure) {
		c.resolve = resolve
	}
}

// Closure stores local values and a parent closure.
type Closure struct {
	parent  *Closure
	resolve ResolveHandler

	mu     sync.RWMutex
	values map[string]SExp
}

// NewClosure creates a new closure with an optional parent.
func NewClosure(opts ...ClosureOpt) *Closure {
	c := &Closure{
		values: map[string]SExp{},
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

// Get returns a value.
//
// It travels up the closures until it finds the key.
func (c *Closure) Get(key string) (SExp, bool) {
	for curr := c; curr != nil; curr = curr.parent {
		curr.mu.RLock()
		if v, ok := curr.values[key]; ok {
			curr.mu.RUnlock()
			return v, true
		}
		curr.mu.RUnlock()
	}

	return nil, false
}

// Local returns a local value.
func (c *Closure) Local(key string) (SExp, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.values[key]

	return v, ok
}

// Set sets a value.
//
// It travels up the closure until it finds the key. If it doesn't find one,
// it sets a local.
func (c *Closure) Set(key string, val SExp) {
	for curr := c; curr != nil; curr = curr.parent {
		curr.mu.Lock()
		if _, ok := curr.values[key]; ok {
			curr.values[key] = val
			curr.mu.Unlock()
			return
		}
		curr.mu.Unlock()
	}

	c.SetLocal(key, val)
}

// SetLocal sets a local value.
func (c *Closure) SetLocal(key string, val SExp) {
	c.mu.Lock()
	c.values[key] = val
	c.mu.Unlock()
}

// Resolve resolves a symbol.
func (c *Closure) Resolve(sym SExp) (SExp, error) {
	v, ok := c.Get(sym.MustSymbolVal())
	if !ok {
		if c.resolve != nil {
			return c.resolve(sym)
		}

		return nil, WrapError(ErrSymNotFound, sym.Meta(), sym.MustSymbolVal())
	}

	return v, nil
}
