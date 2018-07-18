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

// Package benchmark defines integration benchmarks for Indigo Node.
//
// The tests are done by connecting via the API to Indigo nodes launched in
// separate processes.
package benchmark

import "time"

const (
	// NumNodes is the number of nodes launched before each test.
	NumNodes = 11

	// MaxDuration is the maximum allowed duration for each test.
	MaxDuration = 2 * time.Minute

	// SessionDir is the directory where session data will be saved.
	SessionDir = "../tmp/benchmark"
)
