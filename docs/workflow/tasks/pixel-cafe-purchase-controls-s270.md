---
task_id: pixel-cafe-purchase-controls-s270
phase: contract-approved
qa_mode: runtime
---

# Task Contract

## Task ID
pixel-cafe-purchase-controls-s270

## Role
Codex 作为 Planner、Generator 与最终 Evaluator，在当前主工作区按本合同实现。S265 的 Terra worker 仍因模型访问 404 独立阻塞；本任务不调用或修改该 worker/worktree。

## Goal
修正 Pixel Cafe 房间购买与后台操作的信息不对称：购买前显示实际 Key 额度；用已有 `sort_order` 取代“推荐”排序语义；让开团操作按真实团次状态切换，并允许管理员安全暂停一个尚无参与者的 open 团次。

## Success Criteria
- 公共 Room Plan DTO 返回每份 Key 总额度、5H/1D/7D 限额和额度说明；购买弹窗明确展示每份总额度与 7D 限额，零值显示“不限”，不得把 30 天总额度伪装成会自动重置的月限额。
- 后台房间列表把“推荐”列改为“优先级”，编辑弹窗使用 `sort_order` 数字输入并说明数值越小越靠前；公共大厅和后台默认排序均按 `sort_order ASC, id ASC`，保留旧 `featured` 字段/API 兼容但不再参与默认顺序。
- 没有活动团次时显示“开团”；`open` 时显示“暂停”；`awaiting_account`、`activating`、`active`、`refunding` 显示对应不可重复操作的状态文案。
- 暂停接口在事务中锁定当前 Room round，仅允许暂停 `open` 且没有 locked/paid/active/refund Seat 或付费 Membership 的空团；成功后将团次终结为 `cancelled` 并允许重新开团。存在参与者、支付锁或非 open 状态时拒绝，不触发退款、不修改订单。
- focused Go/Vitest、完整受影响编译、前端 typecheck/build、diff/范围检查通过；若执行浏览器验收，必须使用任务专属 profile 并完成进程清理。

## Context
- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`, `knowledge/tasks/current-task.md`
- Current worktree: `main` ahead of `origin/main`; only `outputs/**` is untracked and must remain untouched.
- Existing `GroupBuyRoundStatusCancelled` and Room transaction/row-lock owners are reusable; no新状态或数据库迁移需要引入。

## Allowed Paths
- `backend/internal/service/cafe_public.go`
- `backend/internal/service/cafe_public_test.go`
- `backend/internal/service/cafe_room_service.go`
- `backend/internal/service/cafe_room_service_test.go`
- `backend/internal/repository/cafe_room_repo.go`
- focused Cafe room repository tests under `backend/internal/repository/`
- `backend/internal/handler/admin/cafe_room_handler.go`
- `backend/internal/handler/admin/cafe_room_handler_test.go`
- `backend/internal/server/routes/admin.go`
- `frontend/src/types/pixelCafe.ts`
- `frontend/src/api/admin/cafeRooms.ts`
- `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`
- `frontend/src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts`
- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/i18n/locales/zh/admin/pixelCafe.ts`
- `frontend/src/i18n/locales/en/admin/pixelCafe.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/pixel-cafe-purchase-controls-s270.md`
- `docs/workflow/worker-results/pixel-cafe-purchase-controls-s270-result.md`
- `docs/workflow/qa-reports/pixel-cafe-purchase-controls-s270-qa.md`

## Denied Paths
- Ent schema/generated files and migrations.
- Payment/refund execution, billing, account scheduling, provider traffic, shared/production data, containers/images/Compose, deployment, commit and push.
- Generic Group Buy close/refund behavior outside Cafe Room ownership.
- `outputs/**`, unrelated dirty files, S265 worktree/branch, `knowledge/**` and global memories.

## Constraints
- 保持最小范围，不做无关格式化或字段删除；`featured` 只降级为兼容字段，不做迁移清理。
- “暂停”只代表终结一个空的 open 团次，恢复时创建新团次；不得静默暂停已有付款或支付中的团次。
- 暂停与创建订单必须依赖同一 Round 行锁/事务顺序，避免检查后并发进入 Seat。
- 商业锁状态仍是 `open`、`awaiting_account`、`activating`、`active`、`refunding`；优先级保持可编辑。
- 不把额度展示字段用于新的计费或执行判断，执行仍以现有 Round snapshot/managed Key 字段为准。

## Acceptance Commands
```powershell
Set-Location backend
go test ./internal/service -run 'TestCafePublic|TestCafeRoom' -count=1
go test ./internal/repository -run 'TestCafeRoom' -count=1
go test ./internal/handler/admin -run 'TestCafeRoom' -count=1
go test ./cmd/server -run '^$' -count=0

Set-Location ../frontend
npm.cmd exec -- vitest run src/features/pixelCafe/__tests__/PixelCafePage.spec.ts src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts
npm.cmd run typecheck
npm.cmd run build
```

## Output
- 按 worker-result/qa-report 模板记录 changed files、真实命令、结果、风险和 contract compliance。
- 最终 Evaluator 必须复核暂停事务边界、公共 DTO 无敏感字段扩张、优先级顺序和工作区保护。

## Stop Rules
- 如果安全暂停必须处理已有支付、退款或新增数据库状态/字段，停止并重新拆合同。
- 如果订单创建与暂停不能共享可证明的 Round 行锁顺序，停止实现暂停接口。
- 如果发现 `outputs/**` 或当前已有提交内容需要覆盖/回滚，停止并报告。
- 浏览器进程归属不明确时停止清理，只报告精确 PID、命令行和 profile。

## Contract Review

### PASS: pixel-cafe-purchase-controls-s270

- 额度字段已经是 Room Plan 的现有执行真值，公共 DTO 仅需增加非敏感商业投影，前端不需要推导月度重置。
- `sort_order` 已被后台请求使用且可在活动团次期间编辑，足以替代推荐排序；保留 `featured` 可避免 API 破坏。
- Room repository 的 `CreateOpenRound` 已使用事务和 Round/Room 锁；同一 owner 可在不触碰通用退款路径的前提下原子终结空 open 团次。
- 空团暂停拒绝任何 Seat/Membership，因而无需退款、订单改写或 schema 变更；成功后使用既有 `cancelled` 终态即可重新开团。
