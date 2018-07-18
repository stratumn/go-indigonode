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

package pb

import (
	"context"
	"io/ioutil"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	cryptopb "github.com/stratumn/go-indigonode/core/crypto"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var log = logging.Logger("bootstrap")

// Errors used by the NetworkConfig.
var (
	ErrInvalidConfig       = errors.New("invalid configuration file")
	ErrInvalidSignature    = errors.New("invalid configuration signature")
	ErrInvalidNetworkState = errors.New("invalid network state")
	ErrInvalidPeerID       = errors.New("invalid peer ID")
	ErrMissingPeerAddrs    = errors.New("missing peer addresses")
	ErrInvalidPeerAddr     = errors.New("invalid peer address")
	ErrMissingLastUpdated  = errors.New("missing last updated timestamp")
	ErrInvalidLastUpdated  = errors.New("invalid last updated (maybe outdated)")
)

// NewNetworkConfig creates a NetworkConfig.
func NewNetworkConfig(networkState NetworkState) *NetworkConfig {
	networkConfig := &NetworkConfig{
		NetworkState: networkState,
		Participants: make(map[string]*PeerAddrs),
	}

	return networkConfig
}

// Sign signs the current configuration data.
func (c *NetworkConfig) Sign(ctx context.Context, privKey crypto.PrivKey) error {
	event := log.EventBegin(ctx, "NetworkConfig.Sign")
	defer event.Done()

	c.LastUpdated = types.TimestampNow()
	c.Signature = nil

	b, err := json.Marshal(c)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	c.Signature, err = cryptopb.Sign(ctx, privKey, b)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	return nil
}

// ValidateSignature verifies that the network configuration
// was signed by a given peer.
func (c *NetworkConfig) ValidateSignature(ctx context.Context, peerID peer.ID) bool {
	if c == nil {
		return false
	}

	if c.Signature == nil {
		return false
	}

	// A signed configuration should always have an associated timestamp.
	if c.LastUpdated == nil {
		return false
	}

	pk, err := crypto.UnmarshalPublicKey(c.Signature.PublicKey)
	if err != nil {
		return false
	}

	if !peerID.MatchesPublicKey(pk) {
		return false
	}

	signature := c.Signature
	c.Signature = nil

	payload, err := json.Marshal(c)
	c.Signature = signature
	if err != nil {
		return false
	}

	return c.Signature.Verify(ctx, payload)
}

// ValidateContent validates that the contents of the configuration
// is correctly typed (peerIDs, addresses, etc).
func (c *NetworkConfig) ValidateContent(ctx context.Context) error {
	_, ok := NetworkState_name[int32(c.NetworkState)]
	if !ok {
		return ErrInvalidNetworkState
	}

	for peerID, peerAddrs := range c.Participants {
		_, err := peer.IDB58Decode(peerID)
		if err != nil {
			return ErrInvalidPeerID
		}

		if peerAddrs == nil || len(peerAddrs.Addresses) == 0 {
			return ErrMissingPeerAddrs
		}

		for _, peerAddr := range peerAddrs.Addresses {
			_, err := multiaddr.NewMultiaddr(peerAddr)
			if err != nil {
				return ErrInvalidPeerAddr
			}
		}
	}

	return nil
}

// LoadFromFile loads the given configuration and validates it.
func (c *NetworkConfig) LoadFromFile(ctx context.Context, path string, signerID peer.ID) (err error) {
	event := log.EventBegin(ctx, "NetworkConfig.Load", logging.Metadata{"path": path})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return ErrInvalidConfig
	}

	var confData NetworkConfig
	err = json.Unmarshal(configBytes, &confData)
	if err != nil {
		return ErrInvalidConfig
	}

	err = confData.ValidateContent(ctx)
	if err != nil {
		return err
	}

	if !confData.ValidateSignature(ctx, signerID) {
		return ErrInvalidSignature
	}

	c.NetworkState = confData.NetworkState
	c.LastUpdated = confData.LastUpdated
	c.Participants = confData.Participants
	c.Signature = confData.Signature

	return nil
}
