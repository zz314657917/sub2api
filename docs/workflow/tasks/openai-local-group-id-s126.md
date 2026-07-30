---
task_id: openai-local-group-id-s126
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Prevent the sub2api-local top-level `group_id` request field from reaching
upstream OpenAI-compatible APIs, where strict `/responses`,
`/chat/completions`, and `/images/*` implementations reject it with HTTP 400.

## Success Criteria

- Native and passthrough Responses forwarding removes only the top-level
  `group_id` field before the upstream request is built.
- Chat Completions forwarding, including raw third-party OpenAI-compatible
  passthrough, removes the top-level `group_id` field.
- Images JSON and multipart forwarding removes the top-level/form `group_id`
  field while preserving model rewriting, prompts, image files, and all other
  supported fields.
- Nested application data named `group_id` is not recursively modified.
- Focused regressions prove the upstream request body no longer contains the
  local-only field, and the backend still compiles.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`
- The API key and its resolved group remain the routing authority. A request
  body `group_id` is not accepted as a routing override.
- The primary worktree contains unrelated user changes; preserve them.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_gateway_service_hotpath_test.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/service/openai_images_test.go`
- `docs/workflow/**`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- `Dockerfile*`
- `docker-compose*.yml`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`

## Constraints

- Strip only the exact top-level/form field `group_id`; do not recursively
  delete nested data and do not broaden this Sprint into generic schema
  filtering.
- Do not use request-body `group_id` to select or override an API-key group.
- Preserve existing invalid-JSON handling and multipart file content.
- Do not deploy, update containers, push, create a commit, reset, clean, or
  revert any existing user changes.

## Acceptance Commands

```powershell
cd backend
go test ./internal/service -run "GroupID|Forward_TextResponsesSetsBillingModel|ForwardAsRawChatCompletions_ForcesStreamUsage|TransparentBackgroundAlias" -count=1
go test ./... -run "^$"
cd ..
git diff --check
```

## Output

- A focused implementation and
  `docs/workflow/qa-reports/openai-local-group-id-s126-qa.md`.
- No worker is invoked because the user did not authorize sub-agent work;
  Codex performs implementation, QA, and final evaluation directly.

## Stop Rules

- Stop if the fix requires changing group-selection semantics, authentication,
  schema/migrations, deployment, or container configuration.
- Stop if multipart rewriting cannot preserve uploaded file bytes and headers.
- Stop if an existing user modification overlaps a required source change and
  cannot be preserved.
