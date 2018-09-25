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

package core

import (
	"bytes"
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/cfg"
	"github.com/stratumn/go-indigonode/core/manager"
	"github.com/stratumn/go-indigonode/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	swarmtesting "gx/ipfs/QmeDpqUwwdye8ABKVMPXKuWwPVURFdqTqssbTUB39E2Nwd/go-libp2p-swarm/testing"
)

type coreTest struct {
	name            string
	mockServices    servicesMocker
	servicesToStart []string
	waitForMetrics  bool
	validateStdout  validator
	validateStderr  validator
}

var coreTests = []coreTest{{
	"with metrics",
	withValidServices,
	[]string{"host"},
	true,
	validateCompose(
		validateRegexp("Starting host.+ok"),
		validateRegexp("Starting boot.+ok"),
		validateRegexp(`Listening at \/ip4\/127\.0\.0\.1\/tcp\/[0-9]+`),
		validateRegexp(`Announcing \/ip4\/127\.0\.0\.1\/tcp\/[0-9]+`),
		validateRegexp("Peers: 0"),
		validateRegexp("Conns: 0"),
	),
	validateEmpty,
}}

type servicesMocker func(context.Context, *testing.T) (services []manager.Service, closer func() error)

type mockService struct {
	id     string
	expose interface{}
	active bool
}

func (s *mockService) ID() string   { return s.id }
func (s *mockService) Name() string { return s.id }
func (s *mockService) Desc() string { return s.id }
func (s *mockService) Expose() interface{} {
	if !s.active {
		return nil
	}
	return s.expose
}
func (s *mockService) Run(ctx context.Context, running, stopping func()) error {
	s.active = true
	running()
	<-ctx.Done()
	stopping()
	s.active = false
	return ctx.Err()
}

// withValidServices mocks a valid host service.
func withValidServices(ctx context.Context, t *testing.T) ([]manager.Service, func() error) {
	host := p2p.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))

	services := []manager.Service{
		&mockService{
			"host",
			host,
			false,
		},
	}

	return services, host.Close
}

type validator func(*testing.T, string)

func validateEmpty(t *testing.T, str string) {
	assert.Empty(t, str)
}

func validateCompose(validators ...validator) validator {
	return func(t *testing.T, str string) {
		for _, v := range validators {
			v(t, str)
		}
	}
}

func validateRegexp(pattern string) validator {
	return func(t *testing.T, str string) {
		assert.Regexp(t, regexp.MustCompile(pattern), str)
	}
}

func TestConfigurableSet(t *testing.T) {
	// NewConfigurableSet creates a configuration set for
	// builtins services and adds the core and logging configurables.
	setDefault := NewConfigurableSet(nil)
	assert.Len(t, setDefault, len(BuiltinServices())+2)

	services, close := withValidServices(context.Background(), t)
	defer close()

	// It only adds services that implement the Configurable interface.
	setCustom := NewConfigurableSet(services)
	assert.Len(t, setCustom, 2)
}

func TestCore(t *testing.T) {
	for _, tt := range coreTests {
		t.Run(tt.name, func(t *testing.T) {
			testCore(t, tt)
		})
	}
}

func testCore(t *testing.T, tt coreTest) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	services, close := tt.mockServices(ctx, t)
	defer close()

	config := createTestConfig(ctx, t, services, tt.servicesToStart)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	upCh := make(chan struct{}, 1)
	upHandler := func() {
		upCh <- struct{}{}
	}

	metricsCh := make(chan struct{}, 1000)
	metricsHandler := func() {
		metricsCh <- struct{}{}
	}

	c, err := New(
		config,
		OptServices(services...),
		OptStdout(stdout),
		OptStderr(stderr),
		OptUpHandler(upHandler),
		OptMetricsHandler(metricsHandler),
	)

	require.NoError(t, err, "New()")

	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Boot(ctx)
	}()

	readyCh := upCh
	if tt.waitForMetrics {
		readyCh = metricsCh
	}

	select {
	case <-time.After(5 * time.Second):
		require.Fail(t, "node did not start in time")
	case err := <-errCh:
		// Boot() doesn't return before up handlers are called unless
		// an error occured. In fact it never returns unless there is
		// an error or the context was canceled.
		require.Fail(t, err.Error(), "node failed to start services")
	case <-readyCh:
	}

	if tt.validateStdout != nil {
		tt.validateStdout(t, stdout.String())
	}

	if tt.validateStderr != nil {
		tt.validateStderr(t, stderr.String())
	}

	cancel()
	require.Equal(t, context.Canceled, errors.Cause(<-errCh))
}

func createTestConfig(
	ctx context.Context,
	t *testing.T,
	services []manager.Service,
	boot []string,
) cfg.ConfigSet {
	// Set up boot target to start the given services.
	config := NewConfigurableSet(services).Configs()
	coreConfig := config["core"].(Config)
	coreConfig.ServiceGroups = []ServiceGroupConfig{{
		ID:       "boot",
		Name:     "boot",
		Desc:     "boot",
		Services: boot,
	}}
	config["core"] = coreConfig

	return config
}

func TestDeps(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	services, close := withValidServices(ctx, t)
	defer close()

	config := createTestConfig(ctx, t, services, []string{"host"})
	deps, err := Deps(services, config, "boot")

	require.NoError(t, err)
	assert.Equal(t, []string{"host", "boot"}, deps)
}

func TestDeps_noServiceID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	services, close := withValidServices(ctx, t)
	defer close()

	config := createTestConfig(ctx, t, services, []string{"host"})
	deps, err := Deps(services, config, "")

	require.NoError(t, err)
	assert.Equal(t, []string{"host", "boot"}, deps)
}

func TestFGraph(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	services, close := withValidServices(ctx, t)
	defer close()

	w := bytes.NewBuffer(nil)

	config := createTestConfig(ctx, t, services, []string{"host"})

	require.NoError(t, Fgraph(w, services, config, "boot"))

	assert.Equal(t, `boot─host
`, w.String())
}
