# Indigo Services

The Indigo services provide support for the Indigo platform.
They help you build [Proof-of-Process](https://proofofprocess.org/) networks.
See the [Indigo Core](https://indigocore.org/) documentation for more details.

These services aren't enabled by default, you need to start them yourself:

```sh
StratumnNode> manager-start indigostore
StratumnNode> manager-start indigofossilizer
```

## Configuration

The services come with pre-configured values, but they aren't ready to run
out-of-the-box.
For each Indigo service, you'll need to configure several values yourself.

### Indigo Store

You'll need to configure the following values yourself in the `indigostore`
section in `indigo_node.core.toml`:

- _network_id_: the unique id of your Indigo network
- _storage_type_: the type of underlying storage used (in-memory or postgreSQL)
- _storage_db_url_: if using postgreSQL, the url of the database

### Indigo Fossilizer

You'll need to configure the following values yourself in the `indigofossilizer`
section in `indigo_node.core.toml`:

- _fossilizer_type_: the type of fossilizer used (e.g.: dummy, dummybatch, bitcoin...)
- _bitcoin_WIF_: secret key of your Bitcoin wallet (if using Bitcoin fossilizer)
- _bitcoin_fee_: amount of the fee to use for Bitcoin transactions (if using Bitcoin fossilizer)
