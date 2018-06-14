package protocol

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/stratumn/alice/app/raft/pb"
	"github.com/stratumn/alice/app/raft/protocol/lib"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var log = logging.Logger("circle")

// Circle exports internal circle process structure
type Circle interface {
	Run(ctx context.Context) error
}

type circleProcess struct {
	lib                lib.Lib
	node               lib.Node
	storage            lib.Storage
	peers              []lib.Peer
	nodeID             uint64
	committed          [][]byte
	configChangeID     uint64
	lastNodeID         uint64
	localPeer          []byte
	electionTick       int
	heartbeatTick      int
	maxSizePerMsg      uint64
	maxInflightMsgs    int
	tickerInterval     uint64
	ticker             *time.Ticker
	tickerChan         <-chan time.Time
	circleChan         <-chan lib.Ready
	callStartChan      <-chan hubCallStart
	callStopChan       chan hubCallStop
	callStatusChan     <-chan hubCallStatus
	callPeersChan      <-chan hubCallPeers
	callDiscoverChan   chan<- hubCallDiscover
	callInviteChan     <-chan hubCallInvite
	callJoinChan       <-chan hubCallJoin
	callExpelChan      <-chan hubCallExpel
	callProposeChan    <-chan hubCallPropose
	callLogChan        <-chan hubCallLog
	msgNetToCircleChan <-chan pb.Internode
	msgCircleToNetChan chan<- pb.Internode
}

// NewCircleProcess creates an instance of circleProcess
func NewCircleProcess(
	lib lib.Lib,
	hub *Hub,
	localPeer []byte,
	electionTick int,
	heartbeatTick int,
	maxSizePerMsg uint64,
	maxInflightMsgs int,
	tickerInterval uint64,
) Circle {
	c := &circleProcess{
		lib:             lib,
		localPeer:       localPeer,
		electionTick:    electionTick,
		heartbeatTick:   heartbeatTick,
		maxSizePerMsg:   maxSizePerMsg,
		maxInflightMsgs: maxInflightMsgs,
		tickerInterval:  tickerInterval,
	}
	hub.bindCircle(c)
	return c

}

// Run starts the process
func (c *circleProcess) Run(ctx context.Context) error {
	fmt.Printf("waiting from discovery %+v\n", c.callDiscoverChan)
	for {
		select {
		case <-c.callStartChan:
			c.eventStart(ctx)
		case msg := <-c.callJoinChan:
			c.eventJoin(ctx, msg)
		case <-c.callStopChan:
			c.eventStop(ctx)
		case msg := <-c.callStatusChan:
			c.eventStatus(ctx, msg)
		case msg := <-c.callPeersChan:
			c.eventPeers(ctx, msg)
		case msg := <-c.callInviteChan:
			c.eventInvite(ctx, msg)
		case msg := <-c.callExpelChan:
			c.eventExpel(ctx, msg)
		case msg := <-c.callProposeChan:
			c.eventPropose(ctx, msg)
		case msg := <-c.callLogChan:
			c.eventLog(ctx, msg)
		case <-c.tickerChan:
			c.eventTick(ctx)
		case rd := <-c.circleChan:
			c.eventCircle(ctx, rd)
		case msg := <-c.msgNetToCircleChan:
			c.eventStep(ctx, msg)
		case <-ctx.Done():
			c.eventDone(ctx)
			return nil
		}
	}
}

func (c *circleProcess) eventStart(ctx context.Context) {
	log.Event(ctx, "eventStart")
	if c.node != nil {
		log.Event(ctx, "errorStart", logging.Metadata{
			"error": "node already started",
		})
		return
	}

	var nodeID uint64 = 1

	c.peers = []lib.Peer{c.lib.NewPeer(nodeID, c.localPeer)}

	c.start(nodeID, c.peers)
}

func (c *circleProcess) eventJoin(ctx context.Context, msg hubCallJoin) {
	log.Event(ctx, "eventJoin")
	if c.node != nil {
		log.Event(ctx, "errorJoin", logging.Metadata{
			"error": "node already started",
		})
		return
	}

	peersChan := make(chan pb.Peer)
	c.callDiscoverChan <- hubCallDiscover{
		PeerID:    msg.PeerID,
		PeersChan: peersChan,
	}

	var nodeID uint64
	var peers []lib.Peer

	for peer := range peersChan {
		peers = append(peers, c.lib.NewPeer(peer.Id, peer.Address))
		if bytes.Equal(peer.Address, c.localPeer) {
			nodeID = peer.Id
		}
	}

	if nodeID == 0 {
		log.Event(ctx, "errorJoin", logging.Metadata{
			"error":     "cannot join: cannot find myself in discovered peers",
			"peers":     c.peers,
			"localPeer": c.localPeer,
		})
		return
	}

	c.peers = peers

	c.start(nodeID, nil)
}

