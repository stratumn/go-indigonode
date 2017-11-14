// Copyright © 2017 Stratumn SAS
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

// Package release contains variables describing the current release
// and are overriden during compilation.
package release

import (
	"encoding/hex"
)

var (
	// Protocol is the protocol used by Alice.
	Protocol = "/alice/0.0.1"

	// Version is the release version string.
	Version = "v0.0.1"

	// GitCommit is the hash of the last Git commit.
	GitCommit = "0000000000000000000000000000000000000000"

	// GitCommitBytes bytes, computed at runtime.
	GitCommitBytes []byte

	// BootstrapAddresses is the list of bootstrap addresses.
	BootstrapAddresses = []string{
		"/dnsaddr/impulse.io/ipfs/Qmc1QbSba7RtPgxEw4NqXNeDpB5CpCTwv9dvdZRdTkche1",
		"/dnsaddr/impulse.io/ipfs/QmQVdocY8ZbYxrKRSrff2Vxmm27Mhu6DgXyWXQwmuz1b6P",
		"/dnsaddr/impulse.io/ipfs/QmQJib6mnEMgdCe3bGH1YP7JswHbQQejyNucvW9BjFqmWr",
		"/dnsaddr/impulse.io/ipfs/Qmc1rLFp5stHrjtq4duFg6KakBcDCpB3bTjjMZVSAdnHLj",
		"/dnsaddr/impulse.io/ipfs/QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
	}
)

func init() {
	var err error

	// Cache the hex representation of the Git commit hash.
	GitCommitBytes, err = hex.DecodeString(GitCommit)
	if err != nil {
		panic(err)
	}
}
