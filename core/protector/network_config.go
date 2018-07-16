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
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector/pb"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
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
