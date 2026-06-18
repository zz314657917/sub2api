### DONE: upstream-main-openai-quota-reset-s17

## Summary

Ported upstream `b816949291f972586cd3c37138ca741869b8a3f0 feat(openai-quota): query + reset rate-limit credits for OpenAI accounts` as an independent S17 change. The local port keeps the feature scoped to admin OpenAI OAuth accounts and does not merge `upstream/main`.

## Changed Files

- `backend/internal/service/openai_quota_service.go`
- `backend/internal/service/openai_quota_service_test.go`
- `backend/internal/service/openai_quota_unit_exports.go`
- `backend/internal/service/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/admin/openai_oauth_handler.go`
- `backend/internal/handler/admin/openai_oauth_handler_quota_test.go`
- `backend/internal/server/routes/admin.go`
- `frontend/src/api/admin/accounts.ts`
- `frontend/src/components/account/AccountUsageCell.vue`
- `frontend/src/components/account/OpenAIQuotaResetCell.vue`
- `frontend/src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts`
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
- `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Implementation Notes

- Added `OpenAIQuotaService` with `QueryUsage` and `ResetCredit`, using the existing OpenAI token provider and privacy client factory.
- Upstream endpoints are kept to ChatGPT WHAM usage and reset-credit consume URLs.
- Account proxy handling prefers eager-loaded `account.Proxy` and falls back to `proxyRepo.GetByID` when required.
- Added admin routes:
  - `GET /api/v1/admin/openai/accounts/:id/quota`
  - `POST /api/v1/admin/openai/accounts/:id/reset-quota`
- Kept existing local `/api/v1/admin/accounts/:id/reset-quota` account quota semantics unchanged.
- Added `OpenAIQuotaResetCell.vue` and wired it into the OpenAI OAuth branch of `AccountUsageCell` without duplicating local 5h/7d usage bars.
- Added modular i18n entries under `frontend/src/i18n/locales/{zh,en}/admin/accounts.ts`; root locale files were not modified.

## Commands Run

- `go test -tags=unit ./internal/service -run "TestOpenAIQuota" -count=1`
- `go test -tags=unit ./internal/handler/admin -run "TestOpenAIOAuthHandler.*Quota" -count=1`
- `go test ./internal/service -run "^$" -count=1`
- `go test ./internal/handler/admin -run "^$" -count=1`
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts"`
- `git diff --check`
- Denied-path audit command from the contract.

## Test Output

- Backend service OpenAI quota unit tests passed.
- Backend admin handler quota unit tests passed.
- Normal non-`unit` package compile checks for service and handler/admin passed, confirming `openai_quota_unit_exports.go` does not leak into ordinary builds.
- Targeted frontend Vitest passed for 2 files / 20 tests.
- `git diff --check` passed with only repository line-ending warnings.
- Denied-path audit returned `NO_DENIED_PATHS`.

## Risks

- The local tests mock upstream ChatGPT endpoints; no real `chatgpt.com` quota/reset request was sent.
- The feature depends on OpenAI OAuth accounts having valid `chatgpt_account_id`, `organization_id`, and refreshable credentials.
- Full frontend Vitest was not rerun in S17 because S15 already recorded unrelated product-area failures in Studio/Canvas/navigation/payment tests.

## Knowledge Candidates

- OpenAI upstream quota reset is a full admin feature chain and should stay as its own Sprint, separate from low-risk compatibility patches.
- In this fork, account usage UI i18n belongs in modular `frontend/src/i18n/locales/{zh,en}/admin/accounts.ts`, not upstream root locale files.
