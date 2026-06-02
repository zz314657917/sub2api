# 当前任务快照

最后更新：2026-06-02 17:59 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前主分支 `main` 已完成近期远端与本地分支合并收口。
- 本轮不直接整包合入 `upstream/main`，因为当前本地主线与 `upstream/main` 仍存在较大分叉，需要单独评估。
- `tmp-ui-check/` 仍是未跟踪截图证据目录，不属于源码提交范围。

## 当前目标

- 确认近期本地分支都已合入 `main`。
- 保留合并中发现的必要兼容修复。
- 完成关键后端/前端回归，并留下可恢复的最终快照。

## 本次已完成

- 已执行 `git fetch --all --prune`，刷新 `origin` 与 `upstream` 引用。
- 已确认并合入 `origin/main`。
- 已确认并合入 `codex/upstream-v0.1.133-batch2`：`1f47bd607 Merge branch 'codex/upstream-v0.1.133-batch2'`。
- 已确认并合入 `codex/upstream-v0.1.133-batch3`：`2977ef90c Merge branch 'codex/upstream-v0.1.133-batch3'`。
- 已确认并合入 `codex/upstream-v0.1.133-critical-fixes`：`167ae57d9 Merge branch 'codex/upstream-v0.1.133-critical-fixes'`。
- 已确认并合入 `pge/sub2api-canvas-editor-core`：`d3104f894 Merge branch 'pge/sub2api-canvas-editor-core'`。
- 已确认并合入 `pge/sub2api-canvas-run-control`：`797854f44 Merge branch 'pge/sub2api-canvas-run-control'`。
- 解析 `critical-fixes` 冲突时保留双方后端鉴权测试，并重写 `current-task.md` 为当前事实。
- 修复合并后的编译缺口：`ResolveOpenAIVideoTaskAccount` 按新签名调用 `isOpenAIAccountEligibleForRequest(ctx, account, "", false, OpenAIEndpointCapabilityChatCompletions)`。
- 解析 Canvas editor 冲突时保留主线模块化 i18n、运行取消状态、`ref` 包装的拖拽/平移状态和 pointer listener 防重复保护。

## 已确认事实

- `git branch --no-merged main` 当前为空；已知本地分支均已合入 `main`。
- `307e952ca feat: finalize gateway billing and ticket polish` 已被 `main` 包含。
- `664e9fdcd feat(usage): 用户用量按平台拆分 + UsersView 列设置可配置 + 用量列排序` 仍只在 `upstream/main` 上；当前本地 `batch3` 含同主题本地提交 `382287168`。
- `main` 当前相对 `origin/main` ahead 37。
- `tmp-ui-check/` 仍未跟踪，未纳入提交。

## 待验证点

- Docker 配置改动已随合并进入主线，但尚未做 Docker build 验证。
- 真实上游视频 smoke 未执行；该验证会创建视频任务并消耗上游额度/余额，需要用户明确测试环境和账号后再跑。
- 尚未 push 到远端。

## 当前结论

- 本轮“检查最近更新并合并其他本地分支”已完成。
- 当前 `main` 没有挂起 merge，已知本地分支全部合入；工作区只剩未跟踪 `tmp-ui-check/`。
- 关键后端、前端和 Canvas 目标回归均已通过。

## 下一步

- 如需共享结果：push 当前 `main` 或创建 PR。
- 如需清理工作区：删除或归档 `tmp-ui-check/`。
- 如需补齐剩余风险：跑 Docker build，以及在用户确认账号后跑真实上游视频 smoke。

## 验证记录

- `git diff --cached --check`：通过。
- `go test ./internal/service ./internal/handler ./internal/repository -count=1`：通过。
- `go test ./internal/server/routes ./cmd/server -count=1`：通过。
- `go test -tags=integration ./internal/repository -run TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate -count=1`：通过。
- `corepack.cmd pnpm run typecheck`：通过。
- 工单/公共页/KeysView/UsageFilters/Welfare 目标 Vitest 8 文件 40 用例：通过；仅有 Browserslist 数据较旧提示。
- UsersView/Dashboard/Usage 目标 Vitest 9 文件 33 用例：通过；`UsageView.spec.ts` 的 mock 缺少 `getModelStats` 产生 stderr，但测试通过。
- Canvas 后端目标测试：
  - `go test ./internal/repository -run Canvas -count=1`：通过。
  - `go test ./internal/service -run Canvas -count=1`：通过。
- Canvas 前端目标 Vitest：`corepack.cmd pnpm exec vitest run src/views/user/__tests__/CanvasView.spec.ts src/api/__tests__/canvas.spec.ts`：2 文件 14 用例通过；仅有 Browserslist 数据较旧提示。
