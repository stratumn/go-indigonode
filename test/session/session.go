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

//+build !lint

// Package session defines types for doing system tests.
//
// It facilitates launching a network of nodes each in a separate process and
// run tests against them via the API.
package session

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core"
	bootstrap "github.com/stratumn/go-indigonode/core/app/bootstrap/service"
	grpcapi "github.com/stratumn/go-indigonode/core/app/grpcapi/service"
	metrics "github.com/stratumn/go-indigonode/core/app/metrics/service"
	"github.com/stratumn/go-indigonode/core/cfg"
	logging "github.com/stratumn/go-indigonode/core/log"
	"google.golang.org/grpc"
)

// NumSeeds is the number of seeds in a bootstrap list.
const NumSeeds = 5

// SystemCfg returns a base configuration for system testing.
func SystemCfg() cfg.ConfigSet {
	conf := core.NewConfigurableSet(core.BuiltinServices()).Configs()

	coreConf := conf["core"].(core.Config)
	coreConf.EnableBootScreen = false
	conf["core"] = coreConf

	logConf := conf["log"].(logging.Config)
	logConf.Writers = []logging.WriterConfig{{
		Type:      logging.Stderr,
		Level:     logging.All,
		Formatter: logging.JSON,
	}}
	conf["log"] = logConf

	apiConf := conf["grpcapi"].(grpcapi.Config)
	apiConf.EnableRequestLogger = true
	conf["grpcapi"] = apiConf

	bsConf := conf["bootstrap"].(bootstrap.Config)
	bsConf.Interval = "5s"
	bsConf.ConnectionTimeout = "3s"
	conf["bootstrap"] = bsConf

	metricsConf := conf["metrics"].(metrics.Config)
	metricsConf.PrometheusEndpoint = ""
	conf["metrics"] = metricsConf

	return conf
}

// BenchmarkCfg returns a base configuration for benchmarking.
func BenchmarkCfg() cfg.ConfigSet {
	conf := core.NewConfigurableSet(core.BuiltinServices()).Configs()

	coreConf := conf["core"].(core.Config)
	coreConf.EnableBootScreen = false
	conf["core"] = coreConf

	logConf := conf["log"].(logging.Config)
	logConf.Writers = []logging.WriterConfig{{
		Type:      logging.Stderr,
		Level:     logging.Error,
		Formatter: logging.JSON,
	}}
	conf["log"] = logConf

	apiConf := conf["grpcapi"].(grpcapi.Config)
	apiConf.EnableRequestLogger = false
	conf["grpcapi"] = apiConf

	bsConf := conf["bootstrap"].(bootstrap.Config)
	bsConf.Interval = "5s"
	bsConf.ConnectionTimeout = "3s"
	conf["bootstrap"] = bsConf

	metricsConf := conf["metrics"].(metrics.Config)
	metricsConf.PrometheusEndpoint = ""
	conf["metrics"] = metricsConf

	return conf
}

// WithServices takes a configuration and returns a new one with the same
// settings except that it changes the boot service.
//
// It will also start the signal service so that exit signals are handled
// properly.
func WithServices(config cfg.ConfigSet, services ...string) cfg.ConfigSet {
	services = append(services, "signal")
	conf := deepcopy.Copy(config).(cfg.ConfigSet)
	coreConf := conf["core"].(core.Config)
	coreConf.ServiceGroups = append(coreConf.ServiceGroups, core.ServiceGroupConfig{
		ID:       "test",
		Services: services,
	})
	coreConf.BootService = "test"
	conf["core"] = coreConf

	return conf
}

// MultiErr represents multiple errors as a single error.
type MultiErr []error

// Error returns the errors messages joined by semicolons.
func (e MultiErr) Error() string {
	errs := make([]string, len(e))

	for i, err := range e {
		errs[i] = err.Error()
	}

	return strings.Join(errs, "; ")
}

// NewMultiErr creates a multierror from a slice of errors.
//
// Only non-nil error will be added. If there are no non-nil errors, it returns
// nil.
func NewMultiErr(errs []error) error {
	var merrs MultiErr

	for _, v := range errs {
		if v != nil {
			merrs = append(merrs, v)
		}
	}

	if len(merrs) > 0 {
		return merrs
	}

	return nil
}

