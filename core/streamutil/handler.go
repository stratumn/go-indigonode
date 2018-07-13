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

package streamutil

import (
	"context"

	"github.com/stratumn/go-indigonode/core/monitoring"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
)

// AutoCloseHandler is a specialized stream handler that closes the stream
// when exiting.
// It automatically logs the handler error.
type AutoCloseHandler func(context.Context, *monitoring.Span, inet.Stream, Codec) error

// WithAutoClose transforms an AutoCloseHandler to a StreamHandler.
func WithAutoClose(service string, method string, h AutoCloseHandler) inet.StreamHandler {
	return func(stream inet.Stream) {
		var err error
		ctx, span := monitoring.StartSpan(
			context.Background(),
			service,
			method,
			monitoring.SpanOptionPeerID(stream.Conn().RemotePeer()),
			monitoring.SpanOptionProtocolID(stream.Protocol()),
		)

		defer func() {
			if err != nil {
				span.SetUnknownError(err)
			}

			if closeErr := stream.Close(); closeErr != nil {
				span.Annotate(ctx, "close_err", closeErr.Error())
			}

			span.End()
		}()

		codec := NewProtobufCodec(stream)
		err = h(ctx, span, stream, codec)
	}
}
