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

//go:generate mockgen -package mockcli -destination mockcli/mockcli.go github.com/stratumn/alice/cli CLI

/*
Package cli defines types for Alice's command line interface.

It comes with only of handful of bultin commands. The bulk of commands are
reflected from the API.

The main type is the CLI struct, which wraps everything needed to run the
command line interface. It can, amongst other things, make suggestions for
auto-completion and connect to an Alice node.

The CLI needs a Console and a Prompt. The console is responsible for rendering
text. The Prompt is responsible for getting user input.

The Prompt is also responsible for calling the CLI's Exec method to execute
a command, and the Suggest method to make suggestions for auto-completion.

The Suggest method should be given a Content which allows it to read the
current text and returns a slice suggestions with the type Suggest.

BasicCmd is a type that allows creating simple commands that cover most use
cases.
*/
package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli/script"
	"github.com/stratumn/alice/core/cfg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	manet "gx/ipfs/QmX3U3YXCQ6UYBxq2LVWF8dARS1hPUTEYLrSx654Qyxyw6/go-multiaddr-net"
	ma "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
)

// List all the builtin CLI commands here.
var cmds = []Cmd{
	Addr,
	Bang{},
	Block,
	Connect,
	Car,
	Cdr,
	Cons,
	Disconnect,
	Echo,
	Eval,
	Exit,
	Help,
	If,
	Lambda,
	Let,
	Quote,
	Unless,
	Version,
}

// art is displayed by the init script.
const art = "      .o.       oooo   o8o\n" +
	"     .888.      `888   `\\\"'\n" +
	"    .8\\\"888.      888  oooo   .ooooo.   .ooooo.\n" +
	"   .8' `888.     888  `888  d88' `\\\"Y8 d88' `88b\n" +
	"  .88ooo8888.    888   888  888       888ooo888\n" +
	" .8'     `888.   888   888  888   .o8 888    .o\n" +
	"o88o     o8888o o888o o888o `Y8bod8P' `Y8bod8P'"

// initScript is executed when the CLI is launched.
const initScript = `
echo
echo --log info "` + art + `"
echo
cli-version --git-commit-length 7
echo

; Only visible in debug mode.
echo --log debug Debuf output is enabled.

; Connect to the API.
(unless
	(api-connect)
	(block
		(echo)
		(echo Looks like the API is offline.)
		(echo You can try to connect again using "'api-connect'")))

echo
echo Enter "'help'" to list available commands.
echo Enter "'exit'" to quit the command line interface.
echo Use the tab key for auto-completion.
echo
`

// Available prompts.
var promptsMu sync.Mutex
var prompts = map[string]func(context.Context, CLI){}

// registerPrompt registers a prompt.
func registerPrompt(name string, run func(context.Context, CLI)) {
	promptsMu.Lock()
	prompts[name] = run
	promptsMu.Unlock()
}

// Resolver resolves the name of the symbol if it doesn't begin with a dollar
// sign.
func Resolver(sym script.SExp) (script.SExp, error) {
	name := sym.MustSymbolVal()
	meta := sym.Meta()

	if strings.HasPrefix(name, "$") {
		return nil, errors.Wrapf(
			script.ErrSymNotFound,
			"%d:%d: %s",
			meta.Line,
			meta.Offset,
			name,
		)
	}

	return script.ResolveName(sym)
}

// cli implements the command line interface.
type cli struct {
	conf Config

	cons   *Console
	prompt func(context.Context, CLI)

	closure    *script.Closure
	parser     *script.Parser
	reflector  ServerReflector
	staticCmds []Cmd
	allCmds    []Cmd

	addr string
	conn *grpc.ClientConn

	// Hack to hide suggestions after executing a command.
	executed bool
}

