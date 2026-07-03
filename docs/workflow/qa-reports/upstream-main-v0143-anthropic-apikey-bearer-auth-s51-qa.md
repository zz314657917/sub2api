### PASS: upstream-main-v0143-anthropic-apikey-bearer-auth-s51

# QA Report

## Findings
- No blocking findings.
- Backend auth scheme selection is constrained to Anthropic API Key accounts and defaults to `x-api-key` for missing, invalid, non-Anthropic, and non-API-key accounts.
- Bearer mode is covered for API-key passthrough messages/count_tokens, normal gateway messages/count_tokens, and upstream models sync.
- Frontend Create/Edit modals expose the selector only on Anthropic API Key accounts and write `anthropic_apikey_auth_scheme` only when Bearer is selected.

## Executed Checks
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestAccount_GetAnthropicAPIKeyAuthScheme|TestGatewayService_AnthropicAPIKeyPassthrough_BearerAuthScheme|TestGatewayService_AnthropicAPIKeyBearerAuthScheme|TestBuildUpstreamModelsRequestsForAPIKeyAccounts" -count=1
```

```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "Test.*Anthropic.*APIKey.*Auth|Test.*AnthropicAPIKey.*Bearer|Test.*UpstreamModels.*APIKey|TestBuildUpstreamModelsRequestsForAPIKeyAccounts|TestAccount_GetAnthropicAPIKeyAuthScheme" -count=1
```

```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
```

## Scope Review
- Backend touched only service request construction, helper logic, and focused tests.
- Frontend touched only the account Create/Edit modals and local modular admin account i18n files.
- Denied paths still require final staged audit before commit.

## Unverified Risks
- No live upstream compatibility smoke against Ollama Cloud or another Bearer-only Anthropic-compatible endpoint.
- No browser screenshot for the account modal selector.

## Recommendation
Ship S51 after final `git diff --check`, scoped staging, cached denied-path audit, and commit.
