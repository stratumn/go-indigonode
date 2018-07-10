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
	"github.com/stratumn/go-indigonode/core/monitoring"

	madns "gx/ipfs/QmQMRYmPn77CKRFf4YFjX3M5e6uw6DFAgsQffCX6mwZ4mA/go-multiaddr-dns"
	msmux "gx/ipfs/QmTnsezaB1wWNRHeHnYrm8K4d5i9wtyj3GsqjC3Rt5b5v5/go-multistream"
	metrics "gx/ipfs/QmVvu4bS5QLfS19ePkp5Wgzn2ZUma5oXTT9BgDFyQLxUZF/go-libp2p-metrics"
	mstream "gx/ipfs/QmVvu4bS5QLfS19ePkp5Wgzn2ZUma5oXTT9BgDFyQLxUZF/go-libp2p-metrics/stream"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	bhost "gx/ipfs/QmWsV6kzPaYGBDVyuUfWBvyQygEc9Qrv9vzo8vZ7X4mdLN/go-libp2p/p2p/host/basic"
	identify "gx/ipfs/QmWsV6kzPaYGBDVyuUfWBvyQygEc9Qrv9vzo8vZ7X4mdLN/go-libp2p/p2p/protocol/identify"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	circuit "gx/ipfs/QmZRbCo2gw7ghw5m7L77a8FvvQTVr62J4hmy8ozpdq7dHF/go-libp2p-circuit"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	pstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	mafilter "gx/ipfs/Qmf2UAmRwDG4TvnkQpHZWPAzw7rpCYVhxmRXmYxXr5LD1g/go-maddr-filter"
	ifconnmgr "gx/ipfs/QmfQNieWBPwmnUjXWPZbjJPzhNwFFabTb5RQ79dyVWGujQ/go-libp2p-interface-connmgr"
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

	bwc  metrics.Reporter
	tick *time.Ticker
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

// OptMetricsInterval sets the interval at which metrics are collected.
func OptMetricsInterval(interval time.Duration) HostOption {
	return func(h *Host) {
		h.tick = time.NewTicker(interval)
	}
}

// DefHostOpts are the default options for a host.
//
// These options are set before the options passed to NewHost are processed.
var DefHostOpts = []HostOption{
	OptConnManager(ifconnmgr.NullConnMgr{}),
	OptResolver(madns.DefaultResolver),
	OptNegTimeout(time.Minute),
	OptMetricsInterval(time.Second),
}

// NewHost creates a new host.
func NewHost(ctx context.Context, netw inet.Network, opts ...HostOption) *Host {
	h := &Host{
		ctx:  ctx,
		netw: netw,
		mux:  msmux.NewMultistreamMuxer(),
		bwc:  &MetricsReporter{},
	}

	for _, o := range DefHostOpts {
		o(h)
	}

	for _, o := range opts {
		o(h)
	}

	netw.SetConnHandler(h.newConnHandler)
	netw.SetStreamHandler(h.newStreamHandler)

	// This go-routine will be stopped when Host.Close() is called.
	go h.collectMetrics()

	return h
}

// newConnHandler is the remote-opened conn handler for network.
func (h *Host) newConnHandler(conn inet.Conn) {
	_, span := monitoring.StartSpan(context.Background(), "p2p", "newConnHandler")
	defer span.End()

	pid := conn.RemotePeer()
	span.SetPeerID(pid)

	// Clear protocols on connecting to new peer to avoid issues caused
	// by misremembering protocols between reconnects.
	if err := h.Peerstore().SetProtocols(pid); err != nil {
		span.SetUnknownError(err)
	}

	if h.ids != nil {
		h.ids.IdentifyConn(conn)
	}
}

