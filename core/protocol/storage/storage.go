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

package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/db"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

// ChunkSize size of a chunk of data
const ChunkSize = 1024

// Host represents an Alice host.
type Host = ihost.Host

// ProtocolID is the protocol ID of the service.
var ProtocolID = protocol.ID("/alice/storage/v1.0.0")

// log is the logger for the service.
var log = logging.Logger("storage")

var (
	prefixFilesHashes     = []byte("fh") // prefixFilesHashes + filepath -> filehash
	prefixAuthorizedPeers = []byte("ap") // prefixAuthorizedPeers + filehash -> [peerIDs...]
)

// Storage implements the storage protocol.
type Storage struct {
	db        db.DB
	host      Host
	chunkSize int
}

// NewStorage creates a new storage server.
func NewStorage(host Host, db db.DB) *Storage {
	return &Storage{
		db:        db,
		host:      host,
		chunkSize: ChunkSize,
	}
}

// IndexFile adds the file hash and name to the db.
func (s *Storage) IndexFile(ctx context.Context, file *os.File, fileName string) (fileHash []byte, err error) {
	// go back to the beginning of the file.
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err = io.Copy(h, file); err != nil {
		return
	}

	fileHash, err = mh.Encode(h.Sum(nil), mh.SHA2_256)
	if err != nil {
		return
	}

	if err = s.db.Put(append(prefixFilesHashes, fileHash...), []byte(fileName)); err != nil {
		return
	}

	return
}

// Keep track of the authorized peers in a map
// to have efficient insert.
type authorizedPeersMap map[peer.ID]struct{}

// Authorize adds a list of peers to the authorized peers for a file hash.
func (s *Storage) Authorize(ctx context.Context, pids [][]byte, fileHash []byte) error {

	// Deserialize peer IDs.
	peerIds := make([]peer.ID, len(pids))
	for i, p := range pids {
		pid, err := peer.IDFromBytes(p)
		if err != nil {
			return err
		}
		peerIds[i] = pid
	}
	event := log.EventBegin(ctx, "AuthorizePeer", &logging.Metadata{
		"peerIDs": peerIds,
	})
	defer event.Done()

	// Get authorized peers from DB and update them.
	authorizedPeers := authorizedPeersMap{}
	ap, err := s.db.Get(append(prefixAuthorizedPeers, fileHash...))
	if err != nil && errors.Cause(err) != db.ErrNotFound {
		event.SetError(err)
		return err
	} else if errors.Cause(err) != db.ErrNotFound {
		if err = json.Unmarshal(ap, &authorizedPeers); err != nil {
			event.SetError(err)
			return err
		}
	}

	for _, p := range peerIds {
		authorizedPeers[p] = struct{}{}
	}

	newAp, err := json.Marshal(authorizedPeers)
	if err != nil {
		event.SetError(err)
		return err
	}

	if err = s.db.Put(append(prefixAuthorizedPeers, fileHash...), newAp); err != nil {
		event.SetError(err)
		return err
	}

	return nil
}

// StreamHandler handles incoming messages from peers.
func (s *Storage) StreamHandler(ctx context.Context, stream inet.Stream) {
}
