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

package cli

import (
	"fmt"
	"io"
)

// ANSI color escape sequences.
var ansiReset = "\033[0m"
var ansiGray = "\033[0;37m"
var ansiBlue = "\033[0;34m"
var ansiGreen = "\033[0;32m"
var ansiYellow = "\033[0;33m"
var ansiRed = "\033[0;31m"

// Console is a simple console with optional color output.
type Console struct {
	io.Writer
	color bool
	debug bool
}

// NewConsole creates a new VT100 console.
func NewConsole(out io.Writer, color bool) *Console {
	c := Console{out, color, false}
	return &c
}

// SetDebug enables or disables debug output.
func (c *Console) SetDebug(debug bool) {
	c.debug = debug
}

// Reset resets the text color.
func (c *Console) Reset() {
	if c.color {
		fmt.Fprint(c, ansiReset)
	}
}

// Gray sets the text color to gray.
func (c *Console) Gray() {
	if c.color {
		fmt.Fprint(c, ansiGray)
	}
}

// Blue sets the text color to blue.
func (c *Console) Blue() {
	if c.color {
		fmt.Fprint(c, ansiBlue)
	}
}

// Green sets the text color to green.
func (c *Console) Green() {
	if c.color {
		fmt.Fprint(c, ansiGreen)
	}
}

// Yellow sets the text color to yellow.
func (c *Console) Yellow() {
	if c.color {
		fmt.Fprint(c, ansiYellow)
	}
}

// Red sets the text color to red.
func (c *Console) Red() {
	if c.color {
		fmt.Fprint(c, ansiRed)
	}
}

// Debug outputs gray text that will only be visible in debug mode.
func (c *Console) Debug(v ...interface{}) {
	if !c.debug {
		return
	}

	c.Gray()
	fmt.Fprint(c, v...)
	c.Reset()
}

// Print outputs text.
func (c *Console) Print(v ...interface{}) {
	fmt.Fprint(c, v...)
}

// Info outputs blue text.
func (c *Console) Info(v ...interface{}) {
	c.Blue()
	fmt.Fprint(c, v...)
	c.Reset()
}

// Success outputs green text.
func (c *Console) Success(v ...interface{}) {
	c.Green()
	fmt.Fprint(c, v...)
	c.Reset()
}

// Warning outputs yellow text.
func (c *Console) Warning(v ...interface{}) {
	c.Yellow()
	fmt.Fprint(c, v...)
	c.Reset()
}

// Error outputs red text.
func (c *Console) Error(v ...interface{}) {
	c.Red()
	fmt.Fprint(c, v...)
	c.Reset()
}

// Debugln outputs a line of gray text that will only be visible in debug mode.
func (c *Console) Debugln(v ...interface{}) {
	if !c.debug {
		return
	}

	c.Gray()
	fmt.Fprintln(c, v...)
	c.Reset()
}

// Println outputs a line of text.
func (c *Console) Println(v ...interface{}) {
	fmt.Fprintln(c, v...)
}

// Infoln outputs a line of blue text.
func (c *Console) Infoln(v ...interface{}) {
	c.Blue()
	fmt.Fprintln(c, v...)
	c.Reset()
}

// Successln outputs a line of green text.
func (c *Console) Successln(v ...interface{}) {
	c.Green()
	fmt.Fprintln(c, v...)
	c.Reset()
}

// Warningln outputs a line of yellow text.
func (c *Console) Warningln(v ...interface{}) {
	c.Yellow()
	fmt.Fprintln(c, v...)
	c.Reset()
}

// Errorln outputs a line of red text.
func (c *Console) Errorln(v ...interface{}) {
	c.Red()
	fmt.Fprintln(c, v...)
	c.Reset()
}

// Debugf outputs formatted gray text that will only be visible in debug mode.
func (c *Console) Debugf(format string, v ...interface{}) {
	if !c.debug {
		return
	}

	c.Gray()
	fmt.Fprintf(c, format, v...)
	c.Reset()
}

// Printf outputs formatted text.
func (c *Console) Printf(format string, v ...interface{}) {
	fmt.Fprintf(c, format, v...)
}

// Infof outputs formatted blue text.
func (c *Console) Infof(format string, v ...interface{}) {
	c.Blue()
	fmt.Fprintf(c, format, v...)
	c.Reset()
}

// Successf outputs formatted green text.
func (c *Console) Successf(format string, v ...interface{}) {
	c.Green()
	fmt.Fprintf(c, format, v...)
	c.Reset()
}

// Warningf outputs formatted yellow text.
func (c *Console) Warningf(format string, v ...interface{}) {
	c.Yellow()
	fmt.Fprintf(c, format, v...)
	c.Reset()
}

// Errorf outputs formatted red text.
func (c *Console) Errorf(format string, v ...interface{}) {
	c.Red()
	fmt.Fprintf(c, format, v...)
	c.Reset()
}
