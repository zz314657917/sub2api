### PASS: standard-group-time-rate-s211

# QA Report

## Task ID

`standard-group-time-rate-s211`

## Verdict

`PASS / fresh-review-2`

## Contract Checked

- `docs/workflow/tasks/standard-group-time-rate-s211.md`
- Second-round review scope: the two S211 blockers found during pre-commit
  review, covering generic Gateway per-request billing and preservation of the
  middleware request-start timestamp.

## Findings

- No remaining blocking issue was found in the second-round fresh review.
- Generic Gateway non-image `per_request` billing preserves the original
  effective multiplier before applying the time factor. With an original
  multiplier of `1.5`, a time factor of `0.7`, and a per-request price of
  `0.2`, the charged amount remains `0.3`; the time factor does not affect the
  per-request charge.
- For generic Gateway `per_request` billing, `UsageLog.RateMultiplier` records
  the same original effective multiplier used by the actual deduction. The
  regression above records `1.5`, matching the balance charge calculation.
- Gateway handlers reuse `ctxkey.RequestStartedAt` written by middleware and
  only fall back to `time.Now()` when the timestamp is missing or zero. The
  timestamp is therefore preserved for the usage inputs used by request,
  retry, failover, and asynchronous accounting paths reviewed in this round.

## Executed Checks

- Focused S211 service regressions for peak-rate calculation, validation and
  normalization, standard-group persistence, token billing, generic Gateway
  non-image `per_request` billing, image billing, and video billing were run
  with `-count=10` -> PASS.
- Focused handler regressions for `requestStartedAt` reuse/fallback and API Key
  billing claims were run with `-count=10` -> PASS.
- Focused middleware regression for preserving an existing
  `ctxkey.RequestStartedAt` was run with `-count=10` -> PASS.
- `go test ./internal/service -count=1` -> PASS (`62.527s`).
- `go test ./internal/handler ./internal/server/middleware -count=1` -> PASS
  (`internal/handler` `27.560s`; middleware `0.095s`).
- `go test ./cmd/server -run '^$' -count=0` -> PASS (`0.100s`; no tests to
  run), confirming the server package compiles.
- `gofmt -d` over the reviewed Go files produced no output -> PASS.
- `git diff --check` -> PASS; only existing LF/CRLF warnings were observed.
- `git ls-files -u` produced no entries -> PASS; no unmerged index state was
  present.
- Diff precision review confirmed the two regression tests assert the actual
  charged amount, persisted rate multiplier, middleware timestamp reuse, and
  internal-call fallback behavior.

## Unverified Risks

- No real browser acceptance flow was run in this second-round review.
- No real provider request, shared PostgreSQL/Redis integration, or production
  traffic was exercised.
- No local container build, replacement, health check, deployment, or remote
  push was performed by this review.

## Recommendation

`PASS / fresh-review-2`: the two S211 pre-commit blockers are resolved and the
reviewed S211 changes can proceed to the user-authorized local commit batches.
Container readiness still requires the separate guarded container update and
runtime health verification.
