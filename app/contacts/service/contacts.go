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

// Package service defines types for the contacts service.
package service

import (
	"context"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-node/app/contacts/grpc"
	"google.golang.org/grpc"
)

// log is the logger for the service.
var log = logging.Logger("contacts")

// Service is the Contacts service.
type Service struct {
	config *Config
	mgr    *Manager
}

// Config contains configuration options for the Contacts service.
type Config struct {
	// Filename is the filename of the contacts file.
	Filename string `toml:"filename" comment:"The filename of the contacts file."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "contacts"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Contacts"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Manages contacts."
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

	filename, err := filepath.Abs(filepath.Join(cwd, "data", "contacts.toml"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	return Config{
		Filename: filename,
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)
	s.config = &conf
	return nil
}

// Expose exposes the service to other services.
func (s *Service) Expose() interface{} {
	return s.mgr
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	mgr, err := NewManager(s.config.Filename)
	if err != nil {
		return nil
	}

	s.mgr = mgr

	running()
	<-ctx.Done()
	stopping()

	s.mgr = nil

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterContactsServer(gs, grpcServer{
		GetManager: func() *Manager {
			return s.mgr
		},
	})
}
