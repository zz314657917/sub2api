# 当前任务快照

最后更新：2026-05-27 03:24 +08:00

## 背景

- 项目主仓库：`F:/mcplugins/sub2api`。
- 用户要求：进入 plan 阶段，把“使用记录”迁出到一级。
- 已按 P/G/E 入口初始化本仓库 workflow 事实源，文件位于 `docs/workflow/`。
- 注意：`docs/workflow/*` 当前被仓库 `.gitignore` 的 `docs/*` 规则忽略，是本地流程事实源；`knowledge/00-start-here.md` 已增加 workflow 入口片段。

## 当前目标

- 当前按 Planner 理解：“使用记录”指管理端 `/admin/usage`，目标是从管理端“内容与记录”分组的二级菜单迁出为管理端侧栏一级入口。
- 用户侧 `/usage` 已经是一级导航，本轮不调整。
- 路由、页面业务、后端 API、权限和 i18n 文案不变。

## 本次已完成

- 初始化 P/G/E workflow：
  - `docs/workflow/status.md`
  - `docs/workflow/agent-matrix.md`
  - `docs/workflow/spec.md`
  - `docs/workflow/sprint-01-contract.md`
  - `docs/workflow/sprint-01-review.md`
  - `docs/workflow/sprint-01-qa.md`
  - `docs/workflow/sprint-01-fix-log.md`
  - `docs/workflow/main-log.md`
- 已起草 spec：`docs/workflow/spec.md`。
- 已起草并审核 Sprint 1 contract：`docs/workflow/sprint-01-contract.md`、`docs/workflow/sprint-01-review.md`。
- 已实现管理端侧栏层级迁移：
  - `frontend/src/components/layout/AppSidebar.vue`
  - `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`
- 已将 `docs/workflow/status.md` 推进到 `done`，当前任务 ID 为 `sub2api-admin-usage-top-level-nav`。

## 已确认事实

- `frontend/src/components/layout/AppSidebar.vue` 中用户侧 `/usage` 位于 `primarySelfItems`，已经是一级入口。
- 管理端 `/admin/usage` 已位于 `adminNavItems` 一级数组。
- `/admin/content-records` 分组还包含 `/admin/announcements` 和 `/admin/tutorials`，移出 `/admin/usage` 后不会变成空分组。
- `/admin/usage` 路由在 `frontend/src/router/index.ts` 中已存在，指向 `@/views/admin/UsageView.vue`，需要管理员权限。

## 待验证点

- 动作：如需要视觉复核，用管理员登录态打开 `http://127.0.0.1:62080/admin/usage`。
  验证：`/admin/usage` 显示为管理端一级菜单，“内容与记录”展开后不再显示“使用记录”。

## 当前结论

- 本轮已完成实现和目标测试。
- `/admin/usage` 已迁移为管理端一级菜单；“内容与记录”分组不再包含使用记录。
- 浏览器打开用户给出的地址后当前 in-app browser 显示登录页，说明没有管理员登录态；无法在该浏览器会话直接视觉检查侧栏。

## 下一步

- 动作：如用户提供管理员登录态或在已登录浏览器中刷新页面，做一次视觉复核。
  验证：管理端侧栏一级菜单中可见“使用记录”，内容与记录分组只展示公告和教程。

## 验证记录

- 已执行检查：
  - 读取 `docs/workflow/status.md`、`docs/workflow/spec.md`、`docs/workflow/sprint-01-contract.md`。
  - 搜索并确认 `/admin/usage`、`/usage` 与 `AppSidebar.vue` 导航位置。
  - 读取 `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`，确认有现成侧栏结构测试可扩展。
  - `corepack.cmd pnpm exec vitest run src/components/layout/__tests__/AppSidebar.spec.ts`，工作目录 `F:/mcplugins/sub2api/frontend`，结果 13 passed。
  - 浏览器打开 `http://127.0.0.1:62080/admin/usage`，当前会话被展示登录页，未能直接查看管理端侧栏。
