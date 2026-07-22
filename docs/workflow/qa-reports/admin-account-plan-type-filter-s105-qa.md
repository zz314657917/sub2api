### PASS: admin-account-plan-type-filter-s105

## Findings

- 未发现阻断 S105 本地提交的明确问题。
- 业务 diff 均可追溯到套餐筛选、归一化展示、筛选传播、回归测试或工作流证据；
  contract allowlist 已收紧到实际改动路径。

## Executed Checks

- `go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters" -count=1`：PASS。
- `go test ./internal/service -run "TestAdminService(ListAccountsPropagatesNormalizedPlanType|BulkUpdatePropagatesPlanTypeFilter|BulkShareStatusPropagatesPlanTypeFilter)" -count=1`：PASS。
- `go test ./internal/handler/admin -run "TestListAccountsPassesPlanTypeFilter|TestExportDataPassesAccountFiltersAndSort" -count=1`：PASS。
- `go test ./internal/handler/admin ./internal/repository -count=1`：PASS。
- 聚焦 Vitest：`3 files / 13 tests` PASS。
- `corepack.cmd pnpm --dir frontend run typecheck`：PASS。
- `corepack.cmd pnpm --dir frontend run build`：PASS，`1089 modules transformed`。
- `gofmt -d`、`git diff --check`、冲突标记扫描和未合并索引检查：PASS。
- 人工检查普通列表、共享列表、按筛选批量编辑、按筛选共享状态、按筛选导出、
  ETag 和前端局部行匹配均携带同一个 `plan_type` 条件；selected-ID 操作保持优先。
- 人工检查 SQL 仅使用固定表达式和参数化值，未知筛选值在 service 层被拒绝。

## Unverified Risks

- 未使用真实 K12/Plus/Pro 账号执行登录态浏览器 smoke，也未连接生产数据库。
- `go test ./internal/service -count=1` 仍失败于既有 peak-rate 时区测试：
  `TestPeakMultiplierAt_Boundaries`、`TestPeakMultiplierAt_RespectsTimezoneLocation`、
  `TestPeakMultiplierAt_StandardTypeDegradesToOne`、
  `TestPeakMultiplier_GatewayBillingSequence` 和
  `TestPeakMultiplier_SnapshotRoundTrip`；测试启动日志显示 UTC，与 S105 路径无关。
- Browserslist 数据过期和既有 Vite chunk/dynamic-import 警告仍存在，本 Sprint
  未升级依赖或调整打包策略。

## Recommendation

- `可继续`：按精确 allowlist 创建本地提交。
- 未经单独授权不要推送、部署或更新容器。
