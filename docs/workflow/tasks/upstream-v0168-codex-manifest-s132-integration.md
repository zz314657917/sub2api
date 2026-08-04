---
task_id: upstream-v0168-codex-manifest-s132-integration
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Port the API-key Codex models-manifest adapter from S132 without changing
ordinary `/v1/models`, OAuth manifest payloads, or the current Responses
subpath policy. The API-key-only normalization must preserve Codex web-search
capability for the affected model entries.

## Success Criteria

- Only a request authenticated with an API key assigned to an OpenAI group can
  reach the Codex manifest. Other platforms and unassigned keys fail closed.
- Ordinary `/v1/models` keeps its existing OpenAI-compatible list. The root
  `/models` SPA route is bypassed only for a GET Codex manifest probe carrying
  `client_version`.
- OAuth account manifests pass through unchanged. API-key custom-upstream
  manifests are normalized only at the exact affected Codex model property;
  unrelated models and fields remain unchanged.
- The current Responses subpath forwarding guard remains in place.

## Context

- Repo: `F:/mcplugins/sub2api`
- Integration worktree: `E:/codex-worktrees/sub2api-s132-codex-manifest-integration-20260804`
- Baseline: `main@4ba0ff75a`
- Source behavior: `3c1272d29 feat(codex): add API-key models manifest`
- The source route patch predates the current Responses subpath guard and must
  be adapted, not applied verbatim.

## Allowed Paths

- `backend/internal/handler/openai_codex_models_handler.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_codex_models_test.go`
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_codex_models_service_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/web/embed_on.go`
- `docs/workflow/tasks/upstream-v0168-codex-manifest-s132-integration.md`
- `docs/workflow/qa-reports/upstream-v0168-codex-manifest-s132-integration-qa.md`

## Denied Paths

- `backend/migrations/**`
- `backend/go.mod`
- `backend/go.sum`
- `frontend/**`
- `deploy/**`
- `Dockerfile*`
- `docker-compose*.yml`
- `outputs/**`
- `output/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `C:/Users/Administrator/.codex/memories/**`
- Any real OpenAI/ChatGPT/Codex request, credential, provider account,
  container, deployment, or production state.

## Constraints

- Preserve the current `guardResponsesSubpath` behavior while adding the
  manifest route selection.
- Use only in-process local HTTP test servers; no live external request is
  allowed.
- Do not weaken API-key authentication, group checks, owner boundaries, cache
  separation, or error mapping.
- Do not change the embedded frontend's normal `/models` page behavior.

## Acceptance Commands

Run from `backend`:

```powershell
gofmt -w <changed Go files>
go test ./internal/service -run 'CodexModels' -count=1
go test ./internal/server/routes -run 'CodexModels' -count=1
go test ./... -run '^$'
go build ./...
```

Run from the worktree root:

```powershell
git diff --check
git ls-files -u
git diff --name-only main...HEAD
```

## Output

- One independently reviewable integration commit with contract and evaluator
  QA report. The report must distinguish local fake-upstream evidence from a
  live Codex client or upstream verification.

## Stop Rules

- Stop if the route adaptation removes or weakens the Responses subpath guard,
  needs a public unauthenticated route, changes OAuth payloads, requires a new
  dependency, or needs a real external call.
- Stop before push, remote deletion, Docker operation, deployment, or a
  production write.

## Contract Review

`PASS / contract-approved`: the new handler is behind existing API-key
authentication and OpenAI-group checks; the service's existing local tests use
an in-process upstream. The only current-main conflict is a route addition
that predates `guardResponsesSubpath`; the approved adaptation inserts the
manifest selector alongside it and leaves all current guarded Responses routes
untouched. No migration, dependency, public unauthenticated endpoint, or live
external call is needed.
