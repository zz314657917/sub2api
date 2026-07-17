### PASS: upstream-main-compat-s77

## Task ID

`upstream-main-compat-s77`

## Verdict

`PASS`（S77 post-fix 定向验收通过；完整 service 包保留既有非阻断失败）

## Contract Checked

- `docs/workflow/tasks/upstream-main-compat-s77.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`（相对 `82df0cb412` 的 27 个变更路径全部在 allowlist）
- denied paths touched: `no`
- commands run:

```text
backend: go test ./internal/config -run 'TestLoadDefaultOpenAIWSConfig|TestLoadOpenAIWSClientFirstMessageTimeoutFromEnv|TestValidateConfig_OpenAIWSRules' -count=1 -> PASS
backend: go test ./internal/handler -run 'TestOpenAIResponsesWebSocket_(FirstMessageTimeoutUsesConfig|OpenAIPassiveImageNamespacePreservesLegacyPermissionGate)|TestOpenAIGatewayHandler(Responses_Grok|Responses_OpenAI|ChatCompletions_OpenAI)' -count=1 -> PASS
backend: go test ./internal/service -run 'TestResolveOpenAIWSClientFirstMessageTimeout|TestReadOpenAIWSClientMessage_.*|TestOpenAIGatewayService_(Forward_WSv2RejectsMalformedEvent.*|ProxyResponsesWebSocketFromClient_(RejectsMalformedEventBeforeOutput|InterTurnReadTimeoutClosesClient|PassthroughRejectsMalformedEvent.*))|TestIsImageGenerationIntent.*|TestIsImageGenerationIntentMap.*' -count=1 -> PASS
backend: go test ./internal/server/routes -run 'TestResolveAPIKeyRouteForJSONModel_GrokImageIntent.*' -count=1 -> PASS
backend supplemental: go test ./internal/config ./internal/handler ./internal/service ./internal/server/routes -run 'OpenAIWS|ImageGenerationIntent|ImageIntent' -count=1 -> PASS
frontend: npm.cmd run test:run -- src/components/layout/__tests__/TablePageLayout.spec.ts -> PASS (1 test)
frontend: npm.cmd run typecheck -> PASS
frontend: npm.cmd run build -> PASS
git diff --check 82df0cb412..HEAD -> PASS
git diff --check (working tree) -> PASS
allowlist / denied-path / conflict-marker audit -> PASS
full backend service: go test ./internal/service -count=1 -> FAIL (existing group_peak_rate drift only)
```

- manual checks:

```text
WS 默认/自定义首帧超时与正数校验 -> covered and PASS
WS 取消/超时后阻塞 reader 退出、inter-turn idle close -> covered and PASS
WS ctx_pool/legacy 畸形上游 JSON 在输出前触发 failover，输出后安全终止 -> covered and PASS
WS passthrough 畸形上游 JSON 在输出前触发 failover，输出后安全终止且不下发畸形帧 -> covered and PASS
Grok 被动 image_gen namespace 与显式 image_generation、图片模型、明确 tool choice -> covered and PASS
非 Grok（包括 OpenAI）Responses、Chat Completions 和 WS 原有图片意图语义 -> regression covered and PASS
TablePageLayout 移动模式保留 overflow-x-auto -> source regression covered and PASS
```

## Findings

- 独立评审发现的 passthrough malformed JSON P1 已关闭：adapter 现在在每个上游 text frame 写客户端前执行 `json.Valid`，输出前返回 failover，输出后安全终止；新增两个真实 passthrough relay 用例均通过。
- 未发现残留的 S77 实现缺陷、范围越界或冲突标记。
- 完整 `go test ./internal/service -count=1` 仍失败，但失败集中在未被 S77 修改的 `group_peak_rate_test.go`：峰值边界、时区转换、标准订阅倍率和计费序列断言漂移（`TestPeakMultiplierAt_*`、`TestPeakMultiplier_*`）。S77 定向 service 测试全部通过，因此该结果记录为既有遗留风险，不改变本 Sprint 的 focused acceptance 判定。
- 前端构建保留仓库既有 Browserslist、动态导入、chunk-size 和 Node `DEP0190` 警告；没有新增错误。

## Bug Owner Recommendation

`codex-planner`（仅针对既有 `group_peak_rate` 测试漂移另行拆分；不并入 S77）

## Root Cause

`none`（S77 范围内）；完整 service 失败属于既有测试/时区环境漂移。

## Retest Scope

`none`。若后续修复 `group_peak_rate`，只需重跑完整 service 包及其峰值倍率测试。

## Knowledge Promotion

`none`

## Unverified Risks

- 未执行真实上游 Grok 或 OpenAI WebSocket 联调。
- 未执行带认证会话的浏览器 smoke；前端以组件测试、typecheck 和 production build 验收。
- 未执行 race、部署或容器更新。
