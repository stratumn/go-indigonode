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

package store

import (
	"context"

	"github.com/stratumn/go-indigocore/dummystore"
)

// Store implements github.com/stratumn/go-indigocore/store.Adapter.
type Store struct {
	s *dummystore.DummyStore
}

// New creates a new Indigo store.
func New() *Store {
	return &Store{s: dummystore.New(&dummystore.Config{})}
}

// GetInfo returns information about the underlying store.
func (store *Store) GetInfo(ctx context.Context) (interface{}, error) {
	return store.s.GetInfo(ctx)
}
