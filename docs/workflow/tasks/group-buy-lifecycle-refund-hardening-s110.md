<!-- codex:pge-contract -->
# S110 拼团生命周期与退款加固 Contract

## Task ID

`group-buy-lifecycle-refund-hardening-s110`

## Role

`Generator`

## Goal

修复拼团团次超时、权益过期和原路退款没有真实后台执行的问题，确保余额退款原子且历史不确定记录不会被重复入账。同步收敛用户活动数据，并让管理员能查看团次参与者、订单和退款状态。

## Success Criteria

- 每 60 秒运行的拼团生命周期服务会处理超时团、过期权益和待确认的渠道退款，并在应用退出时停止。
- `provider_refund` 使用既有 `PaymentService.PrepareRefund` / `ExecuteRefund` 发起真实渠道退款；渠道 pending 由后台查询并同步到拼团退款与份额状态。
- 余额退款在一个数据库事务中完成余额、订单、退款记录、份额和事件更新；历史 `processing` 余额退款标为 `needs_review`，禁止自动重放。
- 新订单快照包含有效期和退款模式；成团有效期、累计份额档位和退款模式优先使用购买快照，历史缺失字段只走明确兼容回退。
- 管理端允许 `failed` 与 `cancelled` 团次处理退款，批量结果区分成功、待确认和失败，并可查看参与用户、订单与退款明细及退款汇总。
- 用户 `/group-buy/activity` 只返回 `id`、`event_type`、`message`、`created_at`。
- 后端定向测试、Wire 构建测试、管理端与用户端 Vitest、类型检查、生产构建、浏览器 smoke、格式和 diff 门禁通过。

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/group-buy-lifecycle-refund-hardening-s110`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`
- Existing payment source of truth: `backend/internal/service/payment_refund.go`

## Allowed Paths

- `backend/internal/domain/group_buy.go`
- `backend/internal/service/group_buy.go`
- `backend/internal/service/group_buy_test.go`
- `backend/internal/service/group_buy_lifecycle_service.go`
- `backend/internal/service/group_buy_lifecycle_service_test.go`
- `backend/internal/service/wire.go`
- `backend/internal/handler/group_buy_handler.go`
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
- `frontend/src/views/user/__tests__/GroupBuyView.spec.ts`
- `docs/workflow/**`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/schema/**`、`backend/migrations/**` 和生成 Ent 文件。
- 支付渠道实现、普通充值/订阅退款语义、计费、账号调度和 API Key 路由。
- 拼团总份额或单用户份额上限扩容；20 人团另开 Sprint。
- 部署、容器、生产配置、提交和推送。
- `knowledge/05-current-focus.md` 与全局 memories。

## Constraints

- 支付订单状态是原路退款的渠道真源；不得维护第二套渠道退款状态机。
- 外部退款调用不得包在数据库事务内；并发和重试依赖支付订单条件更新保证只发起一次。
- 已有 `processing` 余额退款无法证明是否加款，必须转人工复核，禁止自动重放。
- 已购买权益按购买快照保护；多个有效批次累计份额，并使用最近激活批次的有效策略快照解析档位。
- 团次超时或手动关闭只产生待退款；仍需管理员点击后发起退款。
- 不做无关重构、格式化或文案迁移。

## Acceptance Commands

```text
go test ./internal/service -run "TestGroupBuy|TestPayment.*Refund" -count=1
go test ./internal/service -run "TestGroupBuyLifecycleService" -count=1
go test ./internal/handler/admin -run "TestGroupBuy" -count=1
go test ./cmd/server -run "TestProvideCleanup_WithMinimalDependencies_NoPanic" -count=1
go test ./internal/service ./internal/handler/admin ./internal/server/routes ./cmd/server -run "^$"
corepack.cmd pnpm exec vitest run src/views/admin/group-buy/__tests__/AdminGroupBuyView.spec.ts src/views/user/__tests__/GroupBuyView.spec.ts
corepack.cmd pnpm run typecheck
corepack.cmd pnpm run build
gofmt -d <S110 Go paths>
git diff --check
git diff --name-only --diff-filter=U
```

## Output

- `docs/workflow/worker-results/group-buy-lifecycle-refund-hardening-s110-result.md`
- `docs/workflow/qa-reports/group-buy-lifecycle-refund-hardening-s110-qa.md`
- Worker report 第一行使用 `### DONE: group-buy-lifecycle-refund-hardening-s110`。
- QA report 第一行使用 `### PASS:`、`### FAIL:` 或 `### BLOCKED:`。

## Stop Rules

- 若原路退款必须修改支付渠道实现、迁移数据库或绕过支付订单状态机，停止并回到 Planner。
- 若无法区分历史余额 `processing` 是否已入账，不得自动重试；保留 `needs_review`。
- 若实现需要把团次上限提高到 20、改变既有份额语义或自动触发退款，停止并拆分新 Sprint。
- 若测试会请求真实支付渠道、部署或刷新容器，停止并改用 provider stub/source-only 验收。

## Contract Review

- `PASS`：当前代码仍存在 contract 所列缺口，既有支付退款状态机和 Runner/Wire 模式可复用。
- JSON 快照扩展不改变数据库列，无 Ent codegen 或 migration 前置。
- 管理 API 是现有 `/admin/group-buy` 资源的窄扩展；用户活动 DTO 收敛是向更少字段的安全变更。
