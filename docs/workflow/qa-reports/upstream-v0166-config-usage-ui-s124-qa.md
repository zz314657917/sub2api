### PASS: upstream-v0166-config-usage-ui-s124

# QA Report

## Task ID
upstream-v0166-config-usage-ui-s124

## Verdict
`PASS / source-only`

## Contract Checked
- `docs/workflow/tasks/upstream-v0166-config-usage-ui-s124.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:
```text
go test ./internal/config -run "ConfigFile|GetServerAddress" -count=1 -> PASS
go test ./internal/handler/admin -run "Usage.*Request" -count=1 -> PASS
go test ./internal/repository -run "UsageLogRepository.*Request" -count=1 -> PASS
go test ./... -run "^$" -> PASS
corepack.cmd pnpm exec vitest run src/components/admin/usage/__tests__/UsageFilters.spec.ts src/views/admin/__tests__/UsageView.spec.ts -> PASS, 8/8
corepack.cmd pnpm run typecheck -> PASS
corepack.cmd pnpm run build -> PASS, 1099 modules
gofmt -d on changed Go files -> PASS
git diff --check -> PASS
conflict-marker scan -> PASS
allowed-path audit -> PASS
primary dirty-worktree overlap audit -> NONE
```

## Findings
- 未发现明确问题。
- `CONFIG_FILE` 同时覆盖完整加载和 bootstrap 地址读取，并保留原有路径和环境变量优先级。
- `request_id` 仅影响管理员 usage 列表的精确 SQL 条件；统计、导出、排序和持久化不变。
- 路由用户邮箱回显会在用户输入发生后拒绝迟到的异步写入，查询失败时仅回显 user ID。

## Unverified Risks
- 未执行真实 PostgreSQL 中的 config/usage 查询、认证态浏览器交互、部署、容器更新或远端推送。
- 前端构建仍输出既有 Browserslist、动态导入和 chunk-size 警告；构建成功，S124 未新增阻断告警。

## Bug Owner Recommendation
`none`

## Root Cause
- `none`

## Retest Scope
- 无；所有 S124 contract acceptance gates 已通过。

## Knowledge Promotion
- `none`
