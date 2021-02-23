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