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

// Runner mainly used to abstract away raftProcess and netProcess
type Runner interface {
	Run(ctx context.Context) error
}

type raftProcess struct {
	nodeID          uint64
	node            raft.Node
	storage         *raft.MemoryStorage
	peers           map[uint64][]byte
	committed       [][]byte
	configChangeID  uint64
	localPeer       []byte
	electionTick    int
	heartbeatTick   int
	maxSizePerMsg   uint64
	maxInflightMsgs int
	ticker          *time.Ticker
	tickerChan      <-chan time.Time
	raftChan        <-chan raft.Ready
	msgStartChan    <-chan MessageStart
	msgStopChan     chan MessageStop
	msgStatusChan   <-chan MessageStatus
	msgPeersChan    <-chan MessagePeers
	msgDiscoverChan chan<- MessageDiscover
	msgInviteChan   <-chan MessageInvite
	msgJoinChan     <-chan MessageJoin
	msgExpelChan    <-chan MessageExpel
	msgProposeChan  <-chan MessagePropose
	msgLogChan      <-chan MessageLog
	msgFromNetChan  <-chan MessageRaft
	msgToNetChan    chan<- MessageRaft
}

// NewRaftProcess creates an instance of raftProcess
func NewRaftProcess(
	localPeer []byte,
	electionTick int,
	heartbeatTick int,
	maxSizePerMsg uint64,
	maxInflightMsgs int,
	msgStartChan <-chan MessageStart,
	msgStopChan chan MessageStop,
	msgStatusChan <-chan MessageStatus,
	msgPeersChan <-chan MessagePeers,
	msgDiscoverChan chan<- MessageDiscover,
	msgInviteChan <-chan MessageInvite,
	msgJoinChan <-chan MessageJoin,
	msgExpelChan <-chan MessageExpel,
	msgProposeChan <-chan MessagePropose,
	msgLogChan <-chan MessageLog,
	msgFromNetChan <-chan MessageRaft,
	msgToNetChan chan<- MessageRaft,
) Runner {
	return &raftProcess{
		localPeer:       localPeer,
		electionTick:    electionTick,
		heartbeatTick:   heartbeatTick,
		maxSizePerMsg:   maxSizePerMsg,
		maxInflightMsgs: maxInflightMsgs,
		msgStartChan:    msgStartChan,
		msgStopChan:     msgStopChan,
		msgStatusChan:   msgStatusChan,
		msgPeersChan:    msgPeersChan,
		msgDiscoverChan: msgDiscoverChan,
		msgInviteChan:   msgInviteChan,
		msgJoinChan:     msgJoinChan,
		msgExpelChan:    msgExpelChan,
		msgProposeChan:  msgProposeChan,
		msgLogChan:      msgLogChan,
		msgFromNetChan:  msgFromNetChan,
		msgToNetChan:    msgToNetChan,
	}

}

