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

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigonode/core/crypto"
	"github.com/stratumn/go-indigonode/core/monitoring"

	ic "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

var (
	// ErrMissingPeerSignature is returned when there is no peer signature
	// on a segment.
	ErrMissingPeerSignature = errors.New("missing peer signature")

	// ErrInvalidPeerSignature is returned when an invalid proof is provided.
	ErrInvalidPeerSignature = errors.New("invalid peer signature")
)

const (
	// PeerSignatureBackend is the name used as the PeerSignature proof backend.
	PeerSignatureBackend = "peer_signature"
)

// SignLink signs a link before publishing it to the network.
func SignLink(ctx context.Context, sk ic.PrivKey, link *cs.Link) (segment *cs.Segment, err error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.audit", "SignLink")
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
		}

		span.End()
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

	proof, err := NewPeerSignature(ctx, sk, segment)
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
func NewPeerSignature(ctx context.Context, sk ic.PrivKey, segment *cs.Segment) (cs.Proof, error) {
	sig, err := crypto.Sign(ctx, sk, segment.GetLinkHash()[:])
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

	return p.Signature.Verify(context.TODO(), linkHash)
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
