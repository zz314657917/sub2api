# Task Contract: upstream-anthropic-grok-usage-s66b

## Task ID

`upstream-anthropic-grok-usage-s66b`

## Status

`approved`

## Role

You are the Generator worker for the isolated usage-compatibility lane. Implement only this contract.

## Goal

Adapt upstream Responses/Anthropic cache-creation token propagation and Grok Responses reasoning-effort preservation while leaving billing policy and gateway routing unchanged.

## Success Criteria

- Anthropic -> Responses non-stream conversion maps `cache_creation_input_tokens` into the local Responses usage representation.
- Responses -> Anthropic non-stream and streaming conversions preserve cache-creation token usage.
- Grok Responses usage metadata accepts both `reasoning.effort` and compatible flat `reasoning_effort` through the shared extractor.
- Existing cache-read/input/output token mapping and Grok request payload behavior remain unchanged.
- Focused apicompat and Grok tests pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Upstream references: `0d28f7f90`, `83f169e4f`, `0fa1eb85e`, `5a0dd510e`.
- `0d28f7f90` dry-applies on the current tree; review it rather than assuming the follow-up also applies unchanged.

## Allowed Paths

- `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic.go`
- `backend/internal/pkg/apicompat/responses_anthropic_cache_creation_test.go`
- `backend/internal/pkg/apicompat/anthropic_responses_test.go`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_gateway_grok_test.go`
- `docs/workflow/worker-results/upstream-anthropic-grok-usage-s66b-result.md`

## Denied Paths

- All paths not listed above.
- Core billing/pricing services, migrations, frontend, routing, deployment, and production configuration.
- `knowledge/**` and global memories.

## Constraints

- Preserve local usage field names and JSON compatibility.
- Do not change pricing, charging formulas, model mappings, or retry behavior.
- Remove a helper only if the patch makes it unused and it is inside an Allowed Path.
- Do not modify another worker's files.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/pkg/apicompat -run "Test.*CacheCreation|Test.*Anthropic.*Usage|Test.*Responses.*Anthropic" -count=1
go test ./internal/service -run "Test.*Grok.*ReasoningEffort|TestForwardGrokResponsesStreaming" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-anthropic-grok-usage-s66b-result.md`.
- First line must be `### DONE: upstream-anthropic-grok-usage-s66b`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit the implementation and report on the assigned worktree branch.

## Stop Rules

- Stop if correct propagation requires changing billing, repository schema, gateway routing, or any path outside Allowed Paths.
- Stop on ambiguous token semantics instead of inventing a new field.
- Do not revert other worktree or user changes.
