### PASS: upstream-main-v0143-anthropic-apikey-bearer-auth-s51

# Worker Result

## Summary
- Ported upstream `7869b7fe3` as a scoped Anthropic API Key auth scheme option.
- Added `GetAnthropicAPIKeyAuthScheme` and `setAnthropicAPIKeyAuthHeader`; default and invalid values continue to use `x-api-key`, while `authorization_bearer` sends `Authorization: Bearer`.
- Applied the scheme to Anthropic API-key messages, count_tokens, passthrough messages, passthrough count_tokens, account connection tests, and upstream models sync.
- Added Create/Edit account modal selector for Anthropic API Key accounts and persisted only non-default Bearer mode in `extra.anthropic_apikey_auth_scheme`.
- Updated local modular i18n files only; did not touch upstream monolithic locale files.

## Changed Files
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
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-v0143-anthropic-apikey-bearer-auth-s51.md`
- `docs/workflow/worker-results/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-anthropic-apikey-bearer-auth-s51-qa.md`

## Commands Run
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestAccount_GetAnthropicAPIKeyAuthScheme|TestGatewayService_AnthropicAPIKeyPassthrough_BearerAuthScheme|TestGatewayService_AnthropicAPIKeyBearerAuthScheme|TestBuildUpstreamModelsRequestsForAPIKeyAccounts" -count=1
go test ./internal/service -run "Test.*Anthropic.*APIKey.*Auth|Test.*AnthropicAPIKey.*Bearer|Test.*UpstreamModels.*APIKey|TestBuildUpstreamModelsRequestsForAPIKeyAccounts|TestAccount_GetAnthropicAPIKeyAuthScheme" -count=1
cd ..
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
```

## Risks
- No live Ollama Cloud or Anthropic-compatible upstream request was executed; validation is code-level request construction and frontend typecheck.
- UI visual screenshot was not required for this scoped settings selector.

## Contract Compliance
- No Ent, migrations, generated wire, deploy, README, `.github`, knowledge, or monolithic locale files were modified.
- Default behavior remains backward-compatible because the new extra field is only written for Bearer mode.
