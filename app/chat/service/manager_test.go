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
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"

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
