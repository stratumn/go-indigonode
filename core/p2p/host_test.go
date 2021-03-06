/*
This code is based on:

	github.com/libp2p/go-libp2p/p2p/host/basic

Which has the license:

	The MIT License (MIT)

And has the copyright:

	Copyright (c) 2014 Juan Batiz-Benet
*/

package p2p

import (
	"bytes"
	"context"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	madns "github.com/multiformats/go-multiaddr-dns"

	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	swarmtesting "github.com/libp2p/go-libp2p-swarm/testing"
	identify "github.com/libp2p/go-libp2p/p2p/protocol/identify"
	testutil "github.com/libp2p/go-testutil"
	ma "github.com/multiformats/go-multiaddr"
)

func TestHostSimple(t *testing.T) {
	ctx := context.Background()
	h1 := NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h2 := NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h1.SetIDService(identify.NewIDService(h1))
	h2.SetIDService(identify.NewIDService(h2))
	defer h1.Close()
	defer h2.Close()

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	err := h1.Connect(ctx, h2pi)
	require.NoError(t, err)

	piper, pipew := io.Pipe()
	h2.SetStreamHandler(protocol.TestingID, func(s inet.Stream) {
		defer s.Close()
		w := io.MultiWriter(s, pipew)
		io.Copy(w, s) // mirror everything
	})

	s, err := h1.NewStream(ctx, h2pi.ID, protocol.TestingID)
	require.NoError(t, err)

	// write to the stream
	buf1 := []byte("abcdefghijkl")
	_, err = s.Write(buf1)
	require.NoError(t, err)

	// get it from the stream (echoed)
	buf2 := make([]byte, len(buf1))
	_, err = io.ReadFull(s, buf2)
	require.NoError(t, err)
	require.True(t, bytes.Equal(buf1, buf2), "buf1 != buf2")

	// get it from the pipe (tee)
	buf3 := make([]byte, len(buf1))
	_, err = io.ReadFull(piper, buf3)
	require.NoError(t, err)
	require.True(t, bytes.Equal(buf1, buf3), "buf1 != buf3")
}

func getHostPair(ctx context.Context, t *testing.T) (host.Host, host.Host) {
	h1 := NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h2 := NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h1.SetIDService(identify.NewIDService(h1))
	h2.SetIDService(identify.NewIDService(h2))

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	err := h1.Connect(ctx, h2pi)
	require.NoError(t, err)

	return h1, h2
}

func assertWait(t *testing.T, c chan protocol.ID, exp protocol.ID) {
	select {
	case proto := <-c:
		require.Equal(t, exp, proto, "proto")
	case <-time.After(time.Second * 10):
		require.Fail(t, "timeout waiting for stream")
	}
}

func TestHostProtoPreference(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, h2 := getHostPair(ctx, t)
	defer h1.Close()
	defer h2.Close()

	protoOld := protocol.ID("/testing")
	protoNew := protocol.ID("/testing/1.1.0")
	protoMinor := protocol.ID("/testing/1.2.0")

	connectedOn := make(chan protocol.ID)

	handler := func(s inet.Stream) {
		connectedOn <- s.Protocol()
		inet.FullClose(s)
	}

	h1.SetStreamHandler(protoOld, handler)

	s, err := h2.NewStream(ctx, h1.ID(), protoMinor, protoNew, protoOld)
	require.NoError(t, err)

	assertWait(t, connectedOn, protoOld)
	inet.FullClose(s)

	mfunc, err := host.MultistreamSemverMatcher(protoMinor)
	require.NoError(t, err)

	h1.SetStreamHandlerMatch(protoMinor, mfunc, handler)

	// remembered preference will be chosen first, even when the other side newly supports it
	s2, err := h2.NewStream(ctx, h1.ID(), protoMinor, protoNew, protoOld)
	require.NoError(t, err)

	// required to force 'lazy' handshake
	_, err = s2.Write([]byte("hello"))
	require.NoError(t, err)

	assertWait(t, connectedOn, protoOld)

	inet.FullClose(s2)

	s3, err := h2.NewStream(ctx, h1.ID(), protoMinor)
	require.NoError(t, err)

	assertWait(t, connectedOn, protoMinor)
	inet.FullClose(s3)
}

func TestHostProtoMismatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, h2 := getHostPair(ctx, t)
	defer h1.Close()
	defer h2.Close()

	h1.SetStreamHandler("/super", func(s inet.Stream) {
		assert.Fail(t, "shouldn't get here")
		s.Reset()
	})

	_, err := h2.NewStream(ctx, h1.ID(), "/foo", "/bar", "/baz/1.0.0")
	require.Error(t, err, "expected new stream to fail")
}

