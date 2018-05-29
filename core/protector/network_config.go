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

package protector

import (
	"context"

	"github.com/pkg/errors"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

const (
	// DefaultConfigPath is the location of the config file.
	DefaultConfigPath = "data/network/config.json"
)

var (
	// ErrInvalidConfig is returned when an invalid configuration file
	// is provided.
	ErrInvalidConfig = errors.New("invalid configuration file")

	// ErrInvalidSignature is returned when an existing configuration file
	// contains an invalid signature.
	ErrInvalidSignature = errors.New("invalid configuration signature")

	// ErrInvalidNetworkState is returned when trying to set an
	// invalid network state.
	ErrInvalidNetworkState = errors.New("invalid network state")
)

// NetworkState is the state of a private network.
type NetworkState string

// NetworkState values.
const (
	// The network is in the bootstrap phase (not fully private yet).
	NetworkStateBootstrap = NetworkState("bootstrap")
	// The network bootstrap phase is complete and the network is now protected.
	NetworkStateProtected = NetworkState("protected")
)

// NetworkStateReader provides read access to the network state.
type NetworkStateReader interface {
	NetworkState(context.Context) NetworkState
}

// NetworkStateWriter provides write access to the network state.
type NetworkStateWriter interface {
	SetNetworkState(context.Context, NetworkState) error
}

// NetworkPeersReader provides read access to the network peers list.
type NetworkPeersReader interface {
	AllowedPeers(context.Context) []peer.ID
}

// NetworkPeersWriter provides write access to the network peers list.
type NetworkPeersWriter interface {
	AddPeer(context.Context, peer.ID, []multiaddr.Multiaddr) error
	RemovePeer(context.Context, peer.ID) error
}

// NetworkConfig manages the private network's configuration.
type NetworkConfig interface {
	NetworkPeersReader
	NetworkPeersWriter

	NetworkStateReader
	NetworkStateWriter
}
