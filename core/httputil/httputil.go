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

// Package httputil provides an utility method to start an HTTP server in the
// context of an Indigo Node service.
package httputil

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/netutil"
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
