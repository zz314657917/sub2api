# Task Contract

## Task ID
upstream-main-apicompat-s6a

## Role
Developer worker for apicompat and Codex transform compatibility fixes.

## Goal
Port the approved upstream apicompat/Codex transform fixes onto branch `codex/upstream-main-apicompat-s6a` from baseline `main@f78566b5d`, without merging `upstream/main` and without touching unrelated product areas.

## Candidate Commits
- `348a48773` fix(codex-transform): preserve underscore when rewriting `call_*` tool-call ids.
- `a729752de` test: align codex tool-call id assertions with `fc_` prefix.
- `e9a25e7b9` fix(apicompat): preserve empty streaming thinking blocks.
- `df82a3bc6` fix(openai): avoid null content when converting chat-completions to responses.
- `276b5c775` fix(apicompat): strip temperature/top_p for reasoning models in Responses conversion.
- `c4d7edba0` fix(apicompat): map developer role to system.

## Allowed Paths
- `backend/internal/pkg/apicompat/**`
- `backend/internal/service/openai_codex_transform*`
- `docs/workflow/tasks/upstream-main-apicompat-s6a.md`
- `docs/workflow/worker-results/upstream-main-apicompat-s6a-result.md`
- `docs/workflow/qa-reports/upstream-main-apicompat-s6a-qa.md`

## Denied Paths
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
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
- Do not modify files owned by S6B/S6C.
- You are not alone in the codebase; do not revert other branches or unrelated existing changes.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
git diff --name-status f78566b5d..HEAD
git diff --name-only f78566b5d..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|deploy/|knowledge/|docs/workflow/main-log\\.md|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"
go test ./internal/pkg/apicompat -run "ChatCompletions|Responses|Anthropic|Thinking|Reasoning|Developer|Tool|Temperature|TopP" -count=1
go test ./internal/service -run "CodexTransform|OpenAICodex" -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-apicompat-s6a-result.md` with first line `### DONE: upstream-main-apicompat-s6a` or `### BLOCKED: upstream-main-apicompat-s6a`.
- Write `docs/workflow/qa-reports/upstream-main-apicompat-s6a-qa.md` with first line `### PASS: upstream-main-apicompat-s6a`, `### FAIL: upstream-main-apicompat-s6a`, or `### BLOCKED: upstream-main-apicompat-s6a`.
