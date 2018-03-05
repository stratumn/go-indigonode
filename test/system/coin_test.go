// Copyright © 2017-2018 Stratumn SAS
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

package system

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"google.golang.org/grpc"

	"github.com/mohae/deepcopy"
	"github.com/stratumn/alice/core/cfg"
	"github.com/stratumn/alice/core/service/coin"
	coinpb "github.com/stratumn/alice/grpc/coin"
	mngrpb "github.com/stratumn/alice/grpc/manager"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stratumn/alice/test/session"
	"github.com/stretchr/testify/assert"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

func TestCoin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// SETUP.

	privKey0, pubKey0, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	assert.NoError(t, err, "GenerateKeyPair()")
	minerID0, err := peer.IDFromPublicKey(pubKey0)
	assert.NoError(t, err, "IDFromPublicKey()")
	balance0 := uint64(42)

	_, pubKey1, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	assert.NoError(t, err, "GenerateKeyPair()")
	minerID1, err := peer.IDFromPublicKey(pubKey1)
	assert.NoError(t, err, "IDFromPublicKey()")
	balance1 := uint64(0)

	_, pubKey2, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	assert.NoError(t, err, "GenerateKeyPair()")
	minerID2, err := peer.IDFromPublicKey(pubKey2)
	assert.NoError(t, err, "IDFromPublicKey()")
	balance2 := uint64(0)

	_, pubKey3, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	assert.NoError(t, err, "GenerateKeyPair()")
	minerID3, err := peer.IDFromPublicKey(pubKey3)
	assert.NoError(t, err, "IDFromPublicKey()")
	balance3 := uint64(43)

	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	assert.NoError(t, err, "GenerateKeyPair()")
	someGuyID, err := peer.IDFromPublicKey(pubKey)
	assert.NoError(t, err, "IDFromPublicKey()")
	balanceGuy := uint64(44)

	txs := []*pb.Transaction{
		&pb.Transaction{
			To:    []byte(minerID0),
			Value: balance0,
		},
		&pb.Transaction{
			To:    []byte(minerID3),
			Value: balance3,
		},
		&pb.Transaction{
			To:    []byte(someGuyID),
			Value: balanceGuy,
		},
	}

	txsMerkleRoot, err := coinutil.TransactionRoot(txs)
	assert.NoError(t, err, "TransactionRoot()")
	ts, err := ptypes.TimestampProto(time.Now())
	assert.NoError(t, err)

	genesis := &pb.Block{
		Header: &pb.Header{
			Nonce:      42,
			Version:    2,
			MerkleRoot: txsMerkleRoot,
			Timestamp:  ts,
		},
		Transactions: txs,
	}

	genBytes, err := genesis.Marshal()
	assert.NoError(t, err, "genesis.Marshal()")
	genesisHash := hex.EncodeToString(genBytes)

	config := session.WithServices(session.SystemCfg(), "boot", "coin")

	configs := []cfg.ConfigSet{
		withCoinConfig(config, minerID0, genesisHash),
		withCoinConfig(config, minerID1, genesisHash),
		withCoinConfig(config, minerID2, genesisHash),
		withCoinConfig(config, minerID3, ""),
	}

	// TEST.

	testFn := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {

		clients := make([]coinpb.CoinClient, 4)
		for i := range conns {
			clients[i] = coinpb.NewCoinClient(conns[i])
		}

		// Check initialization.
		// node3 has not been initialized with the same genesis block.
		assertAccount(ctx, t, clients[0], minerID0, balance0, uint64(0))
		assertAccount(ctx, t, clients[0], minerID1, balance1, uint64(0))
		assertAccount(ctx, t, clients[0], minerID2, balance2, uint64(0))
		assertAccount(ctx, t, clients[0], minerID3, balance3, uint64(0))
		assertAccount(ctx, t, clients[0], someGuyID, balanceGuy, uint64(0))

		assertAccount(ctx, t, clients[1], minerID0, balance0, uint64(0))
		assertAccount(ctx, t, clients[1], minerID1, balance1, uint64(0))
		assertAccount(ctx, t, clients[1], minerID2, balance2, uint64(0))
		assertAccount(ctx, t, clients[1], minerID3, balance3, uint64(0))
		assertAccount(ctx, t, clients[1], someGuyID, balanceGuy, uint64(0))

		assertAccount(ctx, t, clients[2], minerID0, balance0, uint64(0))
		assertAccount(ctx, t, clients[2], minerID1, balance1, uint64(0))
		assertAccount(ctx, t, clients[2], minerID2, balance2, uint64(0))
		assertAccount(ctx, t, clients[2], minerID3, balance3, uint64(0))
		assertAccount(ctx, t, clients[2], someGuyID, balanceGuy, uint64(0))

		assertAccount(ctx, t, clients[3], minerID0, uint64(0), uint64(0))
		assertAccount(ctx, t, clients[3], minerID1, uint64(0), uint64(0))
		assertAccount(ctx, t, clients[3], minerID2, uint64(0), uint64(0))
		assertAccount(ctx, t, clients[3], minerID3, uint64(0), uint64(0))
		assertAccount(ctx, t, clients[3], someGuyID, uint64(0), uint64(0))

		// Add a transaction.
		tx := &pb.Transaction{
			From:  []byte(someGuyID),
			To:    []byte(minerID0),
			Nonce: 1,
			Value: 12,
			Fee:   2,
		}
		doTransaction(ctx, t, clients[0], privKey, tx)
		balanceGuy -= 12 + 2

		// Check that a new block has been accepted.
		assertHeight(ctx, t, clients[0], 1)
		assertHeight(ctx, t, clients[1], 1)

		assertAccount(ctx, t, clients[0], someGuyID, balanceGuy, uint64(1))
		assertAccount(ctx, t, clients[1], someGuyID, balanceGuy, uint64(1))
		assertAccount(ctx, t, clients[2], someGuyID, balanceGuy, uint64(1))
		assertAccount(ctx, t, clients[3], someGuyID, uint64(0), uint64(0))

		// Stop node2.
		mngr2 := mngrpb.NewManagerClient(conns[2])
		_, err := mngr2.Stop(ctx, &mngrpb.StopReq{Id: "test"})
		assert.NoError(t, err, "Stop(test)")
		_, err = mngr2.Stop(ctx, &mngrpb.StopReq{Id: "coin"})
		assert.NoError(t, err, "Stop(coin)")

		// Add bunch of other transactions.
		doTransaction(ctx, t, clients[0], privKey, &pb.Transaction{
			From:  []byte(someGuyID),
			To:    []byte(minerID1),
			Nonce: 2,
			Value: 2,
			Fee:   1,
		})
		balanceGuy -= 2 + 1
		assertHeight(ctx, t, clients[0], 2)
		assertHeight(ctx, t, clients[1], 2)

		doTransaction(ctx, t, clients[0], privKey0, &pb.Transaction{
			From:  []byte(minerID0),
			To:    []byte(minerID2),
			Nonce: 4,
			Value: 9,
			Fee:   2,
		})
		doTransaction(ctx, t, clients[0], privKey, &pb.Transaction{
			From:  []byte(someGuyID),
			To:    []byte(minerID0),
			Nonce: 9000,
			Value: 7,
			Fee:   1,
		})
		balanceGuy -= 7 + 1
		assertHeight(ctx, t, clients[0], 3)
		assertHeight(ctx, t, clients[1], 3)

		// Stop nodes 0 and 1.
		mngr0 := mngrpb.NewManagerClient(conns[0])
		_, err = mngr0.Stop(ctx, &mngrpb.StopReq{Id: "test"})
		assert.NoError(t, err, "Stop(test)")
		_, err = mngr0.Stop(ctx, &mngrpb.StopReq{Id: "coin"})
		assert.NoError(t, err, "Stop(coin)")

		mngr1 := mngrpb.NewManagerClient(conns[1])
		_, err = mngr1.Stop(ctx, &mngrpb.StopReq{Id: "test"})
		assert.NoError(t, err, "Stop(test)")
		_, err = mngr1.Stop(ctx, &mngrpb.StopReq{Id: "coin"})
		assert.NoError(t, err, "Stop(coin)")

		// Start node2.
		_, err = mngr2.Start(ctx, &mngrpb.StartReq{Id: "coin"})
		assert.NoError(t, err, "Start(coin)")
		_, err = mngr2.Start(ctx, &mngrpb.StartReq{Id: "test"})
		assert.NoError(t, err, "Start(test)")

		// Add one block to node2's chain.
		doTransaction(ctx, t, clients[2], privKey, &pb.Transaction{
			From:  []byte(someGuyID),
			To:    []byte(minerID0),
			Nonce: 2,
			Value: 28,
			Fee:   1,
		})
		assertHeight(ctx, t, clients[2], 2)
		assertAccount(ctx, t, clients[2], someGuyID, 1, 2)

		// Start all the nodes and check that the network converges toward
		// the longest chain (which is node0 and node1's chain).

		_, err = mngr0.Start(ctx, &mngrpb.StartReq{Id: "coin"})
		assert.NoError(t, err, "Start(coin)")
		_, err = mngr0.Start(ctx, &mngrpb.StartReq{Id: "test"})
		assert.NoError(t, err, "Start(test)")

		_, err = mngr1.Start(ctx, &mngrpb.StartReq{Id: "coin"})
		assert.NoError(t, err, "Start(coin)")
		_, err = mngr1.Start(ctx, &mngrpb.StartReq{Id: "test"})
		assert.NoError(t, err, "Start(test)")

		doTransaction(ctx, t, clients[0], privKey, &pb.Transaction{
			From:  []byte(someGuyID),
			To:    []byte(minerID3),
			Nonce: 9001,
			Value: 4,
			Fee:   3,
		})
		balanceGuy -= 4 + 3

		assertHeight(ctx, t, clients[0], 4)
		assertAccount(ctx, t, clients[0], someGuyID, balanceGuy, 9001)

		assertHeight(ctx, t, clients[1], 4)
		assertAccount(ctx, t, clients[1], someGuyID, balanceGuy, 9001)

		assertHeight(ctx, t, clients[2], 4)
		assertAccount(ctx, t, clients[2], someGuyID, balanceGuy, 9001)

	}

	err = session.RunWithConfigs(ctx, SessionDir, 4, configs, testFn)
	assert.NoError(t, err, "Session()")

}

