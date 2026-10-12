---
status: approved
review_verdict: PASS
task_id: gpt61-context-subscription-20261012
worker_model: gpt-6.1-sol
base_commit: 167ea06927643856ea1354871761d82f721f07bd
spec_ref: docs/workflow/tasks/gpt61-context-subscription-20261012.md
---

# Task ID: gpt61-context-subscription-20261012

## Task ID
gpt61-context-subscription-20261012

## Role
Controller Planner/Final Evaluator; independent gpt-6.1-sol Contract Reviewer, Developer and QA.

## Goal
Behavior-port upstream 07185812655ec2e0740ca76882879c0b9ae5ee0a and 43768680d38ccdbd8fffc2cb91d3c1fc778a61c6, observed in upstream/main ab80abc6d1f929a6864c8804fd077b509d21fda2 on 2026-10-12. Adapt to the local monolithic gateway without importing the upstream split-file/catalog architecture.

## Success Criteria
- For API-key Codex manifests, fill missing/null context_window with 500000 only for recognized GPT-6.1 Sol model spellings. Preserve explicit smaller/larger windows, maximum window, compaction thresholds, unrelated model fields and unrelated models. The local branch lacks the upstream offline descriptor builder; do not retrofit it or claim an offline catalog feature.
- OAuth model manifests remain byte-for-byte upstream representations, with no hard override of account entitlements. Reuse existing ETag/cache handling for adapted API-key manifests.
- GPT-6.1 Sol upstream subscription credentials (OpenAI OAuth/setup-token, including an applicable shadow parent) are exempt from the model price card's long-context surcharge, even if the old account/group switches are enabled.
- Decide exemption from resolved upstream credential type and final sent model (UpstreamModel, falling back to Model), not customer billing aliases, response model, local Sub2API subscription billing type, or a model-name prefix. API-key credentials, other platforms and other models keep their existing behavior.
- Preserve administrator-defined interval/flat/per-request/image/video pricing and existing time/rate multipliers, cost overrides, deduplication and settlement. Do not delete API long-context prices, edit account/group flags, or change past usage.
- Apply the same exemption to account-statistics model-file pricing while preserving higher-priority custom account-stat rules and ApplyPricingToAccountStats behavior. Existing non-exempt call sites remain source-compatible.
- Default-tag tests exercise actual RecordUsage with fake repositories, usage-log fields, deducted balance and account-stat cost. Cover enabled/disabled gates, no resolver/resolver, 272000 boundaries, cache tokens, priority tier, mappings, shadow credentials, unknown similar model names and explicit administrator intervals.
- Tests fetch manifests through the real local service using httptest/stubs and verify OAuth pass-through, API-key missing/null/explicit fields, aliases, unrelated entries, and adapted ETags.

## Allowed Paths
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_codex_models_service_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/account_stats_pricing.go`
- `backend/internal/service/openai_subscription_long_context_test.go`
- `docs/workflow/worker-results/gpt61-context-subscription-20261012.md`
- `docs/workflow/qa-reports/gpt61-context-subscription-20261012.md`

## Denied Paths
All other business files, especially gateway_service.go, account.go, pricing catalogs, frontend, migrations, dependency/lock files, deployment files and shared knowledge/workflow files.

## Constraints
Implement only in E:/codex-worktrees/sub2api/gpt61-context-subscription-20261012 at the pinned base. No cherry-pick/merge/rebase, commit/push, real provider call, live account/database mutation, container action, global Codex configuration change, or browser. Controller alone updates contract/review/shared workflow and integrates accepted owner files after hash protection checks. Avoid signature churn: additive optional policy arguments or a small internal helper are preferred. Resolve shadow credential failures visibly; do not infer entitlement or silently grant exemption on a failed lookup.

## Acceptance Commands
- From backend: go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestOpenAIGatewayServiceRecordUsage|TestCalculateCostUnified_ModelLongContext' -count=1
- Repeat the new subscription/context tests with -count=10.
- From backend: go test -p 1 ./internal/pkg/openai -count=1
- From backend: go build -p 1 ./...
- Root: git diff --check and gofmt -l on all changed Go files.
- Baseline shared-tree service tests currently fail at gateway_beta_test.go:138 (undefined newTestGatewayServiceForBeta). First try canonical commands on the isolated committed baseline. Only if unrelated committed fixtures block compilation, retain exact first failure and use an external overlay replacing those unrelated test files; never replace changed files, production files, relevant tests or assertions.

## Output
Developer report begins ### DONE/BLOCKED/FAILED: gpt61-context-subscription-20261012, listing exact changed files, executed commands, scope compliance and evidence limitations. Independent QA report begins ### PASS/FAIL/BLOCKED: gpt61-context-subscription-20261012, with Findings, Executed Checks, Unverified Risks and Recommendation. Artifacts/evidence under E:/codex-runtime/gpt61-context-subscription-20261012.

## Stop Rules
Stop for unapproved paths, owner-file drift, unsupported required model, ambiguous credentials/model identity, changes to administrator-defined pricing, or inability to prove settlement/manifest behavior within the controlled test scope. Escalate findings to Controller rather than expanding scope.
