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
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli/script"
	"github.com/stratumn/alice/core/netutil"
	"github.com/stratumn/alice/release"
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
	config := NewConfigSet().Configs()
	delete(config, "cli")

	_, err := New(config)

	if got, want := errors.Cause(err), ErrInvalidConfig; got != want {
		t.Errorf("New(config): error = %s want %s", got, want)
	}
}

func TestCli_Connect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	addr := testAddress()

	serverCh := make(chan error, 2)
	go func() {
		serverCh <- testServer(ctx, t, addr)
	}()

	config := NewConfigSet().Configs()
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	c.Console().Writer = ioutil.Discard

	err = c.Connect(ctx, addr)
	if err != nil {
		t.Errorf("c.Connect(ctx, addr): error: %s", err)
	}

	if got, want := c.Address(), addr; got != want {
		t.Errorf("c.Address() = %s want %s", got, want)
	}

	err = c.Disconnect()
	if err != nil {
		t.Errorf("c.Disconnect(): error: %s", err)
	}

	cancel()
	if err := <-serverCh; err != nil {
		t.Errorf("testServer(ctx, t, addr): error: %s", err)
	}
}

func TestCli_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	addr := testAddress()

	serverCh := make(chan error, 2)
	go func() {
		serverCh <- testServer(ctx, t, addr)
	}()

	config := NewConfigSet().Configs()
	conf := config["cli"].(Config)
	conf.APIAddress = addr
	config["cli"] = conf

	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	c.Console().Writer = ioutil.Discard

	promptCh := make(chan struct{})
	c.(*cli).prompt = func(context.Context, CLI) {
		close(promptCh)
	}

	runCh := make(chan struct{})
	go func() {
		c.Run(ctx)
		close(runCh)
	}()

	select {
	case <-promptCh:
	case <-time.After(10 * time.Second):
		t.Error("c.Run(ctx) did start the prompt")
	}

	cancel()

	select {
	case <-runCh:
	case <-time.After(10 * time.Second):
		t.Error("c.Run(ctx) did not return")
	}

	if err := <-serverCh; err != nil {
		t.Errorf("testServer(ctx, t, addr): error: %s", err)
	}
}

func TestCli_Eval(t *testing.T) {
	config := NewConfigSet().Configs()
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	buf := bytes.NewBuffer(nil)
	c.Console().Writer = buf

	ctx := context.Background()
	err = c.Eval(ctx, "cli-version")
	if err != nil {
		t.Errorf(`c.Eval("cli-version"): error: %s`, err)
	}

	got := buf.String()
	want := release.Version + "@" + release.GitCommit + "\n"

	if got != want {
		t.Errorf("c.Exec(ctx, \"cli-version\") =>\n%s\nwant\n\n%s", got, want)
	}
}

func TestCli_Eval_error(t *testing.T) {
	config := NewConfigSet().Configs()
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	ctx := context.Background()
	err = errors.Cause(c.Eval(ctx, "version"))

	if err != ErrInvalidInstr {
		t.Errorf("c.Exec(ctx, \"version\"): error = %v want %v", err, ErrInvalidInstr)
	}
}

func TestCli_Exec(t *testing.T) {
	config := NewConfigSet().Configs()
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	buf := bytes.NewBuffer(nil)
	c.Console().Writer = buf

	ctx := context.Background()
	c.Exec(ctx, "cli-version")

	got := buf.String()
	want := release.Version + "@" + release.GitCommit + "\n"

	if got != want {
		t.Errorf("c.Exec(ctx, \"cli-version\") =>\n%s\nwant\n\n%s", got, want)
	}
}

func TestCli_Exec_error(t *testing.T) {
	config := NewConfigSet().Configs()
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	buf := bytes.NewBuffer(nil)
	c.Console().Writer = buf

	ctx := context.Background()
	c.Exec(ctx, "version")

	got := buf.String()
	want := ansiRed + "Error: 1:1: version: the instruction is invalid.\n" + ansiReset

	if got != want {
		t.Errorf("c.Exec(ctx, \"version\") =>\n%s\nwant\n\n%s", got, want)
	}
}

func TestCli_Suggest(t *testing.T) {
	config := NewConfigSet().Configs()
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	sugs := c.Suggest(basicContentMock("api"))
	if got, want := len(sugs), 3; got != want {
		t.Fatalf("len(sugs) = %d want %d", got, want)
	}

	if got, want := sugs[0].Text, "api-address"; got != want {
		t.Errorf("sugs[0].Text = %v want %v", got, want)
	}

	if got, want := sugs[1].Text, "api-connect"; got != want {
		t.Errorf("sugs[1].Text = %v want %v", got, want)
	}

	if got, want := sugs[2].Text, "api-disconnect"; got != want {
		t.Errorf("sugs[2].Text = %v want %v", got, want)
	}
}

func TestCli_DidJustExecute(t *testing.T) {
	config := NewConfigSet().Configs()
	c, err := New(config)
	if err != nil {
		t.Errorf("New(config): error: %s", err)
	}

	c.Console().Writer = ioutil.Discard

	ctx := context.Background()

	if got, want := c.DidJustExecute(), false; got != want {
		t.Errorf("c.DidJustExecute() = %v want %v", got, want)
	}

	c.Exec(ctx, "cli-version")

	if got, want := c.DidJustExecute(), true; got != want {
		t.Errorf("c.DidJustExecute() = %v want %v", got, want)
	}

	if got, want := c.DidJustExecute(), false; got != want {
		t.Errorf("c.DidJustExecute() = %v want %v", got, want)
	}
}

func sym(s string) *script.SExp {
	return &script.SExp{Type: script.SExpSym, Str: s}
}

func TestCli_resolver(t *testing.T) {
	got, err := cliResolver(sym("a"))
	if err != nil {
		t.Error(`cliResolver(sym("a")): error: `, err)
	} else {
		if got, want := got.String(), `("a")`; got != want {
			t.Errorf(`cliResolver(sym("a")): v = %s want %s `, got, want)
		}
	}

	if _, err := cliResolver(sym("$a")); err == nil {
		t.Error(`cliResolver(sym("$a")): did not get an error`)
	}
}
