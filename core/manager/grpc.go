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

package manager

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/manager"
)

// GRPCManager represents a manager that can be used by the gRPC server.
type GRPCManager interface {
	Proto(string) (*pb.Service, error)
	List() []string
	Start(string) error
	Stop(string) error
	Prune()
}

// grpcServer is a gRPC server for the manager service.
type grpcServer struct {
	mgr GRPCManager
}

// List lists the available services.
func (s grpcServer) List(req *pb.ListReq, srv pb.Manager_ListServer) error {
	return s.sendServices(srv.Send)
}

// Info returns information on a service.
func (s grpcServer) Info(ctx context.Context, req *pb.InfoReq) (*pb.Service, error) {
	if req.Id == "" {
		return nil, errors.WithStack(ErrMissingServiceID)
	}

	return s.mgr.Proto(req.Id)
}

// Start starts a service.
func (s grpcServer) Start(ctx context.Context, req *pb.StartReq) (*pb.Service, error) {
	if req.Id == "" {
		return nil, errors.WithStack(ErrMissingServiceID)
	}

	if err := s.mgr.Start(req.Id); err != nil {
		return nil, err
	}

	return s.mgr.Proto(req.Id)
}

// Stop stops a service.
func (s grpcServer) Stop(ctx context.Context, req *pb.StopReq) (*pb.Service, error) {
	if req.Id == "" {
		return nil, errors.WithStack(ErrMissingServiceID)
	}

	done := make(chan error, 1)
	go func() {
		done <- s.mgr.Stop(req.Id)
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, err
		}

		if req.Prune {
			pruned := make(chan struct{})
			go func() {
				s.mgr.Prune()
				close(pruned)
			}()

			select {
			case <-pruned:
			case <-ctx.Done():
				return nil, errors.WithStack(ctx.Err())
			}
		}

		return s.mgr.Proto(req.Id)

	case <-ctx.Done():
		return nil, errors.WithStack(ctx.Err())
	}
}

// Prune prunes the services.
func (s grpcServer) Prune(req *pb.PruneReq, srv pb.Manager_PruneServer) error {
	// Avoid race condition if we stop ourself.
	done := make(chan struct{})
	go func() {
		s.mgr.Prune()
		close(done)
	}()

	select {
	case <-done:
		return s.sendServices(srv.Send)
	case <-srv.Context().Done():
		return errors.WithStack(srv.Context().Err())
	}
}

// sendServices is used to send all the services to a stream.
func (s grpcServer) sendServices(send func(*pb.Service) error) error {
	for _, sid := range s.mgr.List() {
		msg, err := s.mgr.Proto(sid)
		if err != nil {
			return err
		}

		if err := send(msg); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
