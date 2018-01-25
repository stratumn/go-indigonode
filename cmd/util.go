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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/core"
	"github.com/stratumn/alice/core/manager"
)

// fail prints an error and exits if the error is not nil.
func fail(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
}

// requireCoreConfigSet loads the core's configuration file and exits on failure.
func requireCoreConfigSet() core.ConfigurableSet {
	set := core.NewConfigurableSet(services)

	if err := core.LoadConfig(set, coreCfgFilename); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the core configuration file %q: %s.\n", coreCfgFilename, err)

		if os.IsNotExist(errors.Cause(err)) {
			fmt.Fprintln(os.Stderr, "You can create one using `alice init`.")
		}

		os.Exit(1)
	}

	return set
}

// requireCLIConfig loads the CLI's configuration file and exits on failure.
func requireCLIConfigSet() cli.ConfigurableSet {
	set := cli.NewConfigurableSet()

	if err := cli.LoadConfig(set, cliCfgFilename); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the command line interface configuration file %q: %s.\n", cliCfgFilename, err)

		if os.IsNotExist(errors.Cause(err)) {
			fmt.Fprintln(os.Stderr, "You can create one using `alice init`.")
		}

		os.Exit(1)
	}

	return set
}

// up launches a node.
func up(mgr *manager.Manager) {
	config := requireCoreConfigSet().Configs()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{}, 1)

	start := func() {
		c, err := core.New(
			config,
			core.OptServices(services...),
			core.OptManager(mgr),
		)
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
}
