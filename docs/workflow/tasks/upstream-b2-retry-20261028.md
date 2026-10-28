---
type: task-contract
status: approved
review_verdict: PASS
task_id: upstream-b2-retry-20261028
worker_model: gpt-6.1-sol
base_commit: b3091c1c9a28884cf4b00fe97e87fc2f295b0512
spec_ref: docs/workflow/spec.md
last_verified: 2026-10-28
---

# Task Contract

## Task ID
upstream-b2-retry-20261028

## Role
Controller plans, freezes inputs and integrates. Independent gpt-6.1-sol Contract Reviewer reviews before implementation. One gpt-6.1-sol Developer implements the allowed slice. A separate gpt-6.1-sol QA repeats checks and probes the frozen candidate in another worktree.

## Goal
Continue the authorized selective upstream integration with B2 from c591412c42ab195dacdede9eb19f9cbfee221144, pinned within upstream f178e5ae32659e8901644a9f2378a0439ffa36ae. Adapt empty Responses stream boundaries and request-scoped overload Retry-After to current local owners. Keep local identity normalization settings, existing retry limits and billing ownership.

## Context
- Main F:/mcplugins/sub2api; candidate E:/codex-worktrees/sub2api/upstream-b2-retry-20261028; QA candidate suffix -qa.
- Evidence E:/codex-runtime/sub2api/upstream-b2-retry-20261028. task-tools.mjs setup freezes original main HEAD/index/file hashes and copies only current backend dirty Go/module prerequisites into a detached candidate. baseline-overlay preserves exact prerequisite bytes. Includes already accepted A1-A7 and user-owned backend changes; these are inputs, not this task's deliverables.
- Production currently marks any non-lifecycle event as output, uses existing overload3 retries at1/2/3 seconds, model-capacity5 retries, generic pre-output stream retry3 and configurable TUI normalization. Reuse these policies rather than introduce a new scheduler.
- Runtime acceptance uses production handlers/functions with in-memory repositories and localhost fake upstreams. No paid providers, real DB, container or deployed service changes.

## Success Criteria
### Stream replay boundaries
- Both Responses streaming processors (ordinary sync/ticker and passthrough) buffer well-formed, explicitly empty string text/reasoning/refusal/tool-argument deltas and empty text/refusal done or known content/summary part added/done events. Retryable error frames before response.failed are also buffered. Queue/lifecycle metadata continues existing buffering.
- Only proven empty known events may become replay-safe. Missing/null/numeric/object string fields, malformed JSON, unexpected part/item types, unknown events, whitespace text, tool completion events even with empty arguments/input, encrypted reasoning and nonempty content remain conservative output boundaries.
- Do not reclassify output_item.added function/custom-tool announcements as empty: these can expose tool IDs/names. Keep structural tool progress conservative. Do not change existing keepalive event classification or bounded pending-buffer limits.
- A subsequent overload/error/EOF before effective output can use the established failover policy. Discard failed-attempt buffered frames. Existing post-output error flow must not return a replayable failover. Retain successful terminal events, cancel handling, first-token timing, comment keepalives and body closure.
### Request-scoped overload and delay
- Add explicit RequestScopedTransient and SameAccountRetryDelay to the existing failover contract. The eligibility/precedence table below is normative, including indirect consumers of shared helpers.
- Preserve current overload3 and model-capacity5 retry counts; request-scoped flag does not grant extra retries, bypass account pin/eligibility/concurrency or change stream-retry wrapping.
- For request-scoped overload read first valid positive Retry-After (seconds or HTTP date), Retry-After-Ms, then error.headers/response.error.headers retry-after seconds/date and milliseconds values. Fixed delay is capped at8 seconds and overrides the existing linear base for all existing handler consumers. Missing/invalid/nonpositive/NaN/Inf/overflow inputs safely fall back to existing backoff, with large finite positive values clamped before time.Duration conversion.
- Preserve original upstream headers in each failover construction path. Parse dates against current time and test deterministically within a safe tolerance. Cancellation interrupts wait; acquired account slots are released before waiting and reacquired through the normal loop.
- Request-scoped overload must not write account error/cooldown/model transient failures, temp-unschedule after exhaustion, or penalize account scheduling success/error counters. Do not emit a synthetic success or clear pre-existing state. Request-local exclusion and bounded existing account switches may proceed. Keep diagnostics and all non-request-scoped error side effects.
### Eligibility and conflict precedence
| Dimension | Eligible | Excluded: preserve old behavior |
| --- | --- | --- |
| Account | PlatformOpenAI, API-key/OAuth/setup-token, ordinary or pool | nil/other platform/other account type, CN provider even when using OpenAI-compatible protocol |
| Actual upstream | OpenAI Responses or Chat HTTP; Messages inbound converted to OpenAI Responses; Images OpenAI HTTP | provider-native Anthropic /v1/messages; Gemini/Anthropic/Grok/CN upstream |
| Transport | Gin context has a real HTTP request and transport HTTP or unset; request is not WebSocket Upgrade | nil request/context, transport WS or WebSocket Upgrade. Existing WS relay and WS->HTTP bridge behavior stay old |
| Raw HTTP status | 400,500,502,503,529 only, already using existing failover branch | 401,402,403,404,408,409,413,429 and other statuses; do not make a new non-failover branch retryable |
| Stream error | Existing pre-output typed failover with authoritative overload, even when public status stays502 | visible-output/partial-result already committed; do not create new replay paths |
| Positive signal | Existing isOpenAIServerOverloadedError code server_is_overloaded/slow_down or exact accepted sentence in authoritative error fields/plain non-JSON error | generic5xx, selected-model-capacity, quoted overload text outside authoritative error fields |
| Higher-priority conflicts | None of the exclusions at right; inspect authoritative nested/root error codes/types/statuses before accepting overload | explicit401/403/429 or other noneligible HTTP status fields; authentication/credentials/access-state, quota/rate-limit/billing/permission, invalid/unsupported/model_not_found, context-window, content/policy/safety/cyber, or model-capacity semantics override an overload signal |

