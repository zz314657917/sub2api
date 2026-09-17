---
type: task-contract
status: approved
review_verdict: PASS
task_id: upstream-v025-compat-fixes
worker_model: gpt-5.6-terra
base_commit: a022d9a3457457463ed06dd025370ad57936e668
spec_ref: docs/workflow/spec.md
---

# Task ID / Role
upstream-v025-compat-fixes. Independent Terra Developer implements only after independent contract PASS. Independent Terra QA validates afterward; controller makes final decision.

## Goal
Adapt four upstream compatibility fixes to the existing local topology. No upstream merge, rebase or cherry-pick. No push/deployment/migration/frontend changes. Preserve main worktree dirt.

## Success Criteria
1. bc8da7815 + 62635532a: Antigravity token keys use account ID; invalidation removes current account key and legacy project key. Preserve expiry, refresh and locking semantics.
2. 8447bdd36: Gemini SSE data frames have exactly one empty separator line; skip redundant empty upstream lines while preserving nonempty lines, heartbeat, cancellation and disconnect behavior.
3. 196c15b5c + e5272c130: remove only input-array object top-level internal_chat_message_metadata_passthrough for OAuth. Map and raw JSON paths work even without prompt/commands. Preserve same-named user content and nested fields, nonobjects, API-key behavior. Check HTTP and WS paths.
4. 087987070: integer sequence_number always serialized including zero. Compact synthesized frames count from zero; standalone errors/failures and fallback JSON include zero. Preserve upstream sequence values; do not renumber passthrough.

## Allowed Paths
All paths relative to isolated worktree E:/codex-worktrees/sub2api/upstream-v025-compat-fixes.
- backend/internal/service/antigravity_token_provider.go
- backend/internal/service/token_cache_invalidator.go
- backend/internal/service/token_cache_key_test.go
- backend/internal/service/token_cache_invalidator_test.go
- backend/internal/service/antigravity_gateway_service.go
- backend/internal/service/antigravity_gateway_service_test.go
- backend/internal/service/openai_responses_compatibility.go
- backend/internal/service/openai_oauth_input_metadata_test.go
- backend/internal/service/upstream_v025_compat_test.go
- backend/internal/service/openai_compact_stream_bridge.go
- backend/internal/service/openai_compact_stream_bridge_test.go
- backend/internal/service/openai_ws_http_bridge.go
- backend/internal/service/openai_ws_http_bridge_test.go
- backend/internal/handler/stream_error_event.go
- backend/internal/handler/stream_error_event_test.go
- backend/internal/pkg/apicompat/types.go
- backend/internal/pkg/apicompat/responses_stream_event_wire_test.go
- docs/workflow/worker-results/upstream-v025-compat-fixes-result.md

## Denied Paths / Constraints
Anything outside allowlist, main worktree, frontend, database migrations, production settings, knowledge and memories. Do not commit as developer. Four independent code commits and evidence commit are controller responsibilities. Existing unrelated test compilation failures must be reproduced on base rather than fixed opportunistically. No no-tests-to-run passes. No real provider or shared Redis. No public listener.

## Acceptance Commands
Run from backend:
```powershell
go test ./internal/pkg/apicompat -count=1
go test ./internal/service ./internal/handler -run 'Test.*(Antigravity|TokenCache|Metadata|Compact|WSHTTPBridge|ResponsesFailed|UpstreamV025)' -count=1
go test -tags unit ./internal/service ./internal/handler -run 'Test.*(Antigravity|TokenCache|Metadata|Compact|WSHTTPBridge|ResponsesFailed|UpstreamV025)' -count=1
go build ./...
```
If baseline test fixtures fail compilation, reproduce on immutable baseline and run isolated tests against actual production sources (no reimplementation/stubs of changed behavior). Document exact commands and exclusions. Missing effective production-source validation blocks acceptance.

## Runtime Acceptance
Dedicated task Redis actual set/get/delete: two accounts sharing a project remain distinct; invalidation removes current/legacy keys, no-project case works. Dedicated task-owned container permitted, no shared service changes; clean only task resources. If unavailable report BLOCKED.
Local httptest upstream exercises LF/CRLF multiple Gemini SSE events, completion, heartbeat/cancel/disconnect; strict frame parser checks no leading extra newline. Metadata coverage includes mixed/nonobject/empty input and nested/text same-name preservation. Sequence tests check zero, increasing compact sequence, standalone error/fallback and upstream preservation. Do not claim real provider acceptance.

## Output / Stop Rules
Developer report first line ### DONE/FAILED/BLOCKED: upstream-v025-compat-fixes, with changed paths, exact commands/results, risks. Independent QA report first line ### PASS/FAIL/BLOCKED: upstream-v025-compat-fixes. No silent model substitution. Stop on needed scope expansion or unavailable required validation; controller resolves. Two repeated worker failures require controller review.

## Review amendments
- Custom wireBase serialization must explicitly test zero sequence; default terminal events, compact sequences, handler and WS error/fallback JSON plus preserved nonzero upstream values are all mandatory.
- Local SSE owner handleGeminiStreamingResponse in antigravity_gateway_service.go must be tested with actual httptest upstream: LF and CRLF exact data frame bytes, no third newline, preserved nonempty non-data lines, heartbeat/cancel/disconnect.
- OAuth tests bind map transform (openai_codex_transform.go), raw HTTP owner (openai_gateway_service.go), and normalizeOpenAIResponsesWebSocketCompatibilityBody; input-only removal and API-key bypass are explicit. Existing call sites are read-only unless contract is amended.
- Untagged focused test TestUpstreamV025RedisRuntime must exercise production cache and invalidator with UPSTREAM_V025_REDIS_ADDR. With the variable set a connection failure is a failure, never skip.
- Dedicated Redis procedure (controller owns lifecycle): docker run -d --name sub2api-v025-compat-redis --label codex.task=upstream-v025-compat-fixes -p 127.0.0.1::6379 redis:7-alpine; check docker exec sub2api-v025-compat-redis redis-cli ping yields PONG; docker port sub2api-v025-compat-redis 6379/tcp returns assigned localhost port; set UPSTREAM_V025_REDIS_ADDR to that address, run go test ./internal/service -run '^TestUpstreamV025RedisRuntime$' -count=1 -v. Test owns unique key prefix/account IDs; no shared cache access. Finally verify container label and docker rm -f sub2api-v025-compat-redis. No persistent volume.
- Baseline failures must include exact reproducible commands/file set in report. Prefer normal untagged production package tests; isolated actual production source fallback only for baseline fixture blockers. Never replace changed production behavior with stubs.