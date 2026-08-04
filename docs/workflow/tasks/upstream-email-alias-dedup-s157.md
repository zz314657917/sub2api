# Upstream Email Alias Dedup S157

## Task ID

`upstream-email-alias-dedup-s157`

## Role

Primary Codex performs Planner, direct implementation, and final Evaluator gates in sequence. No worker is used for this isolated upstream port.

## Goal

Port the final behavior of upstream commits `b6f927751` and `bc3acd6e2` so one real inbox cannot create multiple grant-bearing accounts through plus aliases, Gmail dots, Googlemail aliases, or a trailing DNS root dot.

## Success Criteria

- Local registration, verification-code preflight, and OAuth email account creation reject an existing inbox alias.
- Registration creation rechecks the inbox identity under repository-scoped locks and returns `ErrEmailExists` on a concurrent collision.
- Alias lookup is bounded, escapes SQL LIKE wildcards, preserves soft-delete behavior, and has the matching PostgreSQL partial expression index.
- Admin user creation and email binding/update behavior remain on the existing exact-email path.
- Focused normalization, service, repository, handler, and compile checks pass, or baseline failures are reported separately.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-email-alias-dedup-s157`
- Base: `207e29d31`
- Upstream behavior: `b6f927751`, corrected by `bc3acd6e2`

## Allowed Paths

- `backend/internal/handler/auth_oauth_pending_flow_test.go`
- `backend/internal/handler/user_handler_test.go`
- `backend/internal/repository/user_repo.go`
- `backend/internal/repository/user_repo_email_alias_test.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/server/middleware/admin_auth_test.go`
- `backend/internal/service/admin_service_apikey_test.go`
- `backend/internal/service/admin_service_delete_test.go`
- `backend/internal/service/admin_service_email_identity_sync_test.go`
- `backend/internal/service/auth_oauth_email_flow.go`
- `backend/internal/service/auth_service.go`
- `backend/internal/service/auth_service_email_bind_test.go`
- `backend/internal/service/auth_service_register_test.go`
- `backend/internal/service/content_moderation_test.go`
- `backend/internal/service/group_buy_test.go`
- `backend/internal/service/registration_email_alias.go`
- `backend/internal/service/registration_email_alias_test.go`
- `backend/internal/service/user_service.go`
- `backend/internal/service/user_service_test.go`
- `backend/migrations/190_add_users_email_alias_dedup_index_notx.sql`
- `docs/workflow/tasks/upstream-email-alias-dedup-s157.md`
- `docs/workflow/qa-reports/upstream-email-alias-dedup-s157-qa.md`

## Denied Paths

- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Ent schema/generated files
- production configuration, deployment, containers, or live databases
- notification template, payment, subscription, and frontend modules

## Constraints

- Preserve the final upstream behavior rather than cherry-picking topology-specific patches.
- Do not run migration `190` against any configured database.
- Do not change stored email values, login identity, delivery addresses, admin creation, or email binding/update semantics.
- Keep the existing dirty primary worktree untouched.

## Acceptance Commands

```powershell
go test -tags=unit ./internal/service -run '^(TestNormalizeEmailForAliasDedup|TestNormalizeEmailForAliasDedupKeepsDistinctInboxes|TestEmailAliasDedupProbes|TestExistsByEmailOrAlias|TestAuthService_Register_AliasDuplicateRejected|TestAuthService_Register_UsesAliasGuardedCreate|TestRegisterOAuthEmailAccount.*)$' -count=1
go test -tags=unit ./internal/repository -run '^(TestUserRepositoryExistsByEmailAlias|TestUserRepositoryExistsByEmailAliasIgnoresMalformedInput|TestUserRepositoryCreateWithEmailAliasGuard)$' -count=1
go test -tags=unit ./internal/handler ./internal/server ./internal/server/middleware -run '^(TestCreateOIDCOAuthAccount.*|TestSendPendingOAuthVerifyCodeExistingEmailReturnsBindLoginState|TestAPIContracts|TestAdminAuthJWTValidatesTokenVersion)$' -count=1
go test ./internal/service ./internal/repository -run '^$'
gofmt -d <changed-go-files>
git diff --check
git status --short
```

## Output

- QA report: `docs/workflow/qa-reports/upstream-email-alias-dedup-s157-qa.md`
- Final verdict must be `PASS`, `FAIL`, or `BLOCKED`, with executed checks and unverified risks.

## Stop Rules

- Stop if the final behavior requires Ent/schema changes beyond migration `190`.
- Stop if a registration path cannot distinguish grant-bearing account creation from ordinary email binding/update.
- Stop on migration-number collision or if validation requires writing to a configured/live database.
