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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/manager/grpc"
	mockpb "github.com/stratumn/go-indigonode/core/manager/grpc/mockgrpc"
	"github.com/stratumn/go-indigonode/core/manager/mockmanager"
	"github.com/stretchr/testify/assert"
)

func TestGRPCServer_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockmanager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	serv1 := &pb.Service{Id: "serv1", Status: pb.Service_STOPPED}
	serv2 := &pb.Service{Id: "serv2", Status: pb.Service_RUNNING}

	req, ss := &pb.ListReq{}, mockpb.NewMockManager_ListServer(ctrl)

	mgr.EXPECT().List().Return([]string{"serv1", "serv2"}).Times(1)
	mgr.EXPECT().Proto("serv1").Return(serv1, nil).Times(1)
	mgr.EXPECT().Proto("serv2").Return(serv2, nil).Times(1)

	ss.EXPECT().Send(serv1).Times(1)
	ss.EXPECT().Send(serv2).Times(1)

	err := s.List(req, ss)
	assert.NoError(t, err, "s.List(req, ss)")
}

func TestGRPCServer_Info(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockmanager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serv := pb.Service{
		Id:     "serv1",
		Status: pb.Service_RUNNING,
	}

	mgr.EXPECT().Proto("serv1").Return(&serv, nil).Times(1)

	req := &pb.InfoReq{Id: "serv1"}
	res, err := s.Info(ctx, req)
	assert.NoError(t, err, "s.Info(ctx, req)")

	assert.Equal(t, serv.Id, res.Id, "res.Id")
	assert.Equal(t, serv.Status, res.Status, "res.Status")

	req = &pb.InfoReq{}
	_, err = s.Info(ctx, req)
	assert.Equal(t, ErrMissingServiceID, errors.Cause(err), "s.Info(ctx, req)")

	mgr.EXPECT().Proto("serv2").Return(nil, ErrNotFound).Times(1)

	req = &pb.InfoReq{Id: "serv2"}
	_, err = s.Info(ctx, req)
	assert.Equal(t, ErrNotFound, errors.Cause(err), "s.Info(ctx, req)")
}

func TestGRPCServer_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockmanager.NewMockGRPCManager(ctrl)
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
	assert.NoError(t, err, "s.Start(ctx, req)")

	assert.Equal(t, serv.Id, res.Id, "res.Id")
	assert.Equal(t, serv.Status, res.Status, "res.Status")

	req = &pb.StartReq{}
	_, err = s.Start(ctx, req)
	assert.Equal(t, ErrMissingServiceID, errors.Cause(err), "s.Start(ctx, req)")

	mgr.EXPECT().Start("serv2").Return(ErrNotFound).Times(1)

	req = &pb.StartReq{Id: "serv2"}
	_, err = s.Start(ctx, req)
	assert.Equal(t, ErrNotFound, errors.Cause(err), "s.Start(ctx, req)")
}

func TestGRPCServer_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockmanager.NewMockGRPCManager(ctrl)
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
	assert.NoError(t, err, "s.Stop(ctx, req)")

	assert.Equal(t, serv.Id, res.Id, "res.Id")
	assert.Equal(t, serv.Status, res.Status, "res.Status")

	req = &pb.StopReq{}
	_, err = s.Stop(ctx, req)
	assert.Equal(t, ErrMissingServiceID, errors.Cause(err), "s.Stop(ctx, req)")

	mgr.EXPECT().Stop("serv2").Return(ErrNotFound).Times(1)

	req = &pb.StopReq{Id: "serv2"}
	_, err = s.Stop(ctx, req)
	assert.Equal(t, ErrNotFound, errors.Cause(err), "s.Stop(ctx, req)")
}

func TestGRPCServer_Stop_Prune(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockmanager.NewMockGRPCManager(ctrl)
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
	assert.NoError(t, err, "s.Stop(ctx, req)")

	assert.Equal(t, serv.Id, res.Id, "res.Id")
	assert.Equal(t, serv.Status, res.Status, "res.Status")
}

func TestGRPCServer_Prune(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockmanager.NewMockGRPCManager(ctrl)
	s := grpcServer{mgr}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serv1 := &pb.Service{Id: "serv1", Status: pb.Service_STOPPED}
	serv2 := &pb.Service{Id: "serv2", Status: pb.Service_STOPPED}

	req, ss := &pb.PruneReq{}, mockpb.NewMockManager_ListServer(ctrl)

	mgr.EXPECT().Prune().Times(1)
	mgr.EXPECT().List().Return([]string{"serv1", "serv2"}).Times(1)
	mgr.EXPECT().Proto("serv1").Return(serv1, nil).Times(1)
	mgr.EXPECT().Proto("serv2").Return(serv2, nil).Times(1)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(serv1).Times(1)
	ss.EXPECT().Send(serv2).Times(1)

	err := s.Prune(req, ss)
	assert.NoError(t, err, "s.Prune(req, ss)")
}
