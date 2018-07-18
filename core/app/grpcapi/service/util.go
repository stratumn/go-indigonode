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
