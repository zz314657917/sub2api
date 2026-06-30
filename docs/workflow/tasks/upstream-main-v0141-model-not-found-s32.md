---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-01 00:00 +08:00
---

# Task Contract

## Task ID
upstream-main-v0141-model-not-found-s32

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small upstream port. No external worker is used.

## Goal
Port upstream `fcd3bc1272b3e283c172f153db4d75911cd93357` so gateway account-selection failures return `404 model_not_found` when the group has accounts but none support the requested model, while preserving `503 api_error` for empty pools and transient capacity exhaustion.

## Success Criteria
- Anthropic/Gemini/OpenAI gateway no-account errors are classified through a shared handler helper.
- Unsupported model in a non-empty account pool returns `404` with error type `model_not_found`.
- Empty pool, transient lookup failure, rate limit, quota pause, runtime block, and compact-account unsupported paths stay on `503`.
- Ops routing capacity-limited markers are not set for deterministic `model_not_found` responses.
- No frontend, Ent, migration, wire, deploy, README, VERSION, payment, OAuth completion, Grok quota, user proxy/account, or `knowledge/*` files are included.

## Context
- Repo: `F:/mcplugins/sub2api`
- Local anchor before S32: `b43b51d17 fix(auth): include client ip in acl denials`
- Upstream anchor after refresh: `v0.1.141-1-gdc1bc1545`
- Upstream reference:
  - `fcd3bc1272b3e283c172f153db4d75911cd93357 fix: return 404 model_not_found instead of 503 when no account supports the model`
- Latest `v0.1.141` tail includes frontend/admin user usage, payment display, sponsors/VERSION and related broad changes; they are intentionally out of scope.

## Allowed Paths
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/gemini_v1beta_handler.go`
- `backend/internal/handler/no_account_error.go`
- `backend/internal/handler/no_account_error_test.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_embeddings.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/service/gateway_model_availability.go`
- `backend/internal/service/gateway_model_availability_test.go`
- `backend/internal/service/openai_gateway_model_availability.go`
- `docs/workflow/tasks/upstream-main-v0141-model-not-found-s32.md`
- `docs/workflow/worker-results/upstream-main-v0141-model-not-found-s32-result.md`
- `docs/workflow/qa-reports/upstream-main-v0141-model-not-found-s32-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `knowledge/**`
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/(handler|user_account_handler|wire|user_proxy_handler).go`
- `backend/internal/repository/**`
- `backend/internal/server/routes/user.go`
- `backend/internal/service/(admin_service|openai_oauth_service|openai_oauth_service_state_test|proxy|proxy_service|user_account_service|user_proxy_service|wire).go`
- Payment, subscription, keys UI, ops UI, Grok routing, OAuth email completion flow, proxy/account ownership work, and production configuration paths.

## Constraints
- Do not merge or rebase `upstream/main`.
- Keep local `ForUser` account selection semantics where this branch already enforces user-owned/private pool behavior.
- Diagnose model availability by account mapping only; do not treat transient capacity signals as model-not-found.
- Be conservative on diagnosis errors: return the existing `503` fallback.
- Do not stage existing dirty Ent, migration, proxy/account, frontend, service, handler, or knowledge files unrelated to this task.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test -tags=unit ./internal/handler -run "TestClassifyNoAccountError" -count=1
go test -tags=unit ./internal/service -run "TestDiagnoseModelAvailabilityForPlatform" -count=1
cd F:/mcplugins/sub2api
git diff --check -- backend/internal/handler/gateway_handler.go backend/internal/handler/gateway_handler_chat_completions.go backend/internal/handler/gateway_handler_responses.go backend/internal/handler/gemini_v1beta_handler.go backend/internal/handler/no_account_error.go backend/internal/handler/no_account_error_test.go backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_embeddings.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_images.go backend/internal/service/gateway_model_availability.go backend/internal/service/gateway_model_availability_test.go backend/internal/service/openai_gateway_model_availability.go docs/workflow/tasks/upstream-main-v0141-model-not-found-s32.md docs/workflow/worker-results/upstream-main-v0141-model-not-found-s32-result.md docs/workflow/qa-reports/upstream-main-v0141-model-not-found-s32-qa.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0141-model-not-found-s32-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0141-model-not-found-s32-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementation requires frontend, Ent, migration, wire, route, repository, deploy, README, VERSION, payment, OAuth completion, Grok quota, user proxy/account, or `knowledge/*` changes.
- Stop if classification cannot distinguish non-empty unsupported-model pools from empty pools without broad account-selection refactor.
- Stop if targeted handler/service tests cannot compile without touching denied paths.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`
