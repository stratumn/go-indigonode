# Coin Service

The Coin Service implements a simple proof-of-work blockchain.
It is a proof-of-concept and shouldn't be used for a real crypto-asset.

This service isn't enabled by default, you need to start it yourself:

```bash
node> manager-start coin
```

Once started, it will start mining blocks (this is CPU intensive).

## Configuration

Most of the configuration values will be automatically set.
You can tweak them in the `coin` section in `stratumn_node.core.toml`.
But be careful because other nodes in the network might reject your blocks
if you use invalid configuration values.

In order to be able to start the Coin service, you'll need to provide your
peer ID in that configuration file.
This peer ID will be used to send you miner rewards when you produce blocks.
Only Ed25519 keys are currently supported. Here is how you can generate a key pair:

```go
import (
    crypto "github.com/libp2p/go-libp2p-crypto"
    peer "github.com/libp2p/go-libp2p-peer"
)

func GenerateKeyPair() error {
    privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
    if err != nil {
        return err
    }

    minerID, err := peer.IDFromPublicKey(pubKey)
    if err != nil {
        return err
    }

    // This is how your miner_id should appear in stratumn_node.core.toml
    configMinerID := peer.IDB58Encode(minerID)
}
```

Note that you can directly use your node's default peer ID if you wish.
But you can also set it to any peer ID for which you have the private key.

Make sure you correctly store your private key: if you lose it,
you'll lose all your rewards!
