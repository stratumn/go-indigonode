/*
This code is based on:

	github.com/libp2p/go-libp2p/p2p/host/basic

Which has the license:

	The MIT License (MIT)

And has the copyright:

	Copyright (c) 2014 Juan Batiz-Benet

It has been modified to:

	- make more service friendly
	- combine the routed host package for DHT routing
	- log events
*/

package p2p

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"

	inet "gx/ipfs/QmNa31VPzC561NWwRsJLE7nGYZYuuD2QfpK2b1q9BK54J1/go-libp2p-net"
	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	mafilter "gx/ipfs/QmQBB2dQLmQHJgs2gqZ3iqL2XiuCtUCvXzWt5kMXDf5Zcr/go-maddr-filter"
	metrics "gx/ipfs/QmQbh3Rb7KM37As3vkHYnEFnzkVXNCP8EYGtHz6g2fXk14/go-libp2p-metrics"
	mstream "gx/ipfs/QmQbh3Rb7KM37As3vkHYnEFnzkVXNCP8EYGtHz6g2fXk14/go-libp2p-metrics/stream"
	madns "gx/ipfs/QmS7xUmsTdVNU2t1bPV6o9aXuXfufAjNGYgh2bcN2z9DAs/go-multiaddr-dns"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	msmux "gx/ipfs/QmTnsezaB1wWNRHeHnYrm8K4d5i9wtyj3GsqjC3Rt5b5v5/go-multistream"
	ma "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	ifconnmgr "gx/ipfs/QmYkCrTwivapqdB3JbwvwvxymseahVkcm46ThRMAA24zCr/go-libp2p-interface-connmgr"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	bhost "gx/ipfs/QmefgzMbKZYsmHFkLqxgaTBG9ypeEjrdWRD5WXH4j1cWDL/go-libp2p/p2p/host/basic"
	identify "gx/ipfs/QmefgzMbKZYsmHFkLqxgaTBG9ypeEjrdWRD5WXH4j1cWDL/go-libp2p/p2p/protocol/identify"
	circuit "gx/ipfs/Qmf7GSJ4omRJsvA9uzTqzbnVhq4RWLPzjzW4xJzUta4dKE/go-libp2p-circuit"
)

var (
	// ErrWrongPeerID is returned when the router returns addresses for
	// a different peer ID.
	ErrWrongPeerID = errors.New("addresses do not match peer ID")
)

// Host implements the go-libp2p-host.Host interface.
//
// TODO: Investigate data races durings tests when built with -race.
type Host struct {
	// Used to propagate event metadata.
	ctx context.Context

	netw     inet.Network
	mux      *msmux.MultistreamMuxer
	cmgr     ifconnmgr.ConnManager
	resolver *madns.Resolver

	negTimeout   time.Duration
	addrsFilters *mafilter.Filters

	natmgr bhost.NATManager
	ids    *identify.IDService
	router func(context.Context, peer.ID) (pstore.PeerInfo, error)

	bwc metrics.Reporter
}

// HostOption configures a host.
type HostOption func(*Host)

// OptConnManager adds a connection manager to a host.
func OptConnManager(cmgr ifconnmgr.ConnManager) HostOption {
	return func(h *Host) {
		h.cmgr = cmgr
	}
}

// OptResolver adds a resolver to a host.
func OptResolver(resolver *madns.Resolver) HostOption {
	return func(h *Host) {
		h.resolver = resolver
	}
}

// OptNegTimeout sets the negotiation timeout of a host.
func OptNegTimeout(timeout time.Duration) HostOption {
	return func(h *Host) {
		h.negTimeout = timeout
	}
}

// OptAddrsFilters adds addresses filters to a host.
func OptAddrsFilters(filter *mafilter.Filters) HostOption {
	return func(h *Host) {
		h.addrsFilters = filter
	}
}

// OptBandwidthReporter adds a bandwidth reporter to a host.
func OptBandwidthReporter(bwc metrics.Reporter) HostOption {
	return func(h *Host) {
		h.bwc = bwc
	}
}

