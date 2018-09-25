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

package proposal

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/app/bootstrap/pb"
	"github.com/stratumn/go-node/core/crypto"

	ic "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

// Errors used by the Vote struct.
var (
	ErrMissingChallenge   = errors.New("missing challenge")
	ErrMissingPrivateKey  = errors.New("missing private key")
	ErrInvalidRequestType = errors.New("invalid request type")
	ErrInvalidChallenge   = errors.New("invalid challenge")
	ErrInvalidSignature   = errors.New("invalid signature")
)

// Vote for a network update.
type Vote struct {
	Type   Type
	PeerID peer.ID

	Challenge []byte
	Signature *crypto.Signature
}

// NewVote votes for a given request.
func NewVote(ctx context.Context, sk ic.PrivKey, r *Request) (*Vote, error) {
	if sk == nil {
		return nil, ErrMissingPrivateKey
	}

	if len(r.Challenge) == 0 {
		return nil, ErrMissingChallenge
	}

	if r.Type < AddNode || r.Type > RemoveNode {
		return nil, ErrInvalidRequestType
	}

	v := &Vote{
		Type:      r.Type,
		PeerID:    r.PeerID,
		Challenge: r.Challenge,
	}

	payload, err := json.Marshal(v)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	v.Signature, err = crypto.Sign(ctx, sk, payload)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Verify that vote is valid for the given request.
func (v *Vote) Verify(ctx context.Context, r *Request) error {
	if v.Type != r.Type {
		return ErrInvalidRequestType
	}

	if v.PeerID != r.PeerID {
		return ErrInvalidPeerID
	}

	if !bytes.Equal(v.Challenge, r.Challenge) {
		return ErrInvalidChallenge
	}

	signature := v.Signature
	v.Signature = nil
	payload, err := json.Marshal(v)
	v.Signature = signature

	if err != nil {
		return errors.WithStack(err)
	}

	valid := v.Signature.Verify(ctx, payload)
	if !valid {
		return ErrInvalidSignature
	}

	return nil
}

// ToProtoVote converts to a protobuf message.
func (v *Vote) ToProtoVote() *pb.Vote {
	typeMap := map[Type]pb.UpdateType{
		AddNode:    pb.UpdateType_AddNode,
		RemoveNode: pb.UpdateType_RemoveNode,
	}

	return &pb.Vote{
		UpdateType: typeMap[v.Type],
		PeerId:     []byte(v.PeerID),
		Challenge:  v.Challenge,
		Signature:  v.Signature,
	}
}

// FromProtoVote converts from a protobuf message.
func (v *Vote) FromProtoVote(proto *pb.Vote) error {
	typeMap := map[pb.UpdateType]Type{
		pb.UpdateType_AddNode:    AddNode,
		pb.UpdateType_RemoveNode: RemoveNode,
	}

	peerID, err := peer.IDFromBytes(proto.PeerId)
	if err != nil {
		return ErrInvalidPeerID
	}

	v.Type = typeMap[proto.UpdateType]
	v.PeerID = peerID
	v.Challenge = proto.Challenge
	v.Signature = proto.Signature

	return nil
}

// MarshalJSON marshals the vote to JSON.
func (v *Vote) MarshalJSON() ([]byte, error) {
	toSerialize := struct {
		Type      Type
		PeerID    []byte
		Challenge []byte
		Signature *crypto.Signature
	}{
		Type:      v.Type,
		PeerID:    []byte(v.PeerID),
		Challenge: v.Challenge,
		Signature: v.Signature,
	}

	return json.Marshal(toSerialize)
}

// UnmarshalJSON unmarshals the vote from JSON.
func (v *Vote) UnmarshalJSON(data []byte) error {
	deserialized := struct {
		Type      Type
		PeerID    []byte
		Challenge []byte
		Signature *crypto.Signature
	}{}

	err := json.Unmarshal(data, &deserialized)
	if err != nil {
		return errors.WithStack(err)
	}

	peerID, err := peer.IDFromBytes(deserialized.PeerID)
	if err != nil {
		return errors.WithStack(err)
	}

	v.Type = deserialized.Type
	v.PeerID = peerID
	v.Challenge = deserialized.Challenge
	v.Signature = deserialized.Signature

	return nil
}
