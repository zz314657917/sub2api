# 当前任务快照

最后更新：2026-07-03 22:21 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前分支：`codex/welfare-voucher-image-preflight`。
- `main` 与 `origin/main` 已确认一致：`9abff8fa588415fca795ae53ac09565b04c8edd1`，提交信息为 `merge: add group peak rate multiplier`。
- 当前分支是在福利券、支付退款、OpenAI/Codex 网关与图片预扣费链路上继续收口；不要在未明确批准前把 `main` 的高峰倍率提交 merge/rebase 进来。

## 当前目标

- 收口 `codex/welfare-voucher-image-preflight`：确认代码提交边界、同步 workflow/knowledge 交接文档、推送分支到远端。
- 保持 mixed dirty tree 纪律：只 stage 本轮文档/knowledge 收口文件，不使用 `git add .`。

## 本次已完成

- 已完成 `main -> origin/main` 推送确认；本地和远端 `main` 均为 `9abff8fa`。
- 已回到 `codex/welfare-voucher-image-preflight`。
- 福利券图片预扣费相关源码已分四笔提交收口：
  - `1d3c2e786 fix(openai): preserve codex reasoning and gpt55 pro pricing`
  - `6a9b68bbc fix(payment): finalize pending refunds safely`
  - `f81aab169 feat(welfare): add voucher wallet image preflight billing`
  - `82625e7f0 test(payment): lock public verify and proxy stubs`
- 源码 dirty 已清空；当前剩余 dirty 只在 workflow/knowledge 文档。
- `backend/migrations/182_welfare_vouchers.sql` 已作为本地福利券 migration 编号使用，因为 `main` 已包含 `181_add_group_peak_rate_multiplier.sql`。

## 已确认事实

- OpenAI/Codex 路径已保留 `reasoning` input items，仅剥离 `rs_*` id，并补齐空 `summary`。
- `gpt-5.5-pro` 已作为独立 Codex 模型名保留，并走 GPT-5.5/GPT-5.4 价格 fallback。
- 退款链路已新增 `REFUND_PENDING`，provider pending 退款会持久化并由 admin query/finalize 原子收口。
- 匿名 `out_trade_no` public verify 仅返回最小订单 DTO。
- 福利签到与 milestone 奖励改为发放 voucher；usage billing、Studio Bridge、OpenAI Video 已接入 voucher-first 抵扣。
- OpenAI Images 已在上游派发前做确定性成本估算；估算失败 fail-closed，最终计费使用 `CostOverride` 与 `RequireBalanceCheck` 防并发透支。
- `BillingCacheService.CheckBalanceAmountEligibility` 已按 voucher + balance 总可用额判断。

## 待验证点

- 当前分支尚未与最新 `main` 合并；如果后续准备进主线，需要单独评估高峰倍率与福利券/支付文件的冲突。
- 本地容器仍需在分支推送后按用户要求另行重建/替换；本轮收口不自动更新容器。
- 文档收口提交后需要执行 `git diff --check`、`git diff --cached --check`，并用 `git status --short --branch` 确认工作树。
- 分支推送后需要用 `git ls-remote origin refs/heads/codex/welfare-voucher-image-preflight` 确认远端 head。

## 当前结论

- `codex/welfare-voucher-image-preflight` 的源码实现和定向测试已完成，当前进入文档收口与远端同步阶段。
- 下一步不应继续叠加新功能；先完成文档提交、推送分支，再决定是否开新的 main 合并/容器更新步骤。

## 下一步

1. 运行 `git diff --check`。
2. 精确 stage 当前 workflow/knowledge 文档。
3. 运行 `git diff --cached --check` 和 staged 文件清单核对。
4. 提交 `docs: refresh workflow and task handoff`。
5. 推送 `codex/welfare-voucher-image-preflight` 并确认远端 head。
6. 如继续上线验证，再单独执行本地容器重建与 payment/welfare/image 手动 smoke。

## 验证记录

- `go test ./internal/service -run "Test.*Payment.*|Test.*Refund.*|TestBillingCacheServiceCheckBalanceAmountEligibility|TestOpenAIGatewayServiceEstimateOpenAIImagesCost|TestUsageBillingWalletBalanceCost" -count=1` 通过。
- `go test ./internal/handler -run "Test.*Payment.*|Test.*Refund.*|Test.*OpenAI.*Images|TestVerifyOrderPublic" -count=1` 通过。
- `go test ./internal/repository -run "Test.*UsageBilling|Test.*Studio|Test.*OpenAIVideo|Test.*Welfare" -count=1` 通过。
- `go test -tags=integration ./internal/repository -run "TestUsageBillingRepositoryApply_.*Voucher|TestUsageBillingRepositoryApply_RequireBalanceCheck" -count=1` 通过。
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"` 通过。
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/components/payment/__tests__/currency.spec.ts"` 通过。
- 额外 OpenAI/Codex、payment、admin settings、repository 定向测试已在本轮收口前通过。
