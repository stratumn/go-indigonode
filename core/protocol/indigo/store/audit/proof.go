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

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	"github.com/stratumn/alice/pb/crypto"
	"github.com/stratumn/go-indigocore/cs"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

var (
	// ErrMissingPeerSignature is returned when there is no peer signature
	// on a segment.
	ErrMissingPeerSignature = errors.New("missing peer signature")

	// ErrInvalidPeerSignature is returned when an invalid proof is provided.
	ErrInvalidPeerSignature = errors.New("invalid peer signature")
)

var (
	// PeerSignatureBackend is the name used as the PeerSignature proof backend.
	PeerSignatureBackend = "peer_signature"
)

// SignLink signs a link before publishing it to the network.
func SignLink(ctx context.Context, sk ic.PrivKey, link *cs.Link) (segment *cs.Segment, err error) {
	event := log.EventBegin(ctx, "SignLink")
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	if sk == nil || link == nil {
		return nil, errors.New("secret key or link missing")
	}

	peerID, err := peer.IDFromPrivateKey(sk)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	linkPeerID, err := constants.GetLinkNodeID(link)
	if err != nil {
		return nil, err
	}

	if linkPeerID != peerID {
		return nil, constants.ErrInvalidMetaNodeID
	}

	segment = &cs.Segment{Link: *link}
	err = segment.SetLinkHash()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	proof, err := NewPeerSignature(sk, segment)
	if err != nil {
		return nil, err
	}

	err = segment.Meta.AddEvidence(cs.Evidence{
		Backend:  PeerSignatureBackend,
		Provider: peerID.Pretty(),
		Proof:    proof,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return segment, nil
}

// PeerSignature implements github.com/stratumn/go-indigocore/cs Proof interface.
// A peer signs the link before publishing it to the network.
type PeerSignature struct {
	Timestamp uint64            `json:"timestamp"`
	PeerID    []byte            `json:"peer_id"`
	LinkHash  []byte            `json:"link_hash"`
	Signature *crypto.Signature `json:"signature"`
}

// NewPeerSignature creates a signed proof that the peer
// approves this segment.
func NewPeerSignature(sk ic.PrivKey, segment *cs.Segment) (cs.Proof, error) {
	sig, err := crypto.Sign(sk, segment.GetLinkHash()[:])
	if err != nil {
		return nil, err
	}

	peerID, _ := peer.IDFromPrivateKey(sk)

	return &PeerSignature{
		Timestamp: uint64(time.Now().Unix()),
		PeerID:    []byte(peerID),
		LinkHash:  segment.GetLinkHash()[:],
		Signature: sig,
	}, nil
}

// Time returns the timestamp of the signature.
// This is just for information and isn't signed.
func (p *PeerSignature) Time() uint64 {
	return p.Timestamp
}

// FullProof returns a JSON formatted proof.
func (p *PeerSignature) FullProof() []byte {
	bytes, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		return nil
	}
	return bytes
}

// Verify returns true if the proof is valid.
func (p *PeerSignature) Verify(data interface{}) bool {
	if p.Signature == nil {
		return false
	}

	linkHash, ok := data.([]byte)
	if !ok {
		return false
	}

	if !bytes.Equal(linkHash, p.LinkHash) {
		return false
	}

	peerID, err := peer.IDFromBytes(p.PeerID)
	if err != nil {
		return false
	}

	pk, err := ic.UnmarshalPublicKey(p.Signature.PublicKey)
	if err != nil {
		return false
	}

	if !peerID.MatchesPublicKey(pk) {
		return false
	}

	return p.Signature.Verify(linkHash)
}

// init needs to define a way to deserialize a PeerSignature proof.
func init() {
	cs.DeserializeMethods[PeerSignatureBackend] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := PeerSignature{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}

		return &p, nil
	}
}
