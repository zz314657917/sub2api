### PASS: upstream-billing-rate-sync-s204

> Final verdict: `PASS / scoped` (2026-08-07 22:40 +08:00). This closeout is limited to the S204 contract and local-main integration.

# Post-Merge Reaudit

- The prior FAIL was caused by stale test fixtures and several local integration defects. Those S204 defects are
  repaired and the scoped implementation now passes the default backend, frontend, Wire, and CAS checks below.
- Full integration remains non-green for unrelated baseline tests: usage-log data isolation assertions,
  `deleted_api_key_audits` migration availability, and an old balance-error expectation. These failures do not
  touch S204 files or behavior and are retained as unverified baseline risks.

# Scope

- Behavior-level local adaptation of upstream API-key billing introspection, bounded account probing, and
  opt-in CAS-protected synchronization of a validated declared base rate into `accounts.rate_multiplier`.
- No schema, migration, dependency, profitability scheduler, container, deployment, production traffic, push,
  or remote publication is included.

# Findings

- No remaining contract-scope issue was found. The final integrated branch preserves S203 and local account,
  gateway, CRS, scheduler snapshot, proxy invalidation, and frontend account workflows.
- Final review confirmed the probe uses bounded request/response handling, suppresses official provider domains,
  persists failures without changing the account rate, and writes only a finite validated base rate in `(0, 100]`
  that remains non-zero at four-decimal account precision.
- Snapshot and optional rate write-back share one identity/settings/snapshot CAS. Eligible administrator edits use
  the atomic billing updater even when the rate is omitted, so a concurrent probe write-back is not overwritten.
- Initial full-service QA exposed the missing atomic routing for name-only edits. Frontend QA also exposed missing
  sync state declarations/replay and the multiplier formatter import. These implementation defects were fixed and
  all affected and full suites passed on the final integrated branch.

# Executed Checks

```text
go generate ./cmd/server
PASS; wire_gen.go regenerated with no remaining working-tree diff

go test ./internal/handler -run "TestGatewayKeyBilling|TestAccount.*UpstreamBilling|Test.*Probe" -count=1
PASS

go test ./internal/handler/admin -run "Test.*Account.*|Test.*UpstreamBilling|Test.*Probe" -count=1
PASS

go test ./internal/server/middleware ./internal/server/routes -run "Test.*Billing|Test.*Probe|TestAPIKey" -count=1
PASS

go test ./internal/repository -run "Test.*UpstreamBillingProbe|Test.*Account.*Probe|Test.*Proxy.*Probe" -count=1
PASS

go test ./internal/repository -count=1
PASS (1.585s)

go test ./internal/service -run "Test.*UpstreamBillingProbe|Test.*RateSync|Test.*BulkUpdate|Test.*CRSSync|Test.*Account.*Probe" -count=1
PASS

go test ./internal/service -count=1
PASS (62.237s)

go test -run 'TestUpstreamBillingProbeManualAndScheduledRequestsDoNotShareModeResult|TestUpdateAccountRejectsProbeEnableForNonProbeIdentity' ./internal/service -count=1
PASS; manual/scheduled singleflight modes are isolated and non-probe enable is rejected

go test -tags=integration ./internal/repository -run 'TestAdminAccountEditPreservesRateSynchronizedAfterLoad|TestProbeSnapshotCASProtectsManualRateAfterSyncDisabled' -count=1
PASS; real PostgreSQL/Redis Testcontainers CAS checks

go test -tags=integration ./internal/repository -run '^$' -count=0
PASS; integration package compiles

go test -tags=unit ./internal/server/middleware -run '^TestAPIKeyAuthBillingInfo' -count=1
PASS

go test ./cmd/server -run "^$" -count=0
PASS; compile probe

npm.cmd run test:run -- <six S204 frontend spec files>
PASS (6 files, 80 tests)

npm.cmd run typecheck
PASS

npm.cmd run build
PASS; only existing Browserslist, DEP0190, chunk-size, and import-layout warnings

gofmt -w <changed Go files>
PASS

git diff --check
PASS

git ls-files -u
PASS (empty)

merge-base allowlist and denied-path audit
PASS; no Ent, migration, dependency, deploy, Docker, profitability scheduler, or S203 source authored by S204
```

# Unverified Risks

- No real upstream Sub2API instance, API key, external provider endpoint, live proxy, PostgreSQL/Redis service,
  authenticated browser session, container, deployment, rollback, or production traffic was exercised.
- The complete `go test -tags=integration ./internal/repository -count=1` suite still has five pre-existing,
  non-S204 failures involving cross-test data isolation, the unavailable `deleted_api_key_audits` table, and an
  old balance-error assertion. The S204 integration compile probe and targeted CAS cases pass.
- The complete `go test -tags=unit ./internal/service` package remains blocked by unrelated legacy test symbols;
  default service tests and all S204 service cases pass.
- The probe remains opt-in and disabled by default. A remote declaration is treated as untrusted input, but an
  operator still owns the decision to enable synchronization for a specific account.
- Frontend production build retains pre-existing warnings about stale Browserslist data, large chunks, mixed
  static/dynamic imports, and Node `DEP0190`; none blocked this scoped build.

# Recommendation

The bounded S204 implementation is ready for local-main integration. Do not infer live upstream correctness,
deployment readiness, or production behavior from this local regression evidence.
