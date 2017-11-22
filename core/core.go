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

// Package core defines Alice's core functionality.
package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/hpcloud/tail"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/cfg"
	logger "github.com/stratumn/alice/core/log"
	"github.com/stratumn/alice/core/manager"
	"github.com/stratumn/alice/core/p2p"
	"github.com/stratumn/alice/release"

	metrics "gx/ipfs/QmQbh3Rb7KM37As3vkHYnEFnzkVXNCP8EYGtHz6g2fXk14/go-libp2p-metrics"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	identify "gx/ipfs/QmefgzMbKZYsmHFkLqxgaTBG9ypeEjrdWRD5WXH4j1cWDL/go-libp2p/p2p/protocol/identify"
)

// Set identify version.
func init() {
	identify.ClientVersion = "alice/" + release.Version + "/" + release.GitCommit
}

var (
	// ErrInvalidConfig is returns when the configuration is invalid.
	ErrInvalidConfig = errors.New("the configuration is invalid")
)

// log is the logger for the core module.
var log = logging.Logger("core")

// art is shown upon booting.
const art = "\033[0;34m      .o.       oooo   o8o\n" +
	"     .888.      `888   `\"'\n" +
	"    .8\"888.      888  oooo   .ooooo.   .ooooo.\n" +
	"   .8' `888.     888  `888  d88' `\"Y8 d88' `88b\n" +
	"  .88ooo8888.    888   888  888       888ooo888\n" +
	" .8'     `888.   888   888  888   .o8 888    .o\n" +
	"o88o     o8888o o888o o888o `Y8bod8P' `Y8bod8P'\033[0m"

// Core manages a node. It wraps a service manager, and can display a
// boot screen showing services being started and node metrics.
type Core struct {
	mgr    *manager.Manager
	config Config
}

// New creates a new core.
func New(configSet cfg.ConfigSet) (*Core, error) {
	config, ok := configSet["core"].(Config)
	if !ok {
		return nil, errors.WithStack(ErrInvalidConfig)
	}

	return &Core{manager.New(), config}, nil
}

// Boot starts the node and runs until an exit signal or a fatal error occurs.
//
// It fails if it failed starting the boot service. Otherwise it never returns
// unless the context is canceled.
func (c *Core) Boot(ctx context.Context) error {
	log.Event(ctx, "beginBoot")
	defer log.Event(ctx, "beginBoot")

	workCtx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

	// Start manager queue.
	workCh := make(chan error, 1)
	go func() {
		workCh <- c.mgr.Work(workCtx)
	}()

	registerServices(c.mgr, &c.config)

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
		fmt.Println("\n\nStopping...")
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
	fmt.Println()
	fmt.Println(art)
	fmt.Println()
	fmt.Print(release.Version + "@" + release.GitCommit[:7])
	fmt.Println(" -- Copyright © 2017 Stratumn SAS")
	fmt.Println()

	if err := c.bootStatus(ctx); err != nil {
		return err
	}

	c.hostInfo()
	fmt.Println("\nPress Ctrl^C to shutdown.")
	c.stats(ctx)

	return nil
}

// bootStatus shows the services being started.
func (c *Core) bootStatus(ctx context.Context) error {
	deps, err := doDeps(c.mgr, c.config.BootService)
	if err != nil {
		return err
	}

	for i, sid := range deps {
		line := fmt.Sprintf("Starting %s (%d/%d)...", sid, i+1, len(deps))
		fmt.Print(line)

		running, err := c.mgr.Running(sid)
		if err != nil {
			// Impossible.
			panic(err)
		}

		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		case <-running:
		}

		status := strings.Repeat(" ", 75-len([]rune(line)))
		status += "\033[0;32mok\033[0m"
		fmt.Print(status)
		fmt.Println()
	}

	return nil
}

// hostInfo shows the peer ID and the listen addresses of the host.
func (c *Core) hostInfo() {
	hst := c.findHost()
	if hst == nil {
		return
	}

	fmt.Printf("\nStarted node %s.\n\n", hst.ID().Pretty())

	addrs, err := hst.Network().InterfaceListenAddresses()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get listen addresses: %s.\n", err)
		return
	}

	for _, addr := range addrs {
		fmt.Printf("Listening at %s.\n", addr)
	}

	if len(addrs) > 0 {
		fmt.Println()
	}

	for _, addr := range hst.Addrs() {
		fmt.Printf("Announcing %s.\n", addr)
	}
}

