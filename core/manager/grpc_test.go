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

	"github.com/pkg/errors"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr := createTestMgr(ctx, t)
	defer mgr.StopAll()
	s := grpcServer{mgr}

	req, ss := &pb.ListReq{}, &mockSS{}
	if err := s.List(req, ss); err != nil {
		t.Errorf("s.List(req, ss): error: %s", err)
	}

	list := mgr.List()
	if got, want := len(ss.Res), len(list); got != want {
		t.Errorf("len(ss.Res) = %d want %d", got, want)
	}

	for i, servID := range list {
		res := ss.Res[i]

		if got, want := res.Id, servID; got != want {
			t.Errorf("res.Id = %q want %q", got, want)
		}
	}
}

func TestGRPCServerStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr := createTestMgr(ctx, t)
	defer mgr.StopAll()
	s := grpcServer{mgr}

	req := &pb.StartReq{Id: "fs"}
	res, err := s.Start(ctx, req)
	if err != nil {
		t.Errorf("s.Start(ctx, req): error: %s", err)
	}

	if got, want := res.Id, "fs"; got != want {
		t.Errorf("res.Id = %q want %q", got, want)
	}

	if got, want := res.Status, pb.Service_RUNNING; got != want {
		t.Errorf("res.Status = %d want %d", got, want)
	}

	req = &pb.StartReq{Id: "404"}
	_, err = s.Start(ctx, req)
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf("s.Start(ctx, req): error = %q want %q", got, want)
	}
}

func TestGRPCServerStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr := createTestMgr(ctx, t)
	defer mgr.StopAll()
	s := grpcServer{mgr}

	if err := mgr.Start("fs"); err != nil {
		t.Fatalf(`mgr.Start("fs"): error: %s`, err)
	}

	req := &pb.StopReq{Id: "fs"}
	res, err := s.Stop(ctx, req)
	if err != nil {
		t.Errorf("s.Stop(ctx, req): error: %s", err)
	}

	if got, want := res.Id, "fs"; got != want {
		t.Errorf("res.Id = %q want %q", got, want)
	}

	if got, want := res.Status, pb.Service_STOPPED; got != want {
		t.Errorf("res.Status = %d want %d", got, want)
	}

	req = &pb.StopReq{Id: "404"}
	_, err = s.Stop(ctx, req)
	if got, want := errors.Cause(err), ErrNotFound; got != want {
		t.Errorf("s.Stop(ctx, req): error = %q want %q", got, want)
	}
}

func TestGRPCServerStop_Prune(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr := createTestMgr(ctx, t)
	defer mgr.StopAll()
	s := grpcServer{mgr}

	if err := mgr.Start("api"); err != nil {
		t.Fatalf(`mgr.Start("api"): error: %s`, err)
	}

	status, err := mgr.Status("crypto")
	if err != nil {
		t.Errorf(`mgr.Status("crypto"): error: %s`, err)
	}

	if got, want := status, Stopped; got != want {
		t.Errorf("status = %q want %q", got, want)
	}

	req := &pb.StopReq{Id: "api", Prune: true}
	if _, err := s.Stop(ctx, req); err != nil {
		t.Errorf("s.Stop(ctx, req): error: %s", err)
	}

	status, err = mgr.Status("crypto")
	if err != nil {
		t.Errorf(`mgr.Status("crypto"): error: %s`, err)
	}

	if got, want := status, Stopped; got != want {
		t.Errorf("status = %q want %q", got, want)
	}
}

func TestGRPCServerPrune(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr := createTestMgr(ctx, t)
	defer mgr.StopAll()
	s := grpcServer{mgr}

	if err := mgr.Start("api"); err != nil {
		t.Fatalf(`mgr.Start("api"): error: %s`, err)
	}

	status, err := mgr.Status("crypto")
	if err != nil {
		t.Errorf(`mgr.Status("crypto"): error: %s`, err)
	}

	if got, want := status, Stopped; got != want {
		t.Errorf("status = %q want %q", got, want)
	}

	if err := mgr.Stop("api"); err != nil {
		t.Fatalf(`mgr.Stop("api"): error: %s`, err)
	}

	req, ss := &pb.PruneReq{}, &mockSS{}
	if err := s.Prune(req, ss); err != nil {
		t.Errorf("s.Prune(req, ss): error: %s", err)
	}

	status, err = mgr.Status("crypto")
	if err != nil {
		t.Errorf(`mgr.Status("crypto"): error: %s`, err)
	}

	if got, want := status, Stopped; got != want {
		t.Errorf("status = %q want %q", got, want)
	}
}
