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

package lib

import (
	"context"
)

// EntryType distinguishes normal entry from config change entry
type EntryType int32

const (
	// EntryNormal is a normal entry
	EntryNormal EntryType = 0
	// EntryConfChange is a config change entry
	EntryConfChange EntryType = 1
)

// ConfChangeType distinguishes add node from node removal
type ConfChangeType int32

const (
	//ConfChangeAddNode adds node
	ConfChangeAddNode ConfChangeType = 0
	//ConfChangeRemoveNode removes node
	ConfChangeRemoveNode ConfChangeType = 1
)

//Lib is root-level library abstraction
type Lib interface {
	// StartNode starts a node
	StartNode(Config, []Peer) Node
	// NewConfig creates a configuration
	NewConfig(uint64, int, int, Storage, uint64, int) Config
	// NewPeer creates a peer
	NewPeer(uint64, []byte) Peer
	// NewConfChange creates a config change
	NewConfChange(uint64, ConfChangeType, uint64, []byte) ConfChange
	// NewMemoryStorage creates a memory storage
	NewMemoryStorage() Storage
}

// Node abstracts node
type Node interface {
	// ProposeConfChange proposes configuration change
	ProposeConfChange(context.Context, ConfChange) error
	// Propose proposes some data
	Propose(context.Context, []byte) error
	// Tick is a clock event
	Tick()
	// Advace advances state machine
	Advance()
	// Step tells state machine to process a message
	Step(context.Context, []byte) error
	// Stop stops a node
	Stop()
	// ApplyConfChange applies configuration change
	ApplyConfChange(ConfChange)
	// Ready returns a channel with events
	Ready() <-chan Ready
}

// Storage abstracts storage interface
type Storage interface {
	// Append adds entries to a storage
	Append([]Entry) error
	// InitialState returns initial state
	InitialState() (HardState, ConfState, error)
	// Entries returns entries in a given range
	Entries(uint64, uint64, uint64) ([]Entry, error)
	// Term returns current term
	Term(i uint64) (uint64, error)
	// LastIndex returns last index
	LastIndex() (uint64, error)
	// FirstIndex returns first index
	FirstIndex() (uint64, error)
	// Snapshot returns snapshot
	Snapshot() (Snapshot, error)
}

// HardState abstracts state machine state
type HardState interface {
	// Term returns term
	Term() uint64
	// Vote returns vote
	Vote() uint64
	// Commit returns commit
	Commit() uint64
}

// ConfState abstracts state machine configuration state
type ConfState interface {
	// Nodes returns cluster nodes
	Nodes() []uint64
	// Learners returns non-voting nodes
	Learners() []uint64
}

// Snapshot abstracts snapshot structure
type Snapshot interface {
	// Data represents a serialized snapshot
	Data() []byte
	// Metadata brings additional information about the snapshot
	Metadata() SnapshotMetadata
}

// SnapshotMetadata keeps additional information about the snapshot
type SnapshotMetadata interface {
	// ConfState returns configuration state
	ConfState() ConfState
	// Index returns index
	Index() uint64
	// Term returns term
	Term() uint64
}

// Ready aggregates messages, entries and committed entries
type Ready interface {
	// Messages returns the slice of messages
	Messages() []Message
	// Entries returns the slice of entries
	Entries() []Entry
	// CommittedEntries returns the slice of committed entries
	CommittedEntries() []Entry
}

// Config keeps the configuration of the state-machine
type Config interface {
	// ID returns node ID in a cluster
	ID() uint64
	// ElectionTick is the number of Node.Tick invocations that must pass between elections
	ElectionTick() int
	// HeartbeatTick is the number of Node.Tick invocations that must pass between heartbeats
	HeartbeatTick() int
	// Storage returns storage object
	Storage() Storage
	// MaxSizePerMsg limits the max size of each append message
	MaxSizePerMsg() uint64
	// MaxInflightMsgs limits the max number of in-flight append messages during optimistic replication phase
	MaxInflightMsgs() int
}

// ConfChange represents configuration change
type ConfChange interface {
	// ID returns configuration change ID
	ID() uint64
	// Type returns configuration change type
	Type() ConfChangeType
	// NodeID returns node id concerned
	NodeID() uint64
	// Context returs some associated data
	Context() []byte
	// Unmarshal deserializes the structure
	Unmarshal([]byte) error
}

// Peer represents peer information
type Peer interface {
	// ID returns peer ID
	ID() uint64
	// Context returns peer information, such as address
	Context() []byte
}

// Entry aggregates information about log entry
type Entry interface {
	// Term returns entry term
	Term() uint64
	// Index returns entry index
	Index() uint64
	// Type returns entry type
	Type() EntryType
	// Data returns entry payload
	Data() []byte
}

// Message represents the amount of data exchanged between nodes
type Message interface {
	// To return destination ID
	To() uint64
	// Data returns message payload
	Data() []byte
}
