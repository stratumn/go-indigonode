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

package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/testutil/blocktest"
	"github.com/stratumn/go-node/core/db"
)

func bench(b *testing.B, fn func(*testing.B, db.DB)) {
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

	b.ResetTimer()
	fn(b, db)
}

func BenchmarkState_Read(b *testing.B) {
	bench(b, func(b *testing.B, database db.DB) {
		s := NewState(database)

		for i := 0; i < 10000; i++ {
			err := s.UpdateAccount([]byte(fmt.Sprintf("#%10d", i)), &pb.Account{
				Balance: uint64(i),
			})
			if err != nil {
				b.Fatal(err)
			}
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := s.GetAccount([]byte(fmt.Sprintf("#%10d", i%10000)))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkState_Read_parallel(b *testing.B) {
	bench(b, func(b *testing.B, database db.DB) {
		b.RunParallel(func(p *testing.PB) {
			s := NewState(database)

			for i := 0; i < 10000; i++ {
				err := s.UpdateAccount([]byte(fmt.Sprintf("#%10d", i)), &pb.Account{
					Balance: uint64(i),
				})
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()

			var i int64

			for p.Next() {
				atomic.AddInt64(&i, 1)
				_, err := s.GetAccount([]byte(fmt.Sprintf("#%10d", i)))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkState_Write(b *testing.B) {
	for i := 1; i < 100000; i *= 10 {
		b.Run(fmt.Sprintf("process-transactions-block-size-%d", i), func(b *testing.B) {
			bench(b, func(b *testing.B, database db.DB) {
				benchmarkProcess(b, database, i)
			})
		})
		b.Run(fmt.Sprintf("rollback-transactions-block-size-%d", i), func(b *testing.B) {
			bench(b, func(b *testing.B, database db.DB) {
				benchmarkRollback(b, database, i)
			})
		})
	}
}

// Benchmarks b.N blocks of size n.
func benchmarkProcess(b *testing.B, db db.DB, n int) {
	s := NewState(db)

	for i := 0; i < n; i++ {
		err := s.UpdateAccount([]byte(fmt.Sprintf("#%10d", i)), &pb.Account{
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
	blk := blocktest.NewBlock(b, txs)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		blk.GetHeader().BlockNumber = uint64(i)

		err := s.ProcessBlock(blk)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmarks b.N blocks of size n.
func benchmarkRollback(b *testing.B, db db.DB, n int) {
	s := NewState(db)

	for i := 0; i < n; i++ {
		err := s.UpdateAccount([]byte(fmt.Sprintf("#%10d", i)), &pb.Account{
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
	blk := blocktest.NewBlock(b, txs)

	for i := 0; i < b.N; i++ {
		blk.GetHeader().BlockNumber = uint64(i)

		err := s.ProcessBlock(blk)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := b.N - 1; i >= 0; i-- {
		blk.GetHeader().BlockNumber = uint64(i)

		err := s.RollbackBlock(blk)
		if err != nil {
			b.Fatal(err)
		}
	}
}
