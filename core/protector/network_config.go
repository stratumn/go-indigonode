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

	pb "github.com/stratumn/alice/pb/protector"

	"gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

const (
	// DefaultConfigPath is the location of the config file.
	DefaultConfigPath = "data/network/config.json"
)

// NetworkStateReader provides read access to the network state.
type NetworkStateReader interface {
	NetworkState(context.Context) pb.NetworkState
}

// NetworkStateWriter provides write access to the network state.
type NetworkStateWriter interface {
	SetNetworkState(context.Context, pb.NetworkState) error
}

// NetworkPeersReader provides read access to the network peers list.
type NetworkPeersReader interface {
	IsAllowed(context.Context, peer.ID) bool
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

	// Encode encodes the configuration with the given encoder.
	Encode(multicodec.Encoder) error
}
