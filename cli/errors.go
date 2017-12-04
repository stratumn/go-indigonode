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

	// ErrInvalidInstr is returned when the user entered an invalid
	// instruction.
	ErrInvalidInstr = errors.New("the instruction is invalid")

	// ErrCmdNotFound is returned when a command was not found.
	ErrCmdNotFound = errors.New("the command was not found")

	// ErrInvalidExitCode is returned when an invalid exit code was given.
	ErrInvalidExitCode = errors.New("the exit code is invalid")

	// ErrUnsupportedReflectType is returned when a type is not currently
	// supported by reflection.
	ErrUnsupportedReflectType = errors.New("the type is not currently supported by reflection")

	// ErrParse is returned when a value could not be parsed.
	ErrParse = errors.New("could not parse the value")

	// ErrNotFunc is returned when an S-Expression is not a function.
	ErrNotFunc = errors.New("the expression is not a function")
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
