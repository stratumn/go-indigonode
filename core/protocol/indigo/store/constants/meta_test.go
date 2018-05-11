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

package constants_test

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stretchr/testify/assert"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

func TestNodeID(t *testing.T) {
	peerID, _ := peer.IDB58Decode("QmPEeCgxxX6YbQWqkKuF42YCUpy4GdrqGLPMAFZ8A3A35d")

	t.Run("missing-link", func(t *testing.T) {
		// Should not panic.
		constants.SetLinkNodeID(nil, peerID)

		_, err := constants.GetLinkNodeID(nil)
		assert.Error(t, err)
	})

	t.Run("missing-peer", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta.Data = nil

		constants.SetLinkNodeID(link, "")
		assert.Equal(t, "", link.Meta.Data[constants.NodeIDKey].(string))
	})

	t.Run("missing-node-id", func(t *testing.T) {
		link := cstesting.RandomLink()

		_, err := constants.GetLinkNodeID(link)
		assert.Error(t, err)
	})

	t.Run("invalid-node-id", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta.Data[constants.NodeIDKey] = "spongebob"

		_, err := constants.GetLinkNodeID(link)
		assert.Error(t, err)
	})

	t.Run("correctly-set-and-get", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta.Data[constants.NodeIDKey] = "spongebob"

		constants.SetLinkNodeID(link, peerID)
		id, err := constants.GetLinkNodeID(link)

		assert.NoError(t, err)
		assert.Equal(t, peerID, id)
	})
}

func TestValidatorHash(t *testing.T) {
	validatorHash := testutil.RandomHash()

	t.Run("missing-link", func(t *testing.T) {
		// Should not panic.
		constants.SetValidatorHash(nil, validatorHash)

		_, err := constants.GetValidatorHash(nil)
		assert.Error(t, err)
	})

	t.Run("missing-hash", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta.Data = nil

		constants.SetValidatorHash(link, nil)
		assert.Equal(t, nil, link.Meta.Data[constants.ValidatorHashKey])
	})

	t.Run("missing-validator-hash", func(t *testing.T) {
		link := cstesting.RandomLink()

		_, err := constants.GetValidatorHash(link)
		assert.EqualError(t, err, constants.ErrInvalidValidatorHash.Error())
	})

	t.Run("invalid-validator-hash", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta.Data[constants.ValidatorHashKey] = "spongebob"

		_, err := constants.GetValidatorHash(link)
		assert.EqualError(t, err, constants.ErrInvalidValidatorHash.Error())
	})

	t.Run("correctly-set-and-get", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta.Data[constants.ValidatorHashKey] = "spongebob"

		constants.SetValidatorHash(link, validatorHash)
		h, err := constants.GetValidatorHash(link)

		assert.NoError(t, err)
		assert.Equal(t, validatorHash, h)
	})
}
