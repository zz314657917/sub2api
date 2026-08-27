---
task_id: upstream-email-bind-alias-s267
phase: contract-approved
qa_mode: runtime
---

# Upstream Email-Bind Alias Concurrency Guard S267

## Role

Primary Codex performs the bounded Planner/Generator/Evaluator flow in this
isolated worktree. The change is a manual behavior port; no broad upstream
merge is allowed.

## Goal

Port the missing behavior from upstream `4ca86c52e`: email identity binding
must reject an existing provider alias and must recheck exact and alias
ownership under the existing transaction-scoped uniqueness locks before the
primary email/password update is written.

## Success Criteria

- Email bind-code requests and verified email binds reject an alias that
  belongs to another user, while allowing the current user's own alias.
- The primary email update performs exact and alias recheck inside the caller's
  transaction and serializes both normalized-email and inbox-identity keys;
  concurrent variants of one inbox produce at most one successful bind.
- Existing auth-identity replacement, first-bind grant transactionality,
  exact-email behavior, login/delivery storage, and admin creation semantics
  remain unchanged.
- Alias candidate lookup remains bounded, returns the conflicting owner when
  needed, and keeps the existing soft-delete and normalization behavior.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen base: `main@659ad5d11`
- Upstream behavior source: `4ca86c52e3edbcc5d38247842f045f6975329263`
- Worker worktree: `E:/codex-worktrees/sub2api/upstream-email-bind-alias-s267`

## Allowed Paths

- `backend/internal/repository/user_repo.go`
- `backend/internal/service/auth_email_binding.go`
- `backend/internal/service/auth_service_email_bind_test.go`
- `docs/workflow/tasks/upstream-email-bind-alias-s267.md`
- `docs/workflow/worker-results/upstream-email-bind-alias-s267-result.md`
- `docs/workflow/qa-reports/upstream-email-bind-alias-s267-qa.md`

## Denied Paths

- `F:/mcplugins/sub2api` primary worktree and all Pixel Cafe/user dirty files
- `backend/ent/**`, migrations, schema, handlers, routes, frontend,
  configuration, dependencies, provider calls, containers, deployment,
  databases, `knowledge/**`, `outputs/**`, staging, and push
- Registration alias policy and admin user-creation semantics except for
  read-only regression checks

## Constraints

- Reuse `lockRepositoryScopedKeys`, `emailAliasUniquenessLockKey`, and the
  existing transaction context; do not add a second locking mechanism.
- Keep the extra repository capability behind a narrow service-local interface
  so unrelated UserRepository implementations and test stubs do not change.
- The service preflight may remain a fast rejection, but the transaction
  recheck is authoritative and must occur before the user update.
- Do not expose credentials, raw tokens, or user email data in reports.

## Acceptance Commands

```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Set-Location backend
go test -tags=unit ./internal/service -run '^(TestAuthServiceBindEmailIdentity|TestAuthServiceSendEmailIdentityBindCode)' -count=10
go test -tags=unit ./internal/repository -run '^TestUserRepository(ExistsByEmailAlias|CreateWithEmailAliasGuard)' -count=1
go test ./internal/service ./internal/repository -run '^$' -count=1
gofmt -d internal/repository/user_repo.go internal/service/auth_email_binding.go internal/service/auth_service_email_bind_test.go
Set-Location ..
git diff --check -- backend/internal/repository/user_repo.go backend/internal/service/auth_email_binding.go backend/internal/service/auth_service_email_bind_test.go docs/workflow/tasks/upstream-email-bind-alias-s267.md
git ls-files -u
```

## Stop Rules

- Stop if the lock/recheck requires schema or migration changes, changes the
  registration/admin contract, or cannot preserve transaction rollback.
- Stop before touching any denied path or the primary worktree.
- Stop if a focused test cannot run without enabling unrelated production or
  provider behavior; report the baseline failure separately.

## Output

- Commit one scoped implementation commit and one result report commit on the
  worker branch. Independent QA must write its own report before integration.
- Final evaluation must include exact scope, commands, conflict/index state,
  upstream provenance, and residual runtime risk.
