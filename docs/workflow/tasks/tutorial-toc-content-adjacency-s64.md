# Task Contract

## Task ID
tutorial-toc-content-adjacency-s64

## Role
主 Codex 负责 Planner、实现、浏览器验收与最终裁决。

## Goal
保持左侧教程目录贴近页面左边和正文居中的前提下，把桌面本页目录移动到正文右侧邻接位置，消除正文与 TOC 之间的大段空白。

## Allowed Paths
- `frontend/src/views/public/TutorialView.vue`
- `frontend/src/views/public/__tests__/TutorialView.spec.ts`
- `frontend/src/__tests__/public-pages.spec.ts`（仅同步布局断言）
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/qa-reports/tutorial-toc-content-adjacency-s64-qa.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths
- 教程 DOM 内容结构、CMS/API、Markdown 渲染、路由、模型广场、后端、数据库和部署。

## Required Behavior
- `2560x900` 正文仍保持 `720-820px`，左教程目录仍距左边小于 `96px`。
- 本页目录位于正文右侧，正文右边到 TOC 左边的间距为 `16-40px`。
- `1280x720` 三列不重叠、无横向溢出；`<=1100px` 移动目录与移动 TOC 不回归。

## Acceptance
```powershell
cd frontend
npm.cmd run test:run -- src/views/public/__tests__/TutorialView.spec.ts public-pages public-smoke
npm.cmd run typecheck
cd ..
git diff --check
```

## Stop Rules
- 如需改变正文宽度、移动断点或教程交互，停止并回到 Planner。
