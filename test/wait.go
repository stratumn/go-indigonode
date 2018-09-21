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

// Package test contains a collection of test helpers.
package test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

// WaitUntil waits for duration for the condition function not to return an error
// to return or fails after the delay has elapsed.
func WaitUntil(t *testing.T, duration time.Duration, interval time.Duration, cond func() error, message string) {
	condChan := make(chan struct{})
	var err error
	go func() {
		for {
			if err = cond(); err == nil {
				condChan <- struct{}{}
				return
			}

			<-time.After(interval)
		}
	}()

	select {
	case <-condChan:
	case <-time.After(duration):
		assert.Fail(t, "waitUntil() condition failed:", message, err)
	}
}

// WaitUntilConnected waits until the given host is connected to the given peer.
func WaitUntilConnected(t *testing.T, host ihost.Host, peerID peer.ID) {
	WaitUntil(
		t,
		100*time.Millisecond,
		10*time.Millisecond,
		func() error {
			if host.Network().Connectedness(peerID) == inet.Connected {
				return nil
			}

			return errors.New("peers still not connected")
		}, "peers not connected in time")
}

// WaitUntilDisconnected waits until the given host is disconnected
// from the given peer.
func WaitUntilDisconnected(t *testing.T, host ihost.Host, peerID peer.ID) {
	WaitUntil(
		t,
		100*time.Millisecond,
		10*time.Millisecond,
		func() error {
			if host.Network().Connectedness(peerID) == inet.Connected {
				return errors.New("peers still connected")
			}

			return nil
		}, "peers not disconnected in time")
}
