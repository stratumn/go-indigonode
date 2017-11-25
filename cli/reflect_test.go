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
	"context"
	"io/ioutil"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/stratumn/alice/core/netutil"
	"github.com/stratumn/alice/grpc/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type reflectorTest struct {
	name        string
	r           Reflector
	supported   *desc.FieldDescriptor
	unsupported *desc.FieldDescriptor
	arg         []reflectorParseTest
	flag        []reflectorParseTest
	pretty      []reflectorPrettyTest
}

type reflectorParseTest struct {
	text  string
	want  interface{}
	fails bool
}

type reflectorPrettyTest struct {
	val   interface{}
	want  string
	fails bool
}

func TestReflectors(t *testing.T) {
	msg, err := desc.LoadMessageDescriptorForMessage(&test.Message{})
	if err != nil {
		t.Fatal("failed to load message: ", err)
	}

	tt := []reflectorTest{{
		"string",
		NewStringReflector(),
		msg.FindFieldByName("str"),
		msg.FindFieldByName("buf"),
		[]reflectorParseTest{
			{"test", "test", false},
			{" test", "test", false},
		},
		[]reflectorParseTest{
			{"cmd --str test", "test", false},
		},
		[]reflectorPrettyTest{
			{"test", "test", false},
		},
	}, {
		"string repeated",
		NewStringReflector(),
		msg.FindFieldByName("str_repeated"),
		msg.FindFieldByName("buf"),
		[]reflectorParseTest{
			{"a", []string{"a"}, false},
			{"a,b", []string{"a", "b"}, false},
			{"a, b", []string{"a", "b"}, false},
		},
		[]reflectorParseTest{
			{"cmd --str_repeated a --str_repeated b", []string{"a", "b"}, false},
		},
		[]reflectorPrettyTest{
			{[]string{"a", "b"}, "a,b", false},
		},
	}, {
		"bool",
		NewBoolReflector(),
		msg.FindFieldByName("boolean"),
		msg.FindFieldByName("buf"),
		[]reflectorParseTest{
			{"true", true, false},
			{"True", true, false},
			{"false", false, false},
			{"fa", false, true},
		},
		[]reflectorParseTest{
			{"cmd --boolean true", true, false},
			{"cmd --boolean false", false, false},
			{"cmd --boolean bla", false, true},
			{"cmd", false, false},
		},
		[]reflectorPrettyTest{
			{true, "true", false},
			{false, "false", false},
		},
	}, {
		"bool repeated",
		NewBoolReflector(),
		msg.FindFieldByName("boolean_repeated"),
		msg.FindFieldByName("buf"),
		[]reflectorParseTest{
			{"true,false", []bool{true, false}, false},
			{"true , false", []bool{true, false}, false},
			{"True", []bool{true}, false},
			{"fa", []bool{}, true},
		},
		[]reflectorParseTest{
			{"cmd --boolean_repeated true,false", []bool{true, false}, false},
			{"cmd --boolean_repeated true --boolean_repeated false", []bool{true, false}, false},
			{"cmd", []bool{}, false},
		},
		[]reflectorPrettyTest{
			{[]bool{true, false}, "true,false", false},
		},
	}, {
		"uint32",
		NewUint32Reflector(),
		msg.FindFieldByName("u32"),
		msg.FindFieldByName("buf"),
		[]reflectorParseTest{
			{"10", uint32(10), false},
		},
		[]reflectorParseTest{
			{"cmd --u32 10", uint32(10), false},
			{"cmd --u32 false", uint32(0), true},
			{"cmd", uint32(0), false},
		},
		[]reflectorPrettyTest{
			{uint32(10), "10", false},
		},
	}, {
		"bytes",
		NewBytesReflector(),
		msg.FindFieldByName("buf"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"ff00", []byte{255, 0}, false},
			{"zz", []byte{}, true},
		},
		[]reflectorParseTest{
			{"cmd --buf ff00", []byte{255, 0}, false},
			{"cmd", []byte{}, false},
			{"cmd --buf zz", []byte{}, true},
		},
		[]reflectorPrettyTest{
			{[]byte{255, 0}, "ff00", false},
		},
	}, {
		"bytes repeated",
		NewBytesReflector(),
		msg.FindFieldByName("buf_repeated"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"ff00,ff", [][]byte{{255, 0}, {255}}, false},
			{"ff00, ff", [][]byte{{255, 0}, {255}}, false},
			{"zz", [][]byte{}, true},
		},
		[]reflectorParseTest{
			{"cmd --buf_repeated ff00 --buf_repeated ff", [][]byte{{255, 0}, {255}}, false},
			{"cmd", [][]byte{}, false},
			{"cmd --buf_repeated zz", [][]byte{}, true},
		},
		[]reflectorPrettyTest{
			{[][]byte{{255, 0}, {255}}, "ff00,ff", false},
		},
	}, {
		"enum",
		NewEnumReflector(),
		msg.FindFieldByName("enumeration"),
		msg.FindFieldByName("buf"),
		[]reflectorParseTest{
			{"A", int32(0), false},
			{"B", int32(1), false},
			{"1", int32(0), true},
		},
		[]reflectorParseTest{
			{"cmd --enumeration b", int32(1), false},
			{"cmd", int32(0), false},
			{"cmd --enumeration zz", int32(0), true},
		},
		[]reflectorPrettyTest{
			{int32(1), "B", false},
		},
	}, {
		"duration",
		NewDurationReflector(),
		msg.FindFieldByName("duration"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"1m", int64(time.Minute), false},
			{"1", int64(0), true},
		},
		[]reflectorParseTest{
			{"cmd --duration 1m", int64(time.Minute), false},
			{"cmd", int64(0), false},
			{"cmd --duration zz", int64(0), true},
		},
		[]reflectorPrettyTest{
			{int64(time.Minute), "1m0s", false},
		},
	}, {
		"duration repeated",
		NewDurationReflector(),
		msg.FindFieldByName("duration_repeated"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"1m, 1s", []int64{int64(time.Minute), int64(time.Second)}, false},
			{"a,b", []int64{}, true},
		},
		[]reflectorParseTest{
			{
				"cmd --duration_repeated 1m --duration_repeated 1s",
				[]int64{int64(time.Minute), int64(time.Second)},
				false,
			},
			{"cmd", []int64{}, false},
			{"cmd --duration_repeated zz", []int64{}, true},
		},
		[]reflectorPrettyTest{
			{[]int64{int64(time.Minute), int64(time.Second)}, "1m0s,1s", false},
		},
	}, {
		"base58",
		NewBase58Reflector(),
		msg.FindFieldByName("base58"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9", []byte{
				18, 32, 109, 76, 36, 229, 32, 176,
				53, 31, 169, 189, 190, 37, 119, 25,
				79, 187, 126, 16, 211, 82, 93, 216,
				194, 134, 220, 138, 66, 252, 106, 50,
				58, 192,
			}, false},
			{"!1", []byte{}, true},
		},
		[]reflectorParseTest{
			{"cmd --base58 QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9", []byte{
				18, 32, 109, 76, 36, 229, 32, 176,
				53, 31, 169, 189, 190, 37, 119, 25,
				79, 187, 126, 16, 211, 82, 93, 216,
				194, 134, 220, 138, 66, 252, 106, 50,
				58, 192,
			}, false},
			{"cmd", []byte{}, false},
			{"cmd --base58 !", []byte{}, true},
		},
		[]reflectorPrettyTest{
			{[]byte{
				18, 32, 109, 76, 36, 229, 32, 176,
				53, 31, 169, 189, 190, 37, 119, 25,
				79, 187, 126, 16, 211, 82, 93, 216,
				194, 134, 220, 138, 66, 252, 106, 50,
				58, 192,
			}, "QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9", false},
		},
	}, {
		"bytesize",
		NewBytesizeReflector(),
		msg.FindFieldByName("bytesize"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"1K", uint64(1024), false},
			{"1", uint64(0), true},
		},
		[]reflectorParseTest{
			{"cmd --bytesize 1K", uint64(1024), false},
			{"cmd", uint64(0), false},
			{"cmd --bytesize zz", uint64(0), true},
		},
		[]reflectorPrettyTest{
			{uint64(1024), "1K", false},
		},
	}, {
		"byterate",
		NewByterateReflector(),
		msg.FindFieldByName("byterate"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"1K/s", uint64(1024), false},
			{"1/s", uint64(0), true},
		},
		[]reflectorParseTest{
			{"cmd --byterate 1K/s", uint64(1024), false},
			{"cmd", uint64(0), false},
			{"cmd --byterate zz", uint64(0), true},
		},
		[]reflectorPrettyTest{
			{uint64(1024), "1K/s", false},
		},
	}, {
		"maddr",
		NewMaddrReflector(),
		msg.FindFieldByName("multiaddr"),
		msg.FindFieldByName("boolean"),
		[]reflectorParseTest{
			{"/ip4/127.0.0.1/tcp/80", []byte{
				4, 127, 0, 0, 1, 6, 0, 80,
			}, false},
			{"!1", []byte{}, true},
		},
		[]reflectorParseTest{
			{"cmd --multiaddr /ip4/127.0.0.1/tcp/80", []byte{
				4, 127, 0, 0, 1, 6, 0, 80,
			}, false},
			{"cmd", []byte{}, false},
			{"cmd --multiaddr !", []byte{}, true},
		},
		[]reflectorPrettyTest{
			{[]byte{
				4, 127, 0, 0, 1, 6, 0, 80,
			}, "/ip4/127.0.0.1/tcp/80", false},
		},
	}}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			testReflector(t, test)
		})
	}
}

