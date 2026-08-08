# Task Contract

- Task ID: `upstream-openai-routing-hints-s206`
- Role: Generator and Evaluator (direct Codex implementation; no worker is delegated).
- Goal: Behaviorally adapt upstream `915cc7e7b`, `815035fcc`, and `de349187d` so OpenAI OAuth HTTP and WebSocket traffic carries a gateway-owned final-model routing hint, while API-key paths cannot forward caller/account supplied hints. Also apply upstream `8ad0a5ff5` to remove the audited `nanoid@3.3.16` dependency.

## Success Criteria

- Normal and passthrough OpenAI OAuth HTTP requests remove the legacy `OpenAI-Beta: responses=experimental` injection and synthesize `x-codex-routing-hint` only after local model/tier policy and account header overrides have completed.
- The hint uses the final upstream model and includes only canonical `priority` or `flex` tiers. Empty/unsafe models and caller/account supplied header variants fail closed; API-key paths never retain the hint.
- OpenAI OAuth WebSocket ingress and passthrough use the filtered first-turn model/tier and refresh the hint for later `response.create` turns without making routing affinity a hard continuation-compatibility requirement.
- The WebSocket pool prefers a matching routing hint when a compatible connection is available, falls back to compatible capacity when necessary, and does not publish stale prewarm/dial state after account-pool generation changes.
- `frontend/pnpm-lock.yaml` resolves `nanoid` from `3.3.16` to `3.3.17` with upstream-equivalent integrity metadata.
- Focused HTTP, routing-hint, WebSocket ingress/passthrough/pool, full service, server compile, dependency, formatting, scope and Git-integrity checks pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen base: `3cec8bb904bd880d5b2ef56daee85292e8cfc95a`
- Upstream head: `cc67b1aca1d3b590609abef2fcd3a6ca31c5c651`
- Upstream source: `915cc7e7b`, `815035fcc`, `de349187d`, and `8ad0a5ff5`.
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, `backend/internal/service/openai_gateway_service.go`, `backend/internal/service/openai_ws_forwarder.go`, `backend/internal/service/openai_ws_pool.go`, and `backend/internal/service/openai_ws_v2_passthrough_adapter.go`.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_routing_hint.go`
- `backend/internal/service/openai_routing_hint_test.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_ws_forwarder_success_test.go`
- `backend/internal/service/openai_ws_pool.go`
- `backend/internal/service/openai_ws_pool_test.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter_effort_test.go`
- `frontend/pnpm-lock.yaml`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/upstream-openai-routing-hints-s206.md`
- `docs/workflow/worker-results/upstream-openai-routing-hints-s206-result.md`
- `docs/workflow/qa-reports/upstream-openai-routing-hints-s206-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, `backend/go.mod`, `backend/go.sum`, `deploy/**`, `Dockerfile*`, and every frontend path except `frontend/pnpm-lock.yaml`.
- Any provider credential, shared database/Redis, container, deployment, production, remote push, or unrelated upstream-release change.

## Constraints

- Port behavior into the local monolithic HTTP/WS topology; do not replace local files with upstream split-file versions.
- Preserve current fixed-account routing, agent identity refresh, Fast/Flex policy, billing metadata, prompt-cache/session isolation, WS retry and proxy-circuit behavior.
- The routing hint is advisory and gateway-owned. Do not log its raw value or any authentication header.
- Keep the dependency change byte-equivalent to upstream `8ad0a5ff5`; do not regenerate unrelated lockfile entries.
- Do not fold other `v0.1.172` candidates into S206.

## Acceptance Commands

```powershell
git rev-parse HEAD
go test ./internal/service -run 'TestOpenAICodexRoutingHint|TestOpenAI.*RoutingHint|TestOpenAI.*LegacyBeta|TestOpenAIWS.*Routing|TestOpenAIWS.*Affinity|TestOpenAIWS.*Generation|TestOpenAIWS.*Prewarm' -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=0
gofmt -w <changed Go files>
rg -n 'nanoid@3.3.17|nanoid: 3.3.17' frontend/pnpm-lock.yaml
git diff --check
git diff --name-only 3cec8bb904bd880d5b2ef56daee85292e8cfc95a...HEAD
git diff --name-only --diff-filter=U
rg -n '^(<<<<<<<|=======|>>>>>>>)' <changed files>
```

## Output

- Write `docs/workflow/worker-results/upstream-openai-routing-hints-s206-result.md` with a first-line `### DONE`, `### BLOCKED`, or `### FAILED` verdict.
- Write `docs/workflow/qa-reports/upstream-openai-routing-hints-s206-qa.md` with an evidence-backed `PASS`, `FAIL`, or `BLOCKED` verdict.
- Record P/G/E transitions in `docs/workflow/main-log.md`; update handoff files only after final evidence exists.
- Commit only allowed files to the S206 branch. Do not push or deploy.

## Stop Rules

- Stop if the port requires schema/migration, configuration, provider credentials, production resources, or a non-lockfile frontend source change.
- Stop if routing affinity would have to become a hard WS continuation key or would regress local prompt-cache/session isolation.
- Stop if focused tests show fixed-account, agent-identity, Fast/Flex policy, billing, or WS retry behavior would regress.
