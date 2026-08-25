### PASS: upstream-grok-official-ua-s255

## Findings

未发现阻断问题。`a5326be21` 仅修复第一次独立 QA 指出的两条 unit-tag
Grok OAuth UA 期望：`openai_gateway_grok_test.go:329` 与 `:376` 均改为
`grokOfficialOAuthUserAgent`；没有生产代码、合同、主工作区或范围外文件变更。

## Executed Checks

- 审查 combined business diff `21c854c91^..a5326be21`：7 个既定 service/test
  owner 是唯一业务路径；随结果提交的 worker result 属合同允许文档。业务变更为
  108 additions、5 deletions；remediation 仅在一个 allowed test owner 中替换两条
  UA 断言。
- `git diff --binary 21c854c91^ 21c854c91 | git -C F:/mcplugins/sub2api apply --check --verbose -`：7 个业务文件全部成功。
- `git diff --binary a5326be21^ a5326be21 | git -C F:/mcplugins/sub2api apply --check --verbose -`：两个 hunk 均成功（offset -9）。主线为
  `2407d6d14`，其 staged/unmerged index 为空，7 个 S255 owner 未脏。
- `go test ./internal/service -run "Test(BuildGrokResponsesRequestUsesOfficialCLIUserAgent|ForwardGrok.*|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge|AccountTestService_GrokOAuthUsesOfficialCLIUserAgent|ForwardAsRawChatCompletions_Grok(OAuthUsesOfficialCLIUserAgent|APIKeyDoesNotUseOfficialCLIUserAgent))" -count=10 -v`：PASS，`ok .../internal/service 0.094s`。
- `go test ./internal/service -run "Test(BuildGrokResponsesRequest|ForwardGrok|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge|AccountTestService.*Grok|ForwardAsRawChatCompletions.*Grok)" -count=1 -v`：PASS，`ok .../internal/service 0.073s`。
- `go test ./internal/service -count=1`：PASS，`ok .../internal/service 65.831s`。
- `go test ./cmd/server -run '^$' -count=1`：PASS，`ok .../cmd/server 5.515s [no tests to run]`。
- 对合同 7 文件执行 `gofmt -w` 后，`git diff --check` 通过；conflict-marker
  `rg` 无命中（exit 1）；基线 `5d7c68001..HEAD`、cached、working diff 和
  unmerged index 检查在写本报告前均为空。
- `openai_gateway_grok_test.go` 受 `//go:build unit` 约束；默认
  `go test ./internal/service -list 'Test(ForwardAsChatCompletionsForGrok|ForwardAsAnthropicForGrok)' -count=1`
  无测试输出，确认合同默认标签命令不会发现这两条 unit-tag 测试。
- 静态调用链：`ForwardAsChatCompletions` 对 `PlatformGrok` 进入
  `forwardAsRawChatCompletions`，其 `account.IsGrokOAuth()` 分支设置共享常量，
  与 streaming raw-chat 断言 `:329` 对应；`ForwardAsAnthropic` 的 Grok
  请求经 service Grok 分支进入 `forwardGrokResponses`，再由
  `buildGrokResponsesRequest` 的 OAuth 分支设置同一常量，与 Messages 断言
  `:376` 对应。

## Unverified Risks

- 依合同未运行 `-tags=unit`；已有范围外 unit 基线编译问题仍阻断直接运行这两条
  unit-tag 测试。默认标签 focused、完整 service 与 server 编译均有本次通过证据。
- 未运行真实 provider、数据库、容器、部署或 push，均不在授权范围内。

## Recommendation

可继续 Controller/mainline integration；待范围外 unit 基线问题消除后，补跑对应
`-tags=unit` 测试以移除剩余的直接执行证据缺口。