- For stream raw envelope status200, use existing semantic error fields to determine conflicts; final synthesized502 is not sufficient proof. Root event type response.failed/error is not itself an error-type conflict.
- A common eligibility predicate must be shared by flag creation and HTTP account-side-effect suppression. Preserve the generic handleOpenAIAccountUpstreamError default semantics for callers outside scope (particularly WS bridge); use a new HTTP-scoped wrapper or explicit opt-in at the allowed call sites. Do not add a broad unconditional skip to that generic owner.
- Apply metadata/header augmentation to the existing HTTP failover literals in openai_gateway_service.go, Chat→Responses, raw Chat, Responses→Chat fallback, OpenAI Messages and both image owners (including partial image-result errors); clone original response headers. Shared failoverOpenAIUpstreamHTTPError and newOpenAIStreamFailoverError must obey the same eligibility table. Existing stream usage/failure result policies remain untouched.
- Suppress schedule failure sampling in the four existing HTTP handler branches (Responses, OpenAI Messages, Chat, Images) only when the typed error flag is true. Shared FailoverState and GatewayService.TempUnscheduleRetryableError must also skip temp-unschedule for that flag, including status502 exhaustion. Never modify retry limits/counters/switch budgets or account-selection policy.
### Branch acceptance matrix
The two new test files may reuse existing fixtures but must not modify them. Cover the following production owners; table rows are required outcomes, not permission to refactor.
| Branch | Runtime assertion and dispatch/usage/resource expectations |
| --- | --- |
| Responses processors sync/ticker/passthrough | Empty known frames+retryable error+overload return typed pre-output failover and emit no failed-attempt semantic frames. Valid content/tool completion/unknown/malformed then failure never return replayable error; process one upstream body. All empty/malformed/unknown field cases use each processor mode |
| Responses full HTTP handler, ordinary and passthrough | One eligible account: overload then success gives IDs[1,1], two acquire/release/body-close events; three overload retries then exhaustion gives IDs[1,1,1,1], four of each, zero usage. Header wait occurs with no held account slot. No schedule error/success sample for failed overload, exactly the normal sample for success |
| Responses→Chat fallback, Chat→Responses, raw Chat, OpenAI Messages | Each real forwarding path must retain its raw HTTP headers, flag/delay and existing retry count. At least one real HTTP handler smoke for Chat and Messages proves overload then success follows the expected retry/exclusion budget with closed bodies/released slots and no failed-attempt usage; test both API-key and OAuth classification via table |
| Images API-key and OAuth/Responses bridge; split-image calls | Test real service HTTP construction for each owner; initial overload carries headers/flag/delay. When earlier images exist, preserve ImageCount/result and stop outer replay; internal successful subcalls are not reissued. A handler smoke checks one partial usage record and no restart of the completed image sequence |
| Shared failover loop | With request flag, consume initial+3 retries then bounded switch/exhaustion; no temp-unschedule or false health change. Original capacity5, generic stream3 and genuine429/auth behavior remain unchanged; do not combine their budgets |
| Content/tool-completion after output | Real Responses handler dispatches once and preserves its existing nil,error wrapper policy: it is NOT required to create partial usage. Separately retain raw-Chat adapter existing result/error partial-usage policy and prove one usage record with no second record |
| Cancel/acquire failure | Cancel during wait stops next dispatch; acquired slot/body closed once. Unacquired slot is not released. Test original pin and loss-of-eligibility behavior without bypassing selection. No usage after canceled pre-output failure or terminal exhaustion |
| Exclusions/health | Mixed auth/quota/client/policy/capacity signals never flag request overload; other platforms, native Anthropic and WS/bridge preserve old behavior. Spy real error/cooldown/runtime-block/model-transient/scheduler state: eligible overload is neutral, genuine429/auth retains positive side effects. Request-local exclusion still records bounded switching |

