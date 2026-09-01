### DONE: upstream-v0185-ws-later-turn-429-s285

## 实现摘要

- 在 `openai_ws_http_bridge.go` 中为 HTTP bridge 后续 turn 的 HTTP 429，以及尚未向客户端写出数据的 SSE `error`/`response.failed` 语义 429 返回可识别的 failover；保留首轮、直接上游 WS、非 429 和已写出数据路径行为。
- 在 `OpenAIForwardResult` 与 bridge collector 中保存完整输出项，用于后续 turn 的跨账号 current-turn replay；重放 payload 合并累计 input、移除旧 `previous_response_id`、恢复客户端模型名，并返回隔离副本。
- 对 `function_call_output` 缺少可验证工具调用上下文的情况不构造重放 payload，handler fail-close；避免丢失工具续链语义。
- 在 handler 中切换账号时使用 current-turn payload；即使存在渠道模型映射，也基于 `wsAttemptMessage` 映射，避免回退到首轮消息。
- 仅修改 contract allowlist 内的 handler/service 源码和新增回归测试；未触碰 repository、Ent、migration、frontend、部署或真实 provider/数据库/WebSocket。

## 验证

- `go test ./internal/handler -run '^TestOpenAIWSNextAttemptMessage' -count=10` PASS
- `go test ./internal/service -run 'Test(OpenAIWSCurrentTurn|OpenAIWSHTTPBridge.*LaterTurn|OpenAIWS.*429|S284|S283)' -count=10` PASS
- `go test ./internal/service` PASS（67.196s）
- `go test ./internal/handler` PASS
- `go test ./cmd/server -run '^$' -count=1` PASS
- `gofmt -d`（全部 S285 handler/service 文件）无输出，PASS
- `git diff --check -- backend/internal/handler backend/internal/service` PASS；`git diff --name-only --diff-filter=U` 无冲突
- 受保护六文件 aggregate dirty diff hash 保持 `0e467987fd7aec5fc451983bdb8f8216f97ba69c`
- `go test -tags unit ./internal/service` 未通过既有基线漂移：`stringPtr` 重复定义、旧 `computeTokenBreakdown`/`calculateCostInternal`/`buildCountTokensRequest` 签名、`FallbackMode` 字段缺失等；未修改这些无关测试或源码。

## 未验证风险

- 未执行真实 provider、真实 WebSocket、多账号 scheduler、数据库、容器、部署或浏览器运行态 smoke。
- 未新增 handler 全链路渠道映射重放自动化测试；已通过代码审查确认重放分支使用 `wsAttemptMessage`。

## 交付边界

- 保留工作区原有受保护 dirty changes 与 `outputs/**`，不 push、不 merge/rebase/cherry-pick。
