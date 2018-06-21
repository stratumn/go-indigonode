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

package service

import "net"

// isGRPCShutdownBug is used to detect if an exit error is due to a bug in gRPC
// after shutting down.
//
// TODO: Check if still needed when updating gRPC.
// EDIT: Apparently it's expected behavior.
func isGRPCShutdownBug(err error) bool {
	if e, ok := err.(*net.OpError); ok {
		return e.Op == "accept"
	}
	return false
}