- For success/partial usage checks, install synchronous UsageRecordSubmitter or explicitly drain/wait for all submitted jobs; then assert exact record/charge call counts and payload tokens, including absence of a second record. Merely observing an async queue early is insufficient. Do not introduce new Responses partial-usage behavior: its existing Forward wrappers return nil,error after stream error, while raw-Chat has partial result support.
### Integrated evidence
- Add tests only in the two allowed new files; name all new declarations TestUpstreamB2*. Do not alter original test files/assertions, module files or dependencies.
- Reproduce empty-frame bug before fix through real processors. Cover the normative branch matrix above, Retry-After precedence/numeric boundaries and exact usage token/charge counts. No wall-clock8s sleeps.
- Run fixed original commands in task-tools.mjs check: service and handler relevant suites default/unit plus complete backend build, then every new declaration default/unit three repetitions. Audit declarations against executed test events; zero missing tests or skips. Existing baseline failures cannot be waived without exact independent baseline reproduction.
- Independent QA repeats immutable commands on frozen inputs and adds separate external probes for malformed/unknown output, header overflow, exhaustion side effects and real handler no-replay/usage boundaries.
- Freeze exact target and prerequisite hashes; QA seal binds every target and evidence file. Controller verifies main input/protected hashes unchanged, copies only exact approved targets after independent PASS, then reruns fixed checks and all new declarations on main's real combination. Preserve A1-A7 plus all original unrelated dirty bytes/HEAD/index.

## Allowed Paths
- backend/internal/service/gateway_service.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_images_responses.go
- backend/internal/service/openai_images.go
- backend/internal/service/openai_gateway_chat_completions.go
- backend/internal/service/openai_gateway_chat_completions_raw.go
- backend/internal/service/openai_gateway_responses_chat_fallback.go
- backend/internal/service/openai_gateway_messages.go
- backend/internal/service/openai_request_scoped_retry.go
- backend/internal/service/upstream_b2_retry_20261028_test.go
- backend/internal/handler/failover_loop.go
- backend/internal/handler/openai_gateway_handler.go
- backend/internal/handler/openai_chat_completions.go
- backend/internal/handler/openai_images.go
- backend/internal/handler/upstream_b2_retry_20261028_test.go
- Developer may write a result and new evidence only under the task's external evidence directory. Controller owns docs/workflow and knowledge/tasks/current-task.md; QA may add external probes using -overlay, without editing candidate product/test files.

## Denied Paths
All other business paths, existing tests, identity/version/settings code, frontend, go.mod/go.sum, dependencies, migrations, deployment/config files, knowledge/*, memories/*, protected prerequisite changes and previous task evidence. No staging, commit, push or production changes.

## Constraints
- Adapt behavior; no blanket merge/cherry-pick or upstream architecture replacement. Shared dirty gateway files must start from frozen main inputs.
- Keep original test semantics, billing prices/custom/zero/subscription policies, request/account usage ownership, identity normalization settings and websocket code unchanged.
- No wall-clock sleeps solely for tests; use tiny valid retry headers/cancellation for runtime checks and pure parsing tests for long delays.

## Acceptance Commands
From any directory, using node:
1. node E:/codex-runtime/sub2api/upstream-b2-retry-20261028/task-tools.mjs check <root> <unique-label>
2. node E:/codex-runtime/sub2api/upstream-b2-retry-20261028/task-tools.mjs check <root> <unique-new-label> new
3. node E:/codex-runtime/sub2api/upstream-b2-retry-20261028/task-tools.mjs scope <candidate-root> <unique-scope-label>
4. E:/tools/go1.26.10/go/bin/gofmt.exe -l <all changed Go paths>; git -c core.safecrlf=false diff --check -- <approved paths>
5. Controller/QA independently compare TestUpstreamB2 declarations with JSON run/pass events for both tags x3.

## Output
- Developer: external developer-report.md with first-line verdict, changed files, red/green evidence, executed commands, remaining risks and contract compliance; candidate-freeze.json binding all15 exact target files.
- Reviewer: docs/workflow/contract-reviews/upstream-b2-retry-20261028-review.md with PASS/FAIL, contract draft SHA256, findings and prerequisites.
- QA: external qa/qa-report.md with first-line verdict, Findings, Executed Checks, Unverified Risks, Recommendation; immutable qa/qa-seal.json with taskId/baseCommit/sourceFiles/prerequisites/evidence digests.
- Controller: final docs/workflow/qa-reports/upstream-b2-retry-20261028.md, refreshed status/current-task and main-log entries only when facts change.

## Stop Rules
Missing model or contract gate, scope drift, unexpected prerequisite/HEAD/index changes, new failure/skip/missing declaration or inability to prove no duplicate usage blocks integration. Preserve failures. Developer implementation bugs return to that Developer; oracle/contract errors return to their owner. No silent model substitutions or extra feature scope.
