---
task_id: upstream-baseline-20261003
qa_date: 2026-10-03
verdict: PASS
verdict_scope: selected-production-file-regressions-and-full-build
full_service_unit_suite: BLOCKED
---

## Findings

- 本补丁范围未发现明确问题。独立 PASS 仅适用于下述选定回归集合、完整生产构建和补丁范围审查；不代表 service 全包或全部测试套件通过。
- 原合同 service unit 命令仍在测试编译阶段 BLOCKED，实际执行测试数为 0。既有测试包含 `stringPtr` 重定义、旧 `computeTokenBreakdown` / `calculateCostInternal` 签名、`buildCountTokensRequest` 返回数量不匹配及不存在的 Proxy 字段。未修改这些文件或移除测试。

## Executed Checks

工作目录：`E:/codex-worktrees/sub2api/upstream-integration-20261003`。

1. 在 `backend/internal/service` 执行以下 PowerShell 命令，退出码 0，`ok command-line-arguments 0.660s`：

   ```powershell
   $pkg = go list -json . | ConvertFrom-Json
   if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
   $files = @($pkg.GoFiles) + @(
     'billing_token_cost_request_test.go',
     'pricing_group_id_only_test.go',
     'billing_token_cost_group_id_test.go'
   )
   go test -tags unit -v -count=1 @files
   ```

   `go list` 的 `GoFiles` 共 422 个，使用当前平台实际生产源文件列表，未纳入 `CgoFiles`、`IgnoredGoFiles` 或其他测试文件。没有增加 helper、伪造生产类型或删除测试。3 个原始测试文件共 9 个顶层测试实际执行且全部 PASS：

   - `TestLegacyLongContextRule_OnlyGemini`
   - `TestCalculateTokenCostForRequest_ExplicitPricingWinsOverLegacyRule`
   - `TestCalculateTokenCostForRequest_LegacyRuleFollowsGroupToggle`（内部同时验证 true/false）
   - `TestCalculateTokenCostForRequest_NoResolverFallsBackToCatalog`
   - `TestGatewayTokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated`
   - `TestOpenAITokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated`
   - `TestTokenCostInput_GroupIDPrecedenceAndFallback`
   - `TestCalculateTokenCostForRequest_GroupIDOnlyUsesChannelPricing`
   - `TestCalculateTokenCostForRequest_AbsentGroupIDFallsBackToCatalog`

   GroupID-only service、Gateway、OpenAI 三条实际生产计费路径均计算出输入 0.011、输出 0.0044、合计 0.0154；测试断言比较真实渠道单价与生产返回金额，而非仅检查 GroupID 字段存在。

2. 在 `backend` 执行 `go build ./...`，退出码 0。
3. 在 `backend` 独立复现 `go test -tags unit ./internal/service -run 'Test.*(TokenCost|GroupID|LegacyLongContext)' -count=1`，退出码 1，`[build failed]`。首批错误位于：
   - `ops_health_score_test.go:442` 与 `usage_leaderboard_reward.go:1042`：`stringPtr` 重定义；
   - `billing_service_test.go:1395,1417`：`computeTokenBreakdown` 参数数量；
   - `billing_service_unified_test.go:32`：`calculateCostInternal` 参数数量；
   - `gateway_context_management_test.go:637`：返回值数量；
   - `proxy_update_probe_invalidation_test.go:55-63`：`FallbackMode`、`ExpiryWarnDays` 等类型漂移。
4. 根目录 `git diff --check` 退出码 0，仅已有 `docs/workflow/status.md` 的 LF/CRLF 提示。
5. 精确 diff 审查：业务修改仅 `billing_token_cost_request.go` 的 4 处 GroupID 接线，加新增 `billing_token_cost_group_id_test.go`。已有 `docs/workflow/status.md` 改动保留，本 QA 仅写本报告。无 commit、push、共享工作区修改、部署、数据库修改或 provider 请求。

优先级审查：

- `tokenCostInput` 显式 `GroupID` 优先；nil 时采用 `Group.ID`；两者均无时仍 nil。新增测试实际断言 100 / 200 / nil。
- 已解析的 group/channel 定价继续先进入 unified 路径，优先于 Gemini legacy 规则；预解析 `Resolved` 由 `CalculateCostUnified` 复用。
- legacy 判断与执行位置保持不变；内置价格的 legacy 行为仍遵守 `Group.LongContextPricingEnabled`，既有测试两种开关结果通过。
- 有 resolver 且只有 GroupID 时，现在进入 unified，并将 ID 传给实际 `ModelPricingResolver.Resolve` / 渠道缓存；无 group/ID 或无 resolver 时维持 catalog fallback。

审查业务文件 SHA256：

- `billing_token_cost_request.go`：`3CD3B7C41922337F82F6B06C0E8F3F23025D64542C11F4FAADD6423E8CECA327`
- `billing_token_cost_group_id_test.go`：`0BCD4CC8C0735204F17C90ACCF5E761CF5AA9E291B26EE809FFE5BC6BA428380`

## Unverified Risks

- 显式文件集合产生 `command-line-arguments` 包，绕开的是其他既有测试文件编译失败，不等价于完整包测试或全套测试通过。
- 仅通过 ID 时、`Resolved == nil` 且同时传入 legacy 规则，既有顺序仍先判断 legacy。此次没有改变该顺序；现有 Gateway 会先解析 ID 对应价格，因此选定路径的显式价格优先已获测试证据。不得把本结果扩大到所有未测试的直接调用组合。
- 未启动运行服务、访问真实渠道或验收生产部署；本轮证据为真实生产代码的本地测试及编译。

## Recommendation

本补丁的选定回归与完整生产 build 独立 PASS，可作为 controller 继续精确集成的证据。必须同时保留原 service unit 命令 BLOCKED；如后续门禁要求该全包命令成功，应另立范围修复基线测试后重新验收，不能以此限定 PASS 代替。
