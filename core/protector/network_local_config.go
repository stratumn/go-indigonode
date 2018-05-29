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
	"sync"

	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// LocalConfig implements the NetworkConfig interface.
// It keeps a signed config file on the filesystem.
type LocalConfig struct {
	dataLock sync.RWMutex
	data     *ConfigData
	dataPath string

	peerStore   peerstore.Peerstore
	privKey     crypto.PrivKey
	protect     Protector
	protectChan chan NetworkUpdate
}

// InitLocalConfig loads a NetworkConfig from the given file or creates it if missing.
// It configures the given protector for automatic updates.
func InitLocalConfig(
	ctx context.Context,
	configPath string,
	privKey crypto.PrivKey,
	protect Protector,
	peerStore peerstore.Peerstore,
) (c NetworkConfig, err error) {
	event := log.EventBegin(ctx, "InitLocalConfig", logging.Metadata{"path": configPath})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	// Create the directory if it doesn't exist.
	configDir, _ := filepath.Split(configPath)
	if err = os.MkdirAll(configDir, 0744); err != nil {
		return nil, errors.WithStack(err)
	}

	conf := &LocalConfig{
		data:        NewConfigData(),
		dataPath:    configPath,
		peerStore:   peerStore,
		privKey:     privKey,
		protect:     protect,
		protectChan: make(chan NetworkUpdate),
	}

	// This go routine has the same lifetime as the NetworkConfig object,
	// so it makes sense to launch it here. When the NetworkConfig object
	// is collected by the GC, the channel is closed which stops
	// this go routine.
	go protect.ListenForUpdates(conf.protectChan)

	_, err = os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, ErrInvalidConfig
	}

	// Load previous configuration.
	if err == nil {
		err = conf.data.Load(ctx, configPath, privKey)
		if err != nil {
			return nil, err
		}

		if err = conf.addDataToPeerStore(); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

// addToPeerStore adds peers' addresses to the peer store.
func (c *LocalConfig) addDataToPeerStore() error {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	for peerID, peerAddrs := range c.data.PeersAddrs {
		decodedPeerID, err := peer.IDB58Decode(peerID)
		if err != nil {
			return ErrInvalidConfig
		}

		for _, peerAddr := range peerAddrs {
			decodedPeerAddr, err := multiaddr.NewMultiaddr(peerAddr)
			if err != nil {
				return ErrInvalidConfig
			}

			c.peerStore.AddAddr(decodedPeerID, decodedPeerAddr, peerstore.PermanentAddrTTL)
		}

		c.protectChan <- CreateAddNetworkUpdate(decodedPeerID)
	}

	return nil
}

// AddPeer adds a peer to the network configuration.
// It populates the peer store with the peer's initial addresses.
func (c *LocalConfig) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	defer log.EventBegin(ctx, "LocalConfig.AddPeer", logging.Metadata{
		"peer": peerID.Pretty(),
	}).Done()

	c.peerStore.AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)
	c.protectChan <- CreateAddNetworkUpdate(peerID)

	var marshalledAddrs []string
	for _, addr := range addrs {
		marshalledAddrs = append(marshalledAddrs, addr.String())
	}

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.data.PeersAddrs[peerID.Pretty()] = marshalledAddrs
	return c.data.Flush(ctx, c.dataPath, c.privKey)
}

// RemovePeer removes a peer from the network configuration.
func (c *LocalConfig) RemovePeer(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "LocalConfig.RemovePeer", logging.Metadata{
		"peer": peerID.Pretty(),
	}).Done()

	c.protectChan <- CreateRemoveNetworkUpdate(peerID)

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	delete(c.data.PeersAddrs, peerID.Pretty())
	return c.data.Flush(ctx, c.dataPath, c.privKey)
}

// AllowedPeers returns the IDs of the peers in the network.
func (c *LocalConfig) AllowedPeers(ctx context.Context) []peer.ID {
	return c.protect.AllowedPeers(ctx)
}

// NetworkState returns the current state of the network protection.
func (c *LocalConfig) NetworkState(ctx context.Context) NetworkState {
	event := log.EventBegin(ctx, "LocalConfig.NetworkState")
	defer event.Done()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	event.Append(logging.Metadata{"state": c.data.NetworkState})
	return c.data.NetworkState
}

// SetNetworkState sets the current state of the network protection.
func (c *LocalConfig) SetNetworkState(ctx context.Context, networkState NetworkState) error {
	defer log.EventBegin(ctx, "LocalConfig.SetNetworkState", logging.Metadata{
		"state": networkState,
	}).Done()

	switch networkState {
	case Bootstrap, Protected:
		c.dataLock.Lock()
		defer c.dataLock.Unlock()

		c.data.NetworkState = networkState

		stateAwareProtector, ok := c.protect.(NetworkStateWriter)
		if ok {
			err := stateAwareProtector.SetNetworkState(ctx, networkState)
			if err != nil {
				return err
			}
		}

		return c.data.Flush(ctx, c.dataPath, c.privKey)
	default:
		return ErrInvalidNetworkState
	}
}
