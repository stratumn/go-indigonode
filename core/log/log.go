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

// Package log deals with logging.
//
// It contains configurable writers for the IPFS log, such as for writing to
// files with file rotation, or for writing colored output to a console.
package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// ErrInvalidWriterType is returned when a writer has an invalid type.
	ErrInvalidWriterType = errors.New("log writer has an invalid type")

	// ErrInvalidWriterLevel is returned when a writer has an invalid
	// level.
	ErrInvalidWriterLevel = errors.New("log writer has an invalid level")

	// ErrInvalidWriterFormatter is returned when a writer has an invalid
	// formatter.
	ErrInvalidWriterFormatter = errors.New("log writer has an invalid formatter")
)

// Log writers.
const (
	File   = "file"
	Stdout = "stdout"
	Stderr = "stderr"
)

// Log levels.
const (
	Error = "error"
	All   = "all"
	Info  = "info"
)

// Log formatters.
const (
	JSON     = "json"
	Text     = "text"
	Color    = "color"
	Journald = "journald"
)

// WriterConfig contains configuration for a log writer.
type WriterConfig struct {
	// Type is the type of the writer.
	Type string `toml:"type" comment:"The type of writer (file, stdout, stderr)."`

	// Level is the log level for the writer.
	Level string `toml:"level" comment:"The log level for the writer (info, error, all)."`

	// File is the log formatter for the writer.
	Formatter string `toml:"formatter" comment:"The formatter for the writer (json, text, color, journald)."`

	// Filename is the file for a file logger.
	Filename string `toml:"filename" comment:"The file for a file logger."`

	// MaximumSize is the maximum size of the file in megabytes before a
	// rotation.
	MaximumSize int `toml:"maximum_size" comment:"The maximum size of the file in megabytes before a rotation."`

	// MaximumAge is the maximum age of the file in days before a rotation.
	MaximumAge int `toml:"maximum_age" comment:"The maximum age of the file in days before a rotation."`

	// MaximumBackups is the maximum number of backups.
	MaximumBackups int `toml:"maximum_backups" comment:"The maximum number of backups."`

	// UseLocalTime is whether to use local time instead of UTC.
	UseLocalTime bool `toml:"use_local_time" comment:"Whether to use local time instead of UTC for backups."`

	// Compress is whether to compress the file.
	Compress bool `toml:"compress" comment:"Whether to compress the file."`
}

// Config contains configuration options for the log.
type Config struct {
	// Writers are the writers for the log.
	Writers []WriterConfig `toml:"writers"`
}

// ConfigHandler is the handler of the log configuration.
type ConfigHandler struct {
	config *Config
}

// ID returns the unique identifier of the configuration.
func (h *ConfigHandler) ID() string {
	return "log"
}

