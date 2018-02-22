package raft

import (
	"context"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"

	"github.com/stratumn/alice/core/protocol/circle/lib"
)

type raftImpl struct{}

func (raftImpl) NewMemoryStorage() lib.Storage {
	return raftStorageImpl{
		s: raft.NewMemoryStorage(),
	}
}

func (raftImpl) StartNode(c lib.Config, pp []lib.Peer) lib.Node {
	peers := make([]raft.Peer, len(pp))

	for i := range pp {
		peers[i] = raft.Peer{ID: pp[i].ID(), Context: pp[i].Context()}
	}

	rc := raft.Config{
		ID:            c.ID(),
		ElectionTick:  c.ElectionTick(),
		HeartbeatTick: c.HeartbeatTick(),
		Storage: raftStorageReverseImpl{
			s: c.Storage(),
		},
		MaxSizePerMsg:   c.MaxSizePerMsg(),
		MaxInflightMsgs: c.MaxInflightMsgs(),
	}

	return raftNodeImpl{
		n: raft.StartNode(&rc, peers),
	}
}

func (raftImpl) NewConfChange(id uint64, ccType lib.ConfChangeType, nodeID uint64, context []byte) lib.ConfChange {
	return &raftConfChangeImpl{
		cc: raftpb.ConfChange{
			ID:      id,
			Type:    raftpb.ConfChangeType(ccType),
			NodeID:  nodeID,
			Context: context,
		},
	}
}

func (raftImpl) NewConfig(id uint64, electionTick int, heartbeatTick int, storage lib.Storage, maxSizePerMsg uint64, maxInflightMsgs int) lib.Config {
	return raftConfigImpl{
		c: raft.Config{
			ID:              id,
			ElectionTick:    electionTick,
			HeartbeatTick:   heartbeatTick,
			MaxSizePerMsg:   maxSizePerMsg,
			MaxInflightMsgs: maxInflightMsgs,
		},
		s: storage,
	}
}

func (raftImpl) NewPeer(id uint64, context []byte) lib.Peer {
	return raftPeerImpl{
		p: raft.Peer{
			ID:      id,
			Context: context,
		},
	}
}

type raftNodeImpl struct {
	n raft.Node
}

func (n raftNodeImpl) Ready() <-chan lib.Ready {
	rChan := make(chan lib.Ready)

	rrChan := n.n.Ready()

	go func() {
		for {
			rr, ok := <-rrChan
			if !ok {
				close(rChan)
				break
			}
			rChan <- raftReadyImpl{r: rr}
		}
	}()

	return rChan
}
func (n raftNodeImpl) Propose(ctx context.Context, data []byte) error {
	return n.n.Propose(ctx, data)
}
func (n raftNodeImpl) Tick() {
	n.n.Tick()
}
func (n raftNodeImpl) Advance() {
	n.n.Advance()
}
func (n raftNodeImpl) Step(ctx context.Context, msg []byte) error {
	rm := raftpb.Message{}
	err := rm.Unmarshal(msg)
	if err != nil {
		return err
	}
	return n.n.Step(ctx, rm)
}
func (n raftNodeImpl) Stop() {
	n.n.Stop()
}
func (n raftNodeImpl) ApplyConfChange(cc lib.ConfChange) {
	rcc := raftpb.ConfChange{
		ID:      cc.ID(),
		Type:    raftpb.ConfChangeType(cc.Type()),
		NodeID:  cc.NodeID(),
		Context: cc.Context(),
	}
	n.n.ApplyConfChange(rcc)
}
func (n raftNodeImpl) ProposeConfChange(ctx context.Context, cc lib.ConfChange) error {
	rcc := raftpb.ConfChange{
		ID:      cc.ID(),
		Type:    raftpb.ConfChangeType(cc.Type()),
		NodeID:  cc.NodeID(),
		Context: cc.Context(),
	}
	return n.n.ProposeConfChange(ctx, rcc)
}

type raftReadyImpl struct {
	r raft.Ready
}

