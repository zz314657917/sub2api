# Upstream OpenAI WS Replay S243

## Task ID

`upstream-openai-ws-replay-s243`

## Phase

`contract-draft`

## Role

The Developer Worker implements this contract in an isolated worktree.
Codex remains Controller and Final Evaluator. Independent QA must use the
repository Agent Matrix QA route and produce a separate report.

## Goal

Adapt the behavior from upstream `25da02ddd` and `66808413d` to the local
monolithic OpenAI WebSocket forwarder. HTTP-bridge replay must recognize both
array and object `input` shapes, avoid replaying a full tool history when the
current request already carries matching tool-call outputs and context, and
drop historical orphan tool-call context that has no paired output. Current
tool calls in the request and paired function/custom/MCP/tool-search calls must
remain intact.

## Frozen Base And Provenance

- Local base: current `main` at dispatch time (`28cf22bb1` expected).
- Upstream source commits: `25da02ddd`, `66808413d`.
- Upstream tip: `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54`.
- Both source commits must be ancestors of the checked upstream tip.
- Upstream splits the local monolithic owners; adapt into the existing
  `openai_ws_forwarder.go` and `openai_tool_continuation.go` owners rather than
  copying the upstream file split.

## Success Criteria

1. `AnalyzeToolCallOutputContextCoverageBytes` (or an equivalent local owner)
   handles `input` arrays and single object items, tracks every output
   `call_id`, and reports complete context only when every output is paired by
   a matching call context or `item_reference`.
2. The WS-HTTP bridge does not force replay merely because a request contains
   a complete tool-call/output pair; incomplete or anchored continuation
   requests still replay as required.
3. Historical replay sanitization removes only unpaired historical tool-call
   context items. It preserves ordinary input, paired function/custom/MCP/tool
   calls, item references, and current-turn orphan tool calls supplied by the
   client.
4. Existing OAuth/API-key WS handling, response IDs, SSE framing, usage
   accounting, failover, and custom-tool mapping from S242 remain unchanged.
5. Focused tests cover array/object coverage, orphan historical custom tool
   filtering, paired historical function/custom preservation, item-reference
   non-pairing, current-turn preservation, and bridge request counts/payloads.

## Allowed Paths

- `backend/internal/service/openai_tool_continuation.go`
- `backend/internal/service/openai_tool_continuation_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_ws_http_bridge_test.go`
- `docs/workflow/results/upstream-openai-ws-replay-s243-result.md`

## Denied Paths And Constraints

- Preserve all existing user dirty and untracked paths, including Pixel Cafe,
  API-key service files, room-policy tests, tutorials, knowledge files, and
  `outputs/`; they must remain byte-for-byte unchanged.
- No migrations, schema, repositories, billing/pricing, frontend, config,
  dependencies, generated files, OAuth identity, image generation, 429 retry,
  binary-frame policy, turn-start pricing, provider traffic, database,
  container, deployment, or push operations.
- Do not rename or split the local OpenAI forwarder files.
- Do not weaken validation for malformed JSON or missing `call_id`.

## Acceptance Commands

Run from `backend/` in the isolated worktree:

```powershell
go test ./internal/service -run "TestBuildOpenAIWSReplayInputSequence|TestAnalyzeToolCallOutputContextCoverageBytes|TestOpenAIWSHTTPBridge.*Replay|TestOpenAIWSRawPayloadHasToolCallOutput" -count=10
go test ./internal/service
go test ./cmd/server -run '^$' -count=1
gofmt -w <changed Go files>
git diff --check
```

Controller and QA must additionally verify exact allowlist, no conflict
markers, empty unmerged index, upstream ancestry, patch scope, and protected
main dirty-state preservation. A selector discovering no tests is not evidence.

## Output

Developer output is one business commit plus
`docs/workflow/results/upstream-openai-ws-replay-s243-result.md`, whose first
line is `### DONE: upstream-openai-ws-replay-s243` (or explicit BLOCKED/FAILED).
Independent QA output must begin with `### PASS: upstream-openai-ws-replay-s243`,
`### FAIL: ...`, or `### BLOCKED: ...`.

## Stop Rules

- Stop if the behavior requires a schema/product prerequisite, touches denied
  paths, changes OAuth/Grok contracts, or cannot be isolated.
- Stop if focused tests are undiscoverable, full service/server compile fails,
  or the protected main worktree changes.
- No silent model replacement or widening of the allowlist.

## Evaluator Review

`PENDING`: Controller must verify source ancestry, local topology mapping,
allowlist, acceptance commands, and protected dirty-state boundary before
dispatching implementation and QA.
