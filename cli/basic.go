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
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/stratumn/go-indigonode/script"
)

// BasicCmd is designed for simple commands to just need to parse flags and
// output text.
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

	// NoFlags disables flag parsing as well as the help flag.
	NoFlags bool

	// Exec is a function that executes the command against string
	// arguments and outputs to a writer.
	Exec func(*BasicContext) error
}

// BasicContext is passed to Exec when executing a basic command.
type BasicContext struct {
	Ctx    context.Context
	CLI    CLI
	Writer io.Writer
	Args   []string
	Flags  *pflag.FlagSet
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

	if !cmd.Cmd.NoFlags {
		long += "\n\nFlags:\n"
		long += strings.TrimSuffix(flags.FlagUsages(), "\n")
	}

	return long
}

// Suggest suggests the command whenever of the command's text contains the
// word before the current position of cursor. It also makes suggestions for
// flags.
func (cmd BasicCmdWrapper) Suggest(c Content) []Suggest {
	instrs := strings.Split(c.TextBeforeCursor(), "(")
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
	trimmed := strings.TrimPrefix(word, "(")
	prefix := word[:len(word)-len(trimmed)]

	// Look for flags if the command name is an exact match.
	if !cmd.Cmd.NoFlags && name == cmd.Name() {
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
		canSugCmd = strings.Contains(strings.ToLower(cmd.Cmd.Name), trimmed)
	}

	if canSugCmd {
		return []Suggest{{
			Text: prefix + cmd.Cmd.Name,
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

// Match returns whether the string matches the command name.
func (cmd BasicCmdWrapper) Match(name string) bool {
	return name == cmd.Cmd.Name
}

// Exec executes the basic command.
func (cmd BasicCmdWrapper) Exec(ctx *script.InterpreterContext, cli CLI) (script.SExp, error) {
	argv, err := ctx.EvalListToStrings(ctx, ctx.Args, false)
	if err != nil {
		return nil, err
	}

	var (
		flags *pflag.FlagSet
		help  bool
	)

	if !cmd.Cmd.NoFlags {
		flags = cmd.createFlags()

		// Discard flags output, we do our own error handling.
		flags.SetOutput(ioutil.Discard)

		if err := flags.Parse(argv); err != nil {
			return nil, NewUseError(err.Error())
		}

		help, err = flags.GetBool("help")
		if err != nil {
			return nil, errors.WithStack(err)
		}

	}

	buf := bytes.NewBuffer(nil)

	basicCtx := &BasicContext{
		Ctx:    ctx.Ctx,
		CLI:    cli,
		Writer: buf,
		Flags:  flags,
	}

	if help {
		// Invoke help command, pass the command name and remaining
		// args.
		basicCtx.Args = append([]string{cmd.Name()}, flags.Args()...)
		err = Help.Cmd.Exec(basicCtx)
	} else {
		if flags == nil {
			basicCtx.Args = argv
		} else {
			basicCtx.Args = flags.Args()
		}

		err = cmd.Cmd.Exec(basicCtx)
	}

	if err != nil {
		return nil, err
	}

	return script.String(buf.String(), ctx.Meta), nil
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
	if !cmd.Cmd.NoFlags {
		flags.BoolP("help", "h", false, "Invoke help on command")
	}

	return
}
