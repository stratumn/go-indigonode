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

package fossilizer

import (
	"context"
	"encoding/json"
	"testing"

	rpcpb "github.com/stratumn/go-indigonode/app/indigo/grpc/fossilizer"
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
