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

// Package protocol defines types for the clock protocol.
package protocol

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/app/clock/pb"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
	protobuf "gx/ipfs/QmewJ1Zp9Hwz5HcMd7JYjhLXwvEHTL2UBCCz3oLt1E2N5z/go-multicodec/protobuf"
)

// log is the logger for the protocol.
var log = logging.Logger("clock")

// Host represents a Stratumn Node host.
type Host = ihost.Host

// ProtocolID is the protocol ID of the protocol.
var ProtocolID = protocol.ID("/stratumn/node/clock/v1.0.0")

// Clock implements the clock protocol.
type Clock struct {
	host    Host
	timeout time.Duration
	wg      sync.WaitGroup
}

// NewClock creates a new clock.
func NewClock(host Host, timeout time.Duration) *Clock {
	return &Clock{host: host, timeout: timeout}
}

// ProtocolID returns this protocol ID
func (c *Clock) ProtocolID() protocol.ID {
	return ProtocolID
}

// Wait waits for all the streams to be closed.
func (c *Clock) Wait() {
	c.wg.Wait()
}

// StreamHandler handles incoming messages from a peer.
func (c *Clock) StreamHandler(ctx context.Context, stream inet.Stream) {
	log.Event(ctx, "beginStream", logging.Metadata{
		"stream": stream,
	})
	defer log.Event(ctx, "endStream", logging.Metadata{
		"stream": stream,
	})

	// Protobuf is certainly overkill here, but we use it anyway since this
	// protocol acts as an example.
	enc := protobuf.Multicodec(nil).Encoder(stream)

	c.wg.Add(1)
	defer c.wg.Done()

	buf := make([]byte, 1)

	ch := make(chan error, 1)

	go func() {
		// Everytime we receive any byte, we write the local time.
		_, err := io.ReadFull(stream, buf)
		if err != nil {
			ch <- err
			return
		}

		t := time.Now().UTC()

		err = enc.Encode(&pb.Time{Timestamp: t.UnixNano()})
		if err != nil {
			ch <- errors.WithStack(err)
			return
		}

		ch <- nil
	}()

	select {
	case <-ctx.Done():
		c.closeStream(ctx, stream)
		return

	case <-time.After(c.timeout):
		c.closeStream(ctx, stream)
		return

	case err := <-ch:
		if err != nil {
			c.handleStreamError(ctx, stream, err)
			return
		}
	}
}

// RemoteTime asks a peer for its time.
func (c *Clock) RemoteTime(ctx context.Context, pid peer.ID) (*time.Time, error) {
	event := log.EventBegin(ctx, "RemoteTime", logging.Metadata{
		"peerID": pid.Pretty(),
	})
	defer event.Done()

	timeCh := make(chan *time.Time, 1)
	errCh := make(chan error, 1)

	go func() {
		stream, err := c.host.NewStream(ctx, pid, c.ProtocolID())
		if err != nil {
			event.SetError(err)
			errCh <- errors.WithStack(err)
			return
		}

		_, err = stream.Write([]byte{'\n'})
		if err != nil {
			event.SetError(err)
			errCh <- errors.WithStack(err)
			return
		}

		dec := protobuf.Multicodec(nil).Decoder(stream)

		t := pb.Time{}

		err = dec.Decode(&t)
		if err != nil {
			event.SetError(err)
			errCh <- errors.WithStack(err)
			return
		}
		time := time.Unix(0, t.Timestamp)

		timeCh <- &time
	}()

	select {
	case <-ctx.Done():
		return nil, errors.WithStack(ctx.Err())
	case t := <-timeCh:
		return t, nil
	case err := <-errCh:
		return nil, err
	}
}

// closeStream closes a stream.
func (c *Clock) closeStream(ctx context.Context, stream inet.Stream) {
	if err := stream.Close(); err != nil {
		log.Event(ctx, "streamCloseError", logging.Metadata{
			"error":  err.Error(),
			"stream": stream,
		})
	}
}

// handleStreamError handles errors from streams.
func (c *Clock) handleStreamError(ctx context.Context, stream inet.Stream, err error) {
	log.Event(ctx, "streamError", logging.Metadata{
		"error":  err.Error(),
		"stream": stream,
	})

	if err == io.EOF {
		c.closeStream(ctx, stream)
		return
	}

	if err := stream.Reset(); err != nil {
		log.Event(ctx, "streamResetError", logging.Metadata{
			"error":  err.Error(),
			"stream": stream,
		})
	}
}
