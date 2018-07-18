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

//+build !lint

package session

import (
	"math/rand"
	"runtime"
	"time"
)

// PerCPU executes a function concurrently with one goroutine per cpu.
func PerCPU(fn func(int), n int) {
	numCPU := runtime.NumCPU()
	sem := make(chan struct{}, numCPU)

	for i := 0; i < n; i++ {
		sem <- struct{}{}

		go func(i int) {
			fn(i)
			<-sem
		}(i)
	}

	for i := 0; i < numCPU; i++ {
		sem <- struct{}{}
	}
}

// randPeers creates random peers.
// The number of peers generated should be at least two in case it contains
// its own address.
func randPeers(pool []string, size int) []string {
	if l := len(pool); size > l {
		size = l
	}

	peers := make([]string, size)
	perm := rand.Perm(len(pool))
	rand.Seed(time.Now().UTC().UnixNano())

	for i := range peers {
		peers[i] = pool[perm[i]]
	}

	return peers
}
