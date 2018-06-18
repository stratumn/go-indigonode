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

package fossilizer

import (
	"context"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	"github.com/stratumn/go-indigocore/batchfossilizer"
	"github.com/stratumn/go-indigocore/fossilizer"
)

var log = logging.Logger("indigo.fossilizer")

// Fossilizer implements github.com/stratumn/go-indigocore/fossilizer.Adapter.
type Fossilizer struct {
	fossilizer fossilizer.Adapter
}

// New instantiates a new fossilizer.
func New(f fossilizer.Adapter) *Fossilizer {
	return &Fossilizer{
		fossilizer: f,
	}
}

// GetInfo returns arbitrary information about the adapter.
func (f *Fossilizer) GetInfo(ctx context.Context) (interface{}, error) {
	log.Event(ctx, "GetInfo")
	return f.fossilizer.GetInfo(ctx)
}

// AddFossilizerEventChan adds a channel that receives events from the fossilizer.
func (f *Fossilizer) AddFossilizerEventChan(ctx context.Context, eventChan chan *fossilizer.Event) {
	log.Event(ctx, "AddFossilizerEventChan")
	f.fossilizer.AddFossilizerEventChan(eventChan)
}

// Fossilize requests data to be fossilized.
func (f *Fossilizer) Fossilize(ctx context.Context, data, meta []byte) error {
	event := log.EventBegin(ctx, "Fossilize")
	defer event.Done()

	err := f.fossilizer.Fossilize(ctx, data, meta)
	if err != nil {
		event.SetError(err)
	}
	return err
}

// Start launches the fossilizer if it does use batches and returns nil otherwise.
func (f *Fossilizer) Start(ctx context.Context) error {
	event := log.EventBegin(ctx, "Start")
	defer event.Done()
	if batchFossilizer, ok := f.fossilizer.(batchfossilizer.Adapter); ok {
		return batchFossilizer.Start(ctx)
	}
	return nil
}

// Started returns a channel that will receive an event once the fossilizer has started.
func (f *Fossilizer) Started(ctx context.Context) <-chan struct{} {
	event := log.EventBegin(ctx, "Started")
	defer event.Done()
	if batchFossilizer, ok := f.fossilizer.(batchfossilizer.Adapter); ok {
		return batchFossilizer.Started()
	}
	c := make(chan struct{}, 1)
	c <- struct{}{}
	return c
}
