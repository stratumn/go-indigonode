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
		prompt.OptionTitle("Alice CLI"),
		prompt.OptionPrefix("Alice> "),
	).Run()
}
