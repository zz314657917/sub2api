# Task Contract: upstream-codex-mcp-tool-bridge-s67b-fix1

## Task ID

`upstream-codex-mcp-tool-bridge-s67b-fix1`

## Status

`approved`

## Role

You are the Generator worker fixing the S67b streaming lifecycle defect found by independent QA.

## Goal

Adapt only the response-stream lifecycle subset required from upstream ancestor `f10bca815`: allocate output indices in actual item-open order, emit a complete reasoning item lifecycle, and keep terminal `response.output` indices consistent for ordinary, custom, `tool_search`, and namespace tool calls.

## Success Criteria

- A tool-only stream opens and closes its first tool at `output_index=0`, matching `response.completed.output[0]`.
- A reasoning-plus-tool stream opens the reasoning item before its first delta, closes it before the tool item, and uses indices matching final output order.
- Message text emits the required message/content-part lifecycle with a stable dynamic output index.
- Ordinary function, custom tool, `tool_search`, namespace restore, late tool-name, and parallel tool-call behavior remain compatible with S67b.
- Existing S67a effort metadata and S67c Ops logging remain untouched.
- Focused lifecycle tests assert event ordering, item types, and every streamed/final output index.

## Context

- Integration branch head before fix: `da4fc642e`.
- Failed QA report: `docs/workflow/qa-reports/upstream-v0151-protocol-wave2-s67-qa.md`.
- Upstream prerequisite: `f10bca815`, which is an ancestor of the S67b sequence ending at `90e9d03de`.
- Port only the response-direction state/index/lifecycle subset. Do not import the broad request-normalization refactor from `f10bca815`.

## Allowed Paths

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_custom_tools_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_stream_lifecycle_test.go`
- `docs/workflow/worker-results/upstream-codex-mcp-tool-bridge-s67b-fix1-result.md`

## Denied Paths

- All other paths, including service gateway files, billing/pricing, migrations, frontend, deployment, and production configuration.
- S67a-owned GPT effort files and S67c-owned Ops logger files.
- `knowledge/**` and global memories.

## Constraints

- Do not cherry-pick all of `f10bca815`; its request-direction redesign is outside this fix.
- Preserve the S67b custom/tool-search/namespace metadata maps and late-name classification.
- Do not change non-streaming response conversion semantics.
- Use dynamic indices stored on stream state; do not infer terminal indices again from map keys.
- Do not repair unrelated full-suite drift.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/pkg/apicompat -run "Test.*Stream.*Lifecycle|Test.*ToolOnly|Test.*Reasoning.*Tool|Test.*Custom.*Tool.*Stream|Test.*ToolSearch.*Stream|Test.*Namespace.*Stream|Test.*Late" -count=1
go test ./internal/pkg/apicompat -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-codex-mcp-tool-bridge-s67b-fix1-result.md` with first line `### DONE: upstream-codex-mcp-tool-bridge-s67b-fix1`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit all contract changes on the assigned branch and return the commit hash.

## Stop Rules

- Stop if the fix requires request-direction message normalization or any path outside Allowed Paths.
- Stop if streamed indices still differ from final `response.output` positions in any covered lifecycle.
- Stop if ordinary/custom/tool-search/namespace item types differ between `output_item.added`, `output_item.done`, and terminal output.
- Do not repair unrelated unit-suite drift.
