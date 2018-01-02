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

package log

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
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
	if err != nil {
		t.Error("w.Write([]byte(text)): error: ", err)
	}

	w.Close()
	for _, d := range doners {
		<-d
	}
	doners = nil

	got := trimLines(buf.String())
	want := trimLines(test.want)

	if got != want {
		t.Errorf("w.Write([]byte(test.text)) => \n\n%s\n\nwant\n\n%s ", got, want)
	}
}

func trimLines(s string) string {
	lines := strings.Split(strings.Trim(s, "\n"), "\n")
	for i := range lines {
		lines[i] = strings.Trim(lines[i], " \t")
	}

	return strings.Join(lines, "\n")
}
