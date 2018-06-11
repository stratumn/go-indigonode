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

// Package streamutil provides utility functions
// to make handling streams easier.
package streamutil

import (
	"context"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
)

// AutoCloseHandler is a specialized stream handler that closes the stream
// when exiting.
// It automatically logs the handler error.
type AutoCloseHandler func(context.Context, inet.Stream, *logging.EventInProgress) error

// WithAutoClose transforms an AutoCloseHandler to a StreamHandler.
func WithAutoClose(log logging.EventLogger, name string, h AutoCloseHandler) inet.StreamHandler {
	return func(stream inet.Stream) {
		ctx := context.Background()

		var err error

		event := log.EventBegin(ctx, name, logging.Metadata{
			"remote": stream.Conn().RemotePeer().Pretty(),
		})

		defer func() {
			if err != nil {
				event.SetError(err)
			}

			if closeErr := stream.Close(); closeErr != nil {
				event.Append(logging.Metadata{"close_err": closeErr.Error()})
			}

			event.Done()
		}()

		err = h(ctx, stream, event)
	}
}
