### DONE: upstream-main-ops-attribution-s9

## Summary

- Created isolated worktree `E:/codex-worktrees/sub2api/upstream-main-ops-attribution-s9` and branch `codex/upstream-main-ops-attribution-s9` from baseline `d6c7e4c69`.
- Added Sprint contract at `docs/workflow/tasks/upstream-main-ops-attribution-s9.md`.
- Ported all three approved Ops attribution candidates without directly merging `upstream/main`.
- Kept the implementation inside the approved backend service and workflow paths.
- Preserved local gateway, billing, stream hot path, frontend, Ent, and migration boundaries.

## Candidate Results

- `5d7df678b`: `CHERRY_PICKED` as `745852f7a`. Marks local OpenAI gateway policy and feature-gate denials as client business-limited instead of upstream/provider errors.
- `9c56fe0b0`: `CHERRY_PICKED` as `81daee02e`. Marks OpenAI fast-policy denials across chat-completions, raw chat-completions, messages, WS ingress, and WSv2 passthrough paths as client business-limited.
- `47fe90eab`: `CHERRY_PICKED` as `5b5c033b1`. Marks Antigravity Claude/Gemini whitelist denials as client business-limited.

## Applied Equivalent / Deferred

- `03ae510c6`: `APPLIED_EQUIVALENT`. Current baseline already excludes count_tokens from Ops metrics error counts and includes `ops_metrics_collector_test.go` coverage.
- `00eb3abbe` / `bd1e98ec2`: `APPLIED_EQUIVALENT`. Current baseline already marks API key and Google group denials business-limited.
- `a9c7a3a09`: `APPLIED_EQUIVALENT`. Current baseline already strips Bedrock `context_management` when beta is absent.
- `bf3787de1`: `APPLIED_EQUIVALENT`. Current baseline already allows Claude Code count_tokens by UA.
- `8a999f438`: `APPLIED_EQUIVALENT`. Current baseline already excludes terminal events from OpenAI WS first-token detection.
- `32ea9cfe1`: `APPLIED_EQUIVALENT`. Current baseline already falls back to SSE body for API key responses.
- `e9a2db8e8` / `8e27ff20a` / `86d9b6bff` / `1e2e8b1d6`: `DEFERRED`. Higher behavior blast radius; keep for a dedicated OpenAI stream/billing Sprint.

## Commits

- `2e6ccf455` docs: add ops attribution s9 contract
- `745852f7a` fix(openai): mark local gateway denials business-limited
- `81daee02e` fix(openai): mark fast-policy entrypoints business-limited
- `5b5c033b1` fix(antigravity): mark whitelist denials business-limited

## Changed Files

- `backend/internal/service/antigravity_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `docs/workflow/tasks/upstream-main-ops-attribution-s9.md`
- `docs/workflow/worker-results/upstream-main-ops-attribution-s9-result.md`
- `docs/workflow/qa-reports/upstream-main-ops-attribution-s9-qa.md`
- `docs/workflow/main-log.md`

## Verification

- `git status --short --branch`
- `git diff --check`
- denied path audit against `main...HEAD`
- source diff spot-check for changed gateway rejection paths
- `go test ./internal/service -run "Ops|BusinessLimited|FastPolicy|OpenAI|Antigravity|Whitelist|ImageGeneration|Passthrough" -count=1`
- `go test ./internal/handler ./internal/server/middleware ./internal/service -run "Ops|SLA|BusinessLimited|Denied|Gateway|OpenAI|Antigravity" -count=1`
- `go test ./internal/service ./internal/handler ./internal/server/middleware -count=1`

## Notes

- No candidate required forbidden paths, Ent schema, SQL migration, frontend changes, public API fields, billing recalculation, or server route wiring.
- The implementation only changes local rejection attribution from upstream/provider error paths to client business-limited markers.
