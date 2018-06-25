// Copyright 2017 Stratumn SAS. All rights reserved.
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

// Package storetestcases defines test cases to test audit stores.
package storetestcases

import (
	"testing"

	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
)

// Factory wraps functions to allocate and free an audit store
// and is used to run the test suite.
type Factory struct {
	// New creates an audit store.
	New func() (audit.Store, error)

	// Free is an optional function to free an audit store.
	Free func(audit.Store)
}

// RunTests runs the audit store test suite.
func (f Factory) RunTests(t *testing.T) {
	t.Run("AddSegment", f.TestAddSegment)
	t.Run("GetByPeer", f.TestGetByPeer)
}
