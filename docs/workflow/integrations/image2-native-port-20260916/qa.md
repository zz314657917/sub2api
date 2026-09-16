### PASS: image2-native-port

# Independent Terra QA

## Findings

初轮发现的两项阻断性流式计费/客户端错误语义缺陷已在复测解除；当前未发现范围越界业务改动。

- [P1] `backend/internal/service/openai_images_direct.go:219,262-269,275-302`：原生 SSE 的任意下游 write error 被写入 `processErr`，而 `process` 随即忽略所有后续 upstream data。读循环表面上继续到 EOF，但不再解析 completed 事件或 usage。最小复现序列为：客户端在 `image_generation.partial_image`（或 started）写出时断开，随后上游发送 `image_generation.completed` 和 usage；函数返回 `partialOnly`、停留在断开前的 usage/count（可能为 0），`RecordUsage` 随后会把已经完成的图片作为 preview/零结算，而不是保留实测 usage/完成数。既有 Responses handler 在 `openai_images_responses.go:1442-1456,1486,1544` 使用 `clientDisconnected` 继续解析/drain 以供 billing，原生实现与该既有契约不一致。
- [P1] `backend/internal/service/openai_images_direct.go:226-228,298-304`：native 的 `error`/`response.failed` SSE 只变为内部 `processErr`，没有向仍连接的客户端写 `event: error`。既有 Responses 在 `openai_images_responses.go:1549-1557` 显式转换并发送错误 SSE。partial 后 upstream error 的 native 调用者又在 `openai_images_responses.go:1929-1935` 清除已写输出错误，因此客户端只会得到 partial 而没有终止错误事件。

- 原生分流仅以 OAuth 账号的映射后模型为准；显式白名单、旧/未知快照回退 Responses、404/405 单次回退均有生产函数与 fake transport 覆盖。
- `openai_images_test.go` 的 11 处 `withOpenAIImagesForceResponses` 仅位于既有 OAuth/Responses 协议 fixture；其断言未削弱。APIKey 与 APIMart fixture/生产路径仍保持原样。
- 未把 fake transport / mock 结果当作真实 provider 验收。

## Success Criteria Evidence

1. PASS — `backend/internal/service/openai_images_direct.go:52-60` 明确列出 `gpt-image-1.5`、`gpt-image-2`、两种 2.5 模型与两个 `2026-09-08` alias；`openai_images_responses.go:1766-1823` 先完成账户映射、仅在映射结果属于白名单时选择原生 Images，否则保留配置的 Responses driver。`TestCodexDirectImagesLegacyAndUnknownSnapshotsKeepResponsesDriver` 覆盖 legacy/unknown；映射 JSON/SSE 覆盖在 `TestCodexDirectImagesForwardMappedModelJSONAndSSEOutputFormats`。
2. PASS — `openai_images_direct.go:63-123` 保留解析后的原生选项、`n`、stream，编辑请求将 uploads/mask 变为 data URL；`openai_images_responses.go:1870-1911` 只在原生 404/405 时以 force-Responses context 回退一次。`TestCodexDirectImagesForwardMultipartEditMappedModelAndSSE`、`TestCodexDirectImagesFallbackStopsAfterSingle404Or405`、`TestCodexDirectImagesRetryablePreOutputHTTPFailureIsDelegatedWithoutReplay` 验证这些路径及既有重试边界。
3. PASS（复测）— 非流式转换、尺寸与无输出拒绝正常；`openai_images_direct.go:214-330` 现在在 client disconnect 后继续解析 completed/usage，仍禁止 replay，并规范转发 upstream error SSE。生产 `ForwardImages` fake-transport 回归覆盖 started/partial write failure 后的 completed+cache usage。
4. PASS — `codexDirectImagesUsage`（`openai_images_direct.go:126-145`）对缓存图片 token 做上界拆分；`RecordUsage`（`openai_gateway_service.go:7703-7721,7867-7871`）传入既有账单计算并 clone caller breakdown 后记录 metadata。`TestOpenAIGatewayServiceRecordUsage_ImageCacheReadTokensPersistedAndDeduplicated` 与 subscription 变体使用 usage/balance/subscription fake，断言固定价格、JSON metadata、caller map 不变、同 request ID 仅实际结算一次；无缓存 token 和 nil/empty/populated map 也覆盖。
5. PASS — 变更分为 native helper、direct adapter、Responses dispatch/fallback、RecordUsage metadata/preview 账单与 fake-transport regression；未新增 schema、迁移或替代现有 owner 的 stub。
6. PASS（复测）— 新的 production `ForwardImages` fake-transport regression 覆盖 client write failure 后 completed/usage，及 native `error`/`response.failed` 到 `event: error`；原有 JSON、SSE、edit、404/405、pre-output failure、空输出、plan gate、尺寸、映射、usage/cost 覆盖仍通过。
7. PASS（复测）— JSON/SSE mapped/public model、URL/base64 和 edit prefix 断言继续通过；partial 后 no replay、native error event 和 client disconnect 后 completed usage 现均有直接回归。
8. PASS — `openai_gateway_record_usage_test.go` 的新 captured usage/billing fake 验证准确 cached-image 成本、`image_cache_read_tokens` 持久化 JSON、size entries、caller map 不变和 request-ID dedupe 的单次实际 charge；`TestOpenAIGatewayServiceRecordUsage_PreviewOnlyImagesIgnorePreflightCost` 及 first-resolved-token-candidate 回归验证 image preview 忽略 input 预估、但 token 模式按测得 token 计价，且不改变非 image override 行为。

