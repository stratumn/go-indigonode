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

package service

import (
	"context"
	"testing"

	pb "github.com/stratumn/go-indigonode/core/app/monitoring/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCServer_SetSamplingRatio(t *testing.T) {
	ctx := context.Background()
	server := grpcServer{}

	t.Run("reject-invalid-ratio", func(t *testing.T) {
		_, err := server.SetSamplingRatio(ctx, &pb.SamplingRatio{Value: 4.2})
		assert.EqualError(t, err, ErrInvalidRatio.Error())
	})

	t.Run("set-trace-sampler", func(t *testing.T) {
		_, err := server.SetSamplingRatio(ctx, &pb.SamplingRatio{Value: 0.42})
		require.NoError(t, err)
	})
}
