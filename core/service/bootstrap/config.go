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

package bootstrap

// Config contains configuration options for the Bootstrap service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Needs are services that should be started in addition to the host
	// before bootstrapping.
	Needs []string `toml:"needs" comment:"Services that should be started in addition to the host before bootstrapping."`

	// Addresses is a list of known peer addresses.
	Addresses []string `toml:"addresses" comment:"A list of known peer addresses."`

	// MinPeerThreshold is the number of peers under which to bootstrap
	// connections.
	MinPeerThreshold int `toml:"min_peer_threshold" comment:"The number of peers under which to bootstrap connections."`

	// Interval is the duration of the interval between bootstrap jobs.
	Interval string `toml:"interval" comment:"Interval between bootstrap jobs."`

	// ConnectionTimeout is the connection timeout. It should be less than
	// the interval.
	ConnectionTimeout string `toml:"connection_timeout" comment:"The connection timeout. It should be less than the interval."`
}