// DefHostOpts are the default options for a host.
//
// These options are set before the options passed to NewHost are processed.
var DefHostOpts = []HostOption{
	OptConnManager(ifconnmgr.NullConnMgr{}),
	OptResolver(madns.DefaultResolver),
	OptNegTimeout(time.Minute),
}

// NewHost creates a new host.
func NewHost(ctx context.Context, netw inet.Network, opts ...HostOption) *Host {
	h := &Host{
		ctx:  ctx,
		netw: netw,
		mux:  msmux.NewMultistreamMuxer(),
	}

	for _, o := range DefHostOpts {
		o(h)
	}

	for _, o := range opts {
		o(h)
	}

	netw.SetConnHandler(h.newConnHandler)
	netw.SetStreamHandler(h.newStreamHandler)

	return h
}

// newConnHandler is the remote-opened conn handler for network.
func (h *Host) newConnHandler(conn inet.Conn) {
	ctx := logging.ContextWithLoggable(h.ctx, logging.Metadata{
		"conn": conn,
	})
	defer log.EventBegin(ctx, "newConnHandler").Done()

	pid := conn.RemotePeer()

	// Clear protocols on connecting to new peer to avoid issues caused
	// by misremembering protocols between reconnects.
	if err := h.Peerstore().SetProtocols(pid); err != nil {
		log.Event(ctx, "clearProtocolsError", logging.Metadata{
			"peerID": pid.Pretty(),
			"error":  err.Error(),
		})
	}

	if h.ids != nil {
		h.ids.IdentifyConn(conn)
	}
}

// newStreamHandler is the remote-opened stream handler for inet.Network.
// TODO: this feels a bit wonky
func (h *Host) newStreamHandler(stream inet.Stream) {
	ctx := logging.ContextWithLoggable(h.ctx, logging.Metadata{
		"stream": stream,
	})
	event := log.EventBegin(ctx, "newStreamHandler")
	defer event.Done()

	if h.negTimeout > 0 {
		err := stream.SetDeadline(time.Now().Add(h.negTimeout))
		if err != nil {
			event.SetError(err)

			if err := stream.Reset(); err != nil {
				log.Event(ctx, "streamResetError", logging.Metadata{
					"error": err.Error(),
				})
			}
			return
		}
	}

	lzc, protoID, handle, err := h.Mux().NegotiateLazy(stream)
	if err != nil {
		if err != io.EOF {
			event.SetError(err)
		}

		if err := stream.Reset(); err != nil {
			log.Event(ctx, "streamResetError", logging.Metadata{
				"error": err.Error(),
			})
		}
		return
	}

	stream = &streamWrapper{
		Stream: stream,
		rw:     lzc,
	}

	if h.negTimeout > 0 {
		if err := stream.SetDeadline(time.Time{}); err != nil {
			event.SetError(err)

			if err := stream.Reset(); err != nil {
				log.Event(ctx, "streamResetError", logging.Metadata{
					"error": err.Error(),
				})
			}
			return
		}
	}

	stream.SetProtocol(protocol.ID(protoID))

	if h.bwc != nil {
		stream = mstream.WrapStream(stream, h.bwc)
	}

	// Assumes handle lifecyle is already properly handled.
	go func() {
		err := handle(protoID, stream)
		if err != nil && errors.Cause(err) != context.Canceled {
			log.Event(ctx, "handleFailed")
		}
	}()
}

// ID returns the peer ID of the host.
func (h *Host) ID() peer.ID {
	return h.netw.LocalPeer()
}

// Peerstore returns the repository of known peers.
func (h *Host) Peerstore() pstore.Peerstore {
	return h.netw.Peerstore()
}

