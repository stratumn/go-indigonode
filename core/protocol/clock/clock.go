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

package clock

import (
	"context"
	"io"
	"sync"
	"time"

	ihost "gx/ipfs/QmP46LGWhzVZTMmt5akNNLfoV8qL4h5wTwmzQxLyDafggd/go-libp2p-host"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"

	pb "github.com/stratumn/alice/pb/clock"

	"github.com/pkg/errors"
)

// log is the logger for the protocol.
var log = logging.Logger("clock")

// Host represents an Alice host.
type Host = ihost.Host

// ProtocolID is the protocol ID of the protocol.
var ProtocolID = protocol.ID("/alice/clock/v1.0.0")

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

	for {
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
