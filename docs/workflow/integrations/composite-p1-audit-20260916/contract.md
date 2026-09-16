---
status: approved
review_verdict: PASS
task_id: composite-p1-audit
worker_model: gpt-5.6-terra
base_commit: 8497ecec71985d4903678c056639be2c7a46c93f
spec_ref: docs/workflow/composite-p1-audit/original-contract.md
---

## Task ID
composite-p1-audit

## Role
Fresh independent Terra QA; no Generator or business implementation.

## Goal
Complete the missing independent local QA of already committed Composite P1
(b25cc223a), against current clean base. Keep all original P1 requirements and
denied boundaries. This audit does not approve P2/P3, runtime migration or release.

## Success Criteria
1. Map each original-contract Success Criterion to source/test evidence and
   executed results. Inspect schema/migration without running a database, route
   CRUD validation and group ownership, soft-delete transaction, ordered enabled
   preview, unknown input behavior, and existing admin auth/audit registration.
2. Execute the original focused service/handler/routes/server/build gates; list
   actual selected tests. Distinguish static route registration from real auth
   and database proof. Missing assertions remain gaps, not implicit PASS.
3. Check generator reproducibility in this task-only worktree after source/test
   review and pre-generation build. Run go generate ./ent, inventory exact diff,
   and stop generator work if output touches paths outside the original allowlist.
   Otherwise run go generate ./cmd/server and inventory again. No hand edits to
   generated code, no baseline repair, no application of generated output to main.
   Report failures/drift with exact paths and commands. If applicable rebuild the
   generated allowed result, but distinguish pre/post-generation evidence.
4. Report PASS only if all original local gates are genuinely satisfied. Missing
   generators, forbidden drift or unverified gates -> BLOCKED, product defects ->
   FAIL. Original no-database/provider restriction remains intact, and broader
   roadmap runtime acceptance stays open regardless of local audit result.

## Allowed Paths
- docs/workflow/qa-reports/composite-p1-audit-qa.md (QA write)
- Original contract generated-file allowlist, only as automatic generator output
  in this disposable audit worktree. No manual source modifications.

## Denied Paths
Manual business/test changes, all other worktrees including main, database/Redis,
provider requests, migration execution, frontend, container/service startup,
deployment, commits/push, Git resets/clean, auth changes, P2/P3 resolver/dispatch.

## Constraints
Read original-contract.md, original-review.md and previous-qa.md under
docs/workflow/composite-p1-audit. Preserve old findings as history; revalidate
current state, do not rely on old healthy/build claims. Do not overwrite main's
shared workflow/current-task. Evidence-only report written with apply_patch.

## Acceptance Commands
backend:
`go test ./internal/service -list CompositeRoute`
`go test ./internal/handler/admin -list CompositeRoute`
`go test ./internal/server/routes -list CompositeRoute`
`go test ./internal/service -run CompositeRoute -count=1`
`go test ./internal/handler/admin -run CompositeRoute -count=1`
`go test ./internal/server/routes -run CompositeRoute -count=1`
`go test ./cmd/server -run '^$' -count=1`
`go build ./...`
Then generator steps as Success Criterion 3. Use gofmt -l (not -w) on original
Go allowlist; root git diff --check, git ls-files -u, git status --short and
tracked/untracked inventory before/after. No empty selector counted as PASS.

## Output
docs/workflow/qa-reports/composite-p1-audit-qa.md starts
### PASS/FAIL/BLOCKED: composite-p1-audit, including each criterion, failures,
executed commands, generation drift and explicit remaining runtime gates.

## Stop Rules
No QA dispatch until independent contract review PASS. Stop mutations on denied
generator output; preserve it for inspection, do not revert/reset or repair it.
Terra unavailable -> BLOCKED, no silent alternate model.
