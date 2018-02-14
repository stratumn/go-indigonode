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

type Host = ihost.Host

var discoverProtocol = protocol.ID("/alice/raft-discover/v1.0.0")
var raftProtocol = protocol.ID("/alice/raft/v1.0.0")

type netProcess struct {
	Host         Host
	errorC       chan error
	msgDiscoverC <-chan MessageDiscover
	msgPeersC    chan<- MessagePeers
	msgFromNetC  chan<- MessageRaft
	msgToNetC    <-chan MessageRaft
}

func NewNetProcess(host Host,
	msgDiscoverC <-chan MessageDiscover,
	msgPeersC chan<- MessagePeers,
	msgFromNetC chan<- MessageRaft,
	msgToNetC <-chan MessageRaft,
) Runner {

	n := netProcess{
		Host:         host,
		errorC:       make(chan error),
		msgDiscoverC: msgDiscoverC,
		msgPeersC:    msgPeersC,
		msgFromNetC:  msgFromNetC,
		msgToNetC:    msgToNetC,
	}

	// TODO: wrappers adding context for graceful termination
	host.SetStreamHandler(discoverProtocol, n.discoverHandler)
	host.SetStreamHandler(raftProtocol, n.raftHandler)

	return &n
}

func (n *netProcess) Run(ctx context.Context) error {

	for {
		select {
		case err := <-n.errorC:
			log.Event(ctx, "errorNetProcess", logging.Metadata{
				"error": err.Error(),
			})
		case msg := <-n.msgDiscoverC:
			go n.discoverClient(ctx, msg)
		case msg := <-n.msgToNetC:
			go n.raftClient(ctx, msg)
		case <-ctx.Done():
			// TODO: close all streams?
			return nil
		}
	}
}

// TODO: server stream timeout?
func (n *netProcess) discoverHandler(stream inet.Stream) {
	defer stream.Close()

	// enc := protobuf.Multicodec(nil).Encoder(stream)
	enc := gob.NewEncoder(stream)

	peersC := make(chan pb.Peer)
	n.msgPeersC <- MessagePeers{
		PeersC: peersC,
	}

	peers := []pb.Peer{}
	for peer := range peersC {
		peers = append(peers, peer)
	}

	err := enc.Encode(peers)
	if err != nil {
		n.errorC <- err
		return
	}

}

// TODO: client stream timeout?
func (n *netProcess) discoverClient(ctx context.Context, msgDiscover MessageDiscover) {

	defer close(msgDiscover.PeersC)

	pid, err := peer.IDFromBytes(msgDiscover.PeerID.Address)
	if err != nil {
		n.errorC <- err
		return
	}

	stream, err := n.Host.NewStream(ctx, pid, discoverProtocol)
	if err != nil {
		n.errorC <- err
		return
	}
	defer stream.Close()

	// dec := protobuf.Multicodec(nil).Decoder(stream)
	dec := gob.NewDecoder(stream)

	var peers []pb.Peer

	err = dec.Decode(&peers)
	if err != nil {
		n.errorC <- err
		return
	}

	for i, _ := range peers {
		msgDiscover.PeersC <- peers[i]
	}
}

// TODO: server stream timeout?
func (n *netProcess) raftHandler(stream inet.Stream) {

	defer stream.Close()
	// dec := protobuf.Multicodec(nil).Decoder(stream)
	dec := gob.NewDecoder(stream)

	var message MessageRaft

	err := dec.Decode(&message)
	if err != nil {
		n.errorC <- errors.WithStack(err)
		return
	}

	n.msgFromNetC <- message

}

// TODO: client stream timeout?
func (n *netProcess) raftClient(ctx context.Context, msgRaft MessageRaft) {

	pid, err := peer.IDFromBytes(msgRaft.PeerID.Address)
	if err != nil {
		n.errorC <- errors.WithStack(err)
		return
	}

	stream, err := n.Host.NewStream(ctx, pid, raftProtocol)
	if err != nil {
		n.errorC <- errors.WithStack(err)
		return
	}
	defer stream.Close()

	// enc := protobuf.Multicodec(nil).Encoder(stream)
	enc := gob.NewEncoder(stream)
	if err != nil {
		n.errorC <- errors.WithStack(err)
		return
	}

	err = enc.Encode(msgRaft)
	if err != nil {
		n.errorC <- errors.WithStack(err)
		return
	}

}
