# Task Contract: admin-account-plan-type-filter-s105

- Task ID: `admin-account-plan-type-filter-s105`
- Role: Planner / Generator / Evaluator
- Goal: Let administrators distinguish and filter OpenAI accounts by the
  persisted credential plan type while keeping list pagination, filtered bulk
  actions, and filtered export on one query contract.
- Success Criteria:
  - The account list exposes one plan filter with `Plus`, `Pro`, `K12`,
    `Team`, `Free`, `Other`, and `Unrecognized` categories.
  - Matching is case-insensitive; `pro` and `chatgptpro` both map to `Pro`,
    `k12` always displays as `K12`, non-empty unknown values map to `Other`,
    and missing/blank values map to `Unrecognized`.
  - A non-empty plan filter implicitly limits results to OpenAI accounts so
    another platform's plan-like value cannot match.
  - Filtering uses persisted `credentials.plan_type`; the optional
    `share_display_tier` remains display-only and never changes query results.
  - Repository filtering happens before count, offset, and limit so list total
    and pagination remain correct.
  - The same `plan_type` condition is propagated through ordinary listing,
    owner/share listing, filtered bulk edit, filtered share-status changes,
    and filtered export.
  - Existing filters, selected-ID bulk actions, selected-ID export, sorting,
    and account import behavior remain unchanged.
- Allowed Paths:
  - `backend/internal/service/account_service.go`
  - `backend/internal/service/admin_service.go`
  - `backend/internal/service/admin_account_plan_filter_test.go`
  - `backend/internal/service/admin_service_search_test.go`
  - `backend/internal/service/admin_service_bulk_update_test.go`
  - `backend/internal/service/openai_ws_ratelimit_signal_test.go`
  - `backend/internal/repository/account_repo.go`
  - `backend/internal/repository/account_repo_integration_test.go`
  - `backend/internal/handler/admin/account_handler.go`
  - `backend/internal/handler/admin/account_data.go`
  - `backend/internal/handler/admin/account_data_handler_test.go`
  - `backend/internal/handler/admin/admin_service_stub_test.go`
  - `frontend/src/api/admin/accounts.ts`
  - `frontend/src/components/admin/account/AccountTableFilters.vue`
  - `frontend/src/components/admin/account/__tests__/AccountTableFilters.spec.ts`
  - `frontend/src/components/common/PlatformTypeBadge.vue`
  - `frontend/src/components/common/__tests__/PlatformTypeBadge.spec.ts`
  - `frontend/src/views/admin/AccountsView.vue`
  - `frontend/src/views/admin/__tests__/AccountsView.bulkEdit.spec.ts`
  - `frontend/src/i18n/locales/en/admin/accounts.ts`
  - `frontend/src/i18n/locales/zh/admin/accounts.ts`
  - `docs/workflow/tasks/admin-account-plan-type-filter-s105.md`
  - `docs/workflow/qa-reports/admin-account-plan-type-filter-s105-qa.md`
  - `docs/workflow/spec.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: Ent schema and migrations, account import/enrichment, OAuth
  token handling, scheduler and gateway behavior, billing, deployment,
  containers, VERSION, dependencies, and unrelated workflow history.
- Constraints:
  - Add no database column or migration; query the existing JSONB credential
    field at repository level.
  - Normalize only for comparison and presentation; do not rewrite stored
    credential values and do not add manual plan editing.
  - Reject or neutralize unsupported plan filter values rather than treating
    arbitrary input as an SQL fragment.
  - Preserve selected-ID precedence for bulk operations and export.
  - Do not push, deploy, or update containers without separate authorization.
- Acceptance Commands:
  - `go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters" -count=1`
  - `go test ./internal/service -run "TestAdminService(ListAccountsPropagatesNormalizedPlanType|BulkUpdatePropagatesPlanTypeFilter|BulkShareStatusPropagatesPlanTypeFilter)" -count=1`
  - `go test ./internal/handler/admin -run "TestListAccountsPassesPlanTypeFilter|TestExportDataPassesAccountFiltersAndSort" -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/admin/account/__tests__/AccountTableFilters.spec.ts src/components/common/__tests__/PlatformTypeBadge.spec.ts src/views/admin/__tests__/AccountsView.bulkEdit.spec.ts"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"`
  - `gofmt -d` on the allowed Go files, `git diff --check`, conflict-marker
    scan, exact allowlist audit, and unmerged-index check.
- Output: Scoped implementation, focused backend/frontend regressions, QA
  report, local commit, and final `PASS`, `FAIL`, or `BLOCKED` evidence.
- Stop Rules: Stop on any need for schema/migration changes, persisted plan
  rewrites, import or OAuth behavior changes, scheduler/gateway changes,
  deployment/container work, or a path outside the approved allowlist.

## Contract Review

`PASS`: The existing account list already carries `credentials.plan_type` to
the UI, and list, share filters, filtered bulk actions, and export converge on
`AdminService.ListAccounts`. The repository can apply a bounded JSONB
predicate before count/pagination without a schema change. The duplicated
ordinary/share repository paths and ETag/list/bulk/export propagation are
explicitly covered by the allowlist and acceptance checks.
