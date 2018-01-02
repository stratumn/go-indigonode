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

	"github.com/stratumn/alice/grpc/grpcapi"
	"github.com/stratumn/alice/release"
	"github.com/stratumn/alice/test/session"
	"google.golang.org/grpc"
)

func TestGrcpAPI_Inform(t *testing.T) {
	Test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
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
