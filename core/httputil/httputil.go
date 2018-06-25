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

// Package httputil provides an utility method to start an HTTP server in the
// context of an Alice service.
package httputil

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/netutil"
)

// StartServer starts an HTTP server that listens on the provided address.
// It will block until the context is done.
func StartServer(ctx context.Context, address string, handler http.Handler) error {
	lis, err := netutil.Listen(address)
	if err != nil {
		return err
	}

	srv := http.Server{Handler: handler}

	done := make(chan error, 1)
	go func() {
		err := srv.Serve(lis)
		if err != nil && err != http.ErrServerClosed {
			done <- err
			return
		}

		close(done)
	}()

	shutdown := func() error {
		shutdownCtx, cancelShutdown := context.WithTimeout(
			context.Background(),
			time.Second/2,
		)
		defer cancelShutdown()

		return srv.Shutdown(shutdownCtx)
	}

	select {
	case err := <-done:
		return errors.WithStack(err)

	case <-ctx.Done():
		for {
			if err := shutdown(); err != nil {
				return errors.WithStack(err)
			}

			select {
			case err := <-done:
				if err != nil {
					return errors.WithStack(err)
				}
			case <-time.After(time.Second / 2):
				// Serve will not stop if we call Shutdown
				// before Serve, so in case this happens we
				// try again (for instance during tests).
				continue
			}

			return errors.WithStack(ctx.Err())
		}
	}
}
