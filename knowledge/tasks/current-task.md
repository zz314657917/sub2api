# 当前任务快照

最后更新：2026-07-24 12:38 +08:00

## 背景

- S110 在独立 worktree `E:/codex-worktrees/sub2api/group-buy-lifecycle-refund-hardening-s110` 完成。
- 任务修复拼团生命周期、退款一致性、购买快照、管理员可操作性和用户活动数据暴露问题。
- 主工作树存在并行修改，S110 始终保持独立边界，未触碰部署或容器。

## 当前目标

- 将 Final Evaluator 已判定 PASS 的 S110 精确提交并推送到
  `codex/group-buy-lifecycle-refund-hardening-s110`。
- 推送后验证本地 `HEAD` 与远端分支一致。

## 本次已完成

- 生命周期服务每 60 秒处理超时团次、过期权益和待确认渠道退款，并接入 Wire 启停。
- 余额退款在单事务内更新余额、订单、退款记录、份额和事件；历史不确定记录进入 `needs_review`。
- 原路退款复用现有支付退款管线，支持成功、pending 对账、失败重试和幂等收口。
- 最终复审修复 `Stop()` 无法取消运行中 lifecycle 操作的问题，并补齐对应回归测试。
- 管理端增加退款汇总、参与者/订单/退款详情和 cancelled 团退款；前端修复协议默认值与编辑弹窗误关闭。

## 已确认事实

- 购买快照包含 `validity_days` 与 `refund_mode`；有效权益按最近激活购买快照解析档位。
- `refund_pending` 与 `refund_processing` 都能进入管理员退款扫描，历史不确定记录不会自动重复入账或重复请求渠道。
- 用户活动 DTO 只返回 `id`、`event_type`、`message`、`created_at`。
- 10 份总上限、手动发起退款、数据库 schema、计费、部署和容器语义均未改变。

## 待验证点

- 动作：使用授权管理员登录态执行详情与退款按钮 smoke -> 验证：检查真实列表、弹窗、成功/pending/失败提示。
- 动作：在支付渠道 sandbox 执行真实退款 -> 验证：核对支付订单、拼团退款和份额状态最终一致。
- `go test ./internal/service -count=1` 的既有 peak-rate/worker-pool 基线失败不属于 S110，仍需独立 Sprint。

## 当前结论

- `PASS / source-only`：S110 的代码评审、定向回归、编译、前端检查和静态门禁已通过。
- 当前 P/G/E phase 为 `done`；仅剩精确暂存、提交、推送和远端一致性验证。

## 下一步

1. 动作：按 S110 allowlist 精确暂存 -> 验证：检查 cached name-only、cached diff-check 和 denied path。
2. 动作：提交并推送独立分支 -> 验证：比较 `HEAD`、`origin/codex/group-buy-lifecycle-refund-hardening-s110` 和 `git ls-remote`。
3. 动作：回写 published 证据 -> 验证：`status.md`、`main-log.md` 和本文件一致。

## 验证记录

- `go test ./internal/service -run "TestGroupBuy|TestPayment.*Refund" -count=1` PASS。
- 管理员 handler、Wire cleanup 和四个后端 package 串行编译门禁 PASS。
- 管理端/用户端 Vitest `2 files / 6 tests`、typecheck、production build `1090 modules` PASS。
- lifecycle/provider refund 高风险测试 `count=10` PASS；service 与 `cmd/server` 的 `go vet` PASS。
- admin aggregate `go vet` 仅保留未改动测试桩 `admin_service_stub_test.go:356` 的既有 self-assignment 红项。
- `gofmt -d`、`git diff --check`、冲突标记和未合并索引检查 PASS。
