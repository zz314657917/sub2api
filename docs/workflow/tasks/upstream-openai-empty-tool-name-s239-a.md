# Upstream OpenAI Empty Streamed Tool Name S239-A

## Task ID

`upstream-openai-empty-tool-name-s239-a`

## Role

Developer and QA are independent `gpt-5.6-terra` workers when dispatched. Codex
is Planner and Final Evaluator. This is a local-topology adaptation of the
product diff carried by upstream merge commit `f646a1f97`; do not cherry-pick
the divergent merge tree as a whole.

## Goal

Port the upstream API-compatibility fix so streamed Chat Completions tool-call
argument deltas omit an empty `function.name`, while the initial tool-call
delta still carries a non-empty name. This prevents clients from overwriting a
previously accumulated tool name with `""` and leaves non-empty tool names and
non-streaming responses unchanged.

## Success Criteria

- `ChatFunctionCall.Name` omits only empty JSON values; non-empty names remain
  serialized exactly as before.
- A `response.output_item.added` function call emits its non-empty tool name,
  and subsequent `response.function_call_arguments.delta` chunks contain
  `arguments` without a `name` field.
- The focused regression is default-tag discoverable and passes repeatedly;
  the complete `internal/pkg/apicompat` package and server compile remain
  green.
- No user dirty or untracked file changes are staged or committed.

## Baseline And Provenance

- Frozen base: local `main@f04104623`.
- Upstream source: merge `f646a1f97`, an ancestor of `upstream/main@67380eafd`;
  its first-parent product diff is `fd24923f6..f646a1f97`.
- The first-parent diff applies cleanly to the two local owners after the
  local test path is created; the merge commit itself must not be replayed as
  an unrelated upstream history merge.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`, this contract.
- Local owner: `backend/internal/pkg/apicompat/types.go` and the new focused
  test `backend/internal/pkg/apicompat/responses_to_chatcompletions_tool_name_test.go`.

## Allowed Paths

- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/pkg/apicompat/responses_to_chatcompletions_tool_name_test.go`
- `docs/workflow/worker-results/upstream-openai-empty-tool-name-s239-a-result.md`
- `docs/workflow/qa-reports/upstream-openai-empty-tool-name-s239-a-qa.md`

## Denied Paths

All other product, test, generated, frontend, schema, migration, dependency,
configuration, workflow-status, knowledge, user-dirty, and untracked paths.
No provider traffic, shared database, container, deployment, remote write,
push, force operation, or history rewrite is allowed.

## Constraints

- Keep the change limited to the JSON tag and its focused regression; do not
  redesign stream conversion or tool-name rewrite behavior.
- Preserve non-empty names, IDs, indexes, arguments, finish reasons, and all
  existing request/response fields.
- Do not overwrite or clean the primary worktree's existing changes.
- Use default Go build tags; no real provider, database, or shared state.

## Acceptance Commands

From `backend/`:

```powershell
go test ./internal/pkg/apicompat -run 'TestResponsesEventToChatChunks_ArgumentsDeltaOmitsEmptyName' -count=10
go test ./internal/pkg/apicompat -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -l internal/pkg/apicompat/types.go internal/pkg/apicompat/responses_to_chatcompletions_tool_name_test.go
```

From the repository root:

```powershell
git diff --check
git diff --name-only HEAD
git diff --cached --name-only
```

Also verify the exact product allowlist, clean index, no conflict markers in
the two product/test owners, upstream ancestry, patch scope, and preservation
of the pre-existing dirty/untracked snapshot.

## Output

- One business implementation commit containing only the two product/test
  paths.
- One worker evidence commit containing only the result report.
- The worker report first non-frontmatter line must be
  `### DONE: upstream-openai-empty-tool-name-s239-a`,
  `### BLOCKED: upstream-openai-empty-tool-name-s239-a`, or
  `### FAILED: upstream-openai-empty-tool-name-s239-a`.
- Independent QA may write only
  `docs/workflow/qa-reports/upstream-openai-empty-tool-name-s239-a-qa.md` in
  its separate worktree; its first line must be
  `### PASS: upstream-openai-empty-tool-name-s239-a`,
  `### FAIL: upstream-openai-empty-tool-name-s239-a`, or
  `### BLOCKED: upstream-openai-empty-tool-name-s239-a`.

## Stop Rules

- Stop on any required path outside the allowlist or any need for gateway,
  schema, frontend, dependency, provider, database, deployment, or container
  changes.
- Stop if local stream conversion already omits empty names or if the change
  would alter non-empty tool-name serialization; report the owner evidence
  instead of widening the patch.
- Stop if a test needs a real provider or shared state.
- Stop on any unexpected protected-main change.

## Budget

- worker_mode: native `gpt-5.6-terra`
- qa_worker_mode: native `gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- worktree_root: `E:/codex-worktrees`
