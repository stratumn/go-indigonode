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

// SIGWINCH doesn't compile on Windows.
//+build !windows

package cli

import (
	"syscall"

	"github.com/pkg/errors"
)

// renderPrompt re-renders the prompt.
//
// This is necessary to re-render the prompt ("node>"").
// See https://github.com/c-bata/go-prompt/issues/18.
func renderPrompt(sig Signaler) error {
	return errors.WithStack(sig.Signal(syscall.SIGWINCH))
}
