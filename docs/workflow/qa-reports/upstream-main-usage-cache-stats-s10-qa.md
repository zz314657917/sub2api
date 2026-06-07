### PASS: upstream-main-usage-cache-stats-s10

## Findings

- PASS: Both approved S10 candidates were processed and represented by concrete commits on `codex/upstream-main-usage-cache-stats-s10`.
- PASS: All changed paths are inside the Sprint allowed path set.
- PASS: No `frontend/`, `skills/`, `backend/ent/`, `backend/migrations/`, `deploy/`, `knowledge/`, `docs/workflow/status.md`, `docs/workflow/spec.md`, `.github/`, `assets/`, or `README*` changes are present in `main...HEAD`.
- PASS: The API change is additive: `total_cache_creation_tokens` and `total_cache_read_tokens` were added while `total_cache_tokens` and `total_tokens` remain.
- PASS: Targeted usage/cache/contract tests passed.
- PASS: Service/repository/server regression tests passed.

## Executed Checks

- `git status --short --branch` -> clean on `codex/upstream-main-usage-cache-stats-s10` before workflow report edits.
- `git diff --check` -> PASS.
- denied path audit with `git diff --name-only main...HEAD` -> `DENIED_NONE`.
- `go test ./internal/service ./internal/repository ./internal/server -run "Usage|Stats|Cache|Contract" -count=1` -> PASS.
- `go test ./internal/service ./internal/repository ./internal/server -count=1` -> PASS.

## Not Run

- Frontend tests were not run because frontend paths are denied and untouched.
- `go test ./internal/handler ./cmd/server -count=1` was not run because this Sprint did not touch handler implementation or server startup wiring.

## Risks

- `/api/v1/usage/stats` now returns two extra JSON fields. This is intended as an additive, backward-compatible API contract expansion.
- Repository regression tests passed locally; no external DB blocker was encountered.

## Recommendation

PASS. The branch is ready for integration from current `main`.

## Integration Verification

- 2026-06-07: Created clean integration worktree `E:/codex-worktrees/sub2api/upstream-main-usage-cache-stats-s10-integration` from current `main@c6fefc8c6`.
- Merged `codex/upstream-main-usage-cache-stats-s10` without conflicts.
- `git status --short --branch` -> clean on `codex/upstream-main-usage-cache-stats-s10-integration`.
- `git diff --check main..HEAD` -> PASS.
- denied path audit with `git diff --name-only main..HEAD` -> `DENIED_NONE`.
- `go test ./internal/service ./internal/repository ./internal/server -run "Usage|Stats|Cache|Contract" -count=1` -> PASS.
- `go test ./internal/service ./internal/repository ./internal/server -count=1` -> PASS.
