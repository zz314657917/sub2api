### DONE: upstream-main-auth-repo-s6c

## Verdict
DONE

## Candidate Commits
- `22ff1acde` fix(auth): 停用/删除分组后阻断 API Key
  - Result: SKIPPED_EMPTY
  - Evidence: `git cherry-pick -x 22ff1acde` produced conflicts only inside Allowed Paths; after preserving current branch logic and the candidate group-availability behavior, the cherry-pick became empty. The target branch already contains equivalent or newer behavior.
- `60f6602b8` fix: clear scheduler cache when deleting accounts
  - Result: SKIPPED_EMPTY
  - Evidence: `git cherry-pick -x 60f6602b8` was empty. Current `backend/internal/repository/account_repo.go` already calls `deleteSchedulerAccountSnapshot`, and `backend/internal/repository/account_repo_integration_test.go` already covers `TestDelete_RemovesSchedulerAccountSnapshot`.

## Deferred
- None

## Changed Files
- `docs/workflow/tasks/upstream-main-auth-repo-s6c.md`
- `docs/workflow/worker-results/upstream-main-auth-repo-s6c-result.md`
- `docs/workflow/qa-reports/upstream-main-auth-repo-s6c-qa.md`

## Acceptance Commands
- `git status --short --branch`: PASS, clean before report creation.
- `git diff --check`: PASS.
- `git diff --name-status f78566b5d..HEAD`: PASS, only `docs/workflow/tasks/upstream-main-auth-repo-s6c.md` before report creation.
- `git diff --name-only f78566b5d..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|backend/internal/pkg/apicompat/|backend/internal/service/openai_|backend/internal/service/gemini_|backend/internal/service/ratelimit_|deploy/|knowledge/|docs/workflow/main-log\\.md|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"`: PASS, no matches; `rg` exited 1 for no matches.
- `go test ./internal/server/middleware ./internal/repository ./internal/service -run "APIKeyAuth|AllowedGroups|Group|Delete|Scheduler|Cache|Account" -count=1`: PASS.

## Notes
- No denied paths were modified.
- No new API fields, migrations, frontend changes, or broad refactors were required.