// New create a new command line interface.
func New(configSet cfg.ConfigSet) (CLI, error) {
	config, ok := configSet["cli"].(Config)
	if !ok {
		return nil, errors.WithStack(ErrInvalidConfig)
	}

	cons := NewConsole(os.Stdout, config.EnableColorOutput)

	prompt, ok := prompts[config.PromptBackend]
	if !ok {
		return nil, errors.WithStack(ErrPromptNotFound)
	}

	// Create the top closure with env variables.
	closure := script.NewClosure(
		script.OptEnv("$", os.Environ()),
		script.OptResolver(Resolver),
	)

	c := cli{
		conf:       config,
		cons:       cons,
		prompt:     prompt,
		closure:    closure,
		reflector:  NewServerReflector(cons, 0),
		staticCmds: cmds,
		allCmds:    cmds,
		addr:       config.APIAddress,
	}

	scanner := script.NewScanner(script.OptErrorHandler(c.PrintError))
	c.parser = script.NewParser(scanner)

	sort.Slice(c.staticCmds, func(i, j int) bool {
		return c.staticCmds[i].Name() < c.staticCmds[j].Name()
	})

	c.cons.SetDebug(config.EnableDebugOutput)

	return &c, nil
}

// Start starts the command line interface until the user kills it.
func (c *cli) Start(ctx context.Context) {
	defer func() {
		if err := c.Disconnect(); err != nil {
			if errors.Cause(err) != ErrDisconnected {
				c.cons.Debugf("Could not disconnect: %s.", err)
			}
		}
	}()

	c.Run(ctx, initScript)
	c.prompt(ctx, c)
}

// Config returns the configuration.
func (c *cli) Config() Config {
	return c.conf
}

// Console returns the console.
func (c *cli) Console() *Console {
	return c.cons
}

// Commands returns all the commands.
func (c *cli) Commands() []Cmd {
	return c.allCmds
}

// Address returns the address of the API server.
func (c *cli) Address() string {
	return c.addr
}

// dialOpts builds the options to dial the API.
func (c *cli) dialOpts(ctx context.Context, address string) ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{
		// Makes the call block till the connection is made.
		grpc.WithBlock(),
		// Use multiaddr dialer.
		grpc.WithDialer(func(string, time.Duration) (net.Conn, error) {
			addr, err := ma.NewMultiaddr(address)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			dialer := manet.Dialer{}
			conn, err := dialer.DialContext(ctx, addr)
			return conn, errors.WithStack(err)
		}),
	}

	cert, override := c.conf.TLSCertificateFile, c.conf.TLSServerNameOverride
	if cert != "" {
		c.cons.Successln("TLS is enabled.")
		if override != "" {
			c.cons.Warningln("WARNING: API server authenticity is not checked because TLS server name override is enabled.")
		}
		creds, err := credentials.NewClientTLSFromFile(cert, override)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		c.cons.Warningln("WARNING: API connection is not encrypted because TLS is disabled.")
		opts = append(opts, grpc.WithInsecure())
	}

	return opts, nil
}

// Connect connects to the API server.
func (c *cli) Connect(ctx context.Context, addr string) error {
	if err := c.Disconnect(); err != nil && errors.Cause(err) != ErrDisconnected {
		return err
	}

	if addr != "" {
		c.addr = addr
	}
	if c.addr == "" {
		c.addr = c.conf.APIAddress
	}

	dto, err := time.ParseDuration(c.conf.DialTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	dialCtx, dialCancel := context.WithTimeout(ctx, dto)
	defer dialCancel()

	opts, err := c.dialOpts(dialCtx, c.addr)
	if err != nil {
		return err
	}

	// Connect to the API.
	conn, err := grpc.DialContext(dialCtx, c.addr, opts...)
	if err != nil {
		return errors.WithStack(err)
	}

	apiCmds, err := c.reflector.Reflect(ctx, conn)
	if err != nil {
		if err := conn.Close(); err != nil {
			c.cons.Debugf("Could not close connection: %s.", err)
		}
		return err
	}

	c.allCmds = append(c.staticCmds, apiCmds...)
	sort.Slice(c.allCmds, func(i, j int) bool {
		return c.allCmds[i].Name() < c.allCmds[j].Name()
	})

	c.conn = conn

	return nil
}

// Disconnect closes the API client connection.
func (c *cli) Disconnect() error {
	if c.conn == nil {
		return errors.WithStack(ErrDisconnected)
	}

	conn := c.conn
	c.conn = nil
	c.allCmds = c.staticCmds

	return errors.WithStack(conn.Close())
}

// PrintError prints the error if it isn't nil.
func (c *cli) PrintError(err error) {
	if err == nil {
		return
	}

	c.cons.Errorf("Error: %s.\n", err)

	if stack := StackTrace(err); len(stack) > 0 {
		c.cons.Debugf("%+v\n", stack)
	}

	if useError, ok := errors.Cause(err).(*UseError); ok {
		if use := useError.Use(); use != "" {
			c.cons.Print("\n" + use)
		}
	}
}

// Exec executes the given input.
func (c *cli) Exec(ctx context.Context, in string) error {
	defer func() { c.executed = true }()

	instrs, err := c.parser.Parse(in)
	if err != nil {
		return errors.WithStack(err)
	}

	call := c.callWithClosure(ctx, c.closure)

	if instrs != nil {
		list := instrs.MustToSlice()

		for _, instr := range list {
			v, err := script.Eval(c.closure.Resolve, call, instr)
			if err != nil {
				return err
			}

			if v != nil {
				// Dont't print quotes.
				if v.UnderlyingType() == script.TypeString {
					str := v.MustStringVal()
					if str == "" {
						continue
					}
					c.cons.Print(str)
					if !strings.HasSuffix(str, "\n") {
						c.cons.Println()
					}
				} else {
					c.cons.Println(fmt.Sprint(v))
				}
			}
		}
	}

	return nil
}

// Run executes the given input, handling errors and cancellation signals.
func (c *cli) Run(ctx context.Context, in string) {
	defer func() { c.executed = true }()

	execCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	done := make(chan error, 1)
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		done <- c.Exec(execCtx, in)
	}()

	var err error

	// Handle exit conditions.
	select {
	case <-sigc:
		cancel()
		err = <-done
	case err = <-done:
	}

	c.PrintError(err)
}

