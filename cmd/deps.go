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