// newStreamHandler is the remote-opened stream handler for inet.Network.
// TODO: this feels a bit wonky
func (h *Host) newStreamHandler(stream inet.Stream) {
	ctx, span := monitoring.StartSpan(context.Background(), "p2p", "newStreamHandler")
	defer span.End()

	span.SetPeerID(stream.Conn().RemotePeer())

	if h.negTimeout > 0 {
		err := stream.SetDeadline(time.Now().Add(h.negTimeout))
		if err != nil {
			span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeFailedPrecondition, err.Error()))

			if err := stream.Reset(); err != nil {
				span.Annotate(ctx, "streamResetError", err.Error())
			}

			return
		}
	}

	lzc, protoID, handle, err := h.Mux().NegotiateLazy(stream)
	if err != nil {
		if err != io.EOF {
			span.SetUnknownError(err)
		}

		if err := stream.Reset(); err != nil {
			span.Annotate(ctx, "streamResetError", err.Error())
		}

		return
	}

	stream = &streamWrapper{
		Stream: stream,
		rw:     lzc,
	}

	if h.negTimeout > 0 {
		if err := stream.SetDeadline(time.Time{}); err != nil {
			span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeFailedPrecondition, err.Error()))

			if err := stream.Reset(); err != nil {
				span.Annotate(ctx, "streamResetError", err.Error())
			}

			return
		}
	}

	stream.SetProtocol(protocol.ID(protoID))
	span.SetProtocolID(protocol.ID(protoID))

	if h.bwc != nil {
		stream = mstream.WrapStream(stream, h.bwc)
	}

	// Assumes handle lifecyle is already properly handled.
	go func() {
		err := handle(protoID, stream)
		if err != nil && errors.Cause(err) != context.Canceled {
			span.Annotate(ctx, "handleFailed", err.Error())
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

// Addrs returns the filtered addresses of this host.
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
	ctx, span := monitoring.StartSpan(context.Background(), "p2p", "AllAddrs")
	defer span.End()

	addrs, err := h.netw.InterfaceListenAddresses()
	if err != nil {
		span.Annotate(ctx, "addrsError", err.Error())
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
	ctx, span := monitoring.StartSpan(ctx, "p2p", "Connect")
	defer span.End()

	span.SetPeerID(pi.ID)
	span.SetAddrs(pi.Addrs)

	ps := h.Peerstore()

	// Check if already connected.
	conns := h.Network().ConnsToPeer(pi.ID)
	if len(conns) > 0 {
		span.AddIntAttribute("connections", int64(len(conns)))
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
			span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeFailedPrecondition, err.Error()))
			return err
		}

		// Absorb addresses into peerstore.
		ps.AddAddrs(pi.ID, addrs, pstore.TempAddrTTL)
	}

	pi.Addrs = addrs

	resolved, err := h.resolveAddrs(ctx, ps.PeerInfo(pi.ID))
	if err != nil {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeFailedPrecondition, err.Error()))
		return err
	}

	ps.AddAddrs(pi.ID, resolved, pstore.TempAddrTTL)

	err = h.dialPeer(ctx, pi.ID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	return nil
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
	ctx, span := monitoring.StartSpan(ctx, "p2p", "resolveAddrs")
	defer span.End()

	span.SetPeerID(pi.ID)
	span.SetAddrs(pi.Addrs)

	// Create the IPFS P2P address of the peer.
	protocol := ma.ProtocolWithCode(ma.P_IPFS).Name
	p2pAddr, err := ma.NewMultiaddr("/" + protocol + "/" + pi.ID.Pretty())
	if err != nil {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeFailedPrecondition, err.Error()))
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
			span.Annotate(ctx, "resolveError", err.Error())
			continue
		}

		for _, res := range resAddrs {
			pi, err := pstore.InfoFromP2pAddr(res)
			if err != nil {
				span.Annotate(ctx, "p2pAddrError", err.Error())
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
	ctx, span := monitoring.StartSpan(ctx, "p2p", "dialPeer")
	defer span.End()

	span.SetPeerID(pid)

	conn, err := h.Network().DialPeer(ctx, pid)
	if err != nil {
		span.SetUnknownError(err)
		return errors.WithStack(err)
	}

	// Clear protocols on connecting to new peer to avoid issues caused
	// by misremembering protocols between reconnects.
	if err := h.Peerstore().SetProtocols(pid); err != nil {
		span.SetUnknownError(err)
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
			err := ctx.Err()
			if err != nil {
				span.SetUnknownError(err)
			}
			return errors.WithStack(err)
		}
	}

	return nil
}

// SetStreamHandler sets the stream handler for the given protocol.
func (h *Host) SetStreamHandler(proto protocol.ID, handler inet.StreamHandler) {
	_, span := monitoring.StartSpan(context.Background(), "p2p", "SetStreamHandler")
	span.SetProtocolID(proto)
	defer span.End()

	h.mux.AddHandler(string(proto), func(protoStr string, rwc io.ReadWriteCloser) error {
		stream := rwc.(inet.Stream)
		stream.SetProtocol(protocol.ID(protoStr))

		ctx := monitoring.NewTaggedContext(context.Background()).
			Tag(monitoring.PeerIDTag, stream.Conn().RemotePeer().Pretty()).
			Tag(monitoring.ProtocolIDTag, protoStr).
			Build()
		streamsIn.Record(ctx, 1)

		handler(stream)

		return nil
	})
}

// SetStreamHandlerMatch sets the protocol handler for protocols that match the
// given function.
func (h *Host) SetStreamHandlerMatch(proto protocol.ID, match func(string) bool, handler inet.StreamHandler) {
	_, span := monitoring.StartSpan(context.Background(), "p2p", "SetStreamHandlerMatch")
	span.SetProtocolID(proto)
	defer span.End()

	h.mux.AddHandlerWithFunc(string(proto), match, func(protoStr string, rwc io.ReadWriteCloser) error {
		stream := rwc.(inet.Stream)
		stream.SetProtocol(protocol.ID(protoStr))
		handler(stream)

		return nil
	})
}

