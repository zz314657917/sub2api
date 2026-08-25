### DONE: upstream-codex-oauth-account-identity-s262

# Worker Result

## Task ID
`upstream-codex-oauth-account-identity-s262`

## Status
`done`

## Summary
基于 S259 的 `resolveCredentialAccount` 增加一次性 resolved credential source，并将 Codex OAuth/Setup-token 的 identity namespace 绑定到 `API key + resolved credential account + client-original value`。普通 HTTP、Chat Completions、Anthropic Messages、raw passthrough、普通 WS 和 v2 passthrough 均在现有出站边界执行改写；shadow child 使用 parent，缺 parent 在建立出站前失败，同一 Gin context 的后续账号会覆盖前一账号的 staged source。

## Changed Files
- `backend/internal/service/openai_codex_account_identity.go`
- `backend/internal/service/openai_codex_account_identity_test.go`
- `backend/internal/service/openai_agent_identity_compat_test.go`
- `backend/internal/service/openai_compat_model_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_spark_shadow_credential_outbound_test.go`

## Commands Run
```text
go test ./internal/service -run 'Test.*(CodexAccountIdentity|AccountIdentity|SparkShadowOutbound)' -count=10 -> PASS
go test ./internal/service -run 'Test(ForwardAsChatCompletions|ForwardAsAnthropic|OpenAIGatewayService_Forward_WSv2|OpenAIGatewayService_ProxyResponsesWebSocketFromClient).*' -count=1 -> PASS
go test ./internal/service -count=1 -timeout=3m -> PASS (66.365s)
go test ./cmd/server -run '^$' -count=1 -> PASS
git diff --check -> PASS
```

## Test Output
```text
focused identity/shadow outbound: ok ... 7.561s
protocol-focused service tests: ok ... 17.900s
complete service package: ok ... 66.365s
server compile probe: ok ... [no tests to run]
```

真实出站断言覆盖：HTTP Forward 与 raw passthrough 的 body/header；Chat Completions 与 Anthropic Messages 的实际 upstream session header；普通 WS 的 parent header/body；v2 passthrough 的首帧、`session.update`、后续 `response.create` 三个实际 upstream frame。测试还断言同一原始 session/prompt-cache 值只 scope 一次，并断言 shadow -> 另一 OAuth 账号切换不会遗留 parent namespace。

## Risks
- 未执行真实 provider、数据库、容器、部署、浏览器或外部网络操作；这些均在 contract 外。
- Chat Completions/Anthropic Messages 的转换器会按既有协议删除 `client_metadata`，因此其实际出站证据断言最终 session header；raw/HTTP 和 WS 路径保留并断言 body/frame metadata。
- 完整 service 回归中的日志包含既有测试模拟的 provider 错误，但最终命令退出为 PASS。

## Knowledge Candidates
- 无。稳定结论留待 Codex/Evaluator 审核后决定是否写入知识库。

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Commits
- Business: `a860d3634` (`fix(openai): isolate codex oauth identity by credential`)
- Worker report: `b4be2d932` (`docs(workflow): record S262 developer result`)

## Blocked Reason
- N/A
