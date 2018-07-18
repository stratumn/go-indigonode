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

package log

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type writerTest struct {
	name   string
	create func(io.Writer, sync.Locker) io.WriteCloser
	text   string
	want   string
}

var writerTests = []writerTest{{
	"not filtered",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewFilteredWriter(w, lock, nil)
	},
	`{
		"event": "updatePeer",
		"peerID": "QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	`{"event":"updatePeer","peerID":"QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW","service":"kaddht","system":"dht","time":"2017-11-26T00:59:16.980455Z"}`,
}, {
	"filtered",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewFilteredWriter(w, lock, func(entry map[string]interface{}) bool {
			return false
		})
	},
	`{
		"event": "updatePeer",
		"peerID": "QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	"",
}, {
	"invalid",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewFilteredWriter(w, lock, nil)
	},
	"test",
	"Failed to unmarshal log entry: invalid character 'e' in literal true (expecting 'r').",
}, {
	"pretty info",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewPrettyWriter(w, lock, nil, false)
	},
	`{
		"event": "updatePeer",
		"peerID": "QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	`00:59:16.980 INF dht#updatePeer { 
		"peerID": "QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
		"service": "kaddht" 
	}`,
}, {
	"pretty error",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewPrettyWriter(w, lock, nil, false)
	},
	`{
		"event": "updatePeer",
		"error": "oh no",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	`00:59:16.980 ERR dht#updatePeer { 
		"error": "oh no",
		"service": "kaddht" 
	}`,
}, {
	"pretty filtered",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewPrettyWriter(w, lock, func(entry map[string]interface{}) bool {
			return false
		}, false)
	},
	`{
		"event": "updatePeer",
		"peerID": "QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	"",
}, {
	"pretty invalid",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewPrettyWriter(w, lock, nil, false)
	},
	"test",
	"Failed to unmarshal log entry: invalid character 'e' in literal true (expecting 'r').",
}, {
	"journald",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewJournaldWriter(w, lock, nil)
	},
	`{
		"event": "updatePeer",
		"peerID": "QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	`<6> {"event":"updatePeer","peerID":"QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW","service":"kaddht","system":"dht"}`,
}, {
	"journald error",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewJournaldWriter(w, lock, nil)
	},
	`{
		"event": "updatePeer",
		"error": "oh no",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	`<3> {"error":"oh no","event":"updatePeer","service":"kaddht","system":"dht"}`,
}, {
	"journald filtered",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewJournaldWriter(w, lock, func(entry map[string]interface{}) bool {
			return false
		})
	},
	`{
		"event": "updatePeer",
		"peerID": "QmSSn4cWZZS8oie3sC1MVPrcDfQPFa6LsbvUeyRED3UwBW",
		"service": "kaddht",
		"system": "dht",
		"time": "2017-11-26T00:59:16.980455Z"
	}`,
	"",
}, {
	"journald invalid",
	func(w io.Writer, lock sync.Locker) io.WriteCloser {
		return NewJournaldWriter(w, lock, nil)
	},
	"test",
	"<3> Failed to unmarshal log entry: invalid character 'e' in literal true (expecting 'r').",
}}

func TestWriters(t *testing.T) {
	for _, tt := range writerTests {
		t.Run(tt.name, func(t *testing.T) {
			testWriter(t, tt)
		})
	}
}

func testWriter(t *testing.T, test writerTest) {
	buf := bytes.NewBuffer(nil)
	w := test.create(buf, &sync.Mutex{})

	text := strings.Join(strings.Split(test.text, "\n"), " ") + "\n"
	_, err := w.Write([]byte(text))
	assert.NoError(t, err, "w.Write([]byte(text))")

	w.Close()
	for _, d := range doners {
		<-d
	}
	doners = nil

	got := trimLines(buf.String())
	want := trimLines(test.want)

	assert.Equal(t, want, got, "w.Write([]byte(test.text))")
}

func trimLines(s string) string {
	lines := strings.Split(strings.Trim(s, "\n"), "\n")
	for i := range lines {
		lines[i] = strings.Trim(lines[i], " \t")
	}

	return strings.Join(lines, "\n")
}
