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

package service

import (
	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/event"
)

// grpcServer is a gRPC server for the chat service.
type grpcServer struct {
	GetEmitter func() Emitter
}

func (s grpcServer) Listen(listenReq *pb.ListenReq, ss pb.Emitter_ListenServer) error {
	log.Event(ss.Context(), "Listen")

	emitter := s.GetEmitter()
	if emitter == nil {
		return errors.New("cannot listen to events: event service is not running")
	}

	receiveChan, err := emitter.AddListener(listenReq.Topic)
	if err != nil {
		return err
	}
	defer emitter.RemoveListener(receiveChan)

eventPump:
	for {
		select {
		case ev, ok := <-receiveChan:
			if !ok {
				break eventPump
			}

			err := ss.Send(ev)
			if err != nil {
				log.Errorf("Could not send event to listeners: %v", ev)
				return errors.WithStack(err)
			}
		case <-ss.Context().Done():
			break eventPump
		}
	}

	return nil
}
