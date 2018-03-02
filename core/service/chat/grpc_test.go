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

package chat

import (
	"context"
	"errors"
	"testing"

	pb "github.com/stratumn/alice/grpc/chat"
	"github.com/stretchr/testify/assert"

	pstore "gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
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
			server := &grpcServer{tt.connect, tt.send}
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
