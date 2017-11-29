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

// Package script defines types for a very simple script interpreter.
//
// It is designed with shell-like scripting in mind. It uses simplified
// S-Expressions that can hold either string or list values.
package script

import "github.com/pkg/errors"

var (
	// ErrNil is returned when trying to evaluate an empty list.
	ErrNil = errors.New("cannot evaluate and empty list")

	// ErrSymNotFound is returned when a symbol could not be resolved.
	ErrSymNotFound = errors.New("could not resolve symbol")
)
