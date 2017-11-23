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
