---
type: task-contract
status: approved
review_verdict: PASS
task_id: model-unavailable-failover-20261030
worker_model: gpt-6.1-sol
base_commit: 8f0d51ba8910e80c1ed2a33acf828b895977a9a9
spec_ref: docs/workflow/spec.md
last_verified: 2026-10-30
---

# Task ID
model-unavailable-failover-20261030

## Role
Controller plans and integrates; independent Contract Reviewer reviews before development.
Developer implements in the isolated candidate; a separate QA validates the frozen result.

## Goal
Adapt upstream 5aeb03018 (Codex plan-gated account/model cooldown) and the relevant
missing-model failover behavior of 8363d537e to local owners. Also recognize the
reported HTTP 404 `Model 'gpt-6-luna' is not supported by ...` shape.
An unavailable account/model pair must not trap repeated requests on a sticky
account while another eligible account can serve the model.

## Context
- Primary: F:/mcplugins/sub2api.
- Candidate: E:/codex-worktrees/sub2api/model-unavailable-failover-20261030.
- Immutable baseline: the same path with `-baseline` suffix.
- Evidence: E:/codex-runtime/sub2api/model-unavailable-failover-20261030.
- Upstream pinned: d851a38477597157e78450d30af2d8504764d085.
- The nine existing dirty backend files are exact prerequisites, including
  go.mod/go.sum. Do not modify dependencies or repair unrelated tests.
- Use E:/tools/go1.26.10/go/bin/go.exe with GOTOOLCHAIN=local and
  GOROOT=E:/tools/go1.26.10/go.
- No real provider, application database, container, commit, push or deployment.

## Success Criteria
1. Recognize explicit structured model-not-found codes and model-unavailable
   messages in authoritative error fields for HTTP 400/404. Preserve explicit
   conflicting error codes, ordinary parameter/schema/context errors, endpoint
   404s and JSON that only echoes a matching phrase in request/debug metadata.
   Plain-text upstream errors may be inspected conservatively.
2. Codex plan-gated 400 applies only to OpenAI OAuth accounts. The generic
   missing-model behavior must stay within the OpenAI-compatible HTTP gateway;
   preserve other platform semantics.
3. Persist the existing 30-minute (account, effective upstream model) cooldown
   through SetModelRateLimit, using the existing mapping/scheduler key semantics.
   Other models and other accounts stay schedulable. Existing sticky selection
   must skip the rejected pair and be able to choose another eligible account.
4. Before client output is committed, a deterministic missing-model rejection
   must return the existing UpstreamFailoverError, never retry the same
   unavailable account/model pair. Cover Responses, Chat-to-Responses,
   Chat-native/compat and automatic passthrough owners through shared local
   error handling where possible. No replay after visible output.
   A 404 explicitly rejecting the model must not be mistaken for unknown
   Responses endpoint support and retried on the same account's Chat endpoint.
5. Preserve explicit custom error-code exclusions and pool mode's existing
   no-local-state policy. Request-local switching may still happen under that
   policy, but do not silently enable account-wide punishment or same-account
   retries. Document the lack of persistent cooldown for explicit opt-outs.
6. Preserve current image-driver versus image-model cooldown distinctions and
   avoid poisoning image availability from using a text endpoint incorrectly.
   Preserve request-scoped overload handling, error passthrough rules, existing
   retry budgets, billing and exactly-once release.
7. Failed cooldown persistence still allows the current request to switch.
   Exhaustion remains bounded by existing handler limits, and cannot signal
   business success, record usage or charge for rejected requests.

## Allowed Paths
- backend/internal/service/model_not_found_error.go
- backend/internal/service/ratelimit_service.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_gateway_chat_completions.go
- backend/internal/service/openai_request_scoped_retry.go
- backend/internal/service/openai_model_unavailable.go (new helper if needed)
- backend/internal/service/openai_model_unavailable_20261030_test.go (new)
- backend/internal/service/openai_model_unavailable_cooldown_20261030_test.go (new)
- backend/internal/service/openai_model_unavailable_http_20261030_test.go (new)
- backend/internal/handler/openai_model_unavailable_20261030_test.go (new, if needed)

## Denied Paths
All other product files, existing tests, dependencies, schema/migrations,
frontend, configuration, credentials, primary workspace and knowledge files.
Controller alone updates workflow documents and integrates approved exact bytes.

## Constraints
- Adapt behavior; never merge/cherry-pick upstream topology.
- One Developer owns all allowed product paths; QA writes only external probes
  and evidence. Contract Reviewer and QA remain independent of Developer.
- Preserve the frozen dirty prerequisites byte-for-byte.
- Do not change model mapping defaults, force all 400/404 to retry, or stop an
  entire account for one unavailable model.

## Acceptance Commands
From baseline/candidate backend, GOTOOLCHAIN=local:
- go test ./internal/service -run 'Test.*(ModelNotFound|ModelUnavailable|PlanGated|ImagesRejectedDriver|FailoverReparsesCachedBody|Passthrough.*Error)' -count=1
- The same command with -tags unit.
- go test ./internal/handler -run 'Test.*(ModelUnavailable|Failover|Retry|ChatCompletions)' -count=1
- The same command with -tags unit.
- go build ./...
- gofmt -l on changed Go files; git diff --check.
Add loopback/recorded-upstream fixtures proving the reported 404 and plan-gated
400 switch without writing client output, correct account/model cooldown,
sticky bypass, mapped keys, custom/pool policies, write failure, negatives and
the image guard. Test against the baseline to demonstrate the regression.
Compare broad failures to the immutable baseline instead of editing old tests.

## Output
Developer: external developer-result.md with changed paths, commands, risks.
QA: external qa-result.md, first line `### PASS/FAIL/BLOCKED: model-unavailable-failover-20261030`,
with independent command output, frozen source hashes and negative coverage.
Controller: workflow contract review and final QA report, exact integration
evidence showing original unrelated file bytes and Git index preserved.

## Stop Rules
Stop for allowlist expansion, dependency changes, existing-test edits,
unrelated baseline drift or a change that needs live provider/database access.
Return a concrete gap to the controller for a contract amendment.