func testReflector(t *testing.T, test reflectorTest) {
	if got, want := test.r.Supports(test.supported), true; got != want {
		t.Errorf("test.r.Supports(test.supported) = %v want %v", got, want)
	}

	if got, want := test.r.Supports(test.unsupported), false; got != want {
		t.Errorf("test.r.Supports(unsupported) = %v want %v", got, want)
	}

	testReflectorArg(t, test)
	testReflectorFlag(t, test)
	testReflectorPretty(t, test)
}

func testReflectorArg(t *testing.T, test reflectorTest) {
	if argReflector, ok := test.r.(ArgReflector); ok {
		for _, argTest := range test.arg {
			v, err := argReflector.Parse(test.supported, argTest.text)
			if err != nil && !argTest.fails {
				t.Errorf(
					"argReflector.Parse(test.supported, %q): error: %s",
					argTest.text, err,
				)
			}
			if got, want := v, argTest.want; !reflect.DeepEqual(got, want) {
				t.Errorf(
					"argReflector.Parse(test.supported, %q): v = %v want %v",
					argTest.text, got, want,
				)
			}
		}
	}
}

func testReflectorFlag(t *testing.T, test reflectorTest) {
	if flagReflector, ok := test.r.(FlagReflector); ok {
		for _, flagTest := range test.flag {
			f := pflag.NewFlagSet(test.name, pflag.ContinueOnError)
			flagReflector.Flag(test.supported, f)

			args := strings.Split(flagTest.text, " ")
			f.Parse(args)

			v, err := flagReflector.ParseFlag(test.supported, f)
			if err != nil && !flagTest.fails {
				t.Errorf(
					"flagReflector.ParseFlags(%q, f): error: %s",
					args, err,
				)
			}
			if got, want := v, flagTest.want; !reflect.DeepEqual(got, want) {
				t.Errorf(
					"flagReflector.ParseFlags(%q, f): v = %v want %v",
					args, got, want,
				)
			}
		}
	}
}