// Addrs returns the filtered addresses addresses of this host.
func (h *Host) Addrs() []ma.Multiaddr {
	var addrs []ma.Multiaddr

	allAddrs := h.AllAddrs()

	for _, address := range allAddrs {
		if h.addrsFilters != nil && h.addrsFilters.AddrBlocked(address) {
			continue
		}

		// Filter out relay addresses.
		// TODO: Should the relay service take care of adding a filter? It
		// would be worth it if multiple services need to add filters.
		_, err := address.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			continue
		}

		addrs = append(addrs, address)
	}

	return addrs
}

// AllAddrs returns all the addresses of BasicHost at this moment in time.
//
// It's ok to not include addresses if they're not available to be used now.
func (h *Host) AllAddrs() []ma.Multiaddr {
	addrs, err := h.netw.InterfaceListenAddresses()
	if err != nil {
		log.Event(h.ctx, "addrsError", logging.Metadata{
			"error": err.Error(),
		})
	}

	if h.ids != nil {
		// Add external observed addresses.
		addrs = append(addrs, h.ids.OwnObservedAddrs()...)
	}

	if h.natmgr != nil {
		nat := h.natmgr.NAT()
		if nat != nil {
			addrs = append(addrs, nat.ExternalAddrs()...)
		}
	}

	return addrs
}

// Network returns the network interface.
func (h *Host) Network() inet.Network {
	return h.netw
}

// Mux returns the multistream muxer.
func (h *Host) Mux() *msmux.MultistreamMuxer {
	return h.mux
}

// Connect ensures there is a connection between the host and the given peer.
func (h *Host) Connect(ctx context.Context, pi pstore.PeerInfo) error {
	ctx = logging.ContextWithLoggable(ctx, logging.Metadata(pi.Loggable()))
	event := log.EventBegin(ctx, "Connect")
	defer event.Done()

	ps := h.Peerstore()

	// Check if already connected.
	conns := h.Network().ConnsToPeer(pi.ID)
	if len(conns) > 0 {
		return nil
	}

	if len(pi.Addrs) > 0 {
		// Absorb addresses into peerstore.
		ps.AddAddrs(pi.ID, pi.Addrs, pstore.TempAddrTTL)
	}

	addrs := ps.Addrs(pi.ID)

	if len(addrs) < 1 && h.router != nil {
		// No addrs? Find some with the router.
		var err error
		addrs, err = h.findPeerAddrs(ctx, pi.ID)
		if err != nil {
			event.SetError(err)
			return err
		}

		// Absorb addresses into peerstore.
		ps.AddAddrs(pi.ID, addrs, pstore.TempAddrTTL)

	}

	pi.Addrs = addrs

	resolved, err := h.resolveAddrs(ctx, ps.PeerInfo(pi.ID))
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	ps.AddAddrs(pi.ID, resolved, pstore.TempAddrTTL)

	return h.dialPeer(ctx, pi.ID)
}

