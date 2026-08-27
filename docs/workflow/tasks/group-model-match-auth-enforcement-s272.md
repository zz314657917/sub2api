# Group Model Match Auth Enforcement S272

## Task ID

`group-model-match-auth-enforcement-s272`

## Status

`approved`

## Role

Direct Codex implementation and evidence-first evaluation of a bounded backend
authorization repair. S271 remains independently owned and unchanged.

## Goal

Ensure every grouped API key reaches model-aware resolution after a gateway
handler parses the requested model, so the effective group's administrator-owned
`model_match_patterns` cannot be bypassed by the middleware fast path.

## Success Criteria

- A grouped API key without multi-group routes is rejected with HTTP 403 and
  `NO_MATCHING_GROUP_ROUTE` when its effective group does not match the
  requested model.
- The same single-group path remains allowed when the effective group matches
  the requested model.
- Existing multi-group and pinned-account model-aware resolution remains in the
  same service path and keeps its existing behavior.
- Nil middleware/service/key inputs and ungrouped keys keep their current safe
  behavior.
- Focused middleware tests pass repeatedly, affected service and middleware
  package regressions pass, the server compiles, and diff/index checks pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: current local `main`, preserving the concurrent S271 workflow and
  untracked `outputs/`.
- Root cause: `ResolveAPIKeyForModelRequest` returns success before invoking
  `APIKeyService.ResolveForModelRequest` when both `MultiGroupRoutes` and
  `PinnedAccountID` are empty. The service already enforces
  `Group.MatchesModel` for this single-group case.
- Prior contracts: S88 fail-closed routing and S91 group-owned model matching.

## Allowed Paths

- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s272_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/group-model-match-auth-enforcement-s272.md`
- `docs/workflow/worker-results/group-model-match-auth-enforcement-s272-result.md`
- `docs/workflow/qa-reports/group-model-match-auth-enforcement-s272-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- S271 worktree, contract, reports, implementation, status, or ownership.
- Database schema/migrations, repositories, group/API-key persistence, billing,
  pricing, subscriptions, quotas, account selection, provider adapters,
  frontend, dependencies, lockfiles, deployment, Docker, containers, shared
  data, `outputs/**`, staging, commit, push, and global memories.
- Group rule syntax, normalization, administrator write behavior, route
  priority/weight/cooldown, pinned-account fallback policy, and `/v1/models`
  display behavior.

## Constraints

- Reuse the existing `APIKeyService.ResolveForModelRequest` authorization path;
  do not duplicate model matching in middleware.
- Preserve existing nil guards. An ungrouped key may continue unchanged because
  it has no effective group rule to enforce.
- Return the existing stable 403 code and message on a mismatch.
- Do not broaden the repair into endpoint-specific parsing or routing changes.

## Acceptance Commands

```powershell
Push-Location backend
$found = @(go test ./internal/server/middleware -list '^TestS272')
@(
  'TestS272ResolveAPIKeyForModelRequestRejectsSingleGroupModelMismatch',
  'TestS272ResolveAPIKeyForModelRequestAllowsSingleGroupModelMatch'
) | ForEach-Object {
  if ($found -notcontains $_) { throw "Missing S272 test: $_" }
}
go test ./internal/server/middleware -run '^TestS272' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S272 focused middleware tests failed' }
go test ./internal/service ./internal/server/middleware -count=1
if ($LASTEXITCODE -ne 0) { throw 'S272 affected package regressions failed' }
go test ./cmd/server -run '^$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S272 server compile failed' }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S272 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S272 has unmerged index entries' }
```

The Evaluator additionally verifies the exact changed-path allowlist, confirms
the middleware no longer has a route-count authorization bypass, and checks
that concurrent S271 files and `outputs/` remain otherwise untouched.

## Output

- Worker result:
  `docs/workflow/worker-results/group-model-match-auth-enforcement-s272-result.md`
- QA report:
  `docs/workflow/qa-reports/group-model-match-auth-enforcement-s272-qa.md`
- Workflow status/log entries for contract review, implementation, QA, and
  final verdict.

## Stop Rules

- Stop if the repair requires persistence, schema, frontend, billing, provider,
  deployment, container, or shared-data changes.
- Stop if matching single-group requests are rejected, existing multi-group
  tests regress, or pinned-account no-fallback behavior changes.
- Stop after two failed implementation attempts and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_mode: `fresh focused and package-level Go verification plus independent diff review`
- worktree_root: `current primary checkout with protected user changes`

### PASS: group-model-match-auth-enforcement-s272 contract

The source evidence identifies one middleware fast-path bypass while the
existing service already owns the required model authorization. The two-file
product allowlist, stable error contract, matching-request control, executable
tests, and explicit dirty-worktree protections are decision-complete.
