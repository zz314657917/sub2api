### PASS: upstream-v0184-compat-fixes-s276

## Findings

未发现明确问题。实际业务 diff 的 S276 文件均属于 contract allowlist；受保护脏文件和 `outputs/**` 状态未被本轮修改。四项行为与 contract 一致：Responses 流 item/content 生命周期及 terminal output、Anthropic 空对象工具参数占位、SMTP TLS 三态回退、连字符版本比较。

## Executed Checks

- `go test ./internal/pkg/apicompat -run '^TestAnthropicEventToResponses_' -count=10`：通过。
- 7 类流回归（`TextEmitsContentPart`、`DoneEventsCarryFullText`、`CompletedCarriesOutput`、`ToolCallCompletedCarriesArguments`、`ThinkingAfterTextKeepsMessageOutput`、`MultipleTextBlocksAdvanceContentIndex`、`ItemLifecycleIsBalanced`）以 `-count=10 -v` 独立复跑：全部通过。
- 去除并恢复 allowlist 中 3 个 `unit` 测试文件的 build tag 后，`go test ./internal/service -run 'ToolArgumentsAreValidJSON|TestAppendRawJSON_EmptyObjectPlaceholder' -count=10`：通过；标签已恢复。
- 去除并恢复 `update_service_test.go` 的 `unit` 标签后，`go test ./internal/service -run '^TestCompareVersionsHyphenatedSuffix$' -count=10`：通过；标签已恢复。
- `go test -tags unit ./internal/handler/admin -run '^Test(TestSMTPRequest|SendTestEmailRequest)' -count=10`：通过（省略/false/true 三态）。
- `go test ./internal/pkg/apicompat ./internal/service ./internal/handler/admin`：通过。
- `go test ./cmd/server -run '^$'`：通过。
- contract allowlist 文件 `gofmt -d`：无输出；`git diff --check`：通过；`git diff --name-only --diff-filter=U`：空。
- changed/untracked 路径核对：S276 新增/修改文件在 allowlist；额外变更仅为先存的 workflow、knowledge、保护文件及 `outputs/**`。保护文件 SHA-256 保持预期：`api_key_auth.go`=`F0ED9DB7...36B70`、`api_key_auth_route_breaker_test.go`=`504A7394...0CB77`、`admin_service.go`=`451914FC...53FDE`、`AdminCafeRoomsView.vue`=`4999C158...8EBB7`。
- `go test -tags unit ./internal/service ...`：失败于既有 unit 编译基线（`stringPtr` 重定义、旧函数签名/字段等），未修改或修复；不归因于 S276。

## Unverified Risks

- 未执行真实 provider、数据库、容器、部署或浏览器运行态 smoke，按 contract 明确排除。
- 完整 `unit` service 套件仍受仓库既有测试编译漂移阻断；本次 allowlist 定向用例已独立通过。

## Recommendation

QA 通过，可进入最终 Evaluator/build。提交前继续保持精确 allowlist 暂存；不建议将既有 `unit` 编译基线问题混入 S276 修复。
