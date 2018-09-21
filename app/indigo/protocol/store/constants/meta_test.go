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

package constants_test

import (
	"testing"

	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stretchr/testify/assert"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
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
