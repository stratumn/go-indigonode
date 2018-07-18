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

package script

import (
	"context"
	"fmt"
	"testing"
)

var iter = `
let iter (lambda (n) {
	if (> n 0) (iter (- n 1))
})
`

var seq = `
let seq (lambda(n) {
	let rec (lambda (n tail) {
		if (< n 0) tail (rec (- n 1) (cons n tail))
	})
	rec n ()
})
`

var fib = `
let fib (lambda (n) {
	let fib-rec (lambda (n f1 f2) {
		if (< n 1) f1 else {
			fib-rec (- n 1) f2 (+ f1 f2)
		}
	})
	fib-rec n 0 1
})
`

func BenchmarkInterpreter_EvalInstrs_add(b *testing.B) {
	benchmarkEvalInstrs(b, "", "+ 1 2", b.N)
}

func BenchmarkInterpreter_EvalInstrs_iter(b *testing.B) {
	benchmarkEvalInstrs(b, iter, fmt.Sprintf("iter %d", b.N), 1)
}

func BenchmarkInterpreter_EvalInstrs_seq(b *testing.B) {
	benchmarkEvalInstrs(b, seq, fmt.Sprintf("seq %d", b.N), 1)
}

func BenchmarkInterpreter_EvalInstrs_iter_no_tail_call(b *testing.B) {
	benchmarkEvalInstrs(b, iter, fmt.Sprintf("iter %d", b.N), 1,
		InterpreterOptTailOptimizations(false))
}

func BenchmarkInterpreter_EvalInput_fib(b *testing.B) {
	for i := 10; i <= 10000; i *= 10 {
		b.Run(fmt.Sprintf("n-%d", i), func(b *testing.B) {
			benchmarkEvalInput(b, fib, fmt.Sprintf("fib %d", i), b.N)
		})
	}
}

func BenchmarkInterpreter_EvalInstrs_fib(b *testing.B) {
	for i := 10; i <= 10000; i *= 10 {
		b.Run(fmt.Sprintf("n-%d", i), func(b *testing.B) {
			benchmarkEvalInstrs(b, fib, fmt.Sprintf("fib %d", i), b.N)
		})
	}
}

func benchmarkEvalInput(b *testing.B, src, cmd string, times int, opts ...InterpreterOpt) {
	ctx := context.Background()

	opts = append(opts, InterpreterOptBuiltinLibs)
	opts = append(opts, InterpreterOptValueHandler(func(SExp) {}))

	b.ResetTimer()

	for i := 0; i < times; i++ {
		itr := NewInterpreter(opts...)
		err := itr.EvalInput(ctx, src+"\n"+cmd)
		if err != nil {
			b.Fatalf(err.Error())
		}
	}
}

func benchmarkEvalInstrs(b *testing.B, src, cmd string, times int, opts ...InterpreterOpt) {
	ctx := context.Background()

	opts = append(opts, InterpreterOptBuiltinLibs)
	opts = append(opts, InterpreterOptValueHandler(func(SExp) {}))

	itr := NewInterpreter(opts...)

	if err := itr.EvalInput(ctx, src); err != nil {
		b.Fatal(err)
	}

	parser := NewParser(NewScanner())
	instrs, err := parser.Parse(cmd)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < times; i++ {
		err := itr.EvalInstrs(ctx, instrs)
		if err != nil {
			b.Fatalf(err.Error())
		}
	}
}
