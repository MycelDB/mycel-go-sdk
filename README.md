# Mycel Go SDK

Go connector for MycelDB daemon APIs.

Module path:

```text
github.com/myceldb/mycel-go-sdk
```

The SDK depends on `github.com/myceldb/mycel-api` for protobuf/gRPC stubs and provides:

- daemon dial helpers
- plaintext/TLS/mTLS transport config
- user login and refresh helpers
- operator/admin login helper
- bearer-token metadata injection
- raw generated Admin and Client service clients
- call timeout helpers
- session/transaction helpers
- thin graph/query convenience methods

## Usage

```go
ctx := context.Background()
client, err := mycel.Dial(ctx, mycel.Config{
    Addr:     "127.0.0.1:9091",
    Username: "alice",
    Password: "secret",
})
if err != nil {
    // handle error
}
defer client.Close()

spaces, err := client.Space.ListSpaces(client.AuthContext(ctx), &clientv1.ListSpacesRequest{})
```

Admin APIs use `DialAdmin`:

```go
admin, err := mycel.DialAdmin(ctx, mycel.Config{
    Addr:     "127.0.0.1:9091",
    Username: "operator",
    Password: "secret",
})
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

## Local development

This repository currently uses a local replace for sibling development:

```go
replace github.com/myceldb/mycel-api => ../mycel-api
```

Remove that replace when consuming a public module/tag directly.
