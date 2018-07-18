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

// Package protocol defines types for the storage protocol.
package protocol

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/app/storage/pb"
	"github.com/stratumn/go-indigonode/app/storage/protocol/acl"
	"github.com/stratumn/go-indigonode/app/storage/protocol/file"
	"github.com/stratumn/go-indigonode/app/storage/protocol/p2p"
	"github.com/stratumn/go-indigonode/core/db"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// ChunkSize size of a chunk of data
const ChunkSize = 1024

// Host represents an Indigo Node host.
type Host = ihost.Host

// log is the logger for the service.
var log = logging.Logger("storage")

var (
	// ErrUnauthorized is returned when a peer tries to access a file he
	// is not allowed to get.
	ErrUnauthorized = errors.New("peer not authorized for requested file")
)

// Storage implements the storage protocol.
type Storage struct {
	db          db.DB
	host        Host
	chunkSize   int
	FileHandler file.Handler
	p2p         p2p.P2P
	acl         acl.ACL
}

// NewStorage creates a new storage server.
func NewStorage(host Host, db db.DB, storagePath string) *Storage {
	fh := file.NewLocalFileHandler(storagePath, db)
	return &Storage{
		db:          db,
		host:        host,
		chunkSize:   ChunkSize,
		FileHandler: fh,
		p2p:         p2p.NewP2P(host, fh),
		acl:         acl.NewACL(db),
	}
}

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
	return s.acl.Authorize(ctx, peerIds, fileHash)
}

// PullFile pulls a file from a peer given the file hash if it has not already be downloaded.
func (s *Storage) PullFile(ctx context.Context, fileHash []byte, pid []byte) (err error) {
	exists, err := s.FileHandler.Exists(ctx, fileHash)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	peerID, err := peer.IDFromBytes(pid)
	if err != nil {
		return err
	}
	return s.p2p.PullFile(ctx, fileHash, peerID)
}

// StreamHandler handles incoming messages from peers.
func (s *Storage) StreamHandler(ctx context.Context, stream inet.Stream) {
	from := stream.Conn().RemotePeer()
	event := log.EventBegin(ctx, "SendFile", logging.Metadata{
		"peerID": from.Pretty(),
	})

	defer func() {
		if err := stream.Close(); err != nil {
			event.Append(logging.Metadata{"stream_close_err": err})
		}

		event.Done()
	}()

	dec := protobuf.Multicodec(nil).Decoder(stream)
	enc := protobuf.Multicodec(nil).Encoder(stream)

	var req pb.FileInfo

	err := dec.Decode(&req)
	if err != nil {
		event.SetError(err)
		return
	}

	// Check that the peer is allowed to get the file.
	ok, err := s.acl.IsAuthorized(ctx, from, req.Hash)
	if err != nil {
		event.SetError(err)
		return
	}
	if !ok {
		event.SetError(ErrUnauthorized)
		return
	}

	if err := s.p2p.SendFile(ctx, enc, req.Hash); err != nil {
		event.SetError(err)
		return
	}
}
