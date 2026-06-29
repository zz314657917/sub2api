### DONE: upstream-main-v0139-backend-compat-s24

# Worker Result

## Summary

- Ported `82553c4dc` as a local adaptation: OpenAI handlers now capture request-time `QuotaPlatform` before async usage recording, and `OpenAIRecordUsageInput` carries it into `postUsageBillingParams` / `UsageBillingCommand`.
- Ported `7c2fee6c9`: fallback pricing warn logs are deduplicated per normalized model via `sync.Map`.
- Ported `da810c3b4`: quota-exhausted API keys reactivate when quota is changed to unlimited.

## Changed Files

- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_embeddings.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/handler/openai_quota_platform_contract_test.go`
- `backend/internal/service/openai_quota_platform.go`
- `backend/internal/service/openai_quota_platform_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/usage_billing.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_service_quota_test.go`
- `docs/workflow/tasks/upstream-main-v0139-backend-compat-s24.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Notes

- Local repository does not include upstream's full `UserPlatformQuotaRepository` chain. This Sprint preserves the quota platform in the local unified billing command and tests that propagation; it deliberately does not introduce upstream platform-quota persistence, migrations, or flusher infrastructure.
- `b105cc0fd` Codex JSON/developer-input behavior remains skipped for S25 because local transform semantics need separate review.

## Commands Run

```powershell
cd F:/mcplugins/sub2api/backend
go test -tags=unit ./internal/service -run "TestGetModelPricing_FallbackWarn|TestGetModelPricing_GLM52|TestAPIKeyService_Update_ReactivatesQuotaExhaustedWhenQuotaUnlimited|TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform" -count=1
go test ./internal/handler -run "TestOpenAIRecordUsageInputsCarryQuotaPlatform" -count=1
go test ./internal/service -run "TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform" -count=1

cd F:/mcplugins/sub2api
git diff --check
```

## Risks

- Platform field propagation is covered at handler AST and service command levels, not against a real platform-quota database path, because that upstream infrastructure is not present locally.
