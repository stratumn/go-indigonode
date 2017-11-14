// Copyright © 2017  Stratumn SAS
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

package netutil

import (
	"net"

	manet "gx/ipfs/QmX3U3YXCQ6UYBxq2LVWF8dARS1hPUTEYLrSx654Qyxyw6/go-multiaddr-net"
	maddr "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"

	"github.com/pkg/errors"
)

// Listen returns a listener that is compatible with net.Listener but uses a
// multiaddr listener under the hood.
func Listen(address string) (net.Listener, error) {
	maddr, err := maddr.NewMultiaddr(address)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	addr, err := manet.ToNetAddr(maddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	lis, err := manet.Listen(maddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &listenerWrapper{lis, addr}, nil
}

// listenerWrapper wraps an manet.Listener to make it compatible with
// net.Listener.
type listenerWrapper struct {
	lis  manet.Listener
	addr net.Addr
}

// Accept waits for and returns the next connection to the listener.
func (l *listenerWrapper) Accept() (net.Conn, error) {
	conn, err := l.lis.Accept()
	return conn, errors.WithStack(err)
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return an error.
func (l *listenerWrapper) Close() error {
	return errors.WithStack(l.lis.Close())
}

// Addr returns the listener's network address.
func (l *listenerWrapper) Addr() net.Addr {
	return l.addr
}
