# GPT-5.6 缓存写入计费热修 Sprint：`gpt56-cache-write-billing-s63`

## Contract Status

- Task ID: `gpt56-cache-write-billing-s63`
- Phase: `contract-approved`
- Role: 主 Codex 负责 Planner、Generator 与 Final Evaluator；不调用 worker。
- Review verdict: `APPROVED WITH AMENDMENT`。Evaluator 补齐统一定价 resolver 必需路径及既有测试基线判定；范围仍仅覆盖 GPT-5.6 缓存写入用量解析、互斥计费和定向回归。

## Goal

- 兼容 OpenAI GPT-5.6 官方 `cache_write_tokens` 用量字段，并将其落入现有 `cache_creation_tokens` 账本。
- 将普通输入、缓存写入、缓存读取拆成互斥计费桶，避免少收或重复收费。

## Success Criteria

- HTTP JSON、SSE、raw Chat Completions、Messages bridge、legacy WS forwarder 和 WS v2 均能识别官方嵌套缓存写入字段及兼容别名。
- 官方嵌套字段显式为 `0` 时优先于旧兼容别名，避免陈旧非零值污染账单。
- 普通输入按 `input_tokens - cache_read_tokens - cache_write_tokens` 计算，且下限为 `0`。
- 渠道 flat/interval 显式配置缓存写入价为 `0` 时保持为零，不被 GPT-5.6 默认回退覆盖。
- 缓存写入 token 参与统一计费的上下文区间、按次档位和长上下文阈值选择。
- GPT-5.6 Sol/Terra/Luna 现有价格保持不变，不引入数据库迁移或新账本字段。
- 定向测试、`internal/service` 回归和 diff 边界检查通过。

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: 当前 `main`；目标业务路径在 contract 创建时无未提交改动。
- Upstream references: `4a2b10c94`, `383f61d0e`, `062af81fb`, `de28eba3c`。
- Existing ledger: `cache_creation_tokens` 继续承载 OpenAI `cache_write_tokens`。

## Allowed Paths

- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/pkg/apicompat/responses_to_chatcompletions.go`
- `backend/internal/service/openai_embeddings.go`
- `backend/internal/service/openai_embeddings_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_messages_usage_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_success_test.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay_internal_test.go`
- `docs/workflow/tasks/gpt56-cache-write-billing-s63.md`
- `docs/workflow/qa-reports/gpt56-cache-write-billing-s63-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- 其他未列入 Allowed Paths 的业务、配置和流程文件。

## Constraints

- 不合并整个 `upstream/main`，只按当前代码结构移植必要语义。
- 不覆盖或回滚当前工作树已有 staged/unstaged 改动。
- 官方嵌套字段优先级必须以“字段存在”为判断，不能只取第一个正数。
- 计费修复不得改变非 OpenAI/GPT-5.6 的既有价格或缓存语义。
- 不新增自动缓存策略；本 Sprint 只修用量识别与账本准确性。

## Acceptance Commands

```powershell
go test ./internal/pkg/apicompat -count=1
go test ./internal/service/openai_ws_v2 -count=1
go test ./internal/service -run "GPT56|CacheWrite|CacheCreation|ExtractOpenAIUsage|RecordUsage|ChatCompletions|Messages|Embeddings" -count=1
go test ./internal/service -count=1
go test ./internal/service -skip "PeakMultiplier" -count=1
git diff --check
git diff --name-only -- <allowed paths>
```

## Output

- 独立 QA 报告首行必须为 `### PASS: gpt56-cache-write-billing-s63`、`### FAIL: ...` 或 `### BLOCKED: ...`。
- 最终提交只包含 Allowed Paths，且不得吸收当前工作树其他改动。

## Stop Rules

- 需要数据库迁移、生产配置或未授权业务文件时停止并回 Planner。
- 上游语义与官方文档冲突时停止，不以猜测补字段。
- 定向测试或排除已确认既有 `PeakMultiplier` 基线失败后的 `internal/service` 回归失败时不得提交；完整包失败必须确认仅限既有失败并写入 QA 报告。
