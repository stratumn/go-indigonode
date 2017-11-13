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

// Package core implements Alice's core functionality.
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
	logger "github.com/stratumn/alice/core/log"
	"github.com/stratumn/alice/core/manager"
	"github.com/stratumn/alice/core/service/host"
	"github.com/stratumn/alice/release"

	metrics "gx/ipfs/QmQbh3Rb7KM37As3vkHYnEFnzkVXNCP8EYGtHz6g2fXk14/go-libp2p-metrics"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	identify "gx/ipfs/QmefgzMbKZYsmHFkLqxgaTBG9ypeEjrdWRD5WXH4j1cWDL/go-libp2p/p2p/protocol/identify"
)

// Set identify version.
func init() {
	identify.ClientVersion = "alice/" + release.Version + "/" + release.GitCommit
}

// log is the logger for the core module.
var log = logging.Logger("core")

// globalManagerMu guards globalManager against data races.
var globalManagerMu sync.Mutex

// globalManager is the global service manager.
var globalManager *manager.Manager

// GlobalManager returns the global service manager.
func GlobalManager() *manager.Manager {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()

	if globalManager == nil {
		globalManager = manager.New()
	}

	return globalManager
}

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
	mgr *manager.Manager
}

// New creates a new core.
func New(mgr *manager.Manager) *Core {
	if mgr == nil {
		mgr = GlobalManager()
	}

	return &Core{mgr}
}

// Boot starts the node and runs until an exit signal or a fatal error occurs.
//
// It fails if it failed starting the boot service. Otherwise it never returns
// unless the context is canceled.
func (c *Core) Boot(ctx context.Context) error {
	log.Event(ctx, "beginBoot")
	defer log.Event(ctx, "beginBoot")

	// Also register services as a side-effect. Get them now before
	// starting to avoid map concurrency issues.
	deps, err := Deps("")
	if err != nil {
		return err
	}

	bootScreenCtx, cancelBootScreen := context.WithCancel(context.Background())
	defer cancelBootScreen()

	bootScreenDone := make(chan struct{})
	if configHandler.config.EnableBootScreen {
		go func() {
			c.bootScreen(bootScreenCtx, deps)
			close(bootScreenDone)
		}()
	}

	workCtx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

	// Start manager queue.
	workDone := make(chan error, 1)
	go func() {
		workDone <- c.mgr.Work(workCtx)
	}()

	// Start boot service.
	if err := c.mgr.Start(configHandler.BootService()); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		if configHandler.config.EnableBootScreen {
			cancelBootScreen()
			<-bootScreenDone
		}
		fmt.Println("\n\nStopping...")
		c.mgr.StopAll()
		cancelWork()
		return <-workDone

	case err := <-workDone:
		if configHandler.config.EnableBootScreen {
			cancelBootScreen()
			<-bootScreenDone
		}
		return err
	}
}

// bootScreen displays the boot screen, which shows information about the
// services being started and metrics.
func (c *Core) bootScreen(ctx context.Context, deps []string) {
	fmt.Println()
	fmt.Println(art)
	fmt.Println()
	fmt.Print(release.Version + "@" + release.GitCommit[:7])
	fmt.Println(" -- Copyright © 2017 Stratumn SAS")
	fmt.Println()

	c.bootStatus(ctx, deps)
	c.hostInfo()
	fmt.Println("\nPress Ctrl^C to shutdown.")
	c.stats(ctx)
}

// bootStatus shows the services being started.
func (c *Core) bootStatus(ctx context.Context, deps []string) {
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
			return
		case <-running:
		}

		status := strings.Repeat(" ", 75-len([]rune(line)))
		status += "\033[0;32mok\033[0m"
		fmt.Print(status)
		fmt.Println()
	}
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

	return
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
func (c *Core) findHost() *host.Host {
	exposed, err := c.mgr.Expose("host")
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error:%s.\n", err)
	}

	if hst, ok := exposed.(*host.Host); ok {
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

// Deps returns the dependencies of a service in the order they would be
// started given the current configuration.
//
// The returned slice ends with the service itself.
//
// If no service ID is given, the boot service will be used.
func Deps(serviceID string) ([]string, error) {
	registerServices()

	if serviceID == "" {
		serviceID = configHandler.BootService()
	}

	return GlobalManager().Deps(serviceID)
}

// Fgraph prints the dependency graph of a service given the current
// configuration to a writer.
//
// If no service ID is given, the boot service will be used.
func Fgraph(w io.Writer, serviceID string) error {
	registerServices()

	if serviceID == "" {
		serviceID = configHandler.BootService()
	}

	return GlobalManager().Fgraph(w, serviceID, "")
}

// registerServices registers all the core services with the default service
// manager.
//
// It is safe to call multiple times.
func registerServices() {
	// Register all the services.
	for _, serv := range services {
		GlobalManager().Register(serv)
	}

	// Add services for groups.
	for _, serv := range configHandler.GroupServices() {
		GlobalManager().Register(serv)
	}
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
		MustExist: true,
		Follow:    follow,
		Logger:    tail.DiscardingLogger,
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
