### PASS: upstream-v0185-codex-bootstrap-s289

## Implemented

- Adapted upstream `1be69e56` and `421a83282` as one dependency-aware,
  manually ported handler change.
- Strictly converts only call-less Codex delegation or automation bootstrap
  outputs to Responses user messages before the existing tool-output context
  validation.
- Preserves the rejection path for real call/reference anchors, non-empty call
  IDs, unknown or mixed output types, duplicate JSON members, malformed XML,
  invalid automation IDs, mismatched memory paths, invalid last-run values, and
  empty prompts.
- Preserves JSON numeric precision through `Decoder.UseNumber` and leaves all
  non-target request data in input order.

## Checks

- `go test ./internal/handler -run 'TestNormalizeCodex(Delegation|Automation)Bootstrap' -count=10`: PASS.
- `go test ./internal/handler -count=1`: PASS.
- `go test ./cmd/server -run '^$' -count=1`: PASS.
- `go build ./...`: PASS.
- `gofmt -d`, `git diff --check`, and unmerged-path check: PASS.

## Scope

Only the S289 handler/test allowlist is changed. No route, service, billing,
database, frontend, provider, container, deployment, or protected dirty path
was changed.

## Unverified

Real OpenAI provider, Responses WebSocket, database, container, deployment, and
browser smoke remain outside this Sprint contract.
