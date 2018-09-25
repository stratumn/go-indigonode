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
	chat "github.com/stratumn/go-node/app/chat/service"
	clock "github.com/stratumn/go-node/app/clock/service"
	coin "github.com/stratumn/go-node/app/coin/service"
	contacts "github.com/stratumn/go-node/app/contacts/service"
	raft "github.com/stratumn/go-node/app/raft/service"
	storage "github.com/stratumn/go-node/app/storage/service"
	bootstrap "github.com/stratumn/go-node/core/app/bootstrap/service"
	connmgr "github.com/stratumn/go-node/core/app/connmgr/service"
	event "github.com/stratumn/go-node/core/app/event/service"
	grpcapi "github.com/stratumn/go-node/core/app/grpcapi/service"
	grpcweb "github.com/stratumn/go-node/core/app/grpcweb/service"
	host "github.com/stratumn/go-node/core/app/host/service"
	identify "github.com/stratumn/go-node/core/app/identify/service"
	kaddht "github.com/stratumn/go-node/core/app/kaddht/service"
	monitoring "github.com/stratumn/go-node/core/app/monitoring/service"
	mssmux "github.com/stratumn/go-node/core/app/mssmux/service"
	natmgr "github.com/stratumn/go-node/core/app/natmgr/service"
	ping "github.com/stratumn/go-node/core/app/ping/service"
	pruner "github.com/stratumn/go-node/core/app/pruner/service"
	pubsub "github.com/stratumn/go-node/core/app/pubsub/service"
	relay "github.com/stratumn/go-node/core/app/relay/service"
	signal "github.com/stratumn/go-node/core/app/signal/service"
	swarm "github.com/stratumn/go-node/core/app/swarm/service"
	yamux "github.com/stratumn/go-node/core/app/yamux/service"
	"github.com/stratumn/go-node/core/manager"
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
		&kaddht.Service{},
		&monitoring.Service{},
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
