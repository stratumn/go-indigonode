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

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stratumn/go-indigonode/core"
)

// upCmd represents the up command.
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start a node",
	Run: func(cmd *cobra.Command, args []string) {
		config := requireCoreConfigSet().Configs()

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{}, 1)

		start := func() {
			c, err := core.New(config, core.OptServices(services...))
			fail(err)

			err = c.Boot(ctx)
			if errors.Cause(err) != context.Canceled {
				fail(err)
			}

			done <- struct{}{}
		}

		go start()

		// Reload configuration and restart on SIGHUP signal.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)

		for range sig {
			cancel()
			<-done
			ctx, cancel = context.WithCancel(context.Background())
			config = requireCoreConfigSet().Configs()
			go start()
		}

		// So the linter doesn't complain.
		cancel()
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
}
