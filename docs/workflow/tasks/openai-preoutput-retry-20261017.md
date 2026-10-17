---
status: approved
review_verdict: PASS
task_id: openai-preoutput-retry-20261017
worker_model: gpt-6.1-sol
base_commit: de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
spec_ref: docs/workflow/tasks/openai-preoutput-retry-20261017.md
---

# Task Contract

## Task ID
openai-preoutput-retry-20261017

## Role
Independent Contract Reviewer, Developer and QA per current Agent Matrix. Controller owns scope, integration and final verdict.

## Goal
Reduce user-visible `stream disconnected before completion: Upstream request failed` failures by giving safe OpenAI pre-output transient stream failures up to three additional attempts on an available account before existing failover. Preserve all post-output replay protection.

## Context
- User authorized optimizing the previously investigated stream failure. No commit, push, provider call or deployment is authorized.
- Shared main contains unrelated changes including gateway_service.go; preserve every unrelated byte. Work in the new task-owned managed worktree at stated base.
- Existing stream code buffers response.created/in_progress, returns UpstreamFailoverError before semantic output, but generic disconnect/response.failed lacks same-account retry eligibility. Handler already supports finite per-account retry counts and cancellation-aware delay.
- Existing service default-tag tests compile independently of known stale `-tags unit` fixtures. Default-tag tests and production build are canonical for this bounded task. Run a tagged compile probe and retain any pre-existing errors; do not repair, overlay, hide or claim passing tagged tests. This is independent of upstream-v0216-test-compat and does not change its state.

## Success Criteria
- Before semantic output, connection read failure, missing terminal event and generic transient response.failed (including Upstream request failed) receive up to three additional retry opportunities even on a non-pool single-account group, with cancellation-aware waits of 500ms/1s/1.5s (three seconds total added backoff per request). New generic retries share a request-wide budget of three, even if accounts change. Existing account-switch budget remains finite.
- The new generic policy applies ONLY to non-pool accounts. Pool accounts retain existing count/status-list decisions: count=0, explicitly empty status list, missing default list and custom list; capacity/overloaded/OAuth429 special cases retain their previous priority. Test eligibility, limit and actual dispatch counts for these combinations.
- Preserve explicit capacity, overloaded and OAuth429 policies, including limits, headers and retry-after behavior; do not silently change configured pool-mode retry semantics.
- Do not retry an already committed text/reasoning/tool/image result. Preamble-only events and keepalive comments must not consume the replay boundary. Cancellation/deadline, oversized SSE lines, policy refusals, context-length errors and explicit non-transient client/auth/account-state errors do not acquire the new generic same-account policy. Existing account failover for credentials remains unchanged.
- Do not broaden generic HTTP 403/409/413/502 handling; this change is specifically the pre-output stream path. Bare generic terminal error events retain their conservative classifier.
- Prove no intermediate error or buffered failed-attempt preamble reaches the client before a successful retry. Prove three retries means at most four attempts before same-account exhaustion, bounded existing failover afterward, and no retry after output or cancellation.
- Scope: ordinary and passthrough Responses streaming only. Do not add generic retry eligibility to shared buffered/non-streaming or Anthropic adapter call sites. Messages may reuse existing error handling but no claim of new adapter read-error retry; regress existing Anthropic and non-streaming behavior.
- Request-local retry target must re-enter the existing validated pinned scheduler path; never reuse an old account snapshot or override an original request pin/private-first restriction. If eligibility is lost, preserve the last stream error, exclude the failed target, consume the existing switch budget and permit finite fallback. Normal account slot acquisition, queue, wait and timeout behavior still applies. Release the prior slot before sleeping, never hold it across backoff. Busy-slot or wait failure follows existing error behavior and may end early. No resetting retry/switch counters.
- Handler smoke must exercise actual single-account four-attempt exhaustion, A-fails/A-succeeds despite another account being available, A-ineligible-then-B, slot acquire/wait failure, cancellation during backoff/wait and exactly-once releases of account/user slots and waiting counters. Reuse existing stubs/local upstream, no provider.
- Buffer lifecycle/preamble separately from the 4KiB output buffer until semantic output; only independently tracked keepalive comments may write early. Bound the pending buffer (an existing configured SSE size cap is acceptable); overflow must terminate as too large without acquiring new generic retry eligibility. Cover synchronous/ticker and passthrough processors, accumulated preamble larger than 4KiB, keepalive then failure then success. Assert only the successful attempt's lifecycle IDs reach the client. If a prior write cannot be proven comment-only, conservatively forbid replay. When cancellation and EOF coincide, cancellation wins and no additional upstream call occurs.
- Existing switch cap remains unchanged (default10). For generic failures only, at most eleven initial/account-switch attempts plus three added retries (fourteen attempts total) and three seconds of added backoff. Existing capacity/overload policies remain separate. Request/route cancellation or account availability may terminate earlier. Existing upstream and route timeouts still govern network time; do not claim a new total-duration bound. Distinguish the new stream policy using structured task-local error metadata/wrapper, never by matching human error text or inferring policy from magic delay values.
- Scoped formatting/diff checks, nonzero targeted tests and full production build pass. Runtime smoke uses only a task-owned httptest server or local in-process upstream stub, no real provider.

