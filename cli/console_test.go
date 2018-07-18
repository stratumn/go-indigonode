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

package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equalf(t, "test", buf.String(), "%s: debug: test.fn(\"test\")", tt.name)
		buf.Reset()

		if tt.debug {
			cons.SetDebug(false)

			tt.fn("test")
			assert.Equalf(t, "", buf.String(), "%s: nodebug: test.fn(\"test\")", tt.name)
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
		assert.Equalf(t, want, buf.String(), "%s: color: test.fn(\"test\")", tt.name)
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
		assert.Equalf(t, "test\n", buf.String(), "%s: debug: test.fn(\"test\")", tt.name)
		buf.Reset()

		if tt.debug {
			cons.SetDebug(false)

			tt.fn("test\n")
			assert.Equalf(t, "", buf.String(), "%s: nodebug: test.fn(\"test\")", tt.name)
			buf.Reset()
		}

		cons.SetDebug(true)
		cons.color = true

		tt.fn("test")
		want := "test\n"
		if tt.color != "" {
			want = tt.color + want + ansiReset
		}
		assert.Equalf(t, want, buf.String(), "%s: color: test.fn(\"test\")", tt.name)
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
		assert.Equalf(t, "test", buf.String(), "%s: debug: test.fn(\"test\")", tt.name)
		buf.Reset()

		cons.color = true

		tt.fn("%s", "test")
		want := "test"
		if tt.color != "" {
			want = tt.color + want + ansiReset
		}
		assert.Equalf(t, want, buf.String(), "%s: debug: color.fn(\"test\")", tt.name)
		buf.Reset()
	}
}
