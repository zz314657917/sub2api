### PASS: pixel-cafe-purchase-controls-s270

# QA Report

## Task ID
pixel-cafe-purchase-controls-s270

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/pixel-cafe-purchase-controls-s270.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:
```text
go test ./internal/service -run 'TestCafePublic|TestCafeRoom' -count=1 -> PASS
go test ./internal/repository -run 'TestCafeRoom' -count=1 -> PASS
go test ./internal/handler/admin -run 'TestCafeRoom' -count=1 -> PASS
go test ./cmd/server -run '^$' -count=0 -> PASS; no tests to run
npm.cmd exec -- vitest run src/features/pixelCafe/__tests__/PixelCafePage.spec.ts src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts -> PASS; 28/28
npm.cmd run typecheck -> PASS
npm.cmd run build -> PASS; 1904 modules
git diff --check -> PASS
```
- focused behavior:
```text
public DTO -> 500 USD total, unlimited 5H/1D, 100 USD 7D projected without sensitive account fields
public/admin ordering -> lower sort_order wins even when the later room is featured
pause repository -> locked Seat rejected; paid Membership rejected; empty open Round becomes cancelled; a fresh open Round can then be created
admin UI -> open shows Pause; active shows disabled In use; no current Round shows Open round
purchase UI -> total/5H/1D/7D values shown; no monthly-reset wording
```

## Findings
- 未发现阻断问题。
- `cafe_room_repo_lock_test.go` 的 Plan sqlmock fixture 原为旧 32 列，当前生成模型是 35 列；已仅同步测试数据和 eager-load expectation，业务代码未因此改变。
- 未执行浏览器自动化或真实 PostgreSQL smoke；合同未要求共享数据/容器操作，且这些操作属于明确禁止范围。

## Bug Owner Recommendation
`none`

## Root Cause
- 额度缺失来自公共 DTO 未投影现有 Room Plan 限额字段。
- “推荐”语义来自默认排序仍优先 `featured`，后台仍展示兼容字段而非 `sort_order`。
- 开团按钮只依赖瞬时请求状态，没有消费 `current_round_status`，同时后端缺少空团暂停专用接口。

## Retest Scope
- 常规部署流水线在 PostgreSQL 上 smoke `POST /admin/cafe/rooms/:id/pause-round` 与并发下单互斥。

## Knowledge Promotion
- `none`
