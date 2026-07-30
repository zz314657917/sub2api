---
task_id: upstream-small-fixes-s115
repo: F:/mcplugins/sub2api
phase: done
owner: codex
source: upstream 38cc916f7, 51f354f5c
---

## Task ID

`upstream-small-fixes-s115`

## Role

Planner/Generator/Evaluator by Codex; no worker delegation is needed for these
two isolated backend fixes.

## Goal

Port two independent upstream behaviors into the local topology:

1. Do not start the pricing remote-sync scheduler when `pricing.remote_url` is
   blank.
2. Do not persist a two-minute Grok cooldown for pool-mode accounts after 5xx
   responses; preserve the existing cooldown for non-pool accounts.

## Success Criteria

- Blank or whitespace-only pricing URL leaves the scheduler wait group empty and
  logs that remote sync is disabled.
- Non-empty invalid pricing URLs still reach normal URL validation.
- Pool-mode Grok 5xx leaves runtime and durable scheduling state unchanged.
- Non-pool Grok 5xx keeps the existing two-minute cooldown.
- Focused tests, formatting, diff, conflict, and exact-path checks pass.

## Allowed Paths

- `backend/internal/service/pricing_service.go`
- `backend/internal/service/pricing_service_test.go`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_gateway_grok_test.go`
- `backend/internal/service/openai_gateway_grok_s115_test.go`
- `docs/workflow/tasks/upstream-small-fixes-s115.md`
- `docs/workflow/qa-reports/upstream-small-fixes-s115-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `frontend/**`
- `backend/migrations/**`
- `backend/ent/**`
- `deploy/**`
- `knowledge/**`
- `outputs/**`
- `frontend/src/views/admin/group-buy/**`
- The separate `E:/codex-worktrees/sub2api/group-buy-lifecycle-refund-hardening-s110` worktree
- Any path not listed under Allowed Paths

## Constraints

- Preserve local pricing fallback, scheduler lifecycle, Grok 401/402/403/429
  policies, and non-pool 5xx behavior.
- Keep the new Grok regression in a default-tag test file because the existing
  `//go:build unit` file has unrelated compile drift in this local branch.
- Do not change pool retry status-code configuration or generic rate-limit
  behavior.
- Do not commit, push, deploy, or refresh containers from the primary worktree.

## Acceptance Commands

Run from `F:/mcplugins/sub2api/backend`:

```powershell
go test ./internal/service -run "TestPricingSchedulerBlankRemoteURLDoesNotStart|TestPricingNonEmptyInvalidRemoteURLStillReturnsValidationError" -count=1
go test ./internal/service -run "TestHandleGrokAccountUpstreamError5xxRespectsPoolMode" -count=1
gofmt -d internal/service/pricing_service.go internal/service/pricing_service_test.go internal/service/openai_gateway_grok.go internal/service/openai_gateway_grok_test.go
```

Run from `F:/mcplugins/sub2api`:

```powershell
git diff --check
git diff --name-only --diff-filter=U
```

## Output

- Changed paths, focused test output, and final PASS/FAIL conclusion.
- Any pre-existing aggregate unit-tag compile drift must be recorded rather
  than attributed to S115.

## Stop Rules

- Stop if a fix requires changes outside the four business/test paths.
- Stop if pool-mode 401/403/429 behavior changes.
- Stop if tests require a database migration or deployment change.

## Budget

`worker_mode: codex-direct`
`qa_mode: runtime`
