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

//go:generate mockgen -package mockhandler -destination mockhandler/mockhandler.go github.com/stratumn/alice/core/protocol/storage/file Handler

package file

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/pb/storage"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var log = logging.Logger("storage.file_handler")

var (
	// ErrFileNameMissing is returned when no file name was given.
	ErrFileNameMissing = errors.New("the first chunk should have the filename")

	// ErrUnauthorized is returned when a peer tries to access a file he
	// is not allowed to get
	ErrUnauthorized = errors.New("peer not authorized for requested file")
)

// Handler contains the methods to handle a file on the alice node.
type Handler interface {
	WriteFile(context.Context, <-chan *pb.FileChunk) (*os.File, error)
}

type localFileHandler struct {
	storagePath string
}

// NewLocalFileHandler create a new file Handler.
func NewLocalFileHandler(path string) Handler {
	return &localFileHandler{
		storagePath: path,
	}
}

// SaveFile saves a file locally.
func (h *localFileHandler) WriteFile(ctx context.Context, chunkCh <-chan *pb.FileChunk) (file *os.File, err error) {
	event := log.EventBegin(ctx, "SaveFile")

	defer func() {
		if err != nil {
			// Delete the partially written file.
			if file != nil {
				if err2 := os.Remove(file.Name()); err2 != nil {
					err = errors.Wrap(err, err2.Error())
				}
			}
			event.SetError(err)
		}
		event.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return

		case chunk, ok := <-chunkCh:
			if !ok {
				return
			}
			if file == nil && chunk.FileName == "" {
				err = ErrFileNameMissing
				return
			}

			if file == nil {
				file, err = os.Create(fmt.Sprintf("%s/%s", h.storagePath, chunk.FileName))
				if err != nil {
					return
				}
				event.Append(&logging.Metadata{"filename": file.Name()})
			}

			if _, err = file.Write(chunk.Data); err != nil {
				return
			}
		}
	}
}
