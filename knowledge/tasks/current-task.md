# 当前任务快照

最后更新：2026-06-02 17:30 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前主分支 `main` 正在整理最近本地和远端更新，目标是把已完成的本地批次分支合回主工作线。
- 已先 fast-forward 合入 `origin/main`，随后合入 `codex/upstream-v0.1.133-batch2`。
- 正在合入 `codex/upstream-v0.1.133-batch3`，业务代码已自动合入；仅 `knowledge/tasks/current-task.md` 发生内容冲突并已人工重写为当前快照。
- 工作区仍有未跟踪 `tmp-ui-check/` 截图目录，作为本地视觉证据，不应纳入代码提交。

## 当前目标

- 在当前 `main` 上继续合并近期本地分支：
  - `codex/upstream-v0.1.133-batch3`
  - `codex/upstream-v0.1.133-critical-fixes`
- 合并后运行后端/前端关键验证，确认可提交。
- 不直接整包合入 `upstream/main`，因为当前 `HEAD...upstream/main` 分叉很大，需要另行评估。

## 本次已完成

- `git fetch --all --prune` 更新了 `origin` 和 `upstream` 引用。
- `git merge --no-edit origin/main`：fast-forward 成功。
- `git merge --no-edit codex/upstream-v0.1.133-batch2`：成功生成 merge commit。
- `git merge --no-edit codex/upstream-v0.1.133-batch3`：业务文件自动合入，`knowledge/tasks/current-task.md` 冲突已人工解析。
- `batch3` 已包含 `664e9fdcd feat(usage): 用户用量按平台拆分 + UsersView 列设置可配置 + 用量列排序` 相关本地处理：
  - 后端 dashboard / users 用量统计增加 `by_platform`。
  - 前端用户 Dashboard 增加平台拆分卡。
  - 管理端 UsersView 增加平台用量子列、列设置版本迁移和用量排序。

## 已确认事实

- 当前 `main` 原先落后 `origin/main` 25 个提交，已通过 fast-forward 补齐。
- `upstream/main` 已刷新到 `aa69e3947`，但与当前本地线分叉较大；本轮暂不直接合并整条 `upstream/main`。
- `codex/upstream-v0.1.133-batch2` 和 `codex/upstream-v0.1.133-batch3` 都已经进入当前合并流程。
- `tmp-ui-check/` 仍是未跟踪目录，不属于应提交源码。

## 待验证点

- 继续合并 `codex/upstream-v0.1.133-critical-fixes` 后，需要检查是否与已合入的 `origin/main` / batch2 / batch3 产生冲突。
- 合并全部完成后至少运行：
  - `go test ./internal/service ./internal/handler ./internal/repository -count=1`
  - `go test ./internal/server/routes ./cmd/server -count=1`
  - `go test -tags=integration ./internal/repository -run TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate -count=1`
  - `corepack.cmd pnpm run typecheck`
  - 相关前端 Vitest：工单、公共页、UsersView/Dashboard/usage。
- 若涉及 Docker 配置最终入提交，仍建议补一次 Docker build 或至少记录未验证。

## 当前结论

- 当前处于合并中状态，业务代码冲突已清理到只剩 `current-task.md` 冲突解析；下一步应 `git add knowledge/tasks/current-task.md` 并继续/完成 `batch3` merge。
- `batch3` 合并完成后再合 `codex/upstream-v0.1.133-critical-fixes`。

## 下一步

- `git add knowledge/tasks/current-task.md`。
- 完成当前 `batch3` merge commit。
- 合并 `codex/upstream-v0.1.133-critical-fixes`。
- 运行验证并根据结果修复或提交。

## 验证记录

- 本轮合并前最近已验证：
  - 视频异步任务计费链路：后端 service/handler/repository、routes/server、PostgreSQL 迁移集成测试通过。
  - 工单前后端目标测试和前端 typecheck 通过。
  - 公共页 Vitest 通过。
- 本轮合并后的统一验证尚未执行。
