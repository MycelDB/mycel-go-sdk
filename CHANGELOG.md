# Changelog

All notable changes to the Mycel Go SDK should be documented in this file.

This project follows the spirit of [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and uses Go module semantic versioning. Before `v1.0.0`, exported helper APIs may still evolve, but source-incompatible changes should be called out clearly.

## [Unreleased]

### Added

- Added typed direct-semantic search and hard-eligibility filter helpers over the existing v0.21 contract, preserving opaque continuation tokens and daemon-owned diagnostics.

## [v0.21.0] - 2026-10-27

### Added

- Regenerated Go bindings for channel-owned Search queries; deprecated top-level `SearchRequest.query` is removed and callers must populate `lexical.query` and/or `semantic.query`.
- Regenerated Go bindings for the enum-only `SemanticTriggerPolicy.events` contract; the legacy string representation and transitional `event_types` field are removed. This is intentionally source- and wire-incompatible with earlier unreleased bindings.
- Regenerated Go bindings for direct semantic pagination, hard-eligibility filters, safe result provenance, and Admin semantic reliability/rebuild diagnostics.
- Added semantic provider-budget policy, authenticated enforcement-capability discovery, retained-scope, and immutable attempt-event bindings.
- Regenerated procedure-inference attempt polling, typed automation-run, procedure control-barrier, and procedure/binding Intelligence Access bindings.
- Regenerated vector-store UUID identity, normalized key/display-name uniqueness documentation, and structured catalog-conflict bindings.

### Compatibility

- Removing top-level `SearchRequest.query` is intentionally source-incompatible. Search callers must migrate to the query field for each enabled retrieval channel.
- Generated additions require a daemon advertising the corresponding semantic reliability capability when applications depend on the new RPCs.
- `AdminSemanticMaintenanceServiceClient` and `AdminSemanticMaintenanceServiceServer` add reliability-status, rebuild, provider-budget capability/status, retained-scope, and immutable attempt-event methods. Custom interface implementations, mocks, and test doubles must implement them; servers embedding `UnimplementedAdminSemanticMaintenanceServiceServer` inherit forward-compatible stubs.
- Client and Admin automation service interfaces gain procedure-attempt query/polling and procedure enable/disable methods. Custom interface implementations and mocks must add them; embedded generated unimplemented servers remain forward-compatible.

## [v0.20.1] - 2026-10-20

### Fixed

- Regenerated committed Go protobuf/gRPC bindings from `mycel-api v0.20.1` so session, structured-query, and domain-export comments describe read-only operations as repeatable reads pinned to `base_revision`.

### Compatibility

- Documentation-only generated-binding correction. Go APIs and protobuf wire behavior are unchanged from v0.20.0; use with Mycel daemon/API v0.20.1 or later in the v0.20 line for the corrected public contract documentation.

## [v0.20.0] - 2026-10-19

### Added

- Added durable graph transaction commit idempotency helpers with explicit principal/space/domain-scoped keys, preserved request identity, and access to original commit/replay metadata.
- Added caller-bounded retry helpers that retry only ambiguous commit outcomes by default and preserve cancellation, transaction ID, scope, and key.
- Regenerated committed Go protobuf/gRPC bindings from `mycel-api v0.20.0`, including commit idempotency fields and cluster runtime diagnostics.

### Compatibility

- Best used with Mycel daemon/API v0.20.0. Existing non-idempotent transaction commit helpers remain supported, and retries remain opt-in.

## [v0.19.0] - 2026-10-09

### Removed

- Removed legacy space ACL grant helpers (`GrantSpaceUser`, `GrantSpacePrincipal`, and `SpaceGrantInfo`) and regenerated bindings after the admin space grant RPC was removed from `mycel-api`.

### Compatibility

- Best used with Mycel daemon/API v0.19.0 for matching identity scoped space access semantics. Callers should use identity role/capability grant helpers instead of legacy space ACL helpers.

## [v0.18.0] - 2026-10-07

### Added

- Regenerated Go protobuf/gRPC bindings for `mycel-api` v0.18.0, including asynchronous cluster backup operation contracts, lifecycle states, cancellation flow, retry hints, and readiness blockers.
- Added `AdminClient` helper methods for async cluster backup start, status, cancellation, listing, and backup-set validation.

### Changed

- Updated backup helper documentation to steer cluster backup callers toward the asynchronous operation API.

### Compatibility

- Best used with Mycel daemon/API v0.18.0 for matching async cluster backup operation semantics.

## [v0.17.0] - 2026-09-30

### Added

- Regenerated Go protobuf/gRPC bindings for `mycel-api` v0.17.0, including client space export job APIs, admin Raft snapshot APIs, and graph checkpoint/index status fields.

## [v0.16.0] - 2026-09-24

### Added

- Regenerated Go protobuf/gRPC bindings for `mycel-api` v0.16.0, including graph reference replacement messages and results.
- Added `GraphClient.ReplaceReferences` for transaction-scoped server-side reference reconciliation.

## [v0.15.0] - 2026-09-18

### Added

- Regenerated Go protobuf/gRPC bindings for `mycel-api` v0.15.0, including dimensioned cluster readiness fields.

## [v0.12.0] - 2026-09-09

### Added

- Regenerated Go protobuf/gRPC bindings for hybrid search API additions, including weighted fusion options, metadata filters, and source diagnostics.

## [v0.11.0] - 2026-09-07

### Added

- Regenerated Go protobuf/gRPC bindings for the lexical search APIs and exposed thin `Client.Search` and `AdminClient.LexicalMaintenance` service clients.

## [v0.9.0] - 2026-08-31

### Added

- First public-release baseline for the MycelDB Go SDK.
- Open-source project documentation: contributing guide, security policy, code of conduct, changelog, pull request template, issue templates, and README badges.
- README environment configuration table with variable defaults and descriptions.

### Changed

- Regenerated committed Go protobuf/gRPC bindings from the matching `mycel-api` API contract so tagged Go module releases are self-contained.
- Documented generated-binding policy, SDK compatibility expectations, and repository boundaries.

## Release notes policy

For each release, add a dated section such as:

```md
## [v0.9.0] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
```

Include compatibility notes for exported SDK helpers, authentication behavior, timeout/retry behavior, TLS behavior, generated API bindings, and any required matching `mycel-api` tag or commit.
