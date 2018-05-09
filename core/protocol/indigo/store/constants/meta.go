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

package constants

import (
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// Keys used when enriching links metadata.
const (
	NodeIDKey = "node_id"

	ValidatorHashKey = "validator_hash"
)

// Errors returned when invalid metadata is received.
var (
	// ErrInvalidMetaNodeID is returned when the nodeID in the link
	// meta can't be properly verified.
	ErrInvalidMetaNodeID = errors.New("missing or invalid nodeID in metadata")

	// ErrInvalidValidatorHash is returned when the validator hash in the link
	// meta can't be retrieved.
	ErrInvalidValidatorHash = errors.New("missing or invalid validator hash in metadata")
)

// SetLinkNodeID stores the peerID in the link's metadata.
func SetLinkNodeID(link *cs.Link, peerID peer.ID) {
	if link == nil {
		return
	}

	if link.Meta.Data == nil {
		link.Meta.Data = make(map[string]interface{})
	}

	// This is useful for end users.
	link.Meta.Data[NodeIDKey] = peerID.Pretty()
}

// GetLinkNodeID gets the peerID from the link's metadata.
func GetLinkNodeID(link *cs.Link) (peer.ID, error) {
	if link == nil {
		return "", errors.New("link is nil")
	}

	if link.Meta.Data == nil {
		return "", ErrInvalidMetaNodeID
	}

	nodeID, ok := link.Meta.Data[NodeIDKey].(string)
	if !ok {
		return "", ErrInvalidMetaNodeID
	}

	peerID, err := peer.IDB58Decode(nodeID)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return peerID, nil
}

// SetValidatorHash stores the link validator's hash in the link's metadata.
func SetValidatorHash(link *cs.Link, validatorHash *types.Bytes32) {
	if link == nil || validatorHash == nil {
		return
	}

	if link.Meta.Data == nil {
		link.Meta.Data = make(map[string]interface{})
	}

	link.Meta.Data[ValidatorHashKey] = validatorHash.String()
}

// GetValidatorHash gets the link validator's hash from the link's metadata.
func GetValidatorHash(link *cs.Link) (*types.Bytes32, error) {
	if link == nil {
		return nil, errors.New("link is nil")
	}

	if link.Meta.Data == nil {
		return nil, ErrInvalidValidatorHash
	}

	validatorHash, ok := link.Meta.Data[ValidatorHashKey].(string)
	if !ok {
		return nil, ErrInvalidValidatorHash
	}

	validatorHashBytes, err := types.NewBytes32FromString(validatorHash)
	if err != nil {
		return nil, ErrInvalidValidatorHash
	}

	return validatorHashBytes, nil
}
