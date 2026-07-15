# Task Contract

- Task ID: `upstream-billing-rate-sync-s204`
- Role: Planner, Generator, and Final Evaluator (direct Codex implementation in an isolated worktree; no worker is delegated).
- Goal: Behaviorally port upstream Sub2API declared-rate introspection, periodic account probing, and the per-account opt-in that atomically synchronizes a validated upstream base rate into `accounts.rate_multiplier`.

## Success Criteria

- Authenticated API Keys can read a bounded billing declaration from the local gateway, including the effective downstream rate facts required by another Sub2API instance to probe this deployment.
- Eligible API-key accounts can be probed manually and by the existing CRS synchronization loop without exposing credentials or following an unsafe upstream URL.
- Probe snapshots preserve status, declared base/effective rate, peak window metadata, timestamps, failure count, next probe time, and an optional `synced_rate_multiplier` audit fact.
- A per-account `upstream_billing_rate_sync_enabled` flag is stored in account `extra`. Enabling it also enables probing; disabling or removing probe eligibility disables synchronization.
- Only a successful, finite upstream-declared base rate in `(0, 100]` that remains non-zero at local `DECIMAL(10,4)` precision may update `rate_multiplier`. Failed, unsupported, stale, invalid, or out-of-range declarations leave the account rate unchanged.
- Snapshot and optional rate write-back occur in one compare-and-swap update so stale probe results cannot overwrite concurrent account identity, proxy, settings, probe-state, or administrator changes.
- When synchronization is enabled, single-account and bulk manual rate updates fail with a dedicated conflict. A request that disables synchronization and supplies a manual rate in the same update remains valid.
- The admin account list/editor expose probe status, manual/batch probe actions, the sync toggle, managed-rate indicator, precision-preserving rate display, and clear Chinese/English guidance.
- Existing account duplication, proxy changes, CRS scheduling, billing, gateway authentication, account rate accounting, local modular i18n, Pixel Cafe, and S180/S203 work remain intact.

## Upstream Sources

- `f59a6ed74` API-key billing-rate introspection.
- `0765d10c1` initial upstream billing probe and admin display.
- `d2585097f` probe persistence/display hardening.
- `2447c44f8` nanosecond `next_probe_at` parsing.
- `f3a3d8684` plus `56f3d3c9b` explicit API-key platform eligibility and trust-boundary fixes; port only probe eligibility behavior, not the separate upstream-cost scheduler feature.
- `b0f5007f0` optional account rate write-back.
- `0b6b4ea95` write-back governance, base-rate semantics, traceability, and manual-edit conflict protection.

## Allowed Paths

- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`
- `backend/internal/handler/admin/account_handler*.go`
- `backend/internal/handler/admin/account_upstream_billing_probe*.go`
- `backend/internal/handler/admin/admin_service_stub_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_key_billing*.go`
- `backend/internal/handler/wire.go`
- `backend/internal/repository/account_repo*.go`
- `backend/internal/repository/http_upstream*.go`
- `backend/internal/repository/proxy_repo*.go`
- `backend/internal/repository/upstream_billing_probe*.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/server/middleware/api_key_auth*.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_key_billing_test.go`
- `backend/internal/service/account_service.go`
- `backend/internal/service/admin_account*.go`
- `backend/internal/service/admin_service*.go`
- `backend/internal/service/crs_sync*.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/gateway_usage_billing.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/http_upstream_profile*.go`
- `backend/internal/service/openai_endpoint_url*.go`
- `backend/internal/service/openai_gateway_usage.go`
- `backend/internal/service/ops_advisory_lock.go`
- `backend/internal/service/proxy_update_probe_invalidation_test.go`
- `backend/internal/service/upstream_billing_probe*.go`
- `backend/internal/service/wire.go`
- `frontend/src/api/admin/accounts.ts`
- `frontend/src/api/__tests__/admin.accounts.upstreamBillingProbe.spec.ts`
- `frontend/src/components/account/**`
- `frontend/src/components/admin/account/AccountBulkActionsBar.vue`
- `frontend/src/components/common/DataTable.vue`
- `frontend/src/components/common/__tests__/DataTable.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/admin/AccountsView.vue`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/AccountsView*.spec.ts`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/upstream-billing-rate-sync-s204.md`
- `docs/workflow/qa-reports/upstream-billing-rate-sync-s204-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, `backend/go.mod`, `backend/go.sum`, `deploy/**`, `Dockerfile*`, Pixel Cafe product files, S180 banner files, and S203 account+model transient-breaker files.
- Upstream profitability/admission scheduling (`90ee85f3e`, group profit control, minimum margin, safety buffer), image/video account admission, new dependencies, schema changes, shared database/Redis writes, containers, deployment, production traffic, push, or remote publication.

## Constraints

- Port behavior, not upstream history or whole-file snapshots. Preserve local custom billing, gateway, account, proxy, CRS, modular i18n, frontend column, and Wire topology.
- The declared rate is untrusted remote input. Validate URL, identity, response size, content type, numeric finiteness, range, precision, and stale identity before persistence or write-back.
- Probe failure may update only the bounded probe snapshot/backoff state. It must not disable the account, alter credentials/proxy, or change `rate_multiplier`.
- Probe logs and API responses must not reveal API keys, OAuth credentials, proxy credentials, or raw upstream response bodies.
- No browser automation is required unless focused component tests reveal a presentation issue; S180 browser QA remains separate.

## Acceptance Commands

```powershell
cd backend
go generate ./cmd/server
gofmt -w <changed Go files>
go test ./internal/handler -run 'TestGatewayKeyBilling|TestAccount.*UpstreamBilling|Test.*Probe' -count=1
go test ./internal/server/middleware ./internal/server/routes -run 'Test.*Billing|Test.*Probe|TestAPIKey' -count=1
go test ./internal/repository -run 'Test.*UpstreamBillingProbe|Test.*Account.*Probe|Test.*Proxy.*Probe' -count=1
go test ./internal/service -run 'Test.*UpstreamBillingProbe|Test.*RateSync|Test.*BulkUpdate|Test.*CRSSync|Test.*Account.*Probe' -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=0

cd ../frontend
npm.cmd run test:run -- src/api/__tests__/admin.accounts.upstreamBillingProbe.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts src/components/account/__tests__/BulkEditAccountModal.spec.ts src/components/account/__tests__/UpstreamBillingRateCell.spec.ts src/views/admin/__tests__/AccountsView.bulkEdit.spec.ts src/views/admin/__tests__/AccountsView.usageWindowsHint.spec.ts
npm.cmd run typecheck
npm.cmd run build

cd ..
git diff --check
git diff --name-only
git ls-files -u
```

## Output

- Produce one scoped implementation commit and one workflow/QA closeout commit on `codex/upstream-billing-rate-sync-s204`.
- Write `docs/workflow/qa-reports/upstream-billing-rate-sync-s204-qa.md` with a first-line `PASS`, `FAIL`, or `BLOCKED` verdict and fresh evidence.
- After final review, integrate only the scoped commits into local `main`. Do not push or deploy.

## Stop Rules

- Stop if implementation requires a migration, Ent schema change, dependency change, production/shared resource, or importing the separate upstream profitability scheduler.
- Stop if the gateway introspection endpoint cannot preserve existing API-key auth and group/user rate semantics without exposing private account-cost data.
- Stop if CAS protection cannot distinguish stale probe results from concurrent administrator edits.
- Stop if focused tests show local custom billing, gateway routing, account duplication, proxy invalidation, CRS scheduling, or frontend account workflows regress.
