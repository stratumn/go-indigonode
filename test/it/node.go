// Copyright © 2017 Stratumn SAS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package it

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core"
	"github.com/stratumn/alice/core/cfg"
	"github.com/stratumn/alice/core/service/bootstrap"
	"github.com/stratumn/alice/core/service/grpcapi"
	"github.com/stratumn/alice/core/service/kaddht"
	"github.com/stratumn/alice/core/service/swarm"
	grpcapipb "github.com/stratumn/alice/grpc/grpcapi"
	managerpb "github.com/stratumn/alice/grpc/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	manet "gx/ipfs/QmX3U3YXCQ6UYBxq2LVWF8dARS1hPUTEYLrSx654Qyxyw6/go-multiaddr-net"
	ma "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

// port is atomically incremented each time a port is assigned.
var port = int32(4000)

// randPort returns a random port.
func randPort() int32 {
	for {
		p := atomic.AddInt32(&port, 1)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			if err := l.Close(); err != nil {
				continue
			}
			return p
		}
	}
}

// TestNode represents a test node.
type TestNode struct {
	dir       string
	pid       peer.ID
	swarmAddr string
	apiAddr   string
	conf      cfg.ConfigSet
}

// NewTestNode creates a new test node.
func NewTestNode(dir string, config cfg.ConfigSet) (*TestNode, error) {
	var err error

	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	privKey, _, err := crypto.GenerateKeyPair(crypto.RSA, 1024)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	b, err := privKey.Bytes()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pid, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	swarmAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", atomic.AddInt32(&port, 1))
	apiAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", atomic.AddInt32(&port, 1))

	// Don't mutate the given config.
	config = deepcopy.Copy(config).(cfg.ConfigSet)

	swarmConf := config["swarm"].(swarm.Config)
	swarmConf.PrivateKey = crypto.ConfigEncodeKey(b)
	swarmConf.PeerID = pid.Pretty()
	swarmConf.Addresses = []string{swarmAddr}
	config["swarm"] = swarmConf

	apiConf := config["grpcapi"].(grpcapi.Config)
	apiConf.Address = apiAddr
	config["grpcapi"] = apiConf

	dhtConf := config["kaddht"].(kaddht.Config)
	dhtConf.LevelDBPath = filepath.Join(dir, "data", "kaddht")
	config["kaddht"] = dhtConf

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, errors.WithStack(err)
	}

	confFile := filepath.Join(dir, "alice.core.toml")
	if err := config.Save(confFile, 0600, false); err != nil {
		return nil, err
	}

	return &TestNode{
		dir,
		pid,
		swarmAddr,
		apiAddr,
		config,
	}, nil
}

// FullAddress returns the full address to connect to the node.
func (h *TestNode) FullAddress() string {
	addr := strings.Replace(h.swarmAddr, "0.0.0.0", "127.0.0.1", 1)
	return fmt.Sprintf("%s/ipfs/%s", addr, h.pid.Pretty())
}

// Up launches the node. It blocks until the process exits or the context is
// done.
func (h *TestNode) Up(ctx context.Context) error {
	logf, err := os.OpenFile(filepath.Join(h.dir, "logs"), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.WithStack(err)
	}
	defer logf.Close()
	defer logf.Sync()

	cmd := exec.Command("alice", "up")
	cmd.Dir = h.dir
	cmd.Stdout = logf
	cmd.Stderr = logf
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Detach from current process.
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- errors.WithStack(cmd.Wait())
	}()

	select {
	case err = <-done:
	case <-ctx.Done():
		cmd.Process.Signal(os.Interrupt)
		err = <-done
	}

	return err
}

// Connect returns an API client connected to the node.
func (h *TestNode) Connect(ctx context.Context) (*grpc.ClientConn, error) {
	conn, err := h.WaitForAPI(ctx)
	if err != nil {
		return nil, err
	}

	client := managerpb.NewManagerClient(conn)

	// Wait for boot service to be running.
	boot := h.conf["core"].(core.Config).BootService

	// List services until ready.
CONNECT_LOOP:
	for {
		streamCtx, cancelStream := context.WithCancel(ctx)
		defer cancelStream()

		ss, err := client.List(streamCtx, &managerpb.ListReq{})
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for {
			msg, err := ss.Recv()
			if err == io.EOF {
				return nil, errors.Errorf("boot services not found: %s", boot)
			}
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if msg.Id == boot {
				if msg.Status == managerpb.Service_RUNNING {
					return conn, nil
				}

				time.Sleep(500 * time.Millisecond)

				cancelStream()
				continue CONNECT_LOOP
			}
		}
	}
}

