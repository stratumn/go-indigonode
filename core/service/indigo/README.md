# Indigo Service

The Indigo services on Alice provide support for the Indigo stack.
They allow you to run an Indigo node to build Proof-of-Process networks.
These services aren't enabled by default, you need to start them yourself:

```sh
Alice> manager-start indigostore
Alice> manager-start indigofossilizer
```

## Configuration

The services come with pre-configured values, but they aren't ready to run out-of-the-box.
For each Indigo service, you'll need to configure several values yourself.

### Indigo Store

You'll need to configure the following values yourself in the `indigostore` section in `alice.core.toml`:

* _network_id_: the unique id of your Indigo network
* _storage_type_: the type of underlying storage used (in-memory or postgreSQL)
* _storage_db_url_: if using postgreSQL, the url of the database
