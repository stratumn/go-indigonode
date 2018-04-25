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

package p2p

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/storage/file"
	pb "github.com/stratumn/alice/pb/storage"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	multicodec "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// log is the logger for the service.
var log = logging.Logger("storage.p2p")

// ChunkSize size of a chunk of data
const ChunkSize = 1024

// P2P is where the p2p APIs are defined.
type P2P interface {
	PullFile(ctx context.Context, fileHash []byte, pid peer.ID) error
	SendFile(ctx context.Context, enc multicodec.Encoder, path string) error
}

// P2P is where the p2p APIs are defined.
type p2p struct {
	host       ihost.Host
	protocolID protocol.ID

	fileHandler file.Handler
	chunkSize   int
}

// NewP2P returns a new p2p handler.
func NewP2P(host ihost.Host, p protocol.ID, fh file.Handler) P2P {
	return &p2p{
		host:        host,
		protocolID:  p,
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

	stream, err := p.host.NewStream(ctx, peerID, p.protocolID)
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

	chunkCh := make(chan *pb.FileChunk)
	resCh := make(chan error)
	go func() {
		_, err := p.fileHandler.WriteFile(ctx, chunkCh)
		resCh <- err
	}()

	for {
		select {

		case <-ctx.Done():
			return errors.WithStack(ctx.Err())

		case err = <-resCh:
			fmt.Printf("\n%+v\n", err)
			return err

		default:
			var chunk pb.FileChunk
			if err = dec.Decode(&chunk); err != nil {
				return errors.WithStack(err)
			}

			if len(chunk.Data) == 0 {
				// End of file
				close(chunkCh)
				return nil
			}

			chunkCh <- &chunk
		}
	}
}

func (p *p2p) SendFile(ctx context.Context, enc multicodec.Encoder, path string) error {

	successCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	go func() {

		file, err := os.Open(path)
		if err != nil {
			errCh <- errors.WithStack(err)
			return
		}

		buf := make([]byte, p.chunkSize)
		eof := false
		first := true

		for !eof {
			// put as many bytes as `chunkSize` into the buf array.
			// n is the actual number of bytes read in case we reached end of file.
			n, err := file.Read(buf)

			if n == 0 {
				// No more bytes to send, break loop directly
				break
			}

			if n < p.chunkSize {
				// This is the last chunk of data.
				eof = true
			}

			chunk := &pb.FileChunk{Data: buf[:n]}
			if first {
				// Add the file name in the first message.
				chunk.FileName = filepath.Base(path)

			}

			err = enc.Encode(chunk)
			if err != nil {
				errCh <- errors.WithStack(err)
				return
			}
		}

		// Send an empty chunk to notify end of file.
		err = enc.Encode(&pb.FileChunk{})
		if err != nil {
			errCh <- errors.WithStack(err)
			return
		}

		successCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	case <-successCh:
		return nil
	case err := <-errCh:
		return err
	}
}
