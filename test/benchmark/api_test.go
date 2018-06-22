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

package benchmark

import (
	"context"
	"io"
	"testing"

	grpcapi "github.com/stratumn/alice/core/app/grpcapi/grpc"
	"github.com/stratumn/alice/grpc/manager"
	"github.com/stratumn/alice/test/session"
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
