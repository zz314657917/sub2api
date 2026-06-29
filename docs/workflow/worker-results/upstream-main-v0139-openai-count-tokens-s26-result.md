### DONE: upstream-main-v0139-openai-count-tokens-s26

# Worker Result

## Summary

- Ported upstream `7a38c6621` as a local OpenAI count_tokens bridge.
- OpenAI groups now route `POST /v1/messages/count_tokens` to `OpenAIGatewayHandler.CountTokens` instead of returning local unsupported 404.
- Added `OpenAIGatewayService.ForwardCountTokensAsAnthropic`, converting Anthropic count_tokens requests to OpenAI `/v1/responses/input_tokens` and returning `{"input_tokens": n}`.
- Added custom-base-url handling for API key accounts and unsupported-endpoint fallback for OAuth accounts.

## Changed Files

- `backend/internal/handler/openai_gateway_count_tokens.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_test.go`
- `backend/internal/service/openai_endpoint_url.go`
- `backend/internal/service/openai_gateway_count_tokens.go`
- `backend/internal/service/openai_gateway_count_tokens_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `docs/workflow/tasks/upstream-main-v0139-openai-count-tokens-s26.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Notes

- Local adaptation differs from upstream where required: this codebase does not have upstream's `marshalOpenAIUpstreamJSON`, `WithHTTPUpstreamProfile`, `HTTPUpstreamProfileOpenAI`, `NewRequestBodyRef`, or `classifyNoAccountErrorFromGin` helpers.
- The local handler uses existing `ParseGatewayRequest`, billing eligibility signature, account selection signature, and existing Anthropic-format unavailable errors.
- No usage recording, frontend, migrations, wire generation, payment, Ops/Keys UI, or product settings were touched.

## Commands Run

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestOpenAIGatewayService_ForwardCountTokensAsAnthropic" -count=1
go test ./internal/server/routes -run "TestGatewayRoutesOpenAICountTokensPathIsRegistered|TestGatewayRoutesNonOpenAICountTokensPathStillRegistered" -count=1
go test ./internal/service -run "TestOpenAIGatewayService_ForwardCountTokensAsAnthropic|TestBuildOpenAIEndpointURL" -count=1
go test ./internal/handler -run "TestResolveOpenAIMessagesDispatchMappedModel|TestNewOpenAIModelMappedBodyCache|TestOpenAIGatewayHandler" -count=1
go test ./internal/server/routes -run "TestGatewayRoutes" -count=1

cd F:/mcplugins/sub2api
git diff --check
```

## Risks

- No live OpenAI `/v1/responses/input_tokens` request was sent; behavior is verified with local upstream recorder tests.
- The local route registration tests prove the path no longer returns route-level 404, but full handler runtime still depends on configured billing/account services in production.