func (r raftReadyImpl) Messages() []lib.Message {
	mm := make([]lib.Message, len(r.r.Messages))
	for i := range r.r.Messages {
		mm[i] = raftMessageImpl{m: r.r.Messages[i]}
	}
	return mm
}
func (r raftReadyImpl) Entries() []lib.Entry {
	ee := make([]lib.Entry, len(r.r.Entries))
	for i := range r.r.Entries {
		ee[i] = raftEntryImpl{e: r.r.Entries[i]}
	}
	return ee
}
func (r raftReadyImpl) CommittedEntries() []lib.Entry {
	ee := make([]lib.Entry, len(r.r.CommittedEntries))
	for i := range r.r.CommittedEntries {
		ee[i] = raftEntryImpl{e: r.r.CommittedEntries[i]}
	}
	return ee
}

type raftConfigImpl struct {
	c raft.Config
	s lib.Storage
}

func (c raftConfigImpl) ID() uint64 {
	return c.c.ID
}
func (c raftConfigImpl) ElectionTick() int {
	return c.c.ElectionTick
}
func (c raftConfigImpl) HeartbeatTick() int {
	return c.c.HeartbeatTick
}
func (c raftConfigImpl) Storage() lib.Storage {
	return c.s
}
func (c raftConfigImpl) MaxSizePerMsg() uint64 {
	return c.c.MaxSizePerMsg
}
func (c raftConfigImpl) MaxInflightMsgs() int {
	return c.c.MaxInflightMsgs
}

type raftConfChangeImpl struct {
	cc raftpb.ConfChange
}

func (cc raftConfChangeImpl) ID() uint64 {
	return cc.cc.ID
}
func (cc raftConfChangeImpl) Type() lib.ConfChangeType {
	return lib.ConfChangeType(cc.cc.Type)
}
func (cc raftConfChangeImpl) NodeID() uint64 {
	return cc.cc.NodeID
}
func (cc raftConfChangeImpl) Context() []byte {
	return cc.cc.Context
}
func (cc *raftConfChangeImpl) Unmarshal(data []byte) error {
	rcc := raftpb.ConfChange{}
	err := rcc.Unmarshal(data)
	if err != nil {
		return err
	}
	cc.cc = rcc
	return nil
}

type raftPeerImpl struct {
	p raft.Peer
}

func (p raftPeerImpl) ID() uint64 {
	return p.p.ID
}

func (p raftPeerImpl) Context() []byte {
	return p.p.Context
}

type raftEntryImpl struct {
	e raftpb.Entry
}

func (e raftEntryImpl) Term() uint64 {
	return e.e.Term
}
func (e raftEntryImpl) Index() uint64 {
	return e.e.Index
}
func (e raftEntryImpl) Type() lib.EntryType {
	return lib.EntryType(e.e.Type)
}
func (e raftEntryImpl) Data() []byte {
	return e.e.Data
}

type raftMessageImpl struct {
	m raftpb.Message
}

func (m raftMessageImpl) To() uint64 {
	return m.m.To
}

func (m raftMessageImpl) Data() []byte {
	data, _ := m.m.Marshal()
	return data
}

type raftStorageImpl struct {
	s *raft.MemoryStorage
}

func (s raftStorageImpl) Append(ee []lib.Entry) error {
	ree := make([]raftpb.Entry, len(ee))
	for i := range ee {
		ree[i] = raftpb.Entry{
			Term:  ee[i].Term(),
			Index: ee[i].Index(),
			Type:  raftpb.EntryType(ee[i].Type()),
			Data:  ee[i].Data(),
		}
	}
	return s.s.Append(ree)
}

func (s raftStorageImpl) InitialState() (lib.HardState, lib.ConfState, error) {
	rhs, rcs, err := s.s.InitialState()

	hs := raftHardStateImpl{
		hs: rhs,
	}

	cs := raftConfStateImpl{
		cs: rcs,
	}

	return hs, cs, err
}

func (s raftStorageImpl) Entries(lo uint64, hi uint64, maxSize uint64) ([]lib.Entry, error) {
	ree, err := s.s.Entries(lo, hi, maxSize)
	ee := make([]lib.Entry, len(ree))
	for i := range ree {
		ee[i] = raftEntryImpl{
			e: ree[i],
		}
	}

	return ee, err
}

