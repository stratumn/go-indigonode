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

	grpcapi "github.com/stratumn/go-indigonode/core/app/grpcapi/grpc"
	manager "github.com/stratumn/go-indigonode/core/manager/grpc"
	"github.com/stratumn/go-indigonode/test/session"
	"google.golang.org/grpc"
)

func BenchmarkAPI_Inform(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	tester := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := grpcapi.NewAPIClient(conns[0])
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				_, err := client.Inform(ctx, &grpcapi.InformReq{})
				if err != nil {
					b.Errorf("c.Inform(): error: %+v", err)
				}
			}
		})
	}

	config := session.WithServices(session.BenchmarkCfg(), "grpcapi")

	err := session.Run(ctx, SessionDir, 1, config, tester)
	if err != nil {
		b.Errorf("Session(): error: %+v", err)
	}
}

func BenchmarkAPI_ListServices(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), MaxDuration)
	defer cancel()

	tester := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		client := manager.NewManagerClient(conns[0])

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ss, err := client.List(ctx, &manager.ListReq{})
				if err != nil {
					b.Errorf("c.ListServices(): error: %+v", err)
				}
				for {
					_, err = ss.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						b.Errorf("s.Recv(): error: %+v", err)
					}
				}
			}
		})
	}

	config := session.WithServices(session.BenchmarkCfg(), "grpcapi")

	err := session.Run(ctx, SessionDir, 1, config, tester)
	if err != nil {
		b.Errorf("Session(): error: %+v", err)
	}
}
