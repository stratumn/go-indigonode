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
	mock_pb "github.com/stratumn/alice/grpc/manager/mock_manager"
)

func TestGRPCServerList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mock_manager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	serv1 := &pb.Service{Id: "serv1", Status: pb.Service_STOPPED}
	serv2 := &pb.Service{Id: "serv2", Status: pb.Service_RUNNING}

	req, ss := &pb.ListReq{}, mock_pb.NewMockManager_ListServer(ctrl)

	mgr.EXPECT().List().Return([]string{"serv1", "serv2"}).Times(1)
	mgr.EXPECT().Proto("serv1").Return(serv1, nil).Times(1)
	mgr.EXPECT().Proto("serv2").Return(serv2, nil).Times(1)

	ss.EXPECT().Send(serv1).Times(1)
	ss.EXPECT().Send(serv2).Times(1)

	if err := s.List(req, ss); err != nil {
		t.Errorf("s.List(req, ss): error: %s", err)
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
		Id:     "serv1",
		Status: pb.Service_RUNNING,
	}

	mgr.EXPECT().Start("serv1").Return(nil).Times(1)
	mgr.EXPECT().Proto("serv1").Return(&serv, nil).Times(1)

	req := &pb.StartReq{Id: "serv1"}
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

	mgr.EXPECT().Start("serv2").Return(ErrNotFound).Times(1)

	req = &pb.StartReq{Id: "serv2"}
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
		Id:     "serv1",
		Status: pb.Service_STOPPED,
	}

	mgr.EXPECT().Stop("serv1").Return(nil).Times(1)
	mgr.EXPECT().Proto("serv1").Return(&serv, nil).Times(1)
	mgr.EXPECT().Prune().Times(0)

	req := &pb.StopReq{Id: "serv1"}
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

	mgr.EXPECT().Stop("serv2").Return(ErrNotFound).Times(1)

	req = &pb.StopReq{Id: "serv2"}
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
		Id:     "serv1",
		Status: pb.Service_STOPPED,
	}

	mgr.EXPECT().Stop("serv1").Return(nil).Times(1)
	mgr.EXPECT().Proto("serv1").Return(&serv, nil).Times(1)
	mgr.EXPECT().Prune().Times(1)

	req := &pb.StopReq{Id: "serv1", Prune: true}
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serv1 := &pb.Service{Id: "serv1", Status: pb.Service_STOPPED}
	serv2 := &pb.Service{Id: "serv2", Status: pb.Service_STOPPED}

	req, ss := &pb.PruneReq{}, mock_pb.NewMockManager_ListServer(ctrl)

	mgr.EXPECT().Prune().Times(1)
	mgr.EXPECT().List().Return([]string{"serv1", "serv2"}).Times(1)
	mgr.EXPECT().Proto("serv1").Return(serv1, nil).Times(1)
	mgr.EXPECT().Proto("serv2").Return(serv2, nil).Times(1)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(serv1).Times(1)
	ss.EXPECT().Send(serv2).Times(1)

	if err := s.Prune(req, ss); err != nil {
		t.Errorf("s.Prune(req, ss): error: %s", err)
	}
}
