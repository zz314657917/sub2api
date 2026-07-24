### PASS: group-buy-lifecycle-refund-hardening-s110

# QA Report

## Review Type

S110 提测前验收与 evidence-first 代码复核，范围为当前隔离 worktree 的拼团生命周期、退款、快照、管理员 API/UI 和用户活动 DTO。

## Findings

- 最终复审发现并修复 lifecycle `Stop` 无法取消运行中任务的问题；阻塞操作现在会收到 context cancellation，不再突破服务器清理预算。
- 补齐 S110 原先缺失的渠道退款直接测试：success/idempotency、pending query/reconciliation、failure/retry 均通过窄接口 stub 验证。
- 修复后未发现剩余的明确 P1/P2 阻断问题。
- 复核中发现并修复一个原路退款并发边界：`processing` 记录配合仍为 `completed` 的支付订单不能再次进入渠道退款；现在会进入 `needs_review`，避免并发重试覆盖首个请求状态。
- 业务边界保持不变：份额和分组模板仍分离；3 个用户各买 2 份是同一团次的 3 条购买批次、合计 6 份，不是 6 个独立分组；10 份上限未扩容。

## Executed Checks

```text
go test ./internal/service -run "TestGroupBuy|TestPayment.*Refund" -count=1 -> PASS (includes provider refund success/pending/failure/retry)
go test ./internal/service -run "TestGroupBuyLifecycleService" -count=1 -> PASS
go test ./internal/handler/admin -run "TestGroupBuy" -count=1 -> PASS
go test ./cmd/server -run "TestProvideCleanup_WithMinimalDependencies_NoPanic" -count=1 -> PASS
go test ./internal/service ./internal/handler/admin ./internal/server/routes ./cmd/server -run '^$' -> PASS
go test ./internal/service -run "TestGroupBuy(ProviderRefund|LifecycleService)" -count=10 -> PASS
go vet ./internal/service ./cmd/server -> PASS
corepack.cmd pnpm exec vitest run src/views/admin/group-buy/__tests__/AdminGroupBuyView.spec.ts src/views/user/__tests__/GroupBuyView.spec.ts -> PASS (2 files / 6 tests)
corepack.cmd pnpm run typecheck -> PASS
corepack.cmd pnpm run build -> PASS (1090 modules)
gofmt -d <S110 Go paths> -> PASS
git diff --check -> PASS
git diff --name-only --diff-filter=U -> PASS
S110 allowlist audit -> PASS
```

Manual/browser checks:

- `http://127.0.0.1:5175/admin/group-buy` loads the built Vue app and redirects unauthenticated access to `/login?redirect=/admin/group-buy`.
- Temporary dev server and browser session were stopped; port `5175` is free.
- Browser page snapshot was captured; authenticated admin detail/refund interaction was not attempted without an authorized session.

## Unverified Risks

- Full `go test ./internal/service -count=1` remains red on pre-existing peak-rate timezone/peak multiplier assertions and worker pool tests; these failures are outside S110 and the S110 focused suite is green.
- Aggregate admin `go vet` remains red on the pre-existing self-assignment in `internal/handler/admin/admin_service_stub_test.go:356`; the S110 handler tests and package compile gate pass.
- Real provider calls were not sent to Stripe/Airwallex/WeChat; the S110 state machine is covered with deterministic stubs and production still routes through the existing payment refund pipeline.
- No authenticated browser smoke for admin participant details, cancelled-round refund action, or pending/failed result rendering.
- No production migration, deployment, container refresh, or live payment reconciliation was performed.

## Recommendation

`PASS / source-only`: S110 implementation and focused acceptance evidence are ready for scoped branch publication. Do not treat this as production-release proof until an authorized admin browser smoke and live-provider sandbox verification are run; keep the existing full-service baseline failures tracked separately.