// Config returns the current service configuration or creates one with
// good default values.
//
// The default configuration writes to a single file using the JSON formatter
// and weekly rotations.
func (h *ConfigHandler) Config() interface{} {
	if h.config != nil {
		return *h.config
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(errors.WithStack(err))
	}

	filename, err := filepath.Abs(filepath.Join(cwd, "log.jsonld"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	return Config{
		Writers: []WriterConfig{{
			Type:           File,
			Level:          All,
			Formatter:      JSON,
			Filename:       filename,
			MaximumSize:    128,
			MaximumAge:     7,
			MaximumBackups: 4,
			UseLocalTime:   false,
			Compress:       false,
		}},
	}
}

// files is used to ensure there is only one writer per file.
var files = map[string]io.Writer{}

// SetConfig configures the service handler.
func (h *ConfigHandler) SetConfig(config interface{}) error {
	conf := config.(Config)
	h.config = &conf

	// Use a mutex to make sure stdout and stderr won't overlap.
	var mu sync.Mutex

	for _, writerConf := range conf.Writers {
		var w io.Writer
		switch writerConf.Type {
		case Stdout:
			w = os.Stdout
		case Stderr:
			w = os.Stderr
		case File:
			filename, err := filepath.Abs(writerConf.Filename)
			if err != nil {
				return errors.WithStack(err)
			}

			var ok bool
			if w, ok = files[filename]; ok {
				break // switch
			}

			wc := &lumberjack.Logger{
				Filename:   writerConf.Filename,
				MaxSize:    writerConf.MaximumSize,
				MaxAge:     writerConf.MaximumAge,
				MaxBackups: writerConf.MaximumBackups,
				LocalTime:  writerConf.UseLocalTime,
				Compress:   writerConf.Compress,
			}
			w = wc
			closers = append(closers, wc)
			files[writerConf.Filename] = w
		default:
			return errors.WithStack(ErrInvalidWriterType)
		}

		var filter FilterFunc
		switch writerConf.Level {
		case Info:
			filter = filterInfo
		case Error:
			filter = filterError
		case All:
		default:
			return errors.WithStack(ErrInvalidWriterLevel)
		}

		switch writerConf.Formatter {
		case Text:
			logging.WriterGroup.AddWriter(NewPrettyWriter(w, &mu, filter, false))
		case Color:
			logging.WriterGroup.AddWriter(NewPrettyWriter(w, &mu, filter, true))
		case JSON:
			logging.WriterGroup.AddWriter(NewFilteredWriter(w, &mu, filter))
		case Journald:
			logging.WriterGroup.AddWriter(NewJournaldWriter(w, &mu, filter))
		default:
			return errors.WithStack(ErrInvalidWriterFormatter)
		}
	}

	return nil
}

// closers is a slice of closer to close before exiting.
var closers []io.Closer

// doners are channels that should be closed before exiting.
var doners []chan struct{}

// Close closes all the closers.
func Close() {
	// Wait a bit for log to drain... Couldn't find a better way.
	time.Sleep(100 * time.Millisecond)

	if err := logging.WriterGroup.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
	}

	for _, d := range doners {
		<-d
	}

	for _, c := range closers {
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		}
	}
}

// FilterFunc filters log entries.
type FilterFunc func(map[string]interface{}) bool

// filterInfo removes error entries.
func filterInfo(entry map[string]interface{}) bool {
	_, ok := entry["error"]
	return !ok
}

// filterError removes non-error entries.
func filterError(entry map[string]interface{}) bool {
	_, ok := entry["error"]
	return ok
}

// FilteredWriter filters log output.
type FilteredWriter struct {
	io.WriteCloser
}

// NewFilteredWriter creates a new filtered writer.
func NewFilteredWriter(writer io.Writer, mu sync.Locker, filter FilterFunc) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)
	closers = append(closers, w, r)
	done := make(chan struct{}, 1)
	doners = append(doners, done)

	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}

			entry := map[string]interface{}{}
			if err := json.Unmarshal([]byte(text), &entry); err != nil {
				mu.Lock()
				fmt.Fprintf(writer, "Failed to unmarshal log entry: %s.\n", err)
				mu.Unlock()
				continue
			}

			if filter != nil && !filter(entry) {
				continue
			}

			b, err := json.Marshal(entry)
			if err != nil {
				fmt.Fprintf(writer, "Failed to marshal log entry: %s.\n", err)
			}

			mu.Lock()
			fmt.Fprintln(writer, string(b))
			mu.Unlock()
		}

		close(done)
	}()

	return FilteredWriter{WriteCloser: w}
}

// PrettyWriter writes nicely formatted log output to a console.
type PrettyWriter struct {
	io.WriteCloser
	writer io.Writer
	color  bool
}

// ANSI color escape sequences.
var (
	ansiReset = "\033[0m"
	ansiGray  = "\033[0;37m"
	ansiBlue  = "\033[0;34m"
	ansiGreen = "\033[0;32m"
	ansiRed   = "\033[0;31m"
)

