### PASS: upstream-v027-deepseek

# 独立 QA 报告

## Findings

未发现合同范围内的明确回归或越界修改。相对 `d6e6c3471`，已有文件的
DeepSeek 业务差异仅位于
`backend/internal/service/openai_gateway_service.go` 的
`normalizeDeepSeekResponsesRequestBody`：在原有 `store=false` 和删除
`previous_response_id` 后，以 `json.Decoder.UseNumber` 解码，仅在检测到图片
tool output 时重建请求体。三个新增文件均在合同 allowlist 内。

媒体 helper 先收集连续 tool-output 批次，再将媒体消息置于完整输出批次之后，
最后恢复 developer/system 通知；不会跨 user 消息合并批次。无图片的 plain output
返回原输入，已提升的输出再次执行也不再改变。`UseNumber` 加上现有
`marshalOpenAIUpstreamJSON` 保留 `9007199254740993` 这类大整数，不会转为浮点。

## Executed Checks

- `backend`: `go test ./internal/pkg/apicompat -count=1`，通过。
- `backend`: `go test ./internal/service -run '^TestV027DeepSeek' -count=1 -v`，通过；实际启动 localhost `httptest` 上游并验证请求体：两个 function-call output 连续、两张图片进入随后的 user message、developer/system 通知后移、`store=false`、删除 `previous_response_id`、大整数原样转发。
- `backend`: 对四个合同 Go 文件执行 `gofmt -d`，无输出。
- 相对 `d6e6c3471` 的 scoped `git diff --check` 通过；未合并路径检查为空。
- 当前工作树的 `go build ./...` 因并行 Gemini 改动失败：`gemini_v1beta_handler.go` 引用尚未存在的 `Group.ModelAllowlist*` 和 `filterUpstreamGeminiModelsBody`。按合同在任务专用临时 worktree 检出 `d6e6c3471` 后，用同一 `go build ./...` 成功；临时 worktree 已清理。因此该全量构建失败不归因于本任务，且不构成 DeepSeek scoped QA 的失败证据。

## Scope And Limits

未修改任何业务源文件。本报告仅覆盖 mock localhost 转发、代码差异和构建级证据；真实 DeepSeek 账户、数据库、容器、部署与付费 provider 未验证。当前共享 worktree 的全量构建须待 Gemini 并行任务完成后再整体复跑。
