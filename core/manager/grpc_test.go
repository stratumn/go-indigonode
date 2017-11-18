// Copyright © 2017  Stratumn SAS
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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager/mock_manager"
	pb "github.com/stratumn/alice/grpc/manager"
	"google.golang.org/grpc"
)

type mockSS struct {
	grpc.ServerStream
	Res []*pb.Service
}

func (m *mockSS) Send(serv *pb.Service) error {
	m.Res = append(m.Res, serv)
	return nil
}

func (m *mockSS) Context() context.Context {
	return context.Background()
}

func TestGRPCServerList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mock_manager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	list := []pb.Service{{
		Id:     "crypto",
		Status: pb.Service_STOPPED,
	}, {
		Id:     "fs",
		Status: pb.Service_RUNNING,
	}}

	mgr.EXPECT().List().Return([]string{"crypto", "fs"}).Times(1)
	mgr.EXPECT().Proto("crypto").Return(&list[0], nil).Times(1)
	mgr.EXPECT().Proto("fs").Return(&list[1], nil).Times(1)

	req, ss := &pb.ListReq{}, &mockSS{}
	if err := s.List(req, ss); err != nil {
		t.Errorf("s.List(req, ss): error: %s", err)
	}

	for i, serv := range list {
		res := ss.Res[i]

		if got, want := res.Id, serv.Id; got != want {
			t.Errorf("res.Id = %q want %q", got, want)
		}

		if got, want := res.Status, serv.Status; got != want {
			t.Errorf("res.Id = %q want %q", got, want)
		}
	}
}

func TestGRPCServerStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mock_manager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serv := pb.Service{
		Id:     "fs",
		Status: pb.Service_RUNNING,
	}

	mgr.EXPECT().Start("fs").Return(nil).Times(1)
	mgr.EXPECT().Proto("fs").Return(&serv, nil).Times(1)

	req := &pb.StartReq{Id: "fs"}
	res, err := s.Start(ctx, req)
	if err != nil {
		t.Errorf("s.Start(ctx, req): error: %s", err)
	}

	if got, want := res.Id, serv.Id; got != want {
		t.Errorf("res.Id = %q want %q", got, want)
	}

	if got, want := res.Status, serv.Status; got != want {
		t.Errorf("res.Status = %q want %q", got, want)
	}

	req = &pb.StartReq{}
	_, err = s.Start(ctx, req)
	if got, want := errors.Cause(err), ErrMissingServiceID; got != want {
		t.Errorf("s.Start(ctx, req): error = %q want %q", got, want)
	}

	mgr.EXPECT().Start("404").Return(ErrNotFound).Times(1)

	req = &pb.StartReq{Id: "404"}
	_, err = s.Start(ctx, req)
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf("s.Start(ctx, req): error = %q want %q", got, want)
	}
}

func TestGRPCServerStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mock_manager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serv := pb.Service{
		Id:     "fs",
		Status: pb.Service_STOPPED,
	}

	mgr.EXPECT().Stop("fs").Return(nil).Times(1)
	mgr.EXPECT().Proto("fs").Return(&serv, nil).Times(1)
	mgr.EXPECT().Prune().Times(0)

	req := &pb.StopReq{Id: "fs"}
	res, err := s.Stop(ctx, req)
	if err != nil {
		t.Errorf("s.Stop(ctx, req): error: %s", err)
	}

	if got, want := res.Id, serv.Id; got != want {
		t.Errorf("res.Id = %q want %q", got, want)
	}

	if got, want := res.Status, serv.Status; got != want {
		t.Errorf("res.Status = %q want %q", got, want)
	}

	req = &pb.StopReq{}
	_, err = s.Stop(ctx, req)
	if got, want := errors.Cause(err), ErrMissingServiceID; got != want {
		t.Errorf("s.Stop(ctx, req): error = %q want %q", got, want)
	}

	mgr.EXPECT().Stop("404").Return(ErrNotFound).Times(1)

	req = &pb.StopReq{Id: "404"}
	_, err = s.Stop(ctx, req)
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf("s.Stop(ctx, req): error = %q want %q", got, want)
	}
}

func TestGRPCServerStop_Prune(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mock_manager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serv := pb.Service{
		Id:     "fs",
		Status: pb.Service_STOPPED,
	}

	mgr.EXPECT().Stop("fs").Return(nil).Times(1)
	mgr.EXPECT().Proto("fs").Return(&serv, nil).Times(1)
	mgr.EXPECT().Prune().Times(1)

	req := &pb.StopReq{Id: "fs", Prune: true}
	res, err := s.Stop(ctx, req)
	if err != nil {
		t.Errorf("s.Stop(ctx, req): error: %s", err)
	}

	if got, want := res.Id, serv.Id; got != want {
		t.Errorf("res.Id = %q want %q", got, want)
	}

	if got, want := res.Status, serv.Status; got != want {
		t.Errorf("res.Status = %q want %q", got, want)
	}
}

func TestGRPCServerPrune(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mock_manager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	list := []pb.Service{{
		Id:     "crypto",
		Status: pb.Service_STOPPED,
	}, {
		Id:     "fs",
		Status: pb.Service_RUNNING,
	}}

	mgr.EXPECT().Prune().Times(1)
	mgr.EXPECT().List().Return([]string{"crypto", "fs"}).Times(1)
	mgr.EXPECT().Proto("crypto").Return(&list[0], nil).Times(1)
	mgr.EXPECT().Proto("fs").Return(&list[1], nil).Times(1)

	req, ss := &pb.PruneReq{}, &mockSS{}
	if err := s.Prune(req, ss); err != nil {
		t.Errorf("s.Prune(req, ss): error: %s", err)
	}

	for i, serv := range list {
		res := ss.Res[i]

		if got, want := res.Id, serv.Id; got != want {
			t.Errorf("res.Id = %q want %q", got, want)
		}

		if got, want := res.Status, serv.Status; got != want {
			t.Errorf("res.Id = %q want %q", got, want)
		}
	}
}
