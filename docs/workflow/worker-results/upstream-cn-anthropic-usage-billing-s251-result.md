### DONE: upstream-cn-anthropic-usage-billing-s251

## Scope

- Business commit: `46185fcca0eef682b14c23cb27741072e45609a6` (`fix(billing): normalize CN Anthropic usage tokens`).
- Changed product/test owners only:
  - `backend/internal/pkg/apicompat/types.go`
  - `backend/internal/service/gateway_service.go`
  - `backend/internal/service/gateway_forward_as_responses.go`
  - `backend/internal/service/openai_gateway_messages_anthropic_native.go`
  - `backend/internal/service/kimi_anthropic_usage_test.go`
- No primary-worktree, provider, database, container, deployment, or push operation was performed.

## Upstream Mapping

- Behaviorally adapted upstream `695ebede70e0bed4c8fd4c87b5a426448a08ea4c` (`fix(billing): normalize CN Anthropic usage tokens`).
- The local generic parser remains in `gateway_service.go`; the S229 native parser reuses the shared normalizer from `gateway_service.go`; Responses DTO merging is normalized in `gateway_forward_as_responses.go`.
- `ClaudeUsage.InputTokens` now remains the uncached ordinary-input bucket, while `claudeUsageToOpenAIUsage` makes internal `OpenAIUsage.InputTokens` inclusive of cache-creation and cache-read buckets for the existing `RecordUsage` split.

## Acceptance Evidence

All commands ran in `E:/codex-worktrees/sub2api/upstream-cn-anthropic-usage-billing-s251`:

```powershell
Push-Location backend
go test ./internal/service -run "Test(ParseSSEUsagePassthroughNormalizesKimiPromptUsage|ParseSSEUsagePassthroughKimiFullyCachedInputReplacesStartTotal|ParseClaudeUsageFromResponseBodyNormalizesCNProviderAliases|ParseSSEUsagePassthroughNormalizesGLMAndDeepSeekAliases|MergeAnthropicUsageNormalizesKimiStreamForOpenAIBilling|MergeAnthropicUsageNormalizesGLMAndDeepSeekAliases|ClaudeUsageToOpenAIUsagePreservesCNProviderNativeAnthropicBuckets|CNProviderAnthropicUsageBillsUncachedInput)" -count=10
# PASS: ok github.com/Wei-Shaw/sub2api/internal/service 0.081s

go test ./internal/service -run "TestGatewayService_ParseSSEUsagePassthrough|TestParseClaudeUsageFromResponseBody|TestOpenAIGatewayServiceRecordUsage" -count=1
# PASS: ok github.com/Wei-Shaw/sub2api/internal/service 0.115s
go test ./internal/pkg/apicompat -run "TestAnthropicUsage" -count=1
# PASS: ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat 0.035s

go test ./internal/service -count=1
# PASS: ok github.com/Wei-Shaw/sub2api/internal/service 64.611s
go test ./cmd/server -run '^$' -count=1
# PASS: ok github.com/Wei-Shaw/sub2api/cmd/server 0.063s [no tests to run]
Pop-Location

gofmt -w backend/internal/pkg/apicompat/types.go backend/internal/service/gateway_service.go backend/internal/service/gateway_forward_as_responses.go backend/internal/service/openai_gateway_messages_anthropic_native.go backend/internal/service/kimi_anthropic_usage_test.go
git diff --check
# PASS
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" backend/internal/pkg/apicompat/types.go backend/internal/service/gateway_service.go backend/internal/service/gateway_forward_as_responses.go backend/internal/service/openai_gateway_messages_anthropic_native.go backend/internal/service/kimi_anthropic_usage_test.go
# PASS: no conflict markers (rg exit 1 handled as empty match)
git ls-files -u
# PASS: empty
```

## Risks

- Verification is local and fixture-based; no real provider traffic was authorized or performed.
- Independent QA still must rerun every contract gate in its separate worktree and write only its QA report.

## knowledge_candidates

- None. The result is task-specific behavior/provenance evidence, not a stable cross-repository rule.
