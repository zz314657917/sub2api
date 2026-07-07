# Task Contract: upstream-v0146-small-safe-patches-s55

## Task ID

`upstream-v0146-small-safe-patches-s55`

## Role

Generator / Codex direct integration.

## Goal

Port the next small, safe subset from upstream `v0.1.146` / `upstream/main` after S54:

- `438f17be5`: avoid compact JSON usage loss from a substring-based SSE heuristic.
- `fd64d07e6`: strip illegal `item_*` ids from Codex continuation `function_call` items.
- `cbfeab964`: default Antigravity gateway forwarding to the production endpoint.
- `a1b2b32e0`: prevent silent `usage_logs` drops under queue overflow.
- `f3a3a0869`: optimize stale concurrency slot cleanup.

## Success Criteria

- Non-streaming compact JSON bodies whose text contains `data:` or `event:` are not misrouted into SSE parsing.
- Codex continuation call-input items keep valid `fc*` ids and drop invalid replayed `item_*` ids.
- Antigravity gateway forwarding defaults to production unless explicitly configured otherwise.
- Usage log recording applies backpressure or synchronous fallback instead of silently dropping records on queue overflow.
- Concurrency slot cleanup removes expired slots consistently across service/cache paths.
- Changes are validated with targeted backend checks and committed on the isolated worktree branch.

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/handler/gateway_handler_warmup_intercept_unit_test.go`
- `backend/internal/handler/gateway_helper_fastpath_test.go`
- `backend/internal/handler/gateway_helper_hotpath_test.go`
- `backend/internal/repository/concurrency_cache.go`
- `backend/internal/repository/concurrency_cache_integration_test.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_integration_test.go`
- `backend/internal/service/antigravity_gateway_service.go`
- `backend/internal/service/concurrency_service.go`
- `backend/internal/service/concurrency_service_test.go`
- `backend/internal/service/concurrency_slot_cleanup_test.go`
- `backend/internal/service/gateway_multiplatform_test.go`
- `backend/internal/service/gateway_record_usage_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_codex_function_call_id_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/usage_record_worker_pool.go`
- `backend/internal/testutil/stubs.go`
- `docs/workflow/tasks/upstream-v0146-small-safe-patches-s55.md`
- `docs/workflow/worker-results/upstream-v0146-small-safe-patches-s55-result.md`
- `docs/workflow/qa-reports/upstream-v0146-small-safe-patches-s55-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `.github/**`
- `README*`
- unrelated frontend theme, payment, welfare, public-page, auth, or container files
- batch image, Grok video, messages fallback, or broad upstream feature series not listed in the Goal

## Constraints

- Do not merge `upstream/main` or tag `v0.1.146` directly.
- Keep the work in an isolated worktree based on `origin/main`.
- Preserve local S54 API Key concurrency display behavior while accepting the slot-cleanup backend fix.
- Use `git cherry-pick -x` so each upstream source commit remains traceable.
- Do not touch the dirty main worktree.
- Do not use `git add .`.

## Acceptance Commands

Run from `backend` unless noted:

- `go test ./internal/service -run "Test.*(Compact|SSE|Codex|FunctionCall|Antigravity|Usage|Queue|Concurrency|Slot)" -count=1`
- `go test ./internal/repository -run "Test.*(UsageLog|Concurrency)" -count=1`
- `go test ./internal/config -count=1`
- `go test ./internal/handler -run "Test.*(Gateway|Concurrency|Warmup|Fastpath|Hotpath)" -count=1`
- `git diff --check` from repo root

## Output

- Five upstream cherry-pick commits plus this workflow record.
- QA evidence under `docs/workflow/qa-reports/`.
- Clear final summary of validation and deferred upstream candidates.

## Stop Rules

- Stop if a candidate requires migrations, Ent regeneration, deploy changes, broad frontend work, or unrelated product changes.
- Stop if conflict resolution would overwrite local S54 behavior rather than adapt around it.
- Stop if targeted validation exposes a regression in the selected gateway, usage log, or concurrency paths.
