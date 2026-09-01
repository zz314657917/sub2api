### DONE: upstream-v0184-channel-pricing-s278

# Worker Result

## Task ID
upstream-v0184-channel-pricing-s278

## Status
`done`

## Summary
- 审阅当前 `HEAD 43d109581` 已有的 literal-first 归一化渠道定价实现，未发现需要修改 resolver 的实现 bug，本轮保持 `model_pricing_resolver.go` 不变。
- 在 allowlist 测试文件中补齐两个合同明确负例：未知 OpenAI 变体 `gpt-5.6-luna-ultra` 不得命中 `gpt-5.6-luna`，非 OpenAI 模型 `claude-5.6-luna-high` 不得命中相似基名 `claude-5.6-luna`。
- 两个负例直接断言 `lookupChannelPricingNormalized` 返回 `nil`，不依赖可能漂移的官方价格；已有默认/订阅 `RecordUsage` 正例、effort/date 后缀、精确变体优先和无关模型回退均保留。
- 8 个 `TestChannelPricing_*` 用例连续执行 10 次通过；默认 tag 的完整 service 包和 `cmd/server` 编译通过，合同 Success Criteria 已满足。

## Changed Files
- `backend/internal/service/model_pricing_resolver_channel_normalized_test.go`
- `docs/workflow/worker-results/upstream-v0184-channel-pricing-s278-result.md`

## Commands Run
```text
gofmt -w internal/service/model_pricing_resolver.go internal/service/model_pricing_resolver_channel_normalized_test.go -> passed
go test ./internal/service -run '^TestChannelPricing_' -count=10 -> passed (8 cases per iteration, 5.487s)
go test ./internal/service -> passed (69.627s)
go test ./cmd/server -run '^$' -> passed (cached, no tests to run)
gofmt -d backend/internal/service/model_pricing_resolver.go backend/internal/service/model_pricing_resolver_channel_normalized_test.go -> passed (no diff)
git diff --check -- backend/internal/service/model_pricing_resolver.go backend/internal/service/model_pricing_resolver_channel_normalized_test.go -> passed
git diff --name-only --diff-filter=U -> passed (NO_CONFLICTS)
protected-path status/diff-hash audit -> passed; hash remained 0aac28b2bc965ad343ddba5efac6cf4e47dbb70d
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.487s
ok github.com/Wei-Shaw/sub2api/internal/service 69.627s
ok github.com/Wei-Shaw/sub2api/cmd/server (cached) [no tests to run]
NO_CONFLICTS
```

## Risks
- 本轮未重跑 `-tags unit` 全 service 包；此前已记录的既有测试/实现编译漂移不属于 S278，也不替代本轮全部默认-tag PASS 门禁。
- 未执行真实 provider、数据库、容器、部署、共享数据、commit、push 操作；这些均在合同范围外。

## Knowledge Candidates
- 渠道定价的归一化回退负例应直接验证 lookup/source，而非依赖官方兜底价数值，避免价格表更新造成测试漂移。

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- None.
