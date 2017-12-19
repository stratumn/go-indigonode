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
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stratumn/alice/grpc/grpcapi"
	"github.com/stratumn/alice/grpc/host"
	"github.com/stratumn/alice/release"
	"github.com/stratumn/alice/test/session"
	"google.golang.org/grpc"

	ma "gx/ipfs/QmW8s4zTsUoX1Q6CeYxVKPyqSKbF7H1YDUyTostBtZ8DaG/go-multiaddr"
)

// test wrap session.Session with a context and handles errors.
func test(t *testing.T, fn session.Tester) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	err := session.Run(ctx, SessionDir, NumNodes, session.SystemCfg(), fn)
	if err != nil {
		t.Errorf("Session(): error: %+v", err)
	}
}

func TestGrcpAPI_Inform(t *testing.T) {
	test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		session.PerCPU(func(i int) {
			client := grpcapi.NewAPIClient(conns[i])

			res, err := client.Inform(ctx, &grpcapi.InformReq{})
			if err != nil {
				t.Errorf("node %d: Inform(): error: %+v", i, err)
				return
			}

			if got, want := res.Protocol, release.Protocol; got != want {
				t.Errorf("node %d: r.Protocol = %q want %q", i, got, want)
			}
		}, NumNodes)
	})
}

// TestRouter_Connect checks that the DHT router is working.
func TestRouter_Connect(t *testing.T) {
	test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
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
				fmt.Fprintf(os.Stderr, "Connecting to %q...\n", address)
				ss, err := client.Connect(ctx, &req)
				if err == nil {
					var msgs []*host.Connection

					for {
						msg, err := ss.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failure: %s.\n", err)
							time.Sleep(5 * time.Second)
							continue TEST_ROUTE_CONNECT_LOOP
						}
						msgs = append(msgs, msg)
					}

					remoteAddr, err := ma.NewMultiaddrBytes(msgs[0].RemoteAddress)
					if err != nil {
						t.Errorf("ma.NewMultiaddr(%q): error: %+v", msgs[0].RemoteAddress, err)
						continue
					}
					fmt.Fprintf(os.Stderr, "Success: %s.\n", remoteAddr)
					break
				}
				fmt.Fprintf(os.Stderr, "Failure: %s.\n", err)
				time.Sleep(5 * time.Second)
			}
		}
	})

}
