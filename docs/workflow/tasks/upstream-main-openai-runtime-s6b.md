# Task Contract

## Task ID
upstream-main-openai-runtime-s6b

## Role
Developer worker for OpenAI/Gemini runtime compatibility fixes.

## Goal
Port the approved upstream OpenAI/Gemini runtime fixes onto branch `codex/upstream-main-openai-runtime-s6b` from baseline `main@f78566b5d`, without merging `upstream/main` and without touching unrelated product areas.

## Candidate Commits
- `679c0865a` fix(openai): handle versioned compatible base URLs.
- `a61174291` fix(gateway): detach upstream context unconditionally for image generation.
- `2c14efeaa` fix(openai-images): 修复图片生成 `n` 参数透传.
- `87fac3045` fix: use tier cooldown for google one gemini 429.
- `bec1e2b69` fix(openai): 永久禁用缺失 refresh_token 的 OAuth 账号.

## Allowed Paths
- `backend/internal/service/openai_*`
- `backend/internal/service/gemini_*`
- `backend/internal/service/ratelimit_*`
- `docs/workflow/tasks/upstream-main-openai-runtime-s6b.md`
- `docs/workflow/worker-results/upstream-main-openai-runtime-s6b-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-runtime-s6b-qa.md`

## Denied Paths
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/pkg/apicompat/**`
- `backend/internal/repository/**`
- `backend/internal/server/**`
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
- Do not modify files owned by S6A/S6C.
- You are not alone in the codebase; do not revert other branches or unrelated existing changes.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
git diff --name-status f78566b5d..HEAD
git diff --name-only f78566b5d..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|backend/internal/pkg/apicompat/|backend/internal/repository/|backend/internal/server/|deploy/|knowledge/|docs/workflow/main-log\\.md|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"
go test ./internal/service -run "OpenAI|Images|Gateway|Versioned|BaseURL|Gemini|RefreshToken|RateLimit" -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-openai-runtime-s6b-result.md` with first line `### DONE: upstream-main-openai-runtime-s6b` or `### BLOCKED: upstream-main-openai-runtime-s6b`.
- Write `docs/workflow/qa-reports/upstream-main-openai-runtime-s6b-qa.md` with first line `### PASS: upstream-main-openai-runtime-s6b`, `### FAIL: upstream-main-openai-runtime-s6b`, or `### BLOCKED: upstream-main-openai-runtime-s6b`.
