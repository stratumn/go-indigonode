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

package cli

import (
	"context"

	"github.com/stratumn/alice/script"
)

// Content represents console content used to find suggestions.
type Content interface {
	// TextBeforeCursor returns all the text before the cursor.
	TextBeforeCursor() string

	// GetWordBeforeCursor returns the word before the current cursor
	// position.
	GetWordBeforeCursor() string
}

// Cmd is an interface that must be implemented by commands.
type Cmd interface {
	// Name returns the name of the command (used by `help command`
	// to find the command).
	Name() string

	// Short returns a short description of the command.
	Short() string

	// Long returns a long description of the command.
	Long() string

	// Use returns a short string showing how to use the command.
	Use() string

	// LongUse returns a long string showing how to use the command.
	LongUse() string

	// Suggest gives a chance for the command to add auto-complete
	// suggestions for the current content.
	Suggest(Content) []Suggest

	// Match returns whether the command can execute against the given
	// command name.
	Match(string) bool

	// Exec executes the commands.
	Exec(*script.InterpreterContext, CLI) (script.SExp, error)
}

// Suggest implements a suggestion.
type Suggest struct {
	// Text is the text that will replace the current word.
	Text string

	// Desc is a short description of the suggestion.
	Desc string
}

// CLI represents a command line interface.
type CLI interface {
	// Config returns the configuration.
	Config() Config

	// Console returns the console.
	Console() *Console

	// Commands returns all the commands.
	Commands() []Cmd

	// Address returns the address of the API server.
	Address() string

	// Connect connects to the API server.
	Connect(ctx context.Context, addr string) error

	// Disconnect closes the API client connection.
	Disconnect() error

	// Start starts the command line interface until the user kills it.
	Start(context.Context)

	// Exec executes the given input.
	Exec(ctx context.Context, in string) error

	// Run executes the given input, handling errors and cancellation
	// signals.
	Run(ctx context.Context, in string)

	// Suggest finds all command suggestions.
	Suggest(cnt Content) []Suggest

	// DidJustExecute returns true the first time it is called after a
	// command executed. This is a hack used by the VT100 prompt to hide
	// suggestions after a command was executed.
	DidJustExecute() bool
}