func testReflectorPretty(t *testing.T, test reflectorTest) {
	if resReflector, ok := test.r.(ResponseReflector); ok {
		for _, prettyTest := range test.pretty {
			s, err := resReflector.Pretty(test.supported, prettyTest.val)
			if err != nil && !prettyTest.fails {
				t.Errorf(
					"resReflector.Pretty(test.supported, %v): error: %s",
					prettyTest.val, err,
				)
			}
			if got, want := s, prettyTest.want; got != want {
				t.Errorf(
					"resReflector.Pretty(test.supported, %v): v = %v want %v",
					prettyTest.val, got, want,
				)
			}
		}
	}
}

type reflectServer struct{}

func (reflectServer) UnaryReq(ctx context.Context, req *test.Message) (*test.Message, error) {
	return req, nil
}

func (reflectServer) ServerStream(req *test.Message, ss test.Test_ServerStreamServer) error {
	return ss.Send(req)
}

func (reflectServer) NoExt(ctx context.Context, req *test.Message) (*test.Message, error) {
	return req, nil
}

func testReflectServer(ctx context.Context, t *testing.T, address string) error {
	gs := grpc.NewServer()
	reflection.Register(gs)

	test.RegisterTestServer(gs, reflectServer{})

	lis, err := netutil.Listen(address)
	if err != nil {
		return err
	}

	ch := make(chan error, 1)
	go func() {
		ch <- gs.Serve(lis)
	}()

	select {
	case err = <-ch:
	case <-ctx.Done():
		gs.GracefulStop()
	}

	if err != nil {
		err = errors.Cause(err)

		if e, ok := err.(*net.OpError); ok {
			if e.Op == "accept" {
				// Normal error.
				return nil
			}
		}
	}

	return err
}

