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
	"bytes"

	pb "github.com/stratumn/alice/pb/indigo/store"
)

// ContainsLink returns true if the list contains the given link.
// It only compares the link bytes, not the From and Signature.
func ContainsLink(links []*pb.SignedLink, link *pb.SignedLink) bool {
	if len(links) == 0 {
		return false
	}

	for _, l := range links {
		if bytes.Equal(l.Link, link.Link) {
			return true
		}
	}

	return false
}
