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

package monitoring

// Status codes.
const (
	StatusCodeOK                 = 0
	StatusCodeCancelled          = 1
	StatusCodeUnknown            = 2
	StatusCodeInvalidArgument    = 3
	StatusCodeDeadlineExceeded   = 4
	StatusCodeNotFound           = 5
	StatusCodeAlreadyExists      = 6
	StatusCodePermissionDenied   = 7
	StatusCodeResourceExhausted  = 8
	StatusCodeFailedPrecondition = 9
	StatusCodeAborted            = 10
	StatusCodeOutOfRange         = 11
	StatusCodeUnimplemented      = 12
	StatusCodeInternal           = 13
	StatusCodeUnavailable        = 14
	StatusCodeDataLoss           = 15
	StatusCodeUnauthenticated    = 16
)

// Status is the status of a Span.
type Status struct {
	Code    int32
	Message string
}

// NewStatus creates a new status object.
func NewStatus(code int32, message string) Status {
	return Status{
		Code:    code,
		Message: message,
	}
}
