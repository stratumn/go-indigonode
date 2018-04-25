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

// Package fossilizer contains the Indigo Fossilizer service.
package fossilizer

import (
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/dummyfossilizer"
	"github.com/stratumn/go-indigocore/fossilizer"
)

const (
	// Dummy describes the dummyfossilizer type.
	Dummy = "dummy"
)

var (
	// ErrNotImplemented is returned when trying to instantiate an unknown type of fossilizer.
	ErrNotImplemented = errors.New("fossilizer type is not implemented")
)

// Config contains configuration options for the Fossilizer service.
type Config struct {
	// Version is the version of the Indigo Fossilizer service.
	Version string `toml:"version" comment:"The version of the indigo fossilizer service."`

	// FossilizerType is the fossilizer implementation.
	FossilizerType string `toml:"fossilizer_type" comment:"The type of fossilizer (eg: dummy, dummybatch, bitcoin...)."`

	// Timestamper is the backend of the timestamping mechanism.
	Timestamper string `toml:"timestamper" comment:"The backend for timestamping (only applicable to blockchain fossilizers)."`

	// Interval between batches (if any).
	Interval int64 `toml:"interval" comment:"The time interval between batches expressed in seconds (only applicable to fossilizers using batches)."`

	// Maximum number of leaves of a Merkle tree.
	MaxLeaves int `toml:"max_leaves" comment:"The maximum number of leaves of a merkle tree in a batch (only applicable to fossilizers using batches)."`
}

// CreateIndigoFossilizer creates an indigo fossilizer from the configuration.
func (c *Config) CreateIndigoFossilizer() (fossilizer.Adapter, error) {
	switch c.FossilizerType {
	case Dummy:
		return dummyfossilizer.New(&dummyfossilizer.Config{}), nil
	default:
		return nil, ErrNotImplemented
	}
}
