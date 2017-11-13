# Alice

Alice is virtual infrastructure for interoperable P2P services.

[![Build Status](https://travis-ci.com/stratumn/alice.svg?token=En6rzNysH6Mz2pqepQLy&branch=master)](https://travis-ci.com/stratumn/alice)
[![codecov](https://codecov.io/gh/stratumn/alice/branch/master/graph/badge.svg?token=nVHWHcr5xQ)](https://codecov.io/gh/stratumn/alice)

Copyright © 2017 Stratumn SAS

## Project Status

The current focus is to build a solid architecture to develop P2P services.
Once the design has proven its soundness, some code needs to be cleaned up and
unit tests need to be written (things change way too fast at this point).

### Current features

* uses IPFS' go-libp2p library
* neat and powerful inner-process service based architecture
* core services (P2P, NAT, DHT routing, relay, etc...)
* P2P bootstrapping from seed nodes
* gRPC API
* CLI with gRPC command reflection
* integration test framework
* nice logs

### Next

* ability to create and join multiple private or public P2P services
* a simple, script-less, proof-of-work digital asset
* Indigo integration

### Stuff that needs refactoring

* `cli/reflect.go`: current code feels hacky, it should be properly designed
  to make it easy to support new types
* `core/manager/manager.go`: current implementation isn't bad but a lot of
  functions have high cyclomatic complexity


## Installation

Install Go. On macOS, you can install using `homebrew`:

```bash
$ brew install go
```

Install Go dependencies:

```bash
$ make deps
```

Compile and install `alice`:

```bash
$ go install
```

## Usage

Create a new directory. Open a terminal in that directory then create
configuration files using `alice init`:

```bash
$ alice init
```

Now you can launch a node (from the same directory):

```bash
$ alice up
```

Open another terminal and connect to the node (from the same directory):

```bash
$ alice cli
```

To view streaming logs (from the same directory):

```bash
$ alice log -f log
```

For more information, read the [documentation](doc/README.md).

