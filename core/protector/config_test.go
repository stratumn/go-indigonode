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

package protector_test

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stratumn/alice/core/protector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var (
	testKey crypto.PrivKey
)

func init() {
	testKey, _, _ = crypto.GenerateEd25519Key(rand.Reader)
}

func TestLocalSignedConfig_New(t *testing.T) {
	t.Run("creates-config-folder", func(t *testing.T) {
		configDir, _ := ioutil.TempDir(os.TempDir(), "alice")
		require.NoError(t, os.Remove(configDir))

		conf, err := protector.NewLocalSignedConfig(
			filepath.Join(configDir, "config.json"),
			testKey,
			nil,
		)
		require.NoError(t, err)
		assert.NotNil(t, conf)

		_, err = os.Stat(configDir)
		assert.NoError(t, err)
	})

	t.Run("loads-existing-config", func(t *testing.T) {
		configDir, _ := ioutil.TempDir(os.TempDir(), "alice")
		configPath := filepath.Join(configDir, "config.json")

		peerID := generatePeerID()
		configData := protector.ConfigData{
			PeersAddrs: map[string][]string{
				peerID.Pretty(): []string{"/ip4/127.0.0.1/tcp/8903"},
			},
		}

		require.NoError(t, configData.Flush(configPath, testKey))

		p := protector.NewPrivateNetwork(nil)
		conf, err := protector.NewLocalSignedConfig(configPath, testKey, p)
		require.NoError(t, err)
		assert.NotNil(t, conf)

		waitUntilAllowed(t, p, peerID, 1)
		allowed := conf.AllowedPeers()
		assert.ElementsMatch(t, []peer.ID{peerID}, allowed)
	})

	t.Run("fails-invalid-config", func(t *testing.T) {
		configDir, _ := ioutil.TempDir(os.TempDir(), "alice")
		configPath := filepath.Join(configDir, "invalid.json")
		err := ioutil.WriteFile(
			configPath,
			[]byte("not a json config"),
			0644,
		)
		require.NoError(t, err)

		_, err = protector.NewLocalSignedConfig(configPath, testKey, nil)
		assert.EqualError(t, err, protector.ErrInvalidConfig.Error())
	})

	t.Run("fails-invalid-config-signature", func(t *testing.T) {
		configDir, _ := ioutil.TempDir(os.TempDir(), "alice")
		configPath := filepath.Join(configDir, "invalid_sig.json")

		otherKey, _, _ := crypto.GenerateEd25519Key(rand.Reader)
		configData := protector.ConfigData{
			PeersAddrs: map[string][]string{
				generatePeerID().Pretty(): []string{"/ip4/127.0.0.1/tcp/8903"},
			},
		}

		require.NoError(t, configData.Flush(configPath, otherKey))

		_, err := protector.NewLocalSignedConfig(configPath, testKey, nil)
		assert.EqualError(t, err, protector.ErrInvalidSignature.Error())
	})
}
