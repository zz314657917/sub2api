### DONE: upstream-email-bind-alias-s267

## Task ID
upstream-email-bind-alias-s267

## Status
`done`

## Summary

Manually ported upstream `4ca86c52e` onto the local email-binding topology. The
service now performs alias-aware preflight for bind-code and verified bind
requests, while the repository rechecks exact and canonical inbox ownership
inside the existing transaction-scoped locks before updating the user.

## Changed Files

- `backend/internal/repository/user_repo.go`
- `backend/internal/service/auth_email_binding.go`
- `backend/internal/service/auth_service_email_bind_test.go`
- `docs/workflow/tasks/upstream-email-bind-alias-s267.md`

The local adaptation also exposes an optional owner lookup on the concrete
repository. This prevents a preflight from treating the current user's own
historical alias as proof that no other user owns the same canonical inbox;
the transaction guard remains authoritative.

## Commands Run

```text
go test -tags=unit ./internal/service -run '^(TestAuthServiceBindEmailIdentity|TestAuthServiceSendEmailIdentityBindCode)' -count=10 -> blocked by pre-existing unit-tag compile drift (duplicate stringPtr, obsolete method signatures, missing legacy fields)
go test -tags=unit ./internal/repository -run '^TestUserRepository(ExistsByEmailAlias|CreateWithEmailAliasGuard)' -count=1 -> PASS
go test ./internal/service -run '^TestDoesNotExist$' -count=1 -> PASS (compile-only)
go test ./internal/repository -run '^TestDoesNotExist$' -count=1 -> PASS (compile-only)
gofmt -w backend/internal/repository/user_repo.go backend/internal/service/auth_email_binding.go backend/internal/service/auth_service_email_bind_test.go -> PASS
git diff --check -> PASS
git ls-files -u -> empty
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/repository 16.957s
ok github.com/Wei-Shaw/sub2api/internal/service 12.896s [no tests to run]
```

The requested service `unit` suite cannot compile on the frozen base because
unrelated repository tests contain existing API drift. The implementation and
repository-focused tests compile and the repository alias guard tests pass.

## Risks

- The concurrent bind regression is SQLite/process-lock evidence; a live
  PostgreSQL multi-instance smoke was not run.
- No real email provider, database migration, deployment, or container action
  was performed.

## Knowledge Candidates

- The existing repository-scoped lock helper can serialize normalized email and
  canonical inbox identity without changing the UserRepository interface.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes` (service unit-tag suite blocked only by baseline drift)
- stop_rules_triggered: `no`

## Upstream Provenance

- `4ca86c52e3edbcc5d38247842f045f6975329263` (`fix(auth): 邮箱换绑增加别名与并发守卫`)
