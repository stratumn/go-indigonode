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

package benchmark

import (
	"context"
	"io"
	"testing"

	ping "github.com/stratumn/go-indigonode/core/app/ping/grpc"
	"github.com/stratumn/go-indigonode/test/session"
	"google.golang.org/grpc"
)

func BenchmarkPing(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	tester := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := ping.NewPingClient(conns[0])
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ssCtx, cancelSS := context.WithCancel(ctx)
			ss, err := client.Ping(ssCtx, &ping.PingReq{
				PeerId: []byte(set[1].PeerID()),
				Times:  1,
			})

			_, err = ss.Recv()
			if err != nil {
				b.Errorf("s.Recv(): error: %+v", err)
			}

			cancelSS()
		}
	}

	err := session.Run(ctx, SessionDir, 2, session.BenchmarkCfg(), tester)
	if err != nil {
		b.Errorf("Session(): error: %+v", err)
	}
}

func BenchmarkPing_times(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	tester := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := ping.NewPingClient(conns[0])
		b.ResetTimer()

		ssCtx, cancelSS := context.WithCancel(ctx)
		defer cancelSS()

		ss, err := client.Ping(ssCtx, &ping.PingReq{
			PeerId: []byte(set[1].PeerID()),
			Times:  uint32(b.N),
		})

		for {
			_, err = ss.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				b.Errorf("s.Recv(): error: %+v", err)
			}
		}
	}

	err := session.Run(ctx, SessionDir, 2, session.BenchmarkCfg(), tester)
	if err != nil {
		b.Errorf("Session(): error: %+v", err)
	}
}

func BenchmarkPing_parallel(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	tester := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := ping.NewPingClient(conns[0])
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ssCtx, cancelSS := context.WithCancel(ctx)
				ss, err := client.Ping(ssCtx, &ping.PingReq{
					PeerId: []byte(set[1].PeerID()),
					Times:  1,
				})

				_, err = ss.Recv()
				if err != nil {
					b.Errorf("s.Recv(): error: %+v", err)
				}

				cancelSS()
			}
		})
	}

	err := session.Run(ctx, SessionDir, 2, session.BenchmarkCfg(), tester)
	if err != nil {
		b.Errorf("Session(): error: %+v", err)
	}
}

func BenchmarkPing_times100_parallel(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	tester := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := ping.NewPingClient(conns[0])
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ssCtx, cancelSS := context.WithCancel(ctx)
				ss, err := client.Ping(ssCtx, &ping.PingReq{
					PeerId: []byte(set[1].PeerID()),
					Times:  100,
				})

				for {
					_, err = ss.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						b.Errorf("s.Recv(): error: %+v", err)
					}
				}

				// So we get results per ping.
				for i := 0; i < 99; i++ {
					p.Next()
				}

				cancelSS()
			}
		})
	}

	err := session.Run(ctx, SessionDir, 2, session.BenchmarkCfg(), tester)
	if err != nil {
		b.Errorf("Session(): error: %+v", err)
	}
}
