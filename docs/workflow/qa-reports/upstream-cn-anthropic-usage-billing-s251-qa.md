### PASS: upstream-cn-anthropic-usage-billing-s251

## Findings

- 未发现明确问题。独立复跑确认 Kimi、GLM、DeepSeek 的 CN Anthropic-compatible usage 会归一到互斥计费桶；Kimi `message_delta` 显式未缓存输入 `0` 能覆盖 `message_start` 总量。
- `claudeUsageToOpenAIUsage` 保持网关内部总输入（普通输入 + cache creation + cache read），现有 `RecordUsage` selector 回归通过，未改动其 owner。

## Executed Checks

所有命令均在隔离 QA worktree `E:/codex-worktrees/sub2api/upstream-cn-anthropic-usage-billing-s251-qa` 执行，未访问 provider、数据库、容器、部署或远端写操作。

- `go test ./internal/service -run "Test(ParseSSEUsagePassthroughNormalizesKimiPromptUsage|ParseSSEUsagePassthroughKimiFullyCachedInputReplacesStartTotal|ParseClaudeUsageFromResponseBodyNormalizesCNProviderAliases|ParseSSEUsagePassthroughNormalizesGLMAndDeepSeekAliases|MergeAnthropicUsageNormalizesKimiStreamForOpenAIBilling|MergeAnthropicUsageNormalizesGLMAndDeepSeekAliases|ClaudeUsageToOpenAIUsagePreservesCNProviderNativeAnthropicBuckets|CNProviderAnthropicUsageBillsUncachedInput)" -count=10`：PASS，`ok github.com/Wei-Shaw/sub2api/internal/service 5.437s`。
- `go test ./internal/service -run "TestGatewayService_ParseSSEUsagePassthrough|TestParseClaudeUsageFromResponseBody|TestOpenAIGatewayServiceRecordUsage" -count=1`：PASS，`0.110s`。
- `go test ./internal/pkg/apicompat -run "TestAnthropicUsage" -count=1`：PASS，`0.633s`。
- `go test ./internal/service -count=1`：PASS，`64.574s`。
- `go test ./cmd/server -run '^$' -count=1`：PASS，`5.567s`（`[no tests to run]`，编译门禁通过）。
- `go test ./internal/service -list "...8 named focused tests..."`：八个 contract 指定测试均被发现，避免 selector 空跑。
- 因 QA 规则只允许写本报告，以只读等价检查 `gofmt -d <five product/test owners>` 代替会改写业务文件的 `gofmt -w`；无输出。`git diff --check` 通过，五个 owner 的冲突标记扫描无匹配，`git ls-files -u` 为空。
- QA worktree 在写本报告前无 staged 或 unstaged 改动；业务提交 `46185fcca0eef682b14c23cb27741072e45609a6` 精确只含五个允许的 product/test owners。任务 contract commit `66c2e1343..HEAD` 仅含该五个 owner 与 Developer result；写入后本报告是唯一额外 QA 文件。
- Provenance：业务提交父提交为 `66c2e134331bad3bf273f25cf71cea6713933e24`；源提交 `695ebede70e0bed4c8fd4c87b5a426448a08ea4c` 已确认是 `upstream/main` 的祖先。实现以本地 `gateway_service.go` 共享 normalizer 适配，S229 native parser 调用该 normalizer，Responses DTO merge 同步归一化，未引入上游分叉的 gateway 拓扑。
- 主工作区保护（只读检查）：`F:/mcplugins/sub2api` HEAD 仍为 `1879b42a64d8e480f21860f75c67b86f789d136b`；复跑前后均保留其既有 Pixel Cafe/Groups、`knowledge/*` 脏改及三个未跟踪项，未由本 QA 改动。

## Scope

- 本 QA 只新增本文件；未修改任何业务、前端、知识库或主工作区文件。
- 已核验的业务改动仅为：`backend/internal/pkg/apicompat/types.go`、`backend/internal/service/gateway_service.go`、`backend/internal/service/gateway_forward_as_responses.go`、`backend/internal/service/openai_gateway_messages_anthropic_native.go`、`backend/internal/service/kimi_anthropic_usage_test.go`。

## Unverified Risks

- 证据为本地 fixture/default-tag 测试与编译；按 contract 未进行真实 CN provider 流量、共享数据库、容器、部署或生产环境验证。

## Recommendation

可继续由 Controller 进行 contract review 与集成裁决；本独立 QA gate 通过。
