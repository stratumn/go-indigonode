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
