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

package testutil

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	db "github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stretchr/testify/require"
)

// NewSimpleChain returns a SimpleState ready to use in tests.
func NewSimpleChain(t *testing.T, opts ...chain.Opt) chain.Chain {
	memdb, err := db.NewMemDB(nil)
	require.NoError(t, err, "db.NewMemDB()")

	return chain.NewChainDB(memdb, opts...)
}
