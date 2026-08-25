# Task Contract

## Task ID

upstream-gemini-tool-schema-s254

## Role

Developer Worker (`gpt-5.6-terra`) implements only the approved behavior in the
isolated worktree. Codex Controller reviews the diff and an independent QA
Worker (`gpt-5.6-terra`) reruns the contract gates in a separate worktree.

## Goal

Behavior-level port of upstream `19da0f240` from `upstream/main@e2d9b823f`.
Gemini Messages tool schemas must not retain unsupported `deprecated` or
`exclusiveMinimum` fields, and enum values must be representable as Gemini
strings instead of forwarding unsupported scalar/non-scalar JSON values.

## Success Criteria

- `cleanToolSchema` recursively removes `deprecated` and `exclusiveMinimum`
  while preserving all current removals and type normalization.
- A scalar enum consisting of strings, booleans, JSON numbers, Go numeric
  values, or `nil` becomes an equivalent `[]any` of JSON string values.
- An enum containing an object, array, or another unsupported value is omitted
  in its entirety; no partial enum reaches Gemini.
- Existing tool conversion and Web Search behavior are unchanged.
- The implementation and test changes are one focused business commit. The
  Developer result and independent QA report are separate evidence commits.

## Context

- Repo: `F:/mcplugins/sub2api`
- Isolated worktree: `E:/codex-worktrees/sub2api/upstream-gemini-tool-schema-s254`
- Base: `main@249cbc223`
- Source: `19da0f240` from `upstream/main@e2d9b823f`
- `git apply --check` is expected to fail because local `cleanToolSchema`
  already diverged; adapt only the specified behavior, do not cherry-pick.

## Allowed Paths

- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service_test.go`
- `docs/workflow/worker-results/upstream-gemini-tool-schema-s254-result.md`
- `docs/workflow/qa-reports/upstream-gemini-tool-schema-s254-qa.md`

## Denied Paths

- `frontend/**`, `knowledge/**`, `outputs/**`, all Pixel Cafe and GroupBuy paths
- schema, migrations, Ent generated files, dependency files, configuration,
  provider credentials, containers, deployment, push, and every path not in
  Allowed Paths

## Constraints

- Keep the local Gemini Messages service topology and existing JSON schema
  policy. Do not introduce upstream packages, refactors, unrelated schema
  normalizers, or real provider calls.
- All product edits occur only in the isolated worktree. The primary worktree's
  existing dirty paths must remain untouched.
- No real provider, shared/production database, container, deployment, or push
  operation is authorized.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "TestCleanToolSchema_(DropsAmbiguousExclusiveMinimumWithoutConversion|RemovesNestedDeprecatedAndNormalizesMixedScalarEnum|DropsEnumWithNonScalarValue)" -count=10
go test ./internal/service -run "TestCleanToolSchema|TestConvertClaudeToolsToGeminiTools" -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
Pop-Location

gofmt -w backend/internal/service/gemini_messages_compat_service.go backend/internal/service/gemini_messages_compat_service_test.go
git diff --check
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" backend/internal/service/gemini_messages_compat_service.go backend/internal/service/gemini_messages_compat_service_test.go
git diff --name-only <base>..HEAD
git diff --cached --name-only
git diff --name-only
git ls-files -u
```

## Output

- One business commit limited to the two product/test owners.
- Developer result beginning `### DONE: upstream-gemini-tool-schema-s254`,
  `### FAILED: ...`, or `### BLOCKED: ...`, with changed paths, commands,
  source mapping, risks, and `knowledge_candidates`.
- Independent QA report beginning `### PASS: upstream-gemini-tool-schema-s254`,
  `### FAIL: ...`, or `### BLOCKED: ...`; QA writes only its report.

## Stop Rules

- Stop if the fix needs a path outside Allowed Paths, a new dependency, a
  schema/migration/configuration change, external state, or a provider call.
- Stop before mainline integration if a focused/default-tag/service/server gate
  fails, conflict markers or unmerged index entries appear, or the primary
  worktree changes.
