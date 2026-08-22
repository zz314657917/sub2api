### PASS: upstream-openai-empty-tool-name-s239-a

- QA worktree HEAD: `3cfb2360a` (`fcd7f71e8` business + Controller report).
- The requested contract file `docs/workflow/tasks/upstream-openai-empty-tool-name-s239-a.md`
  is absent from this worktree. The available Controller result report was used
  to recover the stated scope; this is a documentation risk, not a test failure.

## Commands and results

- From `backend/`, `go test ./internal/pkg/apicompat -run 'TestResponsesEventToChatChunks_ArgumentsDeltaOmitsEmptyName' -count=10`:
  PASS.
- From `backend/`, `go test ./internal/pkg/apicompat -count=1`: PASS.
- From `backend/`, `go test ./cmd/server -run '^$' -count=1`: PASS (compile-only; no tests run).
- From `backend/`, `gofmt -l internal/pkg/apicompat/types.go internal/pkg/apicompat/responses_to_chatcompletions_tool_name_test.go`:
  no output.
- `git diff --check`: PASS.
- Business patch scope (`fcd7f71e8^..fcd7f71e8`): exactly
  `backend/internal/pkg/apicompat/types.go` and
  `backend/internal/pkg/apicompat/responses_to_chatcompletions_tool_name_test.go`.
- QA allowlist: only this report is added after the Controller base; no other QA
  worktree paths changed.
- `git ls-files -u`: empty; no conflict markers in the two business owners.
- `git merge-base --is-ancestor f646a1f974c26152160ef8327a7d6b9e3488ee83 upstream/main`:
  PASS (exit code 0).
- The protected primary worktree `F:/mcplugins/sub2api` dirty/untracked path
  snapshot was inspected and remained unchanged during QA.

## Risk and exclusions

The change only adds `omitempty` to streamed Chat Completions tool-call names
and a regression test. The server check is compile-only. Provider, database,
deployment, container, push, and shared-state operations were not run.
