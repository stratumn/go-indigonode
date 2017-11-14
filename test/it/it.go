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

// Package it defines types for doing integration tests.
//
// It facilitates launching a network of nodes each in a separate process and
// run tests against them via the API.
package it

import (
	"fmt"
	"io"
	"strings"

	// Make sure all sevices are registered.
	"github.com/mohae/deepcopy"
	"github.com/stratumn/alice/core"
	"github.com/stratumn/alice/core/cfg"
	logging "github.com/stratumn/alice/core/log"
	"github.com/stratumn/alice/core/service/bootstrap"
	"github.com/stratumn/alice/core/service/grpcapi"
	"github.com/stratumn/alice/core/service/metrics"
)

// NumSeeds is the number of seeds in a bootstrap list.
const NumSeeds = 5

// IntegrationCfg returns a base configuration for integration testing.
func IntegrationCfg() cfg.ConfigSet {
	conf := deepcopy.Copy(core.GlobalConfigSet().Configs()).(cfg.ConfigSet)

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
	conf := deepcopy.Copy(core.GlobalConfigSet().Configs()).(cfg.ConfigSet)

	coreConf := conf["core"].(core.Config)
	coreConf.EnableBootScreen = false
	conf["core"] = coreConf

	logConf := conf["log"].(logging.Config)
	logConf.Writers = []logging.WriterConfig{{
		Type:      logging.Stderr,
		Level:     logging.Error,
		Formatter: logging.JSON,
	}}

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
// settings except that it changes the boot service so that the given services
// will be started.
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
