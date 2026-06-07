### DONE: upstream-main-usage-cache-stats-s10

## Summary

- Created isolated worktree `E:/codex-worktrees/sub2api/upstream-main-usage-cache-stats-s10` and branch `codex/upstream-main-usage-cache-stats-s10` from baseline `c6fefc8c6`.
- Added Sprint contract at `docs/workflow/tasks/upstream-main-usage-cache-stats-s10.md`.
- Ported both approved usage cache stats candidates without directly merging `upstream/main`.
- Kept the implementation inside the approved backend and workflow paths.

## Candidate Results

- `029b6d61a`: `CHERRY_PICKED` as `596a02344`. Adds `total_cache_creation_tokens` and `total_cache_read_tokens` to usage aggregate stats and maps the split values through repository/service layers.
- `7386f38cf`: `CHERRY_PICKED` as `3f9dce82e`. Updates the API contract test expectation for `/api/v1/usage/stats`.

## Deferred / Skipped

- `0760cda92`: `DEFERRED`. Frontend i18n only; frontend paths are denied in this Sprint.
- `9ecfc4e92` / `cb4f0015f`: `DEFERRED`. Adds Codex admin skill assets; not part of this backend usage stats Sprint.
- `8ec448a8f` / `f868f7cb4`: `SKIPPED`. Merge commits; this Sprint does not merge `upstream/main`.

## Commits

- `4b3bca723` docs: add usage cache stats s10 contract
- `596a02344` feat(usage): 聚合统计拆分缓存创建与命中 token
- `3f9dce82e` test(usage): API契约测试补充缓存创建/命中token字段

## Changed Files

- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/service/usage_service.go`
- `backend/internal/server/api_contract_test.go`
- `docs/workflow/tasks/upstream-main-usage-cache-stats-s10.md`
- `docs/workflow/worker-results/upstream-main-usage-cache-stats-s10-result.md`
- `docs/workflow/qa-reports/upstream-main-usage-cache-stats-s10-qa.md`
- `docs/workflow/main-log.md`

## Verification

- `git status --short --branch`
- `git diff --check`
- denied path audit against `main...HEAD`
- `go test ./internal/service ./internal/repository ./internal/server -run "Usage|Stats|Cache|Contract" -count=1`
- `go test ./internal/service ./internal/repository ./internal/server -count=1`

## Notes

- No migration, Ent schema, frontend, skills, deploy, or knowledge paths were modified.
- `total_cache_tokens` remains present; the new fields split the same cache total into creation/read components.
