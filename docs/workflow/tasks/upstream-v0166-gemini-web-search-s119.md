---
task_id: upstream-v0166-gemini-web-search-s119
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Adapt upstream `3e0810611` so a normal Chat Completions function named
`web_search` remains a client-side function declaration when forwarded to
Gemini. Only explicitly typed server-side search tools may become Gemini
built-in Google Search.

## Success Criteria

- A request with `type: function` tools named `web_search` and `read_file`
  forwards both as Gemini `functionDeclarations` and has no `googleSearch`
  tool.
- Explicit `web_search*` and `google_search` tool types retain the existing
  Gemini built-in search conversion.
- Tool routing, response conversion, account selection, retry behavior,
  persistence, frontend, deployment, and container behavior remain unchanged.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-v0166-gemini-web-search-s119`
- Related upstream commit: `3e0810611`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`
- Related files:
  `backend/internal/service/gemini_messages_compat_service.go` and
  `backend/internal/service/gemini_messages_compat_service_test.go`

## Allowed Paths

- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service_test.go`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-v0166-gemini-web-search-s119.md`
- `docs/workflow/worker-results/upstream-v0166-gemini-web-search-s119-result.md`
- `docs/workflow/qa-reports/upstream-v0166-gemini-web-search-s119-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/server/**`
- `backend/internal/service/**` except the two listed Gemini files
- `frontend/**`
- `deploy/**`
- `Dockerfile*`
- `knowledge/**`
- `outputs/**`

## Constraints

- Reuse the existing Gemini tool-conversion path and make the smallest change
  to its server-side search discriminator.
- Do not promote a function based on its `name`; only the explicit tool type
  may identify a Gemini built-in search tool.
- Preserve current explicit search behavior and do not change tool-call
  response conversion.
- Work only in this isolated worktree; do not use or clean the dirty primary
  worktree.

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-v0166-gemini-web-search-s119/backend
go test ./internal/service -run "^TestGeminiForwardAsChatCompletions_FunctionNamedWebSearchStaysClientSide$" -count=1
go test ./internal/service -run "TestGemini" -count=1
go test ./... -run "^$"
gofmt -d internal/service/gemini_messages_compat_service.go internal/service/gemini_messages_compat_service_test.go
cd E:/codex-worktrees/sub2api/upstream-v0166-gemini-web-search-s119
git diff --check
git diff --name-only HEAD
```

## Output

- Narrow Gemini tool-classification adaptation, focused regression evidence,
  Generator result, QA report, and an allowlist-constrained diff.

## Stop Rules

- Stop if preserving the client-side function requires a Gemini request-schema,
  routing, tool-response, scheduler, persistence, or frontend change.
- Stop if explicit server-side search types cannot keep their existing built-in
  Gemini behavior.
