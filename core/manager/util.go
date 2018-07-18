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

// sortedStateKeys returns the keys of a state map sorted alphabetically.
func sortedStateKeys(set map[string]*state) []string {
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

// sortedFriendlyKeys returns the keys of a friendly map sorted alphabetically.
func sortedFriendlyKeys(set map[string]Friendly) []string {
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
