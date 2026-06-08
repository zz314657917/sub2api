### PASS: upstream-main-release135-gateway-auth-s11

## Findings

- PASS: All approved S11 candidates were processed and represented by concrete commits on `codex/upstream-main-release135-gateway-auth-s11`.
- PASS: All changed paths are inside the Sprint allowed path set.
- PASS: No `frontend/`, `backend/ent/`, `backend/migrations/`, `skills/`, `assets/`, `README*`, `.github/`, `deploy/`, `knowledge/`, `docs/workflow/status.md`, or `docs/workflow/spec.md` changes are present in `main..HEAD`.
- PASS: Non-streaming JSON response handling, API Key exclusive group authorization, sticky session group validation, cross-group `previous_response_id` stripping, and `/responses` transport failover all have targeted test coverage.
- PASS: Local transport fault handling intentionally uses `AccountRepository.SetTempUnschedulable`, preserving local scheduler snapshot synchronization semantics.

## Executed Checks

- `git status --short --branch` -> clean on `codex/upstream-main-release135-gateway-auth-s11` before workflow report edits.
- `git diff --check main..HEAD` -> PASS.
- denied path audit with `git diff --name-only main..HEAD` -> `DENIED_NONE`.
- `go test ./internal/service -run "OpenAI|Gateway|Responses|ChatCompletions|Sticky|Previous|Transport|Failover|ContentType" -count=1` -> PASS.
- `go test ./internal/server/middleware -run "APIKey|Group|Exclusive|Allowed" -count=1` -> PASS.
- `go test ./internal/repository -run "APIKey|AllowedGroups" -count=1` -> PASS.
- `go test ./internal/service ./internal/server/middleware ./internal/repository -count=1` -> PASS.
- `go test ./internal/handler ./internal/server -run "OpenAI|Gateway|APIKey|Contract" -count=1` -> PASS.

## Not Run

- Frontend tests were not run because frontend paths are denied and untouched in this Sprint.
- Ent/migration checks were not run because this Sprint intentionally avoided database schema and migration changes.

## Risks

- `af19d4432` remains deferred because it is a larger proxy expiry/fallback feature requiring schema, migration, frontend, and API contract changes.
- Tag-after candidate `d251487da` was not included and should be evaluated separately if prompt cache key propagation is still desired.
- External upstream services were not called; validation is local unit/regression coverage plus diff/path audit.

## Recommendation

PASS. The branch is ready for integration into current `main`.
