### DONE: upstream-v0177-group-pricing-long-context-s220

# Worker Result

## Task ID
upstream-v0177-group-pricing-long-context-s220

## Status
done

## Summary
- Ported the approved OpenAI account long-context billing veto and usage-log
  audit trail into the local group-pricing implementation, including strict
  migration normalization, account/CRS write handling, DTO persistence, and
  Create Account controls.
- Kept group pricing resolution and long-context behavior covered by five
  default-tag contract tests. The OpenAI account veto applies only to OpenAI;
  Grok continues to follow the group switch alone.
- Corrected migration 220 so malformed legacy account values are normalized
  before the strict trigger is installed. A fresh task-owned PostgreSQL
  fixture proved string, numeric, and missing legacy values backfill to false,
  valid true remains true, new OpenAI rows default false, malformed later
  writes fail with SQLSTATE 22023, and a second migration run is idempotent.

## Changed Files
- `backend/ent/migrate/schema.go`
- `backend/ent/mutation.go`
- `backend/ent/runtime/runtime.go`
- `backend/ent/schema/usage_log.go`
- `backend/ent/usagelog.go`
- `backend/ent/usagelog/usagelog.go`
- `backend/ent/usagelog/where.go`
- `backend/ent/usagelog_create.go`
- `backend/ent/usagelog_update.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/repository/openai_long_context_billing_migration_integration_test.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_long_context_billing_test.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/crs_sync_long_context_billing_test.go`
- `backend/internal/service/crs_sync_service.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/group_pricing_long_context_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/usage_log.go`
- `backend/migrations/220_openai_long_context_billing.sql`
- `backend/migrations/openai_long_context_billing_migration_test.go`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/types/index.ts`
- `docs/workflow/worker-results/upstream-v0177-group-pricing-long-context-s220-result.md`

## Commands Run
```text
go generate ./ent -> pass
go test ./migrations -run '^(TestMigration220|TestMigration221|TestOpenAILongContextBillingMigration)' -count=1 -> pass
go test ./internal/service -list <five contract tests> -> all five discovered
go test ./internal/service -run <five contract tests> -count=10 -> pass
go test ./internal/service -count=1 -> pass (61.609s)
go test ./internal/handler -count=1 -> pass (27.449s)
go test ./internal/repository -count=1 -> pass
go test ./internal/server -count=1 -> pass
go test ./cmd/server -run '^$' -count=0 -> pass
pnpm.cmd exec vitest run <five focused files> -> 5 files / 24 tests passed
pnpm.cmd run typecheck -> pass
pnpm.cmd run build -> pass (Vite built in 21.57s)
git diff --check -> pass
```

## Test Output
```text
Fresh disposable PostgreSQL fixture: migration 220 reported UPDATE 3 and
CREATE TRIGGER. Legacy OpenAI values became false, false, false, true; the
non-OpenAI row was unchanged. New OpenAI insert stored false. A string write
was rejected by enforce_openai_long_context_billing_extra with SQLSTATE 22023.
The second migration run reported UPDATE 0 and recreated the trigger safely.
```

## Risks
- `go test -tags=unit ./internal/service` remains a pre-existing unrelated
  compile baseline failure (duplicate `stringPtr`, stale billing signatures,
  and other incompatible test fixtures). The three pricing contract tests were
  moved to a new default-tag file and now execute in the accepted focused and
  complete default-tag service suites.
- The repository's Docker/Testcontainers migration harness was not used because
  no Linux Docker engine was available. The portable PostgreSQL proof used only
  the task-owned `sub2api_s220_m220` database and will be removed after this
  report; no shared or production database was accessed.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes (including approved Amendment 7 test path)
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.