## Executed Checks

- 从 `backend/`：`go test ./internal/service -run 'Test.*(OpenAIImages|CodexDirectImages|ImageOutput|ImageCache|RecordUsage)' -count=1` — PASS（0.164s）。
- 从 `backend/`：`go test ./internal/service -run 'TestOpenAIGatewayServiceForwardImages' -count=1` — PASS（0.071s）。
- 从 `backend/`：`go build ./...` — PASS。
- 从 `backend/`：`go test ./internal/service -count=1` — PASS；该通过不覆盖上述 client-disconnect/completed-after-write-error 场景。
- 从 worktree root：合同 PowerShell allowlist/format gate — PASS；8 个业务路径恰为 `openai_gateway_record_usage_test.go`、`openai_gateway_service.go`、`openai_images_responses.go`、`openai_images_test.go` 和 4 个未跟踪 direct/helper/native test 文件；无 denied business path。
- 从 worktree root：`gofmt -l`（所有基线后 Go 路径）— 无输出；`git diff --check bb3dde9e4` — PASS；`git ls-files -u` — 空。
- 手工 diff 审查：8 个业务文件、原生 dispatch/Responses fallback、账单 metadata/preview 路径，以及 OAuth fixture exact-path amendment。`docs/workflow/main-log.md`、`docs/workflow/status.md` 是 workflow 状态证据，不是本任务业务文件；已按合同列出 ignored task/review/result/report artifacts。

## Unverified Risks

- 本结论仅为本地源码、fake transport 与 repository/billing fake 验收；真实 OAuth Images provider、真实账户计划/冷却、数据库、容器、部署、主树集成、commit/push 均未执行且仍未验证。
- 未执行 `go test -tags unit ./...`；本合同并不要求它，且本轮无理由把未选中的仓库级 fixture 状态归因于该改动。
- 需要修复后至少增加两个 production-`ForwardImages` fake-transport 回归： (1) client write failure 后上游继续发送 completed+usage，断言仍计入 ImageCount/usage 并可供 RecordUsage 结算；(2) native `error`/`response.failed`（含 partial 后）向未断开的客户端发送规范 `event: error`，且不触发 replay。

## Recommendation

可进入最终本地集成裁决；不要将本 PASS 表述为真实 provider 或生产发布通过。

## Retest Update: native streaming drain and error signaling

初轮报告中的两项 P1 已在独立复测中解除，保留上文作为发现历史。

- `backend/internal/service/openai_images_direct.go:214-281` 将 client write failure 独立为 `clientWriteErr`/`clientDisconnected`，停止后续下游写入但不停止 upstream frame 的解析；完成帧的 ImageCount、output token、cache token 会继续进入结果。`TestCodexDirectImagesForwardDrainsCompletedUsageAfterClientDisconnect` 通过生产 `ForwardImages`，对 started/partial 首帧后断开、随后 completed+usage 两个场景断言该行为和单请求不重放。
- `backend/internal/service/openai_images_direct.go:239-249` 对 native `error`、`response.failed` 与 `response.incomplete` 向仍连接客户端写规范 `event: error`。`TestCodexDirectImagesForwardWritesErrorEventForNativeFailures` 使用 partial 前帧覆盖 error/response.failed；原 `TestCodexDirectImagesStreamingPartialClientWritePreventsReadFailover` 仍断言 client write error 优先、不会被后续 read failure 变为可重试错误。
- 复测命令：`go test ./internal/service -run 'Test.*(OpenAIImages|CodexDirectImages|ImageOutput|ImageCache|RecordUsage)' -count=1` PASS（1.310s）；`go test ./internal/service -run 'TestOpenAIGatewayServiceForwardImages' -count=1` PASS（0.082s）；`go build ./...` PASS；`go test ./internal/service -count=1` PASS。
- 复测路径门禁：合同 allowlist/gofmt gate PASS；`git diff --check bb3dde9e4` PASS；`git ls-files -u` 为空。8 个业务文件仍在合同 allowlist 内；workflow 状态/日志为非业务证据。
