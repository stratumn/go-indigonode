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

/*
Package service defines a service that exposes a gRPC API.

The API allows external applications such as Indigo Node's command line interface
to interact with the node without having to connect to the P2P network.

The API is implemented using gRPC. It is possible to automatically generate API
clients for other programming languages such as Javascript and C++.

The Protobuf types used by the API are defined in the `/grpc` folder.

For more information about gRPC, see:

	https://grpc.io
*/
package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/app/grpcapi/grpc"
	"github.com/stratumn/go-indigonode/core/manager"
	"github.com/stratumn/go-indigonode/core/netutil"
	"github.com/stratumn/go-indigonode/release"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

var (
	// ErrNotManager is returned when the connected service is not a
	// manager.
	ErrNotManager = errors.New("connected service is not a manager")

	// ErrPeerNotFound is returned when an interceptor cannot get a peer
	// from a request.
	ErrPeerNotFound = errors.New("could not get peer")
)

// log is the logger the the gRPC API service.
var log = logging.Logger("grpcapi")

// Manager represents a service manager.
type Manager interface {
	List() []string
	Find(string) (manager.Service, error)
}

// Registrable represents something that can add itself to the gRPC server, so
// that other services can add functions to the API.
type Registrable interface {
	AddToGRPCServer(*grpc.Server)
}

// Service is the gRPC API service.
type Service struct {
	config *Config
	mgr    Manager
	ctx    context.Context
	server *grpc.Server
}

// Config contains configuration options for the gRPC API service.
type Config struct {
	// Manager is the name of the manager service.
	Manager string `toml:"manager" comment:"The name of the manager service."`

	// Address is the address to bind to.
	Address string `toml:"address" comment:"Address to bind to."`

	// TLSCertificateFile is a path to a TLS certificate.
	TLSCertificateFile string `toml:"tls_certificate_file" comment:"Path to a TLS certificate."`

	// TLSKeyFile is a path to a TLS key.
	TLSKeyFile string `toml:"tls_key_file" comment:"Path to a TLS key."`

	// EnableRequestLogger is whether to log unary and stream requests.
	EnableRequestLogger bool `toml:"enable_request_logger" comment:"Whether to log requests."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "grpcapi"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "GRPC API"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Starts a gRPC API server."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Manager:             "manager",
		Address:             "/ip4/127.0.0.1/tcp/8904",
		EnableRequestLogger: true,
	}
}

// SetConfig configures the service handler.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	if _, err := ma.NewMultiaddr(conf.Address); err != nil {
		return errors.WithStack(err)
	}

	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Manager] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(handlers map[string]interface{}) error {
	var ok bool

	if s.mgr, ok = handlers[s.config.Manager].(Manager); !ok {
		return errors.Wrap(ErrNotManager, s.config.Manager)
	}

	return nil
}

// Expose exposes the service to other services.
//
// It exposes the type:
//
//	google.golang.org/grpc.Host
func (s *Service) Expose() interface{} {
	return s.server
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.ctx = ctx
	defer func() { s.ctx = nil }()

	// Start the TCP listener.
	lis, err := netutil.Listen(s.config.Address)
	if err != nil {
		return err
	}

	opts, err := s.grpcOpts()
	if err != nil {
		return err
	}

	s.server = grpc.NewServer(opts...)

	// Add the API service.
	pb.RegisterAPIServer(s.server, s)

	// Register reflection service on gRPC server, which is used by the CLI
	// to reflect commands.
	reflection.Register(s.server)

	// Add all registerable services to the server.
	s.addRegistrables(s.server)

	// Launch a goroutine for the gRPC server.
	done := make(chan error, 1)
	go func() {
		done <- errors.WithStack(s.server.Serve(lis))
	}()

	running()

	// Handle exit conditions.
	select {
	case err = <-done:
		stopping()
		return err
	case <-ctx.Done():
		stopping()
		s.server.GracefulStop()
		err = <-done
	}

	cause := errors.Cause(err)

	if isGRPCShutdownBug(cause) || cause == grpc.ErrServerStopped {
		err = nil
	}

	if err == nil {
		return errors.WithStack(ctx.Err())
	}

	return err
}

// Inform returns information about the API.
func (s *Service) Inform(ctx context.Context, req *pb.InformReq) (*pb.Info, error) {
	return &pb.Info{
		Protocol:  release.Protocol,
		Version:   release.Version,
		GitCommit: release.GitCommitBytes,
	}, nil
}

// grpcOpts builds the gRPC server options.
func (s *Service) grpcOpts() ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption

	// Add interceptors for logging if enabled.
	if s.config.EnableRequestLogger {
		log.Event(s.ctx, "requestLoggerEnabled")
		opts = append(opts, grpc.UnaryInterceptor(logRequest))
		opts = append(opts, grpc.StreamInterceptor(logStream))
	} else {
		log.Event(s.ctx, "requestLoggerDisabled")
	}

	// Enable TLS if files are provided.
	cert, key := s.config.TLSCertificateFile, s.config.TLSKeyFile
	if cert != "" && key != "" {
		log.Event(s.ctx, "tlsEnabled")
		creds, err := credentials.NewServerTLSFromFile(cert, key)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		opts = append(opts, grpc.Creds(creds))
	} else {
		log.Event(s.ctx, "tlsDisabled")
	}

	return opts, nil
}

// addRegistrables finds the registrables and adds them to the gRPC server.
func (s *Service) addRegistrables(gs *grpc.Server) {
	for _, sid := range s.mgr.List() {
		service, err := s.mgr.Find(sid)
		if err != nil {
			log.Event(s.ctx, "managerFindError", logging.Metadata{
				"error": err.Error(),
			})
			continue
		}

		reg, ok := service.(Registrable)
		if ok {
			log.Event(s.ctx, "register", logging.Metadata{
				"service": service.ID(),
			})
			reg.AddToGRPCServer(gs)
		}
	}
}
