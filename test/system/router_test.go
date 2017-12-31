// Copyright © 2017 Stratumn SAS
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
	"io"
	"testing"
	"time"

	"github.com/stratumn/alice/grpc/host"
	"github.com/stratumn/alice/test/session"
	"google.golang.org/grpc"

	ma "gx/ipfs/QmW8s4zTsUoX1Q6CeYxVKPyqSKbF7H1YDUyTostBtZ8DaG/go-multiaddr"
)

// TestRouter_Connect checks that the DHT router is working.
func TestRouter_Connect(t *testing.T) {
	Test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := host.NewHostClient(conns[0])

		for _, node := range set[1:] {
			address := "/ipfs/" + node.PeerID().Pretty()
			addr, err := ma.NewMultiaddr(address)
			if err != nil {
				t.Errorf("ma.NewMultiaddr(%q): error: %+v", address, err)
				continue
			}

			req := host.ConnectReq{Address: addr.Bytes()}

		TEST_ROUTE_CONNECT_LOOP:
			for {
				ss, err := client.Connect(ctx, &req)
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}
				var msgs []*host.Connection

				for {
					msg, err := ss.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						time.Sleep(5 * time.Second)
						continue TEST_ROUTE_CONNECT_LOOP
					}
					msgs = append(msgs, msg)
				}

				if err != nil {
					t.Errorf("ma.NewMultiaddr(%q): error: %+v", msgs[0].RemoteAddress, err)
					continue
				}

				break
			}
		}
	})

}
