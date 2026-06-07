# Task Contract

## Task ID
upstream-main-auth-repo-s6c

## Role
Developer worker for API key auth and repository hardening fixes.

## Goal
Port the approved upstream auth/repository hardening fixes onto branch `codex/upstream-main-auth-repo-s6c` from baseline `main@f78566b5d`, without merging `upstream/main` and without touching unrelated product areas.

## Candidate Commits
- `22ff1acde` fix(auth): 停用/删除分组后阻断 API Key.
- `60f6602b8` fix: clear scheduler cache when deleting accounts.

## Allowed Paths
- `backend/internal/repository/**`
- `backend/internal/server/middleware/**`
- `backend/internal/service/api_key_auth_cache_impl.go`
- `backend/internal/service/admin_service_delete_test.go`
- `docs/workflow/tasks/upstream-main-auth-repo-s6c.md`
- `docs/workflow/worker-results/upstream-main-auth-repo-s6c-result.md`
- `docs/workflow/qa-reports/upstream-main-auth-repo-s6c-qa.md`

## Denied Paths
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/pkg/apicompat/**`
- `backend/internal/service/openai_*`
- `backend/internal/service/gemini_*`
- `backend/internal/service/ratelimit_*`
- `deploy/**`
- `knowledge/**`
- `docs/workflow/main-log.md`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `.github/**`
- `assets/**`
- `README*`

## Constraints
- Prefer `git cherry-pick -x`.
- If a candidate requires denied paths, new API fields, schema/migration, frontend changes, or broad gateway refactor, stop that candidate and record `DEFERRED`.
- Do not modify files owned by S6A/S6B.
- You are not alone in the codebase; do not revert other branches or unrelated existing changes.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
git diff --name-status f78566b5d..HEAD
git diff --name-only f78566b5d..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|backend/internal/pkg/apicompat/|backend/internal/service/openai_|backend/internal/service/gemini_|backend/internal/service/ratelimit_|deploy/|knowledge/|docs/workflow/main-log\\.md|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"
go test ./internal/server/middleware ./internal/repository ./internal/service -run "APIKeyAuth|AllowedGroups|Group|Delete|Scheduler|Cache|Account" -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-auth-repo-s6c-result.md` with first line `### DONE: upstream-main-auth-repo-s6c` or `### BLOCKED: upstream-main-auth-repo-s6c`.
- Write `docs/workflow/qa-reports/upstream-main-auth-repo-s6c-qa.md` with first line `### PASS: upstream-main-auth-repo-s6c`, `### FAIL: upstream-main-auth-repo-s6c`, or `### BLOCKED: upstream-main-auth-repo-s6c`.
