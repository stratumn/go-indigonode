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

package manager

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/manager/grpc"
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
