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

// Package core defines Indigo Node's core functionality.
package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/cfg"
	logger "github.com/stratumn/go-node/core/log"
	"github.com/stratumn/go-node/core/manager"
	"github.com/stratumn/go-node/core/p2p"
	"github.com/stratumn/go-node/release"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	identify "gx/ipfs/QmUEqyXr97aUbNmQADHYNknjwjjdVpJXEt1UZXmSG81EV4/go-libp2p/p2p/protocol/identify"
)

// Set the identify protocol client version.
func init() {
	identify.ClientVersion = "indigo-node/" + release.Version + "/" + release.GitCommit
}

var (
	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("the configuration is invalid")
)

// log is the logger for the core module.
var log = logging.Logger("core")

// art is shown upon booting.
const art = "\033[0;34m    ____          ___                _   __          __\n" +
	"   /  _/___  ____/ (_)___ _____     / | / /___  ____/ /__\n" +
	"   / // __ \\/ __  / / __ `/ __ \\   /  |/ / __ \\/ __  / _ \\\n" +
	" _/ // / / / /_/ / / /_/ / /_/ /  / /|  / /_/ / /_/ /  __/\n" +
	"/___/_/ /_/\\__,_/_/\\__, /\\____/  /_/ |_/\\____/\\__,_/\\___/\n" +
	"                  /____/\n\033[0m"

// Opt is a core option.
type Opt func(*Core)

// OptManager sets the service manager.
//
// If this option is not given, a new manager is created.
func OptManager(mgr *manager.Manager) Opt {
	return func(c *Core) {
		c.mgr = mgr
	}
}

// OptServices adds services.
//
// If this option is not given, the builtin services are used.
func OptServices(services ...manager.Service) Opt {
	return func(c *Core) {
		c.services = append(c.services, services...)
	}
}

// OptStdout sets the writer for the boot screen normal output.
//
// If this option is not given, os.Stdout is used.
func OptStdout(w io.Writer) Opt {
	return func(c *Core) {
		c.stdout = w
	}
}

// OptStderr sets the writer for the boot screen error output.
//
// If this option is not given, os.Stderr is used.
func OptStderr(w io.Writer) Opt {
	return func(c *Core) {
		c.stderr = w
	}
}

// OptUpHandler sets a function that will be called once the node has
// started.
//
// This is useful for tests.
func OptUpHandler(h func()) Opt {
	return func(c *Core) {
		c.upHandlers = append(c.upHandlers, h)
	}
}

// OptMetricsHandler sets a function that will be called everytime metrics
// tick.
//
// This is useful for tests.
func OptMetricsHandler(h func()) Opt {
	return func(c *Core) {
		c.metricsHandlers = append(c.metricsHandlers, h)
	}
}

// Core manages a node. It wraps a service manager, and can display a
// boot screen showing services being started and node metrics.
type Core struct {
	mgr             *manager.Manager
	services        []manager.Service
	config          Config
	stdout          io.Writer
	stderr          io.Writer
	upHandlers      []func()
	metricsHandlers []func()
}

// New creates a new core.
func New(configSet cfg.ConfigSet, opts ...Opt) (*Core, error) {
	config, ok := configSet["core"].(Config)
	if !ok {
		return nil, errors.WithStack(ErrInvalidConfig)
	}

	c := &Core{config: config}

	for _, o := range opts {
		o(c)
	}

	if c.mgr == nil {
		c.mgr = manager.New()
	}

	if c.services == nil {
		c.services = BuiltinServices()
	}

	if c.stdout == nil {
		c.stdout = os.Stdout
	}

	if c.stderr == nil {
		c.stderr = os.Stderr
	}

	return c, nil
}

