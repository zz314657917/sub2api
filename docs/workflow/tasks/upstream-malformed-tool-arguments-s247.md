# Task Contract: Upstream Malformed Tool Arguments S247

## Task ID

`upstream-malformed-tool-arguments-s247`

## Role

- Planner / Final Evaluator: Codex Controller
- Implementation owner after two worker stops: Codex Controller
- Independent QA Worker: `gpt-5.6-terra`

## Goal

Behaviorally adapt the final first-parent behavior of upstream merge
`fd6cd474d6d7a4c0c44b9346151376b81fa380cd` so truncated ordinary function-call
arguments cannot poison the Responses-to-Chat bridge or be finalized as a
completed Responses tool call.

The upstream PR contains source commits `e2d9ce0ca` and
`fbc9ee626d7298ddc1ba96c95349625f54737543`. The source branch temporarily
touched `openai_gateway_cc_pipeline.go`, but the final merge resolution does
not. S247 follows the final merge's five-file first-parent scope and adapts it
to the local, earlier bridge topology; it must not port the superseded source
layout mechanically.

The upstream service regression owner is `//go:build unit` in this older local
tree. The repository's unrelated unit-tag suite has existing compile errors,
so the local adaptation replaces that one test owner with a self-contained,
default-tag S247 test file. This is a test-topology substitution only; product
behavior and the other four upstream owners remain unchanged.

Frozen product baseline: local `main@1fe34a329`. The Controller may add only
workflow contract/phase evidence above that baseline before creating the
isolated Developer worktree.

## Success Criteria

1. When Responses history contains an ordinary `function_call` whose non-empty
   `arguments` are not valid JSON, the bridge omits that poisoned call instead
   of forwarding it to a Chat Completions provider.
2. The corresponding `function_call_output` is also omitted. Non-empty call IDs
   match by ID; an invalid call with an empty ID suppresses one corresponding
   empty-ID output without suppressing later unrelated items.
3. Empty or whitespace-only ordinary function arguments retain the existing
   normalization to `{}` and are not rejected.
4. A non-streaming Chat Completions response does not emit an ordinary
   Responses `function_call` with malformed JSON arguments. Other valid output
   items remain present, and an output-limit finish remains `incomplete`.
5. Before streaming finalization, accumulated ordinary tool-call arguments are
   validated. Malformed JSON returns an error and prevents
   `response.function_call_arguments.done`, completed output-item events, and
   the terminal `[DONE]` marker. Usage and request result metadata already
   collected from the upstream stream remain available for billing/error
   handling.
6. Valid tool calls at the output limit still finalize normally while the
   Responses status remains `incomplete`.
7. Existing custom tools, `tool_search`, namespace tools, media extraction,
   reasoning, stream draining, client-disconnect billing, retry/account-state,
   and S242/S243 compatibility behavior do not regress.
8. Business and evidence commits obey the exact allowlists, all acceptance
   commands pass, and the protected primary-worktree snapshot is unchanged.

## Allowed Paths

Developer business commit:

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_stream_lifecycle_test.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback_s247_test.go`

Developer evidence commit only:

- `docs/workflow/worker-results/upstream-malformed-tool-arguments-s247-result.md`

Independent QA evidence commit only:

- `docs/workflow/qa-reports/upstream-malformed-tool-arguments-s247-qa.md`

## Denied Paths

- `backend/internal/service/openai_gateway_cc_pipeline.go`; it is absent from
  the final upstream merge scope and must not be reintroduced from an
  intermediate source commit.
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`;
  this existing owner is unit-tagged locally and must remain unchanged. S247's
  equivalent service regression belongs in the allowed default-tag test file.
- All other backend, frontend, schema, migration, dependency, generated,
  deployment, Docker, knowledge, and workflow files except the exact report
  owner assigned to the active worker.
- All user-owned dirty and untracked primary-worktree paths, including the
  twenty-two current Pixel Cafe tracked paths and `outputs/`.
- Remote writes, push, force operations, history rewrites, real provider
  traffic, shared/production data, browser automation, and container changes.

## Constraints

