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
	"errors"
	"time"

	pb "github.com/stratumn/alice/pb/bootstrap"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// Errors used by the request component.
var (
	ErrInvalidPeerID   = errors.New("invalid peer ID")
	ErrInvalidPeerAddr = errors.New("invalid peer address")
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

	Expires time.Time
}

// NewAddRequest creates a request to add a node to the network.
func NewAddRequest(nodeID *pb.NodeIdentity) (*Request, error) {
	peerID, err := peer.IDFromBytes(nodeID.PeerId)
	if err != nil {
		return nil, ErrInvalidPeerID
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

	return &Request{
		Type:    RemoveNode,
		PeerID:  peerID,
		Expires: time.Now().UTC().Add(DefaultExpiration),
	}, nil
}