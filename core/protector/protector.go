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

//go:generate mockgen -package mocks -destination mocks/mockpeerstore.go gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore Peerstore
//go:generate mockgen -package mocks -destination mocks/mocktransport.go gx/ipfs/QmPUHzTLPZFYqv8WqcBTuMFYTgeom4uHHEaxzk7bd5GYZB/go-libp2p-transport Conn

// Package protector contains implementations of the
// github.com/libp2p/go-libp2p-interface-pnet/ipnet.Protector interface.
//
// Use these implementations in the swarm service to protect a private network.
package protector

import (
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var log = logging.Logger("core.protector")

var (
	// ErrConnectionRefused is returned when a connection is refused.
	ErrConnectionRefused = errors.New("connection refused")
)
