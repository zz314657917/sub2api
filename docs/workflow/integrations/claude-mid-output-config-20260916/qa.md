### PASS: claude-mid-output-config

# Independent QA Report

## Historical First-Pass FAIL

- 首轮完整 service suite（65.812s）除当时仅获批准的 validator 失败外，还出现两项 Codex OAuth transform 与一项 SSE usage 失败，故首轮报告为 `FAIL`。三项随后均在 clean `dc851ea3f` base 定向复现，并已纳入独立合同复审的四项精确基线例外；本轮仍完整保留 suite 的 `FAIL (exit 1)`，没有把它改写为 green。
- 首轮还发现 enabled-CCH 测试将 `gatewayForwardingCache` 留为有效 `cchSigning=true`；Developer 已仅在允许的新测试文件补 `t.Cleanup`，本轮 shuffle 复测通过。

## Findings

- 未发现本 task 允许业务路径中的明确缺陷。此前发现的 enabled-CCH 测试进程级 `gatewayForwardingCache` 污染已修复：测试保存先前缓存，并通过 `t.Cleanup` 恢复；初始为空时写入已过期的同类型缓存，避免保留 `cchSigning=true`。
- 完整 `internal/service` suite **仍为 FAIL，真实 exit code 为 1，绝非 green**。JSON `Action=fail` 的测试失败全集恰为已批准、并在 `dc851ea3f` clean base 精确复现的四项：
  1. `TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation` — `claude_code_validator_test.go:52`，`Should be false`；
  2. `TestApplyCodexOAuthTransform_PreservesAllowedTools` — `openai_codex_transform_test.go:18`，`Should be false`；
  3. `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested` — `openai_codex_transform_test.go:275`，approved expected/actual call-ID hash 差异；
  4. `TestParseSSEUsage_SelectiveParsing` — `openai_gateway_service_test.go:3259`，expected `9`、actual `1`。
  另有同 package 的 package-level `Action=fail` 事件（65.787s），对应上述测试失败；没有第五个具名测试失败事件。

## Executed Checks

- 逐行审阅允许业务 diff：`backend/internal/pkg/claude/constants.go`、`backend/internal/service/gateway_request.go`、`backend/internal/service/gateway_mid_conversation_output_config_test.go`。业务改动只在三条允许路径；`gateway_service.go` 未改。`docs/workflow/main-log.md`、`docs/workflow/status.md` 是既存 controller workflow 修改，未纳入本业务变更。
- 静态核验：新 helper 对精确 beta token 和无 message-level target field 返回原 bytes；完整 JSON 失败时保留原 bytes/`changed=false`；顶层 `output_config` 不删除；未知/non-text content block 与非字符串 `text` 值保守保留。既有 context/thinking/fallback 清理未被替换，新用例覆盖该兼容链。
- `go test ./internal/service -list 'Test.*MidConversation'` — PASS，列出 6 个实际新测试：sanitizer、四类 builder 和 enabled-CCH builder。
- `go test ./internal/service -run 'Test.*MidConversation' -count=1` — PASS（0.084s）。覆盖 OAuth default/policy-drop、API-key client-beta present/absent、两类 count_tokens；均断言最终 beta、消息数量/顺序和顶层 `output_config.effort`。CCH 场景基于最终已清理 body 重签名比对。
- CCH 隔离复测：`go test ./internal/service -run '^TestBuildUpstreamRequestOAuthMimicEnabledCCH_MidConversationOutputConfig$' -shuffle=on -count=30` — PASS（0.078s）；`go test ./internal/service -run 'Test.*MidConversation' -shuffle=on -count=10` — PASS（0.103s）。
- 完整 suite：通过 `cmd /v:on` 后台执行 `go test -json ./internal/service -count=1`，stdout 写入 `E:/codex-runtime/claude-mid-output-config-qa-full-service-exitcode.jsonl`；exit-code 文件真实记录 `1`。完整 JSON 解析的全部 `Action=fail` 事件见 Findings，未依赖截断的终端日志。
- `go build ./...` — PASS；`gofmt -l` 对合同四条 Go 路径无输出；根目录 `git diff --check` — PASS（仅 controller workflow 文件 CRLF warning）；`git ls-files -u` — 空。

## Unverified Risks

- 本结论仅为本地 request-builder/fake 级 scoped integration；未发起真实 Anthropic/provider 请求。native `count_tokens` 按合同排除。
- 完整 service suite 的实际结果为 FAIL；四项失败仅因独立合同复审批准的精确基线例外而不阻断本 scoped 结论，不能用于 release、full-suite green 或运行态健康声明。
- 未执行数据库、容器、部署、提交或推送。

## Recommendation

**scoped local integration PASS with documented baseline exceptions。** 可以交由 controller 进入下一 scoped gate；必须同时保留并展示完整 suite `FAIL (exit 1)` 与四项基线失败，不能将本报告表述为完整回归通过或发布批准。
