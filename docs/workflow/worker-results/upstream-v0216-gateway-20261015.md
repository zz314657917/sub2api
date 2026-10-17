DONE

Developer 实现完成，等待独立 QA；本报告不替代 QA PASS。候选：`E:/codex-worktrees/sub2api/upstream-v0216-gateway-20261015`；base/HEAD：`de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`。

## Changed paths

仅候选树以下六路径（主树仅写本报告）：

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service_test.go`
- `backend/internal/service/gemini_cached_content_not_found_test.go`（新增，default/unit 均执行）
- `backend/internal/service/gemini_error_policy_test.go`

实现：nil/non-OpenAI 在账号/渠道/全局开关前拒绝自动 image bridge，显式用户工具保留；tool-result 支持 base64/URL，按源顺序保留文字/图片、工具名称身份、is_error 及多个结果；native 缺缓存仅匹配结构化 403/PERMISSION_DENIED 的精确消息，支持 OAuth response wrapper/单元素数组，早于 policy/retry/账号写入。响应复用现有 helper，按现有白名单保留头和 request id；OAuth wrapper 按现有约定解包。403 探测上限 2 MiB，readErr/达到上限不确认，最终仅依赖循环 confirmed bool；完整 ForwardNative 大 body/读取失败负例已验证。

## 实际验收

命令均在候选 `backend` 执行，测试添加 `-json` 计数；下列“顶层/全部”包含子测试 pass 事件，全部零 skip、退出码 0。日志根目录统一为 `E:/codex-runtime/upstream-v0216-20261015/`，以下均为该目录下文件。

| 检查 | 顶层/全部 | 日志 |
| --- | --- | --- |
| canonical default 专项 | 66/572 | `gateway-dev-canonical-default-r2.jsonl` |
| apicompat/openai 全包 | 262/387 | `gateway-dev-canonical-packages.jsonl` |
| unit 专项（含两个精确名称） | 11/114 | `gateway-dev-canonical-unit.jsonl` |
| 两个指定 unit 精确运行 | 2/8 | `gateway-dev-exact-unit.jsonl` |
| 新增 9 顶层用例精确匹配，各重复 3 次 | 27/393 | `gateway-dev-new-exact-repeat.jsonl` |
| `go build -p 1 ./...` | PASS | `gateway-dev-canonical-build.log` |
| 六路径 `gofmt -l`、`git diff --check -- <六路径>` | 无输出/0 | `gateway-dev-format-diff.json` |

```powershell
go test -p 1 -json ./internal/service -run 'TestOpenAIGatewayService.*Codex|TestGemini.*(Tool|Image|Cache|Error)|TestConvertClaude|TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestCalculateCostUnified_ModelLongContext' -count=1
go test -p 1 -json ./internal/pkg/apicompat ./internal/pkg/openai -count=1
go test -p 1 -json -tags unit ./internal/service -run 'TestShouldFailoverGeminiUpstreamError|TestGemini.*ErrorPolicy|TestGemini.*CachedContent|^TestCheckErrorPolicy_GeminiAccounts$|^TestHandleGeminiUpstreamError_GoogleOneCapacityExhaustedUsesTierCooldown$' -count=1
go test -p 1 -json -tags unit ./internal/service -run '^TestCheckErrorPolicy_GeminiAccounts$|^TestHandleGeminiUpstreamError_GoogleOneCapacityExhaustedUsesTierCooldown$' -count=1
go test -p 1 -json -tags unit ./internal/service -run '^TestOpenAIGatewayServiceForward_CodexBridgeNonOpenAI$|^TestOpenAIGatewayService_CodexBridgeRejectsNilAndNonOpenAI$|^TestConvertClaudeMessagesToGeminiGenerateContent_ToolResultImagesOrdered$|^TestGeminiForwardNative_CachedContentNotFoundRequestError$|^TestGeminiCachedContentNotFoundRecognition$|^TestGeminiCachedContentNotFoundBoundedProbe$|^TestGeminiForwardNative_OtherErrorPolicyUnchanged$|^TestGeminiForwardNative_CachedContentProbeFailureKeepsErrorPolicy$|^TestGeminiErrorPolicy_CachedContentProbePrecedesPolicy$' -count=3
go build -p 1 ./...
```

首轮失败保留：`gateway-dev-new-default.jsonl`、`gateway-dev-canonical-default.jsonl`。新增 fixture 误设本地支持上游 MiniMax 平台/API，先发生新增测试编译错误，适配后 MiniMax Forward 仍报 `api_key not found in credentials`；完整请求矩阵改用本地 Kimi/DeepSeek/Zhipu，MiniMax 字符串仍在 helper 拒绝矩阵覆盖。没有修改生产平台支持或弱化既有断言。最终 r2 与精确重复运行均通过。

## 边界与风险

- `gateway-dev-scope-audit.json`：六路径最终 SHA-256；候选非授权文件零漂移、六 prerequisite 哈希匹配 seal、主树全文件零漂移（写报告前）、两树 HEAD/index 未变。`gateway-dev-test-counts.json` 记录实际计数和九用例逐个执行次数。
- 生产只改两个 service owner，保留本地两返回值 policy API，handler/依赖/既有 GPT61 owner/前置六测试未改。无 overlay、隐藏测试、改 tag、stage/commit/push/deploy、子代理、真实 provider 或数据库访问。
- 主控告知 main 默认 tag 仍有既有 `gateway_beta_test.go:138` 缺 `newTestGatewayServiceForBeta` 编译失败，本任务未复跑或修改 main；上述通过证据全部来自隔离候选。
- 证据限 mock 转发/转换测试和完整编译；真实 provider 接受性未验证。下一步：独立 QA 核验六路径、前置 seal 及 canonical 验收。
