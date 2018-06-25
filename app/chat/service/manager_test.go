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

package service

import (
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"

	"github.com/stretchr/testify/assert"

	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pb "github.com/stratumn/go-indigonode/app/chat/grpc"
)

var testPID1 peer.ID
var testPID2 peer.ID

func init() {
	pid, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9")
	if err != nil {
		panic(err)
	}
	testPID1 = pid

	pid, err = peer.IDB58Decode("Qmcso1m2v6r9jZv8swVD7LCDPNzRoX8RohHWYqG8dP8cxr")
	if err != nil {
		panic(err)
	}
	testPID2 = pid
}

func TestManager(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	dbPath := filepath.Join(dir, "chat-history.db")

	mgr, err := NewManager(dbPath)
	require.NoError(t, err)
	require.NotNil(t, mgr)

	t.Run("Add a message received to history", func(t *testing.T) {
		receivedMsg := pb.NewDatedMessageReceived(testPID1, "Message received")
		err := mgr.Add(testPID1, receivedMsg)
		require.NoError(t, err)
	})

	t.Run("Add a message sent to history", func(t *testing.T) {
		sentMsg := pb.NewDatedMessageSent(testPID1, "Message sent")
		err := mgr.Add(testPID1, sentMsg)
		require.NoError(t, err)
	})

	t.Run("Retrieve peer history", func(t *testing.T) {
		ph, err := mgr.Get(testPID1)
		require.NoError(t, err)
		require.Len(t, ph, 2)

		msgReceived := ph[0]

		assert.Equal(t, "Message received", msgReceived.Content)
		assert.Equal(t, []byte(testPID1), msgReceived.From)
		assert.WithinDuration(t, time.Now(), time.Unix(msgReceived.Time.Seconds, 0), 1*time.Second)

		msgSent := ph[1]

		assert.Equal(t, "Message sent", msgSent.Content)
		assert.Equal(t, []byte(testPID1), msgSent.To)
		assert.WithinDuration(t, time.Now(), time.Unix(msgSent.Time.Seconds, 0), 1*time.Second)
	})

	t.Run("Retrieve empty peer history", func(t *testing.T) {
		ph, err := mgr.Get(testPID2)
		require.NoError(t, err)
		require.Len(t, ph, 0)
	})
}
