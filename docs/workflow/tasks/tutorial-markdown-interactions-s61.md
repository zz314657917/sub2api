# Task Contract

## Task ID
tutorial-markdown-interactions-s61

## Role
Developer Worker，负责教程 Markdown 渲染与行为单测；主 Codex 负责 diff review 与最终裁决。

## Goal
统一教程 Markdown 代码块交互，让普通 fenced code block 与 `[[command]]` 都具备一致、安全且可测试的复制能力。

## Allowed Paths
- `frontend/src/utils/tutorialMarkdown.ts`
- `frontend/src/utils/__tests__/tutorialMarkdown.spec.ts`
- `docs/workflow/worker-results/tutorial-markdown-interactions-s61-result.md`

## Denied Paths
- `TutorialView.vue`、教程内容、API、路由、后台编辑器、后端和部署。

## Required Behavior
- 所有渲染后的 `<pre><code>` 都包含语言标签、复制按钮和 `data-copy-code`。
- shortcode command 不重复嵌套复制按钮。
- 继续使用 DOMPurify，禁止降低现有清洗边界。
- 保留 heading id、TOC、截图、callout 和 link-button 行为。
- 增加真实渲染单测，不使用纯字符串断言代替行为验证。

## Success Criteria
- fenced code 与 shortcode command 都只渲染一个可用复制按钮。
- 清洗后不保留脚本或危险属性，原有 TOC 与 shortcode 行为不回归。

## Constraints
- 不新增依赖，不降低 DOMPurify 白名单边界，只在 Allowed Paths 内写入。

## Output
- 代码与渲染单测；最终报告首行必须为 `### DONE: tutorial-markdown-interactions-s61`、`### BLOCKED: tutorial-markdown-interactions-s61` 或 `### FAILED: tutorial-markdown-interactions-s61`。

## Acceptance
```powershell
cd frontend
npm.cmd run test:run -- src/utils/__tests__/tutorialMarkdown.spec.ts public-pages
npm.cmd run typecheck
```

## Stop Rules
- 如必须修改页面组件或降低 HTML 清洗策略，返回 `BLOCKED`。
- 不覆盖当前工作区其他未提交改动。
