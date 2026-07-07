# Task Contract: upstream-v0146-backend-safe-patches-s56

## Task ID

`upstream-v0146-backend-safe-patches-s56`

## Role

Generator / Codex direct integration.

## Goal

Port two low-risk backend fixes from upstream `v0.1.146` / `upstream/main` after S55:

- `e76e0499d`: sanitize NUL bytes in payment provider response payloads.
- `f881ff7cb`: support OpenAI models endpoints whose base URL does not end with `/v1`.

## Success Criteria

- Payment order result handling strips embedded NUL bytes before persisting or exposing provider response text.
- OpenAI upstream model listing handles non-`/v1` base URLs without producing malformed models URLs.
- Changes remain limited to backend service files and matching tests.

## Allowed Paths

- `backend/internal/service/payment_order.go`
- `backend/internal/service/payment_order_result_test.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/upstream_models_test.go`
- `docs/workflow/tasks/upstream-v0146-backend-safe-patches-s56.md`
- `docs/workflow/worker-results/upstream-v0146-backend-safe-patches-s56-result.md`
- `docs/workflow/qa-reports/upstream-v0146-backend-safe-patches-s56-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `.github/**`
- frontend files
- unrelated payment UI, scheduler, gateway refactor, batch-image, Grok, image namespace, or messages fallback changes

## Constraints

- Do not merge `upstream/main` or tag `v0.1.146` directly.
- Keep the work in an isolated worktree based on `origin/main`.
- Use `git cherry-pick -x` for source traceability.
- Do not touch the dirty main worktree.
- Do not use `git add .`.

## Acceptance Commands

Run from `backend` unless noted:

- `go test ./internal/service -run "Test.*(Payment.*Order|PaymentOrder|NUL|UpstreamModels|ModelsURL|OpenAIModels)" -count=1`
- `git diff --check origin/main..HEAD` from repo root

## Output

- Two upstream cherry-pick commits plus this workflow record.
- QA evidence under `docs/workflow/qa-reports/`.
- Clear final summary of validation and remaining deferred upstream items.

## Stop Rules

- Stop if a candidate requires frontend UI conflict resolution, migrations, deploy changes, broad gateway refactors, or unrelated product behavior.
- Stop if targeted validation fails in the touched service paths.
