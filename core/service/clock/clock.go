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

//go:generate mockgen -package mockclock -destination mockclock/mockclock.go github.com/stratumn/alice/core/service/clock Host

// Package clock is a simple service that sends the local time to a peer every
// time it receives a byte from that peer.
//
// It is meant to illustrate how to create network services.
package clock

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/clock"
	"google.golang.org/grpc"

	ihost "gx/ipfs/QmP46LGWhzVZTMmt5akNNLfoV8qL4h5wTwmzQxLyDafggd/go-libp2p-host"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	pstore "gx/ipfs/QmYijbtjCxFEjSXaudaQAUz3LN5VKLssm8WCUsRoqzXmQR/go-libp2p-peerstore"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// ProtocolID is the protocol ID of the service.
var ProtocolID = protocol.ID("/alice/clock/v1.0.0")

// Host represents an Alice host.
type Host = ihost.Host

// log is the logger for the service.
var log = logging.Logger("clock")

// Service is the Clock service.
type Service struct {
	config  *Config
	host    Host
	timeout time.Duration
	clock   *Clock
}

// Config contains configuration options for the Clock service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// WriteTimeout sets how long to wait before closing the stream when
	// writing the time to a peer.
	WriteTimeout string `toml:"write_timeout" comment:"How long to wait before closing the stream when writing the time to a peer."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "clock"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Clock"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Returns the time of a node."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Host:         "host",
		WriteTimeout: "10s",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	timeout, err := time.ParseDuration(conf.WriteTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	s.timeout = timeout
	s.config = &conf

	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Host] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.host, ok = exposed[s.config.Host].(Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	return nil
}

// Expose exposes the clock service to other services.
//
// It exposes the type:
//	github.com/stratumn/alice/core/service/*clock.Clock
func (s *Service) Expose() interface{} {
	return s.clock
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.clock = NewClock(s.host, s.timeout)

	// Wrap the stream handler with the context.
	handler := func(stream inet.Stream) {
		s.clock.StreamHandler(ctx, stream)
	}

	s.host.SetStreamHandler(ProtocolID, handler)

	running()
	<-ctx.Done()
	stopping()

	// Stop accepting streams.
	s.host.RemoveStreamHandler(ProtocolID)

	// Wait for all the streams to close.
	s.clock.Wait()

	s.clock = nil

	return errors.WithStack(ctx.Err())
}

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
	// service acts as an example.
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
		stream, err := c.host.NewStream(ctx, pid, ProtocolID)
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

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterClockServer(gs, grpcServer{
		func(ctx context.Context) (*time.Time, error) {
			if s.clock == nil {
				return nil, ErrUnavailable
			}

			t := time.Now()
			return &t, nil
		},
		func(ctx context.Context, pid peer.ID) (*time.Time, error) {
			if s.clock == nil {
				return nil, ErrUnavailable
			}

			return s.clock.RemoteTime(ctx, pid)
		},
		func(ctx context.Context, pi pstore.PeerInfo) error {
			if s.host == nil {
				return ErrUnavailable
			}

			return s.host.Connect(ctx, pi)
		},
	})
}
