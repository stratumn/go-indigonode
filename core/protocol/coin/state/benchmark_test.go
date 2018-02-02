// Copyright © 2017-2018  Stratumn SAS
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

package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	db "github.com/stratumn/alice/core/protocol/coin/db"
	pb "github.com/stratumn/alice/pb/coin"
)

func BenchmarkState(b *testing.B) {
	filename, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(filename)

	db, err := db.NewFileDB(filename, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	for i := 1; i < 100000; i *= 10 {
		b.Run(fmt.Sprintf("process-transactions-block-size-%d", i), func(b *testing.B) {
			benchmarkProcess(b, db, i)
		})
		b.Run(fmt.Sprintf("rollback-transactions-block-size-%d", i), func(b *testing.B) {
			benchmarkRollback(b, db, i)
		})
	}
}

// Benchmarks b.N blocks of size n.
func benchmarkProcess(b *testing.B, db db.DB, n int) {
	s := NewState(db, nil, 14)

	for i := 0; i < n; i++ {
		err := s.UpdateAccount([]byte(fmt.Sprintf("#%10d", i)), Account{
			Balance: 1, // no need to make it more :)
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	txs := make([]*pb.Transaction, n)

	for i := 0; i < n; i++ {
		txs[i] = &pb.Transaction{
			From:  []byte(fmt.Sprintf("#%10d", i%n)),
			To:    []byte(fmt.Sprintf("#%10d", (i+1)%n)),
			Value: 1,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := s.ProcessTransactions([]byte(fmt.Sprintf("job-%10d", i)), txs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmarks b.N blocks of size n.
func benchmarkRollback(b *testing.B, db db.DB, n int) {
	s := NewState(db, nil, 14)

	for i := 0; i < n; i++ {
		err := s.UpdateAccount([]byte(fmt.Sprintf("#%10d", i)), Account{
			Balance: 1, // no need to make it more :)
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	txs := make([]*pb.Transaction, n)

	for i := 0; i < n; i++ {
		txs[i] = &pb.Transaction{
			From:  []byte(fmt.Sprintf("#%10d", i%n)),
			To:    []byte(fmt.Sprintf("#%10d", (i+1)%n)),
			Value: 1,
		}
	}

	for i := 0; i < b.N; i++ {
		err := s.ProcessTransactions([]byte(fmt.Sprintf("job-%10d", i)), txs)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := b.N - 1; i >= 0; i-- {
		err := s.RollbackTransactions([]byte(fmt.Sprintf("job-%10d", i)), txs)
		if err != nil {
			b.Fatal(err)
		}
	}
}