// NewPrettyWriter creates a new pretty writer.
func NewPrettyWriter(writer io.Writer, mu sync.Locker, filter FilterFunc, color bool) io.WriteCloser {
	r, w := io.Pipe()
	pretty := PrettyWriter{WriteCloser: w, writer: writer, color: color}
	scanner := bufio.NewScanner(r)
	closers = append(closers, w, r)
	done := make(chan struct{}, 1)
	doners = append(doners, done)

	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}

			entry := map[string]interface{}{}
			if err := json.Unmarshal([]byte(text), &entry); err != nil {
				mu.Lock()
				fmt.Fprintf(writer, "Failed to unmarshal log entry: %s.\n", err)
				mu.Unlock()
				continue
			}

			if filter != nil && !filter(entry) {
				continue
			}

			date, err := time.Parse(time.RFC3339Nano, entry["time"].(string))
			if err != nil {
				mu.Lock()
				fmt.Fprintf(writer, "Failed to parse log entry date: %s.\n", err)
				mu.Unlock()
				continue
			}

			timestamp := date.Format("15:04:05.000")
			system := entry["system"]
			event := entry["event"]

			mu.Lock()

			fmt.Fprint(writer, timestamp)

			if _, ok := entry["error"]; ok {
				pretty.red(" ERR ")
			} else {
				pretty.green(" INF ")
			}

			pretty.blue(system)

			if subsystem, ok := entry["subsystem"].(string); ok {
				pretty.blue("/" + subsystem)
			}

			pretty.blue(fmt.Sprintf("#%s ", event))

			if duration, ok := entry["duration"].(float64); ok {
				fmt.Fprintf(writer, "(%s) ", time.Duration(duration))
			}

			delete(entry, "time")
			delete(entry, "system")
			delete(entry, "subsystem")
			delete(entry, "event")
			delete(entry, "duration")
			delete(entry, "outcome")

			b, err := json.MarshalIndent(entry, "", "  ")
			if err != nil {
				fmt.Fprintf(writer, "Failed to marshal log entry: %s.\n", err)
			}

			pretty.gray(string(b) + "\n")

			mu.Unlock()
		}

		close(done)
	}()

	return pretty
}

// printColor writers colored text if color is enabled.
func (w *PrettyWriter) printColor(color string, args ...interface{}) {
	if w.color {
		fmt.Fprint(w.writer, color)
	}

	fmt.Fprint(w.writer, args...)

	if w.color {
		fmt.Fprint(w.writer, ansiReset)
	}
}

// gray writes gray text if color is enabled.
func (w *PrettyWriter) gray(args ...interface{}) {
	w.printColor(ansiGray, args...)
}

// blue writes blue text if color is enabled.
func (w *PrettyWriter) blue(args ...interface{}) {
	w.printColor(ansiBlue, args...)
}

// green writes green text if color is enabled.
func (w *PrettyWriter) green(args ...interface{}) {
	w.printColor(ansiGreen, args...)
}

// red writes red text if color is enabled.
func (w *PrettyWriter) red(args ...interface{}) {
	w.printColor(ansiRed, args...)
}

// JournaldWriter prefixes error levels and removes the time.
type JournaldWriter struct {
	io.WriteCloser
}

// NewJournaldWriter creates a new journald writer.
func NewJournaldWriter(writer io.Writer, mu sync.Locker, filter FilterFunc) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)
	closers = append(closers, w, r)
	done := make(chan struct{}, 1)
	doners = append(doners, done)

	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}

			entry := map[string]interface{}{}
			if err := json.Unmarshal([]byte(text), &entry); err != nil {
				mu.Lock()
				fmt.Fprintf(writer, "<3> Failed to unmarshal log entry: %s.\n", err)
				mu.Unlock()
				continue
			}

			if filter != nil && !filter(entry) {
				continue
			}

			delete(entry, "time")

			mu.Lock()

			if _, ok := entry["error"]; ok {
				fmt.Fprint(writer, "<3> ")
			} else {
				fmt.Fprint(writer, "<6> ")
			}

			b, err := json.Marshal(entry)
			if err != nil {
				fmt.Fprintf(writer, "Failed to marshal log entry: %s.\n", err)
			}

			fmt.Fprintln(writer, string(b)+"\n")
			mu.Unlock()
		}

		close(done)
	}()

	return JournaldWriter{WriteCloser: w}
}
