CLI meant to be passed as `--tx-notify` argument to Monero Wallet.

It fetches extrac context about the tx from the Wallet RPC, and then
pushes the tx with its context to NATS

It's an early work in progress

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

### Configuration

`publisher` takes the following env vars:

* `WALLET_RPC_URL`: Points to the wallet RPC interface that can be used to fetch transactions context
* `NATS_URL`: Points the NATS server where events should be pushed to