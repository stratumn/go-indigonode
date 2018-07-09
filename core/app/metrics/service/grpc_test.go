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

package service

import (
	"context"
	"testing"

	pb "github.com/stratumn/go-indigonode/core/app/metrics/grpc"
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
