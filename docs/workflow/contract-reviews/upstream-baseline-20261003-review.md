---
task_id: upstream-baseline-20261003
status: approved
verdict: PASS
reviewer: independent-contract-review
base_commit: 9bb59ba3232b5f53fb20114a19424cc1eaadedb1
---

### PASS: upstream-baseline-20261003

# Contract Review

## Verdict

PASS

## Findings

- The clean main snapshot lacks `TokenCostRequest.GroupID` even though committed gateway code constructs requests with that field.
- The two existing shared-worktree hunks are the minimum implementation surface; a focused test file is permitted.
- Explicit GroupID must win over `Group.ID`; both absent must retain catalog fallback.
- If `Resolved` is absent, resolver-based GroupID-only requests require the one-line condition change to avoid bypassing channel pricing.
- Existing legacy long-context precedence must remain unchanged.
- Integration must compare exact target-file bytes and preserve all unrelated shared dirty files.

## Required Checks

- GroupID-only channel pricing amount assertions through existing Gateway/OpenAI paths.
- Explicit-ID precedence, Group.ID fallback, nil fallback, and legacy pricing tests.
- `go test -tags unit ./internal/service -run 'Test.*(TokenCost|GroupID|LegacyLongContext)' -count=1`
- `go build ./...`, `gofmt`, `git diff --check`, and exact-path audit.
