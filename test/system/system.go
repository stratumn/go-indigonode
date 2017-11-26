// Copyright © 2017 Stratumn SAS
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

// Package system defines system tests for Alice.
//
// The tests are done by connecting via the API to Alice nodes launched in
// separate processes.
package system

import "time"

const (
	// NumNodes is the number of nodes launched before each test.
	NumNodes = 11

	// MaxDuration is the maximum allowed duration for a test.
	MaxDuration = 2 * time.Minute

	// SessionDir is the directory where session data will be saved.
	SessionDir = "../tmp/system"
)
