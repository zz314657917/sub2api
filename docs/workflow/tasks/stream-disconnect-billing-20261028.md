---
type: task-contract
status: approved
review_verdict: PASS
task_id: stream-disconnect-billing-20261028
worker_model: gpt-6.1-sol
qa_worker_model: gpt-6.1-sol
base_commit: 1536c4849533460ab8c6ce81cf9e406c808e1cc4
spec_ref: docs/workflow/spec.md
last_verified: 2026-10-28
---

# Stream disconnect billing and finalization

## Task ID
stream-disconnect-billing-20261028

## Role
Controller is Planner/Final Evaluator. Independent Contract Reviewer, scoped Developers and separate QA use gpt-6.1-sol per repository Agent Matrix. A Developer must not self-approve contract or final acceptance. No model substitution.

## Goal
Selectively adapt upstream disconnect/interrupted-stream usage accounting to the current local architecture, preserving accepted A1-A7/B2 fixes, local multi-group first-response barriers, trial reservations and existing billing policies.

## Context
- Main: `F:/mcplugins/sub2api`, HEAD `1536c4849533460ab8c6ce81cf9e406c808e1cc4`; dirty backend is frozen, not discarded.
- Fixed upstream: `f178e5ae32659e8901644a9f2378a0439ffa36ae`.
- Behavioral references: `2b1f12938`, `94c7d6b02`, `4355236e0`, `bd52e5d77`; inspect with git show. Merge commits are provenance, not patches to replay. `f4f8ff04d` ops/499 classification is excluded.
- Candidate: `E:/codex-worktrees/sub2api/stream-disconnect-billing-20261028`.
- Immutable baseline: candidate path plus `-baseline`. Independent final QA uses separate `-qa`.
- Evidence/tools: `E:/codex-runtime/sub2api/stream-disconnect-billing-20261028`.
- Setup freezes every nonignored original file, HEAD/index and exact backend prerequisites, including existing dirty dependencies. No historical clean-HEAD substitute.

## Success Criteria
1. Gemini native `/v1beta`, Messages and Chat streams distinguish client abort from unrelated upstream failure. Client cancellation with observed tokens or delivered images returns collected billable result with ClientDisconnect; no observed usage/output must not invent a zero-usage bill. Keep Gemini upstream bound to client cancellation (no unbounded detached reads). Preserve malformed-frame and nonstream error behavior.
2. Gemini intermediate usage is cumulative, not additive; ignore non-EOF truncated read-error tails. Before upstream finish, only delivered native images count; Messages/Chat streams that do not deliver images count zero. Normal terminal completion keeps established image accounting.
3. Anthropic Chat/Responses compatibility streams stop client writes after disconnection, continue bounded upstream reading, merge message_start/message_delta usage and cache detail, and flag ClientDisconnect. Missing terminal, read error, configured read-gap timeout or in-band error must not fabricate successful terminal protocol output.
4. Anthropic native and API-key passthrough forwarders preserve nonzero observed partial usage on abnormal EOF/read error/timeout by returning result AND error; ordinary failure without observed usage returns nil result. Wrapped UpstreamFailoverError always returns nil result: retry attempts cannot be billed twice. Existing native/Bedrock ownership is inspected; a shared native stream path must inherit the same result semantics where applicable, without unrelated Bedrock refactoring.
5. Gateway Messages/Chat/Responses handlers submit result usage once on success or eligible partial error, and zero times for failover/no-result/no-usage. Preserve original/mapped model, response model metadata, request ID, duration, first token, cache details and local group/channel/pricing attribution.
6. OpenAI Responses and passthrough must continue bounded detached drain on client cancellation, including cancellation before client bytes or scanner context.Canceled. Once a valid terminal event was received, later transport errors cannot erase final usage or report false incomplete usage. A genuinely interrupted stream retains observed partial result plus error, not success and not failover replay. Unknown/malformed/output events remain B2 no-replay boundaries.
7. Usage worker submission distinguishes stopped pool from explicit drop/sample overflow. Handler billing falls back synchronously only for stopped-pool rejection or absent pool, with detached bounded context and panic handling; explicit drop/sample stays unchanged. Mandatory image billing falls back for either drop reason. All submit-mode consumers/trial flags must recognize both drop reasons, and trial/session/user/account/body release occurs exactly once, with accounting reservations held until billing runs or explicit drop is returned.
8. No accepted A1-A7/B2 behavior regresses. No duplicate records, upstream replay after visible output, goroutine/reader leaks or indefinite drains. Upstream/body closure and slot-release counters, missing usage, cancellation, normal completion, aborted image delivery and handler record-once boundaries are tested.
9. Independent QA verifies actual loopback HTTP/SSE runtime behavior (httptest server/connection abort), not only helper predicates, plus real handler/forwarder usage paths with recording doubles. Real provider/live DB/container acceptance is excluded and must be explicitly distinguished from loopback runtime evidence.
10. Migrate only the three existing `v0215_protocol_ping_test.go` cases whose old completion/accounting expectations conflict with criteria 3 and 8. Preserve immediate ping flush, comment-only encoding, no first-token evidence from pings, zero writes/flushes after a failed write, content/usage/finish-reason checks and exactly one DONE on normal completion. A ping-only EOF must be incomplete with no billable result or fabricated success terminal; a failed ping must still drain subsequent terminal usage (output7, prior input10 where present). A complete terminal ends the stream without requiring another upstream event or EOF. The revised normal-completion fixture must stop sending at its actual terminal; it must not require writes after the consumer closes. No production special-case for test-only nil Requests, no test skips, no edits to unrelated or baseline-failing assertions.

