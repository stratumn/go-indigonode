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

// Mocks for libp2p components:
//go:generate mockgen -package mocks -destination mocks/mockconn.go github.com/libp2p/go-libp2p-net Conn
//go:generate mockgen -package mocks -destination mocks/mockhost.go github.com/libp2p/go-libp2p-host Host
//go:generate mockgen -package mocks -destination mocks/mocklogger.go github.com/ipfs/go-log EventLogger
//go:generate mockgen -package mocks -destination mocks/mocknetwork.go github.com/libp2p/go-libp2p-net Network
//go:generate mockgen -package mocks -destination mocks/mockpeerstore.go github.com/libp2p/go-libp2p-peerstore Peerstore
//go:generate mockgen -package mocks -destination mocks/mockstream.go github.com/libp2p/go-libp2p-net Stream
//go:generate mockgen -package mocks -destination mocks/mockstreammuxer.go github.com/libp2p/go-stream-muxer Transport
//go:generate mockgen -package mocks -destination mocks/mocktransport.go -mock_names Conn=MockTransportConn github.com/libp2p/go-libp2p-transport Conn

// Mocks for opencensus components:
//go:generate mockgen -package mocks -destination mocks/mockexporter.go go.opencensus.io/stats/view Exporter

// Mocks for standard libs:
//go:generate mockgen -package mocks -destination mocks/mocknetconn.go -mock_names Conn=MockNetConn net Conn

package test
