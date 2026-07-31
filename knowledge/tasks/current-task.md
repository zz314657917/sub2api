# 当前任务快照

最后更新：2026-07-31 10:50 +08:00

## 背景

- 用户确认截图所示的“分组复制”功能需要完整合入，而不是只加前端按钮。
- S133 在隔离 worktree `E:/codex-worktrees/sub2api/group-duplicate-s133` 将上游
  `9fc006546` 适配到本地 `main@93b6c75c`；不推送、不部署、不执行生产迁移。
- 功能提交 `05f55d18e` 已通过普通 merge 合入本地 `main`，合并提交为 `43c4ce254`。

## 当前目标

- 保持本地 `main` 的 S133 变更可追溯，等待可选的非生产 PostgreSQL/runtime smoke；
  主工作区未跟踪 `outputs/` 和并行 S132 worktree 不受影响。

## 本次已完成

- 新增管理员 `POST /api/v1/admin/groups/:id/duplicate`、列表复制动作和中英文反馈。
- 副本会保留本地所有已持久化分组配置和合格账号优先级，强制 `inactive`；同一
  Idempotency-Key 可恢复已提交副本。
- 使用迁移 `199_group_duplicate_operation_id.sql`，并重新生成 Ent/Wire。
- 已完成 focused Go 回归、前端 9 项 Vitest、前端 typecheck/build、全仓 Go 编译探针和
  build、格式与 Git 完整性检查；QA 报告为
  `docs/workflow/qa-reports/group-duplicate-s133-qa.md`。

## 已确认事实

- 主工作区当前仅有用户所有的未跟踪 `outputs/`；本次不会将其加入提交。
- 没有遗留 `CHERRY_PICK_HEAD`；S133 功能已在本地 `main`，隔离分支仍保留作回溯。
- 本地 `main` 当前比 `origin/main` 前进 2 个提交，未执行推送。
- `duplicate_operation_id` 为内部字段，不会进入公开 API DTO；部分唯一索引只约束未软删行。

## 待验证点

- 动作：修复既有 API contract 和 repository integration 测试漂移后重跑广义套件 ->
  验证：`TestAPIContracts`、PostgreSQL 事务回滚测试可实际通过。
- 动作：在非生产 PostgreSQL 环境应用迁移并调用复制接口 -> 验证：副本状态、配置、账号
  优先级、scheduler outbox 和同一幂等键恢复均符合 S133 合同。

## 当前结论

- `PASS / local-only`：S133 已合入本地 `main`（`43c4ce254`）。无明确实现缺陷、冲突、
  越界文件或未合并索引项。
- 未验证真实 PostgreSQL 迁移/事务、认证态浏览器和生产运行态；没有推送、部署、容器更新
  或外部提供商调用。

## 下一步

1. 后续获得数据库测试环境时执行真实复制 smoke -> 验证：迁移、绑定优先级、outbox 和
   幂等恢复的持久化结果。
2. 如需回收隔离 worktree/分支，先获得明确清理授权 -> 验证：仅删除 S133 临时链接和
   分支引用，不影响主工作区依赖与并行 S132 worktree。

## 验证记录

- `go generate ./ent`、`go generate ./cmd/server`、服务/管理员 handler 定向测试：PASS。
- `go test ./... -run '^$'`、`go build ./...`、前端 Vitest 9/9、typecheck 和 build：PASS。
- `git diff --check`、暂存检查、冲突标记和未合并索引：PASS。
- 受既有源码漂移阻断：`go test -tags=unit ./internal/server -run '^TestAPIContracts$'` 和
  `go test -tags=integration ./internal/repository ...`；详见 S133 QA 报告。
