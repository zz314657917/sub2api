---
type: task-contract
status: approved
review_verdict: PASS
task_id: upstream-gap-fixes-20261103-test-addendum
worker_model: gpt-6.1-sol
qa_worker_model: gpt-6.1-sol
base_commit: 7ea675c020ff2c5d5e4124ac5e249c9dc65034c5
spec_ref: docs/workflow/spec.md
last_verified: 2026-11-03
---

# Regression expectation alignment

## Task ID
upstream-gap-fixes-20261103-test-addendum

## Role
Current controller is Planner/Final Evaluator. Restore the original independent Contract Reviewer and Developer; the existing independent QA performs final acceptance. All three roles retain the repository matrix model gpt-6.1-sol. Approval metadata is conditional on an independent PASS binding the exact final bytes of this addendum.

## Goal
Align two existing test owners with the already approved future-reset pause and account-cost gate behavior, while retaining the old self-healing and settlement protections. No production implementation change is authorized by this addendum.

## Context
- Parent contract: `docs/workflow/tasks/upstream-gap-fixes-20261103.md`, SHA256 `b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0`. Its constraints, exclusions and stop rules remain effective.
- Runtime root: `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103`.
- Candidate: `E:/codex-worktrees/sub2api/upstream-gap-fixes-20261103`.
- Frozen R1 identity: `candidate-freeze.json` in the runtime root. All nine bound paths, including unchanged account_stats_pricing.go and the three new tests, must remain byte-identical.
- Controller reviewed 20 existing regression files / 248 top-level declarations. Original main passes unit 670 / default 619 leaves; R1 passes unit 653 / default 602 with the same 17 failing leaves under both tags. Evidence: `existing-regression-selection.json`, `main-original-regression-result.json`, `controller-candidate-verification.json` and their immutable command logs.
- One failure is `TestOpenAIGatewayService_SelectAccountForModelWithExclusions_StaleUsageSnapshotSkipsPause_Issue2994`: a three-hour-old 99% snapshot has a reset one hour in the future but expects selection of the exhausted primary.
- Sixteen failures are APIKey / account=false / n=272001 leaves of `TestOpenAISubscriptionGPT61SettlementMatrix`; the old account-stats expectation ignores accountGate. Other customer-charge, credential, group, resolver, tier, cache and precedence assertions must remain.

## Success Criteria
1. Preserve the scheduler test's original top-level name and issue context. Replace its unconditional stale bypass fixture with explicit subcases: stale snapshot without a known future reset selects primary 35901 (self-healing retained); stale snapshot with a known future reset selects secondary 35902 (pause retained). Keep actual account-selection assertions, fixed ID/priority contrast and a fresh per-case fixture. Update only the directly contradictory comment. Keep adjacent fresh-exhausted and reset-reached tests unchanged.
2. In the settlement matrix, only the account-stats long-context multiplier predicate and its explanatory comment may change: long-context account stats require APIKey, input above 272000 and accountGate=true. Group gate and request service tier still must not alter model-file account stats. Preserve customer charge and every other expectation, all loops and case names.
3. Do not delete, skip, xfail, rename away, reduce coverage or loosen tolerances in existing tests. Do not alter helpers or production code to obtain PASS. Keep the original 248 top-level regression declarations discoverable.
4. Run the same frozen regression selection under default and unit tags; all leaves must pass. The explicit scheduler split adds one leaf, so the expected counts are default 620 / unit 671 unless an independently explained inventory proves otherwise.
5. Using aligned tests against original production through a private Go overlay must expose the two intended old behaviors (future-reset primary selection and account-off long-context stats). Preserve that RED evidence; candidate must be GREEN.
6. R1 production and new-test hashes stay unchanged throughout. Keep original failing logs and all initial R1 identities. Form a new R2 identity binding the original nine paths plus these two existing tests (eleven paths; ten actual changed paths).
7. Independent QA reviews the two-test diff, reruns its prepared probes and the same full regression selection against the exact R2 identity. The four already-attributed unrelated failures remain explicit limitations; neither a broad failing suite nor its known failures may be called fully green.

## Allowed Paths
Developer may edit only, inside the candidate:
- `backend/internal/service/openai_account_scheduler_test.go`
- `backend/internal/service/openai_subscription_long_context_test.go`
- `docs/workflow/worker-results/upstream-gap-fixes-20261103-result.md` (append R2 change/evidence, keep R1 history)

Controller may write this addendum, its independent review, task QA/report and bounded status/spec/current-task updates, and external runtime R2 tooling/evidence. Controller integration allowlist expands by exactly the two test owners after independent acceptance; original main-before.json remains the preimage authority.

## Denied Paths
- Any production source, original R1 bound path, dependency/configuration/schema/migration, knowledge file or other existing test.
- Main shared source until independently accepted protected integration.
- All unrelated dirty/untracked work.
- Git staging, commits, push, history operations, deployment, containers, provider/API/database access and credential handling.

## Constraints
- Use the fixed installed Go toolchain and readonly modules from the parent contract; no dependency installs.
- Preserve parent contract bytes, original evidence and all R1 freeze files; create uniquely named R2 artifacts.
- Test overlays must be task-private and must not rewrite main or the QA baseline.
- The role that authors these tests cannot serve as their independent QA.

## Acceptance Commands
From candidate backend, with process-local GOROOT and GOTOOLCHAIN:
```text
go test -mod=readonly -tags unit ./internal/service -run '^(TestOpenAIGatewayService_SelectAccountForModelWithExclusions_StaleUsageSnapshotSkipsPause_Issue2994|TestOpenAISubscriptionGPT61SettlementMatrix)$' -count=1 -json
go test -mod=readonly ./internal/service -run '^(TestOpenAIGatewayService_SelectAccountForModelWithExclusions_StaleUsageSnapshotSkipsPause_Issue2994|TestOpenAISubscriptionGPT61SettlementMatrix)$' -count=1 -json
```
Rerun these commands with the private original-production overlay for RED. Execute the exact `existing-regression-regex.txt` selection under both tags with `-count=1 -json`, check declarations/leaf counts and no skip. Run scoped gofmt/diff-check, verify the original nine path hashes and inspect the exact two-test diff.

Controller/final QA retain parent acceptance requirements and known-failure attribution. Fresh main acceptance follows protected R2 integration; only this batch's ten changed source/test paths may differ from original bytes.

## Output
- Existing Developer report retains first line `### DONE: upstream-gap-fixes-20261103` and adds explicit R2 section, commands, exit codes, RED/GREEN and nine-path preservation.
- Independent addendum review: `docs/workflow/contract-reviews/upstream-gap-fixes-20261103-test-addendum-review.md`, first line `### PASS: upstream-gap-fixes-20261103-test-addendum` or FAIL/BLOCKED, including final `contract_sha256`.
- Final QA report: first line `### PASS: upstream-gap-fixes-20261103`, or FAIL/BLOCKED; binds parent/addendum/R2 freeze, independent probes, existing regression results and evidence hashes.
- Final controller review distinguishes local test/build acceptance from live DB/provider/browser acceptance. No commit, push or deployment.

## Stop Rules
No independent final-byte addendum PASS, any source/identity drift, an additional failing regression, lost original protection or required edit outside the two tests: stop, preserve evidence and return to controller. Do not broaden scope or weaken assertions.

## Budget
worker_mode: native-independent-agent; qa_worker_mode: independent-qa. Reuse original roles and isolated worktrees. No model substitution.
