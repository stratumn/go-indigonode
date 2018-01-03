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

package metrics

import (
	"context"
	"testing"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p2pmetrics "gx/ipfs/QmaL2WYJGbWKqHoLujoi9GQ5jj4JVFrBqHUBWmEYzJPVWT/go-libp2p-metrics"
)

func testGRPCServer() grpcServer {
	mtrx := newMetrics(
		p2pmetrics.NewBandwidthCounter(),
		metrics.NewInmemSink(time.Second, time.Second),
	)
	return grpcServer{func() *Metrics { return mtrx }}
}

func testGRPCServerUnavailable() grpcServer {
	return grpcServer{func() *Metrics { return nil }}
}
func TestGRPCServer_Bandwidth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer()

	srv.GetMetrics().LogRecvMessage(1000)

	req := &pb.BandwidthReq{}
	res, err := srv.Bandwidth(ctx, req)

	require.NoError(t, err)
	require.Equal(t, uint64(1000), res.TotalIn)
}

func TestGRPCServer_Bandwidth_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.BandwidthReq{}
	_, err := srv.Bandwidth(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}
