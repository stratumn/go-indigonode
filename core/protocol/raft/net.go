package raft

import (
	"context"
	"encoding/gob"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/raft"

	inet "gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net"
	// protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	ihost "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

type host = ihost.Host

var discoverProtocol = protocol.ID("/alice/raft-discover/v1.0.0")
var raftProtocol = protocol.ID("/alice/raft/v1.0.0")

type netProcess struct {
	host            host
	errorChan       chan error
	msgDiscoverChan <-chan MessageDiscover
	msgPeersChan    chan<- MessagePeers
	msgFromNetChan  chan<- MessageRaft
	msgToNetChan    <-chan MessageRaft
}

// NewNetProcess creates a runnable instance of netProcess
func NewNetProcess(host host,
	msgDiscoverChan <-chan MessageDiscover,
	msgPeersChan chan<- MessagePeers,
	msgFromNetChan chan<- MessageRaft,
	msgToNetChan <-chan MessageRaft,
) Runner {

	n := netProcess{
		host:            host,
		errorChan:       make(chan error),
		msgDiscoverChan: msgDiscoverChan,
		msgPeersChan:    msgPeersChan,
		msgFromNetChan:  msgFromNetChan,
		msgToNetChan:    msgToNetChan,
	}

	// TODO: wrappers adding context for graceful termination
	host.SetStreamHandler(discoverProtocol, n.discoverHandler)
	host.SetStreamHandler(raftProtocol, n.raftHandler)

	return &n
}

func (n *netProcess) Run(ctx context.Context) error {

	for {
		select {
		case err := <-n.errorChan:
			log.Event(ctx, "errorNetProcess", logging.Metadata{
				"error": err.Error(),
			})
		case msg := <-n.msgDiscoverChan:
			go n.discoverClient(ctx, msg)
		case msg := <-n.msgToNetChan:
			go n.raftClient(ctx, msg)
		case <-ctx.Done():
			// TODO: close all streams?
			return nil
		}
	}
}

func (n *netProcess) closeStream(stream inet.Stream) {
	if err := stream.Close(); err != nil {
		n.errorChan <- err
	}
}

// TODO: server stream timeout?
func (n *netProcess) discoverHandler(stream inet.Stream) {
	defer n.closeStream(stream)

	// enc := protobuf.Multicodec(nil).Encoder(stream)
	enc := gob.NewEncoder(stream)

	peersChan := make(chan pb.Peer)
	n.msgPeersChan <- MessagePeers{
		PeersChan: peersChan,
	}

	peers := []pb.Peer{}
	for peer := range peersChan {
		peers = append(peers, peer)
	}

	err := enc.Encode(peers)
	if err != nil {
		n.errorChan <- err
		return
	}

}

// TODO: client stream timeout?
func (n *netProcess) discoverClient(ctx context.Context, msgDiscover MessageDiscover) {

	defer close(msgDiscover.PeersChan)

	pid, err := peer.IDFromBytes(msgDiscover.PeerID.Address)
	if err != nil {
		n.errorChan <- err
		return
	}

	stream, err := n.host.NewStream(ctx, pid, discoverProtocol)
	if err != nil {
		n.errorChan <- err
		return
	}
	defer n.closeStream(stream)

	// dec := protobuf.Multicodec(nil).Decoder(stream)
	dec := gob.NewDecoder(stream)

	var peers []pb.Peer

	err = dec.Decode(&peers)
	if err != nil {
		n.errorChan <- err
		return
	}

	for i := range peers {
		msgDiscover.PeersChan <- peers[i]
	}
}

// TODO: server stream timeout?
func (n *netProcess) raftHandler(stream inet.Stream) {

	defer n.closeStream(stream)
	// dec := protobuf.Multicodec(nil).Decoder(stream)
	dec := gob.NewDecoder(stream)

	var message MessageRaft

	err := dec.Decode(&message)
	if err != nil {
		n.errorChan <- errors.WithStack(err)
		return
	}

	n.msgFromNetChan <- message

}

// TODO: client stream timeout?
func (n *netProcess) raftClient(ctx context.Context, msgRaft MessageRaft) {

	pid, err := peer.IDFromBytes(msgRaft.PeerID.Address)
	if err != nil {
		n.errorChan <- errors.WithStack(err)
		return
	}

	stream, err := n.host.NewStream(ctx, pid, raftProtocol)
	if err != nil {
		n.errorChan <- errors.WithStack(err)
		return
	}
	defer n.closeStream(stream)

	// enc := protobuf.Multicodec(nil).Encoder(stream)
	enc := gob.NewEncoder(stream)
	if err != nil {
		n.errorChan <- errors.WithStack(err)
		return
	}

	err = enc.Encode(msgRaft)
	if err != nil {
		n.errorChan <- errors.WithStack(err)
		return
	}

}
