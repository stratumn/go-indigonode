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

	"github.com/spf13/cobra"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/core"
	bootstrap "github.com/stratumn/alice/core/app/bootstrap/service"
	swarm "github.com/stratumn/alice/core/app/swarm/service"
	"github.com/stratumn/alice/core/cfg"
	"github.com/stratumn/alice/core/protector"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
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