// ####################################################################################################################
// 																										HELPERS
// ####################################################################################################################

func doTransaction(ctx context.Context, t *testing.T, c coinpb.CoinClient, priv crypto.PrivKey, tx *pb.Transaction) {
	txb, err := tx.Marshal()
	assert.NoError(t, err, "tx.Marshal()")

	sig, err := priv.Sign(txb)
	assert.NoError(t, err, "Sign()")

	pub := priv.GetPublic()
	pkb, err := pub.Bytes()
	assert.NoError(t, err, "pubKey.Bytes()")

	tx.Signature = &pb.Signature{
		KeyType:   pb.KeyType_Ed25519,
		PublicKey: pkb,
		Signature: sig,
	}

	_, err = c.Transaction(ctx, tx)
	assert.NoError(t, err, "client.Transaction()")
}

// assertHeight checks the height of the chain is num.
func assertHeight(ctx context.Context, t *testing.T, c coinpb.CoinClient, num uint64) {
	waitUntil(t, time.Second*2, time.Millisecond*250, func() bool {
		// We have the block of height num
		bcn, _ := c.Blockchain(ctx, &coinpb.BlockchainReq{BlockNumber: num})
		if bcn == nil || bcn.Blocks[0].BlockNumber() != num {
			return false
		}
		// We don't have the block of height num + 1.
		_, err := c.Blockchain(ctx, &coinpb.BlockchainReq{BlockNumber: num + 1})
		if err == nil {
			return false
		}
		return true
	}, "blockChain should update")
}

