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
	"testing"

	"github.com/stratumn/alice/core/service/indigo/fossilizer"
	"github.com/stratumn/go-indigocore/dummyfossilizer"
	"github.com/stretchr/testify/assert"
)

func TestConfig_CreateFossilizer(t *testing.T) {
	config := &fossilizer.Config{FossilizerType: "dummy"}

	indigoFossilizer, err := config.CreateIndigoFossilizer()
	assert.NoError(t, err)
	assert.NotNil(t, indigoFossilizer)
	_, ok := indigoFossilizer.(*dummyfossilizer.DummyFossilizer)
	assert.True(t, ok)
}
