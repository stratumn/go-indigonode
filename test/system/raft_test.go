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

package system

import (
	"context"
	// "sync"
	"testing"
	"time"

	"github.com/stratumn/alice/grpc/raft"
	"github.com/stratumn/alice/test/session"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const numNodes = 4               // please change thoughtfully
const waitTime = 2 * time.Second // increase if test stucks

// TestRaft executes a standart raft test scenario.

func assertNodes(t *testing.T, ctx context.Context,
	nn []raft.RaftClient,
	ss []raft.StatusInfo,
	pp [][]raft.Peer,
	ee [][]raft.Entry,
) {

	assert.Equal(t, len(nn), numNodes, "node array does not match numNodes, check your tests")
	assert.Equal(t, len(ss), numNodes, "status array does not match numNodes, check your tests")
	assert.Equal(t, len(pp), numNodes, "peer array does not match numNodes, check your tests")
	assert.Equal(t, len(ee), numNodes, "log array does not match numNodes, check your tests")

	// TODO: test logs with timestamps
	for i := 0; i < numNodes; i++ {
		s, err := nn[i].Status(ctx, &raft.Empty{})
		assert.NoError(t, err)
		assert.Equalf(t, ss[i], *s, "node %d status does not match", i+1)

		ppStream, err := nn[i].Peers(ctx, &raft.Empty{})
		assert.NoError(t, err)

		p := []raft.Peer{}
		for {
			peer, err := ppStream.Recv()
			if err != nil {
				break
			}
			p = append(p, *peer)
		}

		assert.Equalf(t, pp[i], p, "node %d peer list does not match", i+1)

		eeStream, err := nn[i].Log(ctx, &raft.Empty{})
		assert.NoError(t, err)

		e := []raft.Entry{}
		for {
			entry, err := eeStream.Recv()
			if err != nil {
				break
			}
			e = append(e, *entry)
		}

		assert.Equalf(t, ee[i], e, "node %d log does not match", i+1)

	}

}

// Before running the test reduce Maximum segment lifetime
// On OSX: sudo sysctl net.inet.tcp.msl=500 will set MSL to 500ms
func TestRaft(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	err := session.Run(ctx, SessionDir, numNodes, session.SystemCfg(),
		func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
			assert.Equal(t, numNodes, 4, "numNodes should be changed thoughtfully")

			n1 := raft.NewRaftClient(conns[0])
			n2 := raft.NewRaftClient(conns[1])
			n3 := raft.NewRaftClient(conns[2])
			n4 := raft.NewRaftClient(conns[3])

			a1 := []byte(set[0].PeerID())
			a2 := []byte(set[1].PeerID())
			a3 := []byte(set[2].PeerID())
			a4 := []byte(set[3].PeerID())

			e1 := []byte("\xBA\xDD\xCA\xFE")
			e2 := []byte("\xDE\xAD\xBE\xEF")
			e3 := []byte("\xFE\xED\xFA\xCE")

			var err error

			nn := []raft.RaftClient{n1, n2, n3, n4}
			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
				},
				[][]raft.Peer{
					[]raft.Peer{},
					[]raft.Peer{},
					[]raft.Peer{},
					[]raft.Peer{},
				},
				[][]raft.Entry{
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
				},
			)

			_, err = n1.Start(ctx, &raft.Empty{})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 1},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
					},
					[]raft.Peer{},
					[]raft.Peer{},
					[]raft.Peer{},
				},
				[][]raft.Entry{
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
				},
			)

			_, err = n1.Invite(ctx, &raft.PeerID{Address: a2})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 1},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
					},
					[]raft.Peer{},
					[]raft.Peer{},
					[]raft.Peer{},
				},
				[][]raft.Entry{
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
				},
			)

			_, err = n2.Join(ctx, &raft.PeerID{Address: a1})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 1},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
					},
					[]raft.Peer{},
					[]raft.Peer{},
				},
				[][]raft.Entry{
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
				},
			)

			_, err = n2.Invite(ctx, &raft.PeerID{Address: a3})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 1},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: false, Id: 0},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{},
					[]raft.Peer{},
				},
				[][]raft.Entry{
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
				},
			)

			_, err = n3.Join(ctx, &raft.PeerID{Address: a1})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 1},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: true, Id: 4},
					raft.StatusInfo{Running: false, Id: 0},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{},
				},
				[][]raft.Entry{
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
					[]raft.Entry{},
				},
			)

			_, err = n2.Propose(ctx, &raft.Proposal{Data: e1})
			assert.NoError(t, err)
			time.Sleep(waitTime) // Make sure e1 is added before e2
			_, err = n3.Propose(ctx, &raft.Proposal{Data: e2})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 1},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: true, Id: 4},
					raft.StatusInfo{Running: false, Id: 0},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
					},
					[]raft.Peer{},
				},
				[][]raft.Entry{
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{},
				},
			)

			_, err = n3.Invite(ctx, &raft.PeerID{Address: a4})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			_, err = n4.Join(ctx, &raft.PeerID{Address: a2})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 1},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: true, Id: 4},
					raft.StatusInfo{Running: true, Id: 5},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
					},
					[]raft.Peer{
						raft.Peer{Id: 1, Address: a1},
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
					},
				},
				[][]raft.Entry{
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
				},
			)

			_, err = n2.Expel(ctx, &raft.PeerID{Address: a1})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: false, Id: 0},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: true, Id: 4},
					raft.StatusInfo{Running: true, Id: 5},
				},
				[][]raft.Peer{
					[]raft.Peer{},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
					},
				},
				[][]raft.Entry{
					[]raft.Entry{},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
					},
				},
			)

			dPeersStream, err := n1.Discover(ctx, &raft.PeerID{Address: a4})
			assert.NoError(t, err)

			dPeers := []raft.Peer{}
			for {
				peer, err := dPeersStream.Recv()
				if err != nil {
					break
				}
				dPeers = append(dPeers, *peer)
			}

			assert.Equalf(t, []raft.Peer{
				raft.Peer{Id: 3, Address: a2},
				raft.Peer{Id: 4, Address: a3},
				raft.Peer{Id: 5, Address: a4},
			}, dPeers, "discovered peer list does not match")

			_, err = n2.Propose(ctx, &raft.Proposal{Data: e3})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			_, err = n4.Invite(ctx, &raft.PeerID{Address: a1})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			_, err = n1.Join(ctx, &raft.PeerID{Address: a2})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 7},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: true, Id: 4},
					raft.StatusInfo{Running: true, Id: 5},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 4, Address: a3},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
					},
				},
				[][]raft.Entry{
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
				},
			)

			_, err = n3.Stop(ctx, &raft.Empty{})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			_, err = n2.Expel(ctx, &raft.PeerID{Address: a3})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			_, err = n4.Invite(ctx, &raft.PeerID{Address: a3})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			_, err = n3.Join(ctx, &raft.PeerID{Address: a1})
			assert.NoError(t, err)
			time.Sleep(waitTime)

			assertNodes(t, ctx, nn,
				[]raft.StatusInfo{
					raft.StatusInfo{Running: true, Id: 7},
					raft.StatusInfo{Running: true, Id: 3},
					raft.StatusInfo{Running: true, Id: 9},
					raft.StatusInfo{Running: true, Id: 5},
				},
				[][]raft.Peer{
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
						raft.Peer{Id: 9, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
						raft.Peer{Id: 9, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
						raft.Peer{Id: 9, Address: a3},
					},
					[]raft.Peer{
						raft.Peer{Id: 3, Address: a2},
						raft.Peer{Id: 5, Address: a4},
						raft.Peer{Id: 7, Address: a1},
						raft.Peer{Id: 9, Address: a3},
					},
				},
				[][]raft.Entry{
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
					[]raft.Entry{
						raft.Entry{Index: 0, Data: e1},
						raft.Entry{Index: 1, Data: e2},
						raft.Entry{Index: 2, Data: e3},
					},
				},
			)

		})

	assert.NoError(t, err, "Session()")

}
