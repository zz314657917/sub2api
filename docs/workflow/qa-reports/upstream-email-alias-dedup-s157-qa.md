# Upstream Email Alias Dedup S157 QA

## Verdict

`PASS` for the isolated source-level port. This is not a production rollout or
a proof that migration `190` has run against PostgreSQL.

## Findings

- No new defect was found in the scoped diff.
- The initial compile-only run exposed one unrelated test-stub omission:
  `groupBuyUserRepoStub` did not yet implement the two new `UserRepository`
  methods. Both methods now intentionally panic if the group-buy tests invoke a
  registration-only path; `TestGroupBuy` passes after the adapter change.

## Executed Checks

- `go test ./... -run "^$" -count=1` passed from `backend/` after the stub
  adapter was completed. This compiles all production packages without the
  optional `unit` test build tag.
- `go test -tags=unit ./internal/repository -run '^(TestUserRepositoryExistsByEmailAlias|TestUserRepositoryExistsByEmailAliasIgnoresMalformedInput|TestUserRepositoryCreateWithEmailAliasGuard)$' -count=1 -v` passed.
  It covers Gmail/Googlemail aliases, plus tags, Gmail dots, the DNS root dot,
  soft-delete behavior, and escaped `_` / `%` LIKE literals, as well as the
  lock-protected create path.
- `go test ./internal/handler ./internal/server ./internal/server/middleware -run 'Test.*Alias|Test.*OAuth|Test.*Contract|Test.*Admin' -count=1` passed for the compiled tests (the server package had no matching non-`unit` test).
- `go test -tags=unit ./internal/server/middleware -run '^TestAdminAuthJWTValidatesTokenVersion$' -count=1 -v` passed.
- `go test ./internal/service -run 'TestGroupBuy' -count=1` passed.
- `gofmt` was applied to the changed Go test stub. `git diff --check` passed
  before final staging.
- Manual diff review confirmed that only registration, verification-code
  preflight, and OAuth account-creation paths use alias-aware probes and the
  guarded create path. Administrator creation and email binding/update remain
  on the exact-email `Create` / lookup paths.
- Migration `190` was reviewed as a concurrent partial expression index only;
  it was not executed against any database.

## Unverified Risks

- The focused service tests use `//go:build unit`, but the package currently
  cannot compile with that tag because of pre-existing unrelated failures:
  duplicate `stringPtr`, stale billing helper arities, and removed Grok runtime
  helper expectations. The newly added registration tests now match the local
  three-argument `newAuthService` helper; no S157-specific compile error remains
  in the reported output.
- The handler unit suite is separately blocked by an existing missing
  `strconv` import in `payment_handler_resume_test.go`.
- `TestAPIContracts` runs with `-tags=unit` but has existing snapshot drift for
  fields such as `allow_live`, audit/group-buy settings, and
  `default_fallback_group`; none of those API contracts are changed in S157.
- No real PostgreSQL migration, concurrent registration race, SMTP delivery,
  OAuth provider callback, deployment, or production smoke test was run.

## Recommendation

`可继续合入` this isolated branch for source-level review. Before deploying the
result, validate migration `190` on a non-production PostgreSQL copy and clear
the repository's unrelated `unit`-suite baseline failures so the focused
service and API-contract tests can execute.
