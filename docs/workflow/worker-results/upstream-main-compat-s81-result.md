### PASS: upstream-main-compat-s81-generator

## Changed Files

- `backend/internal/service/subscription_service.go`
- `backend/internal/service/subscription_assign_idempotency_test.go`
- `backend/internal/service/user_subscription_daily_quota_test.go`

## Implemented Behavior

- Admin assignment now renews an existing row when its status is expired, or
  when its term is past expiry and it is not suspended.
- Renewal reuses the existing transaction/term helper, resets status/windows/
  usage, normalizes/caps validity, preserves ID and assignment provenance, and
  invalidates the local L1 plus billing subscription cache using the existing
  local pattern.
- Trim-equivalent admin notes are not duplicated; different notes append once.
- Future and past-expiry suspended subscriptions remain untouched.
- Non-expired active reuse/conflict behavior remains unchanged.
- `AssignOrExtendSubscription` code is unchanged; its equal-note append behavior
  now has an explicit regression test.

## Commands Run

- Baseline before S81: four existing assignment/purchase tests PASS on local
  `main=f62f8bbce`.
- Exact default-tag discovery found all 11 required S81 tests.
- Focused 11-test run: PASS.
- Focused 11-test run with `-count=10`: PASS.
- Broader assignment, bulk, purchase, semantic-conflict, subscription-window,
  and limit regression pattern: PASS.
- `gofmt` over all three business paths: PASS.
- `git diff --check`: PASS (workflow line-ending notices only).

## Risks / Deferred Checks

- Tests use time bounds around `time.Now()` because the local service has no
  injected clock; ten repeated runs showed no boundary flake.
- No live database transaction or Redis/billing-cache integration test was run.
  The service branch reuses existing transaction and cache-invalidating code.
- Concurrent duplicate admin assignment behavior is unchanged from the existing
  check/get/update design and is outside this upstream patch.
