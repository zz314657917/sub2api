### DONE: upstream-main-account-model-sync-s2b

## Task ID
upstream-main-account-model-sync-s2b

## Status
done

## Summary
- Ported upstream `57d9e15e0 feat: 添加账号时支持同步上游模型` onto `codex/upstream-main-account-model-sync-s2b`.
- Added create-flow upstream model sync preview via `POST /admin/accounts/models/sync-upstream-preview`.
- Extended admin accounts API and `ModelWhitelistSelector` so model sync can use either a saved `accountId` or unsaved create-form credentials.
- Wired create account modal whitelist selectors to provide API-key credentials for preview sync.

## Changed Files
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/server/routes/admin.go`
- `frontend/src/api/admin/accounts.ts`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/ModelWhitelistSelector.vue`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-account-model-sync-s2b.md`
- `docs/workflow/worker-results/upstream-main-account-model-sync-s2b-result.md`
- `docs/workflow/qa-reports/upstream-main-account-model-sync-s2b-qa.md`

## Commands Run
```text
git status --short --branch -> clean on codex/upstream-main-account-model-sync-s2b before implementation and before QA
git cherry-pick 57d9e15e0 -> success, commit 764e12073
git diff --check HEAD~1..HEAD -> pass
go test ./internal/handler/admin ./internal/server/routes -run "SyncUpstream|Account|Route|Contract" -count=1 -> pass
go test ./internal/handler ./internal/server/routes -run "SyncUpstream|Account|Route|Contract" -count=1 -> pass
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/BulkEditAccountModal.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts -> pass, 2 files / 26 tests
```

## Scope Notes
- Local branch already had saved-account upstream model sync via `/admin/accounts/:id/models/sync-upstream`; this Sprint added the unsaved create-flow preview path.
- No Ent schema/codegen/migrations were changed.
- No live upstream credential smoke was run.

## Risks
- The preview endpoint depends on existing upstream account test service behavior; real upstream availability and credentials were not exercised.
- The target frontend tests cover related account modal/selector behavior, but no browser smoke was run for the create-account modal.

## Knowledge Candidates
- None.

## Contract Compliance
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.