// findPeerAddrs finds addresses for a peer ID using the router.
func (h *Host) findPeerAddrs(ctx context.Context, id peer.ID) ([]ma.Multiaddr, error) {
	pi, err := h.router(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if pi.ID != id {
		return nil, errors.WithStack(ErrWrongPeerID)
	}

	return pi.Addrs, nil
}

// resolveAddrs resolves the multiaddresses of a peer using the DNS if
// necessary.
func (h *Host) resolveAddrs(ctx context.Context, pi pstore.PeerInfo) ([]ma.Multiaddr, error) {
	event := log.EventBegin(ctx, "resolveAddrs", logging.Metadata(pi.Loggable()))
	defer event.Done()

	// Create the IPFS P2P address of the peer.
	protocol := ma.ProtocolWithCode(ma.P_IPFS).Name
	p2pAddr, err := ma.NewMultiaddr("/" + protocol + "/" + pi.ID.Pretty())
	if err != nil {
		event.SetError(err)
		return nil, errors.WithStack(err)
	}

	var addrs []ma.Multiaddr

	for _, addr := range pi.Addrs {
		addrs = append(addrs, addr)
		if !madns.Matches(addr) {
			// Address doesn't need to be resolved with DNS.
			continue
		}

		reqAddr := addr.Encapsulate(p2pAddr)
		resAddrs, err := h.resolver.Resolve(ctx, reqAddr)
		if err != nil {
			log.Event(ctx, "resolveError", logging.Metadata{
				"error": err.Error(),
			})
			continue
		}

		for _, res := range resAddrs {
			pi, err := pstore.InfoFromP2pAddr(res)
			if err != nil {
				log.Event(ctx, "p2pAddrError", logging.Metadata{
					"error": err.Error(),
				})
				continue
			}

			addrs = append(addrs, pi.Addrs...)
		}
	}

	return addrs, nil
}

// dialPeer opens a connection to peer, and makes sure to identify the
// connection once it has been opened.
func (h *Host) dialPeer(ctx context.Context, pid peer.ID) error {
	event := log.EventBegin(ctx, "dialPeer", logging.Metadata{
		"peerID": pid.Pretty(),
	})
	defer event.Done()

	conn, err := h.Network().DialPeer(ctx, pid)
	if err != nil {
		return errors.WithStack(err)
	}

	// Clear protocols on connecting to new peer to avoid issues caused
	// by misremembering protocols between reconnects.
	if err := h.Peerstore().SetProtocols(pid); err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	// Identify the connection before returning.
	if h.ids != nil {
		done := make(chan struct{})
		go func() {
			h.ids.IdentifyConn(conn)
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
			err := errors.WithStack(ctx.Err())
			if err != nil {
				event.SetError(err)
			}
			return err
		}
	}

	return nil
}

// SetStreamHandler sets the stream handler for the given protocol.
func (h *Host) SetStreamHandler(proto protocol.ID, handler inet.StreamHandler) {
	h.mux.AddHandler(string(proto), func(protoStr string, rwc io.ReadWriteCloser) error {
		stream := rwc.(inet.Stream)
		stream.SetProtocol(protocol.ID(protoStr))
		handler(stream)

		return nil
	})

	log.Event(h.ctx, "setStreamHandler", logging.Metadata{
		"protocol": proto,
	})
}

// SetStreamHandlerMatch sets the protocol handler for protocols that match the
// given function.
func (h *Host) SetStreamHandlerMatch(proto protocol.ID, match func(string) bool, handler inet.StreamHandler) {
	h.mux.AddHandlerWithFunc(string(proto), match, func(protoStr string, rwc io.ReadWriteCloser) error {
		stream := rwc.(inet.Stream)
		stream.SetProtocol(protocol.ID(protoStr))
		handler(stream)

		return nil
	})

	log.Event(h.ctx, "setStreamHandlerMatch")
}

// RemoveStreamHandler removes the stream handler of the given protocol.
func (h *Host) RemoveStreamHandler(proto protocol.ID) {
	h.mux.RemoveHandler(string(proto))

	log.Event(h.ctx, "removeStreamHandler", logging.Metadata{
		"protocol": proto,
	})
}

// NewStream opens a new stream to the given peer for the given protocols.
func (h *Host) NewStream(ctx context.Context, pid peer.ID, protocols ...protocol.ID) (inet.Stream, error) {
	event := log.EventBegin(ctx, "NewStream", logging.Metadata{
		"peerID":    pid.Pretty(),
		"protocols": protocols,
	})
	defer event.Done()

	err := h.Connect(ctx, pstore.PeerInfo{ID: pid})
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	// Try to find a supported protocol.
	pref, err := h.preferredProtocol(pid, protocols)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	if pref != "" {
		return h.newStream(ctx, pid, pref)
	}

	var protoStrs []string
	for _, pid := range protocols {
		protoStrs = append(protoStrs, string(pid))
	}

	stream, err := h.Network().NewStream(ctx, pid)
	if err != nil {
		event.SetError(err)
		return nil, errors.WithStack(err)
	}

	selected, err := msmux.SelectOneOf(protoStrs, stream)
	if err != nil {
		if err := stream.Reset(); err != nil {
			log.Event(ctx, "streamResetError", logging.Metadata{
				"error":  err.Error(),
				"stream": stream,
			})
		}

		event.SetError(err)
		return nil, errors.WithStack(err)
	}

	selfpid := protocol.ID(selected)
	stream.SetProtocol(selfpid)

	if err := h.Peerstore().AddProtocols(pid, selected); err != nil {
		err = errors.WithStack(err)
		if err := stream.Reset(); err != nil {
			log.Event(ctx, "streamResetError", logging.Metadata{
				"error":  err.Error(),
				"stream": stream,
			})
		}

		event.SetError(err)
		return nil, err
	}

	if h.bwc != nil {
		stream = mstream.WrapStream(stream, h.bwc)
	}

	return stream, nil
}

// preferredProtocol finds the first protocol preferred by a peer.
func (h *Host) preferredProtocol(pid peer.ID, protocols []protocol.ID) (protocol.ID, error) {
	protoStrs := protocolsToStrings(protocols)
	supported, err := h.Peerstore().SupportsProtocols(pid, protoStrs...)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var out protocol.ID
	if len(supported) > 0 {
		out = protocol.ID(supported[0])
	}

	return out, nil
}

// newStream opens a stream to a peer for the given protocol.
func (h *Host) newStream(ctx context.Context, pid peer.ID, proto protocol.ID) (inet.Stream, error) {
	event := log.EventBegin(ctx, "newStream", logging.Metadata{
		"peerID":   pid.Pretty(),
		"protocol": proto,
	})
	defer event.Done()

	stream, err := h.Network().NewStream(ctx, pid)
	if err != nil {
		event.SetError(err)
		return nil, errors.WithStack(err)
	}

	stream.SetProtocol(proto)

	if h.bwc != nil {
		stream = mstream.WrapStream(stream, h.bwc)
	}

	lzcon := msmux.NewMSSelect(stream, string(proto))

	return &streamWrapper{
		Stream: stream,
		rw:     lzcon,
	}, nil
}

// Close shuts down the host and its network.
func (h *Host) Close() error {
	h.netw.SetStreamHandler(nil)
	h.netw.SetConnHandler(nil)

	return errors.WithStack(h.netw.Close())
}

// ConnManager returns the connection manager.
func (h *Host) ConnManager() ifconnmgr.ConnManager {
	return h.cmgr
}

// SetNATManager sets the NAT manager.
func (h *Host) SetNATManager(natmgr bhost.NATManager) {
	h.natmgr = natmgr

	if natmgr == nil {
		log.Event(h.ctx, "removeNATManager")
	}

	log.Event(h.ctx, "setNATManager")
}

// SetIDService sets the identity service.
func (h *Host) SetIDService(service *identify.IDService) {
	h.ids = service

	if service == nil {
		log.Event(h.ctx, "removeIDService")
	}

	log.Event(h.ctx, "setIDService")
}

// SetRouter sets the router.
func (h *Host) SetRouter(router func(context.Context, peer.ID) (pstore.PeerInfo, error)) {
	h.router = router

	if router == nil {
		log.Event(h.ctx, "removeRouter")
	}

	log.Event(h.ctx, "setRouter")
}

// protocolsToString converts protocol IDs to strings.
func protocolsToStrings(protocols []protocol.ID) []string {
	out := make([]string, len(protocols))
	for i, proto := range protocols {
		out[i] = string(proto)
	}
	return out
}

// streamWrapper adds lazy-negotiation to a stream.
type streamWrapper struct {
	inet.Stream
	rw io.ReadWriter
}

// Read reads bytes from the stream.
func (s *streamWrapper) Read(b []byte) (int, error) {
	return s.rw.Read(b)
}

// Write writes bytes to the stream.
func (s *streamWrapper) Write(b []byte) (int, error) {
	return s.rw.Write(b)
}
