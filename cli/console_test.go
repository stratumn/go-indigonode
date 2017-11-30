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

package cli

import (
	"bytes"
	"testing"
)

func TestConsole_print(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cons := NewConsole(buf, false)

	tests := []struct {
		name  string
		fn    func(...interface{})
		debug bool
		color string
	}{
		{"Debug", cons.Debug, true, ansiGray},
		{"Print", cons.Print, false, ""},
		{"Info", cons.Info, false, ansiBlue},
		{"Success", cons.Success, false, ansiGreen},
		{"Warning", cons.Warning, false, ansiYellow},
		{"Error", cons.Error, false, ansiRed},
	}

	for _, tt := range tests {
		cons.SetDebug(true)
		cons.color = false

		tt.fn("test")
		if got, want := buf.String(), "test"; got != want {
			t.Errorf(`%s: debug: test.fn("test") => %v want %v`, tt.name, got, want)
		}
		buf.Reset()

		if tt.debug {
			cons.SetDebug(false)

			tt.fn("test")
			if got, want := buf.String(), ""; got != want {
				t.Errorf(`%s: nodebug: test.fn("test") => %v want %v`, tt.name, got, want)
			}
			buf.Reset()
		}

		cons.debug = true
		cons.SetDebug(true)
		cons.color = true

		tt.fn("test")
		want := "test"
		if tt.color != "" {
			want = tt.color + want + ansiReset
		}
		if got := buf.String(); got != want {
			t.Errorf(`%s: color: test.fn("test") => %v want %v`, tt.name, got, want)
		}
		buf.Reset()
	}

}

func TestConsole_println(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cons := NewConsole(buf, false)

	tests := []struct {
		name  string
		fn    func(...interface{})
		debug bool
		color string
	}{
		{"Debugln", cons.Debugln, true, ansiGray},
		{"Println", cons.Println, false, ""},
		{"Infoln", cons.Infoln, false, ansiBlue},
		{"Successln", cons.Successln, false, ansiGreen},
		{"Warningln", cons.Warningln, false, ansiYellow},
		{"Errorln", cons.Errorln, false, ansiRed},
	}

	for _, tt := range tests {
		cons.SetDebug(true)
		cons.color = false

		tt.fn("test")
		if got, want := buf.String(), "test\n"; got != want {
			t.Errorf(`%s: debug: test.fn("test") => %v want %v`, tt.name, got, want)
		}
		buf.Reset()

		if tt.debug {
			cons.SetDebug(false)

			tt.fn("test\n")
			if got, want := buf.String(), ""; got != want {
				t.Errorf(`%s: nodebug: test.fn("test") => %v want %v`, tt.name, got, want)
			}
			buf.Reset()
		}

		cons.SetDebug(true)
		cons.color = true

		tt.fn("test")
		want := "test\n"
		if tt.color != "" {
			want = tt.color + want + ansiReset
		}
		if got := buf.String(); got != want {
			t.Errorf(`%s: color: test.fn("test") => %v want %v`, tt.name, got, want)
		}
		buf.Reset()
	}
}

func TestConsole_printf(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cons := NewConsole(buf, false)

	tests := []struct {
		name  string
		fn    func(string, ...interface{})
		debug bool
		color string
	}{
		{"Debugf", cons.Debugf, true, ansiGray},
		{"Printf", cons.Printf, false, ""},
		{"Infof", cons.Infof, false, ansiBlue},
		{"Successf", cons.Successf, false, ansiGreen},
		{"Warningf", cons.Warningf, false, ansiYellow},
		{"Errorf", cons.Errorf, false, ansiRed},
	}

	for _, tt := range tests {
		cons.SetDebug(true)
		cons.color = false

		tt.fn("%s", "test")
		if got, want := buf.String(), "test"; got != want {
			t.Errorf(`%s: debug: test.fn("test") => %v want %v`, tt.name, got, want)
		}
		buf.Reset()

		cons.color = true

		tt.fn("%s", "test")
		want := "test"
		if tt.color != "" {
			want = tt.color + want + ansiReset
		}
		if got := buf.String(); got != want {
			t.Errorf(`%s: color: test.fn("test") => %v want %v`, tt.name, got, want)
		}
		buf.Reset()
	}
}
