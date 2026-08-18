# Upstream Gemini Typed Tool Config S232

## Task ID

`upstream-gemini-tool-config-s232`

## Role

Controller/Generator: Codex. Independent QA is performed from a separate
worktree before main integration.

## Goal

Behaviorally port upstream `3c3bb2fa1` and `1ba92449c` from the latest upstream
line (`upstream/main@49504adc9`, v0.1.178). Preserve the existing local
Antigravity typed transform and enable
`includeServerSideToolInvocations` only when Gemini built-in Google Search and
function declarations are mixed.

## Success Criteria

- `GeminiToolConfig` serializes the optional typed flag without changing
  function-only or web-search-only requests.
- `TransformClaudeToGeminiWithOptions` emits the flag for mixed built-in and
  function tools, preventing the upstream 400 behavior.
- Focused Antigravity tests, complete package regression, server compile,
  formatting, exact scope, provenance, conflict/index, and dirty worktree
  protection pass.

## Frozen Base

`main@91e7b4f820` (working tree user changes are protected and are not part of
this task).

## Allowed Paths

- `backend/internal/pkg/antigravity/gemini_types.go`
- `backend/internal/pkg/antigravity/request_transformer.go`
- `backend/internal/pkg/antigravity/request_transformer_test.go`
- `docs/workflow/results/upstream-gemini-tool-config-s232-result.md`

## Denied Paths

All other product files, including gateway/provider routing, schema/migrations,
frontend, Codex fingerprint files, dependency files, user-owned dirty files,
`knowledge/*`, `outputs/*`, deployment, containers, databases, provider
traffic, and remote refs.

## Constraints

- Do not merge `main..upstream/main`; history is divergent.
- Keep the flag absent for function-only and web-search-only transforms.
- Do not overwrite, stage, or commit the user's dirty/untracked files.
- No push, deployment, container, database, or real provider operation.

## Acceptance Commands

From `backend/`:

```powershell
go test ./internal/pkg/antigravity -run 'TestGeminiToolConfig_IncludeServerSideToolInvocations|TestTransformClaudeToGeminiWithOptions_PreservesWebSearchAlongsideFunctions' -count=10
go test ./internal/pkg/antigravity -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -l internal/pkg/antigravity/gemini_types.go internal/pkg/antigravity/request_transformer.go internal/pkg/antigravity/request_transformer_test.go
```

Also verify exact allowlist, clean index/conflict markers, upstream ancestry,
patch-id/provenance, and preservation of the frozen user dirty patch IDs and
untracked tutorial files.

## Output

One business implementation commit plus
`docs/workflow/results/upstream-gemini-tool-config-s232-result.md` with a first
line verdict (`### PASS`, `### FAIL`, or `### BLOCKED`), changed files,
commands/evidence, risks, and contract compliance.

## Stop Rules

- Stop if the implementation needs gateway, schema, frontend, fingerprint, or
  provider changes.
- Stop if any denied path changes or the user's dirty/untracked state changes.
- Stop if the focused behavior cannot be independently tested; report the
  blocker instead of claiming product failure.
