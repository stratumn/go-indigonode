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

package httputil

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/core/httputil/mockhttputil"
	"github.com/stratumn/go-node/core/netutil"
	"github.com/stratumn/go-node/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
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
