# Task Contract: upstream-subscription-days-round-up-s100

- Task ID: `upstream-subscription-days-round-up-s100`
- Role: Planner / Generator / Evaluator
- Goal: Port upstream `d0fa8c63f` so any positive partial subscription day is
  reported as one remaining day instead of being truncated to zero.
- Success Criteria:
  - Expired and exactly-expiring subscriptions report zero remaining days.
  - Positive durations round up to the next whole 24-hour day.
  - Exact whole-day durations remain unchanged.
  - Progress DTOs use the same rounded value without changing expiry time,
    quota windows, billing, or subscription status semantics.
- Allowed Paths:
  - `backend/internal/service/user_subscription.go`
  - `backend/internal/service/user_subscription_days_remaining_test.go`
  - `backend/internal/service/subscription_calculate_progress_test.go`
  - `docs/workflow/tasks/upstream-subscription-days-round-up-s100.md`
  - `docs/workflow/qa-reports/upstream-subscription-days-round-up-s100-qa.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: frontend, Ent, migrations, repositories, handlers, billing,
  payment, deployment, containers, VERSION, and unrelated workflow history.
- Constraints:
  - Use a deterministic `daysRemainingAt(now)` helper for boundary tests.
  - Keep the new focused test in the default test set because the local
    service `unit` aggregate has unrelated compile drift.
  - Keep `DaysRemaining()` as the public call site and preserve zero for any
    non-positive duration.
  - Do not change calendar-day or timezone semantics; this remains a 24-hour
    duration calculation matching upstream behavior.
- Acceptance Commands:
  - `go test ./internal/service -run "TestUserSubscriptionDaysRemainingAt|TestCalculateProgress_BasicFields" -count=1`
  - `gofmt -d` on the three allowed Go files.
  - `git diff --check`, conflict-marker scan, exact allowlist audit, and
    unmerged-index check.
- Output: Scoped source diff, focused regression tests, QA report, and final
  `PASS`, `FAIL`, or `BLOCKED` evidence.
- Stop Rules: Stop on any required calendar-day semantic change, persistence
  change, billing change, or path outside the exact allowlist.

## Contract Review

`PASS`: The upstream patch is independently applicable to the local service,
has deterministic boundary tests, and does not require upstream file-layout,
schema, migration, frontend, or runtime deployment prerequisites.
