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

package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-node/app/indigo/pb/store"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
)

func TestGRPCServer_CreateLink(t *testing.T) {
	link := cstesting.RandomLink()
	linkHash, _ := link.Hash()
	linkBytes, _ := json.Marshal(link)
	rpcLink := &pb.Link{Data: linkBytes}

	t.Run("missing-input", func(t *testing.T) {
		server := &grpcServer{}
		_, err := server.CreateLink(context.Background(), nil)
		assert.EqualError(t, err, pb.ErrInvalidArgument.Error(), "server.CreateLink()")
	})

	t.Run("non-json-input", func(t *testing.T) {
		server := &grpcServer{}
		_, err := server.CreateLink(context.Background(), &pb.Link{Data: []byte("{iAm_m4lf0rm3D")})
		assert.EqualError(t, err, pb.ErrInvalidArgument.Error(), "server.CreateLink()")
	})

	t.Run("store-error", func(t *testing.T) {
		storeErr := errors.New("pwn3d")
		server := &grpcServer{
			DoCreateLink: func(_ context.Context, l *cs.Link) (*types.Bytes32, error) {
				assert.Equal(t, link, l, "cs.Link")
				return nil, storeErr
			},
		}

		_, err := server.CreateLink(context.Background(), rpcLink)
		assert.EqualError(t, err, storeErr.Error(), "server.CreateLink()")
	})

	t.Run("store-success", func(t *testing.T) {
		server := &grpcServer{
			DoCreateLink: func(_ context.Context, l *cs.Link) (*types.Bytes32, error) {
				assert.Equal(t, link, l, "cs.Link")
				return linkHash, nil
			},
		}

		lh, err := server.CreateLink(context.Background(), rpcLink)
		assert.NoError(t, err, "server.CreateLink()")
		assert.Equal(t, linkHash[:], lh.Data)
	})
}

func TestGRPCServer_GetSegment(t *testing.T) {
	link := cstesting.RandomLink()
	linkHash, _ := link.Hash()
	rpcLinkHash := &pb.LinkHash{Data: linkHash[:]}

	t.Run("missing-input", func(t *testing.T) {
		server := &grpcServer{}
		_, err := server.GetSegment(context.Background(), nil)
		assert.EqualError(t, err, pb.ErrInvalidArgument.Error(), "server.GetSegment()")
	})

	t.Run("store-error", func(t *testing.T) {
		storeErr := errors.New("pwn3d")
		server := &grpcServer{
			DoGetSegment: func(_ context.Context, lh *types.Bytes32) (*cs.Segment, error) {
				assert.Equal(t, linkHash, lh, "*types.Bytes32")
				return nil, storeErr
			},
		}

		_, err := server.GetSegment(context.Background(), rpcLinkHash)
		assert.EqualError(t, err, storeErr.Error(), "server.GetSegment()")
	})

	t.Run("store-not-found", func(t *testing.T) {
		server := &grpcServer{
			DoGetSegment: func(_ context.Context, lh *types.Bytes32) (*cs.Segment, error) {
				assert.Equal(t, linkHash, lh, "*types.Bytes32")
				return nil, nil
			},
		}

		_, err := server.GetSegment(context.Background(), rpcLinkHash)
		assert.EqualError(t, err, ErrNotFound.Error(), "server.GetSegment()")
	})

	t.Run("store-success", func(t *testing.T) {
		server := &grpcServer{
			DoGetSegment: func(_ context.Context, lh *types.Bytes32) (*cs.Segment, error) {
				assert.Equal(t, linkHash, lh, "*types.Bytes32")
				return link.Segmentify(), nil
			},
		}

		resp, err := server.GetSegment(context.Background(), rpcLinkHash)
		assert.NoError(t, err, "server.GetSegment()")

		var seg cs.Segment
		err = json.Unmarshal(resp.Data, &seg)
		assert.NoError(t, err, "json.Unmarshal()")

		assert.Equal(t, link, &seg.Link)
	})
}
