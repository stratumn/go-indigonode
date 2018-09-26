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

package protector

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/monitoring"
	"github.com/stratumn/go-node/core/protector/pb"

	"gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
)

const (
	// DefaultConfigPath is the location of the config file.
	DefaultConfigPath = "data/network/config.json"
)

// Errors used by the network configuration.
var (
	ErrMissingNonLocalAddr = errors.New("need a non-local peer address")
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
	AllowedAddrs(context.Context, peer.ID) []multiaddr.Multiaddr
}

// NetworkPeersWriter provides write access to the network peers list.
type NetworkPeersWriter interface {
	AddPeer(context.Context, peer.ID, []multiaddr.Multiaddr) error
	RemovePeer(context.Context, peer.ID) error
}

// NetworkConfigReader provides read access to the network configuration.
type NetworkConfigReader interface {
	NetworkStateReader
	NetworkPeersReader
}

// NetworkConfigWriter provides write access to the network configuration.
type NetworkConfigWriter interface {
	NetworkStateWriter
	NetworkPeersWriter
}

// NetworkConfig manages the private network's configuration.
type NetworkConfig interface {
	NetworkPeersReader
	NetworkPeersWriter

	NetworkStateReader
	NetworkStateWriter

	// Sign signs the underlying configuration.
	Sign(context.Context, crypto.PrivKey) error

	// Copy returns a copy of the underlying configuration.
	Copy(context.Context) pb.NetworkConfig

	// Reset clears the current configuration and applies the given one.
	// It assumes that the incoming configuration signature has been validated.
	Reset(context.Context, *pb.NetworkConfig) error
}

// LoadOrInitNetworkConfig loads a NetworkConfig from the given file
// or creates it if missing.
func LoadOrInitNetworkConfig(
	ctx context.Context,
	configPath string,
	privKey crypto.PrivKey,
	protect Protector,
	peerStore peerstore.Peerstore,
) (NetworkConfig, error) {
	ctx, span := monitoring.StartSpan(ctx, "protector", "LoadOrInitNetworkConfig")
	span.AddStringAttribute("path", configPath)
	defer span.End()

	conf, err := NewInMemoryConfig(ctx, pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP))
	if err != nil {
		span.SetUnknownError(err)
		return nil, err
	}

	// Create the directory if it doesn't exist.
	configDir, _ := filepath.Split(configPath)
	if err := os.MkdirAll(configDir, 0744); err != nil {
		span.SetUnknownError(err)
		return nil, errors.WithStack(err)
	}

	_, err = os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		span.SetUnknownError(err)
		return nil, pb.ErrInvalidConfig
	}

	wrappedConf := WrapWithSaver(
		WrapWithProtectUpdater(
			WrapWithSignature(conf, privKey),
			protect,
			peerStore,
		),
		configPath,
	)

	// Load previous configuration.
	if err == nil {
		peerID, err := peer.IDFromPrivateKey(privKey)
		if err != nil {
			span.SetUnknownError(err)
			return nil, errors.WithStack(err)
		}

		previousConf := &pb.NetworkConfig{}
		err = previousConf.LoadFromFile(ctx, configPath, peerID)
		if err != nil {
			span.SetUnknownError(err)
			return nil, err
		}

		err = wrappedConf.SetNetworkState(ctx, previousConf.NetworkState)
		if err != nil {
			span.SetUnknownError(err)
			return nil, err
		}

		// previousConf.LoadFromFile has already validated that the
		// configuration data is ok, so we're ignoring parsing errors.
		for peerID, peerAddrs := range previousConf.Participants {
			decodedPeerID, _ := peer.IDB58Decode(peerID)
			var peerMultiAddrs []multiaddr.Multiaddr
			for _, peerAddr := range peerAddrs.Addresses {
				decodedPeerAddr, _ := multiaddr.NewMultiaddr(peerAddr)
				peerMultiAddrs = append(peerMultiAddrs, decodedPeerAddr)
			}

			err = wrappedConf.AddPeer(ctx, decodedPeerID, peerMultiAddrs)
			if err != nil {
				span.SetUnknownError(err)
				return nil, err
			}
		}
	}

	return wrappedConf, nil
}
