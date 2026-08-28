# Group Model Match Legacy Compatibility S273

## Task ID

`group-model-match-legacy-compat-s273`

## Status

`approved`

## Role

Direct Codex implementation and evidence-first evaluation of a bounded
backend routing compatibility repair following the S272 regression review.

## Goal

Prevent the S272 model-aware middleware handoff from rejecting legacy single-
group API keys whose group predates `model_match_patterns` and still has an
empty rule set, while retaining strict administrator-owned matching for
configured groups and existing fail-closed behavior for multi-group and pinned
account routes.

## Success Criteria

- A non-pinned single-group API key with an empty group rule set remains
  routable for model-aware requests, preserving pre-S272 behavior.
- A legacy single-group image request for a group that has image generation
  disabled reaches the handler's stable `permission_error` gate instead of
  being converted into a routing mismatch.
- A single-group key keeps its configured model match even when the generic
  endpoint platform list does not name the group's provider (for example a
  Grok key on `/v1/responses`).
- A configured single-group rule still rejects mismatches with
  `NO_MATCHING_GROUP_ROUTE` and accepts matches.
- Multi-group routes and pinned-account routes with empty rules retain their
  existing fail-closed behavior; `Group.MatchesModel` semantics do not change.
- Regression tests cover the legacy empty-rule case and the S272 mismatch/match
  controls.
- Focused routing/middleware tests, affected package tests, server compile,
  formatting, diff and unmerged-index checks pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: current local `main` at `bf493099e`, preserving untracked
  `outputs/` and all unrelated worktrees.
- S272 removed the middleware route-count fast path. The existing service now
  calls `canFallbackToDefaultGroup`, which invokes `Group.MatchesModel`; that
  method intentionally returns false for an empty rule set.
- Migration 191 gives the column a `[]` default for legacy rows. Persistence
  and schema changes are outside this repair; compatibility belongs at the
  single-group fallback boundary.

## Allowed Paths

- `backend/internal/service/api_key_routing.go`
- `backend/internal/service/api_key_routing_s273_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s272_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/group-model-match-legacy-compat-s273.md`
- `docs/workflow/worker-results/group-model-match-legacy-compat-s273-result.md`
- `docs/workflow/qa-reports/group-model-match-legacy-compat-s273-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Database schema or migrations, repositories, group/API-key persistence,
  billing, pricing, subscriptions, quotas, account selection, provider
  adapters, frontend, dependencies, lockfiles, deployment, Docker,
  containers, shared data, `outputs/**`, staging, commit, push and global
  memories.
- `Group.MatchesModel` implementation, normalization/validation, route
  priority/weight/cooldown, pinned-account fallback policy and `/v1/models`
  display behavior.
- S271 implementation, contract, reports, status or ownership.

## Constraints

- Keep the existing service-owned matching path; do not duplicate model
  matching in middleware.
- Preserve the direct-handler compatibility boundary for an incomplete,
  single-group, non-pinned snapshot only; if multi-group routes or a pinned
  account are present, middleware must still enter service routing.
- Treat only an empty rule set on the unpinned single-group default fallback as
  legacy compatibility. Do not make empty rules a general wildcard.
- Keep image permission ownership in the gateway handler for the legacy
  single-group default route; do not turn it into a routing wildcard.
- Preserve nil guards, configured mismatch rejection and all multi-group or
  pinned-account semantics.
- Do not perform database writes or automatic migration/backfill.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run 'TestS273|TestS91' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S273 focused service tests failed' }
go test ./internal/server/middleware -run '^TestS272' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S272 middleware controls failed' }
go test ./internal/service ./internal/server/middleware -count=1
if ($LASTEXITCODE -ne 0) { throw 'S273 affected package regressions failed' }
go test ./cmd/server -run '^$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S273 server compile failed' }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S273 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S273 has unmerged index entries' }
```

The Evaluator additionally verifies the exact changed-path allowlist,
confirms that `Group.MatchesModel` remains unchanged, and confirms that
`outputs/` and unrelated worktrees remain untouched.

## Output

- Worker result:
  `docs/workflow/worker-results/group-model-match-legacy-compat-s273-result.md`
- QA report:
  `docs/workflow/qa-reports/group-model-match-legacy-compat-s273-qa.md`
- Workflow status/log entries for contract review, implementation, QA and
  final verdict.

## Stop Rules

- Stop if the repair requires schema, migration, persistence, provider,
  deployment, container or shared-data changes.
- Stop if configured mismatches are accepted, multi-group/pinned behavior
  changes, or empty rules become a wildcard outside the legacy single-group
  fallback.
- Stop after two failed implementation attempts and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_mode: `fresh focused and package-level Go verification plus independent
  diff review`
- worktree_root: `current primary checkout with protected user changes`

### PASS: group-model-match-legacy-compat-s273 contract

The reviewed S272 regression is isolated to the single-group fallback boundary.
The compatibility exception is narrow, preserves strict configured matching,
avoids persistence changes, and has executable regression and protection gates.
