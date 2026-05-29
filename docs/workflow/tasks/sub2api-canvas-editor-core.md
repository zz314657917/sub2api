---
task_id: sub2api-canvas-editor-core
owner: developer-worker
repo: F:/mcplugins/sub2api
---

# Task Contract

## Task ID
sub2api-canvas-editor-core

## Role
你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。你不是唯一开发者，不得回滚或覆盖他人改动；如遇到并行改动，必须在允许范围内适配。

## Goal
补齐 `/canvas` 前端核心编辑能力：节点拖拽、连线创建/删除、画布缩放/平移/适配视图，并确保保存刷新后节点坐标、边和 viewport 能保持。

## Success Criteria
- 用户可以在 canvas stage 内拖拽节点，节点 `x/y` 更新并进入保存 payload。
- 用户可以选择起点节点和终点节点创建连线；重复边不重复创建；删除节点会删除相关边；可删除选中边。
- 用户可以缩放、平移、适配视图；viewport 写入 `canvasDocument.viewport` 并进入保存 payload。
- 保留现有节点参数编辑、运行按钮、ImageCreator task 轮询结果展示，不破坏现有 `CanvasView.spec.ts` 用例。
- 增加前端测试覆盖：拖拽节点保存、创建/删除连线、viewport 保存。

## Context
- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, `knowledge/tasks/current-task.md`
- Current facts:
  - `frontend/src/views/user/CanvasView.vue` 当前是单文件 Canvas 页面，已有节点列表、节点类型、保存、运行、结果轮询。
  - `frontend/src/api/canvas.ts` 已支持 document nodes/edges/viewport 映射。
  - 节点数据结构使用 `CanvasNode.x/y/width/height`；边使用 `CanvasEdge.source_node_id/target_node_id`。

## Allowed Paths
- `frontend/src/views/user/CanvasView.vue`
- `frontend/src/views/user/__tests__/CanvasView.spec.ts`
- `frontend/src/components/canvas/**`
- `frontend/src/composables/useCanvas*.ts`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `docs/workflow/worker-results/**`

## Denied Paths
- `backend/**`
- `frontend/src/api/canvas.ts`
- `frontend/src/api/imageCreator.ts`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `C:/Users/Administrator/.codex/memories/**`
- 模板库、高级图像编辑、裁剪、外扩、mask、公开图库、账号/RBAC。

## Constraints
- 保持现有视觉风格，不引入新大型依赖。
- 优先把复杂交互拆到小组合函数或局部 helper，但不要做大规模重构。
- 拖拽和连线交互必须可用鼠标完成；测试可通过事件模拟完成。
- 不依赖后端新接口；只修改 Canvas document 并走现有保存。
- 避免 UI 文案说明“如何使用”的长教程；按钮使用简短标签或图标/tooltip。

## Acceptance Commands
```powershell
cd frontend
npm.cmd run test:run -- CanvasView canvas
npm.cmd run lint:check
```

## Output
- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 worker report 到 `docs/workflow/worker-results/sub2api-canvas-editor-core-result.md`。
- Worker report 第一行必须是 `### DONE: sub2api-canvas-editor-core`、`### BLOCKED: sub2api-canvas-editor-core` 或 `### FAILED: sub2api-canvas-editor-core`。
- 必须列出 changed files、commands run、key test output、risks、contract compliance、knowledge_candidates。
- 不允许直接写长期知识库；只提交候选结论。

## Stop Rules
- 如果需要改后端、API client、数据库 migration 或生产配置，停止并报告 BLOCKED。
- 如果交互实现需要大规模重写整个页面，先实现最小可用拖拽/连线/viewport，不做视觉重构。
- 如果验收命令无法运行，报告原因和最小替代验证。

## Budget
- worker_mode: `claude-bare-deepseek-v4-pro`
- worker_model: `deepseek-v4-pro`
- max_budget_usd: `0.12`
- worktree_root: `E:/codex-worktrees`

## Worker Output
- 同 `Output`。