- Adapt behavior to local `responsesInputToChatMessages`,
  `chatMessageToResponsesOutput`, `ChatCompletionsToResponsesStreamState`, and
  `streamChatCompletionsAsResponses`; do not import unrelated upstream bridge
  refactors.
- The six named focused regression tests are required S247 deliverables in the
  three allowed test owners. Their absence on the frozen baseline is expected;
  implement the tests before running focused discovery and acceptance.
- JSON validity enforcement applies only to ordinary function calls. Preserve
  the existing custom-tool input extraction, tool-search fallback semantics,
  namespace restoration, and empty-arguments normalization.
- Do not turn historical poison cleanup into request rejection. Invalid
  historical call/output pairs are skipped so a later user turn can self-heal.
- Do not emit final success events or `[DONE]` after streaming validation
  fails. Do not discard already-collected usage/result metadata.
- Do not redesign stream parsing, account scheduling, retry/error policy,
  billing, tool schemas, IDs, or media behavior.
- Do not install/update dependencies, call external services, or stage,
  overwrite, revert, or format unrelated work.

## Acceptance Commands

From `backend/` in the isolated worktree:

```powershell
go test ./internal/pkg/apicompat -list '^(TestResponsesInputToChatMessages_SkipsInvalidHistoricalFunctionCall|TestResponsesInputToChatMessages_SkipsInvalidEmptyCallIDOutput|TestChatCompletionsResponseToResponses_SkipsInvalidFunctionArguments|TestStream_InvalidToolArgumentsAreRejectedBeforeFinalize|TestStream_ValidToolCallAtOutputLimitKeepsIncompleteResponse)$'
go test ./internal/pkg/apicompat -run '^(TestResponsesInputToChatMessages_SkipsInvalidHistoricalFunctionCall|TestResponsesInputToChatMessages_SkipsInvalidEmptyCallIDOutput|TestChatCompletionsResponseToResponses_SkipsInvalidFunctionArguments|TestStream_InvalidToolArgumentsAreRejectedBeforeFinalize|TestStream_ValidToolCallAtOutputLimitKeepsIncompleteResponse)$' -count=10
go test ./internal/service -list '^TestStreamChatCompletionsAsResponses_RejectsInvalidToolArgumentsAtOutputLimit$'
go test ./internal/service -run '^TestStreamChatCompletionsAsResponses_RejectsInvalidToolArgumentsAtOutputLimit$' -count=10
go test ./internal/pkg/apicompat -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -l internal/pkg/apicompat/chatcompletions_responses_bridge.go internal/pkg/apicompat/chatcompletions_responses_bridge_test.go internal/pkg/apicompat/chatcompletions_responses_stream_lifecycle_test.go internal/service/openai_gateway_responses_chat_fallback.go internal/service/openai_gateway_responses_chat_fallback_s247_test.go
```

From the worktree root:

```powershell
git diff --check
git diff --cached --name-only
git ls-files -u
git merge-base --is-ancestor e2d9ce0ca upstream/main
git merge-base --is-ancestor fbc9ee626d7298ddc1ba96c95349625f54737543 upstream/main
git merge-base --is-ancestor fd6cd474d6d7a4c0c44b9346151376b81fa380cd upstream/main
git log --oneline fd6cd474d6d7a4c0c44b9346151376b81fa380cd..upstream/main -- backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go backend/internal/pkg/apicompat/chatcompletions_responses_stream_lifecycle_test.go backend/internal/service/openai_gateway_responses_chat_fallback.go backend/internal/service/openai_gateway_responses_chat_fallback_test.go
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go backend/internal/pkg/apicompat/chatcompletions_responses_stream_lifecycle_test.go backend/internal/service/openai_gateway_responses_chat_fallback.go backend/internal/service/openai_gateway_responses_chat_fallback_s247_test.go
```

The Controller must additionally verify the exact business/evidence commit
allowlists, the final merge's five-file first-parent scope and the one-for-one
local service-test owner substitution, absence of later upstream product-owner
touches, preservation of S242/S243 custom-tool/replay behavior, empty
index/conflict state, and the primary-worktree protected snapshot.

The protected primary-worktree patch ID is scoped to these twenty-two tracked
user-owned paths only:

