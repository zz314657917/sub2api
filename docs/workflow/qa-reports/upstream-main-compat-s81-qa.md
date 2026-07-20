### PASS: upstream-main-compat-s81

## Findings

- No blocking finding remains. The service diff is confined to the existing
  admin-assignment reuse branch and follows upstream `1db10dc55` behavior while
  adapting cache invalidation to local topology.
- Expired-status and active-but-time-expired rows renew; suspended rows do not.
- Renewal reuses the existing transaction/term-reset helper, preserves ID and
  assignment provenance, resets windows/usage, deduplicates trim-equivalent
  admin notes, and returns the refreshed row as reused.
- `AssignOrExtendSubscription`, `updateExistingSubscriptionTerm`,
  `renewedSubscriptionTerm`, and `appendSubscriptionNotes` implementations are
  unchanged. A new regression locks the purchase/redeem equal-note behavior.
- No handler, DTO, repository, schema, migration, billing/payment/redeem,
  frontend, deploy, dependency, VERSION, or `knowledge/**` path is changed.

## Executed Checks

- Pre-change local-main baseline: four existing assignment/purchase tests PASS.
- Exact default-tag discovery found all 11 contract tests.
- Focused 11-test run: PASS.
- Focused 11-test run with `-count=10`: PASS.
- Broader assignment, bulk, purchase, semantic-conflict, subscription-window,
  and limit regression selection: PASS.
- Final Evaluator used an independent workspace-local `GOCACHE/GOTMPDIR` and
  reran exact discovery plus all 11 tests: PASS.
- Line-level service diff review: only 27 lines inside
  `assignSubscriptionWithReuse`; frozen helper implementations have no diff.
- `gofmt` over all business files: PASS.
- Pre-report eight-path allowlist audit: PASS.
- `git diff --check`, unmerged-index check, and exact conflict-marker scan: PASS.
- All three primary-checkout protected SHA-256 values match the contract.

## Unverified Risks

- No live database transaction or Redis/billing-cache integration smoke was
  performed. Existing transaction and invalidation paths are reused directly.
- Concurrent duplicate admin assignments retain the repository's existing
  check/get/update concurrency behavior; S81 does not introduce locking.
- Tests bound `time.Now()` before/after renewal because the service does not
  inject a clock. Ten repeated focused runs did not expose flakiness.
- The full service package was not used as S81 evidence; this repository has
  known unrelated broad-suite drift. Exact and broader subscription selections
  are green.

## Recommendation

`PASS` — create the scoped S81 commit after force-tracking all three workflow
evidence files and passing the exact nine-path staged gate. Keep the branch
isolated; do not merge, push, deploy, or update containers automatically.
