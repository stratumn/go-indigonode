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
	"gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
)

// Codec implements an Encoder and a Decoder.
type Codec interface {
	multicodec.Encoder
	multicodec.Decoder
}

// ProtobufCodec uses protobuf to encode messages.
type ProtobufCodec struct {
	multicodec.Encoder
	multicodec.Decoder
}

// NewProtobufCodec creates a Codec over the given stream,
// using protobuf as the message format.
func NewProtobufCodec(stream inet.Stream) Codec {
	enc := protobuf.Multicodec(nil).Encoder(stream)
	dec := protobuf.Multicodec(nil).Decoder(stream)

	return &ProtobufCodec{
		Encoder: enc,
		Decoder: dec,
	}
}
