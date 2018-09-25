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

	grpcapi "github.com/stratumn/go-node/core/app/grpcapi/grpc"
	"github.com/stratumn/go-node/release"
	"github.com/stratumn/go-node/test/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
)

func TestGrcpAPI_Inform(t *testing.T) {
	Test(t, func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		session.PerCPU(func(i int) {
			client := grpcapi.NewAPIClient(conns[i])

			res, err := client.Inform(ctx, &grpcapi.InformReq{})
			require.NoError(t, err)

			assert.Equal(t, release.Protocol, res.Protocol)
		}, NumNodes)
	})
}
