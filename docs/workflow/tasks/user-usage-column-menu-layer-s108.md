---
task_id: user-usage-column-menu-layer-s108
phase: contract-approved
qa_mode: runtime
---

# S108 User Usage Column Menu Layer Contract

## Task ID

user-usage-column-menu-layer-s108

## Role

你是 P/G/E 流程中的 Generator。只修复用户端使用记录筛选卡片与表格的层叠关系，不改变功能行为。

## Goal

让用户端使用记录右侧列设置菜单在打开时完整显示在固定表头和记录行之上。

## Success Criteria

- `UsageView` 仅在 `showColumnMenu` 打开时把筛选卡提升到 `z-[221]`，高于 `DataTable` 固定表头最高 `z-index: 220`；关闭后移除临时层级。
- 列设置菜单的开关、列显隐持久化、筛选、刷新、重置和 CSV 导出行为保持不变。
- 定向视图测试锁定菜单开关时筛选卡的动态 overlay layer。
- 桌面浏览器 smoke 中，菜单边界和菜单项位于固定表头之上，没有被表格容器裁切或遮挡。

## Context

- Repo: `F:/mcplugins/sub2api`
- Related paths: `frontend/src/views/user/UsageView.vue`, `frontend/src/components/layout/TablePageLayout.vue`, `frontend/src/components/common/DataTable.vue`
- Root cause: 菜单自身 `z-50` 低于 `DataTable` 固定表头的 `z-index: 210/220`，筛选卡没有在菜单打开时建立更高层叠上下文。
- Existing precedent: 管理端 S72 已使用只在列菜单打开时提升父筛选卡到 `z-[221]` 的最小修复。

## Allowed Paths

- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `docs/workflow/worker-results/user-usage-column-menu-layer-s108-result.md`
- `docs/workflow/qa-reports/user-usage-column-menu-layer-s108-qa.md`
- `output/playwright/**`

## Denied Paths

- `frontend/src/components/common/DataTable.vue`
- `frontend/src/components/layout/TablePageLayout.vue`
- `backend/**`
- `knowledge/**`
- deployment and container files
- `C:/Users/Administrator/.codex/memories/**`

## Constraints

- 复用管理端已验证的 `z-[221]` 层级，只绑定现有 `showColumnMenu` 状态。
- 不引入 Teleport，不修改全局表格层级、菜单定位、关闭逻辑、列显隐或 API 调用。
- 不覆盖当前 worktree 中 S106/S107 及其他既有改动。

## Acceptance Commands

```powershell
Push-Location frontend
node_modules/.bin/vitest.cmd run src/views/user/__tests__/UsageView.spec.ts
npm.cmd run typecheck
npm.cmd run build
Pop-Location

git diff --check
```

## Output

- QA report 第一行必须是 `### PASS: user-usage-column-menu-layer-s108`、`### FAIL: user-usage-column-menu-layer-s108` 或 `### BLOCKED: user-usage-column-menu-layer-s108`。
- 列出 changed files、executed checks、unverified risks 和 contract compliance。

## Stop Rules

- 如果动态提升筛选卡仍无法使菜单压过表格，停止并重新评审；不得直接改成 Teleport 或修改 `DataTable` 全局层级。
- 如果需要修改 Allowed Paths 之外的业务文件，停止并请求 Codex 裁决。
