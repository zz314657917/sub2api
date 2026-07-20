### PASS: upstream-main-compat-s85

## Findings

- No blocking finding remains. The upstream `a2acbf553` behavior is ported to
  the local failover state machine without changing retry or switch mechanics.
- A bound session no longer forces cache billing during same-account retries;
  the first actual switch still sets `ForceCacheBilling`.
- Explicit `failoverErr.ForceCacheBilling` remains effective during retries.
- The pre-existing integration test assertion was updated because it encoded
  the exact old behavior corrected by S85.
- No other handler, service, billing, account-selection, or deployment path
  changed.

## Executed Checks

- Generator focused `TestHandleFailoverError_CacheBilling`: PASS.
- Generator broader `TestHandleFailoverError_` selection: PASS.
- Fresh Evaluator focused test with independent cache: PASS.
- Fresh Evaluator broader handler selection with independent cache: PASS.
- `gofmt`: PASS.
- ForceCacheBilling state-transition static check: PASS.
- Line-level business diff, `git diff --check`, unmerged-index, and
  conflict-marker scans: PASS.
- All three primary-checkout protected SHA-256 values: PASS.

## Unverified Risks

- No live billing/account-switch smoke or upstream integration was run; this
  patch only changes the existing handler flag decision.
- No deployment/container operation was performed.
- Primary Usage S82 changes and workflow files remain external dirty work and
  are not included in S85.

## Recommendation

`PASS` — create the scoped S85 commit after the exact eight-path tracking,
conflict-marker, cached-diff, and protected-hash gates pass. Keep the branch
isolated until the primary S82 workflow conflict is reconciled.
