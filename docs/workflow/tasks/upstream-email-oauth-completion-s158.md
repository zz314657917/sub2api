# Upstream Email OAuth Completion S158

## Task ID

`upstream-email-oauth-completion-s158`

## Role

Primary Codex performs Planner, direct implementation, and final Evaluator gates in sequence. No worker is used for this isolated upstream port.

## Goal

Port the final behavior of upstream commit `260fda19b` so OAuth email completion remains usable when email verification is required, backend mode is enabled, or the API base URL is configured with a relative/non-default prefix.

## Success Criteria

- LinuxDO OAuth pending signup clears synthetic/resolved email and reports `create_account_required` with email binding required when verification or forced email signup applies without a compatible local account.
- The pending create-account flow can bind the verified email and create the LinuxDO identity; existing identity, compatible-account, current-user binding, and ordinary signup paths retain their current behavior.
- Backend mode allows GitHub and Google `complete-registration` endpoints while continuing to block OAuth start endpoints.
- OAuth callback redirection uses the shared API URL builder and normalizes a relative `VITE_API_BASE_URL` without duplicating `/api/v1`.
- Focused Go and frontend tests, typecheck, formatting, diff, and scoped integrity checks pass.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-email-alias-dedup-s157`
- Base: `d100d7592`
- Upstream behavior: `260fda19b`
- This branch already contains isolated S156/S157 email ports; the dirty primary worktree remains untouched.

## Allowed Paths

- `backend/internal/handler/auth_linuxdo_oauth.go`
- `backend/internal/handler/auth_linuxdo_oauth_test.go`
- `backend/internal/server/middleware/backend_mode_guard.go`
- `backend/internal/server/middleware/backend_mode_guard_test.go`
- `frontend/src/api/__tests__/client.spec.ts`
- `frontend/src/api/url.ts`
- `frontend/src/views/auth/OAuthCallbackView.vue`
- `frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts`
- `docs/workflow/tasks/upstream-email-oauth-completion-s158.md`
- `docs/workflow/qa-reports/upstream-email-oauth-completion-s158-qa.md`

## Denied Paths

- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Ent schema/generated files, migrations, production configuration, deployment, containers, live databases, notification templates, subscription mail, payment, and unrelated frontend/backend modules
- The primary worktree `F:/mcplugins/sub2api`

## Constraints

- Preserve the final upstream behavior rather than cherry-picking topology-specific patches.
- Do not weaken backend-mode restrictions or expose synthetic email as a user-selected verified address.
- Keep the existing dirty primary worktree untouched.
- Do not run migrations, provider calls, deployment, or production smoke.

## Acceptance Commands

```powershell
go test ./internal/handler -run '^TestLinuxDoOAuthCallbackEmailVerificationCompletesWithBoundEmail$' -count=1
go test ./internal/server/middleware -run '^TestBackendModeAuthGuard$' -count=1
go test ./internal/handler ./internal/server/middleware -run 'Test(LinuxDoOAuthCallback|BackendModeAuthGuard)' -count=1
go test ./internal/handler ./internal/server/middleware -run '^$' -count=1
npm.cmd run test:run -- src/views/auth/__tests__/OAuthCallbackView.spec.ts src/api/__tests__/client.spec.ts
npm.cmd run typecheck
gofmt -d backend/internal/handler/auth_linuxdo_oauth.go backend/internal/handler/auth_linuxdo_oauth_test.go backend/internal/server/middleware/backend_mode_guard.go backend/internal/server/middleware/backend_mode_guard_test.go
git diff --check
git status --short
```

## Output

- QA report: `docs/workflow/qa-reports/upstream-email-oauth-completion-s158-qa.md`
- Final verdict must be `PASS`, `FAIL`, or `BLOCKED`, with executed checks and unverified risks.

## Stop Rules

- Stop if the completion flow requires schema/migration or changes outside the allowlist.
- Stop if a path cannot distinguish verified-email account creation from ordinary binding/update.
- Stop if validation requires writing to a configured/live database or provider.
