# 当前任务快照

最后更新：2026-07-04 02:40 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前分支：`main`。
- 本轮目标：把前置 `redeem-invitation-reject-s45` 和 S46-S52 小补丁链按顺序重放到 `main`，完成验证、no-ff 合并、推送确认。
- 集成分支：`codex/upstream-main-v0143-s45-s52-batch`。
- 合并基线：`485eaf801 docs: record affiliate risk merge`。
- 集成分支最终 head：`40cfe08e7 docs: record s45 s52 batch validation`。
- `main` merge commit：`d231da13d merge: add upstream s45 s52 patch batch`。

## 当前目标

- 提交合并后的 workflow / handoff 状态同步。
- 推送 `main`，确认 `origin/main` 指向最新主线提交。
- 推送后进入 release validation 或下一个已批准 Sprint。

## 本次已完成

- 已从干净 `main` 创建集成分支 `codex/upstream-main-v0143-s45-s52-batch`。
- 已按计划使用 `cherry-pick -x` 顺序重放：
  - `544accdd3` redeem 普通兑换拒绝 invitation/internal marker code。
  - `af6c8fdeb` Codex import 优先按用户身份匹配，并将 import API 请求超时提升到 120s。
  - `9aa85e59e` ops realtime stats 使用缩减账号查询并抑制 request canceled 噪音。
  - `6888e9da5` Codex OAuth 保留 reasoning items，剥离 replay-unsafe `rs_*` id。
  - `512f44c13` / `6abeb0796` 规划并实现 compact 跳过 Codex image bridge 注入。
  - `248bf80dc` / `c558b6eda` 规划并实现 Claude Code streaming idle keepalive。
  - `fed128046` / `5ce438fa7` 规划并实现 Anthropic API Key Bearer auth scheme。
  - `b6970cdc6` / `4f9542e34` 规划并实现 Antigravity OAuth 401 可恢复临时不可调度。
- 已按规则处理 workflow 冲突：`docs/workflow/status.md` / `main-log.md` 不接受旧分支过期状态，保留新增 task/result/QA 文件，最终统一更新为 S52 done / `total_sprints: 52`。
- 已将集成分支 no-ff 合入 `main`，merge commit 为 `d231da13d`。
- 已更新 `docs/workflow/status.md` 和 `docs/workflow/main-log.md`，记录 S45-S52 批量重放、验证和主线合并。

## 已确认事实

- 普通 redeem endpoint 不再消耗 invitation/internal marker code，统一走 unsupported type。
- Codex import 匹配优先级改为 user identity 优先，shared account id 最后 fallback。
- Ops realtime 账号统计新增内部缩减查询路径，不作为通用账号列表接口复用。
- Codex reasoning input 会跨轮保留，但剥离 `id`，缺失 `summary` 时补空数组。
- `/responses/compact` 不再被注入 image bridge tool、tool_choice 或 instructions。
- Claude Code `>= claude-cli/2.1.193` streaming keepalive 使用空 content delta；旧客户端继续 ping。
- Anthropic API Key 账号新增 `extra.anthropic_apikey_auth_scheme = "authorization_bearer"` 可选行为，默认仍是 `x-api-key`。
- Antigravity OAuth 401 有 refresh token 时进入临时不可调度恢复路径；无 refresh token 时仍置 error。
- `origin/main..HEAD` denied-path audit 未命中 Ent、migrations、deploy、README、`.github`、`knowledge/**`、无关 payment/welfare 路径。

## 待验证点

- 推送后需要确认 `origin/main` 指向最新提交。
- S52 unit-tag 命令仍受已知无关 service unit 编译基线阻挡；后续若要恢复完整 `-tags=unit` service 测试，需要单独修旧 billing 单测。
- 本轮未做完整发布验证；上线前建议补后端启动 smoke、迁移检查、前端 build 或容器构建，以及关键网关真实请求 smoke。

## 当前结论

- S45-S52 小补丁链已完成代码级集成、计划内验证和 no-ff 主线合并。
- 唯一未通过项是计划内允许记录的 S52 unit-tag 已知无关编译 blocker；其余定向测试、typecheck、diff check、冲突标记扫描和 denied-path audit 均通过。

## 下一步

1. 提交本 handoff / workflow 状态同步。
2. 推送 `main` 并确认 `origin/main`。
3. 推送确认后执行 release validation 或进入下一个已批准 Sprint。

## 验证记录

- `go test ./internal/service -run "TestRedeemRejects.*BeforeTransaction|TestFulfillPaidOrder.*Redeem|TestPaymentRechargePackage|TestFirstRechargeBonus|TestMonthlyRecharge" -count=1` 通过。
- `go test ./internal/handler/admin -run "TestCodexIdentity|TestParseCodexSessionImport|TestNormalizeCodexImport|TestResolveCodexImport|TestMergeCodexImport|TestImportCodex|TestOpsRealtimeRequestCanceled|TestOpsRealtime|TestGetConcurrencyStats|TestGetAccountAvailability|TestGetRealtimeTrafficSummary" -count=1` 通过。
- `go test ./internal/service -run "TestListAllAccountsForOps|TestOps.*Concurrency|TestOps.*Availability|TestFilterCodexInput|TestApplyCodexOAuthTransform|TestOpenAIGatewayServiceForward_CodexBridge|TestOpenAIGatewayServiceForward_.*Image|TestOpenAIGatewayService_CodexImageGenerationBridge|TestGatewayService_StreamingKeepalive|TestGatewayService_StreamingReusesScannerBufferAndStillParsesUsage|TestDetachUpstreamContextIgnoresClientCancel|TestAccount_GetAnthropicAPIKeyAuthScheme|TestGatewayService_AnthropicAPIKeyPassthrough_BearerAuthScheme|TestGatewayService_AnthropicAPIKeyBearerAuthScheme|TestBuildUpstreamModelsRequestsForAPIKeyAccounts" -count=1` 通过。
- `go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListOpsAccountsForStats|TestAccountRepoSuite/TestListWithFilters" -count=1` 通过。
- `go test -tags=unit ./internal/service -run "TestRateLimitService_HandleUpstreamError_OAuth401SetsTempUnschedulable|TestRateLimitService_HandleUpstreamError_OAuth401NoRefreshTokenSetsError|TestTokenRefreshService_RefreshWithRetry_Antigravity|TestTokenRefreshService_RefreshWithRetry_ClearsTempUnschedulable" -count=1` 已尝试；失败仅命中已知无关 service unit 编译基线：旧 billing tests 引用 `ImageOutputPriceExplicit`、旧 `computeTokenBreakdown` 和旧 `calculateCostInternal` 签名。
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"` 通过。
- `git diff --check` 通过。
- `rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" .` 无命中。
- denied-path audit over `git diff --name-only origin/main..HEAD` 输出 `DENIED_PATH_AUDIT_PASS`。
