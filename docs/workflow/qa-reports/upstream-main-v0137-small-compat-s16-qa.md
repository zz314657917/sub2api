### PASS: upstream-main-v0137-small-compat-s16

## Executed Checks

- `go test -tags=unit ./internal/service -run "TestParseGatewayRequest_ResponsesInput|TestGenerateSessionHash_ResponsesInputProducesHash|TestDecideResponsesProbeSupportRequiresFunctionCallOn2xx|TestOpenAIResponsesProbePayloadForcesFunctionCall|TestSelectResponsesProbeModelUsesMappedUpstreamModel|TestProbeOpenAIAPIKeyResponsesSupportPersistsToolCapability" -count=1` passed.
- `go test -tags=unit ./internal/handler -run "TestDetectInterceptType_MaxTokensOneHaiku|TestSendMockInterceptResponse_MaxTokensOneHaiku" -count=1` passed.
- `go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionIncludesClientIPForBlacklistDenial" -count=1` passed.
- `go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback|Test.*GenerateSessionHash|TestParseGatewayRequest" -count=1` passed.
- `go test ./internal/repository -run "Test.*Decompress|Test.*HTTPUpstream" -count=1` passed.
- `go test ./internal/pkg/apicompat -count=1` passed.
- `go test ./internal/service -run "TestOpenAIGatewayServiceRecordUsage_MissingPricingRecordsZeroCostUsageLog|TestExtractOpenAIReasoningEffortFromBody|TestIsNonRetryableRefreshError|TestResolveThinkingProtocol|TestThinkingFilters|TestNormalizeChineseLLMThinking|TestApplyThinkingEnabledFallback" -count=1` passed.
- `git diff --check` passed.
- Denied-path audit returned `NO_DENIED_PATHS`.
- Lockfile scan returned `NO_FORM_DATA_405`.

## Diff Review

- Responses sticky hash is scoped to `protocol=="responses"` and only falls back to `input` when system/messages content did not already anchor the hash.
- Haiku probe intercept still requires a Claude Code client signal; it now covers streaming probes as upstream intended.
- OpenAI APIKey `/responses` probe now avoids false positives from upstreams whose endpoint exists but tool calls do not work.
- ACL denial response now reports the IP used by the existing trusted-IP restriction check.

## Residual Risk

- Full frontend Vitest was not rerun in S16 because S15 already recorded unrelated product-area failures in Studio/Canvas/navigation/payment tests.
- Remaining upstream candidates still need separate contracts because they touch image generation behavior, repository batching, rate-limit windows, migrations, or product settings/UI.

## Recommendation

PASS for the contracted S16 small-compat scope. Do not treat this as a full upstream merge.
