package raft

import (
	"bytes"
	"context"
	"sort"
	"time"

	pb "github.com/stratumn/alice/grpc/raft"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var log = logging.Logger("raft")

type raftProcess struct {
	nodeID       uint64
	node         raft.Node
	storage      *raft.MemoryStorage
	peers        map[uint64][]byte
	committed    [][]byte
	ccID         uint64
	localPeer    []byte
	ticker       *time.Ticker
	tickerC      <-chan time.Time
	raftC        <-chan raft.Ready
	msgStartC    <-chan MessageStart
	msgStopC     chan MessageStop
	msgStatusC   <-chan MessageStatus
	msgPeersC    <-chan MessagePeers
	msgDiscoverC chan<- MessageDiscover
	msgInviteC   <-chan MessageInvite
	msgJoinC     <-chan MessageJoin
	msgExpelC    <-chan MessageExpel
	msgProposeC  <-chan MessagePropose
	msgLogC      <-chan MessageLog
	msgFromNetC  <-chan MessageRaft
	msgToNetC    chan<- MessageRaft
}

func NewRaftProcess(localPeer []byte,
	msgStartC <-chan MessageStart,
	msgStopC chan MessageStop,
	msgStatusC <-chan MessageStatus,
	msgPeersC <-chan MessagePeers,
	msgDiscoverC chan<- MessageDiscover,
	msgInviteC <-chan MessageInvite,
	msgJoinC <-chan MessageJoin,
	msgExpelC <-chan MessageExpel,
	msgProposeC <-chan MessagePropose,
	msgLogC <-chan MessageLog,
	msgFromNetC <-chan MessageRaft,
	msgToNetC chan<- MessageRaft,
) *raftProcess {
	return &raftProcess{
		localPeer:    localPeer,
		msgStartC:    msgStartC,
		msgStopC:     msgStopC,
		msgStatusC:   msgStatusC,
		msgPeersC:    msgPeersC,
		msgDiscoverC: msgDiscoverC,
		msgInviteC:   msgInviteC,
		msgJoinC:     msgJoinC,
		msgExpelC:    msgExpelC,
		msgProposeC:  msgProposeC,
		msgLogC:      msgLogC,
		msgFromNetC:  msgFromNetC,
		msgToNetC:    msgToNetC,
	}

}

func (r *raftProcess) Run(ctx context.Context) {
	for {
		select {
		case <-r.msgStartC:
			r.eventStart(ctx)
		case msg := <-r.msgJoinC:
			r.eventJoin(ctx, msg)
		case <-r.msgStopC:
			r.eventStop(ctx)
		case msg := <-r.msgStatusC:
			r.eventStatus(ctx, msg)
		case msg := <-r.msgPeersC:
			r.eventPeers(ctx, msg)
		case msg := <-r.msgInviteC:
			r.eventInvite(ctx, msg)
		case msg := <-r.msgExpelC:
			r.eventExpel(ctx, msg)
		case msg := <-r.msgProposeC:
			r.eventPropose(ctx, msg)
		case msg := <-r.msgLogC:
			r.eventLog(ctx, msg)
		case <-r.tickerC:
			r.eventTick(ctx)
		case rd := <-r.raftC:
			r.eventRaft(ctx, rd)
		case msg := <-r.msgFromNetC:
			r.eventStep(ctx, msg)
		case <-ctx.Done():
			r.eventDone(ctx)
			return
		}
	}
}

func (r *raftProcess) eventStart(ctx context.Context) {
	log.Event(ctx, "raftStart")
	if r.node != nil {
		log.Event(ctx, "errorStart", logging.Metadata{
			"error": "node already started",
		})
		return
	}

	var nodeID uint64 = 1

	r.peers = map[uint64][]byte{nodeID: r.localPeer}

	var peers []raft.Peer
	for id, address := range r.peers {
		peers = append(peers, raft.Peer{ID: id, Context: address})
	}

	r.start(nodeID, peers)
}

func (r *raftProcess) eventJoin(ctx context.Context, msg MessageJoin) {
	log.Event(ctx, "raftJoin")
	if r.node != nil {
		log.Event(ctx, "errorJoin", logging.Metadata{
			"error": "node already started",
		})
		return
	}

	peersC := make(chan pb.Peer)
	r.msgDiscoverC <- MessageDiscover{
		PeerID: msg.PeerID,
		PeersC: peersC,
	}

	peers := map[uint64][]byte{}
	var nodeID uint64
	// TODO: don't block
	for peer := range peersC {
		peers[peer.Id] = peer.Address
		if bytes.Compare(peer.Address, r.localPeer) == 0 {
			nodeID = peer.Id
		}
	}

	if nodeID == 0 {
		log.Event(ctx, "errorJoin", logging.Metadata{
			"error":     "cannot join: cannot find myself in discovered peers",
			"peers":     r.peers,
			"localPeer": r.localPeer,
		})
		return
	}

	r.peers = peers

	r.start(nodeID, nil)
}

func (r *raftProcess) eventStop(ctx context.Context) {
	log.Event(ctx, "raftStop")
	if r.node == nil {
		log.Event(ctx, "errorStop", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	r.stop()
}

func (r *raftProcess) eventStatus(ctx context.Context, msg MessageStatus) {
	log.Event(ctx, "raftStatus")
	msg.StatusInfoC <- pb.StatusInfo{Running: (r.node != nil), Id: r.nodeID}
}

func (r *raftProcess) eventPeers(ctx context.Context, msg MessagePeers) {
	log.Event(ctx, "raftPeers")

	defer close(msg.PeersC)

	keys := make([]uint64, len(r.peers))
	i := 0
	for k := range r.peers {
		keys[i] = k
		i += 1
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		msg.PeersC <- pb.Peer{Id: k, Address: r.peers[k]}
	}
}

func (r *raftProcess) eventInvite(ctx context.Context, msg MessageInvite) {
	log.Event(ctx, "raftInvite")
	if r.node == nil {
		log.Event(ctx, "errorInvite", logging.Metadata{
			"error": "node not started",
		})
		return
	}

	for _, address := range r.peers {
		if bytes.Compare(address, msg.PeerID.Address) == 0 {
			log.Event(ctx, "errorInvite", logging.Metadata{
				"error":   "peer already invited",
				"peers":   r.peers,
				"address": address,
			})
			return
		}
	}

	cc := raftpb.ConfChange{
		ID:      r.ccID + 1,
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  r.ccID + 2,
		Context: msg.PeerID.Address,
	}
	r.node.ProposeConfChange(ctx, cc)
}

func (r *raftProcess) eventExpel(ctx context.Context, msg MessageExpel) {
	log.Event(ctx, "raftExpel")
	if r.node == nil {
		log.Event(ctx, "errorExpel", logging.Metadata{
			"error": "node not started",
		})
		return
	}

	var expelID uint64
	for id, address := range r.peers {
		if bytes.Compare(address, msg.PeerID.Address) == 0 {
			expelID = id
			break
		}
	}

	if expelID == 0 {
		log.Event(ctx, "errorExpel", logging.Metadata{
			"error": "cannot resolve raft ID",
		})
		return
	}

	r.ccID += 1
	cc := raftpb.ConfChange{
		ID:     r.ccID,
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: expelID,
	}
	err := r.node.ProposeConfChange(ctx, cc)
	if err != nil {
		log.Event(ctx, "errorExpelProposeConfChange", logging.Metadata{
			"error": err,
		})
	}
}

func (r *raftProcess) eventPropose(ctx context.Context, msg MessagePropose) {
	log.Event(ctx, "raftPropose")
	if r.node == nil {
		log.Event(ctx, "errorPropose", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	r.node.Propose(ctx, msg.Proposal.Data)
}

func (r *raftProcess) eventLog(ctx context.Context, msg MessageLog) {
	log.Event(ctx, "raftLog")

	defer close(msg.EntriesC)

	for i := range r.committed {
		msg.EntriesC <- pb.Entry{Index: uint64(i), Data: r.committed[i]}
	}
}

func (r *raftProcess) eventTick(ctx context.Context) {
	log.Event(ctx, "raftTick")
	if r.node == nil {
		log.Event(ctx, "errorTick", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	r.node.Tick()
}

func (r *raftProcess) eventRaft(ctx context.Context, rd raft.Ready) {
	log.Event(ctx, "raftRaft")
	for i := range rd.Messages {
		rmsg := rd.Messages[i]
		address, ok := r.peers[rmsg.To]
		if !ok {
			log.Event(ctx, "errorResolve", logging.Metadata{
				"error": "cannot resolve address",
				"to":    rmsg.To,
				"peers": r.peers,
			})
			return
		}
		message := MessageRaft{
			PeerID:  pb.PeerID{Address: address},
			Message: rmsg,
		}
		r.msgToNetC <- message
	}
	r.storage.Append(rd.Entries)
	for _, entry := range rd.CommittedEntries {
		switch entry.Type {
		case raftpb.EntryConfChange:
			r.ccID += 1

			var cc raftpb.ConfChange
			cc.Unmarshal(entry.Data)
			r.node.ApplyConfChange(cc)
			log.Event(ctx, "confChangeApply", logging.Metadata{
				"newConfChange": cc,
			})

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				r.peers[cc.NodeID] = cc.Context
			case raftpb.ConfChangeRemoveNode:
				if bytes.Compare(r.peers[cc.NodeID], r.localPeer) == 0 && cc.NodeID == r.nodeID {
					log.Event(ctx, "stopDueToExpel")
					r.stop()
					return
				}
				delete(r.peers, cc.NodeID)
			}
		case raftpb.EntryNormal:
			// Ignore empty entries
			if len(entry.Data) == 0 {
				break
			}
			r.committed = append(r.committed, entry.Data)
		}
	}
	r.node.Advance()
}

func (r *raftProcess) eventStep(ctx context.Context, msg MessageRaft) {
	log.Event(ctx, "raftStep")
	if r.node == nil {
		log.Event(ctx, "errorStep", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	r.node.Step(ctx, msg.Message)
}

func (r *raftProcess) eventDone(ctx context.Context) {
	log.Event(ctx, "raftCtxDone")
	r.stop()
}

func (r *raftProcess) stop() {
	if r.node == nil {
		return // already stopped
	}
	r.node.Stop()
	r.ticker.Stop()
	r.node = nil
	r.storage = nil
	r.ticker = nil
	r.tickerC = nil
	r.raftC = nil
	r.nodeID = 0
	r.ccID = 0
	r.peers = nil
	r.committed = nil
}

func (r *raftProcess) start(nodeID uint64, peers []raft.Peer) {
	if r.node != nil {
		return // already started
	}

	r.nodeID = nodeID
	r.storage = raft.NewMemoryStorage()

	c := &raft.Config{
		ID:              nodeID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         r.storage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
	}

	r.ticker = time.NewTicker(100 * time.Millisecond)
	r.tickerC = r.ticker.C

	r.node = raft.StartNode(c, peers)
	r.raftC = r.node.Ready()
}
