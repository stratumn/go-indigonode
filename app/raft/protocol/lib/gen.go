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

//go:generate mockgen -package mocklib -destination mocklib/mocklib.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Lib
//go:generate mockgen -package mocklib -destination mocklib/mocknode.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Node
//go:generate mockgen -package mocklib -destination mocklib/mockstorage.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Storage
//go:generate mockgen -package mocklib -destination mocklib/mockhardstate.go github.com/stratumn/go-indigonode/app/raft/protocol/lib HardState
//go:generate mockgen -package mocklib -destination mocklib/mockconfstate.go github.com/stratumn/go-indigonode/app/raft/protocol/lib ConfState
//go:generate mockgen -package mocklib -destination mocklib/mocksnapshot.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Snapshot
//go:generate mockgen -package mocklib -destination mocklib/mocksnapshotmetadata.go github.com/stratumn/go-indigonode/app/raft/protocol/lib SnapshotMetadata
//go:generate mockgen -package mocklib -destination mocklib/mockready.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Ready
//go:generate mockgen -package mocklib -destination mocklib/mockconfig.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Config
//go:generate mockgen -package mocklib -destination mocklib/mockconfchange.go github.com/stratumn/go-indigonode/app/raft/protocol/lib ConfChange
//go:generate mockgen -package mocklib -destination mocklib/mockpeer.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Peer
//go:generate mockgen -package mocklib -destination mocklib/mockentry.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Entry
//go:generate mockgen -package mocklib -destination mocklib/mockmessage.go github.com/stratumn/go-indigonode/app/raft/protocol/lib Message

package lib
