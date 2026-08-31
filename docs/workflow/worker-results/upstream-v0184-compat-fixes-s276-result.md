### DONE: upstream-v0184-compat-fixes-s276

# Worker Result

## Task ID
upstream-v0184-compat-fixes-s276

## Status
`done`

## Summary
按 contract 完成四项本地兼容修复并补充定向回归测试：Anthropic->Responses item 生命周期与 text content index、空对象 tool input 占位、SMTP TLS 省略回退、连字符版本后缀比较。Controller review 后将流转换对齐为上游验证过的 current-buffer 状态模式，并补齐 7 类事件不变式测试。未触碰 provider、计费、schema、依赖或外部状态。

## Changed Files
- `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`
- `backend/internal/pkg/apicompat/anthropic_to_responses_stream_test.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/gateway_forward_as_chat_completions_test.go`
- `backend/internal/service/gateway_forward_as_responses_test.go`
- `backend/internal/service/openai_gateway_anthropic_native_pump_test.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/admin/setting_handler_email_test.go`
- `backend/internal/service/update_service.go`
- `backend/internal/service/update_service_test.go`
- `docs/workflow/worker-results/upstream-v0184-compat-fixes-s276-result.md`

## Commands Run
```text
gofmt -w <contract allowlist Go files> -> pass
go test ./internal/pkg/apicompat -run <7 controller lifecycle regressions> -count=10 -> pass
go test ./internal/pkg/apicompat -> pass
go test ./internal/service -run <5 real tool argument paths> -count=10 -> pass (temporarily removed and then restored unit tags in the two allowlist test files so the focused tests actually executed)
go test ./internal/handler/admin -run SMTP tests -tags unit -count=1 -> pass
go test -tags unit ./internal/handler/admin -run '^Test(TestSMTPRequest|SendTestEmailRequest)' -count=10 -> pass (omitted/false/true matrix for both endpoints)
go test ./internal/service -run '^TestCompareVersionsHyphenatedSuffix$' -count=10 -> pass (temporarily removed and restored the allowlist test file's unit tag; covers hyphenated and ordinary versions)
go test ./internal/pkg/apicompat ./internal/service ./internal/handler/admin -> pass
go test ./cmd/server -run '^$' -> pass
git diff --check <allowlist> -> pass; git diff --name-only --diff-filter=U -> empty
```

## Risks
- Provider/database/container/browser runtime smoke was not run per contract.
- Existing service tests are generally behind the `unit` build tag.
- `go test -tags unit ./internal/service ...` 被仓库既有 unit-tag 编译基线阻断（重复 `stringPtr`、旧测试函数签名等，与 S276 diff 无关）；contract 的无 tag focused/broad 命令通过。

## Upstream Provenance
- `8f5451587`: `ported` - current-buffer 流状态、完整 done/terminal output、content index 及 7 类回归已按本地类型适配。
- `da10822d7`: `ported` - 空对象 tool input 占位及 Chat Completions、Responses buffered/streaming、native Anthropic bridge 真实路径回归已适配。
- `c31fe2ed9`: `ported` - SMTP handler 保持本地单文件拓扑，以 `*bool` 区分省略和显式覆盖。
- `9e7aff59d`: `ported` - `parseVersion` 在数字段解析前剥离首个连字符后缀。

## Knowledge Candidates
- `appendRawJSON` treats only a zero-key JSON object as the Anthropic placeholder.
- Stream item state closes before switching item types and increments message content indexes per text block.

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- None.
