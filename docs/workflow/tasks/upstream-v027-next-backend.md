---
status: approved
review_verdict: PASS
task_id: upstream-v027-next-backend
worker_model: gpt-5.6-terra
base_commit: 38a5a6040faa065a795c2ee81c5b11e12902164c
spec_ref: docs/workflow/tasks/upstream-v027-next-backend.md
---
## Task ID
upstream-v027-next-backend
## Role
Independent Terra contract reviewer, Terra developer, separate Terra QA; controller integrates.
## Goal
Behavior-port bcc73f8d4 DeepSeek reasoning placeholders and 8e34ca5e3 paused OAuth refresh using actual local owners. No upstream split-file imports.
## Success Criteria
Responses-to-Chat fallback sends single-space reasoning_content for missing/empty assistant reasoning only for DeepSeek platform or exact api.deepseek.com host. Preserve nonempty reasoning, non-assistant messages, original input and other providers. Apply after local Responses conversion and before HTTP request creation; do not expand to unrelated native Chat paths. Test actual forwardResponsesViaRawChatCompletions against fake/local HTTP upstream for encrypted-only and plaintext history, streaming and nonstreaming, and strict host boundaries.
Local TokenRefreshService.listActiveAccounts must retain active accounts even when administrator set Schedulable=false; ListActive already owns status filtering. Refresh must not set schedulable=true or clear administrator pause. Preserve existing invalid-grant/error handling, retries, temporary cooldown and disabled/error account exclusion. Trace postRefreshActions and refresh loop before implementing; if local owner would unpause an account, report necessary scope amendment. Test actual candidate selection and refresh/persistence with fake repo/refresher, including paused account flag unchanged, active enabled, repository error and non-OAuth exclusion at existing loop boundary.
## Allowed Paths
backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go
backend/internal/pkg/apicompat/deepseek_reasoning_v027_test.go
backend/internal/service/openai_gateway_responses_chat_fallback.go
backend/internal/service/openai_gateway_deepseek_reasoning_v027_test.go
backend/internal/service/token_refresh_service.go
backend/internal/service/token_refresh_service_test.go
backend/internal/service/token_refresh_paused_v027_test.go
docs/workflow/worker-results/upstream-v027-next-backend.md
docs/workflow/qa-reports/upstream-v027-next-backend.md
## Denied Paths
Everything else: main checkout, frontend, repositories, dependencies/locks, migrations, databases/containers, unrelated dirty files, temporary patch cleanup. Workers must not commit/push/deploy.
## Constraints
Minimal behavior adaptation. No cherry-pick, patch import or whole-file replacement. Preserve current gateway schema/media/role fixes. Tests use local fakes only, no real account/network credential refresh. Controller may commit/integrate after independent QA; no production deployment.
## Acceptance Commands
From backend: go test ./internal/service -run 'TestV027(DeepSeekReasoning|PausedRefresh)' -count=1 -v; go test -tags unit ./internal/service -run 'TestForwardResponses_ForceChatCompletions|TestTokenRefreshService|TestV027' -count=1; go build ./.... Confirm tests actually run; include unit tag when necessary. Run git diff HEAD --check, gofmt diff only allowed Go files and empty unmerged-path query. Reviewer checks added test coverage, exact hostname matching and preserved plaintext.
## Output
First line ### PASS/FAIL/BLOCKED: upstream-v027-next-backend; files, commands, runtime payload/state evidence, scope and unverified real-provider limits.
## Stop Rules
Missing APIs, contract ambiguity or out-of-allowlist change: report to controller before edits. Terra unavailable: BLOCKED, no silent model substitution. Preserve unrelated work. Do not claim real provider validation from fakes.

## Local owner clarification
The DeepSeek injection owner is exactly forwardResponsesViaRawChatCompletions in openai_gateway_responses_chat_fallback.go, after chatBody serialization and before http.NewRequestWithContext. No sendCCUpstreamRequest or new split pipeline file is needed. Existing code lacking the requested behavior is the reason for this task, not a contract defect. Token refresh changes remove the Schedulable filter in listActiveAccounts; postRefreshActions remains unchanged and must not enable scheduling. Regression tests must prove these boundaries.

## Plaintext adaptation clarification
The local shared Responses-to-Chat bridge does not currently retain input reasoning summaries. The fallback owner may restore plaintext for DeepSeek only before serialization, then fill remaining empty assistant fields. No cache subsystem is added. Source-to-output alignment must handle merged parallel calls, skipped invalid calls, media-generated user messages and distinct reasoning across turns without assigning one turn's reasoning to another. If implementing this safely requires changing the shared bridge, obtain an explicit path amendment and independent contract review first.

## Opt-in bridge amendment
DeepSeek implementation waits for independent review of this amendment. Add an explicitly named opt-in conversion entry in the existing bridge; its default ResponsesToChatCompletionsRequest behavior is unchanged. Only DeepSeek fallback chooses opt-in. In the existing single-pass input conversion, consume reasoning items instead of turning them into empty user messages, accumulate plaintext summary_text/reasoning_text for the following assistant turn, preserve explicit nonempty assistant reasoning_content, and carry the reasoning with the assistant through final tool/media normalization. Encrypted-only items contribute no plaintext and are never forwarded as reasoning text. User/tool-output boundaries reset pending reasoning to prevent cross-turn leaks; tests must cover consecutive reasoning, distinct multiple turns, parallel calls, skipped invalid calls, missing replies and media insertion. No encrypted cache or new service is introduced. Default conversion regression tests must demonstrate unchanged non-target semantics. Also run go test ./internal/pkg/apicompat -count=1.
