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
	"bytes"
	"context"
	"io"
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
	Disconnect,
	Echo,
	Eval,
	Exit,
	Help,
	If,
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
echo --stream info "` + art + `"
echo
echo (cli-version --git-commit-length 7) :: Copyright © 2017 Stratumn SAS
echo

; Only visible in debug mode.
echo --stream debug Debuf output is enabled.

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

var (
	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("the configuration is invalid")

	// ErrPromptNotFound is returned when the requested prompt backend was
	// not found.
	ErrPromptNotFound = errors.New("the requested prompt was not found")

	// ErrDisconnected is returned when the CLI is not connected to the
	// API.
	ErrDisconnected = errors.New("the client is not connected to API")

	// ErrInvalidInstr is returned when the user entered an invalid
	// instruction.
	ErrInvalidInstr = errors.New("the instruction is invalid")

	// ErrCmdNotFound is returned when a command was not found.
	ErrCmdNotFound = errors.New("the command was not found")

	// ErrInvalidExitCode is returned when an invalid exit code was given.
	ErrInvalidExitCode = errors.New("the exit code is invalid")

	// ErrUnsupportedReflectType is returned when a type is not currently
	// supported by reflection.
	ErrUnsupportedReflectType = errors.New("the type is not currently supported by reflection")

	// ErrParse is returned when a value could not be parsed.
	ErrParse = errors.New("could not parse value")

	// ErrCarNil is retrurned when calling car on an S-Expression that doesn't
	// have one.
	ErrCarNil = errors.New("car is nil")

	// ErrCdrNil is retrurned when calling cdr on an S-Expression that doesn't
	// have one.
	ErrCdrNil = errors.New("cdr is nil")
)

// Content represents console content used to find suggestions.
type Content interface {
	// TextBeforeCursor returns all the text before the cursor.
	TextBeforeCursor() string

	// GetWordBeforeCursor returns the word before the current cursor
	// position.
	GetWordBeforeCursor() string
}

// Cmd is an interface that must be implemented by commands.
type Cmd interface {
	// Name returns the name of the command (used by `help command`
	// to find the command).
	Name() string

	// Short returns a short description of the command.
	Short() string

	// Long returns a long description of the command.
	Long() string

	// Use returns a short string showing how to use the command.
	Use() string

	// LongUse returns a long string showing how to use the command.
	LongUse() string

	// Suggest gives a chance for the command to add auto-complete
	// suggestions for the current content.
	Suggest(Content) []Suggest

	// Match returns whether the command can execute against the given
	// command name.
	Match(string) bool

	// Exec executes the given S-Expression.
	//
	// The result should be written to the writer. Other messages should
	// be written the CLI's console.
	Exec(
		context.Context,
		CLI,
		io.Writer,
		*script.Closure,
		script.Evaluator,
		*script.SExp,
	) error
}

// Suggest implements a suggestion.
type Suggest struct {
	// Text is the text that will replace the current word.
	Text string

	// Desc is a short description of the suggestion.
	Desc string
}

// UseError represents a usage error.
//
// Make it a pointer receiver so that errors.Cause() can be used to retrieve
// and modify the use string after the error is created. This is actually
// being done by the executor after a command if it returns a usage error.
type UseError struct {
	msg string
	use string
}

// NewUseError creates a new usage error.
func NewUseError(msg string) error {
	return errors.WithStack(&UseError{msg: msg})
}

// Error returns the error message.
func (err *UseError) Error() string {
	return "invalid usage: " + err.msg
}

// Use returns the usage message.
func (err *UseError) Use() string {
	return err.use
}

// CLI represents a command line interface.
type CLI interface {
	// Config returns the configuration.
	Config() Config

	// Console returns the console.
	Console() *Console

	// Commands returns all the commands.
	Commands() []Cmd

	// Address returns the address of the API server.
	Address() string

	// Connect connects to the API server.
	Connect(ctx context.Context, addr string) error

	// Disconnect closes the API client connection.
	Disconnect() error

	// Run runs the command line interface until the user kills it.
	Run(context.Context)

	// Eval executes the given input, but does not handle errors or exit
	// signals.
	Eval(ctx context.Context, in string) error

	// Exec executes a command.
	// It handles signals to cancel the command if the user presses Ctrl-C.
	Exec(ctx context.Context, in string)

	// Suggest finds all command suggestions.
	Suggest(cnt Content) []Suggest

	// PrintError prints an error if it isn't nil.
	PrintError(error)

	// DidJustExecute returns true the first time it is called after a command
	// executed. This is a hack used by the VT100 prompt to hide suggestions
	// after a command was executed.
	DidJustExecute() bool
}