// Boot starts the node and runs until an exit signal or a fatal error occurs.
//
// It returns an error early if it failed to start the boot service. Otherwise
// it never returns unless the context is canceled.
func (c *Core) Boot(ctx context.Context) error {
	log.Event(ctx, "beginBoot")
	defer log.Event(ctx, "endBoot")

	workCtx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

	// Start manager queue.
	workCh := make(chan error, 1)
	go func() {
		workCh <- c.mgr.Work(workCtx)
	}()

	registerServices(c.mgr, c.services, &c.config)

	bootScreenCtx, cancelBootScreen := context.WithCancel(context.Background())
	defer cancelBootScreen()

	bootScreenCh := make(chan error, 1)
	if c.config.EnableBootScreen {
		go func() {
			bootScreenCh <- c.bootScreen(bootScreenCtx)
		}()
	}

	// Start boot service.
	if err := c.mgr.Start(c.config.BootService); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		if c.config.EnableBootScreen {
			cancelBootScreen()
			<-bootScreenCh
		}
		fmt.Fprintln(c.stdout, "\n\nStopping...")
		c.mgr.StopAll()
		cancelWork()
		return <-workCh

	case err := <-bootScreenCh:
		c.mgr.StopAll()
		cancelWork()
		<-workCh
		return err

	case err := <-workCh:
		if c.config.EnableBootScreen {
			cancelBootScreen()
			<-bootScreenCh
		}
		return err
	}
}

// bootScreen displays the boot screen, which shows information about the
// services being started and metrics.
func (c *Core) bootScreen(ctx context.Context) error {
	fmt.Fprintln(c.stdout)
	fmt.Fprintln(c.stdout, art)
	fmt.Fprintln(c.stdout)
	fmt.Fprintln(c.stdout, release.Version+"@"+release.GitCommit[:7])
	fmt.Fprintln(c.stdout)

	if err := c.bootStatus(ctx); err != nil {
		return err
	}

	c.hostInfo()
	fmt.Fprintln(c.stdout, "\nPress Ctrl^C to shutdown.")

	for _, f := range c.upHandlers {
		f()
	}

	c.nodeMetrics(ctx)

	return nil
}

// bootStatus shows the services being started.
func (c *Core) bootStatus(ctx context.Context) error {
	deps, err := c.mgr.Deps(c.config.BootService)
	if err != nil {
		return err
	}

	for i, sid := range deps {
		line := fmt.Sprintf("Starting %s (%d/%d)...", sid, i+1, len(deps))
		fmt.Fprint(c.stdout, line)

		running, err := c.mgr.Running(sid)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		case <-running:
		}

		status := strings.Repeat(" ", 75-len([]rune(line)))
		status += "\033[0;32mok\033[0m"
		fmt.Fprint(c.stdout, status)
		fmt.Fprintln(c.stdout)
	}

	return nil
}

// hostInfo shows the peer ID and the listen addresses of the host.
func (c *Core) hostInfo() {
	hst := c.findHost()
	if hst == nil {
		return
	}

	fmt.Fprintf(c.stdout, "\nStarted node %s.\n\n", hst.ID().Pretty())

	addrs, err := hst.Network().InterfaceListenAddresses()
	if err != nil {
		fmt.Fprintf(c.stderr, "Failed to get listen addresses: %s.\n", err)
		return
	}

	for _, addr := range addrs {
		fmt.Fprintf(c.stdout, "Listening at %s.\n", addr)
	}

	if len(addrs) > 0 {
		fmt.Fprintln(c.stdout)
	}

	for _, addr := range hst.Addrs() {
		fmt.Fprintf(c.stdout, "Announcing %s.\n", addr)
	}
}

const (
	// metricsInterval is the interval between metrics ticks.
	metricsInterval = time.Second

	// metricsFmt is the format string for metrics.
	metricsFmt = "\033[2K\rPeers: %d | Conns: %d"
)

