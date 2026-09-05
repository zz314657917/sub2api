### DONE: prompt-audit-policy-matrix-s293-r2-lifecycle

## Changed Files

- `backend/internal/securityaudit/prompt_policy_history.go`
- `backend/internal/securityaudit/prompt_policy_lifecycle_test.go`

## Implemented

- Publish now rejects a draft whose `BaseConfigVersion` no longer equals the
  current active configuration version.
- Existing advisory-lock, transaction, history and Redis publication ordering is
  preserved; ordinary configuration PUT remains a separate lifecycle.

## Checks

- `go test ./internal/securityaudit ./internal/server/routes ./internal/server/middleware ./migrations -count=1` PASS
- `go build ./...` PASS
- Isolated lifecycle tests use `sqlmock` plus task-private `miniredis` to cover
  save/publish/rollback success, stale config/draft/base conflicts, missing
  drafts, write rollback, and commit-gated snapshot/notification ordering.
- `go test ./internal/securityaudit -run '^TestPolicyLifecycle' -count=1` PASS
- `go test ./internal/securityaudit -count=1` PASS
- `git diff --check -- backend/internal/securityaudit/prompt_policy_lifecycle_test.go docs/workflow/worker-results/prompt-audit-policy-matrix-s293-r2-lifecycle-result.md` PASS
- Isolated PostgreSQL concurrency/rollback coverage BLOCKED: no dedicated
  PostgreSQL resource was available and shared fixtures are destructive.

## Risks / Unverified

- No shared DB, Redis, deployment or provider operation performed.
- The new isolated tests are simulated transaction evidence only; they do not
  establish PostgreSQL transaction behavior or multi-instance runtime PASS.
- Full UI stale-draft conflict workflow is covered by R3 and remains subject to
  independent browser/runtime verification.