// cli implements the command line interface.
type cli struct {
	conf   Config
	cons   *Console
	prompt func(context.Context, CLI)

	closure   *script.Closure
	reflector ServerReflector
	parser    *script.Parser
	cmds      []Cmd
	allCmds   []Cmd

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

	closure := script.NewClosure(
		script.OptEnv("$", os.Environ()),
		script.OptResolver(Resolver),
	)

	c := cli{
		conf:      config,
		cons:      cons,
		prompt:    prompt,
		closure:   closure,
		reflector: NewServerReflector(cons, 0),
		cmds:      cmds,
		allCmds:   cmds,
		addr:      config.APIAddress,
	}

	scanner := script.NewScanner(script.OptErrorHandler(c.PrintError))
	c.parser = script.NewParser(scanner)

	sort.Slice(c.cmds, func(i, j int) bool {
		return c.cmds[i].Name() < c.cmds[j].Name()
	})

	c.cons.SetDebug(config.EnableDebugOutput)

	return &c, nil
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

	c.allCmds = append(c.cmds, apiCmds...)
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
	c.allCmds = c.cmds

	return errors.WithStack(conn.Close())
}

// Run runs the command line interface until the user kills it.
func (c *cli) Run(ctx context.Context) {
	defer func() {
		if err := c.Disconnect(); err != nil && errors.Cause(err) != ErrDisconnected {
			c.cons.Debugf("Could not disconnect: %s.", err)
		}
	}()

	c.Exec(ctx, initScript)

	// Start the input prompt.
	c.prompt(ctx, c)
}

// evalSExp evaluate an S-Expression.
//
// If capture is false, the result of the command is printed to the console
// and an empty string is returned.
//
// Otherwise the result of the command is returned.
func (c *cli) evalSExp(
	ctx context.Context,
	closure *script.Closure,
	exp *script.SExp,
	capture bool,
) (string, error) {
	if exp == nil {
		return "", nil
	}

	if exp.Type == script.TypeStr {
		return exp.Str, nil
	}

	for _, cmd := range c.allCmds {
		if cmd.Match(exp.Str) {
			return c.execCmd(ctx, closure, cmd, exp, capture)
		}
	}

	return "", errors.Wrapf(
		ErrInvalidInstr,
		"%d:%d: %s",
		exp.Line,
		exp.Offset,
		exp.Str,
	)
}

// execCmd executes the given command.
func (c *cli) execCmd(
	ctx context.Context,
	closure *script.Closure,
	cmd Cmd,
	sexp *script.SExp,
	capture bool,
) (string, error) {
	buf := bytes.NewBuffer(nil)
	w := c.cons.Writer

	if capture {
		w = buf
	}

	eval := c.sexpEvaluator(ctx, closure, true)
	err := cmd.Exec(ctx, c, w, closure, eval, sexp)

	if err != nil {
		cause := errors.Cause(err)

		// If it is a usage error, add the use string to it.
		if userr, ok := cause.(*UseError); ok {
			userr.use = cmd.LongUse()
		}

		err = errors.Wrapf(
			err,
			"%d:%d: %s",
			sexp.Line,
			sexp.Offset,
			sexp.Str,
		)

		// Always print errors of nested commands.
		if capture {
			c.PrintError(err)
		}

		return "", err
	}

	return strings.TrimSuffix(buf.String(), "\n"), nil
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

	if userr, ok := errors.Cause(err).(*UseError); ok {
		c.cons.Print("\n" + userr.Use())
	}
}

// sexpEvaluator returns an S-Expression evaluator.
func (c *cli) sexpEvaluator(
	ctx context.Context,
	closure *script.Closure,
	capture bool,
) script.Evaluator {
	return func(resolve script.Resolver, exp *script.SExp) (string, error) {
		return c.evalSExp(ctx, closure, exp, capture)
	}
}

// Eval executes the given input, but does not handle errors or exit signals.
func (c *cli) Eval(ctx context.Context, in string) error {
	head, err := c.parser.Parse(in)
	if err != nil {
		return errors.WithStack(err)
	}

	for ; head != nil; head = head.Cdr {
		_, err = c.sexpEvaluator(ctx, c.closure, false)(nil, head.List)
		if err != nil {
			return err
		}
	}

	return nil
}

// Exec executes a command.
//
// It handles signals to cancel the command if the user presses Ctrl-C.
func (c *cli) Exec(ctx context.Context, in string) {
	defer func() { c.executed = true }()

	execCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	done := make(chan error, 1)
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		done <- c.Eval(execCtx, in)
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

// Resolver resolves the name of the symbol if it doesn't begin with a dollar
// sign.
func Resolver(sym *script.SExp) (*script.SExp, error) {
	if strings.HasPrefix(sym.Str, "$") {
		return nil, errors.Wrapf(
			script.ErrSymNotFound,
			"%d:%d: %s",
			sym.Line, sym.Offset, sym.Str,
		)
	}

	atom := sym.Clone()
	atom.Type = script.TypeStr

	return atom, nil
}
