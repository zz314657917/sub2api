# Task Contract

## Task ID
upstream-cn-anthropic-usage-billing-s251

## Role
Developer Worker 在隔离 worktree 中按本 contract 实现；Codex Controller 负责 contract review、范围审计和最终裁决。独立 QA 必须使用单独 worktree，且不得修改业务文件。

## Goal
行为级移植上游 `695ebede7`（合入 `b8651947c`）的 CN Anthropic-compatible 用量归一化，修正 Kimi、GLM、DeepSeek 缓存明细下的普通输入 token 与缓存 token 重复扣减。不得合并已长期分叉的上游历史。

## Success Criteria
- 对提供 `prompt_tokens`、`cached_tokens`、`prompt_tokens_details.cached_tokens`、`prompt_cache_hit_tokens` 或 `prompt_cache_miss_tokens` 的 Anthropic-compatible usage，`ClaudeUsage.InputTokens` 始终表示不含缓存读写的普通输入；缓存读/写分别保留在互斥字段中。
- Kimi 流式 `message_start` 的总输入可被 `message_delta` 中的未缓存输入（包括显式 `0`）正确覆盖，不会再从该未缓存值二次扣减缓存命中。
- GLM 的嵌套缓存明细和 DeepSeek 的 hit/miss 明细均得到相同的互斥计费桶；没有这些扩展字段的原生 Anthropic 响应行为不变。
- 供 OpenAI Chat Completions、Responses、原生 Messages 网关共用的内部 `OpenAIUsage.InputTokens` 保持“普通输入 + cache creation + cache read”的总输入语义，令既有 `RecordUsage` 恰好拆回一次普通输入。
- 新的 default-tag 回归同时覆盖通用 Anthropic 解析、S229 国产原生流式路径、Anthropic SSE DTO 合并、OpenAI 总输入映射和基于本地模型定价的普通输入计费。

## Context
- Repo: `F:/mcplugins/sub2api`
- Base: `main@de360f464`
- Upstream source: `695ebede7` from `upstream/main@03e8ab413`; the source is an ancestor of current upstream.
- `git apply --check` proves direct cherry-pick is invalid: local `gateway_anthropic_passthrough.go` is represented by `gateway_service.go`, and local S229 native paths plus current DTO topology require behavior-level adaptation.
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`, and this contract.

## Allowed Paths
- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/openai_gateway_messages_anthropic_native.go`
- `backend/internal/service/kimi_anthropic_usage_test.go`
- `docs/workflow/worker-results/upstream-cn-anthropic-usage-billing-s251-result.md`
- `docs/workflow/qa-reports/upstream-cn-anthropic-usage-billing-s251-qa.md`

## Denied Paths
- `backend/internal/service/openai_gateway_service.go` and all existing billing pricing/resolver owners.
- `backend/internal/service/openai_gateway_{chat_completions,responses}_anthropic_native.go` (they must receive the correction through the shared DTO merge/converter only).
- `frontend/**`, `knowledge/**`, `outputs/**`, current Pixel Cafe/Groups edits, schema/migrations, dependency files, provider configuration, containers, deployment, push, and every path not explicitly Allowed.

## Constraints
- Preserve the local S229 protocol/routing/timeout behavior. Do not reintroduce upstream gateway topology or copy its missing file paths.
- Reuse the existing `ClaudeUsage` and `OpenAIUsage` contracts: Claude ordinary input is exclusive of caches; OpenAI internal input is inclusive of cache details. Do not change `RecordUsage`.
- Use presence-aware handling for `prompt_cache_hit_tokens` and `prompt_cache_miss_tokens`, so explicit zero means zero rather than retaining a preceding `message_start` total.
- Prefer one shared normalizer for the generic and S229 native SSE parsers; do not duplicate diverging logic. `mergeAnthropicUsage` must normalize DTO-decoded data consistently.
- All product work occurs only in `E:/codex-worktrees/sub2api/upstream-cn-anthropic-usage-billing-s251`. The primary worktree's user-owned changes must remain byte-for-byte untouched.
- No real provider, shared/production database, container, deployment, or push operation is authorized.

## Acceptance Commands
```powershell
Push-Location backend
go test ./internal/service -run "Test(ParseSSEUsagePassthroughNormalizesKimiPromptUsage|ParseSSEUsagePassthroughKimiFullyCachedInputReplacesStartTotal|ParseClaudeUsageFromResponseBodyNormalizesCNProviderAliases|ParseSSEUsagePassthroughNormalizesGLMAndDeepSeekAliases|MergeAnthropicUsageNormalizesKimiStreamForOpenAIBilling|MergeAnthropicUsageNormalizesGLMAndDeepSeekAliases|ClaudeUsageToOpenAIUsagePreservesCNProviderNativeAnthropicBuckets|CNProviderAnthropicUsageBillsUncachedInput)" -count=10
go test ./internal/service -run "TestGatewayService_ParseSSEUsagePassthrough|TestParseClaudeUsageFromResponseBody|TestOpenAIGatewayServiceRecordUsage" -count=1
go test ./internal/pkg/apicompat -run "TestAnthropicUsage" -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
Pop-Location

gofmt -w backend/internal/pkg/apicompat/types.go backend/internal/service/gateway_service.go backend/internal/service/gateway_forward_as_responses.go backend/internal/service/openai_gateway_messages_anthropic_native.go backend/internal/service/kimi_anthropic_usage_test.go
git diff --check
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" backend/internal/pkg/apicompat/types.go backend/internal/service/gateway_service.go backend/internal/service/gateway_forward_as_responses.go backend/internal/service/openai_gateway_messages_anthropic_native.go backend/internal/service/kimi_anthropic_usage_test.go
git diff --name-only <base>..HEAD
git diff --cached --name-only
git diff --name-only
git ls-files -u
```

## Output
- One focused business commit plus a separate Developer result commit. The business commit may change only the five product/test owners listed above.
- Developer result first line: `### DONE: upstream-cn-anthropic-usage-billing-s251`, `### BLOCKED: ...`, or `### FAILED: ...`; list exact files, commands/results, upstream mapping, risks, and `knowledge_candidates`.
- Independent QA report first line: `### PASS: upstream-cn-anthropic-usage-billing-s251`, `### FAIL: ...`, or `### BLOCKED: ...`; it must rerun all contract gates independently and state that only its report was written.

## Stop Rules
- Stop if the fix requires `RecordUsage`, pricing, routing, schema/migration, dependency, provider configuration, or a denied native endpoint owner to change.
- Stop if any focused test fails, if a default-tag test cannot exercise the named behavior, or if a candidate changes the protected primary worktree.
- Stop on unmerged index, conflict markers, external state access, scope expansion, or absence of the required Worker/QA verdict.

## Budget
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
