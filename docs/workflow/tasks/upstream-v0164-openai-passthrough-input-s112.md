---
task_id: upstream-v0164-openai-passthrough-input-s112
repo: F:/mcplugins/sub2api
phase: contract-draft
owner: codex
source: v0.1.164
---

## Task ID

`upstream-v0164-openai-passthrough-input-s112`

## Role

Planner/Generator by Codex; no worker delegation is needed for this narrow
two-file patch. The final review remains the Evaluator gate.

## Goal

Port upstream `851436c55` and `3e26dfa5b` into the local OpenAI OAuth
passthrough normalization path. Any non-array top-level `input` accepted by the
local passthrough request must be normalized to the list shape required by the
ChatGPT Codex endpoint before forwarding.

## Success Criteria

- A non-blank string `input` becomes one user message object while preserving
  the original string as message content.
- A whitespace-only string becomes an empty array, and a single JSON object is
  wrapped as a one-element array.
- An existing array remains unchanged; missing `input`, compact stream/store
  normalization, unsupported-field removal, and non-OAuth paths retain their
  current behavior.
- Focused normalization and OAuth passthrough forwarding tests pass, with no
  changes outside the allowed paths.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`
- Local implementation: `backend/internal/service/openai_gateway_service.go`
- Existing tests: `backend/internal/service/openai_passthrough_normalization_test.go`
  and `backend/internal/service/openai_oauth_passthrough_test.go`
- The normal OAuth transform already implements string input conversion in
  `backend/internal/service/openai_codex_transform.go`.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_passthrough_normalization_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-v0164-openai-passthrough-input-s112.md`

## Denied Paths

- `frontend/**`
- `backend/migrations/**`
- `backend/ent/**`
- `deploy/**`
- `backend/cmd/server/VERSION`
- `knowledge/**`
- `C:/Users/Administrator/.codex/memories/**`
- `frontend/src/views/admin/group-buy/**`
- The separate `E:/codex-worktrees/sub2api/group-buy-lifecycle-refund-hardening-s110` worktree
- Any path not listed under Allowed Paths

## Constraints

- Keep the patch local to `normalizeOpenAIPassthroughOAuthBody` and focused
  regression tests; do not refactor shared JSON helpers.
- Use the existing `gjson`/`sjson` behavior and local Go test conventions.
- Do not alter model routing, billing, scheduling, auth policy, stream/store
  semantics, or request handling outside the input-shape normalization.
- Do not stage, commit, push, deploy, refresh containers, or modify the
  parallel group-buy changes as part of this contract.

## Acceptance Commands

Run from `F:/mcplugins/sub2api/backend`:

```powershell
go test ./internal/service -run "TestNormalizeOpenAIPassthroughOAuthBody" -count=1
go test ./internal/service -run "TestOpenAIGatewayService_OAuthPassthrough_(StreamKeepsToolNameAndBodyNormalized|CompactUsesJSONAndKeepsNonStreaming)" -count=1
gofmt -d internal/service/openai_gateway_service.go internal/service/openai_passthrough_normalization_test.go
```

Run from `F:/mcplugins/sub2api`:

```powershell
git diff --check -- backend/internal/service/openai_gateway_service.go backend/internal/service/openai_passthrough_normalization_test.go
git diff --name-only -- backend/internal/service/openai_gateway_service.go backend/internal/service/openai_passthrough_normalization_test.go
```

The final path audit must show only the two business/test files plus explicitly
listed workflow evidence files.

## Output

- Changed files and focused test output in the final response.
- Evaluator conclusion in the final response as `PASS`, `FAIL`, or `BLOCKED`.
- No worker report is required because Codex owns this narrow implementation;
  if a worker is introduced later, it must use the standard worker-result
  template and stop rules.

## Stop Rules

- Stop before implementation if the local function topology or request contract
  requires changes outside Allowed Paths.
- Stop and return for review if a test exposes a change to routing, billing,
  auth policy, or compact stream/store semantics.
- Stop if any parallel group-buy or S110 path appears in the working diff.

## Budget

- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree: current `F:/mcplugins/sub2api`; no new worktree
