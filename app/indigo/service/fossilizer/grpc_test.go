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

package fossilizer

import (
	"context"
	"encoding/json"
	"testing"

	rpcpb "github.com/stratumn/alice/app/indigo/grpc/fossilizer"
	"github.com/stretchr/testify/assert"
)

func TestGRPCServer_GetInfo(t *testing.T) {

	t.Run("returns valid information", func(t *testing.T) {
		info := "test"
		server := &grpcServer{
			DoGetInfo: func(ctx context.Context) (interface{}, error) {
				return info, nil
			},
		}
		infoResp, err := server.GetInfo(context.Background(), nil)
		assert.NoError(t, err, "server.GetInfo()")
		var got string
		err = json.Unmarshal(infoResp.Data, &got)
		assert.NoError(t, err)
		assert.Equal(t, got, info)
	})
}

func TestGRPCServer_Fossilize(t *testing.T) {

	t.Run("invalid argument", func(t *testing.T) {
		server := &grpcServer{
			DoFossilize: func(ctx context.Context, data, meta []byte) error {
				return nil
			},
		}
		_, err := server.Fossilize(context.Background(), nil)
		assert.EqualError(t, err, ErrInvalidArgument.Error(), "server.Fossilize()")
	})

	t.Run("valid argument", func(t *testing.T) {
		server := &grpcServer{
			DoFossilize: func(ctx context.Context, data, meta []byte) error {
				return nil
			},
		}
		_, err := server.Fossilize(context.Background(), &rpcpb.FossilizeReq{})
		assert.NoError(t, err, "server.Fossilize()")
	})
}