// Suggest finds all command suggestions.
func (c *cli) Suggest(cnt Content) []Suggest {
	var sug []Suggest

	for _, v := range c.allCmds {
		sug = append(sug, v.Suggest(cnt)...)
	}

	// Sort by text.
	sort.Slice(sug, func(i, j int) bool {
		return sug[i].Text < sug[j].Text
	})

	return sug
}

// DidJustExecute returns true the first time it is called after a command
// executed. This is a hack used by the VT100 prompt to hide suggestions
// after a command was executed.
func (c *cli) DidJustExecute() bool {
	defer func() { c.executed = false }()

	return c.executed
}

// execCmd executes the given command.
func (c *cli) execCmd(ctx *ExecContext, cmd Cmd) (script.SExp, error) {
	val, err := cmd.Exec(ctx)

	if err != nil {
		cause := errors.Cause(err)

		// If it is a usage error, add the use string to it.
		if useError, ok := cause.(*UseError); ok {
			useError.use = cmd.LongUse()
		}

		return nil, errors.Wrapf(
			err,
			"%d:%d: %s",
			ctx.Meta.Line,
			ctx.Meta.Offset,
			ctx.Name,
		)
	}

	return val, nil
}

// call handles function calls.
func (c *cli) call(ctx *ExecContext) (script.SExp, error) {
	lambda, ok := ctx.Closure.Get("$" + ctx.Name)
	if ok {
		return ExecFunc(&FuncContext{
			Ctx:               ctx.Ctx,
			Closure:           ctx.Closure,
			CallerWithClosure: c.callWithClosure,
			Name:              ctx.Name,
			Lambda:            lambda,
			Args:              ctx.Args,
		})
	}

	for _, cmd := range c.allCmds {
		if cmd.Match(ctx.Name) {
			return c.execCmd(ctx, cmd)
		}
	}

	return nil, errors.Wrapf(
		ErrInvalidInstr,
		"%d:%d: %s",
		ctx.Meta.Line,
		ctx.Meta.Offset,
		ctx.Name,
	)
}

// callWithClosure creates an S-Expression call handler bound to a closure.
func (c *cli) callWithClosure(
	ctx context.Context,
	closure *script.Closure,
) script.CallHandler {
	return func(
		resolve script.ResolveHandler,
		name string,
		args script.SCell,
		meta script.Meta,
	) (script.SExp, error) {
		return c.call(&ExecContext{
			Ctx:     ctx,
			CLI:     c,
			Name:    name,
			Closure: closure,
			Call:    c.callWithClosure(ctx, closure),
			Args:    args,
			Meta:    meta,
		})
	}
}