func TestHostProtoPreknowledge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1 := NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h2 := NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h1.SetIDService(identify.NewIDService(h1))
	h2.SetIDService(identify.NewIDService(h2))

	conn := make(chan protocol.ID)
	handler := func(s inet.Stream) {
		conn <- s.Protocol()
		inet.FullClose(s)
	}

	h1.SetStreamHandler("/super", handler)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	err := h1.Connect(ctx, h2pi)
	require.NoError(t, err)

	defer h1.Close()
	defer h2.Close()

	// wait for identify handshake to finish completely
	select {
	case <-h1.ids.IdentifyWait(h1.Network().ConnsToPeer(h2.ID())[0]):
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for identify")
	}

	select {
	case <-h2.ids.IdentifyWait(h2.Network().ConnsToPeer(h1.ID())[0]):
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for identify")
	}

	h1.SetStreamHandler("/foo", handler)

	s, err := h2.NewStream(ctx, h1.ID(), "/foo", "/bar", "/super")
	require.NoError(t, err)

	select {
	case p := <-conn:
		require.Fail(t, "shouldn't have gotten connection yet, we should have a lazy stream: ", p)
	case <-time.After(time.Millisecond * 50):
	}

	_, err = s.Read(nil)
	require.NoError(t, err)

	assertWait(t, conn, "/super")

	inet.FullClose(s)
}

func TestNewDialOld(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, h2 := getHostPair(ctx, t)
	defer h1.Close()
	defer h2.Close()

	connectedOn := make(chan protocol.ID)
	h1.SetStreamHandler("/testing", func(s inet.Stream) {
		connectedOn <- s.Protocol()
		inet.FullClose(s)
	})

	s, err := h2.NewStream(ctx, h1.ID(), "/testing/1.0.0", "/testing")
	require.NoError(t, err)

	assertWait(t, connectedOn, "/testing")

	require.EqualValues(t, "/testing", s.Protocol(), "s.Protocol()")

	inet.FullClose(s)
}

func TestProtoDowngrade(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, h2 := getHostPair(ctx, t)
	defer h1.Close()
	defer h2.Close()

	connectedOn := make(chan protocol.ID)
	h1.SetStreamHandler("/testing/1.0.0", func(s inet.Stream) {
		connectedOn <- s.Protocol()
		inet.FullClose(s)
	})

	s, err := h2.NewStream(ctx, h1.ID(), "/testing/1.0.0", "/testing")
	require.NoError(t, err)

	assertWait(t, connectedOn, "/testing/1.0.0")

	require.EqualValues(t, "/testing/1.0.0", s.Protocol(), "s.Protocol()")
	inet.FullClose(s)

	h1.Network().ConnsToPeer(h2.ID())[0].Close()

	time.Sleep(time.Millisecond * 100) // allow notifications to propagate
	h1.RemoveStreamHandler("/testing/1.0.0")
	h1.SetStreamHandler("/testing", func(s inet.Stream) {
		connectedOn <- s.Protocol()
		inet.FullClose(s)
	})

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	err = h1.Connect(ctx, h2pi)
	require.NoError(t, err)

	s2, err := h2.NewStream(ctx, h1.ID(), "/testing/1.0.0", "/testing")
	require.NoError(t, err)

	_, err = s2.Write(nil)
	require.NoError(t, err)

	assertWait(t, connectedOn, "/testing")

	require.EqualValues(t, "/testing", s2.Protocol(), "s2.Protocol()")
	inet.FullClose(s2)
}

func TestAddrResolution(t *testing.T) {
	ctx := context.Background()

	p1, err := testutil.RandPeerID()
	assert.NoError(t, err)

	p2, err := testutil.RandPeerID()
	assert.NoError(t, err)

	addr1 := ma.StringCast("/dnsaddr/example.com")
	addr2 := ma.StringCast("/ip4/192.0.2.1/tcp/123")
	p2paddr1 := ma.StringCast("/dnsaddr/example.com/ipfs/" + p1.Pretty())
	p2paddr2 := ma.StringCast("/ip4/192.0.2.1/tcp/123/ipfs/" + p1.Pretty())
	p2paddr3 := ma.StringCast("/ip4/192.0.2.1/tcp/123/ipfs/" + p2.Pretty())

	backend := &madns.MockBackend{
		TXT: map[string][]string{"_dnsaddr.example.com": []string{
			"dnsaddr=" + p2paddr2.String(), "dnsaddr=" + p2paddr3.String(),
		}},
	}
	resolver := &madns.Resolver{Backend: backend}

	h := NewHost(ctx, swarmtesting.GenSwarm(t, ctx), OptResolver(resolver))
	h.SetIDService(identify.NewIDService(h))

	defer h.Close()

	pi, err := pstore.InfoFromP2pAddr(p2paddr1)
	assert.NoError(t, err)

	tctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()
	_ = h.Connect(tctx, *pi)

	addrs := h.Peerstore().Addrs(pi.ID)
	sort.Sort(sortedMultiaddrs(addrs))

	require.Equal(t, 2, len(addrs), "len(addrs)")
	require.True(t, addrs[0].Equal(addr1), "addrs[0]")
	require.True(t, addrs[1].Equal(addr2), "addrs[1]")
}

type sortedMultiaddrs []ma.Multiaddr

func (sma sortedMultiaddrs) Len() int      { return len(sma) }
func (sma sortedMultiaddrs) Swap(i, j int) { sma[i], sma[j] = sma[j], sma[i] }
func (sma sortedMultiaddrs) Less(i, j int) bool {
	return bytes.Compare(sma[i].Bytes(), sma[j].Bytes()) == 1
}
