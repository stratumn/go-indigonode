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

package fossilizer

import (
	"context"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	"github.com/stratumn/go-indigocore/batchfossilizer"
	"github.com/stratumn/go-indigocore/fossilizer"
	"github.com/stratumn/go-indigonode/core/monitoring"
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
	ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ErrorTag, "success").Build()
	event := log.EventBegin(ctx, "Fossilize")
	defer event.Done()

	err := f.fossilizer.Fossilize(ctx, data, meta)
	if err != nil {
		ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ErrorTag, err.Error()).Build()
		event.SetError(err)
	}

	fossils.Record(ctx, 1)

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
