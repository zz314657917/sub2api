### DONE: upstream-v0185-ws-semantic-429-s284

# Worker Result

## Task ID

`upstream-v0185-ws-semantic-429-s284`

## Status

`done`

## Summary

- 完成上游 `571d1e1d9` 在本地统一 WebSocket forwarder owner 上的手工适配。
- 已建立连接后的语义 429 仅为 Spark OAuth 保留握手 quota headers；普通模型和非 OAuth 账号清空 headers 后进入既有错误处理。
- 握手 HTTP 429（`responseBody=nil`）继续保留响应 headers；补齐 `response.failed` 的嵌套 `response.error` 字段解析和 429 处理。
- 新增 S284 定向测试，覆盖普通模型、Spark OAuth、握手路径、`response.failed` 嵌套 429、API key 边界，并复用既有 S283 shadow 覆盖。

## Changed Files

- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_semantic_429_test.go`

## Commands Run

```text
go test ./internal/service -run 'Test(S284|OpenAIWSSemantic429|OpenAIWS.*RateLimit|S283|OpenAI.*Spark)' -count=10 -> PASS
go test ./internal/service -> PASS
go test ./cmd/server -run '^$' -count=1 -> PASS
gofmt -d internal/service/openai_ws_forwarder.go internal/service/openai_ws_semantic_429_test.go -> PASS
git diff --check -- backend/internal/service -> PASS
git diff --name-only --diff-filter=U -> PASS
```

## Test Output

```text
定向 service 回归 PASS（x10）；完整 service PASS；server compile PASS；gofmt 无输出。
```

## Risks

- 未执行真实 provider/WebSocket、数据库、scheduler 多进程、容器、部署或生产环境 smoke。
- `go test -tags unit ./internal/service` 仍受既有 fixture/API drift 影响，未修改或弱化该基线测试。

## Knowledge Candidates

- 无。

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- 无。
