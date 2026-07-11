### PASS: upstream-v0151-protocol-wave2-s67-retest1

# QA Report

## Scope

- Retest head: `35d2b6be0`; Fix1 implementation commit: `d1c858392`.
- S67 audit baseline: `09bfc7e9b`; original failed QA: `upstream-v0151-protocol-wave2-s67-qa`.
- Fix contract/result reviewed: `upstream-codex-mcp-tool-bridge-s67b-fix1`; the result begins with the exact required `### DONE:` verdict.
- QA worktree: `E:/codex-worktrees/sub2api/upstream-v0151-protocol-wave2-s67-retest1`; all Go commands used `GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-protocol-wave2-s67-retest1/.tmp/go-build`.

## Findings

- The original tool-only index failure is closed. `TestStreamLifecycle_ToolOnlyIndicesMatchTerminalForEveryToolKind` verifies ordinary function, custom tool, `tool_search`, and namespace tool streams: the first tool opens/closes at `output_index=0`, the terminal item is `response.completed.output[0]`, and streamed/terminal type and item ID agree.
- The original reasoning-plus-tool lifecycle failure is closed. `TestStreamLifecycle_ReasoningClosesBeforeToolAndIndicesMatchTerminal` verifies `reasoning output_item.added < reasoning delta < reasoning output_item.done < tool output_item.added`, with terminal indices `reasoning=0` and `function_call=1`.
- Manual implementation audit agrees with the tests: `chatcompletions_responses_bridge.go:806-874` owns sequential allocation, `:902-982` opens/closes reasoning before later item types and stores each tool index, `:1099-1150` emits the complete reasoning lifecycle, `:1195-1230` reuses the stored tool index for added/delta/done, and `:1304-1367` places terminal items by those same stored indices.
- Ordinary/custom/`tool_search`/namespace, late-name, and parallel behavior is consistent. The lifecycle helper validates every streamed event against the terminal item type/index; focused legacy late custom/function/namespace and custom/tool-search/namespace stream tests also pass. `TestStreamLifecycle_LateCustomNameKeepsAllocatedIndex` preserves the pre-name allocation, while `TestStreamLifecycle_ParallelToolCallsUseOpenOrderIndices` verifies stable open-order indices `0/1`.
- Message text now has a complete item/content-part lifecycle on one stable dynamic index, verified by `TestStreamLifecycle_MessageContentPartUsesStableDynamicIndex`.
- No S67a GPT-5.6 effort or S67c Ops regression was found: their combined service and handler checks pass.
- The Fix1 implementation commit touches exactly three approved artifacts: the apicompat bridge, the new lifecycle test, and the Fix1 worker result. Across `09bfc7e9b..HEAD`, the union audit classifies 23 business/test paths as contract-allowed, four worker results, four workflow artifacts, and zero Denied Paths.
- Per the retest instruction, the full service suite was not rerun. The initial QA already isolated its only failures to unchanged `group_peak_rate_test.go` timezone/peak-window drift; Fix1 changes no service path, so that known package-level limitation remains unrelated to S67.

## Executed Checks

- `go test ./internal/pkg/apicompat -run "Test.*Stream.*Lifecycle|Test.*ToolOnly|Test.*Reasoning.*Tool|Test.*Custom.*Tool.*Stream|Test.*ToolSearch.*Stream|Test.*Namespace.*Stream|Test.*Late" -count=1 -v` - PASS (`0.677s`).
- `go test ./internal/pkg/apicompat -count=1` - PASS (`0.136s`).
- Combined S67 service contract run covering GPT-5.6 Max/candidates, raw Chat, compact, Responses fallback, and Messages usage - PASS (`6.067s`).
- `go test ./internal/handler -run "TestLogOpsStreamError|TestMarkOpsStreamError|TestOpsCaptureWriter|TestOpsErrorLogger" -count=1` - PASS (`7.917s`).
- `go test ./internal/handler -count=1` - PASS (`28.542s`).
- Static lifecycle audit of ordinary/custom/`tool_search`/namespace, late-name, and parallel streams - PASS; stored `ToolOutputIndex` and item IDs are reused by streamed and terminal construction.
- `git show --name-status d1c858392` - PASS ownership audit: two approved apicompat paths plus the approved Fix1 worker result only.
- `git diff --check 09bfc7e9b..HEAD` - PASS.
- Conflict-boundary scan across all changed paths - PASS (`NO_CONFLICT_MARKERS`).
- Allowed/Denied union audit for `09bfc7e9b..HEAD` - PASS: 23 contract-allowed business/test paths, four worker results, four workflow artifacts, zero Denied Path changes.

## Unverified Risks

- No real Codex CLI plus MCP/custom/tool-search request was sent through a Chat-only upstream. Retest evidence is deterministic protocol-event testing and package regression coverage.
- The full service package remains non-green according to the initial QA because of unrelated peak-rate timezone drift; this retest intentionally did not repeat that command.
- No race-detector run was required by the Fix1 or S67 contracts.

## Recommendation

- PASS. The two initial blocking lifecycle defects are closed, the S67 combination regression is green within contract scope, and S67 may proceed to the next workflow gate while preserving the unrelated full-service drift as a separate follow-up.
