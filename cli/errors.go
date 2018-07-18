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

package cli

import "github.com/pkg/errors"

var (
	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("the configuration is invalid")

	// ErrPromptNotFound is returned when the requested prompt backend was
	// not found.
	ErrPromptNotFound = errors.New("the requested prompt was not found")

	// ErrDisconnected is returned when the CLI is not connected to the
	// API.
	ErrDisconnected = errors.New("the client is not connected to API")

	// ErrCmdNotFound is returned when a command was not found.
	ErrCmdNotFound = errors.New("the command was not found")

	// ErrInvalidExitCode is returned when an invalid exit code was given.
	ErrInvalidExitCode = errors.New("the exit code is invalid")

	// ErrUnsupportedReflectType is returned when a type is not currently
	// supported by reflection.
	ErrUnsupportedReflectType = errors.New("the type is not currently supported by reflection")

	// ErrReflectParse is returned when a value could not be parsed by a
	// reflector.
	ErrReflectParse = errors.New("could not parse the value")

	// ErrFieldNotFound is returned when the field of a message was not
	// found.
	ErrFieldNotFound = errors.New("the field was not found")
)

// UseError represents a usage error.
//
// Make it a pointer receiver so that errors.Cause() can be used to retrieve
// and modify the use string after the error is created. This is actually
// being done by the executor after a command if it returns a usage error.
type UseError struct {
	msg string
	use string
}

// NewUseError creates a new usage error.
func NewUseError(msg string) error {
	return errors.WithStack(&UseError{msg: msg})
}

// Error returns the error message.
func (err *UseError) Error() string {
	return "invalid usage: " + err.msg
}

// Use returns the usage message.
func (err *UseError) Use() string {
	return err.use
}

// stackTracer is used to get the stack trace from errors created by the
// github.com/pkg/errors package.
type stackTracer interface {
	StackTrace() errors.StackTrace
}

// StackTrace returns the stack trace from an error created using the
// github.com/pkg/errors package.
func StackTrace(err error) errors.StackTrace {
	if err, ok := err.(stackTracer); ok {
		return err.StackTrace()
	}

	return nil
}
