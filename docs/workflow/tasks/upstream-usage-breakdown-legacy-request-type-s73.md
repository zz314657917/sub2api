# Task Contract: upstream-usage-breakdown-legacy-request-type-s73

## Task ID

`upstream-usage-breakdown-legacy-request-type-s73`

## Status

`draft`

## Role

Generator adapts only the aliased legacy request-type fallback used by `GetUserBreakdownStats` from upstream `de28eba3c`.

## Goal

Make UserBreakdown request-type filters include legacy rows whose canonical `request_type` is unknown but stream/WS flags identify Sync, Stream, or WS v2.

## Success Criteria

- Existing `buildRequestTypeFilterCondition` delegates to a new alias-aware helper with an empty alias and preserves its current SQL.
- `GetUserBreakdownStats` uses the helper with alias `ul`.
- Sync fallback includes only `request_type=0, stream=false, openai_ws_mode=false`.
- Stream fallback includes only `request_type=0, stream=true, openai_ws_mode=false`.
- WS v2 fallback includes only `request_type=0, openai_ws_mode=true`.
- Explicit nonzero request types remain authoritative; legacy flags do not reclassify them.
- Current RequestType+Stream extra-AND behavior, seven-column scan, actual-cost ordering, limits, dimensions, and S70 leaderboard exclusion behavior remain unchanged.

## Allowed Paths

- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `docs/workflow/worker-results/upstream-usage-breakdown-legacy-request-type-s73-result.md`

## Denied Paths

- `usage_log_repo_breakdown_test.go`, Cyber migration, frontend, service/handler, leaderboard queries, Ent/Wire, payment/subscription, deployment, `knowledge/**`, and global memories.

## Constraints

- Do not copy upstream `usage_log_repo_trend.go`; local query lives in `usage_log_repo.go`.
- Use only untagged tests. Do not repair unrelated unit-tag drift.
- Ordinary breakdown continues to include leaderboard-excluded users.

## Acceptance Commands

```powershell
Push-Location backend
$pattern = "^(TestBuildRequestTypeFilterConditionWithAliasLegacyFallbackMatrix|TestGetUserBreakdownStatsRequestTypeLegacyFallbackMatrix)$"
$listed = @(go test ./internal/repository -list $pattern | Where-Object { $_ -match '^Test' })
if ($LASTEXITCODE -ne 0 -or $listed.Count -ne 2) { throw "S73 required test discovery failed: $($listed -join ', ')" }
go test ./internal/repository -run $pattern -count=1
if ($LASTEXITCODE -ne 0) { throw "S73 required tests failed" }
go test ./internal/repository -run "^(TestBuildRequestTypeFilterConditionLegacyFallback|TestBuildRequestTypeFilterConditionWithAliasLegacyFallbackMatrix|TestGetUserBreakdownStatsRequestTypeLegacyFallbackMatrix)$" -count=1
if ($LASTEXITCODE -ne 0) { throw "S73 request-type regressions failed" }
go test ./internal/repository -run "^$" -count=1
if ($LASTEXITCODE -ne 0) { throw "S73 repository compile failed" }
Pop-Location
$dirty = @(git status --porcelain --untracked-files=all)
if ($dirty.Count -gt 0) { throw "S73 acceptance requires a clean committed worktree: $($dirty -join ', ')" }
$base = (git merge-base HEAD codex/upstream-v0151-followups-s71-s73).Trim()
$allowed = @(
  "backend/internal/repository/usage_log_repo.go", "backend/internal/repository/usage_log_repo_request_type_test.go",
  "docs/workflow/worker-results/upstream-usage-breakdown-legacy-request-type-s73-result.md"
)
$unexpected = @(git diff --name-only "$base..HEAD" | Where-Object { $_ -notin $allowed })
if ($unexpected.Count -gt 0) { throw "S73 path audit failed: $($unexpected -join ', ')" }
git diff --check "$base..HEAD"
if ($LASTEXITCODE -ne 0) { throw "S73 diff check failed" }
```

## Output

- Write `docs/workflow/worker-results/upstream-usage-breakdown-legacy-request-type-s73-result.md` with the required verdict first line.
- Commit all contract changes and return the commit hash.

## Stop Rules

- Stop if another repository file, migration, frontend, handler, or service change is required.
- Stop if explicit request types can be reclassified by legacy flags.
- Stop if S70 leaderboard queries or ordinary user inclusion change.
