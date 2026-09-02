### DONE: upstream-v0200-group-pricing-layout-s290

# Worker Result

## Changed Files
- `frontend/src/components/admin/channel/IntervalRow.vue`
- `frontend/src/components/admin/channel/PricingEntryCard.vue`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/groupsModelsListLayout.spec.ts`

## Implementation
- 将本地 token 与按次区间行都收敛到 `pricing-interval-grid`，以
  `repeat(auto-fit, minmax(7.5rem, 1fr))` 适应可用宽度；删除按钮保持
  固定宽度，既有字段、输入事件和价格语义未变。
- 六项默认 token 价格使用 `pricing-default-grid`，最小列宽为 `8rem`。
- 新建/编辑分组模型定价弹窗改为既有 `wide` 宽度；标题行可换行，说明可收缩，添加
  控件不收缩且不换行。
- 扩展源码级回归，断言两个弹窗、两个标题行、添加控件和两种响应式网格均存在。

## Commands Run
```text
frontend: pnpm.cmd exec vitest run src/views/admin/__tests__/groupsModelsListLayout.spec.ts -> PASS (2 tests)
frontend: pnpm.cmd run typecheck -> PASS
frontend: pnpm.cmd run build -> PASS
root: git diff --check -- <four allowed frontend files> -> PASS
root: protected dirty diff hash -> 0e467987fd7aec5fc451983bdb8f8216f97ba69c
```

## Browser Attempt
- Vite was launched with `pnpm.cmd exec vite --host 127.0.0.1 --port 5174 --strictPort`.
- Playwright CLI session `sub2api-s290-pricing-layout` launched Chrome with a task-only
  `playwright_chromiumdev_profile-*` directory, but page creation failed with
  `TargetClosedError` before a snapshot or login interaction.
- The CLI session reported not open after the failure; no browser/cliDaemon process matching
  the task profile/session remained, and the Vite terminal process/listener was stopped.

## Risks
- Actual desktop/mobile screenshots of the authenticated group pricing dialogs remain
  unverified because the task browser closed before navigation. No production or user-profile
  browser session was used.

## Knowledge Candidates
- `pnpm.cmd run <script> -- <args>` forwards a literal `--` for this Vite script; use
  `pnpm.cmd exec vite --host ... --port ...` for an exact task-owned port.
