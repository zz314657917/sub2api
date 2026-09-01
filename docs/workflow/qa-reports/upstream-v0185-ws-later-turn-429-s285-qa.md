### PASS: upstream-v0185-ws-later-turn-429-s285

## Findings

- Generator 交付物已复核：`docs/workflow/worker-results/upstream-v0185-ws-later-turn-429-s285-result.md` 首行为 `### DONE: upstream-v0185-ws-later-turn-429-s285`，其实现范围、保护基线、聚焦测试和 unit-tag 基线漂移说明与当前工作区及本次独立检查一致。
- 当前源码的关键换号边界已人工复核：`openai_gateway_handler.go` 用 `wsAttemptMessage` 作为每次渠道模型映射的输入，而非固定的 `firstMessage`。因此 HTTP bridge 后续 turn 的 retry payload（累计 `input`、移除 `previous_response_id`、恢复客户端模型）在新账号再次映射时不会被首轮 payload 覆盖。
- 未发现当前已实现代码的明确行为错误：HTTP 429 和写出前的 SSE `error` / `response.failed` 429 均在 HTTP bridge 的后续 turn 返回可 failover 的错误；已写出下游数据、首轮、非 429 和直接 WS 路径保持既有分支。工具输出上下文不完整时不生成 retry payload，handler 关闭连接而不换号重放。

## Executed Checks

- 静态 diff / 路径审查：S285 代码及新增测试只落在 contract 的 handler/service allowlist；其他工作区脏文件为既有受保护路径、workflow/user state 与 `outputs/`，本轮未修改。
- 人工审查 `backend/internal/handler/openai_gateway_handler.go`：retry payload 经 `openAIWSNextAttemptMessage` 复制后写入 `wsAttemptMessage`；渠道映射调用为 `ReplaceModelInBody(wsAttemptMessage, channelMappingWS.MappedModel)`。
- 人工审查 `backend/internal/service/openai_ws_forwarder.go`：仅 HTTP bridge 的 `turn > 1` 的 `UpstreamFailoverError` 包装 current-turn payload；payload 合并历史/当前 input，删除 `previous_response_id`，恢复 `originalModel`，并对工具输出上下文执行 fail-close 覆盖校验和副本隔离。
- 人工审查 `backend/internal/service/openai_ws_http_bridge.go`：后续 HTTP 429、写出前 SSE `error` 与 `response.failed` 429 走 failover；首轮、非 429、已写出下游事件与直接 WS 不进入该分支。
- `go test ./internal/handler -run '^TestOpenAIWSNextAttemptMessage' -count=10` -> PASS。
- `go test ./internal/service -run 'Test(OpenAIWSCurrentTurn|OpenAIWSHTTPBridge.*LaterTurn|OpenAIWS.*429|S284|S283)' -count=10` -> PASS。
- `go test ./internal/service` -> PASS。
- `go test ./internal/handler` -> PASS。
- `go test ./cmd/server -run '^$' -count=1` -> PASS。
- `gofmt -d` on all seven S285 Go source/test files -> PASS (no output)。
- `git diff --check -- backend/internal/handler backend/internal/service` -> PASS；`git diff --name-only --diff-filter=U` -> PASS (no unmerged paths)。
- `git diff --no-ext-diff --binary -- <six protected files> | git hash-object --stdin` -> `0e467987fd7aec5fc451983bdb8f8216f97ba69c`，与 contract 一致。
- `go test -tags unit ./internal/service` -> FAIL（既有 unit-tag 测试/API 漂移）：`stringPtr` 重复定义、旧 `computeTokenBreakdown` / `calculateCostInternal` / `buildCountTokensRequest` 签名及过期 `FallbackMode` 字段；未触及 S285 allowlist，未修改相关测试或源码。

## Unverified Risks

- 未进行真实 provider、数据库、容器或浏览器/WebSocket 端到端调用，符合 contract 排除范围。
- 完整 `unit` tag service 套件因上述既有编译漂移无法作为通过证据；默认 tag 的完整 service/handler 与 server 编译已通过。
- 当前新增单元测试覆盖 HTTP 429、SSE `error` 429、payload 重建/复制和孤立工具输出拒绝；未建立从 handler 的账号切换到第二账号映射的全链路测试，也未对 `response.failed` 和“已写出数据后 429”新增专属断言。静态分支审查未发现其语义偏移。

## Recommendation

允许进入下一 P/G/E 阶段；不应把 unit-tag 基线漂移归因于 S285。建议后续补一条 handler 级回归：触发后续 turn 429 后验证第二账号请求保留重建的 current-turn input，并使用第二账号的 mapped model。
