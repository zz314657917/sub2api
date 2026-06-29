### PASS: upstream-main-v0139-backend-compat-s24

# QA Report

## Findings

- PASS: OpenAI handler usage input literals now include `QuotaPlatform`; AST contract test covers Responses, Messages, WebSocket, Chat Completions, Embeddings, and Images.
- PASS: OpenAI `RecordUsage` preserves explicit quota platform and falls back to API key group platform when unset.
- PASS: fallback pricing warn is deduplicated per normalized model while GLM fallback pricing remains unchanged.
- PASS: quota-exhausted API keys reactivate when quota is changed to unlimited.
- PASS: `git diff --check` returned no whitespace errors; only existing CRLF conversion warnings were printed.

## Executed Checks

```powershell
cd F:/mcplugins/sub2api/backend
go test -tags=unit ./internal/service -run "TestGetModelPricing_FallbackWarn|TestGetModelPricing_GLM52|TestAPIKeyService_Update_ReactivatesQuotaExhaustedWhenQuotaUnlimited|TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform" -count=1
go test ./internal/handler -run "TestOpenAIRecordUsageInputsCarryQuotaPlatform" -count=1
go test ./internal/service -run "TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform" -count=1

cd F:/mcplugins/sub2api
git diff --check
```

## Unverified Risks

- No real OpenAI/OAuth upstream or real platform-quota persistence smoke was run.
- Local codebase lacks upstream's full `UserPlatformQuotaRepository` infrastructure; S24 intentionally did not add migrations or flusher behavior.

## Recommendation

PASS for S24 as a scoped backend adaptation. Commit only S24 allowed paths and keep existing `knowledge/*` dirty files unstaged.
