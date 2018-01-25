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

//+build gui

package cmd

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/spf13/cobra"
	"github.com/stratumn/alice/core/manager"

	metrics "gx/ipfs/QmaL2WYJGbWKqHoLujoi9GQ5jj4JVFrBqHUBWmEYzJPVWT/go-libp2p-metrics"
)

// upCmd represents the up command.
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start a node",
	Run: func(cmd *cobra.Command, args []string) {
		systray.Run(onReady, onExit)

	},
}

func init() {
	RootCmd.AddCommand(upCmd)
}

func onReady() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	systray.SetIcon(icon.Data)
	systray.SetTitle("Alice")

	metrics := systray.AddMenuItem("Unavailable", "Bandwidth usage")
	metrics.Disable()

	systray.AddSeparator()

	quit := systray.AddMenuItem("Quit", "Terminate the node")

	mgr := manager.New()

	metricsDone := make(chan struct{})
	go func() {
		nodeMetrics(ctx, mgr, metrics)
		close(metricsDone)
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				<-metricsDone
				systray.Quit()
				return
			case <-quit.ClickedCh:
				cancel()
			}
		}
	}()

	up(mgr)

}

func onExit() {
}

const (
	// metricsInterval is the interval between metrics ticks.
	metricsInterval = 500 * time.Millisecond

	// metricsFmt is the format string for metrics.
	metricsFmt = "Total ↓%s ↑%s Rate: ↓%s/s ↑%s/s"
)

// nodeMetrics periodically displays some node metrics.
func nodeMetrics(ctx context.Context, mgr *manager.Manager, menu *systray.MenuItem) {
	ticker := time.NewTicker(metricsInterval)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return

		case <-ticker.C:
			// We have to fetch these every tick because the
			// services could be stopped or restarted.
			bwc := findMetrics(mgr)

			if bwc == nil {
				menu.SetTitle("Unavailable")
				continue
			}

			bw := bwc.GetBandwidthTotals()
			totalIn := bytefmt.ByteSize(uint64(bw.TotalIn))
			totalOut := bytefmt.ByteSize(uint64(bw.TotalOut))
			rateIn := bytefmt.ByteSize(uint64(bw.RateIn))
			rateOut := bytefmt.ByteSize(uint64(bw.RateOut))

			str := fmt.Sprintf(
				metricsFmt,
				totalIn, totalOut,
				rateIn, rateOut,
			)

			menu.SetTitle(str)
		}
	}
}

// findMetrics finds the metrics service.
func findMetrics(mgr *manager.Manager) metrics.Reporter {
	exposed, err := mgr.Expose("metrics")
	if err != nil {
		return nil
	}

	if bwc, ok := exposed.(metrics.Reporter); ok {
		return bwc
	}

	return nil
}
