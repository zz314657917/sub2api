### FAIL: upstream-v0151-protocol-wave2-s67

# QA Report

## Scope

- Integration head: `6442264a1` (integrated implementation head `fbe5dd123`).
- Audit baseline: `09bfc7e9b`.
- Contracts reviewed: `upstream-gpt56-max-effort-s67a`, `upstream-codex-mcp-tool-bridge-s67b`, and `upstream-compact-ops-stream-log-s67c`.
- Worker results reviewed: all three corresponding reports begin with the required exact `### DONE:` verdict.
- QA worktree: `E:/codex-worktrees/sub2api/upstream-v0151-protocol-wave2-s67-qa`; all Go commands used `GOTMPDIR=E:/codex-worktrees/sub2api/upstream-v0151-protocol-wave2-s67-qa/.tmp/go-build`.

## Findings

- **Blocking S67b defect: tool-only SSE uses an output index that disagrees with the terminal response.** `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go:1093`, `:1105`, and `:1128` derive tool indices as `idx + 1`, so a tool-only response whose upstream tool index is zero is announced and closed at output index `1`. The terminal builder starts at `:1202`; with no reasoning and at least one tool, the message condition at `:1214` is false, so the same tool is the first and only `response.completed.output[0]`. A strict Responses client can therefore associate the streamed lifecycle with a nonexistent second output item.
- **Minimal reproducer / missing assertion:** `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_custom_tools_test.go:128` already constructs the minimal case: a fresh state, one custom `exec` call at upstream index `0`, then `FinalizeChatCompletionsResponsesStream`. Its tool `response.output_item.added` has `OutputIndex == 1`, while the terminal response has one custom-tool item at `Output[0]`. Adding `assert.Equal(t, 0, added.OutputIndex)` to that existing test fails against the integrated implementation; the test currently checks item type/name/input but never the index.
- **Blocking bridge lifecycle defect on reasoning plus tool calls.** At `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go:896`, a `reasoning_content` delta is emitted at output index `0` without first opening a reasoning `response.output_item.added`, and the local bridge has no matching reasoning summary/item completion lifecycle. `backend/internal/pkg/apicompat/chatcompletions_responses_test.go:524` verifies only that no visible-text fallback is synthesized and that the terminal output contains reasoning plus a function call; it never records opened output items or asserts that each reasoning delta references an open reasoning item.
- Comparative audit confirms this is not the upstream protocol shape on which the S67b source sequence was based. `git merge-base --is-ancestor f10bca815 90e9d03de` exits `0`, and commit timestamps place `f10bca815` on 2026-05-31 before all seven S67b references dated 2026-07-07 through 2026-07-10. In `git show upstream/main:backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`, lines `1000-1003` define `nextOutputIndex`, line `1025` defines `ToolOutputIndex`, lines `1076-1079` allocate item-order indices, line `1110` opens reasoning before its delta, and line `1157` stores the allocated tool index. The local port retained the older fixed-index lifecycle while adding the later tool types.
- No defect was found in the S67a GPT-5.6 effort paths or the S67c Ops stream logging paths. Their focused tests and the combined service/handler tests passed.
- The full `internal/service` package remains red only in existing `group_peak_rate_test.go` assertions: `TestPeakMultiplierAt_Boundaries` (`14:00`, `15:30`, `17:59`), `TestPeakMultiplierAt_RespectsTimezoneLocation`, `TestPeakMultiplierAt_StandardTypeDegradesToOne`, `TestPeakMultiplier_GatewayBillingSequence` (`peak_hour_amplifies_token_multiplier_only`, `image_independent_mode_decoupled_from_peak`), and `TestPeakMultiplier_SnapshotRoundTrip`. Neither `group_peak_rate.go` nor `group_peak_rate_test.go` is changed by `09bfc7e9b..HEAD`, so these failures are unrelated pre-existing suite drift, but the package command is accurately recorded as FAIL.
- Allowed/Denied audit found 27 changed paths: 22 business/test paths are in the union of the three contracts, three are contract-owned worker results, and two are workflow integration metadata (`docs/workflow/status.md`, `docs/workflow/main-log.md`). There are zero Denied Path changes.

