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
	"testing"
	"time"

	clock "github.com/stratumn/go-indigonode/app/clock/grpc"
	"github.com/stratumn/go-indigonode/test/session"
	"github.com/stratumn/go-indigonode/test/system"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestClock(t *testing.T) {
	system.Test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
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
				assert.WithinDuration(t, time.Now().UTC(), ts, time.Second, "client.Remote(ctx, &req): unexpected time")

				break
			}
		}
	})
}
