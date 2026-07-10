### DONE: tutorial-markdown-interactions-s61

## Changed Files
- `frontend/src/utils/tutorialMarkdown.ts`
- `frontend/src/utils/__tests__/tutorialMarkdown.spec.ts`

## Implementation
- 普通 fenced code、callout 内 fenced code 与 `[[command]]` 统一使用语言标签、单一复制按钮和 `data-copy-code` payload。
- shortcode command 不产生重复嵌套按钮；heading、TOC、截图、callout 与 link-button 行为保持。
- DOMPurify 清洗边界未放宽。

## Commands Run
- `npm.cmd run test:run -- src/utils/__tests__/tutorialMarkdown.spec.ts public-pages`
  - PASS：2 files / 14 tests。
- `npm.cmd run typecheck`
  - PASS。
- scoped `git diff --check`
  - PASS。

## Contract Compliance
- 只修改 S61 Allowed Paths，未触碰 `TutorialView.vue`、API、路由、后台或后端。

## Risks
- 页面级复制反馈由 S60 的 `TutorialView.vue` 统一处理，需在集成 QA 中一起验证。
