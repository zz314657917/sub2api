### PASS: upstream-main-v0137-safe-patches-s15

## Executed Checks

- `go test -tags=unit ./internal/service -run "TestGetFallbackPricing_FamilyMatching|TestGetModelPricing_DoubaoEmbeddingVisionImageInputRate|TestCalculateCost_DoubaoEmbeddingVisionDifferentialInput|TestHandleNonStreamingResponse|TestHandleStreamingResponse_SSEErrorEvent|TestIsNonRetryableRefreshError|TestResolveThinkingProtocol|TestThinkingFilters|TestNormalizeChineseLLMThinking|TestApplyThinkingEnabledFallback|TestExtractOpenAIReasoningEffortFromBody" -count=1` passed.
- `go test ./internal/service -run "TestOpenAIGatewayServiceRecordUsage_MissingPricingRecordsZeroCostUsageLog|TestExtractOpenAIReasoningEffortFromBody|TestIsNonRetryableRefreshError|TestResolveThinkingProtocol|TestThinkingFilters|TestNormalizeChineseLLMThinking|TestApplyThinkingEnabledFallback" -count=1` passed.
- `go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback" -count=1` passed.
- `go test ./internal/repository -run "Test.*Decompress|Test.*HTTPUpstream" -count=1` passed.
- `go test ./internal/pkg/apicompat -count=1` passed.
- `git diff --check` passed.
- Lockfile scan for `form-data@4.0.5` / `form-data: 4.0.5` returned no matches.

## Failed Or Blocked Checks

- `npm.cmd run test:run -- --runInBand` failed because Vitest 2.1.9 does not support Jest's `--runInBand` option.
- Replacement command `npm.cmd run test:run -- --pool=threads --poolOptions.threads.singleThread=true` ran, but the existing full frontend suite failed in unrelated product areas:
  - `ChatImageStudioView.spec.ts`
  - `CanvasView.spec.ts`
  - `ChatStudioView.spec.ts`
  - `navigation.spec.ts`
  - `ImageCreatorView.spec.ts`
  - `useTableLoader.spec.ts`
  - `UsersView.spec.ts`
  - `AirwallexPaymentView.spec.ts`
  - `StudioBridgeSessionProbeView.spec.ts`
- These failures are outside the changed frontend dependency metadata and overlap with paths denied by this Sprint contract, so they were recorded as unverified existing frontend-suite risk rather than fixed here.

## Diff Review

- Denied paths were not touched.
- New thinking protocol filtering is gated on mapped upstream model families:
  - Anthropic strict models keep the existing signature cleanup behavior.
  - DeepSeek/Kimi/GLM/Moonshot/MiniMax M/Qwen thinking variants preserve thinking blocks.
  - Unknown model ids are conservative when an explicit model is provided.
- `frontend/pnpm-lock.yaml` now resolves axios/jsdom transitive `form-data` to `4.0.6`.

## Residual Risk

- Full frontend suite still needs a separate frontend stabilization pass.
- Migration-heavy upstream changes, compliance gates, OpenAI quota UI, channel monitor jitter, and Claude OAuth system prompt blocks were intentionally skipped.

## Recommendation

PASS for the contracted safe-patch scope. Do not treat this as a full upstream merge.
