---
task_id: upstream-main-v0139-backend-compat-s24
phase: contract-approved
owner: Codex
qa_mode: runtime
created_at: 2026-06-29 22:05 +08:00
---

# Task Contract: upstream v0.1.139 backend compatibility S24

## Goal

Port the next low-risk backend billing and key-state fixes from `upstream/main` after S23, without wholesale merging upstream or changing local product surfaces.

## Success Criteria

- OpenAI async usage billing preserves the request-time quota platform, including `ForcePlatform`, when posting usage after the original request context is detached.
- Fallback model pricing warning logs are emitted at most once per normalized model per process, without changing pricing results.
- API keys in `quota_exhausted` status are reactivated when their quota is changed to unlimited (`quota <= 0`).
- Targeted backend tests, `git diff --check`, and staged denied-path audit pass.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/usage_billing.go`
- `backend/internal/service/openai_quota_platform.go`
- `backend/internal/service/openai_quota_platform_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_embeddings.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/handler/openai_quota_platform_contract_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_service_delete_test.go`
- `backend/internal/service/api_key_service_quota_test.go`
- `docs/workflow/tasks/upstream-main-v0139-backend-compat-s24.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/worker-results/upstream-main-v0139-backend-compat-s24-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-backend-compat-s24-qa.md`

## Denied Paths

- Existing dirty `knowledge/*` files, unless separately requested.
- Frontend, public pages, Studio/Canvas, Model Plaza, payment pages, Ops/Keys UI.
- Ent schema/generated files, migrations, wire generation, VERSION, README, sponsors, CI/deploy.
- Payment/subscription/order currency/provider supported-types/affiliate rebate behavior beyond passing the quota platform into existing usage billing params.
- Codex JSON/developer-input transform changes from `b105cc0fd`; keep for a later S25 if needed.

## Constraints

- Do not cherry-pick upstream commits blindly; adapt selected behavior to local billing/trial/prepaid/account-share code.
- Do not alter local trial, prepaid balance, Studio Bridge, or account-share billing semantics.
- Handler changes may only capture request-time quota platform and pass it into `OpenAIRecordUsageInput`.
- Do not stage or commit unrelated dirty `knowledge/*` files.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test -tags=unit ./internal/service -run "TestGetModelPricing_FallbackWarn|TestGetModelPricing_GLM52|TestAPIKeyService_Update_ReactivatesQuotaExhaustedWhenQuotaUnlimited|TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform" -count=1
go test ./internal/handler -run "TestOpenAIRecordUsageInputsCarryQuotaPlatform" -count=1

cd F:/mcplugins/sub2api
git diff --check
git diff --cached --name-only | rg "^(knowledge/|frontend/|backend/ent/|backend/migrations/|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|deploy/|README|README_|assets/partners/)" || echo NO_DENIED_PATHS
```

## Output

- Code diff in allowed backend and workflow paths only.
- Worker-style result and QA report.
- Final report with ported/equivalent/skipped candidate accounting.

## Stop Rules

- Stop if the selected patch needs Ent/migration/wire generation, frontend UI, or product pricing policy changes.
- Stop if preserving quota platform conflicts with local trial/prepaid/account-share billing behavior.
- Stop if tests require real external OpenAI/OAuth/APIMart upstreams.
