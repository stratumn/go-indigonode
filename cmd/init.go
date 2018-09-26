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

	"github.com/spf13/cobra"
	"github.com/stratumn/go-node/cli"
	"github.com/stratumn/go-node/core"
	bootstrap "github.com/stratumn/go-node/core/app/bootstrap/service"
	swarm "github.com/stratumn/go-node/core/app/swarm/service"
	"github.com/stratumn/go-node/core/cfg"
	"github.com/stratumn/go-node/core/protector"

	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
)

// Flags used by the init command.
const (
	PrivateCoordinatorFlagName     = "private-coordinator"
	PrivateWithCoordinatorFlagName = "private-with-coordinator"
	CoordinatorAddrFlagName        = "coordinator-addr"
)

var (
	initPrivateCoordinator     bool
	initPrivateWithCoordinator bool
	initCoordinatorAddr        string
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		validateFlags()

		configSet := core.NewConfigurableSet(services)

		if initPrivateCoordinator {
			configurePrivateCoordinator(configSet)
		} else if initPrivateWithCoordinator {
			configurePrivateWithCoordinatorMode(configSet)
		}
		if err := configSet.Save(coreCfgFilename(), 0600, cfg.ConfigSaveOpts{
			Overwrite: false,
			Backup:    false,
		}); err != nil {
			osExit(1, fmt.Sprintf("Could not save the core configuration file: %s.", err))
		}

		fmt.Printf("Created configuration file %q.\n", coreCfgFilename())
		fmt.Println("Keep this file private!!!")

		if err := cli.NewConfigurableSet().Save(cliCfgFilename(), 0600, cfg.ConfigSaveOpts{
			Overwrite: false,
			Backup:    false,
		}); err != nil {
			osExit(1, fmt.Sprintf("Could not save the command line interface configuration file: %s.", err))
		}

		fmt.Printf("Created configuration file %q.\n", cliCfgFilename())
		fmt.Println("Keep this file private!!!")
	},
}

func validateFlags() {
	if initPrivateCoordinator && initPrivateWithCoordinator {
		osExit(1, fmt.Sprintf("Invalid flag combination. "+
			"You need to chose between private-coordinator "+
			"and private-with-coordinator"))
	}
}

func getSwarmConfig(configSet cfg.Set) swarm.Config {
	swarmConfig, ok := configSet[swarm.ServiceID].Config().(swarm.Config)
	if !ok {
		osExit(1, fmt.Sprintf("Invalid or missing swarm configuration."))
	}

	return swarmConfig
}

func setSwarmConfig(configSet cfg.Set, swarmConfig swarm.Config) {
	if err := configSet[swarm.ServiceID].SetConfig(swarmConfig); err != nil {
		osExit(1, fmt.Sprintf("Could not set swarm configuration: %s.", err))
	}
}

func getBootstrapConfig(configSet cfg.Set) bootstrap.Config {
	bootstrapConfig, ok := configSet[bootstrap.ServiceID].Config().(bootstrap.Config)
	if !ok {
		osExit(1, fmt.Sprintf("Invalid or missing bootstrap configuration."))
	}

	return bootstrapConfig
}

func setBootstrapConfig(configSet cfg.Set, bootstrapConfig bootstrap.Config) {
	if err := configSet[bootstrap.ServiceID].SetConfig(bootstrapConfig); err != nil {
		osExit(1, fmt.Sprintf("Could not set bootstrap configuration: %s.", err))
	}
}

func configurePrivateCoordinator(configSet cfg.Set) {
	swarmConfig := getSwarmConfig(configSet)
	bootstrapConfig := getBootstrapConfig(configSet)

	swarmConfig.ProtectionMode = protector.PrivateWithCoordinatorMode
	swarmConfig.CoordinatorConfig = &swarm.CoordinatorConfig{
		ConfigPath:    protector.DefaultConfigPath,
		IsCoordinator: true,
	}

	// No bootstrapping to other nodes, we are the bootstrapping node.
	bootstrapConfig.Addresses = nil
	bootstrapConfig.MinPeerThreshold = 0

	setSwarmConfig(configSet, swarmConfig)
	setBootstrapConfig(configSet, bootstrapConfig)
}

func configurePrivateWithCoordinatorMode(configSet cfg.Set) {
	coordinatorAddr, err := multiaddr.NewMultiaddr(initCoordinatorAddr)
	if err != nil {
		osExit(1, fmt.Sprintf("Invalid coordinator address (%s): %s.", initCoordinatorAddr, err))
	}

	coordinatorInfo, err := peerstore.InfoFromP2pAddr(coordinatorAddr)
	if err != nil {
		osExit(1, fmt.Sprintf("Invalid coordinator address (%s): %s.", initCoordinatorAddr, err))
	}

	swarmConfig := getSwarmConfig(configSet)
	bootstrapConfig := getBootstrapConfig(configSet)

	swarmConfig.ProtectionMode = protector.PrivateWithCoordinatorMode
	swarmConfig.CoordinatorConfig = &swarm.CoordinatorConfig{
		ConfigPath:    protector.DefaultConfigPath,
		CoordinatorID: coordinatorInfo.ID.Pretty(),
	}

	for _, addr := range coordinatorInfo.Addrs {
		swarmConfig.CoordinatorConfig.CoordinatorAddresses = append(swarmConfig.CoordinatorConfig.CoordinatorAddresses, addr.String())
	}

	// Bootstrap only with the coordinator node.
	bootstrapConfig.Addresses = []string{coordinatorAddr.String()}
	bootstrapConfig.MinPeerThreshold = 1

	setSwarmConfig(configSet, swarmConfig)
	setBootstrapConfig(configSet, bootstrapConfig)
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(
		&initPrivateCoordinator,
		PrivateCoordinatorFlagName,
		false,
		"initialize the coordinator of a private network",
	)

	initCmd.Flags().BoolVar(
		&initPrivateWithCoordinator,
		PrivateWithCoordinatorFlagName,
		false,
		"initialize a private network using a coordinator node",
	)

	initCmd.Flags().StringVar(
		&initCoordinatorAddr,
		CoordinatorAddrFlagName,
		"",
		"multiaddr of the coordinator node",
	)
}
