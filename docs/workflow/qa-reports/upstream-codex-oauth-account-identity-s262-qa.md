### PASS: upstream-codex-oauth-account-identity-s262

## Findings

未发现明确问题。`a860d3634` 的业务提交相对冻结 base `5cca17b14` 精确包含 contract 允许的 10 个 service/test 路径；流程文件（contract、`main-log`、worker result）单独归类，不计入业务 allowlist。未发现 conflict marker 或 unmerged index。

## Executed Checks

- `go test ./internal/service -list 'Test.*(CodexAccountIdentity|AccountIdentity|SparkShadowOutbound)'`
  - PASS；列出 8 个身份/协议/shadow 用例。
- `go test ./internal/service -run 'Test.*(CodexAccountIdentity|AccountIdentity|SparkShadowOutbound)' -count=10`
  - PASS，`2.164s`。
- `go test ./internal/service -run 'Test(ForwardAsChatCompletions|ForwardAsAnthropic|OpenAIGatewayService_Forward_WSv2|OpenAIGatewayService_ProxyResponsesWebSocketFromClient).*' -count=1`
  - PASS，`17.900s`。
- `go test ./internal/service -count=1 -timeout=3m`
  - PASS，`66.888s`。
- `go test ./cmd/server -run '^$' -count=1`
  - PASS，`0.060s`，server compile probe。
- 对业务提交的全部 Go 文件执行 `gofmt -d`：PASS，无格式差异。
- `git diff --check`：PASS。
- conflict marker 扫描：PASS，无命中。
- `git ls-files -u`：PASS，无未合并索引。
- exact allowlist：PASS，`a860d3634` 的 10 个业务路径全部在 contract allowlist 内，无越界业务路径。

## Fake-Upstream Outbound Evidence

- HTTP Forward/raw passthrough：测试捕获真实 fake upstream request body/header，断言 `client_metadata.session_id`、`prompt_cache_key` 使用同一原始值生成同一 scoped 值，并断言 `session_id`/`conversation_id` 出站 header 按 account namespace 隔离。
- Chat Completions：测试捕获真实 upstream request，断言出站 `session_id` 为由原始 prompt-cache/session 值派生的 account-scoped UUID。
- Anthropic Messages：测试捕获真实 upstream request，断言出站 `session_id` 为由原始 prompt-cache/session 值派生的 account-scoped UUID；转换器按既有协议移除 `client_metadata`，因此证据以最终 header 为准。
- 普通 WebSocket：测试捕获真实 upstream frame/header，断言 `response.create` 的 metadata 使用 parent namespace；后续切换到另一 OAuth account 后使用另一 namespace，未遗留 parent；普通 WS 对原始值只 scope 一次。
- v2 passthrough WebSocket：测试捕获真实 upstream 的首个 `response.create`、`session.update`、后续 `response.create` 三帧，三帧均使用同一 parent-scoped metadata；dial header 使用 parent token/account-id。
- shadow：shadow child 通过 `resolveCredentialAccount` 使用 parent credential namespace；同一 Gin context 后续切换到另一 OAuth account 会覆盖 staged source；缺 parent 在 HTTP/WS 建立 outbound 前 fail-closed，dial/request 数为零。
- API-key fallback：身份 helper 测试确认 API-key 继续使用既有 `isolateOpenAISessionID` 行为。

## Unverified Risks

- 未执行真实 provider、数据库、容器、部署、浏览器或外部网络操作；均不在本 contract 范围内。
- 完整 service 测试日志包含既有模拟 provider 错误，但命令退出码为 0。
- 主工作区只读快照：QA 前后业务提交 patch-id 均为 `d4defdf2e68c21d20f433eb5eb654f3ff4a8d551`，已有 Pixel Cafe 修改与 `outputs/` 保持不变；QA 期间未跟踪的 `sub2api` 条目从 porcelain 中消失，属于外部漂移，未由本 QA 修改或恢复。

## Recommendation

可继续进入 Evaluator/集成门禁；本轮代码级、fake-upstream 运行态和 contract 规定的编译/回归检查均通过。集成前继续保留主工作区脏文件边界，并将 `sub2api` 未跟踪条目的外部漂移交由主流程确认。

## Scope and Provenance

- Reviewed business commit: `a860d3634` (`fix(openai): isolate codex oauth identity by credential`)
- Frozen base: `5cca17b14`
- Worker worktree: `E:/codex-worktrees/sub2api/upstream-codex-oauth-account-identity-s262`
- QA only added this report file; no business file was modified.
