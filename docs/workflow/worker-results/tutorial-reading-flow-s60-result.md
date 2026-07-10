### DONE: tutorial-reading-flow-s60

## Changed Files
- `frontend/src/views/public/TutorialView.vue`
- `frontend/src/views/public/__tests__/TutorialView.spec.ts`
- `frontend/src/__tests__/public-pages.spec.ts`（只同步备用教程用户文案断言）

## Implementation
- `/tutorial` 保留总览并增加搜索、分类分组；详情路由改为紧凑文章阅读布局。
- 桌面保留 sticky 教程目录；平板/移动改为当前教程折叠目录，本页 TOC 位于正文之前。
- 章节导航同步 URL hash，支持直接深链、刷新和前进后退定位。
- 增加篇间进度、上一篇/下一篇、loading/error/notFound 分离。
- 普通与 shortcode 复制按钮提供当前按钮反馈；图片预览支持键盘打开、Escape 关闭与焦点恢复。
- Evaluator 退修补齐 `aria-current`、`aria-pressed` 和非运营化备用教程文案。
- Browser QA 退修让正文截图按自然尺寸显示，避免低分辨率素材被桌面放大；移动索引隐藏重复路线卡但保留主 CTA，使搜索进入首屏。

## Commands Run
- `npm.cmd run test:run -- src/views/public/__tests__/TutorialView.spec.ts public-pages`
  - PASS：2 files / 15 tests。
- `npm.cmd run typecheck`
  - PASS。
- scoped `git diff --check`
  - PASS。
- Browser QA 修正后复跑 `TutorialView.spec.ts + public-pages`：PASS，2 files / 15 tests；`npm.cmd run typecheck`：PASS。

## Contract Compliance
- 业务代码只修改 `TutorialView.vue`；未修改 API、路由、后台、后端或 fallback 正文。

## Risks
- 真实 CMS 的延迟/500/404 浏览器注入仍需在统一 QA 中复核；组件测试已覆盖三种状态。
