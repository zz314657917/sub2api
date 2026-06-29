### PASS: upstream-main-v0139-openai-count-tokens-s26

# QA Report

## Findings

- PASS: OpenAI count_tokens service converts Anthropic messages payloads to Responses input_tokens payloads and returns `input_tokens`.
- PASS: API key account tests confirm custom `base_url` builds `/v1/responses/input_tokens` and sends the API key bearer token.
- PASS: OAuth unsupported endpoint tests map upstream 401 to Anthropic-format 404 without leaking ChatGPT account headers.
- PASS: route tests confirm OpenAI count_tokens path is registered and non-OpenAI count_tokens remains registered.
- PASS: related service, handler, and routes test subsets pass.
- PASS: `git diff --check` returned no whitespace errors; only existing `knowledge/*` CRLF warnings were printed.

## Executed Checks

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

## Unverified Risks

- No real OpenAI upstream smoke was run.
- Account scheduler and billing eligibility were compile-checked through handler package tests, but no integrated live account selection scenario was run.

## Recommendation

PASS for S26 as a scoped backend bridge. Commit only S26 allowed paths and keep existing `knowledge/*` dirty files unstaged.
