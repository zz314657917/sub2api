# Model-unavailable failover: scheduler amendment

Task ID: `model-unavailable-failover-20261030`

This amendment becomes effective only after an independent PASS review binds
the exact bytes. The original approved contract remains immutable at SHA256
`da964c460e0b768d12b136ab5821c547bd037532b0ec181a771eba1b9199a6f3`.

## Evidence and problem

R2 source freeze `33f54374aa329aa114acf5729a48f268b9914d97e7e429473f4d1c3f90f82962`
and all original FAIL logs remain preserved. Independent handler probes show
the next request returns 503 when the advanced scheduler uses LBTopK=1:
cooled A occupies the ranked shortlist; healthy B is omitted. The immutable
baseline has the same problem for its existing structured model_not_found.
`lb_top_k=1` is a valid production configuration; default 7 can encounter the
same issue when the first seven candidates are model-cooled.

Changing the test configuration from 1 to 2 demonstrates the root cause but
does not satisfy original success criterion 3. The controller rejects that as
a final acceptance workaround.

## Scope and ownership

Add only `backend/internal/service/openai_account_scheduler.go` to the original
product allowlist. The original Developer owns the amendment; QA remains
independent. Existing new model-unavailable test files may gain regression
cases. All seven R2 files remain unchanged unless required to add tests; the
three R2 error-owner files and helper must retain their R2 bytes.

Filter an account's existing account/model cooldown before ranking/top-K
truncation. Use the existing context-aware request model lookup semantics.
Retain DB rechecks and current capability, compact, transport, pool, quota,
channel, priority, load, sticky, concurrency and billing rules. Do not widen
the candidate list or change LBTopK defaults. Do not add a new cooldown store
or modify global account eligibility, dependencies or existing tests.

## Acceptance

- A valid LBTopK=1 with cooled A and eligible B must select B on the next
  actual handler request, including stale sticky rebound to A. Preserve
  `[A,B,B]` HTTP trace, successful B usage only and exactly-once releases.
- With default top-K=7 and at least eight accounts, seven higher-ranked
  model-cooled accounts must not hide the healthy eighth account.
- Another model remains eligible on the cooled accounts; account eligibility
  restrictions, compact/transport/capability checks and request-model mapping
  remain intact. Include non-idempotent alias mapping and OAuth normalized keys.
- Original priority four-case probe remains unchanged and passes both tags.
  Preserve original R1/R2 logs and R2 top-K=2 diagnostic.
- Correct only the external OAuth cooldown oracle: require exactly the actual
  upstream key plus the existing scheduler lookup mirror if distinct, same
  account/reset time, and no twice-mapped key. No test configuration workaround.
- Rerun original contract four commands and build/format, plus model-limit,
  sticky, scheduler, rate-limit and image cross regression. Compare failures
  and test inventory to the immutable baseline; retain original WS failure.
- QA freezes and validates the complete new eight-file source set, reports
  any remaining limitations, and binds evidence before protected integration.

## Protection and integration

Keep original protect.mjs and source-freeze-r1/r2.json unchanged. The versioned
protect-r3.mjs adds only the scheduler path and requires an independent PASS
binding to this amendment and its own script hash. Its original whole-tree,
HEAD/index, target baseline and candidate hash checks remain in force.
Any primary drift must fail closed. Preserve the concurrent task's documents
and product bytes; no whole-file overwrite after concurrent target drift.

Controller writes source-freeze-r3.json, makes it the active source-freeze.json
only after original-file protection passes, and independently audits the
candidate delta against the immutable baseline. Controller alone integrates.
No commit, push, container update, database operation or deployment.

## Stop rules

Any extra production path, existing-test edit, changed R2 production byte
outside the added scheduler path, dependency change or unresolved protected
primary drift returns to the controller before implementation/integration.
