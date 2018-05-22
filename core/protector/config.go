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

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
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
func (c *ConfigData) Flush(configPath string, privKey crypto.PrivKey) error {
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

	return ioutil.WriteFile(configPath, signedBytes, os.ModePerm)
}

// LocalSignedConfig implements the Config interface.
// It keeps a signed config file on the filesystem.
type LocalSignedConfig struct {
	privKey crypto.PrivKey
	protect Protector
}

// NewLocalSignedConfig creates a Config instance.
// It loads the previous configuration if it exists.
func NewLocalSignedConfig(configPath string, privKey crypto.PrivKey, protect Protector) (Config, error) {
	_, err := os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, ErrInvalidConfig
	}

	conf := &LocalSignedConfig{
		privKey: privKey,
		protect: protect,
	}

	if err == nil {
		if err = conf.load(configPath, privKey); err != nil {
			return nil, err
		}
	} else {
		dir, _ := filepath.Split(configPath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return conf, nil
}

func (c *LocalSignedConfig) load(configPath string, privKey crypto.PrivKey) error {
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ErrInvalidConfig
	}

	var confData ConfigData
	err = json.Unmarshal(configBytes, &confData)
	if err != nil {
		return ErrInvalidConfig
	}

	confBytes, err := json.Marshal(confData.PeersAddrs)
	if err != nil {
		return ErrInvalidConfig
	}

	valid, err := privKey.GetPublic().Verify(confBytes, confData.Signature)
	if !valid || err != nil {
		return ErrInvalidSignature
	}

	return nil
}

func (c *LocalSignedConfig) AddPeer(context.Context, peer.ID) error {
	return nil
}

func (c *LocalSignedConfig) RemovePeer(context.Context, peer.ID) error {
	return nil
}

func (c *LocalSignedConfig) AllowedPeers() []peer.ID {
	return nil
}
