CLI meant to be passed as `--tx-notify` argument to Monero Wallet, or `--block-notify` argument to Monero Daemon.

It fetches extra context about the tx from the Wallet RPC, or the block from the Monero Daemon RPC, and then
pushes it to NATS

**It's an early work in progress**

### Compiling

* Statically (production ready)

```bash
GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o publisher
```

* Dynamically (dev/testing)

```bash
go build -o publisher
```

* Run tests

```bash
go test -v
```

### Usage

Run `./publisher help` for detailed help.

It implements a couple of CLI commands:

* `./publisher tx <txid>`: Gathers extra context about the Tx and publishes it to NATS
* `./publisher block <blockHash>`: Gathers extra context about the Block and publishes it to NATS

It takes the following optional flags:

* `--wallet`: URL to the Monero Wallet RPC
* `--daemon`: URL to the Monero Daemon RPC
* `--nats`: URL to the NATS Streaming server
* `--ancestors`: Max number of ancestor blocks' hashes to include with every published block