func (r *raftProcess) Run(ctx context.Context) error {
	for {
		select {
		case <-r.msgStartChan:
			r.eventStart(ctx)
		case msg := <-r.msgJoinChan:
			r.eventJoin(ctx, msg)
		case <-r.msgStopChan:
			r.eventStop(ctx)
		case msg := <-r.msgStatusChan:
			r.eventStatus(ctx, msg)
		case msg := <-r.msgPeersChan:
			r.eventPeers(ctx, msg)
		case msg := <-r.msgInviteChan:
			r.eventInvite(ctx, msg)
		case msg := <-r.msgExpelChan:
			r.eventExpel(ctx, msg)
		case msg := <-r.msgProposeChan:
			r.eventPropose(ctx, msg)
		case msg := <-r.msgLogChan:
			r.eventLog(ctx, msg)
		case <-r.tickerChan:
			r.eventTick(ctx)
		case rd := <-r.raftChan:
			r.eventRaft(ctx, rd)
		case msg := <-r.msgFromNetChan:
			r.eventStep(ctx, msg)
		case <-ctx.Done():
			r.eventDone(ctx)
			return nil
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

	peersChan := make(chan pb.Peer)
	r.msgDiscoverChan <- MessageDiscover{
		PeerID:    msg.PeerID,
		PeersChan: peersChan,
	}

	peers := map[uint64][]byte{}
	var nodeID uint64

	for peer := range peersChan {
		peers[peer.Id] = peer.Address
		if bytes.Equal(peer.Address, r.localPeer) {
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
	msg.StatusInfoChan <- pb.StatusInfo{Running: (r.node != nil), Id: r.nodeID}
}

func (r *raftProcess) eventPeers(ctx context.Context, msg MessagePeers) {
	log.Event(ctx, "raftPeers")

	defer close(msg.PeersChan)

	keys := make([]uint64, len(r.peers))
	i := 0
	for k := range r.peers {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		msg.PeersChan <- pb.Peer{Id: k, Address: r.peers[k]}
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
		if bytes.Equal(address, msg.PeerID.Address) {
			log.Event(ctx, "errorInvite", logging.Metadata{
				"error":   "peer already invited",
				"peers":   r.peers,
				"address": address,
			})
			return
		}
	}

	cc := raftpb.ConfChange{
		ID:      r.configChangeID + 1,
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  r.configChangeID + 2,
		Context: msg.PeerID.Address,
	}
	err := r.node.ProposeConfChange(ctx, cc)
	if err != nil {
		log.Event(ctx, "errorInvite", logging.Metadata{
			"error": err,
		})
	}
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
		if bytes.Equal(address, msg.PeerID.Address) {
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

	r.configChangeID++
	cc := raftpb.ConfChange{
		ID:     r.configChangeID,
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
	err := r.node.Propose(ctx, msg.Proposal.Data)
	if err != nil {
		log.Event(ctx, "errorPropose", logging.Metadata{
			"error": err,
		})
	}
}

func (r *raftProcess) eventLog(ctx context.Context, msg MessageLog) {
	log.Event(ctx, "raftLog")

	defer close(msg.EntriesChan)

	for i := range r.committed {
		msg.EntriesChan <- pb.Entry{Index: uint64(i), Data: r.committed[i]}
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
			continue
		}
		message := MessageRaft{
			PeerID:  pb.PeerID{Address: address},
			Message: rmsg,
		}
		r.msgToNetChan <- message
	}
	err := r.storage.Append(rd.Entries)
	if err != nil {
		log.Event(ctx, "errorStorageAppend", logging.Metadata{
			"error": err,
		})
	}
	for _, entry := range rd.CommittedEntries {
		switch entry.Type {
		case raftpb.EntryConfChange:
			r.configChangeID++

			var cc raftpb.ConfChange
			err = cc.Unmarshal(entry.Data)
			if err != nil {
				log.Event(ctx, "errorUnmarshal", logging.Metadata{
					"error": err,
				})
				continue
			}
			r.node.ApplyConfChange(cc)
			log.Event(ctx, "confChangeApply", logging.Metadata{
				"newConfChange": cc,
			})

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				r.peers[cc.NodeID] = cc.Context
			case raftpb.ConfChangeRemoveNode:
				if bytes.Equal(r.peers[cc.NodeID], r.localPeer) && cc.NodeID == r.nodeID {
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
	err := r.node.Step(ctx, msg.Message)
	if err != nil {
		log.Event(ctx, "errorStep", logging.Metadata{
			"error": err,
		})
	}
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
	r.tickerChan = nil
	r.raftChan = nil
	r.nodeID = 0
	r.configChangeID = 0
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
		ElectionTick:    r.electionTick,
		HeartbeatTick:   r.heartbeatTick,
		Storage:         r.storage,
		MaxSizePerMsg:   r.maxSizePerMsg,
		MaxInflightMsgs: r.maxInflightMsgs,
	}

	r.ticker = time.NewTicker(100 * time.Millisecond)
	r.tickerChan = r.ticker.C

	r.node = raft.StartNode(c, peers)
	r.raftChan = r.node.Ready()
}