const (
	// statsInterval is the interval between stats ticks.
	statsInterval = 500 * time.Millisecond

	// statsFmt is the format string for stats.
	statsFmt = "\033[2K\rPeers: %d | Conns: %d | Total: ↓%s ↑%s | Rate: ↓%s/s ↑%s/s"
)

// stats shows host statistics every second.
func (c *Core) stats(ctx context.Context) {
	ticker := time.NewTicker(statsInterval)

	fmt.Println("\nStats:")

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return

		case <-ticker.C:
			// We have to fetch these every tick because the
			// services could be stopped or restarted.
			hst := c.findHost()
			bwc := c.findMetrics()

			if hst == nil || bwc == nil {
				fmt.Print("\033[2K\rMetrics are not available.")
				continue
			}

			peers := len(hst.Network().Peers())
			conns := len(hst.Network().Conns())

			stats := bwc.GetBandwidthTotals()
			totalIn := bytefmt.ByteSize(uint64(stats.TotalIn))
			totalOut := bytefmt.ByteSize(uint64(stats.TotalOut))
			rateIn := bytefmt.ByteSize(uint64(stats.RateIn))
			rateOut := bytefmt.ByteSize(uint64(stats.RateOut))

			fmt.Printf(
				statsFmt,
				peers, conns,
				totalIn, totalOut,
				rateIn, rateOut,
			)
		}
	}
}

// findHost finds the host service.
func (c *Core) findHost() *p2p.Host {
	exposed, err := c.mgr.Expose("host")
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error:%s.\n", err)
	}

	if hst, ok := exposed.(*p2p.Host); ok {
		return hst
	}

	return nil
}

// findMetrics finds the metrics service.
func (c *Core) findMetrics() metrics.Reporter {
	exposed, err := c.mgr.Expose("metrics")
	if err != nil {
		return nil
	}

	if bwc, ok := exposed.(metrics.Reporter); ok {
		return bwc
	}

	return nil
}

// doWithManager runs a function with with a freshly created service manager.
func doWithManager(config *Config, fn func(*manager.Manager)) error {
	workCtx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

	mgr := manager.New()

	// Start manager queue.
	ch := make(chan error, 1)
	go func() {
		ch <- mgr.Work(workCtx)
	}()

	registerServices(mgr, config)

	fn(mgr)

	cancelWork()
	err := <-ch

	if err != nil && errors.Cause(err) != context.Canceled {
		return err
	}

	return nil
}

// Deps returns the dependencies of a service in the order they would be
// started given the current configuration.
//
// The returned slice ends with the service itself.
//
// If no service ID is given, the boot service will be used.
func Deps(configSet cfg.ConfigSet, servID string) (deps []string, err error) {
	config, ok := configSet["core"].(Config)
	if !ok {
		return nil, errors.WithStack(ErrInvalidConfig)
	}

	if servID == "" {
		servID = config.BootService
	}

	mgrError := doWithManager(&config, func(mgr *manager.Manager) {
		deps, err = doDeps(mgr, servID)
	})

	if mgrError != nil {
		return nil, mgrError
	}

	return
}

func doDeps(mgr *manager.Manager, servID string) ([]string, error) {
	return mgr.Deps(servID)
}

// Fgraph prints the dependency graph of a service given the current
// configuration to a writer.
//
// If no service ID is given, the boot service will be used.
func Fgraph(w io.Writer, configSet cfg.ConfigSet, servID string) error {
	config, ok := configSet["core"].(Config)
	if !ok {
		return errors.WithStack(ErrInvalidConfig)
	}

	if servID == "" {
		servID = config.BootService
	}

	var err error

	mgrError := doWithManager(&config, func(mgr *manager.Manager) {
		err = doFgraph(w, mgr, servID)
	})

	if mgrError != nil {
		err = mgrError
	}

	return err
}

func doFgraph(w io.Writer, mgr *manager.Manager, servID string) error {
	return mgr.Fgraph(w, servID, "")
}

// PrettyLog pretty prints log output.
func PrettyLog(filename, level, system string, follow, json, color bool) error {
	var w io.Writer

	filter := func(entry map[string]interface{}) bool {
		_, isError := entry["error"]

		switch {
		case level == logger.Info && isError:
			return false
		case level == logger.Error && !isError:
			return false
		case system != "" && system != entry["system"].(string):
			return false
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
