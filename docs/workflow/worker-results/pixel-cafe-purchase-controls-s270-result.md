### DONE: pixel-cafe-purchase-controls-s270

# Worker Result

## Task ID
pixel-cafe-purchase-controls-s270

## Status
`done`

## Summary
- 公共 Room Plan DTO 直接投影现有每份 Key 总额度、5H/1D/7D 限额与说明；购买弹窗按原值展示，零值显示“不限”，30 天仍只表示有效期。
- 公共大厅和后台默认顺序改为 `sort_order ASC, id ASC`；后台“推荐”列与开关改为数字优先级，保留 `featured` 字段兼容旧 API。
- 后台团次按钮由服务端 `current_round_status` 驱动：无活动团次可开团，`open` 可暂停，后续商业状态只显示不可重复操作的状态文案。
- 新增 Cafe 专属暂停接口。事务按订单创建相同的 Room -> Round 锁顺序执行，只允许关闭无占用计数、无 locked/paid/active/refund Seat、无付费/预留 Membership 的 `open` 团次；成功写入现有 `cancelled`、`closed_at`、`close_reason`，不改订单和退款，并可重新开团。

## Changed Files
- `backend/internal/service/cafe_public.go`
- `backend/internal/service/cafe_public_test.go`
- `backend/internal/service/cafe_room_service.go`
- `backend/internal/service/cafe_room_service_test.go`
- `backend/internal/repository/cafe_room_repo.go`
- `backend/internal/repository/cafe_room_owned_plan_repo_test.go`
- `backend/internal/repository/cafe_room_repo_lock_test.go`
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

## Commands Run
```text
cd backend; go test ./internal/service -run 'TestCafePublic|TestCafeRoom' -count=1 -> PASS
cd backend; go test ./internal/repository -run 'TestCafeRoom' -count=1 -> PASS
cd backend; go test ./internal/handler/admin -run 'TestCafeRoom' -count=1 -> PASS
cd backend; go test ./cmd/server -run '^$' -count=0 -> PASS
cd frontend; npm.cmd exec -- vitest run src/features/pixelCafe/__tests__/PixelCafePage.spec.ts src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts -> PASS (28/28)
cd frontend; npm.cmd run typecheck -> PASS
cd frontend; npm.cmd run build -> PASS (1904 modules)
git diff --check -> PASS
```

## Risks
- 未连接共享/生产 PostgreSQL，也未创建测试容器；并发安全依据同一事务中的 Room -> Round 行锁顺序、SQLite repository 行为测试和 server 编译证据。正常部署前仍应由常规 PostgreSQL 流水线覆盖接口 smoke。
- 前端 build 保留仓库既有 Browserslist 过期、动态/静态 import 和大 chunk 警告；构建退出码为 0，本任务未扩展处理这些全局告警。

## Knowledge Candidates
- none

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`（未触碰 Ent schema/migration、支付退款实现、容器、共享数据、S265 worktree 或 `outputs/**`）
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- none
