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

//go:generate mockgen -package mocklib -destination mocklib/mocklib.go github.com/stratumn/alice/app/raft/protocol/lib Lib
//go:generate mockgen -package mocklib -destination mocklib/mocknode.go github.com/stratumn/alice/app/raft/protocol/lib Node
//go:generate mockgen -package mocklib -destination mocklib/mockstorage.go github.com/stratumn/alice/app/raft/protocol/lib Storage
//go:generate mockgen -package mocklib -destination mocklib/mockhardstate.go github.com/stratumn/alice/app/raft/protocol/lib HardState
//go:generate mockgen -package mocklib -destination mocklib/mockconfstate.go github.com/stratumn/alice/app/raft/protocol/lib ConfState
//go:generate mockgen -package mocklib -destination mocklib/mocksnapshot.go github.com/stratumn/alice/app/raft/protocol/lib Snapshot
//go:generate mockgen -package mocklib -destination mocklib/mocksnapshotmetadata.go github.com/stratumn/alice/app/raft/protocol/lib SnapshotMetadata
//go:generate mockgen -package mocklib -destination mocklib/mockready.go github.com/stratumn/alice/app/raft/protocol/lib Ready
//go:generate mockgen -package mocklib -destination mocklib/mockconfig.go github.com/stratumn/alice/app/raft/protocol/lib Config
//go:generate mockgen -package mocklib -destination mocklib/mockconfchange.go github.com/stratumn/alice/app/raft/protocol/lib ConfChange
//go:generate mockgen -package mocklib -destination mocklib/mockpeer.go github.com/stratumn/alice/app/raft/protocol/lib Peer
//go:generate mockgen -package mocklib -destination mocklib/mockentry.go github.com/stratumn/alice/app/raft/protocol/lib Entry
//go:generate mockgen -package mocklib -destination mocklib/mockmessage.go github.com/stratumn/alice/app/raft/protocol/lib Message

package lib