func (c *circleProcess) eventStop(ctx context.Context) {
	log.Event(ctx, "eventStop")
	if c.node == nil {
		log.Event(ctx, "errorStop", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	c.stop()
}

func (c *circleProcess) eventStatus(ctx context.Context, msg hubCallStatus) {
	log.Event(ctx, "eventStatus")
	msg.StatusInfoChan <- pb.StatusInfo{Running: (c.node != nil), Id: c.nodeID}
}

func (c *circleProcess) eventPeers(ctx context.Context, msg hubCallPeers) {
	log.Event(ctx, "eventPeers")

	defer close(msg.PeersChan)

	sort.Slice(c.peers, func(i, j int) bool { return c.peers[i].ID() < c.peers[j].ID() })

	for _, peer := range c.peers {
		msg.PeersChan <- pb.Peer{Id: peer.ID(), Address: peer.Context()}
	}
}

func (c *circleProcess) eventInvite(ctx context.Context, msg hubCallInvite) {
	log.Event(ctx, "eventInvite")
	if c.node == nil {
		log.Event(ctx, "errorInvite", logging.Metadata{
			"error": "node not started",
		})
		return
	}

	for _, peer := range c.peers {
		if bytes.Equal(peer.Context(), msg.PeerID.Address) {
			log.Event(ctx, "errorInvite", logging.Metadata{
				"error":   "peer already invited",
				"peers":   c.peers,
				"address": peer.Context(),
			})
			return
		}
	}

	log.Event(ctx, "inviteConfigChange", logging.Metadata{
		"confChangeID": c.configChangeID,
		"lastNodeID":   c.lastNodeID,
	})
	cc := c.lib.NewConfChange(
		c.configChangeID,
		lib.ConfChangeAddNode,
		c.lastNodeID+1,
		msg.PeerID.Address,
	)
	err := c.node.ProposeConfChange(ctx, cc)
	if err != nil {
		log.Event(ctx, "errorInvite", logging.Metadata{
			"error": err,
		})
		return
	}
}

func (c *circleProcess) eventExpel(ctx context.Context, msg hubCallExpel) {
	log.Event(ctx, "eventExpel")
	if c.node == nil {
		log.Event(ctx, "errorExpel", logging.Metadata{
			"error": "node not started",
		})
		return
	}

	var expelID uint64
	for _, peer := range c.peers {
		if bytes.Equal(peer.Context(), msg.PeerID.Address) {
			expelID = peer.ID()
			break
		}
	}

	if expelID == 0 {
		log.Event(ctx, "errorExpel", logging.Metadata{
			"error": "cannot resolve ID",
		})
		return
	}

	cc := c.lib.NewConfChange(
		c.configChangeID,
		lib.ConfChangeRemoveNode,
		expelID,
		nil,
	)
	err := c.node.ProposeConfChange(ctx, cc)
	if err != nil {
		log.Event(ctx, "errorExpelProposeConfChange", logging.Metadata{
			"error": err,
		})
		return
	}
}

func (c *circleProcess) eventPropose(ctx context.Context, msg hubCallPropose) {
	log.Event(ctx, "eventPropose")
	if c.node == nil {
		log.Event(ctx, "errorPropose", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	err := c.node.Propose(ctx, msg.Proposal.Data)
	if err != nil {
		log.Event(ctx, "errorPropose", logging.Metadata{
			"error": err,
		})
	}
}

func (c *circleProcess) eventLog(ctx context.Context, msg hubCallLog) {
	log.Event(ctx, "eventLog")

	defer close(msg.EntriesChan)

	for i := range c.committed {
		msg.EntriesChan <- pb.Entry{Index: uint64(i), Data: c.committed[i]}
	}
}

func (c *circleProcess) eventTick(ctx context.Context) {
	log.Event(ctx, "eventTick")
	if c.node == nil {
		log.Event(ctx, "errorTick", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	c.node.Tick()
}

func (c *circleProcess) eventCircle(ctx context.Context, rd lib.Ready) {
	log.Event(ctx, "eventCircle")
	for _, rmsg := range rd.Messages() {
		c.processMessage(ctx, rmsg)
	}
	err := c.storage.Append(rd.Entries())
	if err != nil {
		log.Event(ctx, "errorStorageAppend", logging.Metadata{
			"error": err,
		})
	}
	for _, entry := range rd.CommittedEntries() {
		c.processCommittedEntry(ctx, entry)
	}
	c.node.Advance()
}

func (c *circleProcess) eventStep(ctx context.Context, msg pb.Internode) {
	log.Event(ctx, "eventStep")
	if c.node == nil {
		log.Event(ctx, "errorStep", logging.Metadata{
			"error": "node not started",
		})
		return
	}
	err := c.node.Step(ctx, msg.Message)
	if err != nil {
		log.Event(ctx, "errorStep", logging.Metadata{
			"error": err,
		})
	}
}

func (c *circleProcess) eventDone(ctx context.Context) {
	log.Event(ctx, "eventDone")
	c.stop()
}

func (c *circleProcess) processMessage(ctx context.Context, rmsg lib.Message) {
	var address []byte
	for _, peer := range c.peers {
		if peer.ID() == rmsg.To() {
			address = peer.Context()
			break
		}
	}

	if address == nil {
		log.Event(ctx, "errorResolve", logging.Metadata{
			"error": "cannot resolve address",
			"to":    rmsg.To(),
			"peers": c.peers,
		})
		return
	}
	message := pb.Internode{
		PeerId:  &pb.PeerID{Address: address},
		Message: rmsg.Data(),
	}
	c.msgCircleToNetChan <- message
}

func (c *circleProcess) processCommittedEntry(ctx context.Context, entry lib.Entry) {
	switch entry.Type() {
	case lib.EntryNormal:
		// Ignore empty entries
		if len(entry.Data()) != 0 {
			c.committed = append(c.committed, entry.Data())
		}
	case lib.EntryConfChange:
		c.configChangeID++

		cc := c.lib.NewConfChange(0, lib.ConfChangeAddNode, 0, nil)
		err := cc.Unmarshal(entry.Data())
		if err != nil {
			log.Event(ctx, "errorUnmarshal", logging.Metadata{
				"error": err,
			})
			return
		}
		c.node.ApplyConfChange(cc)
		log.Event(ctx, "confChangeApply", logging.Metadata{
			"newConfChange": cc,
			"confChangeID":  c.configChangeID,
			"lastNodeID":    c.lastNodeID,
		})

		switch cc.Type() {
		case lib.ConfChangeAddNode:
			c.lastNodeID++

			var id uint64
			for _, peer := range c.peers {
				if peer.ID() == cc.NodeID() {
					id = cc.NodeID()
					break
				}
			}

			if id != 0 {
				log.Event(ctx, "confChangeAddNodeError", logging.Metadata{
					"error": "adding existing peer",
					"id":    id,
				})
				return
			}

			c.peers = append(c.peers, c.lib.NewPeer(cc.NodeID(), cc.Context()))

		case lib.ConfChangeRemoveNode:

			var id uint64
			var address []byte
			var i int
			for _, peer := range c.peers {
				if peer.ID() == cc.NodeID() {
					id = cc.NodeID()
					address = peer.Context()
					break
				}
				i++
			}

			if id == 0 {
				log.Event(ctx, "confChangeRemoveNodeError", logging.Metadata{
					"error": "removing non-existing peer",
					"id":    id,
				})
				return
			}

			if bytes.Equal(address, c.localPeer) && cc.NodeID() == c.nodeID {
				log.Event(ctx, "stopDueToExpel")
				go func() {
					c.callStopChan <- hubCallStop{}
				}()
				return
			}

			c.peers = append(c.peers[:i], c.peers[i+1:]...)
		}
	}
}

func (c *circleProcess) stop() {
	if c.node == nil {
		return // already stopped
	}
	c.node.Stop()
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}
	c.node = nil
	c.storage = nil
	c.tickerChan = nil
	c.circleChan = nil
	c.nodeID = 0
	c.configChangeID = 0
	c.lastNodeID = 0
	c.peers = nil
	c.committed = nil
}

func (c *circleProcess) start(nodeID uint64, peers []lib.Peer) {
	if c.node != nil {
		return // already started
	}

	c.nodeID = nodeID

	c.ticker = time.NewTicker(time.Duration(c.tickerInterval) * time.Millisecond)
	c.tickerChan = c.ticker.C

	c.storage = c.lib.NewMemoryStorage()

	config := c.lib.NewConfig(
		c.nodeID,
		c.electionTick,
		c.heartbeatTick,
		c.storage,
		c.maxSizePerMsg,
		c.maxInflightMsgs,
	)

	c.node = c.lib.StartNode(config, peers)
	c.circleChan = c.node.Ready()
}
