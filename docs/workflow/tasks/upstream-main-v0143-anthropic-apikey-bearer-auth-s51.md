---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 17:58 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-anthropic-apikey-bearer-auth-s51

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `7869b7fe38d7e9c0341c9bb0f69af3dbcef6f465` so Anthropic API Key accounts can choose whether upstream requests authenticate with the historical `x-api-key` header or `Authorization: Bearer <token>`.

## Success Criteria
- Missing, invalid, non-Anthropic, and non-API-key accounts continue to use `x-api-key`.
- Anthropic API Key accounts with `extra.anthropic_apikey_auth_scheme = "authorization_bearer"` use `Authorization: Bearer <token>` and do not send `x-api-key`.
- The selected auth scheme is honored by:
  - normal Anthropic gateway request construction,
  - Anthropic API-key passthrough messages,
  - Anthropic API-key passthrough count_tokens,
  - account connection tests,
  - upstream model sync requests.
- Create/Edit account modals expose a scoped selector only for Anthropic API Key accounts and persist only non-default `authorization_bearer` in `extra`.
- Local modular i18n is updated under `frontend/src/i18n/locales/{zh,en}/admin/accounts.ts`; upstream monolithic `zh.ts` / `en.ts` are not introduced.
- No Ent, migration, wire, deploy, README, `.github`, knowledge, or unrelated product surface is modified.

## Allowed Paths
- `backend/internal/service/account.go`
- `backend/internal/service/anthropic_apikey_auth.go`
- `backend/internal/service/account_anthropic_passthrough_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/gateway_anthropic_apikey_passthrough_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/upstream_models_test.go`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `docs/workflow/tasks/upstream-main-v0143-anthropic-apikey-bearer-auth-s51.md`
- `docs/workflow/worker-results/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/cmd/server/wire_gen.go`
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/src/i18n/locales/zh.ts`
- `frontend/src/i18n/locales/en.ts`
- `deploy/**`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of `upstream/main` or the full release.
- Do not change Anthropic OAuth/setup-token auth behavior.
- Do not change billing, account selection, failover, model mapping, web-search emulation, passthrough body filtering, or cache TTL semantics.
- Preserve existing `anthropic_passthrough` and `web_search_emulation` extra fields.
- Keep default behavior backward-compatible by omitting `anthropic_apikey_auth_scheme` unless Bearer is selected.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestAccount_GetAnthropicAPIKeyAuthScheme|TestGatewayService_AnthropicAPIKeyPassthrough_BearerAuthScheme|TestBuildUpstreamModelsRequestsForAPIKeyAccounts" -count=1
cd ..
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
git diff --check -- backend/internal/service/account.go backend/internal/service/anthropic_apikey_auth.go backend/internal/service/account_anthropic_passthrough_test.go backend/internal/service/account_test_service.go backend/internal/service/gateway_anthropic_apikey_passthrough_test.go backend/internal/service/gateway_service.go backend/internal/service/upstream_models.go backend/internal/service/upstream_models_test.go frontend/src/components/account/CreateAccountModal.vue frontend/src/components/account/EditAccountModal.vue frontend/src/i18n/locales/zh/admin/accounts.ts frontend/src/i18n/locales/en/admin/accounts.ts docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0143-anthropic-apikey-bearer-auth-s51.md docs/workflow/worker-results/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-result.md docs/workflow/qa-reports/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-qa.md
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|frontend/src/i18n/locales/(zh|en)\.ts|deploy/|knowledge/|\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0143-anthropic-apikey-bearer-auth-s51`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if backend support requires Ent schema, migration, generated wire, or account DTO changes beyond existing `extra`.
- Stop if frontend support requires replacing the local modular i18n structure with upstream monolithic locale files.
- Stop if staged diff includes denied paths.

## Review Result
- Reviewed at: 2026-07-03 17:58 +08:00.
- Verdict: approved.
- Reason: the upstream change is a narrow compatibility toggle for Anthropic-compatible upstreams such as Ollama Cloud, defaults remain backward-compatible, and the scope can be verified with targeted service tests plus frontend typecheck.
