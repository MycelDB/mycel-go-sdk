# Agent instructions for mycel-go-sdk

This repository is the Go SDK for MycelDB daemon APIs. It contains Go helper code and committed generated Go protobuf/gRPC bindings from `mycel-api`.

## Scope

Agents may edit:

- Go SDK source and tests in the repository root.
- Documentation such as `README.md`, `CONTRIBUTING.md`, and examples.
- Build scripts and CI configuration.
- Generated files under `gen/go/` only by running the generation workflow.

Agents must not add:

- MycelDB daemon implementation code.
- Source `.proto` contract changes. Change `mycel-api` first, then regenerate here.
- Generated bindings for non-Go languages.
- Product-specific application code.

## Generated code

Generated bindings live under `gen/go/` and are committed so Go module releases are self-contained.

Do not hand-edit files under `gen/go/`. To update generated bindings, use:

```sh
make generate
```

By default this reads protobuf definitions from `../mycel-api/api/proto`. Use `MYCEL_API_ROOT=/path/to/mycel-api make generate` when needed.

## Go SDK rules

When changing SDK helper code:

- Preserve exported Go APIs unless a breaking change is explicitly requested.
- Prefer additive helpers and backward-compatible behavior.
- Keep helper behavior close to the underlying gRPC contract.
- Use `context.Context` for blocking/network operations.
- Preserve context cancellation and timeout behavior.
- Do not log or expose access tokens, refresh tokens, passwords, TLS private keys, or other secrets.
- Keep insecure transport/TLS options explicit opt-ins.
- Add or update tests for helper behavior, error mapping, auth refresh, transaction/session behavior, query helpers, and stream setup.
- Run `gofmt` on changed Go files.

## API contract changes

If a requested change requires new or changed protobuf definitions:

1. Make the contract change in `mycel-api` first.
2. Validate the API repo according to its `AGENTS.md` / `CONTRIBUTING.md`.
3. Regenerate Go bindings here with `make generate`.
4. Update SDK helpers/tests/docs if needed.

Do not invent generated-code changes manually.

## Validation

Before handing off changes, run:

```sh
make test
```

If generation is intentionally skipped during exploration, say so explicitly and run at least:

```sh
go test ./...
```

Report exactly which commands were run and their results.

## Documentation

When public SDK usage changes, update `README.md` or examples in the same change. Keep generated-code policy and API-version alignment notes accurate.
