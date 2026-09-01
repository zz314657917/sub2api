### PASS: upstream-v0185-ws-semantic-429-s284

# QA Report

## Task ID

`upstream-v0185-ws-semantic-429-s284`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0185-ws-semantic-429-s284.md`
- `docs/workflow/contract-reviews/upstream-v0185-ws-semantic-429-s284-review.md`

## Findings

- 未发现明确的实现问题。已建立 WebSocket 连接后的语义 429 仅在 Spark OAuth 模型保留握手 quota header；普通模型和非 OAuth 账号清空该 header，避免错误建立账号级 cooldown。
- 握手 HTTP 429 仍保留响应 header；`response.failed` 的嵌套 `response.error.status_code`、code/type/message 均能作为语义限流信号解析。
- S284 worker result `docs/workflow/worker-results/upstream-v0185-ws-semantic-429-s284-result.md` 已存在；其记录的变更范围、通过命令和未验证风险与本次独立实现复核及执行证据一致。

## Executed Checks

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`。S284 业务变更限于 `backend/internal/service/openai_ws_forwarder.go` 和新增的 `backend/internal/service/openai_ws_semantic_429_test.go`；QA 仅新增本报告。
- protected dirty baseline: `PASS`。六个受保护文件的 aggregate diff hash 保持 `0e467987fd7aec5fc451983bdb8f8216f97ba69c`，各文件 SHA-256 与任务基线一致；无 staged paths。
- commands run:
```text
go test ./internal/service -run 'Test(S284|OpenAIWSSemantic429|OpenAIWS.*RateLimit|S283|OpenAI.*Spark)' -count=10 -> PASS (2.149s)
go test ./internal/service -> PASS (cached; 首次执行自然结束后终端桥未回传最终文本，随即复跑明确通过)
go test ./cmd/server -run '^$' -count=1 -> PASS
gofmt -d internal/service/openai_ws_forwarder.go internal/service/openai_ws_semantic_429_test.go -> PASS (no output)
git diff --check -- backend/internal/service -> PASS
git diff --name-only --diff-filter=U -> PASS (no conflicts)
go test -tags unit ./internal/service -> FAIL (existing baseline fixture/API drift; not weakened or changed)
```
- manual checks:
```text
普通模型 semantic 429 -> handshake global quota headers cleared before existing error handling; PASS
Spark OAuth semantic 429 -> headers retained and S283 model-scoped rate limit written; PASS
握手 HTTP 429 -> responseBody=nil preserves headers; PASS
response.failed nested status_code/code/type/message -> semantic rate-limit fields parsed; PASS
API key and Spark shadow boundary -> no semantic global-header account cooldown; PASS
```

## Unverified Risks

- 未执行真实 provider/WebSocket、数据库、scheduler 多进程、容器、部署或生产环境 smoke；这些外部运行时链路仍未验证。
- `go test -tags unit ./internal/service` 受既有测试/API 漂移阻断：`stringPtr` 重复定义、billing 函数签名及 `buildCountTokensRequest` 返回值不匹配、Proxy/UpdateProxyInput 旧字段缺失。该失败与 S284 变更无直接证据关联，但保留为基线风险。

## Recommendation

`PASS`。允许进入集成阶段；在可控环境执行一次真实 OAuth 普通模型、Spark OAuth 模型及握手 429 的 WebSocket smoke。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- 若后续修改 WebSocket 限流处理，重测普通模型语义 429、Spark OAuth 语义 429、握手 HTTP 429、`response.failed` 嵌套错误字段与 API key/Spark shadow 边界。

## Knowledge Promotion

- `none`
