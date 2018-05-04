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

// Package storage is a simple service that allows one peer to share a file to another peer.
package storage

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/stratumn/alice/core/db"
	"github.com/stratumn/alice/core/protocol/storage"
	"github.com/stratumn/alice/core/protocol/storage/constants"
	grpcpb "github.com/stratumn/alice/grpc/storage"

	"google.golang.org/grpc"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

var log = logging.Logger("storage.service")

// Host represents an Alice host.
type Host = ihost.Host

// Service is the Storage service.
type Service struct {
	config        *Config
	host          Host
	storage       *storage.Storage
	uploadTimeout time.Duration
}

// Config contains configuration options for the Storage service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Path to the local storage folder.
	LocalStorage string `toml:"local_storage" comment:"The path to the local storage."`

	// DbPath is the path to the database used for file hashes and permissions.
	DbPath string `toml:"db_path" comment:"The path to the database used for the state and the chain."`

	// UploadTimeout is the time after which an upload session will be reset (and the partial file deleted).
	UploadTimeout string `toml:"upload_timeout" comment:"The time after which an upload session will be reset (and the partial file deleted)"`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "storage"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Storage"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "A basic p2p file storage service."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	// Set the default configuration settings of your service here.
	return Config{
		Host:          "host",
		LocalStorage:  "data/storage/files",
		DbPath:        "data/storage/db",
		UploadTimeout: "10m",
	}
}

// SetConfig configures the service handler.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	timeout, err := time.ParseDuration(conf.UploadTimeout)
	if err != nil {
		return errors.WithStack(err)
	}
	s.uploadTimeout = timeout

	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Host] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(services map[string]interface{}) error {
	var ok bool

	if s.host, ok = services[s.config.Host].(Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	return nil
}

// Expose exposes the storage service to other services.
// It is currently not exposed.
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {

	// Create local folder if not exists.
	if _, err := os.Stat(s.config.LocalStorage); os.IsNotExist(err) {
		if err := os.MkdirAll(s.config.LocalStorage, 0766); err != nil {
			return err
		}
	}

	db, err := db.NewFileDB(s.config.DbPath, nil)
	if err != nil {
		return err
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Event(ctx, "ErrorClosingDB", logging.Metadata{"error": err.Error()})
		}
	}()

	s.storage = storage.NewStorage(s.host, db, s.config.LocalStorage)

	// Wrap the stream handler with the context.
	s.host.SetStreamHandler(constants.ProtocolID, func(stream inet.Stream) {
		s.storage.StreamHandler(ctx, stream)
	})

	running()
	<-ctx.Done()
	stopping()

	// Stop accepting streams.
	s.host.RemoveStreamHandler(constants.ProtocolID)
	s.storage = nil

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	grpcpb.RegisterStorageServer(
		gs,
		&grpcServer{
			beginWrite: func(ctx context.Context, name string) (uuid.UUID, error) {
				if s.storage == nil {
					return uuid.Nil, ErrUnavailable
				}
				return s.storage.FileHandler.BeginWrite(ctx, name)
			},

			writeChunk: func(ctx context.Context, id uuid.UUID, data []byte) error {
				if s.storage == nil {
					return ErrUnavailable
				}
				return s.storage.FileHandler.WriteChunk(ctx, id, data)
			},

			endWrite: func(ctx context.Context, id uuid.UUID) ([]byte, error) {
				if s.storage == nil {
					return nil, ErrUnavailable
				}
				return s.storage.FileHandler.EndWrite(ctx, id)
			},

			abortWrite: func(ctx context.Context, id uuid.UUID) error {
				if s.storage == nil {
					return ErrUnavailable
				}
				return s.storage.FileHandler.AbortWrite(ctx, id)
			},

			authorize: func(ctx context.Context, peerIds [][]byte, fileHash []byte) error {
				if s.storage == nil {
					return ErrUnavailable
				}
				return s.storage.Authorize(ctx, peerIds, fileHash)
			},
			download: func(ctx context.Context, fileHash []byte, peerId []byte) error {
				if s.storage == nil {
					return ErrUnavailable
				}
				return s.storage.PullFile(ctx, fileHash, peerId)
			},
			readChunks: func(ctx context.Context, fileHash []byte, chunkSize int, cr *chunkReader) error {
				if s.storage == nil {
					return ErrUnavailable
				}
				return s.storage.FileHandler.ReadChunks(ctx, fileHash, chunkSize, cr)
			},
			uploadTimeout: s.uploadTimeout,
		})
}
