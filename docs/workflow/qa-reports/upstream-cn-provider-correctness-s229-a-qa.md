### PASS: upstream-cn-provider-correctness-s229-a

## Scope

- QA base: `fb391fd08` (`docs(workflow): record S229-A controller result`).
- QA worktree had no product diff; only this report was added.
- Exact allowlist, diff check, conflict/index, upstream ancestry, and protected-main checks passed.

## Independent Verification

- `go test ./internal/handler -run "TestAllowOpenAICompatibleMessagesDispatch_CNProvidersExempt" -count=10` - PASS.
- `go test ./internal/service -run "TestResolveMessagesDispatchModel_CNProvidersBypassOpenAIDefaults" -count=10` - PASS.
- `go test ./internal/service -run "TestOpenAIGatewayService_ForwardCountTokensAsAnthropic_CNProviderAllProtocolsUseLocalEstimate" -count=10` - PASS.
- `go test ./internal/handler ./internal/service -count=1` - PASS.
- `go test ./cmd/server -run "^$" -count=1` - PASS.
- `git diff --check`, exact allowlist, conflict/index, and `10c8b7020` ancestry - PASS.

## Review

- CN and Grok bypass only the OpenAI-specific Messages switch; OpenAI false/true behavior remains covered.
- CN Kimi/Zhipu/DeepSeek dispatch mapping returns empty rather than GPT defaults; OpenAI default mapping remains covered.
- Kimi chat_completions, Zhipu anthropic, and DeepSeek responses count_tokens cases all returned a positive local estimate with zero recorded upstream requests.
- Main protection remained unchanged: `main@ff241be81`, `origin/main@a865d8b6e`, `upstream/main@8869775ed`; user dirty patch IDs, untracked tutorial migration/test hashes, and `outputs/` were preserved.

## Risks

- No real provider, database, deployment, container, or remote operation was used.
- The deferred billing, 403, partial-result usage, and response-stream drain/finalize slices remain unverified by this task.
