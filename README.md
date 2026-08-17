# Mycel Go SDK

Go connector for MycelDB daemon APIs.

Module path:

```text
github.com/myceldb/mycel-go-sdk
```

The SDK generates Go protobuf/gRPC stubs from the language-independent `mycel-api` protobuf contracts and provides:

- daemon dial helpers
- plaintext/TLS/mTLS transport config
- user login, refresh, and logout helpers
- operator/admin login, refresh, and logout helpers
- automatic access-token refresh with one retry on expired-token `Unauthenticated`
- bearer-token metadata injection
- generated Admin and Client service clients under `gen/go/` after generation
- call timeout helpers
- session/transaction helpers
- thin graph/query convenience methods
- structured query builders and explain helpers for common indexed/text/semantic/path/aggregate shapes
- graph-change watch helpers
- Admin backup policy/status/list/trigger/delete helpers
- Admin cluster backup trigger/status/list/validate helpers

Generated code is committed under `gen/go/` so tagged Go module releases are self-contained.

## Generate protobuf stubs

The committed stubs on release branches are generated from the matching `mycel-api` release tag. The current `develop` branch is aligned with `mycel-api` `v0.8.0`.

By default generation reads protobufs from a sibling checkout:

```text
../mycel-api/api/proto
```

For release-aligned regeneration, check out `mycel-api` at the matching release tag, then run:

```sh
make generate
```

Or set a custom API checkout path:

```sh
MYCEL_API_ROOT=/path/to/mycel-api make generate
```

Generated files are written to `gen/go/` and should be committed when API contracts change.

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

`Dial` and `DialAdmin` store access-token expiry and refresh tokens returned by login. Before protected RPCs, the SDK refreshes near-expiry tokens automatically. If a protected unary RPC or stream setup fails with `Unauthenticated` because the access token is expired, the SDK refreshes once and retries once. You can also call `Refresh`, `Refresh`, `Logout`, or `Logout` directly.

Transaction operation IDs can be generated client-side and passed when beginning a transaction. They are correlation metadata only, not idempotency keys:

```go
operationID := mycel.NewOperationID()
tx, err := client.BeginReadWriteTransactionWithOperationID(ctx, sessionID, operationID)
if err != nil {
    panic(err)
}
// Perform graph writes against tx.GetTransactionId().
commit, err := client.CommitTransactionResult(ctx, tx.GetTransactionId())
if err != nil {
    panic(err)
}
_ = commit.GetOperationId() // matches operationID
```

Common structured query shapes can be built without hand-assembling every protobuf field:

```go
query, err := mycel.IndexedNodeLookupQuery("n", "Note", "title", "Roadmap", "note")
if err != nil {
    panic(err)
}
diag, err := client.ExplainQuery(ctx, txID, query)
if err != nil {
    panic(err)
}
_ = diag.GetPlan()
res, err := client.ExecuteQuery(ctx, txID, query, 50)
if err != nil {
    panic(err)
}
_ = res.GetResult().GetRows()
```

Other helpers include `OrderedNodeQuery`, `TextPredicateQuery`, `SemanticPredicateQuery`, `PathQuery`, `AggregateCountQuery`, `AggregatePropertyQuery`, and `ExplainGQL`.

Graph changes can be watched with `GraphChangeService.WatchGraphChanges` through the SDK helper. Persist the last processed `event.revision` and use it as `after_revision` when reconnecting:

```go
var lastRevision int64 // load from durable client checkpoint storage
watchCtx, stopWatch := context.WithCancel(ctx)
defer stopWatch()
stream, err := client.WatchGraphChanges(watchCtx, &clientv1.WatchGraphChangesRequest{
    SpaceId:        spaceID,
    DomainId:       domainID,
    AfterRevision:  &lastRevision,
    IncludeCurrent: true,
})
if err != nil {
    panic(err)
}
for {
    msg, err := stream.Recv()
    if err != nil {
        break // reconnect later with the last persisted revision
    }
    switch payload := msg.GetMessage().(type) {
    case *clientv1.WatchGraphChangesResponse_Event:
        event := payload.Event
        if event.GetOrigin().GetOperationId() == operationID {
            continue // optional: ignore a write issued by this workflow
        }
        // Apply event to local cache or derived state.
        lastRevision = event.GetRevision()
        // Persist lastRevision after successful processing.
    case *clientv1.WatchGraphChangesResponse_Gap:
        // Requested history is unavailable. Rebuild/resync local state, persist
        // a fresh checkpoint, and open a new stream.
        return
    }
}
```

Watch helpers refresh/retry only while opening the stream. They do not automatically reconnect or resume if a long-lived stream ends later. Track the last received `event.revision`, reconnect with `after_revision`, and handle `gap` by invalidating or rebuilding local derived state. If you stop reading early, cancel the parent context. Global `CallTimeout` / `MYCEL_CALL_TIMEOUT` applies to watch streams, so long-lived watchers should usually avoid a short global call timeout.

Admin backup helpers wrap `mycel.admin.v1.AdminBackupService`. Cluster backup archive formats use `adminv1` from `github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1`:

```go
policy, err := admin.GetBackupPolicy(ctx)
status, err := admin.GetBackupStatus(ctx)
trigger, err := admin.TriggerBackup(ctx, "before upgrade")
cluster, err := admin.TriggerClusterBackup(ctx, "before upgrade", "/mnt/mycel-backups", adminv1.BackupArchiveFormat_BACKUP_ARCHIVE_FORMAT_TAR_ZST)
_ = policy
_ = status
_ = trigger
_ = cluster
```

## Environment config

`ConfigFromEnv` reads:

- `MYCELD_GRPC_ADDR`
- `MYCEL_USERNAME`
- `MYCEL_PASSWORD`
- `MYCEL_ACCESS_TOKEN`
- `MYCEL_ACCESS_TOKEN_EXPIRE_TIME` (RFC3339)
- `MYCEL_REFRESH_TOKEN`
- `MYCEL_REFRESH_BEFORE` (duration, default `30s`)
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
