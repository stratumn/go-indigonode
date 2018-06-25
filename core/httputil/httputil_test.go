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

package httputil

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigonode/core/httputil/mockhttputil"
	"github.com/stratumn/go-indigonode/core/netutil"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
)

func TestStartServer_Starts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockHandler := mockhttputil.NewMockHandler(mockCtrl)
	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Do(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("ok"))
	})

	ctx, cancel := context.WithCancel(context.Background())

	address := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", netutil.RandomPort())
	ma, err := ma.NewMultiaddr(address)
	require.NoError(t, err, "ma.NewMultiaddr")

	addr, err := manet.ToNetAddr(ma)
	require.NoError(t, err, "manet.ToNetAddr")

	serverDone := make(chan error, 1)

	go func() {
		serverDone <- StartServer(ctx, address, mockHandler)
	}()

	test.WaitUntil(t, 20*time.Millisecond, 5*time.Millisecond, func() error {
		res, err := http.Get(fmt.Sprintf("http://%s", addr.String()))
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err, "ioutil.ReadAll(res.Body)")
		assert.Equal(t, "ok", string(body), "Unexpected body")

		return nil
	}, "HTTP server did not start fast enough")

	cancel()
	select {
	case <-time.After(time.Second):
		require.Fail(t, "Server shutdown is too slow")
	case err := <-serverDone:
		require.Fail(t, err.Error())
	case <-ctx.Done():
		assert.EqualError(t, ctx.Err(), context.Canceled.Error())
	}
}
