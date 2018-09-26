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

package streamutil

import (
	"context"

	"github.com/stratumn/go-node/core/monitoring"

	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
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

			if closeErr := inet.FullClose(stream); closeErr != nil {
				span.Annotate(ctx, "close_err", closeErr.Error())
			}

			span.End()
		}()

		codec := NewProtobufCodec(stream)
		err = h(ctx, span, stream, codec)
	}
}
