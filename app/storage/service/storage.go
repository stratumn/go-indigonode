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

// Package service defines types for the storage service.
package service

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	grpcpb "github.com/stratumn/go-node/app/storage/grpc"
	"github.com/stratumn/go-node/app/storage/protocol"
	"github.com/stratumn/go-node/app/storage/protocol/constants"
	"github.com/stratumn/go-node/core/db"

	"google.golang.org/grpc"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

var log = logging.Logger("storage.service")

// Host represents an Indigo Node host.
type Host = ihost.Host

// Service is the Storage service.
type Service struct {
	config        *Config
	host          Host
	storage       *protocol.Storage
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

	cwd, err := os.Getwd()
	if err != nil {
		panic(errors.WithStack(err))
	}

	localStoragePath, err := filepath.Abs(filepath.Join(cwd, "data", "storage", "files"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	dbPath, err := filepath.Abs(filepath.Join(cwd, "data", "storage", "db"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	// Set the default configuration settings of your service here.
	return Config{
		Host:          "host",
		LocalStorage:  localStoragePath,
		DbPath:        dbPath,
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

	s.storage = protocol.NewStorage(s.host, db, s.config.LocalStorage)

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
