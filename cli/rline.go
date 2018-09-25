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
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
)

// Register the prompt.
func init() {
	registerPrompt("readline", func(ctx context.Context, cli CLI) {
		p := Rline{c: cli}
		p.Run(ctx, nil)
	})
}

// Rline implements a quick-and-dirty prompt using readline.
type Rline struct {
	c     CLI
	line  []rune
	pos   int
	start int
}

// NewRline creates a new readline prompt.
func NewRline(cli CLI) *Rline {
	return &Rline{c: cli}
}

// Run launches the prompt until it is killed.
func (p *Rline) Run(ctx context.Context, rc io.ReadCloser) {
	rl, err := readline.NewEx(&readline.Config{
		Stdin:             rc,
		Prompt:            "node> ",
		HistorySearchFold: true,
		AutoComplete:      p,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := rl.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close readline: %s.\n", err)
		}
	}()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		p.c.Run(ctx, line)
	}
}

// Do finds suggestions.
func (p *Rline) Do(line []rune, pos int) (newLine [][]rune, length int) {
	p.line = line
	p.pos = pos
	p.begin()
	sugs := p.c.Suggest(p)
	offset := pos - p.start

	var matches [][]rune

	for _, s := range sugs {
		// Only accept prefix.
		if strings.HasPrefix(s.Text, p.GetWordBeforeCursor()) {
			matches = append(matches, []rune(s.Text)[offset:])
		}
	}

	return matches, pos
}

// TextBeforeCursor returns the content of the line before the the current
// cursor position.
func (p *Rline) TextBeforeCursor() string {
	return string(p.line[:p.pos])
}

// GetWordBeforeCursor return the word before the current cursor position.
func (p *Rline) GetWordBeforeCursor() string {
	return string(p.line[p.start:p.pos])
}

// begin finds the start offset of the current word.
func (p *Rline) begin() {
	p.start = p.pos

	for p.start > 0 {
		if unicode.IsSpace(p.line[p.start-1]) {
			break
		}
		p.start--
	}
}
