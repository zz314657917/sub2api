### DONE: upstream-cn-provider-correctness-s229-a

## Changed Files

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_count_tokens.go`
- `backend/internal/handler/openai_gateway_cn_dispatch_test.go`
- `backend/internal/service/openai_messages_dispatch.go`
- `backend/internal/service/openai_messages_dispatch_test.go`
- `backend/internal/service/openai_gateway_count_tokens.go`
- `backend/internal/service/openai_gateway_count_tokens_test.go`

## Manual Adaptation

- The local handler has direct `AllowMessagesDispatch` checks rather than the upstream helper owner. Added one nil-safe helper and reused it in both Messages and count_tokens handlers; CN providers and Grok bypass the OpenAI-only switch while OpenAI remains controlled.
- The local group dispatch resolver did not have the upstream Grok branch. Added only the CN early return, leaving existing OpenAI behavior unchanged.
- The local count_tokens service already had a deterministic CN estimator, but its `IsAnthropicProtocol` branch ran first. CN is now excluded from that branch so chat_completions, responses, and anthropic protocols all use the local path; the native helper remains untouched for non-CN Anthropic accounts.

## Commands

- `go test ./internal/handler -run "TestAllowOpenAICompatibleMessagesDispatch_CNProvidersExempt" -count=10` - PASS.
- `go test ./internal/service -run "TestResolveMessagesDispatchModel_CNProvidersBypassOpenAIDefaults" -count=10` - PASS.
- `go test ./internal/service -run "TestOpenAIGatewayService_ForwardCountTokensAsAnthropic_CNProviderAllProtocolsUseLocalEstimate" -count=10` - PASS.
- `go test ./internal/handler ./internal/service -count=1` - PASS.
- `go test ./cmd/server -run "^$" -count=1` - PASS.
- `gofmt -d`, `git diff --check`, exact allowlist, conflict/index, and upstream ancestry checks - PASS.

## Risks

- No real provider or network request was used; the zero-egress behavior is proven by the `httpUpstreamRecorder` test.
- Billing candidate filtering, 403 policy, partial-result usage submission, and response-stream drain/finalize remain separate deferred slices from `10c8b7020`.

## Knowledge Candidates

- None.
