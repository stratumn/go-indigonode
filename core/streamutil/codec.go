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
