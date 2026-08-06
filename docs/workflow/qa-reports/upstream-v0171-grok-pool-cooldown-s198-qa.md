### PASS: upstream-v0171-grok-pool-cooldown-s198

## Findings

- The local Grok handler only bypassed the default `5xx` cooldown for pool-mode accounts. It still temporarily unscheduled pool-mode accounts for `401`, `402`, `403`, and `429` responses.
- S198 moves the existing pool-mode bypass before default status handling. Non-pool cooldown reasons, durations, and `Retry-After` parsing remain on their existing branches.
- The focused regression covers `401`, `402`, `403`, `429`, and `502` for both pool and non-pool API-key accounts.

## Executed Checks

- `go test ./internal/service -run '^TestHandleGrokAccountUpstreamError.*PoolMode' -count=1`: passed.
- `go test ./internal/service -run '^TestHandleGrokAccountUpstreamError' -count=1`: passed.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0`: passed.
- `gofmt -w internal/service/openai_gateway_grok.go internal/service/openai_gateway_grok_s115_test.go` and `git diff --check`: passed.

## Unverified Risks

- No live Grok request, OAuth credential, scheduler database state, container, deployment, primary-worktree modification, push, or merge to `main` was performed.
- The local branch has no equivalent of upstream configurable forbidden policies or quota-snapshot rate-limit persistence; those independent features remain intentionally out of scope.

## Recommendation

Commit the scoped local mapping to the isolated integration branch. It has focused source-level coverage but not live Grok certification.
