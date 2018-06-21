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

// Package service defines a service for the Yamux stream multiplexer.
//
// For more information about Yamux, see:
//
//  	https://github.com/hashicorp/yamux
package service

import (
	"io/ioutil"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"

	yamux "gx/ipfs/QmNWCEvi7bPRcvqAV8AKLGVNoQdArWi7NJayka2SM4XtRe/go-smux-yamux"
	smux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
)

// Service is the Yamux service.
type Service struct {
	config *Config
	tpt    smux.Transport
}

// Config contains configuration options for the Yamux service.
type Config struct {
	// AcceptBacklog is the size of the accept backlog.
	AcceptBacklog int `toml:"accept_backlog" comment:"The size of the accept backlog."`

	// ConnectionWriteTimeout is the connection write timeout.
	ConnectionWriteTimeout string `toml:"connection_write_timeout" comment:"The connection write timeout."`

	// KeepAliveInterval is the keep alive interval.
	KeepAliveInterval string `toml:"keep_alive_interval" comment:"The keep alive interval."`

	// MaxStreamWindowSize is the maximum stream window size.
	MaxStreamWindowSize string `toml:"max_stream_window_size" comment:"The maximum stream window size."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "yamux"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Yamux"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Multiplexes streams using Yamux."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		AcceptBacklog:          512,
		ConnectionWriteTimeout: "10s",
		KeepAliveInterval:      "30s",
		MaxStreamWindowSize:    "512KB",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	var err error

	conf := config.(Config)

	tpt := yamux.Transport{
		AcceptBacklog:   conf.AcceptBacklog,
		EnableKeepAlive: true,
		LogOutput:       ioutil.Discard, // TODO: debug
	}

	tpt.ConnectionWriteTimeout, err = time.ParseDuration(conf.ConnectionWriteTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	tpt.KeepAliveInterval, err = time.ParseDuration(conf.KeepAliveInterval)
	if err != nil {
		return errors.WithStack(err)
	}

	size, err := bytefmt.ToBytes(conf.MaxStreamWindowSize)
	if err != nil {
		return errors.WithStack(err)
	}

	tpt.MaxStreamWindowSize = uint32(size)

	s.tpt = &tpt
	s.config = &conf

	return nil
}

// Expose exposes the stream muxer to other services.
//
// It exposes the type:
//	github.com/libp2p/go-stream-muxer.Transport
func (s *Service) Expose() interface{} {
	return s.tpt
}
