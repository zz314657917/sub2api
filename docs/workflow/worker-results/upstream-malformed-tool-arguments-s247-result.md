### DONE: upstream-malformed-tool-arguments-s247

## Changed files

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_stream_lifecycle_test.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback_s247_test.go`
- `docs/workflow/worker-results/upstream-malformed-tool-arguments-s247-result.md`

## Commands run

- Reviewed final upstream merge `fd6cd474d`, source ancestry, final five-owner
  scope, local bridge/fallback topology, and absence of later upstream owner
  changes.
- Ran focused discovery and x10 for five `apicompat` regressions and the
  default-tag service regression.
- Ran complete `internal/pkg/apicompat`, complete `internal/service`, server
  compile, gofmt, diff/scope/conflict/index/provenance, S242/S243 preservation,
  and protected-primary gates.

## Key output

- Five `apicompat` tests and one service test were discovered.
- Focused `apicompat` x10: PASS, `0.059s`; focused service x10: PASS, `0.082s`.
- Complete `apicompat`: PASS, `0.077s`; complete service: PASS, `64.866s`;
  server compile: PASS, `5.551s`.
- Malformed historical ordinary calls and matching outputs are skipped while a
  later valid call/output remains. Non-stream malformed calls are omitted.
- Streaming validation returns usage/result metadata and an error before final
  argument/output completion events or `[DONE]`; valid output-limit calls still
  finalize with an incomplete response.
- Business commit: `a8ce875c23c1e2bf4555ab46e376c7b92f9566a6`.

## Risks

- The repository-wide `-tags=unit` service suite has unrelated stale compile
  errors. After two attributed worker stops, the amended contract used a
  self-contained default-tag service regression and Controller takeover.
- No real provider, upload/download, database, container, browser, deployment,
  or push operation was exercised; all remain outside scope.

## Contract compliance

- Business commit contains exactly the five amended local owners. The original
  unit-tag fallback test and intermediate upstream `cc_pipeline` owner are
  unchanged.
- Custom tools, tool search, namespace behavior, S242/S243 compatibility, and
  empty-arguments normalization remain covered by complete package regression.
- The primary twenty-two-path patch ID and five untracked SHA-256 values remain
  unchanged; staged/unmerged indexes are empty.

## knowledge_candidates

- none; the unit-tag mismatch is recorded as Sprint-specific workflow evidence.
