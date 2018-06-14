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
	"time"

	"github.com/stratumn/alice/app/indigo/service/fossilizer"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testService(ctx context.Context, t *testing.T) *fossilizer.Service {
	serv := &fossilizer.Service{}
	config := serv.Config().(fossilizer.Config)
	config.Version = "1.0.0"

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	return serv
}

func TestService_Strings(t *testing.T) {
	testservice.CheckStrings(t, &fossilizer.Service{})
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serv := testService(ctx, t)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_Run_Error(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name   string
		config fossilizer.Config
	}{{
		"invalid-fossilizer-type",
		fossilizer.Config{
			Version:        "1.0.0",
			FossilizerType: "on-the-moon",
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := &fossilizer.Service{}
			require.NoError(t, serv.SetConfig(tt.config), "serv.SetConfig(config)")
			assert.Error(t, serv.Run(ctx, func() {}, func() {}))
		})
	}
}
