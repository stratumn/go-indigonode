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
	"crypto/rand"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/pb"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// Errors used by the request component.
var (
	ErrInvalidPeerID   = errors.New("invalid peer ID")
	ErrInvalidPeerAddr = errors.New("invalid peer address")
	ErrMissingPeerAddr = errors.New("missing peer address")
	ErrMissingRequest  = errors.New("no request for the given peer ID")
)

const (
	// DefaultExpiration expires requests after a week.
	DefaultExpiration = 24 * 7 * time.Hour
)

// Type defines the types of proposals supported.
type Type int

const (
	// AddNode adds a node to the network.
	AddNode Type = 0
	// RemoveNode removes a node from the network.
	RemoveNode Type = 1
)

// String returns a friendly name for the type of proposal.
func (t Type) String() string {
	names := []string{
		"Add",
		"Remove",
	}

	if t < AddNode || t > RemoveNode {
		return "Unknown"
	}

	return names[t]
}

// Request packages all the elements of a network update request.
type Request struct {
	Type     Type
	PeerID   peer.ID
	PeerAddr multiaddr.Multiaddr
	Info     []byte

	Challenge []byte
	Expires   time.Time
}

// NewAddRequest creates a request to add a node to the network.
func NewAddRequest(nodeID *pb.NodeIdentity) (*Request, error) {
	peerID, err := peer.IDFromBytes(nodeID.PeerId)
	if err != nil {
		return nil, ErrInvalidPeerID
	}

	if nodeID.PeerAddr == nil {
		return nil, ErrMissingPeerAddr
	}

	peerAddr, err := multiaddr.NewMultiaddrBytes(nodeID.PeerAddr)
	if err != nil {
		return nil, ErrInvalidPeerAddr
	}

	return &Request{
		Type:     AddNode,
		PeerID:   peerID,
		PeerAddr: peerAddr,
		Info:     nodeID.IdentityProof,
		Expires:  time.Now().UTC().Add(DefaultExpiration),
	}, nil
}

// NewRemoveRequest creates a request to remove a node from the network.
func NewRemoveRequest(nodeID *pb.NodeIdentity) (*Request, error) {
	peerID, err := peer.IDFromBytes(nodeID.PeerId)
	if err != nil {
		return nil, ErrInvalidPeerID
	}

	challenge := make([]byte, 32)
	_, err = rand.Read(challenge)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Request{
		Type:      RemoveNode,
		PeerID:    peerID,
		Challenge: challenge,
		Expires:   time.Now().UTC().Add(DefaultExpiration),
	}, nil
}

// ToUpdateProposal converts to a protobuf format.
func (r *Request) ToUpdateProposal() *pb.UpdateProposal {
	typeMap := map[Type]pb.UpdateType{
		AddNode:    pb.UpdateType_AddNode,
		RemoveNode: pb.UpdateType_RemoveNode,
	}

	prop := pb.UpdateProposal{
		UpdateType: typeMap[r.Type],
		Challenge:  r.Challenge,
		NodeDetails: &pb.NodeIdentity{
			PeerId:        []byte(r.PeerID),
			IdentityProof: r.Info,
		},
	}

	if r.PeerAddr != nil {
		prop.NodeDetails.PeerAddr = r.PeerAddr.Bytes()
	}

	return &prop
}

// FromUpdateProposal converts from a protobuf format.
func (r *Request) FromUpdateProposal(p *pb.UpdateProposal) error {
	typeMap := map[pb.UpdateType]Type{
		pb.UpdateType_AddNode:    AddNode,
		pb.UpdateType_RemoveNode: RemoveNode,
	}

	if p.NodeDetails == nil {
		return ErrInvalidPeerID
	}

	peerID, err := peer.IDFromBytes(p.NodeDetails.PeerId)
	if err != nil {
		return ErrInvalidPeerID
	}

	var peerAddr multiaddr.Multiaddr
	if len(p.NodeDetails.PeerAddr) > 0 {
		peerAddr, err = multiaddr.NewMultiaddrBytes(p.NodeDetails.PeerAddr)
		if err != nil {
			return ErrInvalidPeerAddr
		}
	}

	r.Type = typeMap[p.UpdateType]
	r.PeerID = peerID
	r.PeerAddr = peerAddr
	r.Info = p.NodeDetails.IdentityProof
	r.Challenge = p.Challenge
	r.Expires = time.Now().UTC().Add(DefaultExpiration)

	return nil
}

// MarshalJSON marshals the request to JSON.
func (r *Request) MarshalJSON() ([]byte, error) {
	toSerialize := struct {
		Type      Type
		PeerID    []byte
		PeerAddr  []byte
		Info      []byte
		Challenge []byte
		Expires   time.Time
	}{
		Type:      r.Type,
		PeerID:    []byte(r.PeerID),
		Info:      r.Info,
		Challenge: r.Challenge,
		Expires:   r.Expires,
	}

	if r.PeerAddr != nil {
		toSerialize.PeerAddr = r.PeerAddr.Bytes()
	}

	return json.Marshal(toSerialize)
}

// UnmarshalJSON unmarshals the request from JSON.
func (r *Request) UnmarshalJSON(data []byte) error {
	deserialized := struct {
		Type      Type
		PeerID    []byte
		PeerAddr  []byte
		Info      []byte
		Challenge []byte
		Expires   time.Time
	}{}

	err := json.Unmarshal(data, &deserialized)
	if err != nil {
		return errors.WithStack(err)
	}

	peerID, err := peer.IDFromBytes(deserialized.PeerID)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(deserialized.PeerAddr) > 0 {
		peerAddr, err := multiaddr.NewMultiaddrBytes(deserialized.PeerAddr)
		if err != nil {
			return errors.WithStack(err)
		}

		r.PeerAddr = peerAddr
	}

	r.Type = deserialized.Type
	r.PeerID = peerID
	r.Info = deserialized.Info
	r.Challenge = deserialized.Challenge
	r.Expires = deserialized.Expires

	return nil
}