func (s raftStorageImpl) Term(i uint64) (uint64, error) {
	return s.s.Term(i)
}
func (s raftStorageImpl) LastIndex() (uint64, error) {
	return s.s.LastIndex()
}
func (s raftStorageImpl) FirstIndex() (uint64, error) {
	return s.s.FirstIndex()
}
func (s raftStorageImpl) Snapshot() (lib.Snapshot, error) {
	rs, err := s.s.Snapshot()

	return raftSnapshotImpl{
		s: rs,
	}, err
}

type raftSnapshotImpl struct {
	s raftpb.Snapshot
}

func (s raftSnapshotImpl) Data() []byte {
	return s.s.Data
}
func (s raftSnapshotImpl) Metadata() lib.SnapshotMetadata {
	return raftSnapshotMetadataImpl{
		m: s.s.Metadata,
	}
}

type raftHardStateImpl struct {
	hs raftpb.HardState
}

func (hs raftHardStateImpl) Term() uint64 {
	return hs.hs.Term
}
func (hs raftHardStateImpl) Vote() uint64 {
	return hs.hs.Vote
}
func (hs raftHardStateImpl) Commit() uint64 {
	return hs.hs.Commit
}

type raftConfStateImpl struct {
	cs raftpb.ConfState
}

func (cs raftConfStateImpl) Nodes() []uint64 {
	return cs.cs.Nodes
}
func (cs raftConfStateImpl) Learners() []uint64 {
	return cs.cs.Learners
}

type raftSnapshotMetadataImpl struct {
	m raftpb.SnapshotMetadata
}

func (m raftSnapshotMetadataImpl) ConfState() lib.ConfState {
	return raftConfStateImpl{
		cs: m.m.ConfState,
	}
}
func (m raftSnapshotMetadataImpl) Index() uint64 {
	return m.m.Index
}
func (m raftSnapshotMetadataImpl) Term() uint64 {
	return m.m.Term
}

type raftStorageReverseImpl struct {
	s lib.Storage
}

func (s raftStorageReverseImpl) InitialState() (raftpb.HardState, raftpb.ConfState, error) {
	hs, cs, err := s.s.InitialState()

	rhs := raftpb.HardState{
		Term:   hs.Term(),
		Vote:   hs.Vote(),
		Commit: hs.Commit(),
	}

	rcs := raftpb.ConfState{
		Nodes:    cs.Nodes(),
		Learners: cs.Learners(),
	}

	return rhs, rcs, err
}

func (s raftStorageReverseImpl) Entries(lo uint64, hi uint64, maxSize uint64) ([]raftpb.Entry, error) {
	ee, err := s.s.Entries(lo, hi, maxSize)

	ree := make([]raftpb.Entry, len(ee))
	for i, e := range ee {
		ree[i] = raftpb.Entry{
			Term:  e.Term(),
			Index: e.Index(),
			Type:  raftpb.EntryType(e.Type()),
			Data:  e.Data(),
		}
	}
	return ree, err
}
func (s raftStorageReverseImpl) Term(i uint64) (uint64, error) {
	return s.s.Term(i)
}
func (s raftStorageReverseImpl) LastIndex() (uint64, error) {
	return s.s.LastIndex()
}
func (s raftStorageReverseImpl) FirstIndex() (uint64, error) {
	return s.s.FirstIndex()
}
func (s raftStorageReverseImpl) Snapshot() (raftpb.Snapshot, error) {
	snap, err := s.s.Snapshot()
	m := snap.Metadata()
	cs := m.ConfState()

	rsnap := raftpb.Snapshot{
		Data: snap.Data(),
		Metadata: raftpb.SnapshotMetadata{
			ConfState: raftpb.ConfState{
				Nodes:    cs.Nodes(),
				Learners: cs.Learners(),
			},
			Index: m.Index(),
			Term:  m.Term(),
		},
	}

	return rsnap, err
}
