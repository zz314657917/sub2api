# Task Contract

## Task ID
tutorial-reading-flow-s60

## Role
Developer Worker，负责公共教程阅读流实现；主 Codex 负责 diff review 与最终裁决。

## Goal
重构公共教程的信息架构和连续阅读路径，让教程详情在桌面、平板和移动首屏都直接进入当前文章，并提供可发现、可分享、可连续阅读的目录导航。

## Allowed Paths
- `frontend/src/views/public/TutorialView.vue`
- `frontend/src/views/public/__tests__/TutorialView.spec.ts`
- `frontend/src/__tests__/public-pages.spec.ts`（仅允许同步 S60 用户可见文案和行为断言，不得改动 S58/S59 其他断言）
- `docs/workflow/worker-results/tutorial-reading-flow-s60-result.md`

## Denied Paths
- 教程 CMS API、后台教程编辑器、路由定义、数据库、后端、部署和 fallback 教程正文。

## Required Behavior
- `/tutorial` 保留总览、推荐路线、搜索和分类目录。
- `390x844` 教程索引首屏可见搜索入口；移动端可隐藏与目录重复的四张路线卡，但必须保留“新手最快路线”主入口。
- `/tutorial/:slug` 不再重复显示大 Hero 与四步路线，首屏直接展示紧凑文章头和正文。
- 移动端使用“当前教程 + 展开目录”控制，当前项始终可见；桌面保留稳定侧栏。
- 文章目录点击同步 URL hash，直接打开 hash、刷新和浏览器前进后退均能定位。
- 文章头展示篇间进度，文末提供上一篇/下一篇。
- 详情 loading、error、notFound 分离；错误可重试，404 才显示不存在。
- 普通和 shortcode 代码复制均提供当前按钮成功/失败反馈。
- 图片预览支持键盘打开、Escape 关闭和关闭后焦点恢复。
- 正文低分辨率截图不得被放大超过固有尺寸；窄屏仍需按容器缩小且不产生横向溢出。

## Success Criteria
- `1024x768` 和 `1280x720` 直达详情时首屏可见文章标题与第一段正文。
- `390x844` 下当前教程始终可见，目录无需盲目横滑；本页目录位于正文之前。
- 章节深链、篇间导航、复制反馈和详情状态均满足 Required Behavior。
- 桌面截图渲染宽度不超过 `naturalWidth`，移动端图片右边界不超过正文容器。
- `390x844` 索引页搜索控件顶部小于视口高度，页面无横向溢出。

## Constraints
- 不新增依赖，不改变 CMS/API 数据结构，保持现有暖白/陶土公共页设计语言。
- 只在 Allowed Paths 内写入，兼容当前 mixed dirty tree。
- `public-pages.spec.ts` 属于共享 dirty 文件，只允许最小化同步本 Sprint 对应断言，禁止重排或覆盖其他改动。

## Contract Amendment
- 2026-07-11：Evaluator 退修要求把备用教程文案改为面向用户的语言，对应静态 contract 测试必须同步；Planner 补批 `frontend/src/__tests__/public-pages.spec.ts` 的单条断言范围。该修订不扩大业务行为或生产代码范围。
- 2026-07-11：浏览器 QA 发现 `278x176` 等低分辨率截图在桌面被放大到 600px 以上；补充固有尺寸上限作为原“多视口阅读体验”目标的视觉验收项。
- 2026-07-11：浏览器 QA 发现移动索引四张路线卡把搜索入口推到 `y=1229px`；允许在 `<=640px` 隐藏重复路线卡，保留主 CTA，并要求搜索进入首屏。

## Output
- 代码与定向测试；最终报告首行必须为 `### DONE: tutorial-reading-flow-s60`、`### BLOCKED: tutorial-reading-flow-s60` 或 `### FAILED: tutorial-reading-flow-s60`。

## Acceptance
```powershell
cd frontend
npm.cmd run test:run -- src/views/public/__tests__/TutorialView.spec.ts public-pages public-smoke
npm.cmd run typecheck
```

## Stop Rules
- 如需修改 API 合约、路由、后台编辑器或 fallback 正文，返回 `BLOCKED`。
- 不覆盖当前工作区 S58/S59 或其他未提交改动。
