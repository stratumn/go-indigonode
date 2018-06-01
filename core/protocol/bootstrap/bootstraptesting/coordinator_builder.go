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

package bootstraptesting

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap"
	"github.com/stretchr/testify/require"

	netutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	bhost "gx/ipfs/Qmc64U41EEB4nPG7wxjEqFwKJajS2f8kk5q2TvUrQf78Xu/go-libp2p-blankhost"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// CoordinatorBuilder lets you configure a test coordinator host.
type CoordinatorBuilder struct {
	t             *testing.T
	h             *bhost.BlankHost
	handler       bootstrap.Handler
	networkConfig protector.NetworkConfig
}

// NewCoordinatorBuilder returns a CoordinatorBuilder with default configuration.
func NewCoordinatorBuilder(ctx context.Context, t *testing.T) *CoordinatorBuilder {
	return &CoordinatorBuilder{
		t: t,
		h: bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx)),
	}
}

// WithNetworkConfig configures the NetworkConfig used.
func (c *CoordinatorBuilder) WithNetworkConfig(networkConfig protector.NetworkConfig) *CoordinatorBuilder {
	c.networkConfig = networkConfig
	return c
}

// Unavailable configures the host to be unavailable.
func (c *CoordinatorBuilder) Unavailable() *CoordinatorBuilder {
	require.NoError(c.t, c.h.Close(), "h.Close()")
	return c
}

// PeerID returns the ID of the coordinator node.
func (c *CoordinatorBuilder) PeerID() peer.ID {
	return c.h.ID()
}

// Host returns the node's host.
func (c *CoordinatorBuilder) Host() *bhost.BlankHost {
	return c.h
}

// Close closes resources used by the node.
func (c *CoordinatorBuilder) Close(ctx context.Context) {
	c.handler.Close(ctx)
	require.NoError(c.t, c.h.Close(), "h.Close()")
}

// Build builds the coordinator node and returns its handler.
func (c *CoordinatorBuilder) Build() bootstrap.Handler {
	var err error
	c.handler, err = bootstrap.NewCoordinatorHandler(c.h, c.networkConfig)
	require.NoError(c.t, err, "bootstrap.NewCoordinatorHandler()")

	return c.handler
}
