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
	"github.com/pkg/errors"
)

var (
	// ErrTypeMissingArg is emitted when a function argument is missing.
	ErrTypeMissingArg = errors.New("missing function argument")

	// ErrTypeExtraArg is emitted when a function call has too many
	// arguments.
	ErrTypeExtraArg = errors.New("extra function argument")

	// ErrCarNil is emitted when the car of a cell is nil.
	ErrCarNil = errors.New("car cannot be nil")

	// ErrNotSymbol is emitted when a symbol is expected.
	ErrNotSymbol = errors.New("not a symbol")

	// ErrNotCell is emitted when a cell is expected.
	ErrNotCell = errors.New("not a cell")

	// ErrNotList is emitted when a list is expected.
	ErrNotList = errors.New("not a list")

	// ErrNil is emitted when a value cannot be nil.
	ErrNil = errors.New("value cannot be nil")

	// ErrCouldBeNil is emitted when a value could be nil.
	ErrCouldBeNil = errors.New("value could be nil")

	// ErrBound is emitted when a symbol is already locally bound.
	ErrBound = errors.New("symbol is already bound locally")

	// ErrNotBound is emitted when a symbol is not bound to a value.
	ErrNotBound = errors.New("symbol is not bound to a value")
)

// PrimitiveType is a primitive type.
type PrimitiveType uint8

// Primitive types.
const (
	PrimitiveInvalid PrimitiveType = iota
	PrimitiveString
	PrimitiveInt64
	PrimitiveBool
	PrimitiveList
)

// Maps primitive types to their names.
var primitiveTypeMap = map[PrimitiveType]string{
	PrimitiveInvalid: "invalid",
	PrimitiveString:  "string",
	PrimitiveInt64:   "int64",
	PrimitiveBool:    "bool",
	PrimitiveList:    "cell",
}

// String returns a string representation of the primitive type.
func (t PrimitiveType) String() string {
	return primitiveTypeMap[t]
}

// TypeInfo contains information about a type.
type TypeInfo struct {
	// Type is the primitive type, for instance bool, function, or list.
	Type PrimitiveType

	// Subtype is for compound types, such as list<int64> or
	// function()<bool>.
	Subtype *TypeInfo

	// Params is currently used for function arguments, for instance
	// function(<string>,<int64>).
	Params []*TypeInfo
}
