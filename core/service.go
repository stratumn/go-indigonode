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
	chat "github.com/stratumn/go-indigonode/app/chat/service"
	clock "github.com/stratumn/go-indigonode/app/clock/service"
	coin "github.com/stratumn/go-indigonode/app/coin/service"
	contacts "github.com/stratumn/go-indigonode/app/contacts/service"
	indigofossilizer "github.com/stratumn/go-indigonode/app/indigo/service/fossilizer"
	indigostore "github.com/stratumn/go-indigonode/app/indigo/service/store"
	raft "github.com/stratumn/go-indigonode/app/raft/service"
	storage "github.com/stratumn/go-indigonode/app/storage/service"
	bootstrap "github.com/stratumn/go-indigonode/core/app/bootstrap/service"
	connmgr "github.com/stratumn/go-indigonode/core/app/connmgr/service"
	event "github.com/stratumn/go-indigonode/core/app/event/service"
	grpcapi "github.com/stratumn/go-indigonode/core/app/grpcapi/service"
	grpcweb "github.com/stratumn/go-indigonode/core/app/grpcweb/service"
	host "github.com/stratumn/go-indigonode/core/app/host/service"
	identify "github.com/stratumn/go-indigonode/core/app/identify/service"
	kaddht "github.com/stratumn/go-indigonode/core/app/kaddht/service"
	metrics "github.com/stratumn/go-indigonode/core/app/metrics/service"
	mssmux "github.com/stratumn/go-indigonode/core/app/mssmux/service"
	natmgr "github.com/stratumn/go-indigonode/core/app/natmgr/service"
	ping "github.com/stratumn/go-indigonode/core/app/ping/service"
	pruner "github.com/stratumn/go-indigonode/core/app/pruner/service"
	pubsub "github.com/stratumn/go-indigonode/core/app/pubsub/service"
	relay "github.com/stratumn/go-indigonode/core/app/relay/service"
	signal "github.com/stratumn/go-indigonode/core/app/signal/service"
	swarm "github.com/stratumn/go-indigonode/core/app/swarm/service"
	yamux "github.com/stratumn/go-indigonode/core/app/yamux/service"
	"github.com/stratumn/go-indigonode/core/manager"
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
		&grpcweb.Service{},
		&host.Service{},
		&identify.Service{},
		&indigostore.Service{},
		&indigofossilizer.Service{},
		&kaddht.Service{},
		&metrics.Service{},
		&mssmux.Service{},
		&natmgr.Service{},
		&ping.Service{},
		&pruner.Service{},
		&pubsub.Service{},
		&raft.Service{},
		&relay.Service{},
		&signal.Service{},
		&swarm.Service{},
		&storage.Service{},
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
