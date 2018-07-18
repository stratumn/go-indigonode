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
	"time"

	"github.com/stratumn/go-indigonode/app/indigo/service/fossilizer"
	"github.com/stratumn/go-indigonode/core/manager/testservice"
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
