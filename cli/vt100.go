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

// vt100 doesn't compile on Windows.
//+build !windows

package cli

import (
	"context"

	prompt "github.com/c-bata/go-prompt"
)

// Register the prompt.
func init() {
	registerPrompt("vt100", vt100Run)
}

// vt100Run launches the prompt until it is killed.
func vt100Run(ctx context.Context, c CLI) {
	prompt.New(
		func(in string) {
			c.Run(ctx, in)
		},
		func(d prompt.Document) []prompt.Suggest {
			// Don't show suggestions right after a command was
			// executed.
			if c.DidJustExecute() {
				return nil
			}

			csugs := c.Suggest(&d)

			// Convert cli suggestions to go-prompt suggestions.
			psugs := make([]prompt.Suggest, len(csugs))

			for i, v := range csugs {
				psugs[i] = prompt.Suggest{
					Text:        v.Text,
					Description: v.Desc,
				}
			}

			return psugs
		},
		prompt.OptionTitle("Stratumn Node CLI"),
		prompt.OptionPrefix("node> "),
	).Run()
}
