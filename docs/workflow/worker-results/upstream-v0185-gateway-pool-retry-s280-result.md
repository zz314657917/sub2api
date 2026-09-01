### DONE: upstream-v0185-gateway-pool-retry-s280

# Worker Result

## Task ID

`upstream-v0185-gateway-pool-retry-s280`

## Status

`done`

## Summary

- 用户明确授权主控接替连续失败的 Terra Developer 调度；主控按原 approved contract 完成最小实现，没有扩大业务范围。
- 两条 Anthropic 兼容转发路径都会把 `mappedModel` 传给 `HandleUpstreamError`，保存其 `shouldDisable` 结果，并仅对仍可调度且命中账号池重试状态码的池账号设置 `RetryableOnSameAccount`。
- 新增四个真实转发入口用例，覆盖 Chat/Responses 池模式 429 正例、非池账号和显式空重试状态码负例。

## Changed Files

- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/gateway_pool_mode_retry_test.go`
- `docs/workflow/worker-results/upstream-v0185-gateway-pool-retry-s280-result.md`

## Commands Run

```text
gofmt -w internal/service/gateway_forward_as_chat_completions.go internal/service/gateway_forward_as_responses.go internal/service/gateway_pool_mode_retry_test.go -> PASS
go test ./internal/service -run '^TestGatewayCompatPoolModeRetry$' -count=10 -> PASS, 5.471s
go test ./internal/service -> PASS, 66.903s
go test ./cmd/server -run '^$' -> PASS
git diff --check -- <three product/test paths> -> PASS
git diff --name-only --diff-filter=U -> empty
protected SHA-256 manifest and dirty diff hash check -> PASS, aggregate hash 0e467987fd7aec5fc451983bdb8f8216f97ba69c unchanged
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.471s
ok github.com/Wei-Shaw/sub2api/internal/service 66.903s
ok github.com/Wei-Shaw/sub2api/cmd/server (cached) [no tests to run]
```

## Risks

- 没有真实 Anthropic provider、真实多账号 scheduler 或部署 smoke；本轮只证明本地状态机返回值和转发错误语义。
- 独立 QA 尚未执行；本报告不是最终 PASS。

## Knowledge Candidates

- 兼容转发路径返回 failover error 时，必须把 rate-limit handler 的 disable 决策与账号池重试状态码一起合成，不能只调用 handler 丢弃结果。

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`（Terra 路由阻塞已由用户明确授权主控接替）

## Blocked Reason

- none