type reflectorServerTest struct {
	name string
	cmd  string
	want string
	err  error
}

var serverReflectorTT = []reflectorServerTest{{
	"unary",
	"test-unaryreq hello --boolean true --bytesize_repeated 1k --bytesize_repeated 1m",
	`
NOEXT
REQUIRED FIELD            hello
STRING FIELD
STRING REPEATED FIELD
BOOL FIELD                true
BOOL REPEATED FIELD
UINT32 FIELD              0
UINT32 REPEATED FIELD
BYTES FIELD
BYTES REPEATED FIELD
ENUM FIELD                A
ENUM REPEATED FIELD
BASE58 FIELD
BASE58 REPEATED FIELD
MULTIADDR FIELD
MULTIADDR REPEATED FIELD
DURATION FIELD            0s
DURATION REPEATED FIELD
BYTESIZE FIELD            0
BYTESIZE REPEATED FIELD   1K,1M
BYTERATE FIELD            0/s
BYTERATE REPEATED FIELD
`,
	nil,
}, {
	"stream",
	"test-serverstream hello",
	`
NOEXT | REQUIRED FIELD | STRING FIELD | STRING REPEATED FIELD | BOOL FIELD ...
------+----------------+--------------+-----------------------+------------+--
      | hello          |              |                       | false      ...
`,
	nil,
}, {
	"missing arg",
	"test-unaryreq",
	"",
	errUse,
}, {
	"extra arg",
	"test-unaryreq a b",
	"",
	errUse,
}, {
	"invalid flag",
	"test-unaryreq --boolean 1",
	"",
	errUse,
}}

func TestServerReflector_Reflect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	addr := testAddress()

	serverCh := make(chan error, 2)
	go func() {
		serverCh <- testReflectServer(ctx, t, addr)
	}()

	config := NewConfigSet().Configs()
	conf := config["cli"].(Config)
	conf.APIAddress = addr
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	// Set truncate width.
	c.(*cli).reflector = NewServerReflector(c.Console(), 78)

	c.Console().Writer = ioutil.Discard

	err = c.Connect(ctx, addr)
	if err != nil {
		t.Errorf("c.Connect(ctx, addr): error: %s", err)
	}

	for _, test := range serverReflectorTT {
		t.Run(test.name, func(t *testing.T) {
			testServerReflectorReflect(ctx, t, c, test)
		})
	}

	cancel()

	if err := <-serverCh; err != nil {
		t.Errorf("testServer(ctx, t, addr): error: %s", err)
	}
}

var (
	errAny = errors.New("any error")
	errUse = errors.New("usage error")
)

func testServerReflectorReflect(ctx context.Context, t *testing.T, c CLI, test reflectorServerTest) {
	buf := bytes.NewBuffer(nil)
	c.Console().Writer = buf

	err := errors.Cause(c.Eval(ctx, test.cmd))

	switch {
	case test.err == errAny && err != nil:
		// Pass.
	case test.err == errUse:
		if _, ok := err.(*UseError); !ok {
			t.Errorf("%s: error = %v want %v", test.cmd, err, test.err)
		}
	case err != test.err:
		t.Errorf("%s: error = %v want %v", test.cmd, err, test.err)
	}

	got, want := trimLines(buf.String()), trimLines(test.want)

	if got != want {
		t.Errorf("%s =>\n\n%s\n\nwant\n\n%s\n", test.cmd, got, want)
	}
}

func trimLines(s string) string {
	lines := strings.Split(strings.Trim(s, "\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}

	return strings.Join(lines, "\n")
}
