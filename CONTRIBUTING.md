# Contributing to Mycel Go SDK

Thank you for contributing to the MycelDB Go SDK. This repository provides Go bindings and convenience helpers for the MycelDB daemon APIs defined in [`mycel-api`](https://github.com/MycelDB/mycel-api).

## Repository scope

This repo **does** contain:

- Go SDK connection, authentication, session, transaction, graph, query, backup, and watch helpers.
- Generated Go protobuf/gRPC bindings under `gen/go/`.
- Tests and examples for Go SDK behavior.
- Build scripts for regenerating bindings from `mycel-api`.

This repo **does not** contain:

- MycelDB daemon/server implementation.
- The source `.proto` API contract. API contract changes belong in `mycel-api` first.
- Generated bindings for other languages.
- Product-specific application code.

## Local validation

Run the full test target before opening a PR:

```sh
make test
```

`make test` regenerates protobuf bindings and then runs:

```sh
go test ./...
```

If you are only checking Go tests after generation is already current, you can run `go test ./...` directly, but PRs should be validated with `make test`.

## Regenerating protobuf bindings

Generated Go bindings are committed under `gen/go/` so tagged Go module releases are self-contained.

By default, generation reads protobuf sources from a sibling checkout:

```text
../mycel-api/api/proto
```

To regenerate from the default location:

```sh
make generate
```

To use a different `mycel-api` checkout:

```sh
MYCEL_API_ROOT=/path/to/mycel-api make generate
```

When API contracts change:

1. Land or check out the matching `mycel-api` changes.
2. Run `make generate` in this repository.
3. Commit the resulting `gen/go/` changes together with any SDK helper/test updates.
4. Document compatibility or migration notes when public SDK behavior changes.

Do not hand-edit files under `gen/go/`. Regenerate them from `mycel-api` instead.

## Go SDK compatibility

The generated bindings follow the protobuf contract from `mycel-api`. SDK helper APIs should be kept stable and idiomatic.

Guidelines:

- Prefer additive helpers over breaking existing callers.
- Keep exported names, function signatures, and public struct fields stable unless a breaking change is intentional.
- Return typed or wrapped errors that callers can inspect with the helpers in this repo.
- Preserve context cancellation and timeout behavior.
- Avoid hiding authentication, authorization, retry, or stream-resume semantics that callers need to understand.
- Keep operation IDs as correlation metadata only; do not treat them as idempotency keys or credentials.
- Keep generated code and helper code clearly separated.

Before a stable `v1.0.0` release, the SDK may still evolve, but PRs should still call out any source-incompatible change. After `v1.0.0`, follow Go module semantic import versioning for breaking changes.

## Coding guidelines

- Run `gofmt` on Go files.
- Add or update tests for new helpers, auth behavior, error handling, stream setup, and query builders.
- Keep helper APIs small and close to the underlying gRPC contract.
- Do not log credentials, refresh tokens, bearer tokens, TLS key material, or private data.
- Prefer `context.Context` parameters for operations that can block.
- Keep default network behavior safe; insecure TLS settings must remain explicit opt-ins.

## Pull request expectations

A good SDK PR includes:

- A clear summary of the change and why it belongs in the SDK.
- Tests or an explanation for why tests are not applicable.
- `make test` results.
- Regenerated `gen/go/` files when the underlying API changed.
- README or example updates when public usage changes.
- Compatibility notes for exported helper/API changes.

## Security issues

Do not report security vulnerabilities in public issues. See [SECURITY.md](SECURITY.md) for private reporting instructions.