## Executed Checks

- `go test ./internal/pkg/apicompat -count=1` - PASS (`1.264s`).
- Combined S67 service contract run covering GPT-5.6 Max/candidates, raw Chat, compact, Responses fallback, and Messages usage - PASS (`11.176s`). The run selected `TestNormalizeOpenAIReasoningEffortForGPT56Max`, `TestNormalizeOpenAICodexCompactReasoningEffortGPT56MaxScopes`, mapped GPT-5.6 forward/raw-Chat tests, reasoning-candidate and WS passthrough tests, `TestForwardResponses_*`, `TestCopyOpenAIUsage*`, `TestExtractCCStreamUsageCapturesCacheWriteTokens`, and `TestExtractOpenAIReasoningEffortFromFallbackCandidatesUsesSuffixFallback`.
- `go test ./internal/handler -run "TestLogOpsStreamError|TestMarkOpsStreamError|TestOpsCaptureWriter|TestOpsErrorLogger" -count=1` - PASS (`5.493s`).
- `go test ./internal/handler -count=1` - PASS (`20.720s`).
- `go test ./internal/service -count=1` - FAIL (`52.107s`) only in the exact existing peak-rate cases listed under Findings; no S67-targeted test failed.
- Focused SSE/tool wire run - PASS at the assertion level, but implementation audit exposed the missing lifecycle assertions described under Findings. Covered tests include:
  - tool-only: `TestChatCompletionsChunkToResponsesEvents_CustomToolCallStream`;
  - reasoning plus tool: `TestChatCompletionsStreamToResponses_DeepSeekReasoningToolCallDoesNotFallbackToVisibleText`;
  - late names: `TestChatCompletionsChunkToResponsesEvents_CustomToolNameArrivesLate`, `TestChatCompletionsChunkToResponsesEvents_FunctionToolNameArrivesLate`, and `TestChatCompletionsChunkToResponsesEvents_NamespacedToolNameArrivesLate`;
  - custom/tool-search/namespace conversion and restoration: `TestResponsesToChatCompletionsRequest_CustomToolBecomesFunctionTool`, `TestResponsesToChatCompletionsRequest_ToolSearchToolBecomesProxyFunction`, `TestResponsesToChatCompletionsRequest_NamespaceToolFlattensChildren`, `TestResponsesToChatCompletionsRequest_ToolSearchToolChoiceMapsToProxy`, `TestChatCompletionsResponseToResponses_ToolSearchCallOutputItem`, `TestChatCompletionsResponseToResponses_NamespacedToolCallRestored`, and `TestChatCompletionsChunkToResponsesEvents_NamespacedToolCallStream`;
  - wire shape: all `TestWire_*` cases in `responses_stream_event_wire_test.go`.
- Static comparison of local bridge lifecycle against `upstream/main` and predecessor commit `f10bca815` - FAIL for local output-index/reasoning lifecycle equivalence.
- `git diff --check 09bfc7e9b..HEAD` - PASS.
- Conflict-boundary scan across all changed paths - PASS (`NO_CONFLICT_MARKERS`).
- Allowed/Denied union audit for `09bfc7e9b..HEAD` - PASS: 22 contract-allowed business/test paths, three worker result paths, two workflow metadata paths, zero Denied Path changes.
- Worker result first-line audit - PASS: all three are exact `### DONE: <task-id>` lines.

## Unverified Risks

- No real Codex CLI plus MCP/custom/tool-search request was sent through a Chat-only upstream. The blocking lifecycle findings are derived from deterministic event construction and terminal-output ordering, not a live upstream session.
- No race-detector run was required by the S67 contracts. Concurrency behavior beyond the executed package tests was not requalified here.
- The unrelated peak-rate timezone drift still prevents treating the full service package as green.

## Recommendation

- **FAIL; do not merge S67 yet.** Return the bridge lifecycle to the S67b owner. Allocate output indices in item-open order, preserve those indices through delta/done events and `response.completed.output`, and open/close reasoning items around reasoning deltas. Add assertions for tool-only index `0`, reasoning-plus-tool indices/lifecycle, and final-output consistency, then rerun the complete S67 QA matrix.
