---
task_id: upstream-main-v0137-postfixes-s19
role: Generator
phase: contract-draft
qa_mode: runtime
owner: codex
upstream_base: v0.1.136
upstream_target: v0.1.137
created_at: 2026-06-18
---

# Task Contract: upstream-main-v0137-postfixes-s19

## Goal

Port the next small, independent backend fixes from upstream `v0.1.137` after S15-S17, while keeping the current S18 APIMart webhook work and local product customizations untouched.

## Success Criteria

- Preserve Studio Bridge / LuoyeAI, APIMart billing, payment packages, model market, Canvas, tickets, public pages, and current S18 webhook contract state.
- Port or equivalently implement these upstream fixes:
  - `46bd7968a fix: reuse OpenAI failover error body`.
  - `f6e0ebc6b fix: preserve Anthropic window cooldowns`.
  - `8b698ff4c fix account list parameter limit`.
- Evaluate `acaffe29e fix(account-repo): refresh candidates SQL excluded healthy accounts; fix CI build`; port only if the local branch has the corresponding `ListOAuthRefreshCandidates` SQL path. Otherwise record it as skipped/not applicable without pulling the broader token refresh retry amplification chain.
- Do not merge or rebase `upstream/main`.
- Do not change `backend/cmd/server/VERSION` merely to match upstream.
- Do not port image failover, token refresh retry amplification, OAuth promo signup, scheduler outbox dedup/cleanup, cyber policy, channel monitor jitter, Claude OAuth system prompt blocks, or migration-heavy chains in this Sprint.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_failover_cached_body_test.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/ratelimit_service_anthropic_window_limit_test.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_test.go`
- `backend/internal/repository/account_repo_temp_unsched_test.go`
- `backend/internal/server/api_contract_test.go`
- `docs/workflow/tasks/upstream-main-v0137-postfixes-s19.md`
- `docs/workflow/worker-results/upstream-main-v0137-postfixes-s19-result.md`
- `docs/workflow/qa-reports/upstream-main-v0137-postfixes-s19-qa.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/VERSION`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/service/wire.go`
- `frontend/**`
- Studio Bridge / Canvas / payment / public page / model market / ticket product files.
- APIMart webhook implementation files unless they are already listed in Allowed Paths.
- Production deployment secrets or environment files.

## Constraints

- Prefer manual equivalent ports over raw cherry-picks because local gateway, image, account repository, and scheduler code may differ from upstream.
- Keep S18 as the current workflow Sprint; S19 is only a drafted follow-up contract until explicitly reviewed and approved.
- If any selected patch requires new DB migrations, Ent regeneration, wire regeneration, frontend changes, or broader token refresh semantics, document it as skipped and stop before changing those paths.
- Preserve local behavior for APIMart pricing, Studio Bridge billing, API key routing, and custom public pages.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "Test.*Failover.*Body|Test.*Cached.*Body|Test.*Anthropic.*Window|Test.*Cooldown|TestOpenAI.*Images" -count=1
go test ./internal/repository -run "Test.*Account.*List|Test.*Refresh.*Candidate|Test.*Temp.*Unscheduled" -count=1
go test ./internal/server -run "Test.*APIContract" -count=1

cd F:/mcplugins/sub2api
git diff --check
git diff --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/VERSION|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|frontend/|backend/internal/service/studio_bridge|backend/internal/repository/studio_bridge|frontend/src/views/public/|frontend/src/views/payment/|frontend/src/views/canvas/|frontend/src/components/studio/|frontend/src/views/admin/ModelMarket|frontend/src/views/admin/Payment)" || echo NO_DENIED_PATHS
```

Commands may be narrowed or expanded if the local tests use different names; any change must be listed in the QA report.

## Output

- Implementation diff with short notes mapping each upstream commit to `ported`, `equivalent`, or `skipped`.
- Worker/result note at `docs/workflow/worker-results/upstream-main-v0137-postfixes-s19-result.md`.
- QA note at `docs/workflow/qa-reports/upstream-main-v0137-postfixes-s19-qa.md`.
- Result must explicitly list skipped candidates and why they remain out of scope.

## Stop Rules

- Stop before broad merge/rebase/reset of `upstream/main`.
- Stop before changing Ent schema, migrations, VERSION, wire generation, frontend, payment, Studio Bridge, Canvas, public pages, model market, tickets, or production secrets.
- Stop if a selected patch depends on scheduler outbox migrations, token refresh retry amplification, OpenAI image failover chain, or product behavior gates not approved in this contract.
