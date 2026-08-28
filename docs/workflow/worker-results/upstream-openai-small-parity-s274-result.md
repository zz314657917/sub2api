### DONE: upstream-openai-small-parity-s274

# Worker Result

## Task ID
upstream-openai-small-parity-s274

## Status
`done`

## Summary
- 已按 contract 移植三项行为：入站 OpenAI WebSocket 客户端正常关闭/取消不再记录账号失败；非流式合成与透传 SSE 对 `response.failed`、`error` 复用流式 failover 分类；EasyPay `mapi.php` 根相对 `payurl/payurl2/qrcode` 在合法 HTTP(S) base 下补全绝对 URL。
- 每个切片均新增 focused 回归测试；未修改 schema、依赖、配置、容器、数据库或生产行为。

## Changed Files
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_ws_client_close_attribution_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_nonstreaming_terminal_failure_failover_test.go`
- `backend/internal/payment/provider/easypay.go`
- `backend/internal/payment/provider/easypay_url_test.go`
- `docs/workflow/worker-results/upstream-openai-small-parity-s274-result.md`

## Commands Run
```text
gofmt -w <six changed Go files> -> passed
go test ./internal/handler -run 'TestOpenAIWSIngressEndedByClient|TestShouldReportOpenAIWSProxyAccountFailure' -count=10 -> PASS
go test ./internal/service -run 'TestHandleSSEToJSON|TestHandlePassthroughSSEToJSON|TestNonStreamingTerminalFailureFailover|TestOpenAIStream.*Failover' -count=10 -> PASS
go test ./internal/payment/provider -run 'TestResolveEasyPayReturnedRef|TestEasyPayCreatePayment' -count=10 -> PASS
go test ./internal/handler -> PASS
go test ./internal/service -> PASS
go test ./internal/payment/provider -> PASS
go test ./cmd/server -run '^$' -> PASS
go test ./internal/service -run 'TestNonStreamingTerminalFailureFailover|TestOpenAIStreamErrorEventShouldFailover|TestHandleSSEToJSON|TestHandlePassthroughSSEToJSON|TestOpenAIStream.*Failover' -count=10 -> PASS (复审修复后)
gofmt -w backend/internal/service/openai_gateway_service.go backend/internal/service/openai_nonstreaming_terminal_failure_failover_test.go -> PASS (复审修复后)
go test ./internal/handler -run 'TestOpenAIWSIngressEndedByClient|TestShouldReportOpenAIWSProxyAccountFailure' -count=10 -> PASS (QA bounded fix)
go test ./internal/payment/provider -run 'TestResolveEasyPayReturnedRef|TestEasyPayCreatePayment' -count=10 -> PASS (QA bounded fix)
gofmt -w backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_ws_client_close_attribution_test.go backend/internal/payment/provider/easypay.go backend/internal/payment/provider/easypay_url_test.go -> PASS (QA bounded fix)
go test ./internal/handler -run 'TestOpenAIWSIngressEndedByClient|TestShouldReportOpenAIWSProxyAccountFailure' -count=10 -> PASS (final bounded-fix rerun)
go test ./internal/payment/provider -run 'TestResolveEasyPayReturnedRef|TestEasyPayCreatePayment' -count=10 -> PASS (final bounded-fix rerun)
git diff --check -- <allowlist> -> PASS
git diff --name-only --diff-filter=U -> empty
allowlist status check -> empty outside-allowlist result
```

## Test Output
```text
internal/handler: ok (27.426s)
internal/service: ok (65.375s)
internal/payment/provider: ok (1.392s)
cmd/server: ok [no tests to run] (5.533s)
Focused suites passed repeatedly with -count=10.
```

## Risks
- 未执行真实 provider 请求、部署、容器或共享数据库验证，均在 contract 外。
- `error` 终止帧采用保守瞬时标记分类；显式 cyber policy 不重试，账号/workspace 状态、凭据认证失败及 401/429/529 保持上游 failover 语义；已提交响应或无账号仍保持协议错误路径。
- EasyPay 协议相对 `//host/path` 保持原样，仅补全单斜杠根相对引用；WS 归因测试提供 contract 精确名称 `TestShouldReportOpenAIWSProxyAccountFailure`。

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- N/A
