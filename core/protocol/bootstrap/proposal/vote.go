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

package proposal

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/pb/crypto"

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
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
func NewVote(sk ic.PrivKey, r *Request) (*Vote, error) {
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

	v.Signature, err = crypto.Sign(sk, payload)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Verify that vote is valid for the given request.
func (v *Vote) Verify(r *Request) error {
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

	valid := v.Signature.Verify(payload)
	if !valid {
		return ErrInvalidSignature
	}

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