- `backend/internal/handler/admin/cafe_room_handler.go`
- `backend/internal/handler/admin/cafe_room_handler_test.go`
- `backend/internal/repository/cafe_room_repo.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/service/cafe_public.go`
- `backend/internal/service/cafe_public_test.go`
- `backend/internal/service/cafe_room_service.go`
- `backend/internal/service/cafe_room_service_test.go`
- `frontend/src/api/admin/cafeRooms.ts`
- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/features/pixelCafe/components/CafeScene.vue`
- `frontend/src/features/pixelCafe/components/SceneFallback.vue`
- `frontend/src/features/pixelCafe/components/__tests__/CafeScene.spec.ts`
- `frontend/src/features/pixelCafe/renderer/assetManifest.ts`
- `frontend/src/features/pixelCafe/renderer/createCafeRenderer.ts`
- `frontend/src/features/pixelCafe/renderer/sceneLayout.ts`
- `frontend/src/i18n/locales/en/admin/pixelCafe.ts`
- `frontend/src/i18n/locales/zh/admin/pixelCafe.ts`
- `frontend/src/types/pixelCafe.ts`
- `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`
- `frontend/src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts`

Their combined stable patch ID must remain
`941b1edf9df9e465a6100007edfc4a6715e38b5e`. These untracked files must retain
their SHA-256 values:

- `e6cd621c9f2df7b5d4a5521e8904c95731996533761e01add8ba544b014e0952`
  `backend/internal/repository/cafe_room_account_option_test.go`
- `1e3830c11e13b586f09c254c1a468878a84f932a8615be58fb479cfd607d66ff`
  `frontend/src/views/admin/pixelCafe/components/CafeRoomAccountPicker.vue`
- `49ec0eaadeb4d49f0eb01853629769be601e8896c5eb3ee2d5ae98db83717c32`
  `frontend/src/views/admin/pixelCafe/components/__tests__/CafeRoomAccountPicker.spec.ts`
- `f21e77c5d3cc82727a516bc4b2cb901e53c2d7505a448d5dd551b74ddfb3ece0`
  `outputs/20260725-static-residential-socks5/静态住宅 IP (1)-sub2-socks5.json`
- `438fdda26586fa3a5857b927d7dbbfac4868bb55a6a1e8bfdac540296a497f4c`
  `outputs/20260731-static-residential-sub2/静态住宅 IP (3)-sub2api.json`

The primary staged and unmerged indexes must remain empty.

## Output

- Controller produces one business commit containing only the five local
  product/test paths and one separate evidence commit containing only
  `docs/workflow/worker-results/upstream-malformed-tool-arguments-s247-result.md`.
- Controller result first line must be exactly
  `### DONE: upstream-malformed-tool-arguments-s247`,
  `### BLOCKED: upstream-malformed-tool-arguments-s247`, or
  `### FAILED: upstream-malformed-tool-arguments-s247`.
- Independent QA modifies only
  `docs/workflow/qa-reports/upstream-malformed-tool-arguments-s247-qa.md`; its
  first line must be exactly
  `### PASS: upstream-malformed-tool-arguments-s247`,
  `### FAIL: upstream-malformed-tool-arguments-s247`, or
  `### BLOCKED: upstream-malformed-tool-arguments-s247`.
- Reports list changed files, commands, key output, risks, contract compliance,
  and `knowledge_candidates` without unrelated long logs.

## Stop Rules

- Stop if `gpt-5.6-terra` is unavailable; do not silently replace the model.
- Stop if implementation requires `openai_gateway_cc_pipeline.go`, any path
  outside the allowlist, a bridge/scheduler/retry redesign, dependency changes,
  frontend/schema work, external traffic, or shared state.
- Stop if, after adding the six required S247 regression tests, focused
  selectors still discover fewer than six total tests; also stop if a baseline
  failure belongs outside this contract, custom/tool-search/namespace behavior
  must change, or any protected-primary fingerprint changes unexpectedly.

## Budget

- worker_mode: stopped after two attributed failures; Controller takeover
- qa_worker_mode: native `gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- developer_max_budget_usd: exhausted for the stopped worker loop
- qa_max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Status

`contract-approved`

## Worker Output

Same requirements as `Output`; this compatibility heading is retained for the
worker dispatcher.
