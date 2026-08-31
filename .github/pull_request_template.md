## Summary

<!-- What changed and why? -->

## SDK compatibility

- [ ] This change is backward compatible for exported SDK APIs.
- [ ] This change may be source-incompatible and has maintainer approval.
- [ ] README/examples are updated if public usage changed.
- [ ] Error, auth refresh, timeout, TLS, retry, or stream behavior changes are documented.

## Generated API bindings

- [ ] No generated files were changed by hand.
- [ ] `gen/go/` was regenerated with `make generate` if the API contract changed.
- [ ] The matching `mycel-api` branch/tag/commit is identified when generated files changed.
- [ ] Non-Go generated bindings were not added.

## Repository boundaries

- [ ] No daemon implementation code is added.
- [ ] Source `.proto` contract changes are not added here.
- [ ] Product-specific application behavior is not added.
- [ ] Secrets, tokens, passwords, and TLS key material are not logged or exposed.

## Validation

- [ ] `make test` passes.

## Notes

<!-- Migration notes, downstream impact, or follow-up work. -->
