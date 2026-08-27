### DONE: group-model-match-auth-enforcement-s272

## Changed Files

- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s272_test.go`

## Implementation

- Removed the multi-group-route/pinned-account fast-path condition from
  `ResolveAPIKeyForModelRequest` while preserving its nil guards.
- Every non-nil API key now enters the existing
  `APIKeyService.ResolveForModelRequest` path, which already applies the
  effective group's `MatchesModel` rule.
- Added a focused mismatch regression using group 41 rules that exclude
  `gpt-5.6-luna`, plus a wildcard-match control that remains allowed.

## Commands Run

- Pre-fix: `go test ./internal/server/middleware -run '^TestS272' -count=1`
  - Expected FAIL: mismatch request incorrectly returned `ok=true`.
- Post-fix: `go test ./internal/server/middleware -run '^TestS272' -count=10`
  - PASS.
- `gofmt -w internal/server/middleware/api_key_auth.go internal/server/middleware/api_key_auth_s272_test.go`
- `git diff --check`
  - PASS.
- `git ls-files -u`
  - PASS, no unmerged entries.

## Contract Compliance

- Product changes are limited to the two allowed middleware paths.
- S271 implementation/worktree, schema, persistence, routing policy, billing,
  frontend, dependencies, provider traffic, containers, shared data, commit,
  push and `outputs/**` were not changed by S272.

## Risks

- Full affected package regressions and server compilation remain for the
  independent QA gate.
- No provider traffic, deployment, container update, or production API probe
  is authorized or claimed.

## Knowledge Candidates

- None. The repair is directly captured by its regression and workflow
  artifacts.
