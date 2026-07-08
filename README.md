# Mycel Go SDK

Go connector for MycelDB daemon APIs.

Module path:

```text
github.com/myceldb/mycel-go-sdk
```

The SDK generates Go protobuf/gRPC stubs from the language-independent `mycel-api` protobuf contracts and provides:

- daemon dial helpers
- plaintext/TLS/mTLS transport config
- user login and refresh helpers
- operator/admin login helper
- bearer-token metadata injection
- generated Admin and Client service clients under `gen/go/` after generation
- call timeout helpers
- session/transaction helpers
- thin graph/query convenience methods
- Admin backup policy/status/list/trigger/delete helpers

Generated code is not committed. Run generation before testing or building from a fresh checkout.

## Generate protobuf stubs

By default generation reads protobufs from a sibling checkout:

```text
../mycel-api/api/proto
```

Run:

```sh
make generate
```

Or set a custom API checkout path:

```sh
MYCEL_API_ROOT=/path/to/mycel-api make generate
```

Generated files are written to `gen/go/` and ignored by git.

## Test

```sh
make test
```

`make test` runs generation first, then `go test ./...`.

## Usage

```go
package main

import (
    "context"

    mycel "github.com/myceldb/mycel-go-sdk"
    clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
)

func main() {
    ctx := context.Background()
    client, err := mycel.Dial(ctx, mycel.Config{
        Addr:     "127.0.0.1:9091",
        Username: "alice",
        Password: "secret",
    })
    if err != nil {
        panic(err)
    }
    defer client.Close()

    _, err = client.Space.ListSpaces(client.AuthContext(ctx), &clientv1.ListSpacesRequest{})
    if err != nil {
        panic(err)
    }
}
```

Admin APIs use `DialAdmin`:

```go
admin, err := mycel.DialAdmin(ctx, mycel.Config{
    Addr:     "127.0.0.1:9091",
    Username: "operator",
    Password: "secret",
})
```

Admin backup helpers wrap `mycel.admin.v1.AdminBackupService`:

```go
policy, err := admin.GetBackupPolicy(ctx)
status, err := admin.GetBackupStatus(ctx)
trigger, err := admin.TriggerBackup(ctx, "before upgrade")
_ = policy
_ = status
_ = trigger
```

## Environment config

`ConfigFromEnv` reads:

- `MYCELD_GRPC_ADDR`
- `MYCEL_USERNAME`
- `MYCEL_PASSWORD`
- `MYCEL_ACCESS_TOKEN`
- `MYCEL_CALL_TIMEOUT`
- `MYCELD_TLS`
- `MYCELD_TLS_CA_FILE`
- `MYCELD_TLS_SERVER_NAME`
- `MYCELD_TLS_INSECURE_SKIP_VERIFY`
- `MYCELD_TLS_CLIENT_CERT_FILE`
- `MYCELD_TLS_CLIENT_KEY_FILE`
- `MYCEL_CLIENT_NAME`
- `MYCEL_CLIENT_VERSION`
- `MYCEL_CLIENT_PLATFORM`
- `MYCEL_CLIENT_DEVICE_LABEL`
