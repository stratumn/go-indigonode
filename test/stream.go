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

package test

import (
	"errors"
	"io"
	"testing"
	"time"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
)

// WaitUntilStreamClosed waits until a stream is closed by the other party.
func WaitUntilStreamClosed(t *testing.T, stream inet.Stream) {
	WaitUntil(
		t,
		25*time.Millisecond,
		5*time.Millisecond,
		func() error {
			_, err := stream.Read([]byte{42})
			if err == io.EOF {
				return nil
			}

			return errors.New("stream not closed yet")
		},
		"stream not closed in time",
	)
}

// Stream wraps an existing stream to provide additional testing information.
type Stream struct {
	inet.Stream

	closeChan chan struct{}
	resetChan chan struct{}
}

// WrapStream wraps an existing stream to provide additional testing information.
func WrapStream(stream inet.Stream) *Stream {
	return &Stream{
		Stream:    stream,
		closeChan: make(chan struct{}, 1),
		resetChan: make(chan struct{}, 1),
	}
}

// CloseChan returns the channel used to notify of close events.
func (s *Stream) CloseChan() <-chan struct{} {
	return s.closeChan
}

// Close closes the stream and notifies a channel.
func (s *Stream) Close() error {
	err := s.Stream.Close()
	s.closeChan <- struct{}{}
	return err
}

// ResetChan returns the channel used to notify of reset events.
func (s *Stream) ResetChan() <-chan struct{} {
	return s.resetChan
}

// Reset resets the stream and notifies a channel.
func (s *Stream) Reset() error {
	err := s.Stream.Reset()
	s.resetChan <- struct{}{}
	return err
}
