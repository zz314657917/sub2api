---
status: approved
review_verdict: PASS
task_id: upstream-v0216-gateway-20261015
worker_model: gpt-6.1-sol
base_commit: de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
spec_ref: docs/workflow/tasks/upstream-v0216-gateway-20261015.md
---

# Task Contract

## Task ID
upstream-v0216-gateway-20261015

## Role
This implementation contract is executed by an independent gpt-6.1-sol Developer after separate Contract Reviewer approval. A different gpt-6.1-sol QA independently validates the result; controller integrates exact validated paths.

## Goal
Behavior-port the small upstream gateway fixes f48866958 (OpenAI-only auto image bridge), ba2a71733 (Gemini tool-result images), 3bfd40f5a (missing explicit cache is a request error, not account failure). Re-run already integrated GPT-6.1 Sol context/subscription regression; do not reimplement those two policies.

## Context
- Shared main has unrelated gateway_service.go, beta/apicompat/admin/claude edits; none can be overwritten or imported into this batch.
- Work in a fresh isolated tree on stated base, separate from frontend candidate. Do not import upstream split gateway architecture, explicit cache management or billing reservation system.
- Execution prerequisite (2026-10-17 amendment): upstream-v0216-test-compat-20261015 must first receive independent QA PASS and a six-file seal. Controller copies only those exact sealed test files into E:/codex-worktrees/sub2api/upstream-v0216-gateway-20261015 on this base and verifies hashes. They remain read-only prerequisites, outside this Developer's writable allowlist. QA checks prerequisite hashes before/after; further test compilation gaps require a separate scope decision, never an overlay or silent fix.
- A fixture file absent in local baseline may be added if within allowlist. Developer must inspect original source diff, adapt to local API, and preserve existing error-policy priorities other than precise missing-cache request errors.
- Missing-cache owner is GeminiMessagesCompatService.ForwardNative and its helpers in gemini_messages_compat_service.go, including cachedContentNotFoundInLoop/isGeminiCachedContentNotFound to be adapted. Handler callers are outside scope; if request attribution needs a handler change, stop BLOCKED.

## Success Criteria
- Nil and non-OpenAI accounts never receive automatic hosted image_generation tool, tool_choice or injected instructions, despite global/channel enablement. OpenAI account/channel/global precedence remains unchanged; explicit user tools retain existing behavior.
- Gemini conversion preserves supported base64 and URL tool-result images, text, ordering, tool call identity, error markers and multiple results without duplicating/loss of parts.
- Precisely recognized missing-cache errors in Gemini native forwarding pass through as request-scoped errors before any disable/cooldown/retry/failover; unrelated credential 403 and overload responses retain existing policies. Normalize only as required by local API contract, no explicit cache resource CRUD.
- All production service packages build. Focused tests for these fixes and existing GPT-6.1 Sol policies execute on isolated baseline without hiding failures.
- Independent QA executes checks and seals exact changed-file hashes before controller integration; no unvalidated fixture overlays in canonical candidate.

## Allowed Paths
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_image_generation_controls_test.go
- backend/internal/service/gemini_messages_compat_service.go
- backend/internal/service/gemini_messages_compat_service_test.go
- backend/internal/service/gemini_cached_content_not_found_test.go
- backend/internal/service/gemini_error_policy_test.go
- docs/workflow/worker-results/upstream-v0216-gateway-20261015.md
- docs/workflow/qa-reports/upstream-v0216-gateway-20261015.md

## Denied Paths
All other source/test files including main dirty files; backend/internal/handler, schemas/migrations, all billing/ledger code, account status APIs, package manifests, dependency locks, shared knowledge/memories, container/production/configuration. No edits to GPT-6.1 Sol existing implementation/tests.

## Constraints
No new git commits/staging/push/merge/rebase/cherry-pick. No real providers, database or deployment. Distinguish tests/build evidence from runtime acceptance. External logs/artifacts only under E:/codex-runtime/upstream-v0216-20261015/.

## Acceptance Commands
From candidate backend:
```powershell
go test -p 1 ./internal/service -run 'TestOpenAIGatewayService.*Codex|TestGemini.*(Tool|Image|Cache|Error)|TestConvertClaude|TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestCalculateCostUnified_ModelLongContext' -count=1
go test -p 1 ./internal/pkg/apicompat ./internal/pkg/openai -count=1
go test -p 1 -tags unit ./internal/service -run 'TestShouldFailoverGeminiUpstreamError|TestGemini.*ErrorPolicy|TestGemini.*CachedContent' -count=1
go build -p 1 ./...
```
Discover exact new test names and run each explicitly with -v (or count JSON pass events), require nonzero executions/no skips. The tagged command must execute the tests in gemini_error_policy_test.go; discover any names not covered by the example expression and run them explicitly. Repeat new focused cases if needed for flake risk. Run scoped gofmt/diff checks; baseline failures must remain visible, never claim zero matching tests as PASS.

## Output
Developer DONE/FAILED/BLOCKED and independent QA PASS/FAIL/BLOCKED reports with actual commands, changed paths, risks, evidence and hash manifest.

## Stop Rules
If model unavailable, denied changes needed (including handler request-scoped attribution), or canonical baseline tests cannot compile, stop and report exact dependency/proof gap for controller decision. Do not expand architecture or weaken assertions.
