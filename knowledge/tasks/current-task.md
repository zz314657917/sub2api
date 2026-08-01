# 当前任务快照

最后更新：2026-08-01 14:28 +08:00

## 背景

- S135 后端基础已提交为 `d48370f75`，S136 用户 Usage 前端已提交为 `4f4d61008`。
- S137 位于隔离 worktree `E:/codex-worktrees/sub2api/usage-admin-frontend-s137`，分支为 `codex/usage-admin-frontend-s137`。
- 主工作树存在重叠用户改动，本轮未 merge、stash、回滚、push、部署或更新容器。

## 当前目标

- 管理员 `/admin/usage` 三 Tab、错误请求工作台和用户 Token 排行已完成并在隔离分支形成一个本地提交。
- 后续仅在用户明确授权后评估主工作树集成、push 或部署。

## 本次已完成

- 管理员 Usage 增加用量明细、错误请求和用户 Token 排行三个 Tab；错误和排行均按需懒加载并防止旧响应覆盖新请求。
- 错误请求接入日期、用户、API Key、账号、模型、分组、类型、分类和状态筛选，支持服务端稳定排序、分页、独立列设置、移动卡片、详情和用户操作。
- 用户排行返回输入、输出、缓存和总 Token，支持请求数、各 Token 分段、实际消费排序以及 Top 20/50/100/200。
- 排行用户下钻会切回用量明细、写入 `user_id`、恢复用户标签并刷新共用分析和表格。
- Ops 原弹窗保持默认不可排序；管理员 Usage 显式开启服务端排序，避免共享组件产生无效排序交互。
- 管理员错误 handler 对用户/API Key 身份、分类和排序参数做校验；错误列表与排行 SQL 均使用固定排序白名单。

## 已确认事实

- 首屏不会请求 `/admin/ops/errors` 或 `/admin/dashboard/user-breakdown`。
- Playwright 请求包含错误 `status_codes=400`、`sort_by=status_code`、`page=2`，排行 `sort_by=input_tokens&limit=100`，以及下钻后的 `user_id=9`。
- 桌面 `1440x1000` 和移动 `390x844` 均无页面级横向溢出；最终控制台 0 error / 0 warning。
- 改动未触及 schema、migration、route、permission、authentication、deployment、container 或 production configuration。

## 待验证点

- 真实管理员后端 smoke -> 验证：真实数据下切换三个 Tab、组合筛选、排序、翻页、详情和用户下钻。
- PostgreSQL 大数据量性能 -> 验证：错误和排行查询的真实执行计划与响应时间。
- 部署/容器/生产 migration -> 仅在后续明确授权后执行；S137 本身没有 migration。

## 当前结论

- S137 最终裁决为 `PASS / source-level + mocked-browser`，详见 `docs/workflow/qa-reports/usage-admin-frontend-s137-qa.md`。
- 当前分支只保留一个本地功能提交；不得直接合入脏主工作树、push、部署或更新容器。

## 下一步

1. 等待明确的主工作树集成授权 -> 集成前重新检查主工作树重叠文件、基线关系和冲突风险。
2. 如授权 push -> 先做 fresh publication preflight，再只推指定分支或用户指定目标。
3. 如授权部署 -> 单独走容器锁、构建、替换、健康检查和真实管理员 smoke。

## 验证记录

- Handler Go：`go test ./internal/handler/admin -run 'Test(GetUserBreakdown|OpsErrorHandler)' -count=1` PASS。
- Repository Go：`go test ./internal/repository -run 'TestGetUserBreakdownStats' -count=1` PASS。
- 聚焦 Vitest：4 files、18/18 tests PASS。
- typecheck、changed-file ESLint、production build PASS；build 共 1109 modules。
- Playwright mock：桌面/移动、懒加载、筛选、排序、分页、列、详情、排行和用户下钻 PASS。
- `git diff --check`、冲突标记、未合并索引、allowlist、禁止触面、Ops 调用兼容和排序白名单检查 PASS。
