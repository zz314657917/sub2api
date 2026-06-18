---
task_id: upstream-main-v0137-safe-patches-s15
role: Generator
phase: contract-approved
qa_mode: runtime
owner: codex
upstream_base: e34ad2b19
upstream_target: 4a5665da5
created_at: 2026-06-16
---

# Task Contract: upstream-main-v0137-safe-patches-s15

## Goal

Port selected low-risk fixes from upstream `v0.1.137` into the local customized Sub2API branch without merging `upstream/main` wholesale.

## Success Criteria

- Preserve local Studio Bridge, LuoyeAI, configurable recharge packages, model market, Canvas, tickets, and routing customizations.
- Apply or equivalently port the safe patch set:
  - frontend `form-data` override security bump from `bbd970249`.
  - token refresh non-retryable error additions from `fa8f1749f` and `727ac3f68`.
  - zstd upstream response decompression from `c1c28ac7b`.
  - non-JSON 2xx failover and SSE `event:error` body preservation from `ab9987b2e` and `6c7203d83`.
  - API compat `tool strict` default false from `edfd5e373`.
  - fallback pricing additions from `c906bf000`, `a4ce73391`, `4f5f2788e`, and `262fe1230`.
  - reasoning/thinking compatibility from `142d8c361`, `6baf00d78`, `56c6325d1`, and `a05d9e87c`, if it can be ported without rewriting local gateway architecture.
- Do not change `backend/cmd/server/VERSION` merely to match upstream.
- Do not port product-gate or migration-heavy changes in this Sprint.

## Allowed Paths

- `backend/internal/service/**`
- `backend/internal/repository/http_upstream.go`
- `backend/internal/repository/decompress_response_test.go`
- `backend/internal/pkg/apicompat/**`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/handler/openai_images_failover_test.go`
- `frontend/package.json`
- `frontend/pnpm-lock.yaml`
- `docs/workflow/**`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/VERSION`
- `frontend/src/views/user/ChatImageStudioView.vue`
- `frontend/src/views/user/StudioBridgeSessionProbeView.vue`
- `frontend/src/views/user/PaymentView.vue`
- `frontend/src/views/admin/SettingsView.vue` except if strictly needed for listed safe patches
- `frontend/src/views/public/**`
- `knowledge/**` except task handoff files listed above
- Production deployment secrets or environment files.

## Constraints

- Prefer manual equivalent ports over raw `git merge` or large cherry-picks when upstream files diverge.
- Keep migration numbers untouched in this Sprint.
- If a patch depends on a larger upstream feature chain, document it as skipped instead of pulling the chain.
- Preserve local behavior for APIMart pricing, Studio Bridge billing, API key routing, and custom public pages.

## Acceptance Commands

- `cd backend && go test ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh" -count=1`
- `cd backend && go test ./internal/repository -run "Test.*Decompress|Test.*HTTPUpstream" -count=1`
- `cd backend && go test ./internal/pkg/apicompat -count=1`
- `cd frontend && npm.cmd run test:run -- --runInBand`
- `git diff --check`

Commands may be narrowed if the touched code has a more precise existing test target; any narrowing must be listed in the QA report.

## Output

- Implementation diff with short notes mapping each upstream commit to `ported`, `equivalent`, or `skipped`.
- Worker/result note at `docs/workflow/worker-results/upstream-main-v0137-safe-patches-s15-result.md`.
- QA note at `docs/workflow/qa-reports/upstream-main-v0137-safe-patches-s15-qa.md`.

## Stop Rules

- Stop and report if a safe patch requires Ent schema regeneration or new DB migrations.
- Stop before altering Studio Bridge, payment package, Canvas, tickets, or public page product flows.
- Stop before broad `upstream/main` merge, rebase, reset, or forced history rewrite.
