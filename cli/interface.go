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

import (
	"context"

	"github.com/stratumn/go-indigonode/script"
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

// EventListener is an interface that must be implemented by event listeners.
type EventListener interface {
	// Start connects to the corresponding event emitter and continuously
	// listens for events and displays them.
	// It disconnects from the event emitter when the context is cancelled.
	Start(context.Context) error

	// Connected returns true if the listener is connected to the event
	// emitter and ready to receive events.
	Connected() bool
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