// nodeMetrics periodically displays some node metrics.
func (c *Core) nodeMetrics(ctx context.Context) {
	ticker := time.NewTicker(metricsInterval)

	fmt.Fprintln(c.stdout, "\nStats:")

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return

		case <-ticker.C:
			// We have to fetch these every tick because the
			// services could be stopped or restarted.
			hst := c.findHost()

			if hst == nil {
				fmt.Fprint(c.stdout, "\033[2K\rMetrics are not available.")

				for _, f := range c.metricsHandlers {
					f()
				}

				continue
			}

			peers := len(hst.Network().Peers())
			conns := len(hst.Network().Conns())

			fmt.Fprintf(c.stdout, metricsFmt, peers, conns)

			for _, f := range c.metricsHandlers {
				f()
			}
		}
	}
}

// findHost finds the host service.
func (c *Core) findHost() *p2p.Host {
	exposed, err := c.mgr.Expose(c.config.BootScreenHost)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %s.\n", err)
	}

	if hst, ok := exposed.(*p2p.Host); ok {
		return hst
	}

	return nil
}

// doWithManager runs a function with a freshly created service manager.
func doWithManager(
	services []manager.Service,
	config *Config,
	fn func(*manager.Manager),
) error {
	workCtx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

	mgr := manager.New()

	// Start manager queue.
	ch := make(chan error, 1)
	go func() {
		ch <- mgr.Work(workCtx)
	}()

	registerServices(mgr, services, config)

	fn(mgr)

	cancelWork()
	err := <-ch

	if err != nil && errors.Cause(err) != context.Canceled {
		return err
	}

	return nil
}

// Deps returns the dependencies of a service in the order they would be
// started with the given configuration.
//
// The returned slice ends with the service itself.
//
// If no service ID is given, the boot service will be used.
func Deps(
	services []manager.Service,
	configSet cfg.ConfigSet,
	servID string,
) (deps []string, err error) {
	config, ok := configSet["core"].(Config)
	if !ok {
		return nil, errors.WithStack(ErrInvalidConfig)
	}

	if servID == "" {
		servID = config.BootService
	}

	mgrError := doWithManager(services, &config, func(mgr *manager.Manager) {
		deps, err = mgr.Deps(servID)
	})

	if mgrError != nil {
		return nil, mgrError
	}

	return
}

// Fgraph prints the dependency graph of a service given the current
// configuration to a writer.
//
// If no service ID is given, the boot service will be used.
func Fgraph(
	w io.Writer,
	services []manager.Service,
	configSet cfg.ConfigSet,
	servID string,
) error {
	config, ok := configSet["core"].(Config)
	if !ok {
		return errors.WithStack(ErrInvalidConfig)
	}

	if servID == "" {
		servID = config.BootService
	}

	var err error

	mgrError := doWithManager(services, &config, func(mgr *manager.Manager) {
		err = doFgraph(w, mgr, servID)
	})

	if mgrError != nil {
		return mgrError
	}

	return err
}

func doFgraph(w io.Writer, mgr *manager.Manager, servID string) error {
	return mgr.Fgraph(w, servID, "")
}

// PrettyLog pretty prints log output.
func PrettyLog(filename, level, system string, follow, json, color bool) error {
	var w io.Writer

	var systemRegexp *regexp.Regexp
	if system != "" {
		systemRegexp, _ = regexp.Compile(system)
	}

	filter := func(entry map[string]interface{}) bool {
		_, isError := entry["error"]

		switch {
		case level == logger.Info && isError:
			return false
		case level == logger.Error && !isError:
			return false
		case systemRegexp != nil:
			return systemRegexp.MatchString(entry["system"].(string))
		case system != "":
			return system == entry["system"].(string)
		}

		return true
	}

	mu := sync.Mutex{}

	if json {
		w = logger.NewFilteredWriter(os.Stderr, &mu, filter)
	} else {
		w = logger.NewPrettyWriter(os.Stderr, &mu, filter, color)
	}

	t, err := tail.TailFile(filename, tail.Config{
		Follow: follow,
		Logger: tail.DiscardingLogger,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	defer t.Cleanup()

	for line := range t.Lines {
		if _, err := w.Write([]byte(line.Text + "\n")); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
