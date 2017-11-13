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

package it

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/cfg"
	"google.golang.org/grpc"
)

// SessionFunc represents a function used to interact with a test network
// during a session.
type SessionFunc func(context.Context, TestNodeSet, []*grpc.ClientConn)

// Session launches a test network and calls a user defined function that can
// interact with the nodes via API clients.
//
// The configuration files, data files, and logs will be saved in a directory
// within the given directory.
func Session(ctx context.Context, dir string, numNodes int, config cfg.ConfigSet, fn SessionFunc) error {
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
