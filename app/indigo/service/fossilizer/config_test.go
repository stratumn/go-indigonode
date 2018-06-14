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

package fossilizer_test

import (
	"context"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/batchfossilizer"
	"github.com/stratumn/go-indigocore/bcbatchfossilizer"
	"github.com/stratumn/go-indigocore/blockchain/btc"
	"github.com/stratumn/go-indigocore/dummyfossilizer"

	indigofossilizer "github.com/stratumn/go-indigocore/fossilizer"

	"github.com/stratumn/alice/app/indigo/service/fossilizer"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name       string
	cfg        *fossilizer.Config
	entityType indigofossilizer.Adapter
	err        string
}

func runTestCases(t *testing.T, tests []testCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := tt.cfg.CreateIndigoFossilizer(context.Background())
			if tt.err == "" {
				assert.NoError(t, err)
				assert.NotNil(t, f)
				assert.IsType(t, tt.entityType, f)
			} else {
				assert.EqualError(t, err, tt.err)
			}
		})
	}
}
func TestConfig_CreateFossilizer(t *testing.T) {
	t.Run("Unknown fossilizer", func(t *testing.T) {
		runTestCases(t, []testCase{
			{
				name: "returns an error",
				cfg: &fossilizer.Config{
					FossilizerType: "unknown",
				},
				err: fossilizer.ErrNotImplemented.Error(),
			},
		})
	})

	t.Run("Dummy fossilizer", func(t *testing.T) {
		runTestCases(t, []testCase{
			{
				name: "Cannot fail",
				cfg: &fossilizer.Config{
					FossilizerType: fossilizer.Dummy,
				},
				entityType: &dummyfossilizer.DummyFossilizer{},
			},
		})
	})

	t.Run("Dummy Batch fossilizer", func(t *testing.T) {
		runTestCases(t, []testCase{
			{
				name: "Cannot fail",
				cfg: &fossilizer.Config{
					FossilizerType: fossilizer.DummyBatch,
				},
				entityType: &batchfossilizer.Fossilizer{},
			},
		})
	})

	t.Run("Blockchain Batch fossilizer", func(t *testing.T) {
		tests := []testCase{
			{
				name: "using dummy timestamping",
				cfg: &fossilizer.Config{
					FossilizerType: fossilizer.BlockchainBatch,
					Timestamper:    fossilizer.DummyTimestamper,
				},
				entityType: &bcbatchfossilizer.Fossilizer{},
			},
			{
				name: "using bitcoin timestamping",
				cfg: &fossilizer.Config{
					FossilizerType: fossilizer.BlockchainBatch,
					Timestamper:    fossilizer.BitcoinTimestamper,
					BtcWIF:         "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
				},
				entityType: &bcbatchfossilizer.Fossilizer{},
			},
			{
				name: "using bitcoin timestamping with a bad WIF",
				cfg: &fossilizer.Config{
					FossilizerType: fossilizer.BlockchainBatch,
					Timestamper:    fossilizer.BitcoinTimestamper,
					BtcWIF:         "bad wif",
				},
				entityType: &bcbatchfossilizer.Fossilizer{},
				err:        errors.Wrap(btcutil.ErrMalformedPrivateKey, btc.ErrBadWIF.Error()).Error(),
			},
		}
		runTestCases(t, tests)
	})

}