func assertAccount(ctx context.Context, t *testing.T, c coinpb.CoinClient, pid peer.ID, balance, nonce uint64) {
	acc, err := c.Account(ctx, &coinpb.AccountReq{PeerId: []byte(pid)})
	assert.NoError(t, err, "Account()")
	if acc != nil {
		assert.Equal(t, balance, acc.Balance, "account.Balance")
		assert.Equal(t, nonce, acc.Nonce, "account.Balance")
	}
}

func waitUntil(t *testing.T, duration time.Duration, interval time.Duration, cond func() bool, message string) {
	condChan := make(chan struct{})
	go func() {
		for {
			if cond() {
				condChan <- struct{}{}
				return
			}

			<-time.After(interval)
		}
	}()

	select {
	case <-condChan:
	case <-time.After(duration):
		assert.Fail(t, "waitUntil() condition failed:", message)
	}
}

func withCoinConfig(config cfg.ConfigSet, minerID peer.ID, genesis string) cfg.ConfigSet {
	conf := deepcopy.Copy(config).(cfg.ConfigSet)

	coinConf := conf["coin"].(coin.Config)
	coinConf.MinerID = peer.IDB58Encode(minerID)
	coinConf.GenesisBlock = genesis
	coinConf.BlockDifficulty = 12
	conf["coin"] = coinConf

	return conf
}
