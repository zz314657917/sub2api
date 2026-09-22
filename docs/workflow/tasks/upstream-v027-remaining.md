---
status: draft
review_verdict: pending
task_id: upstream-v027-remaining
worker_model: gpt-5.6-terra
base_commit: d6e6c34718e6eee6388391b346f40e1492e81d57
spec_ref: docs/workflow/tasks/upstream-v027-remaining.md
---

# Task ID: upstream-v027-remaining

## Role
Planner/controller owns scope and final evaluation. Independent Terra contract reviewer, Developer and QA are separate agents.

## Goal
Continue user-approved v0.2.7 selective behavior ports in this isolated worktree. Preserve local monolithic gateway and customized behavior. Prior Anthropic schema port is baseline and must remain intact.

## Success Criteria
1. DeepSeek native Responses tool images lifted into following user message, parallel tool results remain contiguous (881ab1b0c + acc05620c); non-DeepSeek/non-Responses unaffected. Preserve numeric precision and store=false. Cover real localhost upstream wire payload and plain/multimodal/multiple-call cases.
2. Antigravity Gemini thinkingConfig bare-model mapping and SDK SSE comment compatibility (0f4d8acaa, f79b8bf96), plus account-mapped native model listing (da74bf13a), adapted without erasing local model-list and routing permissions. Use fake HTTP runtime tests.
3. Strict-provider developer role normalization (18bfa4bf2), manifest parse/validation pair (2f16e0984 + 31f3003ff), cancellation-safe bounded affinity writes (d7ee1ab6b), using local equivalents and request ownership. Add focused regression/runtime tests.
4. CN Coding Plan quota-exhausted 403 temporary cooldown (db8692d67), preserving ordinary 403, concurrency limits, per-model circuit breaker and reset semantics. Add table-driven classification and persistence-failure tests.
5. Independently evaluate low-risk UI fixes: subscriptions clear loading 406be7c518, config await b51f0759d4, dialog IDs be313735dc, refund warning 017e9e98d1, promo flicker fe36f4a914, model input Tab f8e5a0d94f, negative quota 0838e0e610, payment amount 98321a054e, TOTP sync ee9ac3e45f/errors 14636d2f0e, pagination 1a32b91eb2, clipboard 130ba634ad, announcements 6a4938bdfb, proxy test guards 0ed735d3ba, order page 21532add46. Port only missing applicable behavior; document equivalent/deferred cases with evidence. UI changes need focused Vitest/typecheck/build and runtime/browser evidence before full PASS.

## Allowed Paths
Backend: backend/internal/pkg/apicompat/responses_tool_output_media*.go; backend/internal/service/openai_gateway_service.go and relevant new scoped *_v027_test.go; backend/internal/service/antigravity_gateway_gemini.go; backend/internal/service/antigravity_gateway_streaming.go; new Gemini thinking/SSE helpers/tests; backend/internal/handler/gemini_v1beta_handler.go and related tests; backend/internal/service/gemini_messages_compat_service.go; backend/internal/service/openai_gateway_chat_completions_raw.go; new openai_chat_roles.go/tests; backend/internal/service/openai_codex_models_service.go/tests; backend/internal/service/ratelimit_service.go and CN quota helper/tests. Existing local owner files may be substituted only after controller approval recorded in review.
Frontend: only production paths and related test files touched by the listed UI commits; no global style/locale/router rewrites.
Evidence: docs/workflow/tasks/upstream-v027-*.md, docs/workflow/contract-reviews/upstream-v027-*.md, docs/workflow/worker-results/upstream-v027-*.md, docs/workflow/qa-reports/upstream-v027-*.md. Controller owns isolated docs/workflow/status.md and main-log.md.

## Denied Paths
Main F:/mcplugins/sub2api workspace; migrations, dependency manifests/locks, plugin system, Seedance, shared containers/DBs, secrets, global memory. No commit/push/merge/deployment in this phase.

## Constraints
Behavior adaptations only; no cherry-pick/whole-file upstream overwrite. Preserve existing changes. No paid provider calls. Do not call a clean patch application validation. Independent development/QA using gpt-5.6-terra. Task-owned isolated browser profile and exact cleanup if used.

## Acceptance Commands
From backend: go test ./internal/pkg/apicompat -count=1; focused service/handler test selectors exercising named scenarios; go build ./.... Distinguish known baseline failures using immutable base comparison; no silent skipped tests. Frontend from frontend: pnpm exec vitest run <changed tests>; pnpm exec vue-tsc -b; pnpm build (use available native Windows entry point and isolated output). git diff HEAD --check. QA independently reruns focused tests and checks actual runtime payloads.

## Output
Worker report per batch with first-line PASS/FAIL/BLOCKED, paths, commands/exit codes, upstream-to-local mapping and remaining limitations. Independent QA report and overall coverage table for every listed candidate.

## Stop Rules
Missing dependencies or topology requiring out-of-allowlist change: report to controller first. No silent model fallback. Unavailable Terra means BLOCKED. Test failures must be repaired/retested or specifically documented; full acceptance requires runtime evidence. Keep useful partial work and reports without staging/committing.
