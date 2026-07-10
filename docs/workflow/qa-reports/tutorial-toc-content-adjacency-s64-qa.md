### PASS: tutorial-toc-content-adjacency-s64

## Findings
- 未发现布局重叠、横向溢出或移动端回归。
- 正文位置和 `800px` 宽度保持不变；右侧本页目录从视口边缘移动到正文邻接列。

## Executed Checks
- `npm.cmd run test:run -- src/views/public/__tests__/TutorialView.spec.ts public-pages public-smoke`：3 files / 21 tests PASS。
- `npm.cmd run typecheck`：PASS。
- `git diff --check`：PASS，仅既有 CRLF future-conversion warning。
- `2560x900`：左目录 `left=32px`，正文 `904-1704px`，TOC `1736-1976px`，正文到 TOC 间距 `32px`。
- `1280x720`：正文宽 `800px`，TOC 宽 `154px`，正文到 TOC 间距 `19px`，无横向溢出。
- `390x844`：桌面左目录与 TOC 隐藏，移动目录与移动 TOC 显示，`scrollWidth == clientWidth == 390`。

## Evidence
- `output/playwright/tutorial-s64-2560x900.png`
- `output/playwright/tutorial-s64-1280x720.png`
- `output/playwright/tutorial-s64-390x844.png`

## Recommendation
- PASS。右侧本页目录已成为正文的邻接导航，不再贴视口右边。
