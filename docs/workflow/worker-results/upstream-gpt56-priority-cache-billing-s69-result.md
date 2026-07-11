### DONE: upstream-gpt56-priority-cache-billing-s69

## Worker Deviation

- Agent Matrix 指定的 `deepseek-v4-pro` 外部 worker 返回 model 404。
- 经主控确认用户已明确授权多智能体后，本次改用当前可用协作 agent 作为 Generator fallback；实现范围、contract 和最终 Evaluator 归属均未改变。

## Summary

- 为 `ModelPricing` 和 LiteLLM 定价结构补充 GPT-5.6 Priority cache-write 专用价格，并贯通 embedded JSON、动态解析和 billing 映射。
- Priority 先选择专用 cache-write 价，再应用既有长上下文输入倍率；Standard 和 Flex 继续从 Standard cache-write 价计算。
- channel/interval `CacheWritePrice` 同时覆盖 Standard/Priority 且保留显式零；5m/1h breakdown 继续只读取原 breakdown 字段。
- 新增 contract 指定的 6 个精确 untagged 测试，覆盖 Sol/Terra/Luna、272k exclusive 边界、override、显式零、breakdown、解析和真实 `RecordUsage` 持久化路径。

## Changed Files

- `backend/internal/service/billing_service.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/pricing_service.go`
- `backend/internal/service/pricing_service_test.go`
- `backend/internal/service/gpt56_priority_cache_billing_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/resources/model-pricing/model_prices_and_context_window.json`
- `docs/workflow/worker-results/upstream-gpt56-priority-cache-billing-s69-result.md`

## Commands Run

- Required test discovery: exactly 6 tests found, PASS.
- Required six-test execution: PASS (`go test ./internal/service -run <requiredPattern> -count=1`).
- Service pricing and RecordUsage regressions: PASS.
- WebSocket cache creation / usage regressions: PASS.
- Default service package compile: PASS.
- Full clean-worktree Acceptance Commands: PASS (`base=6283f1aa66b1a467c59fa8a99c9d7351b61d0dcc`, `changed=8`, `required tests=6`).
- `git diff --check`: PASS.

## Contract Compliance

- No denied path, schema, payment, subscription, frontend, deployment, alias/catalog, usage extraction, or repository change.
- Standard/Flex pricing, the exclusive `> 272000` boundary, explicit-zero override, and cache breakdown semantics remain covered by assertions.

## Risks

- No live upstream billing request was issued; evidence is deterministic service and RecordUsage tests.
- Unit-tag suites were intentionally excluded per contract because of unrelated existing compile drift.
