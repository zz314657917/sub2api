# Task Contract

## Task ID
upstream-main-account-usage-window-hints-s2c

## Role
Codex acts as Planner, Generator, and Final Evaluator for this small frontend-only Sprint. Implement only the account usage-window explanatory tooltip selected here, and stop if conflicts require broader account table, i18n structure, backend, schema, or gateway changes.

## Goal
Port upstream `c256a5441 feat(admin): 账号用量窗口 5h/7d 增加说明 tooltip` onto the current upstream-sync branch after Sprint 2b. This Sprint adds a single help tooltip to the admin accounts table usage-window column header, explaining that `5h / 7d` are upstream account rolling usage windows.

## Success Criteria
- Admin accounts table keeps the existing `Usage Windows` / `用量窗口` column label.
- A `HelpTooltip` appears beside the usage-window column header.
- Tooltip content is exposed through the existing modular i18n files.
- A focused frontend test verifies the header slot renders the hint key.
- No backend, schema, billing, account scheduling, gateway, or public API behavior changes are introduced.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-account-model-sync-s2b`
- Work branch: `codex/upstream-main-account-usage-window-hints-s2c`
- Upstream source commit: `c256a5441`
- Main worktree `F:/mcplugins/sub2api` has unrelated dirty Model Plaza/reference-pricing changes and must not be modified.
- Local i18n is modular. Do not adopt upstream monolithic `frontend/src/i18n/locales/en.ts` / `zh.ts` structure.

## Allowed Paths
- `frontend/src/views/admin/AccountsView.vue`
- `frontend/src/views/admin/__tests__/AccountsView.usageWindowsHint.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/tasks/upstream-main-account-usage-window-hints-s2c.md`
- `docs/workflow/worker-results/upstream-main-account-usage-window-hints-s2c-result.md`
- `docs/workflow/qa-reports/upstream-main-account-usage-window-hints-s2c-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/**`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Payment, subscription notify, redeem expiry, channel monitor API mode, DingTalk, `user_platform_quotas`, OpenAI WS/Responses bridge redesign, or unrelated local UI/public page changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Prefer cherry-pick of `c256a5441` only if it does not pull in monolithic i18n structure. Otherwise port manually.
- Preserve local modular i18n and existing account table behavior.
- Do not introduce Ent schema/codegen/migrations.
- Do not add real upstream smoke tests requiring live credentials.
- If the selected patch requires touching denied paths or broader account table architecture, stop and split a new Sprint.

## Candidate Commit
- Primary: `c256a5441 feat(admin): 账号用量窗口 5h/7d 增加说明 tooltip`

## Explicitly Deferred
- OpenAI OAuth refresh enrichment.
- Admin usage performance/deleted-user history.
- `user_platform_quotas`, DingTalk OAuth, payment/subscription/redeem/channel monitor migrations.
- OpenAI gateway / WS / Responses bridge redesign and response.failed stream handling.
- Any backend/runtime Docker smoke outside this frontend-only Sprint.

## Acceptance Commands
```powershell
git status --short --branch
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run lint:check
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/AccountsView.usageWindowsHint.spec.ts
```

## Output
- Write `docs/workflow/worker-results/upstream-main-account-usage-window-hints-s2c-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-account-usage-window-hints-s2c-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval, implementation, and QA events.

## Stop Rules
- Stop if implementing the selected commit requires touching denied paths.
- Stop if resolving conflicts requires adopting upstream monolithic i18n or broader account/OAuth architecture.
- Stop if tests fail for reasons requiring backend/schema/API/config changes beyond this tooltip patch.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`
