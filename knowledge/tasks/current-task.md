# 当前任务快照

最后更新：2026-06-02 17:50 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前主分支 `main` 正在整理近期远端与本地批次分支，目标是把已完成的本地合并线收口到主工作线。
- `tmp-ui-check/` 仍是未跟踪截图证据目录，不属于源码提交范围。
- 本轮不直接整包合入 `upstream/main`，因为当前本地主线与 `upstream/main` 分叉较大，需要单独评估。

## 当前目标

- 完成 `codex/upstream-v0.1.133-critical-fixes` 合并，并确保之前的 `origin/main`、`batch2`、`batch3` 合并结果仍然有效。
- 合并后运行后端/前端关键验证，确认可提交。
- 保持交接快照记录为当前事实，不把旧的视频 smoke 待办覆盖为当前主目标。

## 本次已完成

- 已执行 `git fetch --all --prune`，刷新 `origin` 与 `upstream` 引用。
- 已确认当前分支为 `main`，没有挂起的旧 merge。
- 已确认 `codex/upstream-v0.1.133-batch3` 的 merge commit 已存在：`2977ef90c Merge branch 'codex/upstream-v0.1.133-batch3'`。
- 已确认 `codex/upstream-v0.1.133-batch2` 的 merge commit 已存在：`1f47bd607 Merge branch 'codex/upstream-v0.1.133-batch2'`。
- 已确认 `307e952ca feat: finalize gateway billing and ticket polish` 只在 `codex/upstream-v0.1.133-critical-fixes` 上，当前 `main` 合并前未包含。
- 已尝试合并 `codex/upstream-v0.1.133-critical-fixes`，冲突文件只有：
  - `backend/internal/server/middleware/api_key_auth_test.go`
  - `knowledge/tasks/current-task.md`
- `api_key_auth_test.go` 冲突按保留双方新增测试解析：保留未分组 key business-limited 测试，同时加入模型感知多分组 key 延迟分组计费测试。
- `current-task.md` 已重写为当前合并快照。
- 修复合并后的编译缺口：`ResolveOpenAIVideoTaskAccount` 按新签名调用 `isOpenAIAccountEligibleForRequest(ctx, account, "", false, OpenAIEndpointCapabilityChatCompletions)`。
- 已完成后端核心包、routes/server、迁移集成测试、前端 typecheck 与目标 Vitest。

## 已确认事实

- `codex/upstream-v0.1.133-batch2` 与 `codex/upstream-v0.1.133-batch3` 已进入当前 `main`。
- `664e9fdcd feat(usage): 用户用量按平台拆分 + UsersView 列设置可配置 + 用量列排序` 仍只在 `upstream/main` 上；当前本地 `batch3` 含同主题本地提交 `382287168`。
- `codex/upstream-v0.1.133-critical-fixes` 相对合并前 `main` 还有 7 个提交未进入，包含模型感知分组路由、工单后端/前端、usage filter、视频计费与工单收尾提交。
- 本轮合并带入 `.dockerignore` 与 `deploy/docker-compose.dev.yml` 改动，但尚未做 Docker build 验证。

## 待验证点

- Docker 配置改动已随合并进入暂存区，但尚未做 Docker build 验证。
- 真实上游视频 smoke 未执行；该验证会创建视频任务并消耗上游额度/余额，需要用户明确测试环境和账号后再跑。
- `tmp-ui-check/` 仍是未跟踪截图证据目录，不应纳入提交。

## 当前结论

- 当前处于 `codex/upstream-v0.1.133-critical-fixes` merge 过程中，冲突已解析，关键回归已通过。
- 下一步应完成最终 diff/status 检查，然后创建 `critical-fixes` merge commit。

## 下一步

- 执行最终 `git diff --cached --check` 和 `git status` 检查。
- 提交当前 `critical-fixes` merge commit。
- 提交后确认 `main` 包含 `307e952ca`，并确认 `tmp-ui-check/` 仍未跟踪。

## 验证记录

- 本轮合并前最近已验证：
  - 视频异步任务计费链路 service/handler/repository、routes/server、PostgreSQL 迁移集成测试通过。
  - 工单前后端目标测试和前端 typecheck 通过。
  - 公共页 Vitest 通过。
- 本轮 `critical-fixes` 合并后的统一验证：
  - `go test ./internal/service ./internal/handler ./internal/repository -count=1`：通过。
  - `go test ./internal/server/routes ./cmd/server -count=1`：通过。
  - `go test -tags=integration ./internal/repository -run TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate -count=1`：通过。
  - `corepack.cmd pnpm run typecheck`：通过。
  - 工单/公共页/KeysView/UsageFilters/Welfare 目标 Vitest 8 文件 40 用例：通过；仅有 Browserslist 数据较旧提示。
  - UsersView/Dashboard/Usage 目标 Vitest 9 文件 33 用例：通过；`UsageView.spec.ts` 的 mock 缺少 `getModelStats` 产生 stderr，但测试通过。
