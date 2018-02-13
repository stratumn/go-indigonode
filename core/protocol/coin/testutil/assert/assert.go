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
