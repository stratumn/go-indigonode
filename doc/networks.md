# Networks

## Joining the public network

The default behavior is to create a public node and join the public network.

If you just run:

```bash
indigo-node init
indigo-node up
```

This is what will happen.
Your node will connect to known public bootstrap nodes (defined in the
configuration's `bootstrap.addresses` field).

## Creating a private network

For deployment convenience, private networks currently require a coordinator
node that will have more capabilities than other nodes in the network.

This introduces some centralization of power in the system, which is why
we plan on adding fully decentralized private network in the future.

We are also working on a UI to simplify network deployment and management.

### Creating the coordinator node

If plan on being the network coordinator, create your node with the following
command:

```bash
indigo-node init --private-with-coordinator --private-coordinator
indigo-node up
```

Your node won't connect to the public network and will sit idle, waiting
for connections.

Network participants will send a request to your node to join the network.
You can see the pending requests in the `data/network/proposals.json` file.

Once you've confirmed that a given ID is correctly owned by someone that
you want to allow in your network, you can use the accept command:

```bash
IndigoNode> bootstrap-accept <PeerID>
```

The node will be added to the network and all participants notified.

### Joining a private network

If you're not the network coordinator, you'll need her ID and address to be
able to join the network.

Create your node with the following command:

```bash
indigo-node init --private-with-coordinator --coordinator-addr <multiaddr>
```

The coordinator address should uses the IPFS multiformat and specify the ID
(e.g. `ip4/127.0.0.1/tcp/8903/ipfs/12D3KooWN35NseW9wy5MdBbWSAb1CaJuUTeZAPXjRs82URjdo1jE`).

When you run `indigo-node up` your node will try to connect to the coordinator
and wait until it gets approved.

### Updating a private network

Once the coordinator has accepted the initial participants, it should complete
the bootstrap phase with the following command:

```bash
IndigoNode> bootstrap-complete
```

Participants can now send proposals to add or remove participants.

Adding a participant requires the approval of the coordinator with the
`bootstrap-accept` command.

Removing a participant requires the approval of all current participants with
the `bootstrap-accept` command.

The coordinator has special rights: using `bootstrap-addnode` and
`bootstrap-removenode` on the coordinator node applies the change immediately.
