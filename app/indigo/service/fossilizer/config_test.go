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

	"github.com/stratumn/go-node/app/indigo/service/fossilizer"
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
