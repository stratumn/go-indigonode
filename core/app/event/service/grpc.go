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

package service

import (
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/app/event/grpc"
)

// grpcServer is a gRPC server for the chat service.
type grpcServer struct {
	GetEmitter func() Emitter
}

func (s grpcServer) Listen(listenReq *grpc.ListenReq, ss grpc.Emitter_ListenServer) error {
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
