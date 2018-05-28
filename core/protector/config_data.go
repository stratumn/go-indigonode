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

	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// ConfigData describes the data stored in config file.
type ConfigData struct {
	PeersAddrs map[string][]string `json:"peers_addresses"`
	Signature  []byte              `json:"signature"`
}

// NewConfigData creates a new ConfigData object.
func NewConfigData() *ConfigData {
	return &ConfigData{PeersAddrs: make(map[string][]string)}
}

// Load loads the given configuration and validates it.
func (c *ConfigData) Load(ctx context.Context, configPath string, privKey crypto.PrivKey) (err error) {
	event := log.EventBegin(ctx, "ConfigData.Load", logging.Metadata{"path": configPath})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

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

	c.PeersAddrs = confData.PeersAddrs
	c.Signature = confData.Signature

	return nil
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
