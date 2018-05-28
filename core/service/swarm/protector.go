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

package swarm

import (
	"context"

	"github.com/stratumn/alice/core/protector"

	"gx/ipfs/Qmd3oYWVLCVWryDV6Pobv6whZcvDXAHqS3chemZ658y4a8/go-libp2p-interface-pnet"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

// ProtectorConfig configures the right protector for the swarm service.
type ProtectorConfig interface {
	Configure(context.Context, *Service, peerstore.Peerstore) (ipnet.Protector, error)
}

// NewProtectorConfig returns the right protector configuration
// depending on the network parameters.
func NewProtectorConfig(config *Config) (ProtectorConfig, error) {
	switch config.ProtectionMode {
	case "":
		return &noProtectorConfig{}, nil
	case PrivateWithCoordinatorMode:
		if config.CoordinatorConfig.IsCoordinator {
			return &coordinatorConfig{}, nil
		}

		return &withCoordinatorConfig{}, nil
	default:
		return nil, ErrInvalidProtectionMode
	}
}

// noProtectorConfig is used when no protection is needed for the network.
type noProtectorConfig struct{}

func (c *noProtectorConfig) Configure(_ context.Context, _ *Service, _ peerstore.Peerstore) (ipnet.Protector, error) {
	return nil, nil
}

// coordinatorConfig is used for the coordinator of a private network.
type coordinatorConfig struct{}

func (c *coordinatorConfig) Configure(ctx context.Context, s *Service, pstore peerstore.Peerstore) (ipnet.Protector, error) {
	// TODO: refactor the Protector/NetworkConfig interfaces
	// The NetworkConfig should expose the state of the network (bootstrapped/bootstrapping)
	// and the protector should be able to access that state.
	p, _ := protector.NewPrivateNetworkWithBootstrap(pstore)

	var err error
	s.networkConfig, err = protector.InitLocalConfig(ctx, s.config.CoordinatorConfig.ConfigPath, s.privKey, p, pstore)
	if err != nil {
		return nil, err
	}

	if err = s.networkConfig.AddPeer(ctx, s.peerID, s.addrs); err != nil {
		return nil, err
	}

	return p, nil
}

// withCoordinatorConfig is used for a node in a private network
// that uses a coordinator.
type withCoordinatorConfig struct{}

func (c *withCoordinatorConfig) Configure(ctx context.Context, s *Service, pstore peerstore.Peerstore) (ipnet.Protector, error) {
	p := protector.NewPrivateNetwork(pstore)

	var err error
	s.networkConfig, err = protector.InitLocalConfig(ctx, s.config.CoordinatorConfig.ConfigPath, s.privKey, p, pstore)
	if err != nil {
		return nil, err
	}

	pstore.AddAddrs(s.coordinatorID, s.coordinatorAddrs, peerstore.PermanentAddrTTL)

	if err = s.networkConfig.AddPeer(ctx, s.coordinatorID, s.coordinatorAddrs); err != nil {
		return nil, err
	}

	return p, nil
}
