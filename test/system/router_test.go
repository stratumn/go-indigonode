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

package system

import (
	"context"
	"io"
	"testing"
	"time"

	host "github.com/stratumn/go-indigonode/core/app/host/grpc"
	"github.com/stratumn/go-indigonode/test/session"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
)

// TestRouter_Connect checks that the DHT router is working.
func TestRouter_Connect(t *testing.T) {
	Test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := host.NewHostClient(conns[0])

		for _, node := range set[1:] {
			address := "/ipfs/" + node.PeerID().Pretty()
			addr, err := ma.NewMultiaddr(address)
			if err != nil {
				assert.NoError(t, err, "ma.NewMultiaddr(address)")
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
					assert.NoError(t, err, "ma.NewMultiaddr(address)")
					continue
				}

				break
			}
		}
	})

}