## Allowed Paths
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_gateway_service_test.go
- backend/internal/service/openai_stream_retry.go
- backend/internal/service/openai_stream_retry_test.go
- backend/internal/handler/openai_gateway_handler.go
- backend/internal/handler/openai_stream_retry_test.go
- docs/workflow/worker-results/openai-preoutput-retry-20261017.md
- docs/workflow/qa-reports/openai-preoutput-retry-20261017.md

## Denied Paths
All other business files, stale tagged fixtures, other workflow tasks, frontend, dependencies/locks, migrations, configuration, containers, live databases/providers, shared knowledge and memories. No staging/commits/push/reset/rebase/cherry-pick.

## Constraints
Keep the smallest implementation using current retry infrastructure; no new dependency or administrator setting. Preserve shared main and index. Do not import unrelated fixes. Developer and QA must be separate agents. Logs and supplemental artifacts belong under E:/codex-runtime/openai-preoutput-retry-20261017/.

## Acceptance Commands
From isolated backend, using exact test names discovered in this task when necessary:
```powershell
go test -p 1 ./internal/service -run 'TestOpenAI(Stream|Streaming|Passthrough|NonStreaming)|TestNonStreamingTerminalFailureFailover|TestOpenAI.*Retry' -count=1
go test -p 1 ./internal/service -run 'TestForwardAsAnthropic_|TestHandleAnthropic|TestNonStreamingTerminalFailureFailover' -count=1
go test -p 1 ./internal/handler -run 'Test.*(StreamRetry|PreOutput|SameAccount|Failover|Streaming)' -count=1
go test -p 1 ./internal/service ./internal/handler -run 'TestOpenAIStreamRetry|TestOpenAIPreOutputRetry' -count=5
go build -p 1 ./...
git diff --check
```
First run a `-tags unit` no-test compile probe to establish baseline status and retain failures separately. If canonical default-tag builds cannot run without denied edits, stop. Do not report `no tests to run` or filtered/file-list builds as acceptance.

## Output
Developer report starts `### DONE: openai-preoutput-retry-20261017` (or FAILED/BLOCKED), lists exact paths, commands, nonzero counts, boundaries, risks. Independent QA report starts PASS/FAIL/BLOCKED and includes dispatch smoke evidence, denied-path protection, exact accepted file SHA256 map. Controller copies only accepted files into main after drift checks.

## Stop Rules
Stop on unavailable configured model, denied edits, new billing/auth/scheduler architecture needs, unexpected main-owner drift or missing nonzero default-tag coverage. Report stale tagged fixture failures without expanding scope. Never weaken existing assertions to force success; any changed expected retry eligibility must be justified by this contract.