// WaitForAPI waits for the API to be available.
func (h *TestNode) WaitForAPI(ctx context.Context) (*grpc.ClientConn, error) {
	conn, err := h.conn(ctx)
	if err != nil {
		return nil, err
	}

	client := grpcapipb.NewAPIClient(conn)

	for {
		streamCtx, cancelStream := context.WithCancel(ctx)
		defer cancelStream()

		_, err := client.Inform(streamCtx, &grpcapipb.InformReq{})
		if err == nil {
			return conn, nil
		}

		conn.Close()

		if grpc.Code(err) != codes.Unavailable {
			return nil, errors.WithStack(err)
		}

		cancelStream()
		time.Sleep(500 * time.Millisecond)
	}
}

// conn creates a connection to the API.
func (h *TestNode) conn(ctx context.Context) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(
		ctx,
		h.apiAddr,
		// Use multiaddr dialier.
		grpc.WithDialer(func(addr string, t time.Duration) (net.Conn, error) {
			a, err := ma.NewMultiaddr(addr)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			dialer := manet.Dialer{}
			conn, err := dialer.DialContext(ctx, a)
			return conn, errors.WithStack(err)
		}),
		// Makes the call block till the connection is made.
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)

	return conn, errors.WithStack(err)
}

// PeerID returns the peer ID of the node.
func (h *TestNode) PeerID() peer.ID {
	return h.pid
}

// TestNodeSet represents a set of test nodes.
type TestNodeSet []*TestNode

// NewTestNodeSet creates a new set of test nodes.
func NewTestNodeSet(dir string, n int, config cfg.ConfigSet) (TestNodeSet, error) {
	var err error

	if dir, err = filepath.Abs(dir); err != nil {
		return nil, errors.WithStack(err)
	}

	nodes, errs := make(TestNodeSet, n), make([]error, n)

	PerCPU(func(i int) {
		name := fmt.Sprintf("node%d", i)
		dir := filepath.Join(dir, name)
		node, err := NewTestNode(dir, config)
		nodes[i], errs[i] = node, errors.Wrap(err, name)
	}, n)

	if err = NewMultiErr(errs); err != nil {
		return nil, err
	}

	nodes.randSeeds()

	for _, node := range nodes {
		confFile := filepath.Join(node.dir, "alice.core.toml")
		if err := node.conf.Save(confFile, 0600, true); err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

// Up launches the set of nodes.
func (s TestNodeSet) Up(ctx context.Context) error {
	errs := make([]error, len(s))
	wg := sync.WaitGroup{}
	wg.Add(len(s))

	PerCPU(func(i int) {
		go func(i int) {
			errs[i] = errors.Wrapf(s[i].Up(ctx), "node %d", i)
			wg.Done()
		}(i)

		conn, err := s[i].WaitForAPI(ctx)
		if err == nil {
			conn.Close()
		}
	}, len(s))

	wg.Wait()

	return NewMultiErr(errs)
}

// randSeeds creates random seeds.
func (s TestNodeSet) randSeeds() {
	seeds := make([]string, len(s))
	for i, h := range s {
		seeds[i] = h.FullAddress()
	}

	// Exclude the first node so that tests can use it as a node that is
	// not part of the bootstrap list. We select random one but make sure
	// that all the nodes are connected by taking a sample of 2*NumSeed-1.
	// This guarantees that at least one node will be connected to every
	// node.
	seeds = randPeers(seeds[1:], 2*NumSeeds-1)
	if NumSeeds < len(seeds) {
		seeds = seeds[:NumSeeds]
	}

	for _, h := range s {
		bsConf := h.conf["bootstrap"].(bootstrap.Config)
		bsConf.Addresses = seeds
		h.conf["bootstrap"] = bsConf
	}
}

// Connect returns an API client for each node in the set.
func (s TestNodeSet) Connect(ctx context.Context) ([]*grpc.ClientConn, error) {
	conns, errs := make([]*grpc.ClientConn, len(s)), make([]error, len(s))

	PerCPU(func(i int) {
		conn, err := s[i].Connect(ctx)
		conns[i], errs[i] = conn, errors.Wrapf(err, "node %d", i)
	}, len(s))

	return conns, NewMultiErr(errs)
}
