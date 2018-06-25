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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stratumn/go-indigonode/core"
)

// depsService is the ID of the service of which to show dependencies.
var depsService string

// depsGraph is whether to show the dependency graph.
var depsGraph bool

// depsCmd represents the deps command.
var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Display the dependencies of services",
	Run: func(cmd *cobra.Command, args []string) {
		config := requireCoreConfigSet().Configs()

		if depsGraph {
			fail(core.Fgraph(os.Stdout, services, config, depsService))
			return
		}

		deps, err := core.Deps(services, config, depsService)
		fail(err)

		fmt.Println(strings.Join(deps, "\n"))
	},
}

func init() {
	RootCmd.AddCommand(depsCmd)

	depsCmd.Flags().StringVar(
		&depsService,
		"service",
		"",
		"the service of which to show dependencies (defaults to boot service)",
	)

	depsCmd.Flags().BoolVar(
		&depsGraph,
		"graph",
		false,
		"whether to show the dependency graph",
	)
}
