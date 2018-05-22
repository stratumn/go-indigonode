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
	"io/ioutil"
	"os"
	"path/filepath"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

const (
	// DefaultConfigPath is the location of the config file.
	DefaultConfigPath = "data/network/config.json"
)

var (
	// ErrInvalidConfig is returned when an existing configuration file
	// exists and is invalid.
	ErrInvalidConfig = errors.New("invalid configuration file")

	// ErrInvalidSignature is returned when an existing configuration file
	// contains an invalid signature.
	ErrInvalidSignature = errors.New("invalid configuration signature")
)

// Config manages the network participants list.
type Config interface {
	AddPeer(context.Context, peer.ID) error
	RemovePeer(context.Context, peer.ID) error
	AllowedPeers() []peer.ID
}

// ConfigData describes the data stored in config file.
type ConfigData struct {
	PeersAddrs map[string][]string `json:"peers_addresses"`
	Signature  []byte              `json:"signature"`
}

// Flush signs the configuration data and writes it to disk.
func (c *ConfigData) Flush(ctx context.Context, configPath string, privKey crypto.PrivKey) (err error) {
	event := log.EventBegin(ctx, "ConfigData.Flush", logging.Metadata{"path": configPath})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	b, err := json.Marshal(c.PeersAddrs)
	if err != nil {
		return errors.WithStack(err)
	}

	c.Signature, err = privKey.Sign(b)
	if err != nil {
		return errors.WithStack(err)
	}

	signedBytes, err := json.Marshal(c)
	if err != nil {
		return errors.WithStack(err)
	}

	return ioutil.WriteFile(configPath, signedBytes, 0644)
}

// LocalConfig implements the Config interface.
// It keeps a signed config file on the filesystem.
type LocalConfig struct {
	privKey     crypto.PrivKey
	protect     Protector
	protectChan chan NetworkUpdate
}

// InitLocalConfig loads a Config from the given file or creates it if missing.
// It configures the given protector for automatic updates.
func InitLocalConfig(
	ctx context.Context,
	configPath string,
	privKey crypto.PrivKey,
	protect Protector,
	peerStore peerstore.Peerstore,
) (c Config, err error) {
	event := log.EventBegin(ctx, "InitLocalConfig", logging.Metadata{"path": configPath})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	// Create the directory if it doesn't exist.
	configDir, _ := filepath.Split(configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return nil, errors.WithStack(err)
	}

	conf := &LocalConfig{
		privKey:     privKey,
		protect:     protect,
		protectChan: make(chan NetworkUpdate),
	}

	// This go routine has the same lifetime as the Config object,
	// so it makes sense to launch it here. When the Config object
	// is collected by the GC, the channel is closed which stops
	// this go routine.
	go protect.ListenForUpdates(conf.protectChan)

	_, err = os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, ErrInvalidConfig
	}

	// Load previous configuration.
	if err == nil {
		previousData, err := conf.load(configPath, privKey)
		if err != nil {
			return nil, err
		}

		for peerID, peerAddrs := range previousData.PeersAddrs {
			decodedPeerID, err := peer.IDB58Decode(peerID)
			if err != nil {
				return nil, ErrInvalidConfig
			}

			for _, peerAddr := range peerAddrs {
				decodedPeerAddr, err := multiaddr.NewMultiaddr(peerAddr)
				if err != nil {
					return nil, ErrInvalidConfig
				}

				peerStore.AddAddr(decodedPeerID, decodedPeerAddr, peerstore.PermanentAddrTTL)
			}

			conf.protectChan <- CreateAddNetworkUpdate(decodedPeerID)
		}
	}

	return conf, nil
}

func (c *LocalConfig) load(configPath string, privKey crypto.PrivKey) (*ConfigData, error) {
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, ErrInvalidConfig
	}

	var confData ConfigData
	err = json.Unmarshal(configBytes, &confData)
	if err != nil {
		return nil, ErrInvalidConfig
	}

	confBytes, err := json.Marshal(confData.PeersAddrs)
	if err != nil {
		return nil, ErrInvalidConfig
	}

	valid, err := privKey.GetPublic().Verify(confBytes, confData.Signature)
	if !valid || err != nil {
		return nil, ErrInvalidSignature
	}

	return &confData, nil
}

func (c *LocalConfig) AddPeer(context.Context, peer.ID) error {
	return nil
}

func (c *LocalConfig) RemovePeer(context.Context, peer.ID) error {
	return nil
}

// AllowedPeers returns the IDs of the peers in the network.
func (c *LocalConfig) AllowedPeers() []peer.ID {
	return c.protect.AllowedPeers()
}
