### BLOCKED: upstream-email-bind-alias-s267

## Scope and provenance

- QA worktree: `E:/codex-worktrees/sub2api/upstream-email-bind-alias-s267`.
- Tested HEAD: `bfecbfd9155cd6c0f520dcc9268a9ba7a1a71931`.
- Frozen base: `659ad5d1161589ed80bf505afd75e345376669e7`.
- Upstream behavior source named by the contract: `4ca86c52e3edbcc5d38247842f045f6975329263`.
- Relative changed paths are the three approved backend files plus the contract and worker result; no denied path, conflict, or unmerged index entry was found.

## Commands and evidence

- `go test -tags=unit ./internal/repository -run '^TestUserRepository(ExistsByEmailAlias|CreateWithEmailAliasGuard)' -count=1` -> PASS.
- `go test ./internal/service ./internal/repository -run '^$' -count=1` -> PASS (both packages compile).
- `go test ./cmd/server -run '^$' -count=1` -> PASS (server compiles).
- `go test ./internal/service -run '^(TestAuthServiceBindEmailIdentity|TestAuthServiceSendEmailIdentityBindCode)' -count=1` -> command completed with `[no tests to run]` because these tests require the `unit` build tag; this does not constitute service coverage.
- Contract command `go test -tags=unit ./internal/service -run '^(TestAuthServiceBindEmailIdentity|TestAuthServiceSendEmailIdentityBindCode)' -count=10` was attempted. It is BLOCKED before test execution by pre-existing package compile drift: duplicate `stringPtr` (`ops_health_score_test.go` vs `usage_leaderboard_reward.go`) and obsolete billing/gateway/proxy test APIs. These errors are outside this change and were not repaired.
- `gofmt -d internal/repository/user_repo.go internal/service/auth_email_binding.go internal/service/auth_service_email_bind_test.go` -> no output.
- `git diff --check -- <approved paths>` -> PASS.
- `git ls-files -u` -> empty; conflict/index check PASS.

## Behavioral review

The diff implements bounded alias candidate probing with normalized final matching, preserves self-alias allowance, performs exact/alias recheck inside the caller transaction, and acquires both normalized-email and inbox-identity repository-scoped locks before updating email/password. Auth-identity replacement remains after the guarded update in the same transaction, so rollback paths remain transactional. Added tests cover foreign alias rejection, one-winner concurrent variants, self/foreign shared inbox behavior, and first-bind rollback paths; the service package compile drift prevents executing them under the required unit tag.

## Residual risk

Runtime service alias/concurrency tests are not executable until the unrelated existing unit-test compile drift is repaired. No provider, database, container, deployment, staging, push, or primary-worktree operation was performed.

This report therefore proves repository checks and compile gates only; it does not constitute a complete alias/concurrency QA PASS. Re-run acceptance after the unit-tag service tests become executable.
