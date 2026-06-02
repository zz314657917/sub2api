### DONE: upstream-main-account-usage-window-hints-s2c

## Task ID
upstream-main-account-usage-window-hints-s2c

## Status
done

## Summary
- Ported upstream `c256a5441 feat(admin): 账号用量窗口 5h/7d 增加说明 tooltip` onto `codex/upstream-main-account-usage-window-hints-s2c`.
- Added a `HelpTooltip` beside the admin accounts table usage-window column header.
- Added modular i18n keys under `admin.accounts.usageWindowsHint` for English and Chinese locales.
- Added a focused `AccountsView` Vitest that renders the `header-usage` slot and verifies the hint key.

## Changed Files
- `frontend/src/views/admin/AccountsView.vue`
- `frontend/src/views/admin/__tests__/AccountsView.usageWindowsHint.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-account-usage-window-hints-s2c.md`
- `docs/workflow/worker-results/upstream-main-account-usage-window-hints-s2c-result.md`
- `docs/workflow/qa-reports/upstream-main-account-usage-window-hints-s2c-qa.md`

## Commands Run
```text
git status --short --branch -> clean on codex/upstream-main-account-usage-window-hints-s2c before implementation
git cherry-pick --no-commit c256a5441 -> conflict in frontend/src/i18n/locales/en.ts and zh.ts
git restore --source=HEAD --staged --worktree <cherry-pick touched paths> -> clean reset of the failed manual attempt
git diff --check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/AccountsView.usageWindowsHint.spec.ts -> pass, 1 file / 1 test
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
```

## Scope Notes
- Upstream uses monolithic `frontend/src/i18n/locales/en.ts` and `zh.ts`; local branch uses modular `frontend/src/i18n/locales/*/admin/accounts.ts`.
- The direct cherry-pick was intentionally not completed because it conflicted in denied monolithic i18n paths.
- Final implementation was manually ported to preserve local modular i18n and existing account table behavior.
- No backend, Ent schema/codegen/migrations, billing, account scheduling, gateway, or public API paths were changed.

## Risks
- No browser screenshot or live runtime smoke was run for the table header tooltip.
- Tooltip behavior relies on the existing shared `HelpTooltip` component, which is already covered separately in this frontend project.

## Knowledge Candidates
- None.

## Contract Compliance
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.
