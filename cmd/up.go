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
