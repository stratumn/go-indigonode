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

package service

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/app/contacts/grpc"
	mockpb "github.com/stratumn/go-indigonode/app/contacts/grpc/mockcontacts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGRPCServer(ctx context.Context, t *testing.T) grpcServer {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	mgr, err := NewManager(filepath.Join(dir, "contacts.toml"))
	require.NoError(t, err, "NewManager()")

	return grpcServer{func() *Manager { return mgr }}
}

func testGRPCServerUnavailable() grpcServer {
	return grpcServer{func() *Manager { return nil }}
}

func TestGRPCServer_List(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contact := &Contact{PeerID: testPID}

	srv := testGRPCServer(ctx, t)
	mgr := srv.GetManager()
	mgr.Set("alice", contact)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ListReq{}, mockpb.NewMockContacts_ListServer(ctrl)

	ss.EXPECT().Send(&pb.Contact{
		Name:   "alice",
		PeerId: []byte(testPID),
	})

	assert.NoError(t, srv.List(req, ss))
}

func TestGRPCServer_List_unavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := testGRPCServerUnavailable()

	req, ss := &pb.ListReq{}, mockpb.NewMockContacts_ListServer(ctrl)

	assert.Equal(t, ErrUnavailable, errors.Cause(srv.List(req, ss)))
}

func TestGRPCServer_Get(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contact := &Contact{PeerID: testPID}

	srv := testGRPCServer(ctx, t)
	mgr := srv.GetManager()
	mgr.Set("alice", contact)

	req := &pb.GetReq{Name: "alice"}
	res, err := srv.Get(ctx, req)
	require.NoError(t, err, "srv.Get(ctx, req)")

	assert.Equal(t, &pb.Contact{Name: "alice", PeerId: []byte(testPID)}, res)
}

func TestGRPCServer_Get_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.GetReq{}
	_, err := srv.Get(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}

func TestGRPCServer_Set(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contact := &Contact{PeerID: testPID}

	srv := testGRPCServer(ctx, t)

	req := &pb.SetReq{Name: "alice", PeerId: []byte(testPID)}
	res, err := srv.Set(ctx, req)
	require.NoError(t, err, "srv.Set(ctx, req)")

	assert.Equal(t, &pb.Contact{Name: "alice", PeerId: []byte(testPID)}, res)

	mgr := srv.GetManager()
	record, err := mgr.Get("alice")
	require.NoError(t, err, "mgr.Get()")

	assert.Equal(t, contact, record, "mgr.Get()")
}

func TestGRPCServer_Set_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.SetReq{}
	_, err := srv.Set(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}

func TestGRPCServer_Delete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contact := &Contact{PeerID: testPID}

	srv := testGRPCServer(ctx, t)
	mgr := srv.GetManager()
	mgr.Set("alice", contact)

	req := &pb.DeleteReq{Name: "alice"}
	res, err := srv.Delete(ctx, req)
	require.NoError(t, err, "srv.Delete(ctx, req)")

	assert.Equal(t, &pb.Contact{Name: "alice", PeerId: []byte(testPID)}, res)
}

func TestGRPCServer_Delete_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.DeleteReq{}
	_, err := srv.Delete(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}
