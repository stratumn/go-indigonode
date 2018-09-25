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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	pb "github.com/stratumn/go-node/app/chat/grpc"
	mockpb "github.com/stratumn/go-node/app/chat/grpc/mockchat"
	"github.com/stretchr/testify/assert"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	pstore "gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
)

var testPID peer.ID

func init() {
	pid, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9")
	if err != nil {
		panic(err)
	}
	testPID = pid
}

type grpcTestCase struct {
	name     string
	connect  func(context.Context, pstore.PeerInfo) error
	send     func(context.Context, peer.ID, string) error
	validate func(error)
}

func TestGRPCServer(t *testing.T) {
	testCases := []grpcTestCase{{
		"connect-error",
		func(_ context.Context, _ pstore.PeerInfo) error {
			return errors.New("Connect()")
		},
		func(_ context.Context, _ peer.ID, _ string) error {
			assert.Fail(t, "Send()")
			return nil
		},
		func(err error) {
			assert.Error(t, err)
			assert.Equal(t, "Connect()", err.Error())
		},
	}, {
		"send-error",
		func(_ context.Context, _ pstore.PeerInfo) error {
			return nil
		},
		func(_ context.Context, _ peer.ID, _ string) error {
			return errors.New("Send()")
		},
		func(err error) {
			assert.Error(t, err)
			assert.Equal(t, "Send()", err.Error())
		},
	}, {
		"send-success",
		func(_ context.Context, pi pstore.PeerInfo) error {
			assert.Equal(t, testPID, pi.ID, "PID")
			return nil
		},
		func(_ context.Context, pid peer.ID, message string) error {
			assert.Equal(t, testPID, pid, "PID")
			assert.Equal(t, "hello", message, "message")
			return nil
		},
		func(err error) {
			assert.NoError(t, err)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			server := &grpcServer{Connect: tt.connect, Send: tt.send}
			_, err := server.Message(
				context.Background(),
				&pb.ChatMessage{
					PeerId:  []byte(testPID),
					Message: "hello",
				})
			tt.validate(err)
		})
	}
}

func TestGRPCServer_GetHistory(t *testing.T) {
	msg := pb.NewDatedMessageReceived(testPID, "Test")

	server := &grpcServer{GetPeerHistory: func(peerID peer.ID) (ph PeerHistory, err error) {
		ph = append(ph, *msg)
		return
	}}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.HistoryReq{PeerId: []byte(testPID)}, mockpb.NewMockChat_GetHistoryServer(ctrl)

	ss.EXPECT().Send(msg)

	assert.NoError(t, server.GetHistory(req, ss))
}
