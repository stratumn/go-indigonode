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
	"path/filepath"
	"testing"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/app/coin/protocol/coinutil"
	"github.com/stratumn/alice/test"
	"google.golang.org/grpc"

	"github.com/mohae/deepcopy"
	coinpb "github.com/stratumn/alice/app/coin/grpc"
	"github.com/stratumn/alice/app/coin/pb"
	coin "github.com/stratumn/alice/app/coin/service"
	"github.com/stratumn/alice/core/cfg"
	mngrpb "github.com/stratumn/alice/grpc/manager"
	"github.com/stratumn/alice/test/session"
	system "github.com/stratumn/alice/test/system"
	"github.com/stretchr/testify/assert"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

func TestCoin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// SETUP.

	_, pubKey0, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
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
		goodClients := clients[0:2]

		// Check initialization.
		// node3 has not been initialized with the same genesis block.
		assertAllAccount(ctx, t, goodClients, minerID0, balance0, uint64(0))
		assertAllAccount(ctx, t, goodClients, minerID1, balance1, uint64(0))
		assertAllAccount(ctx, t, goodClients, minerID2, balance2, uint64(0))
		assertAllAccount(ctx, t, goodClients, minerID3, balance3, uint64(0))
		assertAllAccount(ctx, t, goodClients, someGuyID, balanceGuy, uint64(0))

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
		assertAllHeight(ctx, t, goodClients, 1)

		assertAllAccount(ctx, t, goodClients, someGuyID, balanceGuy, uint64(1))
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

		assertAllHeight(ctx, t, goodClients, 4)
		assertAllAccount(ctx, t, goodClients, someGuyID, balanceGuy, 9001)
	}

	err = session.RunWithConfigs(ctx, system.SessionDir, 4, configs, testFn)
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

	_, err = c.SendTransaction(ctx, tx)
	assert.NoError(t, err, "client.Transaction()")
}

// assertHeight checks the height of the chain is num.
func assertHeight(ctx context.Context, t *testing.T, c coinpb.CoinClient, num uint64) {
	test.WaitUntil(t, time.Second*2, time.Millisecond*250, func() error {
		// We have the block of height num
		bcn, _ := c.Blockchain(ctx, &coinpb.BlockchainReq{BlockNumber: num})
		if bcn == nil {
			return errors.New("Got nil blockchain")
		}
		if bcn.Blocks[0].BlockNumber() != num {
			return errors.Errorf("Block had number %d, expected %d", bcn.Blocks[0].BlockNumber(), num)
		}
		// We don't have the block of height num + 1.
		bcn, err := c.Blockchain(ctx, &coinpb.BlockchainReq{BlockNumber: num + 1})
		if err == nil {
			bcn, _ := c.Blockchain(ctx, &coinpb.BlockchainReq{BlockNumber: num + 1, Count: uint32(num + 1)})
			return errors.Errorf("Expected block %d not to exist: %v", num+1, bcn)
		}
		return nil
	}, "blockChain should update: %s")
}

func assertAllHeight(ctx context.Context, t *testing.T, cs []coinpb.CoinClient, num uint64) {
	for _, c := range cs {
		assertHeight(ctx, t, c, num)
	}
}

func assertAccount(ctx context.Context, t *testing.T, c coinpb.CoinClient, pid peer.ID, balance, nonce uint64) {
	acc, err := c.GetAccount(ctx, &coinpb.AccountReq{PeerId: []byte(pid)})
	assert.NoError(t, err, "Account()")
	if acc != nil {
		assert.Equal(t, balance, acc.Balance, "account.Balance")
		assert.Equal(t, nonce, acc.Nonce, "account.Balance")
	}
}

func assertAllAccount(ctx context.Context, t *testing.T, cs []coinpb.CoinClient, pid peer.ID, balance, nonce uint64) {
	for _, c := range cs {
		assertAccount(ctx, t, c, pid, balance, nonce)
	}
}

func withCoinConfig(config cfg.ConfigSet, minerID peer.ID, genesis string) cfg.ConfigSet {
	conf := deepcopy.Copy(config).(cfg.ConfigSet)

	coinConf := conf["coin"].(coin.Config)
	coinConf.MinerID = peer.IDB58Encode(minerID)
	coinConf.GenesisBlock = genesis
	coinConf.BlockDifficulty = 6
	coinConf.DbPath = filepath.Join("data", "coin", "db")
	conf["coin"] = coinConf

	return conf
}
