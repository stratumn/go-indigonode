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

package core

import (
	"github.com/stratumn/alice/core/manager"
	"github.com/stratumn/alice/core/service/bootstrap"
	"github.com/stratumn/alice/core/service/chat"
	"github.com/stratumn/alice/core/service/clock"
	"github.com/stratumn/alice/core/service/coin"
	"github.com/stratumn/alice/core/service/connmgr"
	"github.com/stratumn/alice/core/service/contacts"
	"github.com/stratumn/alice/core/service/event"
	"github.com/stratumn/alice/core/service/grpcapi"
	"github.com/stratumn/alice/core/service/host"
	"github.com/stratumn/alice/core/service/identify"
	"github.com/stratumn/alice/core/service/kaddht"
	"github.com/stratumn/alice/core/service/metrics"
	"github.com/stratumn/alice/core/service/mssmux"
	"github.com/stratumn/alice/core/service/natmgr"
	"github.com/stratumn/alice/core/service/ping"
	"github.com/stratumn/alice/core/service/pruner"
	"github.com/stratumn/alice/core/service/relay"
	"github.com/stratumn/alice/core/service/signal"
	"github.com/stratumn/alice/core/service/swarm"
	"github.com/stratumn/alice/core/service/yamux"
)

// BuiltinServices returns all the builtin services.
func BuiltinServices() []manager.Service {
	return []manager.Service{
		&bootstrap.Service{},
		&chat.Service{},
		&clock.Service{},
		&coin.Service{},
		&connmgr.Service{},
		&contacts.Service{},
		&event.Service{},
		&grpcapi.Service{},
		&host.Service{},
		&identify.Service{},
		&kaddht.Service{},
		&metrics.Service{},
		&mssmux.Service{},
		&natmgr.Service{},
		&ping.Service{},
		&pruner.Service{},
		&relay.Service{},
		&signal.Service{},
		&swarm.Service{},
		&yamux.Service{},
	}
}

// registerServices registers all the given services as well as groups defined
// in the configuration on the given manager.
//
// It assumes that the manager's Work function has been called.
//
// It is safe to call multiple times.
func registerServices(mgr *manager.Manager, services []manager.Service, config *Config) {
	// Register the manager service.
	mgr.RegisterService()

	// Register all the services.
	for _, serv := range services {
		mgr.Register(serv)
	}

	// Add services for groups.
	for _, config := range config.ServiceGroups {
		group := manager.ServiceGroup{
			GroupID:   config.ID,
			GroupName: config.Name,
			GroupDesc: config.Desc,
			Services:  map[string]struct{}{},
		}

		for _, dep := range config.Services {
			group.Services[dep] = struct{}{}
		}

		mgr.Register(&group)
	}
}
