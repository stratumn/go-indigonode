// Copyright © 2017 Stratumn SAS
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
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// BasicCmd implements a basic command that matches the first word of the
// input.
type BasicCmd struct {
	// Name is the name of the command, used to match a word to the command
	// amongst other things.
	Name string

	// Short is a short description of the command.
	Short string

	// Use is a short usage string, for example `help [command]`. If not
	// defined, the name of the command will be used.
	Use string

	// Flags can be defined to return a set of flags for the command.
	Flags func() *pflag.FlagSet

	// Exec is a function that must be defined to execute the command.
	Exec func(context.Context, *CLI, []string, *pflag.FlagSet) error
}

// BasicCmdWrapper wraps a basic command to make it compatible with the Cmd
// interface. It also adds a help flag and deals with flag errors.
type BasicCmdWrapper struct {
	Cmd BasicCmd
}

// Name returns the name string.
func (cmd BasicCmdWrapper) Name() string {
	return cmd.Cmd.Name
}

// Short returns the short description string.
func (cmd BasicCmdWrapper) Short() string {
	return cmd.Cmd.Short
}

// Long returns the long description which is the short description followed
// by the long usage string of the command.
func (cmd BasicCmdWrapper) Long() string {
	return cmd.Short() + "\n\n" + cmd.LongUse()
}

// Use returns the short usage string or the name string if not specified.
func (cmd BasicCmdWrapper) Use() string {
	if cmd.Cmd.Use != "" {
		return cmd.Cmd.Use
	}

	return cmd.Cmd.Name
}

// LongUse returns the long usage string of the command which is the short
// usage string followed by the flags usage string.
func (cmd BasicCmdWrapper) LongUse() string {
	flags := cmd.createFlags()

	long := "Usage:\n"
	long += "  " + cmd.Use()
	long += "\n\nFlags:\n"
	long += flags.FlagUsages()

	return long
}

// Suggest suggests the command whenever of the command's text contains the
// word before the current position of cursor. It also makes suggestions for
// flags.
func (cmd BasicCmdWrapper) Suggest(c Content) []Suggest {
	// Find instruction and split argument vector.
	instrs := strings.Split(c.TextBeforeCursor(), ";")
	if len(instrs) < 1 {
		return nil
	}

	instr := strings.TrimSpace(instrs[len(instrs)-1])
	args := strings.Split(instr, " ")

	// Get the command name (first word before space) and ignore case.
	name := ""
	if len(args) > 0 {
		name = strings.ToLower(strings.TrimSpace(args[0]))
	}

	// Get the word being typed.
	word := strings.ToLower(c.GetWordBeforeCursor())

	// Look for flags if the command name is an exact match.
	if name == cmd.Name() {
		// If the the current word is empty or starts with a dash,
		// make suggestions for flags.
		if word == "" || word[0] == '-' {
			return cmd.suggestFlags(word)
		}

		// Not much we can suggest at this point.
		return nil
	}

	canSugCmd := false

	if instr == "" {
		// Suggest command if instruction is blank.
		canSugCmd = true
	} else if len(args) < 2 && word != "" || name == "help" {
		// Match command if command is being typed or the instruction
		// is a help command.
		canSugCmd = strings.Contains(strings.ToLower(cmd.Cmd.Name), word)
	}

	if canSugCmd {
		return []Suggest{{
			Text: cmd.Cmd.Name,
			Desc: cmd.Cmd.Short,
		}}
	}

	return nil
}

// suggestFlags makes suggestions for flags.
func (cmd BasicCmdWrapper) suggestFlags(word string) []Suggest {
	var matches []Suggest

	flags := cmd.createFlags()

	if word == "-" || word == "--" {
		// If the word is just one or two dashes, add all available
		// flags.
		flags.VisitAll(func(f *pflag.Flag) {
			matches = append(matches, Suggest{
				Text: "--" + f.Name,
				Desc: f.Usage,
			})
		})
	} else {
		// Otherwise match flag names.
		fname := strings.TrimPrefix(strings.TrimPrefix(word, "-"), "-")

		flags.VisitAll(func(f *pflag.Flag) {
			if strings.Contains(strings.ToLower(f.Name), fname) {
				matches = append(matches, Suggest{
					Text: "--" + f.Name,
					Desc: f.Usage,
				})
			}
		})
	}

	return matches
}

// Match returns whether the first word exacly matches the text of the command.
func (cmd BasicCmdWrapper) Match(in string) bool {
	return strings.Split(in, " ")[0] == cmd.Cmd.Name
}

// Exec executes the basic command.
func (cmd BasicCmdWrapper) Exec(ctx context.Context, cli *CLI, in string) error {
	argv := strings.Split(in, " ")
	for i, v := range argv {
		argv[i] = strings.TrimSpace(v)
	}

	flags := cmd.createFlags()

	// Discard flags output, we do our own error handling.
	flags.SetOutput(ioutil.Discard)

	if err := flags.Parse(argv[1:]); err != nil {
		return NewUseError(err.Error())
	}

	help, err := flags.GetBool("help")
	if err != nil {
		return errors.WithStack(err)
	}

	if help {
		// Invoke help command, pass the command name and remaining
		// args.
		args := append([]string{cmd.Name()}, flags.Args()...)
		return Help.Cmd.Exec(ctx, cli, args, flags)
	}

	return cmd.Cmd.Exec(ctx, cli, flags.Args(), flags)
}

// createFlags creates a set of flags for the command. There will be at least
// the help flag.
func (cmd BasicCmdWrapper) createFlags() (flags *pflag.FlagSet) {
	if cmd.Cmd.Flags != nil {
		flags = cmd.Cmd.Flags()
	}

	if flags == nil {
		flags = pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
	}

	// Add help flag.
	flags.BoolP("help", "h", false, "Invoke help on command")

	return
}