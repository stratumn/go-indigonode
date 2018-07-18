// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package script

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// ResolveHandler resolves symbols.
//
// If it returns an error of the type ErrSymNotFound, the symbol will be
// resolved from its value in the closure. If it returns any other error,
// the symbol will not be resolved and an error will be produced.
//
// It makes it possible to handle special cases such as if you want variables
// to have a prefix like a dollar sign.
type ResolveHandler func(closure *Closure, sym SExp) (SExp, error)

// ResolveName resolves symbols with their names.
func ResolveName(_ *Closure, sym SExp) (SExp, error) {
	return String(sym.MustSymbolVal(), sym.Meta()), nil
}

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
func ClosureOptEnv(env []string) ClosureOpt {
	return func(c *Closure) {
		for _, e := range env {
			parts := strings.Split(e, "=")
			c.values[parts[0]] = String(parts[1], Meta{})
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
	if c.resolve != nil {
		v, err := c.resolve(c, sym)
		switch {
		case errors.Cause(err) == ErrSymNotFound:
			// Keep going, resolve from closure instead.
		case err != nil:
			return nil, err
		default:
			return v, nil
		}
	}

	v, ok := c.Get(sym.MustSymbolVal())
	if !ok {
		return nil, WrapError(ErrSymNotFound, sym.Meta(), sym.MustSymbolVal())
	}

	return v, nil
}
