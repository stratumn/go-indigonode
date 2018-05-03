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

//go:generate mockgen -package mockencoder -destination mockencoder/mockencoder.go github.com/stratumn/alice/core/protocol/storage/p2p Encoder

package p2p

import (
	"context"
	"encoding/hex"
	"io"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/storage/constants"
	"github.com/stratumn/alice/core/protocol/storage/file"
	pb "github.com/stratumn/alice/pb/storage"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// log is the logger for the service.
var log = logging.Logger("storage.p2p")

// ChunkSize size of a chunk of data
const ChunkSize = 1024

// Encoder is an interface that implements an Encode method.
type Encoder interface {
	Encode(interface{}) error
}

// P2P is where the p2p APIs are defined.
type P2P interface {
	PullFile(ctx context.Context, fileHash []byte, pid peer.ID) error
	SendFile(ctx context.Context, enc Encoder, fileHash []byte) error
}

type p2p struct {
	host        ihost.Host
	fileHandler file.Handler
	chunkSize   int
}

// NewP2P returns a new p2p handler.
func NewP2P(host ihost.Host, fh file.Handler) P2P {
	return &p2p{
		host:        host,
		chunkSize:   ChunkSize,
		fileHandler: fh,
	}
}

// PullFile pulls a file from a peer given the file hash.
func (p *p2p) PullFile(ctx context.Context, fileHash []byte, peerID peer.ID) error {

	event := log.EventBegin(ctx, "PullFile", logging.Metadata{
		"fileHash": hex.EncodeToString(fileHash),
		"peerID":   peerID,
	})
	defer event.Done()

	stream, err := p.host.NewStream(ctx, peerID, constants.ProtocolID)
	if err != nil {
		return errors.WithStack(err)
	}

	// Send the request
	req := &pb.FileInfo{Hash: fileHash}
	enc := protobuf.Multicodec(nil).Encoder(stream)
	if err = enc.Encode(req); err != nil {
		return errors.WithStack(err)
	}

	// Get the response
	dec := protobuf.Multicodec(nil).Decoder(stream)

	var chunk pb.FileChunk
	if err = dec.Decode(&chunk); err != nil {
		return errors.WithStack(err)
	}

	sessionID, err := p.fileHandler.BeginWrite(ctx, chunk.FileName)
	if err != nil {
		event.SetError(err)
		return err
	}

LOOP:
	for {
		select {

		case <-ctx.Done():
			err := ctx.Err()
			if err2 := p.fileHandler.AbortWrite(ctx, sessionID); err2 != nil {
				err = errors.Wrap(err, err2.Error())
			}
			event.SetError(err)
			return err

		default:
			err := p.fileHandler.WriteChunk(ctx, sessionID, chunk.Data)
			if err != nil {
				event.SetError(err)
				return err
			}

			if err = dec.Decode(&chunk); err != nil {
				if err2 := p.fileHandler.AbortWrite(ctx, sessionID); err2 != nil {
					err = errors.Wrap(err, err2.Error())
				}
				event.SetError(err)
				return err
			}

			if len(chunk.Data) == 0 {
				break LOOP
			}
		}
	}

	// finalize the writing.
	_, err = p.fileHandler.EndWrite(ctx, sessionID)
	if err != nil {
		event.SetError(err)
		return err
	}
	return nil
}

func (p *p2p) SendFile(ctx context.Context, enc Encoder, fileHash []byte) error {
	event := log.EventBegin(ctx, "SendFile", &logging.Metadata{
		"file_hash": hex.EncodeToString(fileHash),
	})

	sessionID, filePath, err := p.fileHandler.BeginRead(ctx, fileHash)
	if err != nil {
		event.SetError(err)
		return err
	}
	first := true

LOOP:
	for {
		select {
		case <-ctx.Done():
			event.SetError(ctx.Err())
			return ctx.Err()

		default:
			data, err := p.fileHandler.ReadChunk(ctx, sessionID, p.chunkSize)
			if err == io.EOF {
				break LOOP
			}
			if err != nil {
				event.SetError(err)
				return err
			}

			chunk := &pb.FileChunk{Data: data}
			if first {
				first = false
				chunk.FileName = filepath.Base(filePath)
			}

			if err = enc.Encode(chunk); err != nil {
				if err2 := p.fileHandler.EndRead(ctx, sessionID); err2 != nil {
					err = errors.Wrap(err, err2.Error())
				}
				event.SetError(err)
				return err
			}
		}

	}

	if err := p.fileHandler.EndRead(ctx, sessionID); err != nil {
		event.Append(&logging.Metadata{"end_read_error": err.Error()})
	}

	// Send an empty chunk to notify end of file.
	return enc.Encode(&pb.FileChunk{})
}
