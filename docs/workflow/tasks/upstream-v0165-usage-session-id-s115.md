---
task_id: upstream-v0165-usage-session-id-s115
status: done
owner: Codex
qa_mode: runtime
---

# Task Contract

## Goal

Persist an explicit client session identifier on usage records so operators
can correlate retries and conversations without deriving identity from prompt
content or `prompt_cache_key`.

## Success Criteria

- Accept `session_id`, `X-Session-Id`, and `X-Conversation-ID` request headers
  with deterministic precedence and bounded/trimmed values.
- Propagate the identifier through gateway, OpenAI-compatible, Gemini, Grok,
  embedding, image, and response usage write paths.
- Persist and scan the value for both single-row and batch usage inserts;
  existing rows and omitted headers remain `NULL`/empty-compatible.
- Expose the field through usage DTO mapping where the local API already
  exposes usage metadata, without changing billing or routing semantics.
- Add focused service/repository tests and a forward-only migration using the
  next unused local migration number.

## Scope Boundary

- Adapt upstream `1c0cb24c7` to the local repository topology.
- Do not port upstream batch-image propagation when the local batch-image
  tables/services are absent; document that omission in the QA report.
- Do not derive identifiers from prompt contents, cache keys, API-key IDs, or
  request hashes.
- Do not modify account selection, billing calculations, routing, deployment,
  containers, or unrelated dirty files.

## Allowed Paths

- backend/internal/handler/dto/types.go
- backend/internal/handler/dto/mappers.go
- backend/internal/handler/gateway_handler.go
- backend/internal/handler/gateway_handler_chat_completions.go
- backend/internal/handler/gateway_handler_responses.go
- backend/internal/handler/gemini_v1beta_handler.go
- backend/internal/handler/grok_media.go
- backend/internal/handler/openai_alpha_search.go
- backend/internal/handler/openai_chat_completions.go
- backend/internal/handler/openai_embeddings.go
- backend/internal/handler/openai_gateway_handler.go
- backend/internal/handler/openai_images.go
- backend/internal/repository/usage_log_repo.go
- backend/internal/repository/usage_log_repo_insert.go
- backend/internal/repository/usage_log_repo_query.go
- backend/internal/service/session_id.go
- backend/internal/service/session_id_test.go
- backend/internal/service/usage_log.go
- backend/internal/service/gateway_usage_billing.go
- backend/internal/service/openai_gateway_usage.go
- backend/internal/service/openai_gateway_scheduling.go
- backend/internal/service/openai_gateway_grok_cache.go
- backend/internal/repository/usage_log_session_id_unit_test.go
- backend/internal/repository/usage_log_session_id_integration_test.go
- backend/migrations/195_add_usage_log_session_id.sql
- docs/workflow/tasks/upstream-v0165-usage-session-id-s115.md

## Denied Paths

- backend/ent/**
- backend/internal/service/openai_live*.go
- backend/internal/handler/openai_live.go
- backend/internal/repository/concurrency_cache.go
- backend/internal/repository/gateway_cache.go
- frontend/**
- deploy/**
- Dockerfile*
- knowledge/**
- outputs/**

## Constraints

- Preserve existing usage insert ordering and nullable semantics.
- Header extraction must be independent of body parsing and safe for retries.
- Keep the local usage repository's explicit column/type/scan lists in sync.
- Do not overwrite unrelated user changes in the working tree.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service ./internal/repository ./internal/handler/... -run "Test(SessionID|UsageLog.*Session|.*SessionID)" -count=1
gofmt -d internal/service/session_id.go internal/service/session_id_test.go internal/repository/usage_log_repo*.go
cd F:/mcplugins/sub2api
git diff --check
```

## Output

- Source changes, focused tests, migration, and a QA report with PASS/FAIL,
  executed commands, and the batch-image applicability decision.

## Stop Rules

- Stop and split the work if the implementation requires Ent regeneration,
  billing schema changes, or frontend API redesign.
- Stop if a requested identifier would be synthesized rather than supplied by
  the client.
