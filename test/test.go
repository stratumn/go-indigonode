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

package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// WaitUntil waits for duration for the condition function to evaluate to true
// to return or fails after the delay has elapsed.
func WaitUntil(t *testing.T, duration time.Duration, interval time.Duration, cond func() bool, message string) {
	condChan := make(chan struct{})
	go func() {
		for {
			if cond() {
				condChan <- struct{}{}
				return
			}

			<-time.After(interval)
		}
	}()

	select {
	case <-condChan:
	case <-time.After(duration):
		assert.Fail(t, "waitUntil() condition failed:", message)
	}
}
