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

// Package netutil defines useful networking types.
package netutil

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/pkg/errors"

	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
)

// Listen returns a listener that is compatible with net.Listener but uses a
// multiaddr listener under the hood.
func Listen(address string) (net.Listener, error) {
	ma, err := ma.NewMultiaddr(address)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	addr, err := manet.ToNetAddr(ma)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	lis, err := manet.Listen(ma)
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

// port is atomically incremented each time a port is assigned.
var port = int32(4000)

// RandomPort returns a random port.
func RandomPort() uint16 {
	for {
		p := atomic.AddInt32(&port, 1)
		if p >= 65535 {
			// TODO: handle this better.
			panic("no more ports")
		}

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err == nil {
			if err := l.Close(); err != nil {
				continue
			}
			return uint16(p)
		}
	}
}
