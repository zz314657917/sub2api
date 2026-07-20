### PASS: upstream-main-compat-s85-generator

## Changed Files

- `backend/internal/handler/failover_loop.go`
- `backend/internal/handler/failover_loop_test.go`

## Implemented Behavior

- Computes local `sameAccountRetry` before the cache-billing decision.
- Bound sessions no longer force cache billing during same-account retries.
- After retry exhaustion and account switch, bound sessions still force cache
  billing.
- Explicit `failoverErr.ForceCacheBilling` remains effective during retries.
- Retry counts/delays, temporary unscheduling, switch counts, and account
  selection behavior are unchanged.
- Added focused cache-billing subtests and updated the existing integration
  scenario to assert the new transition.

## Commands Run

- Focused `TestHandleFailoverError_CacheBilling`: PASS.
- Broader `TestHandleFailoverError_` selection: PASS after updating the
  pre-existing old-semantic integration assertion.
- `gofmt`: PASS.
- Static ForceCacheBilling state gate, `git diff --check`, unmerged-index, and
  conflict-marker scans: PASS.

## Risks / Deferred Checks

- The broader handler test run initially failed only because its old assertion
  expected bound-session billing during same-account retries; the test was
  updated within the allowlist and the focused/broader rerun passed.
- No live billing or upstream integration smoke was run; the change only
  controls the existing ForceCacheBilling flag.
- Primary Usage S82 changes and workflow files remain external dirty work.
