### DONE: upstream-streaming-audit-s210

## Summary

- Adapted upstream `2f109e74c`: compact keepalive-only header/comment commits
  no longer suppress the Responses `response.failed` terminal event.
- Adapted upstream `c418fd522`: WebSocket security-audit results reuse only a
  matching allow decision for the same stage, turn, and SHA-256 payload.
- Added default-tag handler regressions for compact keepalive terminal output,
  same-turn deduplication, different payload/turn evaluation, and
  unavailable/flagged decision non-caching.

## Changed Files

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/handler/security_audit_helper.go`
- `backend/internal/handler/security_audit_helper_test.go`
- S210 workflow status, result, QA, and main log.

## Commands Run

- `gofmt -w` and `gofmt -d` over the four changed Go files: PASS.
- S210 focused handler regressions (`-count=10`): PASS.
- `go test ./internal/handler -count=1`: PASS.
- `go test ./cmd/server -run '^$' -count=0`: PASS.
- `git diff --check`, upstream ancestry, conflict-marker, and unmerged-index
  checks: PASS.

## Contract Compliance

- Product changes are limited to the four allowed handler/test paths.
- No route/cooldown, billing, persistence, schema/migration, frontend,
  configuration, dependency, container, push, deployment, or production path
  was changed.
- No cherry-pick, provider call, shared resource operation, remote push, or
  deployment was performed.

## Risks

- This is source-level and local-regression evidence only; it does not run a
  deployed WebSocket client, provider, shared infrastructure, or production
  traffic.
