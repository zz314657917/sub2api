### PASS: upstream-main-ops-alert-temp-unscheduled-s14

## Findings

- PASS: The approved S14 candidate `f20e6bf76` was cherry-picked as `0fb09933c`.
- PASS: Backend validation and evaluator support the new `account_temp_unscheduled_count` metric.
- PASS: Unit-tagged service tests verify active temp-unscheduled windows are counted and expired windows are ignored.
- PASS: Frontend type union, alert rule selector, and modular en/zh locale strings were updated.
- PASS: No denied paths were changed.

## Executed Checks

- `git status --short --branch` -> clean on `codex/upstream-main-ops-alert-temp-unscheduled-s14` before workflow report edits.
- `git diff --check main...HEAD` -> PASS.
- denied path audit with `git diff --name-only main...HEAD` -> `DENIED_NONE`.
- `go test ./internal/service -run "OpsAlert|TempUnscheduled|AccountTemp|RuleMetric" -count=1` -> PASS, but reported `[no tests to run]` because `ops_alert_evaluator_service_test.go` is behind the `unit` build tag.
- `go test -tags unit ./internal/service -run "ComputeRuleMetric|TempUnscheduled|OpsAlert" -count=1` -> PASS.
- `go test ./internal/handler/admin -run "OpsAlert|Metric" -count=1` -> PASS.
- `go test ./internal/service ./internal/handler/admin -count=1` -> PASS.
- `corepack.cmd pnpm --dir frontend install --frozen-lockfile` -> PASS, created ignored `frontend/node_modules` in the isolated worktree.
- `corepack.cmd pnpm --dir frontend run typecheck` -> PASS.

## Not Run

- Full frontend build was not run; this Sprint only touched the Ops alert rule editor type/locale surface and `vue-tsc` passed.
- Live Ops alert evaluation against production data was not run; behavior is covered by local service tests.

## Risks

- The evaluator uses current wall-clock time to decide whether `TempUnschedulableUntil` is active, matching upstream behavior. Tests use a future/past window with enough slack to avoid boundary flake.
- Alert rules using the new metric require operators/thresholds chosen by admins; no default rule is created.

## Recommendation

PASS. The branch is ready for integration into current `main`.
