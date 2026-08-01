# 当前任务快照

最后更新：2026-08-01 11:51 +08:00

## 当前目标

- 完成 `usage-full-alignment-s135` 后端基础 Sprint 的验收，并为后续分阶段合并提供清晰边界。
- 当前分支：`codex/usage-full-alignment-s135`。
- 隔离 worktree：`E:/codex-worktrees/sub2api/usage-full-alignment-s135`。
- 基线：`main@1c1021133`；上游参考：`upstream/main@7ceabb3fd`。

## 本次已完成

- Usage 列表、统计、趋势、模型、分组和 `snapshot-v2` 统一继承用户安全过滤条件，并强制 API Key 所有权。
- 新增用户错误请求列表/详情 API，服务层强制 `user_id` 归属、排除 `count_tokens`、支持分类/状态/模型/日期/API Key 过滤和稳定排序。
- 用户错误响应收紧为白名单字段；详情对非拥有者返回 NotFound 语义。
- 新增 `allow_user_view_error_requests`，默认关闭，设置读取失败 fail closed，并接入 admin/public settings、SSR injection、Wire 和用户路由。
- 新增 `backend/migrations/200_add_ops_error_logs_user_time_index_notx.sql`，为用户错误分页创建幂等并发部分索引。
- 新增路由静态断言、归属/脱敏/where/order/设置测试。

## 验收结论

- QA：`PASS / source-level`，详见 `docs/workflow/qa-reports/usage-full-alignment-s135-qa.md`。
- 聚焦用户错误、设置、过滤、路由、迁移和 public settings schema 测试通过。
- `go test -run '^$' ./...` 编译探针通过。
- 完整四包测试仍有既有失败：`group_peak_rate_test.go` 峰值时区断言，以及 `auth_rate_limit_test.go` Redis 不可用路由 panic；未改动这些文件。
- `-tags=unit TestAPIContracts` 仍受既有精确 payload 漂移阻断；本 Sprint 新增的 `allow_user_view_error_requests=false` 期望已补齐。

## 尚未执行

- 未执行真实 PostgreSQL migration、生产数据库、部署、容器更新、浏览器/API 登录态 smoke 或 push。
- 删除 API Key owner 快照归属恢复未实现；当前按现有 `user_id` 强制归属，需后续独立 schema contract。
- 前端用户 Usage、用户错误表、管理员错误请求表和用户排行 UI 延后至 S136/S137。

## 下一步

1. 用户已授权本地提交并继续合入；先精确提交 S135，再从该提交创建 S136 隔离分支。主工作树存在重叠脏文档，暂不直接 merge。
2. S136：先合入用户 Usage 页面、筛选器、趋势/模型/分组图表和用户错误表，复用 S135 API。
3. S137：再合入管理员错误请求视图、用户排行/Token 排名和完整前端回归，单独评估管理员数据口径与权限。