## Allowed Paths
The exact machine-readable list is `config.json.groups` under the external evidence root, generated from controller-authored task_tools.py. It has four disjoint write sets:

### Gemini Developer
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_stream_client_abort.go`
- `backend/internal/service/gemini_stream_client_abort_test.go`
- `backend/internal/service/stream_billing_gemini_20261028_test.go`

### Anthropic Developer
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/gateway_forward_compat_stream.go`
- `backend/internal/service/gateway_stream_partial_usage.go`
- `backend/internal/service/stream_billing_anthropic_20261028_test.go`
- `backend/internal/service/gateway_forward_compat_client_disconnect_test.go`
- `backend/internal/service/v0215_protocol_ping_test.go` (only the three cases described in criterion 10)

### OpenAI Developer
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/stream_billing_openai_20261028_test.go`

### Handler/Worker Developer
- `backend/internal/service/usage_record_worker_pool.go`
- `backend/internal/service/usage_record_worker_pool_test.go`
- `backend/internal/service/stream_billing_worker_20261028_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/stream_billing_usage_20261028_test.go`

Existing worker-pool test edits may only update stopped-rejection expectations to the new explicit stopped mode; never relax behavioral assertions. Existing OAuth passthrough tests retain require.NoError unchanged and must be fixed by production behavior, not by changing expectations.
The protocol ping test migration is the sole additional existing-test exception. All original bytes, the first combined FAIL, original freeze23 and original QA snapshot are retained under external `contract-reviews/pre-ping-addendum/` and `candidate/`. The newly allowed path's starting hash must equal its entry in main-before.json and the immutable baseline; no other original target record changes. The fixed command filters, tags, counts, failure comparator and original failure assertions remain unchanged.

Controller-only workflow files: this contract, its review/result/QA reports, docs/workflow/status.md, spec.md, main-log.md and knowledge/tasks/current-task.md. Worker evidence reports/scripts/probes go outside repository under their role folder.

## Denied Paths
- All production paths not explicitly listed, frontend/**, migrations/**, schema/Ent, go.mod/go.sum and lockfiles; these are frozen prerequisites, never implementation targets.
- Production/live config, credentials, secrets, shared databases, cloud tunnels, deployments, containers, git index/HEAD/history, all global memories and knowledge/** (except controller current-task).
- Existing tests outside allowlist; no skip/relax/rewrite of known failures.

## Constraints
- No merge, cherry-pick, reset, clean, blanket staging, commit, push or deployment. Edit candidate only, not main or immutable baseline. No new dependency or schema change.
- Minimal changes; do not overwrite unrelated dirty hunks within allowed monolithic owners. Compare against frozen current bytes, not only git HEAD.
- All new task-owned tests begin TestStreamBilling (upstream Gemini abort tests may retain descriptive names). Meaningful assertions must execute under default and unit tags; zero-run/skip is not PASS.
- Cross-worker ABI: existing ForwardResult/streamingResult contracts remain; Anthropic worker owns ClaudeUsage.hasObservedTokens helper if needed. Gemini may call it but not define a conflicting method; adding local Gemini-only predicate is also valid. Handler worker owns stopped-mode/Dropped helper and all handler consumers.
- Bounded line pump must cancel and close its task-owned response body so goroutines cannot outlive timeout/terminal exit. Never mutate Gin state from the read goroutine.
- Pump ABI is fixed: keep `openai_gateway_anthropic_native_pump.go` and all its existing callers unchanged. Anthropic Developer implements an independent, non-conflicting reader in allowed `gateway_forward_compat_stream.go`, not calls to upstream-only nextOrClientGone/enableInterval methods.
- Drain bounds for Anthropic native/passthrough/compat and OpenAI Responses/passthrough: while connected, preserve existing configured positive read-gap timeout (zero means disabled). Detect request cancellation even before first upstream byte. Once disconnected by request cancellation or failed client write, enable idle timeout equal to the configured positive gap or 180 seconds if disabled, and a 180-second total drain deadline; whichever expires first closes this attempt body and returns collected partial usage plus error. Terminal/in-band error/reader error exit also cancels reader and closes this attempt body; tests must prove goroutine exit. Deterministic tests may use task-owned helper duration parameters; production bounds remain as specified.
- Gemini delivered-image tests must cover zero-byte failed Write, partial Write, prior successfully delivered image followed by failed write, and disabling default-one image fallback on unfinished disconnect; no observer owner expansion is needed.
- Keep local reservation/failover/first-response timeout contracts: request context errors before an attempt is committed are not user billable; effective-output barrier forbids replay. Preserve account-model policy, explicit pricing and image mandatory recorder logic.
- No live SQL required: this task changes protocol/result/submission semantics, not repositories. QA may use actual recorder code with in-memory repositories, and localhost network; must not present this as live DB/provider validation.
- Baseline failures must be reproduced independently in immutable baseline with same command, named assertions and execution sets. Candidate may FIX known defects in scope, never add failures. Other known failures must match exactly and be named, not waived wholesale.

## Acceptance Commands
Run with `GOROOT=E:/tools/go1.26.10/go`, `GOTOOLCHAIN=local`, `GOMAXPROCS=4`; Go working directory is backend.
```powershell
python E:/codex-runtime/sub2api/stream-disconnect-billing-20261028/task_tools.py check E:/codex-worktrees/sub2api/stream-disconnect-billing-20261028-baseline baseline
python E:/codex-runtime/sub2api/stream-disconnect-billing-20261028/task_tools.py check E:/codex-worktrees/sub2api/stream-disconnect-billing-20261028 candidate
python E:/codex-runtime/sub2api/stream-disconnect-billing-20261028/task_tools.py check E:/codex-worktrees/sub2api/stream-disconnect-billing-20261028 candidate-new new
python E:/codex-runtime/sub2api/stream-disconnect-billing-20261028/verify_results.py compare baseline candidate
python E:/codex-runtime/sub2api/stream-disconnect-billing-20261028/verify_results.py new E:/codex-worktrees/sub2api/stream-disconnect-billing-20261028 candidate-new
```
The fixed broad filters are defined in task_tools.py SERVICE/HANDLER and include Gemini/Anthropic/OpenAI streaming, usage/worker, failover and B2. Both tags plus complete backend build are mandatory. New declarations must also be enumerated from changed test files and individually checked for count=3 and zero skip (supplement command if any name is not matched by the fixed new filter). Run complete apicompat protocol package under both tags, and A1-A7/B2 task tests under both tags. Run gofmt -l on all changed Go paths and git diff --check on allowlist; changed/new tests do not replace old tests.
Independent QA must add separate external contract-bound regression probes to a task-owned QA worktree/overlay without changing frozen final source. Report exact executed tests and baseline assertion comparisons. Stress timeout/drain/shutdown/release tests with count=10 and targeted Go race detector when Windows C compiler is available; otherwise explicitly report unavailable race instrumentation and still execute the deterministic concurrency stress tests.

## Output
- Reviewer canonical file: `F:/mcplugins/sub2api/docs/workflow/contract-reviews/stream-disconnect-billing-20261028-review.md`, first line `### PASS: stream-disconnect-billing-20261028` or FAIL/BLOCKED; bind exact contract/task_tools.py/config.json/verify_results.py hashes. External `contract-reviews/review-rN.md` is archival only; preserve every original failed round.
- Developers: external `<role>/report.md`, first line DONE/BLOCKED/FAILED; list changed paths, tests/commands, red-before/green-after evidence, risks, upstream adaptation differences and knowledge_candidates (do not write knowledge).
- Independent QA: `qa/report.md` first line PASS/FAIL/BLOCKED, and `qa/seal.json` with verdict, sourceFiles exactly equal candidate-freeze files, evidenceFiles relative to evidence root with SHA256. Only independently observed PASS permits integration.
- Controller reviews pure task delta, verifies seal and exact source/evidence hashes, preserves all main originals/HEAD/index, then copies only frozen changed targets and reruns fresh main combined checks. No product files are integrated on FAIL/BLOCKED.

## Stop Rules
- Model unavailable: BLOCKED, do not substitute. Scope/ABI uncertainty or required denied edit: stop that write set and request controller revision/review.
- Missing contract approval, immutable baseline, freeze or independent QA: no next gate.
- Repeated implementation failures: preserve failed logs, correct via original Developer, and re-test original failing cases by QA; contract defect returns to controller.
- Main/candidate/baseline/source/evidence drift: stop integration and reconcile exact changes without blanket exemptions or overwrites.
