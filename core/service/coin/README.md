# Coin service

The coin service implements a simple proof-of-work blockchain.
This service isn't enabled by default, you need to start it yourself:

```bash
Alice> manager-start coin
```

Once started, it will start mining blocks (this is CPU intensive).

## Configuration

Most of the configuration values will be automatically set.
You can tweak them in the `coin` section in `alice.core.toml`.
But be careful because other nodes in the network might reject your blocks if you use invalid configuration values.

In order to be able to start the coin service, you'll need to provide your public key in that configuration file.
This public key will be used to send you miner rewards when you produce blocks.
Only Ed25519 keys are currently supported. Here is how you can generate a key pair:

```go
import (
    crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

func GenerateKeyPair() error {
    privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
    if err != nil {
        return err
    }

    pubKeyBytes, err := pubKey.Bytes()
    if err != nil {
        return err
    }

    // This is how your public key should be encoded in alice.core.toml
    configPublicKey := crypto.ConfigEncodeKey(pubKeyBytes)
}
```

Make sure you correctly store your private key: if you lose it, you'll lose all your rewards!
