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

package cli

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/netutil"
	"github.com/stratumn/go-indigonode/release"
	"github.com/stratumn/go-indigonode/script"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func testAddress() string {
	return fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", netutil.RandomPort())
}

func testServer(ctx context.Context, t *testing.T, address string) error {
	gs := grpc.NewServer()
	reflection.Register(gs)

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

func TestCli_New_invalidConfig(t *testing.T) {
	config := NewConfigurableSet().Configs()
	delete(config, "cli")

	_, err := New(config)

	assert.Equal(t, ErrInvalidConfig, errors.Cause(err), "invalid error")
}

func TestCli_Connect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	addr := testAddress()

	serverCh := make(chan error, 2)
	go func() {
		serverCh <- testServer(ctx, t, addr)
	}()

	assert := assert.New(t)

	config := NewConfigurableSet().Configs()
	c, err := New(config)

	assert.NoError(err, "New(config)")

	c.Console().Writer = ioutil.Discard

	err = c.Connect(ctx, addr)
	assert.NoError(err, "c.Connect(ctx, addr)")

	assert.Equal(addr, c.Address(), "c.Address()")

	err = c.Disconnect()
	assert.NoError(err, "c.Disconnect()")

	cancel()
	err = <-serverCh
	assert.NoError(err, "testServer(ctx, t, addr)")
}

func TestCli_Start(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	addr := testAddress()

	serverCh := make(chan error, 2)
	go func() {
		serverCh <- testServer(ctx, t, addr)
	}()

	assert := assert.New(t)

	config := NewConfigurableSet().Configs()
	conf := config["cli"].(Config)
	conf.APIAddress = addr
	config["cli"] = conf

	c, err := New(config)
	assert.NoError(err, "New(config) should not fail")

	c.Console().Writer = ioutil.Discard

	promptCh := make(chan struct{})
	c.(*cli).prompt = func(context.Context, CLI) {
		close(promptCh)
	}

	runCh := make(chan struct{})
	go func() {
		c.Start(ctx)
		close(runCh)
	}()

	select {
	case <-promptCh:
	case <-time.After(10 * time.Second):
		assert.Fail("c.Start(ctx) did not start the prompt")
	}

	cancel()

	select {
	case <-runCh:
	case <-time.After(10 * time.Second):
		assert.Fail("c.Run(ctx) did not return")
	}

	err = <-serverCh
	assert.NoError(err, "testServer(ctx, t, addr)")
}

func TestCli_Exec(t *testing.T) {
	assert := assert.New(t)

	config := NewConfigurableSet().Configs()
	c, err := New(config)
	assert.NoError(err, "New(config)")

	buf := bytes.NewBuffer(nil)
	c.Console().Writer = buf

	ctx := context.Background()
	err = c.Exec(ctx, "cli-version")
	assert.NoError(err, "c.Exec(ctx, \"cli-version\")")

	got := buf.String()
	want := release.Version + "@" + release.GitCommit + "\n"

	assert.Equal(want, got, "c.Exec(ctx, \"cli-version\")")
}

func TestCli_Exec_error(t *testing.T) {
	config := NewConfigurableSet().Configs()
	c, err := New(config)
	assert.NoError(t, err, "New(config)")

	c.Console().Writer = ioutil.Discard

	ctx := context.Background()
	err = errors.Cause(c.Exec(ctx, "version"))

	assert.Equal(t, script.ErrUnknownFunc, err, "c.Exec(ctx, \"version\")")
}

func TestCli_Run(t *testing.T) {
	config := NewConfigurableSet().Configs()
	c, err := New(config)
	assert.NoError(t, err, "New(config)")

	buf := bytes.NewBuffer(nil)
	c.Console().Writer = buf

	ctx := context.Background()
	c.Exec(ctx, "cli-version")

	got := buf.String()
	want := release.Version + "@" + release.GitCommit + "\n"

	assert.Equal(t, want, got, "c.Run(ctx, \"cli-version\")")
}

func TestCli_Run_error(t *testing.T) {
	config := NewConfigurableSet().Configs()
	c, err := New(config)
	assert.NoError(t, err, "New(config)")

	buf := bytes.NewBuffer(nil)
	c.Console().Writer = buf

	ctx := context.Background()
	c.Run(ctx, "version")

	got := buf.String()
	want := ansiRed + "Error: 1:1: version: unknown function.\n" + ansiReset

	assert.Equal(t, want, got, "c.Run(ctx, \"version\")")
}

func TestCli_Suggest(t *testing.T) {
	assert := assert.New(t)

	config := NewConfigurableSet().Configs()
	c, err := New(config)
	assert.NoError(err, "New(config)")

	sugs := c.Suggest(basicContentMock("api"))
	require.Equal(t, 3, len(sugs), "len(sugs)")

	assert.Equal("api-address", sugs[0].Text, "sugs[0].Text")
	assert.Equal("api-connect", sugs[1].Text, "sugs[1].Text")
	assert.Equal("api-disconnect", sugs[2].Text, "sugs[2].Text")
}

func TestCli_DidJustExecute(t *testing.T) {
	assert := assert.New(t)

	config := NewConfigurableSet().Configs()
	c, err := New(config)
	assert.NoError(err, "New(config)")

	c.Console().Writer = ioutil.Discard

	ctx := context.Background()

	assert.False(c.DidJustExecute(), "c.DidJustExecute()")

	c.Run(ctx, "cli-version")

	assert.True(c.DidJustExecute(), "c.DidJustExecute()")
	assert.False(c.DidJustExecute(), "c.DidJustExecute()")
}

func sym(s string) script.SExp {
	return script.Symbol(s, script.Meta{})
}

func TestResolver(t *testing.T) {
	c := script.NewClosure()

	got, err := Resolver(c, sym("a"))
	assert.NoError(t, err, "cliResolver(sym(\"a\"))")
	assert.Equal(t, "\"a\"", got.String(), "cliResolver(sym(\"a\"))")

	_, err = Resolver(c, sym("$a"))
	assert.Error(t, err, "cliResolver(sym(\"$a\"))")
}
