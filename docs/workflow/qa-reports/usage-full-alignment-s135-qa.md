### PASS: usage-full-alignment-s135

# QA Report

## Findings

- 未发现与 S135 合同直接相关的实现缺陷。用户 Usage 过滤链路、用户错误请求归属/脱敏、设置 fail-closed、路由顺序和 migration 200 均通过聚焦验收。
- 用户错误 DTO 已收紧为白名单字段；不包含账户、上游地址、用户邮箱、API Key 前缀、重试控制、客户端 IP、User-Agent 或内部 owner/source。
- 完整四包测试命令仍命中两处既有基线失败：`group_peak_rate_test.go` 的峰值时区/倍率断言，以及 `auth_rate_limit_test.go` 的 Redis 不可用路由 panic。相关文件不在本 Sprint 改动范围，未扩展修复。
- `go test -tags=unit ./internal/server -run '^TestAPIContracts$'` 仍因历史精确 payload 漂移失败；本 Sprint 新增的 `allow_user_view_error_requests=false` 已补入期望值，剩余差异是 `allow_live`、audit/group-buy/session 等既有字段。

## Executed Checks

- `go test ./internal/service -run 'Test(UserError|IsUserErrorViewAllowed|SettingService_GetPublicSettings)' -count=1`：PASS。
- `go test ./internal/repository -run 'Test(BuildOpsErrorLogsWhere|OpsErrorLogsOrderByWhitelist|ValidateMigrationExecutionMode|LatestMigrationBaseline)' -count=1`：PASS。
- `go test ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`：PASS。
- `go test ./internal/server/routes -run 'TestUserUsageStaticRoutesRegisteredBeforeIDWildcard' -count=1`：PASS。
- `go test ./migrations -count=1`：PASS。
- `go test -run '^$' ./...`：PASS，全仓库编译探针通过。
- `go test ./internal/service ./internal/handler ./internal/repository ./internal/server/routes`：FAIL，失败仅为上述既有 group peak 和 auth rate-limit 测试。
- `go test -tags=unit ./internal/server -run '^TestAPIContracts$' -count=1`：FAIL，新增设置字段期望已对齐，剩余为既有 payload 漂移。
- `gofmt`：PASS，所有变更 Go 文件已格式化。
- `git diff --check`：PASS；冲突标记扫描：PASS；允许路径审计：PASS。
- migration 200 内容检查：`CREATE INDEX CONCURRENTLY IF NOT EXISTS`、`user_id, created_at DESC`、非事务 `_notx.sql` 命名及 `user_id IS NOT NULL` 部分索引均符合合同。

## Unverified Risks

- 未执行真实 PostgreSQL migration、生产数据库、部署、容器更新、浏览器/API 登录态 smoke、commit 或 push。
- 本 Sprint 按合同未引入删除 API Key owner 快照字段；历史上 `user_id IS NULL` 的删除 Key 认证失败记录不会被用户侧归属恢复，需后续独立 schema contract。
- 前端用户 Usage 分析页、用户错误表、管理员错误请求表、管理员排行 UI 仍属于 S136/S137，尚未合入本 Sprint。

## Contract Compliance

- S135 允许路径内完成后端 Usage 过滤、用户错误请求列表/详情、设置 DTO/持久化、Wire/routes 和 migration 200；未修改 `frontend/**`、既有 migration、部署或容器。
- 用户错误列表强制 `user_id`、`View=all`、排除 `count_tokens`，并使用模型模糊匹配及排序白名单；详情对非拥有者返回 NotFound 语义。
- `allow_user_view_error_requests` 默认关闭，设置读取失败 fail closed，并已接入 admin/public settings 及 SSR injection。

## Recommendation

`PASS / source-level`。S135 可以进入单独授权的本地提交/合并步骤；暂不 push、部署或执行生产 migration。合并后按 S136（用户 Usage 页面与错误表）再按 S137（管理员错误请求、用户排行和完整 UI）分阶段集成，并保留上述历史测试和删除 Key 归属风险。
