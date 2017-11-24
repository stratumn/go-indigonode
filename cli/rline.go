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
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
)

// Register the prompt.
func init() {
	registerPrompt("readline", func(ctx context.Context, cli CLI) {
		p := rline{c: cli}
		p.Run(ctx)
	})
}

// rline implements a quick-and-dirty prompt using readline.
type rline struct {
	c     CLI
	line  []rune
	pos   int
	start int
}

// Run launches the prompt until it is killed.
func (p *rline) Run(ctx context.Context) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "alice> ",
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
		p.c.Exec(ctx, line)
	}
}

// Do finds suggestions.
func (p *rline) Do(line []rune, pos int) (newLine [][]rune, length int) {
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
func (p *rline) TextBeforeCursor() string {
	return string(p.line[:p.pos])
}

// GetWordBeforeCursor return the word before the current cursor position.
func (p *rline) GetWordBeforeCursor() string {
	return string(p.line[p.start:p.pos])
}

// begin finds the start offset of the rurrent word.
func (p *rline) begin() {
	p.start = p.pos

	for p.start > 0 {
		if unicode.IsSpace(p.line[p.start-1]) {
			break
		}
		p.start--
	}
}
