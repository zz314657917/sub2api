# Task Contract

## Task ID
upstream-main-account-model-sync-s2b

## Role
Codex acts as Planner, Generator, and Final Evaluator for this Sprint. Implement only the create-flow upstream model sync preview patch selected here, and stop if conflicts require broader account/OAuth/gateway architecture changes.

## Goal
Port upstream `57d9e15e0 feat: 添加账号时支持同步上游模型` onto the current upstream-sync branch after Sprint 2. This Sprint adds the ability to sync upstream-supported models while creating an account, using unsaved form credentials, without changing database schema, billing, account scheduling, or existing saved-account sync behavior.

## Success Criteria
- Admin create-account flow can request upstream-supported models before the account is saved.
- Backend exposes a preview endpoint that builds a temporary account from request credentials and calls the existing upstream model fetch service.
- Existing saved-account upstream model sync remains intact.
- Frontend `ModelWhitelistSelector` can sync via either existing `accountId` or new unsaved `syncCredentials`.
- Create account modal passes API-key credentials into the selector for supported account forms.
- Changes are scoped to the selected upstream patch and documented if local code already contains equivalent saved-account sync behavior.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-migration-patches-s2`
- Work branch: `codex/upstream-main-account-model-sync-s2b`
- Upstream source commit: `57d9e15e0`
- Main worktree `F:/mcplugins/sub2api` has unrelated dirty Model Plaza changes and must not be modified.
- Local branch already contains saved-account sync via `POST /admin/accounts/:id/models/sync-upstream`; this Sprint only adds create-flow preview sync.

## Allowed Paths
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/admin/*sync_upstream*test.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/api_contract_test.go`
- `frontend/src/api/admin/accounts.ts`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/ModelWhitelistSelector.vue`
- `frontend/src/components/account/__tests__/*ModelWhitelist*.spec.ts`
- `frontend/src/components/account/__tests__/*CreateAccount*.spec.ts`
- `docs/workflow/tasks/upstream-main-account-model-sync-s2b.md`
- `docs/workflow/worker-results/upstream-main-account-model-sync-s2b-result.md`
- `docs/workflow/qa-reports/upstream-main-account-model-sync-s2b-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
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
- Prefer cherry-pick of `57d9e15e0`, but resolve conflicts by preserving local account modal, saved-account sync, Canvas/tickets/billing/public UI work.
- Do not introduce Ent schema/codegen/migrations.
- Do not add real upstream smoke tests requiring live credentials.
- If the selected patch requires touching denied paths or broad OAuth/account wiring, stop and split a new Sprint.

## Candidate Commit
- Primary: `57d9e15e0 feat: 添加账号时支持同步上游模型`

## Explicitly Deferred
- `eba204632` OpenAI OAuth refresh enrichment.
- `bf24b6113`, `b60d8bb4c` admin usage performance/deleted-user history.
- `user_platform_quotas`, DingTalk OAuth, payment/subscription/redeem/channel monitor migrations.
- OpenAI gateway / WS / Responses bridge redesign and response.failed stream handling.

## Acceptance Commands
```powershell
git status --short --branch
go test ./internal/handler/admin ./internal/server/routes -run "SyncUpstream|Account|Route|Contract" -count=1
go test ./internal/handler ./internal/server/routes -run "SyncUpstream|Account|Route|Contract" -count=1
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run lint:check
corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/BulkEditAccountModal.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts
```

## Output
- Write `docs/workflow/worker-results/upstream-main-account-model-sync-s2b-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-account-model-sync-s2b-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval, implementation, and QA events.

## Stop Rules
- Stop if implementing the selected commit requires touching denied paths.
- Stop if resolving conflicts requires adopting broader upstream account/OAuth architecture.
- Stop if tests fail for reasons requiring schema/API/config changes beyond this preview-sync patch.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`
