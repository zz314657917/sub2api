### DONE: upstream-v0180-responses-lite-s253

## Scope

- Implemented only the approved Responses Lite prerequisite chain, raw Chat Completions empty tool-identity sanitizer, OAuth manifest model sync with `slug` fallback, Ops list-return behavior, and the fresh-table priority default.
- Kept the primary S252 worktree untouched; all source changes were made and committed in the isolated S253 worktree.

## Business commits

- `093a662ef fix(openai): normalize responses lite requests`
- `2c0d6f720 fix(openai): preserve streamed tool identities and oauth models`
- `75e2b804a fix(admin): preserve ops list context and show priority`
- `8c6378b8a fix(openai): enforce serial lite tool calls`

## Implemented behavior

- Lite HTTP, ctx-pool WS, WS v2 passthrough, and WS HTTP bridge normalize namespace tools to `input.additional_tools`, force `reasoning.context=all_turns`, and set `parallel_tool_calls=false` only when tools exist.
- Invalid Lite tool/parallel/reasoning shapes reject before upstream I/O; raw SSE preserves all fields except present empty `tool_calls[].id` and `tool_calls[].function.name`.
- OAuth model sync reuses the local Codex manifest endpoint/authentication conventions and accepts model `slug` values.
- Ops detail returns to its source list without resetting its state; saved account column preferences remain authoritative while new tables expose priority.

## Controller checks

- Focused Go Lite, WS bridge/ingress, raw SSE, and OAuth model-sync tests passed.
- UI focused Vitest, typecheck, and production build passed.
- Diff check, unresolved-index check, allowlist review, and primary-worktree protection review passed.

## Risks and knowledge candidates

- No real provider request, login-state UI interaction, push, deployment, or database action was performed.
- No durable knowledge candidate: this is a scoped, upstream-compatibility port with behavior covered in local regression tests.
