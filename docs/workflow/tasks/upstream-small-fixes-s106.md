# Task Contract: upstream-small-fixes-s106

- Task ID: `upstream-small-fixes-s106`
- Role: Planner / Generator / Evaluator
- Goal: Manually port five isolated upstream fixes without merging upstream
  history: scheduler quota metadata (`e51e61008`), channel-monitor decrypt
  failure unscheduling (`bd62c6f66`), subscription validity-unit display
  (`a6ecc202f`), usage multiplier precision (`d0bdd7e77`), and promo-code
  local-time expiry prefill (`eba7289a8`).
- Success Criteria:
  - Scheduler cache snapshots preserve every field required by the local
    total, daily, weekly, and monthly quota checks while still dropping
    unrelated account `extra` data.
  - Enabled channel monitors with an API-key decryption failure are not
    scheduled; a decrypt failure returned by `RunCheck` removes the existing
    task; replacing the key allows scheduling again.
  - User subscription cards and payment selection display singular/plural
    `day`, `week`, and `month` units consistently with backend validity
    billing, and unknown units retain the backend day fallback.
  - Usage rate multipliers retain meaningful precision through four decimal
    places without losing the existing two-decimal minimum.
  - Editing a promo code prefills `datetime-local` using local wall-clock
    time rather than an ISO UTC slice.
  - Existing scheduler filters, ordinary monitor retry behavior, billing,
    promo persistence, and unrelated frontend presentation remain unchanged.
- Allowed Paths:
  - `backend/internal/repository/scheduler_cache.go`
  - `backend/internal/repository/scheduler_cache_unit_test.go`
  - `backend/internal/service/channel_monitor_runner.go`
  - `backend/internal/service/channel_monitor_runner_test.go`
  - `frontend/src/components/payment/SubscriptionPlanCard.vue`
  - `frontend/src/components/payment/__tests__/SubscriptionPlanCard.spec.ts`
  - `frontend/src/components/payment/validity.ts`
  - `frontend/src/components/payment/__tests__/validity.spec.ts`
  - `frontend/src/views/user/PaymentView.vue`
  - `frontend/src/i18n/locales/en/payment.ts`
  - `frontend/src/i18n/locales/zh/payment.ts`
  - `frontend/src/utils/formatters.ts`
  - `frontend/src/utils/__tests__/formatMultiplier.spec.ts`
  - `frontend/src/views/admin/PromoCodesView.vue`
  - `frontend/src/utils/__tests__/formatDateTimeLocalInput.spec.ts`
  - `docs/workflow/tasks/upstream-small-fixes-s106.md`
  - `docs/workflow/qa-reports/upstream-small-fixes-s106-qa.md`
  - `docs/workflow/spec.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: Schema and migrations, billing calculations, account quota
  persistence, scheduler selection algorithms, monitor CRUD/encryption,
  payment execution, dependency files, deployment, containers, VERSION, and
  unrelated upstream changes.
- Constraints:
  - Adapt behavior to the local topology; do not cherry-pick or merge upstream
    history.
  - Preserve local monthly quota semantics in addition to the upstream total,
    daily, and weekly cache fields.
  - Only `ErrChannelMonitorAPIKeyDecryptFailed` is terminal for a monitor task;
    all ordinary check failures keep their existing retry behavior.
  - Keep validity display aligned with `psComputeValidityDays`: week(s) means
    seven-day units, month(s) means 30-day units, and unknown values mean days.
  - Do not commit, push, deploy, or update containers without separate
    authorization.
- Acceptance Commands:
  - `go test -tags=unit ./internal/repository -run TestBuildSchedulerMetadataAccount_KeepsQuotaStateForCachedAccounts -count=1`
  - `go test -tags=unit ./internal/service -run "Test(Schedule_(DecryptFailedRedirectsToUnschedule|RepairedAPIKeyCanBeScheduled)|RunOne_DecryptFailureUnschedulesTask|Start_SkipsDecryptFailedMonitor)" -count=1`
  - `go test -tags=unit ./internal/repository ./internal/service -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/components/payment/__tests__/validity.spec.ts src/utils/__tests__/formatMultiplier.spec.ts src/utils/__tests__/formatDateTimeLocalInput.spec.ts"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"`
  - `gofmt -d` on the four allowed Go files, `git diff --check`, conflict-marker
    scan, exact allowlist audit, and unmerged-index check.
- Output: Scoped source changes, focused regressions, QA report, and final
  `PASS`, `FAIL`, or `BLOCKED` evidence. No automatic commit or push.
- Stop Rules: Stop on any need for schema/migration changes, quota persistence
  changes, billing changes, non-decrypt monitor retry changes, dependency
  updates, deployment/container work, or a business path outside the approved
  allowlist.

## Contract Review

`PASS`: All five upstream fixes have existing local behavior boundaries and
can be migrated without schema or API changes. Two patches apply directly;
the scheduler, monitor, and subscription patches require local-topology
adaptation. The local scheduler supports monthly quotas that upstream does not
preserve, so the approved cache field set explicitly includes that local
dimension and tests its scheduling effect.
