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

package storetestcases

import (
	"github.com/stratumn/go-indigocore/cs"
)

// ContainsSegment returns true if the list contains the given segment.
// It only compares the link hash.
func ContainsSegment(segments []cs.Segment, segment cs.Segment) bool {
	if len(segments) == 0 {
		return false
	}

	for _, s := range segments {
		if s.GetLinkHashString() == segment.GetLinkHashString() {
			return true
		}
	}

	return false
}
