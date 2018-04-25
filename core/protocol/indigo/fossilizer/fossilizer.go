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

//go:generate mockgen -package mockfossilizer -destination mockfossilizer/mockfossilizer.go github.com/stratumn/go-indigocore/fossilizer Adapter

package fossilizer

import (
	"context"

	"github.com/stratumn/go-indigocore/fossilizer"
)

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
	return nil, nil
}

// AddFossilizerEventChan adds a channel that receives events from the fossilizer.
func (f *Fossilizer) AddFossilizerEventChan(chan *fossilizer.Event) {
}

// Fossilize requests data to be fossilized.
func (f *Fossilizer) Fossilize(ctx context.Context, data []byte, meta []byte) error {
	return nil
}

// Start launches the fossilizer it it does use batches and return nil otherwise.
func (f *Fossilizer) Start(ctx context.Context) error {
	return nil
}
