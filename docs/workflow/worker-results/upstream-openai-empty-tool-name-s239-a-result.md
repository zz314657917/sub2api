### DONE: upstream-openai-empty-tool-name-s239-a

- Business commit: `fcd7f71e8` (`fix(apicompat): omit empty streamed tool names`).
- Changed only `backend/internal/pkg/apicompat/types.go` and
  `backend/internal/pkg/apicompat/responses_to_chatcompletions_tool_name_test.go`.
- `ChatFunctionCall.Name` now uses `json:"name,omitempty"`; non-empty names
  remain unchanged, while arguments-only streamed deltas omit `name`.
- Focused regression verifies the initial `response.output_item.added` delta
  carries `exec`, and repeated `response.function_call_arguments.delta` chunks
  contain `arguments` without a `name` field.

Acceptance evidence from `backend/`:

- `go test ./internal/pkg/apicompat -run 'TestResponsesEventToChatChunks_ArgumentsDeltaOmitsEmptyName' -count=10` PASS
- `go test ./internal/pkg/apicompat -count=1` PASS
- `go test ./cmd/server -run '^$' -count=1` PASS
- `gofmt -l internal/pkg/apicompat/types.go internal/pkg/apicompat/responses_to_chatcompletions_tool_name_test.go` produced no output

Repository checks:

- `git diff --check` PASS.
- Business commit staged path audit contained exactly the two allowed product/test paths.
- No unmerged index entries or conflict markers in the two owners.
- `f646a1f974c26152160ef8327a7d6b9e3488ee83` is an ancestor of `upstream/main`.
- Protected primary worktree dirty/untracked snapshot was inspected before and
  after work and was not modified.

Risks: no provider, database, deployment, container, push, or shared-state
operation was performed. The server command is compile-only (`-run '^$'`).
