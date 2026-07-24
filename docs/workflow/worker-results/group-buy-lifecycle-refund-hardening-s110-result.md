### DONE: group-buy-lifecycle-refund-hardening-s110

# Worker Result

## Status

`done / source-only`

## Summary

- 拼团快照新增 `validity_days`、`refund_mode`，成团有效期、累计份额档位和退款方式优先读取购买时快照。
- 新增 60 秒 `GroupBuyLifecycleService`，负责超时团次、过期权益和待确认原路退款对账，并接入 Wire 与应用清理。
- 生命周期停止会取消正在执行的操作，不会让单次 30 秒任务突破服务器清理预算。
- 余额退款改为单事务，余额、订单、退款记录、份额和事件一起提交；历史不确定的 `processing` 余额退款转为 `needs_review`，禁止自动重放。
- 原路退款复用 `PaymentService.PrepareRefund` / `ExecuteRefund`，渠道 pending 由既有订单状态和查询链路收口；原路退款历史 `processing` 且订单仍为 `completed` 时隔离为人工复核，避免并发重复发起。
- 管理端增加退款汇总、参与者/订单/退款明细，`failed` 与 `cancelled` 团次均可处理退款，批量结果区分成功、待确认和失败。
- 用户活动 DTO 收敛为 `id`、`event_type`、`message`、`created_at`。

## Changed Files

- `backend/internal/domain/group_buy.go`
- `backend/internal/service/group_buy.go`
- `backend/internal/service/group_buy_lifecycle_service.go`
- `backend/internal/service/group_buy_test.go`
- `backend/internal/service/group_buy_lifecycle_service_test.go`
- `backend/internal/service/wire.go`
- `backend/internal/handler/admin/group_buy_handler.go`
- `backend/internal/handler/admin/group_buy_handler_test.go`
- `backend/internal/server/routes/admin.go`
- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`
- `frontend/src/api/admin/groupBuy.ts`
- `frontend/src/types/groupBuy.ts`
- `frontend/src/views/admin/group-buy/AdminGroupBuyView.vue`
- `frontend/src/views/admin/group-buy/__tests__/AdminGroupBuyView.spec.ts`
- `docs/workflow/**` S110 contract/status/spec/result/QA evidence

## Tests

```text
go test ./internal/service -run "TestGroupBuy|TestPayment.*Refund" -count=1 -> PASS (includes direct provider refund state-machine coverage)
go test ./internal/service -run "TestGroupBuyLifecycleService" -count=1 -> PASS
go test ./internal/handler/admin -run "TestGroupBuy" -count=1 -> PASS
go test ./cmd/server -run "TestProvideCleanup_WithMinimalDependencies_NoPanic" -count=1 -> PASS
go test ./internal/service ./internal/handler/admin ./internal/server/routes ./cmd/server -run '^$' -> PASS
corepack.cmd pnpm exec vitest run <admin/user group-buy specs> -> PASS (2 files / 6 tests)
corepack.cmd pnpm run typecheck -> PASS
corepack.cmd pnpm run build -> PASS (1090 modules)
gofmt -d <S110 Go paths> -> PASS (no output)
git diff --check -> PASS
git diff --name-only --diff-filter=U -> PASS (no output)
S110 allowlist audit -> PASS
```

## Regression Coverage Added

- lifecycle 立即执行、tick 执行、错误隔离和 `Stop` 停止回收。
- 阻塞中的 lifecycle 操作会在 `Stop` 时收到取消并及时退出。
- 原路退款成功幂等、渠道 pending 后台对账、渠道失败后重试。
- 最近激活批次的购买策略快照优先于当前计划。
- 历史余额 `processing` 隔离到 `needs_review` 且不重复入账。
- 余额入账后故意失败时事务回滚余额、退款记录和份额状态。
- 历史原路退款 `processing` + 订单 `completed` 在重新发起前隔离为人工复核。
- 管理员团次无效 ID 在访问 service 前返回 `INVALID_ID`。

## Known Baseline / Risks

- `go test ./internal/service -count=1` 仍复现仓库已有的 `group_peak_rate` 时区/峰值断言和 worker pool 基线失败；S110 focused tests 未复现这些失败。
- 浏览器 smoke 能打开前端并把未登录 `/admin/group-buy` 正确导向 `/login?redirect=/admin/group-buy`；当前没有授权管理员登录态，因此未验证真实团次列表、详情弹窗和退款点击链路。
- 渠道状态机使用测试 stub 验证；未向 Stripe、Airwallex 或微信支付发送真实退款请求。
- 未请求真实支付渠道、未部署、未刷新容器、未执行迁移、未提交或推送。

## Contract Compliance

- allowed paths only: `yes`
- denied paths touched: `no`
- schema/migrations changed: `no`
- 20-share expansion: `no`
- automatic refund initiation: `no`
- deployment/container refresh: `no`
