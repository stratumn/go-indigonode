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

//+build !lint

// Package system defines system tests for Stratumn Node.
//
// The tests are done by connecting via the API to Stratumn nodes launched in
// separate processes.
package system

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/go-node/test/session"
	"github.com/stretchr/testify/assert"
)

const (
	// NumNodes is the number of nodes launched before each test.
	NumNodes = 11

	// MaxDuration is the maximum allowed duration for a test.
	MaxDuration = 2 * time.Minute

	// SessionDir is the directory where session data will be saved.
	SessionDir = "../tmp/system"
)

// Test wrap session.Session with a context and handles errors.
func Test(t *testing.T, fn session.Tester) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	err := session.Run(ctx, SessionDir, NumNodes, session.SystemCfg(), fn)
	assert.NoError(t, err, "Session()")
}
