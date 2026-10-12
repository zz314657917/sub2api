---
status: approved
review_verdict: PASS
task_id: upstream-v0215-ws-oauth-20261009
worker_model: current-session-default
base_commit: ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b
spec_ref: docs/workflow/tasks/upstream-v0215-20261009-spec.md
---
# Task Contract
## Task ID
upstream-v0215-ws-oauth-20261009
## Role
Developer; only approved contract, disjoint ownership of WS/OAuth files.
## Goal
Adapt per-turn group billing snapshots and passthrough image input accounting; recover precisely rejected encrypted reasoning signatures; declare web search tools for OAuth history replay including Lite.
## Success Criteria
- Later WS turns re-read the key through existing auth cache; only refresh pricing group for same key ID, group ID, platform and subscription type; lookup errors or changed identity retain connection snapshot.
- Each turn's RecordUsage gets its own immutable group snapshot, including overlapping Before/After hooks, preserving local routing, subscription/trial, ledger idempotency and async context fields. No adopting upstream profit-gate architecture absent locally.
- Parse image input tokens from both Responses and Chat usage and managed image tool fallback; merge/cumulate/adapter propagation correct with no double counting.
- HTTP recovery for specific thinking_signature_invalid verification/decrypt text only; retry bounded to once and existing no-encrypted-content/no-replay safeguards retained. Do not treat all thinking errors as retryable.
- OAuth web_search_call history declares web_search in normal tools or Lite additional_tools as appropriate. Respect existing declarations/tool_choice, don't mutate API-key or compact requests, handle malformed/empty input safely.
- Unit and fake-local-upstream tests prove boundary cases and production build passes; preserve current GroupID-only pricing, image/cached-image pricing and model bridge controls.
## Context
Upstream: 73381fc9c, b6a157c19, a4ea8059b, dab3b87ea, 64caa9af8, 85e311095. Local gateway is monolithic: adapt into existing owners instead of importing split topology.
## Allowed Paths
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/*v0215_ws*test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_oauth_web_search_history.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay.go`
- `backend/internal/service/openai_ws_v2/*v0215*test.go`
- `backend/internal/service/*v0215_ws*test.go`
- `backend/internal/service/*v0215_oauth*test.go`
- `docs/workflow/worker-results/upstream-v0215-ws-oauth-20261009-result.md`
## Denied Paths
- Every other path, especially pricing/account ledger/schema, other worker protocol owners, knowledge, dependencies, config, migrations, frontend and original shared tree.
## Constraints
- Minimal behavior adaptation, no real providers, shared DB/Redis/container calls, commits/pushes or deployments.
- No replacing whole files. Don't amend baseline tests to hide failures. External compile-only fixture overlay may be used only with original errors and exact replacements documented.
- Preserve connection ownership and local first-response failover cancellation/output boundaries; refresh is pricing-only, not a new authorization/revocation policy.
## Acceptance Commands
From backend: `go test -tags unit ./internal/service/openai_ws_v2`; `go test -tags unit ./internal/handler ./internal/service -run V0215 -count=1`; `go build ./cmd/server`; `git diff --check`.
## Output
First-line DONE/FAILED/BLOCKED result for task ID, changed files, commands/exits/counts, runtime simulations, limitations, risks, baseline blockers, knowledge_candidates.
## Stop Rules
No approved review; model unavailable; missing handler architecture requiring pricing/security redesign; need denied files: stop and report. Don't silently broaden scope.