// Format formats each error individually.
func (e MultiErr) Format(s fmt.State, verb rune) {
	for i, err := range e {
		if f, ok := err.(fmt.Formatter); ok {
			f.Format(s, verb)
		} else {
			switch verb {
			case 'q':
				fmt.Fprintf(s, "%q", err)
			default:
				_, ioErr := io.WriteString(s, err.Error())
				if ioErr != nil {
					fmt.Fprintf(s, "%q", err)
				}
			}
		}
		if i < len(e)-1 {
			_, ioErr := io.WriteString(s, "\n\n")
			if ioErr != nil {
				fmt.Fprintf(s, "%q", err)
			}
		}
	}
}

// Tester represents a function used to interact with a test network during a
// session.
type Tester func(context.Context, TestNodeSet, []*grpc.ClientConn)

// Run launches a test network and calls a user defined function that can
// interact with the nodes via API clients.
//
// The configuration files, data files, and logs will be saved in a directory
// within the given directory.
func Run(ctx context.Context, dir string, numNodes int, config cfg.ConfigSet, fn Tester) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.WithStack(err)
	}

	prefix := time.Now().UTC().Format(time.RFC3339) + "--"
	dir, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		return errors.WithStack(err)
	}

	// Create the test network.
	set, err := NewTestNodeSet(dir, numNodes, config)
	if err != nil {
		return err
	}
	return run(ctx, dir, set, fn)
}

// RunWithConfigs launches a test network and calls a user defined function that can
// interact with the nodes via API clients.
// Each node has a specific config.
//
// The configuration files, data files, and logs will be saved in a directory
// within the given directory.
func RunWithConfigs(ctx context.Context, dir string, numNodes int, configs []cfg.ConfigSet, fn Tester) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.WithStack(err)
	}

	prefix := time.Now().UTC().Format(time.RFC3339) + "--"
	dir, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		return errors.WithStack(err)
	}

	// Create the test network.
	set, err := NewTestNodeSetWithConfigs(dir, numNodes, configs)
	if err != nil {
		return err
	}

	return run(ctx, dir, set, fn)
}

func run(ctx context.Context, dir string, set TestNodeSet, fn Tester) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle exit signals.
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	// Start the test network in a goroutine.
	upDone := make(chan error, 1)
	go func() {
		upDone <- set.Up(ctx)
	}()

	var errs []error

	// Create the API clients.
	conns, err := set.Connect(ctx)
	if err != nil {
		errs = append(errs, err)
		return NewMultiErr(errs)
	}

	// Run the user function.
	fnDone := make(chan struct{}, 1)
	if err == nil {
		go func() {
			fn(ctx, set, conns)
			fnDone <- struct{}{}
		}()
		defer func() {
			for _, conn := range conns {
				conn.Close()
			}
		}()
	}

	exiting := false

	// Handle exit conditions.
	for {
		select {
		case err := <-upDone:
			errs = append(errs, err)
			if ctx.Err() != context.Canceled {
				errs = append(errs, ctx.Err())
			}
			if err = NewMultiErr(errs); err != nil {
				printLog(dir, err)
			}
			return err
		case <-fnDone:
			cancel()
		case <-sigc:
			if exiting {
				os.Exit(1)
			}
			exiting = true
			cancel()
		}
	}
}

func printLog(dir string, err error) {
	if multi, ok := err.(MultiErr); ok {
		for _, nodeErr := range multi {
			if nodeErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %s.\n", nodeErr)
				fmt.Fprintln(os.Stderr, "Here's the content of the log file:")
				var nodeNum int
				n, scanErr := fmt.Sscanf(nodeErr.Error(), "node %d:", &nodeNum)
				if scanErr != nil {
					fmt.Fprintf(os.Stderr, "Error: %s.\n", scanErr)
				}
				if n == 0 {
					return
				}
				nodeDir := fmt.Sprintf("node%d", nodeNum)
				logf, logErr := os.OpenFile(
					filepath.Join(dir, nodeDir, "logs"),
					os.O_RDONLY, 0700,
				)
				if logErr != nil {
					fmt.Fprintf(os.Stderr, "Could not open log file: %s.\n", logErr)
				}
				defer logf.Close()
				_, copyErr := io.Copy(os.Stderr, logf)
				if copyErr != nil {
					fmt.Fprintf(os.Stderr, "Could not print log file: %s.\n", copyErr)
				}
			}
		}
	}
}
