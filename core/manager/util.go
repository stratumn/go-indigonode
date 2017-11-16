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

package manager

import (
	"sort"
)

// sortedSetKeys returns the keys of a string set sorted alphabetically.
func sortedSetKeys(set map[string]struct{}) []string {
	if set == nil {
		return nil
	}

	var keys []string
	for k := range set {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// sortedSetKeys returns the keys of a process map sorted alphabetically.
func sortedProcessKeys(set map[string]*process) []string {
	if set == nil {
		return nil
	}

	var keys []string
	for k := range set {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
