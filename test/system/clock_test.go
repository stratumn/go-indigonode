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
	"testing"
	"time"

	"github.com/stratumn/alice/grpc/clock"
	"github.com/stratumn/alice/test/session"
	"google.golang.org/grpc"
)

func TestClock(t *testing.T) {
	Test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := clock.NewClockClient(conns[0])

		for _, node := range set[1:] {
			req := clock.RemoteReq{PeerId: []byte(node.PeerID())}

			// It might take a few attempts before the router can
			// connect to the peer.
			for {
				res, err := client.Remote(ctx, &req)
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}

				ts := time.Unix(0, res.Timestamp)
				if time.Now().UTC().Sub(ts) > time.Second {
					t.Errorf("client.Remote(ctx, &req): unexpected time: %v", ts)
				}

				break
			}
		}
	})
}
