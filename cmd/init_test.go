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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	Execute()

	// Panic can be caught in tests, os.Exit cannot.
	osExit = func(code int) { panic(code) }
}

func TestInitPrivateNetwork_InvalidArgs(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(*testing.T)
	}{{
		"missing-coordinator-addr",
		func(t *testing.T) {
			require.NoError(t, initCmd.Flags().Set(
				PrivateWithCoordinatorFlagName,
				"true",
			))
		},
	}, {
		"invalid-coordinator-addr",
		func(t *testing.T) {
			require.NoError(t, initCmd.Flags().Set(
				PrivateWithCoordinatorFlagName,
				"true",
			))
			require.NoError(t, initCmd.Flags().Set(
				CoordinatorAddrFlagName,
				"/not/a/multi/addr",
			))
		},
	}, {
		"invalid-flags-combination",
		func(t *testing.T) {
			require.NoError(t, initCmd.Flags().Set(
				PrivateWithCoordinatorFlagName,
				"true",
			))
			require.NoError(t, initCmd.Flags().Set(
				CoordinatorAddrFlagName,
				"/ip4/127.0.0.1/tcp/8903/ipfs/QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9",
			))
			require.NoError(t, initCmd.Flags().Set(
				PrivateCoordinatorFlagName,
				"true",
			))
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Reset all flags
			require.NoError(t, initCmd.Flags().Set(PrivateCoordinatorFlagName, "false"))
			require.NoError(t, initCmd.Flags().Set(PrivateWithCoordinatorFlagName, "false"))
			require.NoError(t, initCmd.Flags().Set(CoordinatorAddrFlagName, ""))

			tt.setup(t)

			require.Panics(t, func() { initCmd.Run(RootCmd, nil) })
		})
	}
}
