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

package assert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// WaitUntil waits for a condition to happen or fails after a small time.
// This is useful when you want to assert that asynchronous conditions need
// to be verified.
func WaitUntil(t *testing.T, cond func() bool, message string) {
	condChan := make(chan struct{})
	go func() {
		for {
			if cond() {
				condChan <- struct{}{}
				return
			}

			<-time.After(10 * time.Millisecond)
		}
	}()

	select {
	case <-condChan:
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "waitUntil() condition failed:", message)
	}
}
