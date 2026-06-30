---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-01 01:26 +08:00
---

# Task Contract

## Task ID
upstream-main-v0141-payment-validity-units-s33

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small upstream port. No external worker is used.

## Goal
Port upstream `147c1879d95f520b6db3af0291a52921322de421` so subscription plan validity units saved as `weeks` or `months` are converted correctly when creating payment orders.

## Success Criteria
- `psComputeValidityDays` treats `week` and `weeks` as 7-day units.
- `psComputeValidityDays` treats `month` and `months` as 30-day units.
- Existing `days` / unknown-unit fallback remains unchanged.
- A targeted regression test covers singular and plural unit values.
- No frontend, Ent, migration, wire, deploy, README, VERSION, user proxy/account, Ops, security redaction, or `knowledge/*` files are included.

## Context
- Repo: `F:/mcplugins/sub2api`
- Local anchor before S33: `8d1f94098 fix(gateway): return model not found for unsupported models`
- Upstream anchor after refresh: `v0.1.141-1-gdc1bc1545`
- Upstream reference:
  - `147c1879d95f520b6db3af0291a52921322de421 fix(payment): support plural subscription validity units`
- Local frontend already stores `weeks` / `months` from the plan editor, while backend payment order conversion only recognized singular `week` / `month`.

## Allowed Paths
- `backend/internal/service/payment_service.go`
- `backend/internal/service/payment_order_result_test.go`
- `docs/workflow/tasks/upstream-main-v0141-payment-validity-units-s33.md`
- `docs/workflow/worker-results/upstream-main-v0141-payment-validity-units-s33-result.md`
- `docs/workflow/qa-reports/upstream-main-v0141-payment-validity-units-s33-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `knowledge/**`
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/server/**`
- `backend/internal/service/(admin_service|openai_oauth_service|openai_oauth_service_state_test|proxy|proxy_service|user_account_service|user_proxy_service|wire).go`
- Ops classification, admin credential redaction, image billing metadata, Gemini chat-completions routing, channel-monitor, email notification, user proxy/account ownership work, production configuration paths, README, deploy, Dockerfile, and assets.

## Constraints
- Do not merge or rebase `upstream/main`.
- Do not change pricing, payment provider, order amount, subscription stacking, refund, or fulfillment semantics outside the validity-unit conversion.
- Keep the change backend-only and compatible with existing frontend values.
- Do not stage existing dirty Ent, migration, proxy/account, frontend, service, handler, or knowledge files unrelated to this task.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestComputeValidityDaysSupportsSingularAndPluralUnits" -count=1
cd F:/mcplugins/sub2api
git diff --check -- backend/internal/service/payment_service.go backend/internal/service/payment_order_result_test.go docs/workflow/tasks/upstream-main-v0141-payment-validity-units-s33.md docs/workflow/worker-results/upstream-main-v0141-payment-validity-units-s33-result.md docs/workflow/qa-reports/upstream-main-v0141-payment-validity-units-s33-qa.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0141-payment-validity-units-s33-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0141-payment-validity-units-s33-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementation requires frontend, Ent, migration, wire, route, repository, broader payment fulfillment/refund changes, user proxy/account work, or `knowledge/*`.
- Stop if tests cannot compile because of this patch's allowed paths.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`