// RemoveStreamHandler removes the stream handler of the given protocol.
func (h *Host) RemoveStreamHandler(proto protocol.ID) {
	_, span := monitoring.StartSpan(context.Background(), "p2p", "RemoveStreamHandler")
	span.SetProtocolID(proto)
	defer span.End()

	h.mux.RemoveHandler(string(proto))
}

// NewStream opens a new stream to the given peer for the given protocols.
func (h *Host) NewStream(ctx context.Context, pid peer.ID, protocols ...protocol.ID) (s inet.Stream, err error) {
	ctx, span := monitoring.StartSpan(ctx, "p2p", "NewStream", monitoring.SpanOptionPeerID(pid))
	defer func() {
		if err != nil {
			ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ErrorTag, err.Error()).Build()
			streamsErr.Record(ctx, 1)
			span.SetUnknownError(err)
		} else {
			streamsOut.Record(ctx, 1)
		}

		span.End()
	}()

	if h.router != nil {
		err := h.Connect(ctx, pstore.PeerInfo{ID: pid})
		if err != nil {
			return nil, err
		}
	}

	// Try to find a supported protocol.
	pref, err := h.preferredProtocol(pid, protocols)
	if err != nil {
		return nil, err
	}

	if pref != "" {
		ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ProtocolIDTag, string(pref)).Build()
		return h.newStream(ctx, pid, pref)
	}

	var protoStrs []string
	for _, pid := range protocols {
		protoStrs = append(protoStrs, string(pid))
	}

	stream, err := h.Network().NewStream(ctx, pid)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	selected, err := msmux.SelectOneOf(protoStrs, stream)
	if err != nil {
		if err := stream.Reset(); err != nil {
			span.Annotate(ctx, "streamResetError", err.Error())
		}

		return nil, errors.WithStack(err)
	}

	selfpid := protocol.ID(selected)
	stream.SetProtocol(selfpid)
	span.SetProtocolID(selfpid)

	ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ProtocolIDTag, selected).Build()

	if err := h.Peerstore().AddProtocols(pid, selected); err != nil {
		if err := stream.Reset(); err != nil {
			span.Annotate(ctx, "streamResetError", err.Error())
		}

		return nil, errors.WithStack(err)
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
	ctx, span := monitoring.StartSpan(ctx, "p2p", "newStream",
		monitoring.SpanOptionPeerID(pid),
		monitoring.SpanOptionProtocolID(proto),
	)
	defer span.End()

	stream, err := h.Network().NewStream(ctx, pid)
	if err != nil {
		span.SetUnknownError(err)
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

	h.tick.Stop()

	return errors.WithStack(h.netw.Close())
}

// ConnManager returns the connection manager.
func (h *Host) ConnManager() ifconnmgr.ConnManager {
	return h.cmgr
}

// SetNATManager sets the NAT manager.
func (h *Host) SetNATManager(natmgr bhost.NATManager) {
	h.natmgr = natmgr

	eventName := "SetNATManager"
	if natmgr == nil {
		eventName = "RemoveNATManager"
	}

	_, span := monitoring.StartSpan(h.ctx, "p2p", eventName)
	span.End()
}

// SetIDService sets the identity service.
func (h *Host) SetIDService(service *identify.IDService) {
	h.ids = service

	eventName := "SetIDService"
	if service == nil {
		eventName = "RemoveIDService"
	}

	_, span := monitoring.StartSpan(h.ctx, "p2p", eventName)
	span.End()
}

// SetRouter sets the router.
func (h *Host) SetRouter(router func(context.Context, peer.ID) (pstore.PeerInfo, error)) {
	h.router = router

	eventName := "SetRouter"
	if router == nil {
		eventName = "RemoveRouter"
	}

	_, span := monitoring.StartSpan(h.ctx, "p2p", eventName)
	span.End()
}

// collectMetrics periodically reports p2p metrics.
func (h *Host) collectMetrics() {
	for range h.tick.C {
		ctx := context.Background()
		connCount := len(h.Network().Conns())
		peerCount := len(h.Network().Peers())

		connections.Record(ctx, int64(connCount))
		peers.Record(ctx, int64(peerCount))

		for _, peerID := range h.Peerstore().Peers() {
			if peerID == h.ID() {
				continue
			}

			ctx := monitoring.NewTaggedContext(ctx).Tag(monitoring.PeerIDTag, peerID.Pretty()).Build()
			peerLatency := ((float64)(h.Peerstore().LatencyEWMA(peerID).Nanoseconds())) / 1e6
			latency.Record(ctx, peerLatency)
		}
	}
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
