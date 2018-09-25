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

// Package netutil defines useful networking types.
package netutil

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/pkg/errors"

	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
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

	return &listenerWrapper{lis: lis, addr: addr}, nil
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
