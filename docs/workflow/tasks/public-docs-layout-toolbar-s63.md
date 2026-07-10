# Task Contract

## Task ID
public-docs-layout-toolbar-s63

## Role
两个 Developer Worker 分别负责教程详情布局和模型工具栏；主 Codex 负责集成、浏览器验收与最终裁决。

## Goal
把教程详情改成宽屏真正三栏的文档阅读布局，并把模型广场搜索、分类与结果统计改成扁平、清晰、无嵌套卡片的工具栏。

## Allowed Paths
- `frontend/src/views/public/TutorialView.vue`
- `frontend/src/views/public/ModelPlazaView.vue`
- `frontend/src/views/public/__tests__/TutorialView.spec.ts`
- `frontend/src/views/public/__tests__/ModelPlazaView.spec.ts`
- `frontend/src/__tests__/public-pages.spec.ts`（仅同步 S63 布局断言）
- `docs/workflow/worker-results/public-docs-layout-toolbar-s63-result.md`
- `docs/workflow/qa-reports/public-docs-layout-toolbar-s63-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths
- 教程 CMS/API、fallback 教程正文、模型目录 API、计价逻辑、账号分组权限、路由、后端、数据库和部署。

## Required Behavior
- 宽屏教程详情使用左侧教程目录、中间正文、右侧本页目录三个同级列；页面容器不再受 104rem 居中窄框限制。
- 左右目录 sticky 并靠近页面两侧；中间正文宽度保持约 46-50rem，去掉整篇文章外层卡片边框、阴影和实底背景。
- 代码块、callout、截图等局部内容可以继续使用边框；教程索引和 `<=1100px` 移动/平板交互不得回归。
- 模型搜索框独占一行；分类 tabs 与“共 N 个型号/规格，来自 M 个分组”位于同一条扁平 filter row。
- 模型 toolbar 和分类 tabs 不再使用嵌套面板/毛玻璃卡片；分类仍有明确选中态、键盘焦点和移动适配。
- 不改变 S60-S62 搜索、hash、复制、移动展开、价格说明与 stale refresh 行为。

## Success Criteria
- `2560x900` 教程详情：左目录距视口左边小于 96px，右 TOC 距视口右边小于 96px，中间正文宽度为 720-820px；三列互不重叠。
- `1280x720` 教程详情仍可读且无横向溢出；`390x844` 保持当前折叠目录和移动 TOC。
- `1280x720` 模型页工具栏无外层卡片与 tabs 内层卡片，搜索、filter row、统计层级清晰；`390x844` 不溢出。

## Constraints
- 不新增依赖；保持暖白、陶土、黑灰设计语言；兼容当前 mixed dirty tree。
- 只做布局和视觉层级调整，不扩大产品功能。

## Acceptance
```powershell
cd frontend
npm.cmd run test:run -- src/views/public/__tests__/TutorialView.spec.ts src/views/public/__tests__/ModelPlazaView.spec.ts public-pages public-smoke
npm.cmd run typecheck
npm.cmd run build
cd ..
git diff --check
```

## Output
- Worker 首行必须为 `### DONE: public-docs-layout-toolbar-s63-*`、`BLOCKED` 或 `FAILED`。

## Stop Rules
- 如需改模板数据结构以外的 API/路由/后端行为，返回 `BLOCKED`。
- 不覆盖 S58-S62 或其他并行 dirty 改动。
