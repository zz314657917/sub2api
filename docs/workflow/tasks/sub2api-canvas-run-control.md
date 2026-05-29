---
task_id: sub2api-canvas-run-control
owner: developer-worker
repo: F:/mcplugins/sub2api
---

# Task Contract

## Task ID
sub2api-canvas-run-control

## Role
你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。你不是唯一开发者，不得回滚或覆盖他人改动；如遇到并行改动，必须在允许范围内适配。

## Goal
补齐 Canvas 运行队列和取消能力的后端/API-client 基础，让前端可以调用现有 Canvas run cancel 路由，并让 run 数据完整表达 canceled 时间与状态。后端已有 cancel handler/service/repository 时，不重复造新入口，只补缺口和测试。

## Success Criteria
- `frontend/src/api/canvas.ts` 导出 `cancelCanvasRun(id: string): Promise<CanvasRun>`，调用 `POST /user/canvas-runs/:id/cancel` 并复用现有 run mapper。
- `CanvasRun` 前端类型和 mapper 保留 `canceled_at` 字段，`pending` 仍归一为 `queued`，`canceled` 保持 `canceled`。
- Canvas 后端 cancel 行为若已有，补齐或修正测试覆盖：当前用户隔离、pending/running 可取消、已 canceled 幂等返回、终态 succeeded/failed 不能取消并返回冲突。
- 不改 `CanvasView.vue`；本任务只提供前端可调用 API 和后端稳定性保障。

## Context
- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, `knowledge/tasks/current-task.md`
- Current facts:
  - `backend/internal/handler/canvas_handler.go` 已有 `CancelRun` handler。
  - `backend/internal/service/canvas_service.go` 已有 `CancelRun` 方法。
  - `backend/internal/repository/canvas_repo.go` 已有 `CancelRun` SQL 更新迹象。
  - `backend/internal/server/routes/user.go` 已有 `canvasRuns.POST("/:id/cancel", h.Canvas.CancelRun)`。
  - `frontend/src/api/canvas.ts` 当前缺少 `cancelCanvasRun` 导出，且 `CanvasRun` 未暴露 `canceled_at`。

## Allowed Paths
- `backend/internal/handler/canvas_handler.go`
- `backend/internal/handler/canvas_handler_test.go`
- `backend/internal/service/canvas_service.go`
- `backend/internal/service/canvas_service_test.go`
- `backend/internal/repository/canvas_repo.go`
- `backend/internal/repository/canvas_repo_test.go`
- `backend/internal/server/routes/user.go`
- `frontend/src/api/canvas.ts`
- `frontend/src/api/__tests__/canvas.spec.ts`
- `docs/workflow/worker-results/**`

## Denied Paths
- `frontend/src/views/user/CanvasView.vue`
- `frontend/src/views/user/__tests__/CanvasView.spec.ts`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `C:/Users/Administrator/.codex/memories/**`
- 数据库 migration、生产配置、账号/RBAC、ImageCreator task 执行逻辑。

## Constraints
- 保持最小改动；不要格式化无关文件。
- 不新增新的 cancel URL，使用既有 `/user/canvas-runs/:id/cancel`。
- 不级联取消 ImageCreator task，除非当前代码已有明确安全接口且无需改其执行逻辑；默认只取消 Canvas run 本身。
- 后端错误语义遵循现有 infraerrors/response 模式。
- 前端 API helper 继续使用 `apiClient` 和现有 mapper 风格。

## Acceptance Commands
```powershell
go test ./internal/service ./internal/handler ./internal/repository -run "Canvas" -count=1
cd frontend
npm.cmd run test:run -- canvas
```

## Output
- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 worker report 到 `docs/workflow/worker-results/sub2api-canvas-run-control-result.md`。
- Worker report 第一行必须是 `### DONE: sub2api-canvas-run-control`、`### BLOCKED: sub2api-canvas-run-control` 或 `### FAILED: sub2api-canvas-run-control`。
- 必须列出 changed files、commands run、key test output、risks、contract compliance、knowledge_candidates。
- 不允许直接写长期知识库；只提交候选结论。

## Stop Rules
- 如果需要修改 `CanvasView.vue`、数据库 migration、生产配置、安全边界或 ImageCreator 执行逻辑，停止并报告 BLOCKED。
- 如果后端 cancel 已完整且只需前端 API helper，也只做 API helper 和测试，不扩大范围。
- 如果验收命令无法运行，报告原因和最小替代验证。

## Budget
- worker_mode: `claude-bare-deepseek-v4-pro`
- worker_model: `deepseek-v4-pro`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Worker Output
- 同 `Output`。
