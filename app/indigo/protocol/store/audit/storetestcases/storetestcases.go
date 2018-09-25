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

// Package storetestcases defines test cases to test audit stores.
package storetestcases

import (
	"testing"

	"github.com/stratumn/go-node/app/indigo/protocol/store/audit"
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
