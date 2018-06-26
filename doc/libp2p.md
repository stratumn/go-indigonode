# Libp2p Overview

## Network Packages

An quick summary of the libp2p network packages.

### go-multistream (msmux)

Defines a stream router that routes streams to a handler based on their
protocol string.

### go-stream-muxer (smux)

Common interfaces for stream muxers:

- Stream: can be read from and written to and provides access to a Conn
- Conn: can open and accept multiple streams between two peers
- Transport: can create Conn from a standard net.Conn

### go-smux-yamux (yamux)

Implements a smux.Transport for the Yamux stream multiplexer.

### go-smux-multistream (mssmux)

Implements a smux.Transport which selects an underlying smux.Transport based
on the protocol. It uses go-multistream to select the appropriate transport.
For instance you can add the Yamux transport for the "/yamux/v1.0.0" protocol
string.

### go-libp2p-swarm (swarm)

Holds connections to peers. It enables handling incoming streams and opening a
stream with a peer.

### go-libp2p-host

Defines the Host interface. Implementations expose functions such as
SetStreamHandler, which adds a handler for a protocol. This is typically done
by encapsulating a msmux and a swarm.

### go-libp2p-kad-dht

A Kademlia distributed hash table and protocol. It is typically used to store
addresses of node IDs so that nodes can be found.

### go-floodsub

A decentralized Publish-Subscribe system.
