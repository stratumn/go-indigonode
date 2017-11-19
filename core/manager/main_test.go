// Copyright © 2017  Stratumn SAS
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

package manager

import (
	"flag"
	"os"
	"sync"
	"testing"

	logger "github.com/stratumn/alice/core/log"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

func TestMain(m *testing.M) {
	enableLog := flag.Bool("log", false, "Enable log output to stderr")
	flag.Parse()

	if *enableLog {
		mu := sync.Mutex{}
		w := logger.NewPrettyWriter(os.Stderr, &mu, nil, true)
		logging.WriterGroup.AddWriter(w)
	}

	os.Exit(m.Run())
}
