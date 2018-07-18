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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitPrivateNetwork_InvalidArgs(t *testing.T) {
	prevOsExit := osExit
	osExit = func(code int, _ string) {
		// Panic can be caught in tests, os.Exit cannot.
		panic(code)
	}
	defer func() { osExit = prevOsExit }()

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
