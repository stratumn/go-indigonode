// Copyright © 2017-2018  Stratumn SAS
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

// SIGWINCH doesn't compile on Windows.
//+build !windows

package cli

import (
	"syscall"

	"github.com/pkg/errors"
)

// renderPrompt re-renders the prompt.
//
// This is necessary to re-render the prompt ("IndigoNode>"").
// See https://github.com/c-bata/go-prompt/issues/18.
func renderPrompt(sig Signaler) error {
	return errors.WithStack(sig.Signal(syscall.SIGWINCH))
}
