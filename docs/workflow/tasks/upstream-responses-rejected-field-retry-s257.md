# Task Contract

## Task ID

upstream-responses-rejected-field-retry-s257

## Role

Developer Worker (`gpt-5.6-terra`) implements this approved, isolated
Responses compatibility slice. Codex reviews the diff before a separately
created `gpt-5.6-terra` QA Worker reruns every gate.

## Goal

Behaviorally adapt the reusable rejected-field retry portion of upstream
`5e4da92d`, `cf3577a3`, and `e440ac48c` to the local monolithic
`OpenAIGatewayService` HTTP `/responses` forward loop. In particular, a 400
that explicitly rejects `input[n].status` must remove `status` from every
same-type input item in one bounded retry, without importing the unrelated
41-file Responses refactor.

## Success Criteria

- Add a request-scoped, cross-failover bounded retry budget of exactly six
  transformed bodies; each account attempt also deduplicates bodies it has
  already sent. It must never retry an empty, unchanged, or previously seen
  body.
- Apply it only after an HTTP 400 explicitly identifies an unsupported or
  unknown Responses field. Preserve the existing error/failover behavior for
  all other status codes and error payloads.
- Port the compatibility transformations needed by the selected source chain:
  callable-item `input[n].namespace`, top-level `max_output_tokens` and
  `truncation`, indexed `prompt_cache_breakpoint`, the documented null/zero
  reasoning-content repairs, and indexed `input[n].status` removal. Existing
  S254 tool-schema sanitization remains its current pre-send behavior; do not
  add a second schema pipeline.
- For `input[n].status`, determine the rejected item's exact `type` and remove
  `status` from every object of that same type in the `input` array. Other
  types retain their status. If the named item has no usable type, remove only
  the named index. Never shift input indexes or delete unrelated fields.
- Error `param` and a parseable error-message field name must agree; a missing
  `param` may use the message fallback. Malformed JSON, mismatched signals,
  unsupported item types, missing fields, or unsupported error shapes must not
  mutate or retry the request.
- The existing local one-time `invalid_encrypted_content` retry must coexist:
  its rewritten request body is registered with the rejected-field guard so it
  cannot be reissued as a duplicate transformed retry.
- Add default-tag unit/integration tests proving all transformations, strict
  no-retry guards, dedupe/budget behavior across reused `gin.Context`, the
  same-type batch status rule/fallback, and an `httptest`-backed non-WS HTTP
  forward retry that succeeds on the second request with the correctly
  rewritten body.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-responses-rejected-field-retry-s257`
- Base: `main@962db2c11`
- Upstream sources: `5e4da92d` (initial retry), `cf3577a3` (needed helper
  hardening only), `e440ac48c` (same-type status batch removal).
- Upstream's original `openai_gateway_forward.go` and most of the 41-file
  `cf3577a3` topology do not exist locally. The local HTTP loop and existing
  `invalid_encrypted_content` recovery are in
  `backend/internal/service/openai_gateway_service.go` around its non-WS
  error handling. This is a behavior-level adaptation, not a cherry-pick.
- The primary worktree contains user-owned S252 Pixel Cafe/Group/Wire/docs and
  untracked assets. It is protected: no primary worktree write, staging,
  cleanup, provider call, database, container, deployment, or push is allowed.

## Allowed Paths

- `backend/internal/service/openai_responses_rejected_field_retry.go`
- `backend/internal/service/openai_responses_rejected_field_retry_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_rejected_field_retry_test.go`
- `docs/workflow/worker-results/upstream-responses-rejected-field-retry-s257-result.md`

## Denied Paths

- `backend/internal/handler/**`, `backend/cmd/**`, `backend/internal/service/wire.go`,
  `frontend/**`, `knowledge/**`, `outputs/**`, dependencies/lockfiles,
  schema/migrations/Ent, account scheduling, billing, OAuth identity,
  WebSocket forwarding, real provider/database/container/deployment, push,
  and every product path not explicitly Allowed.
- The whole primary-worktree S252/Pixel Cafe/Group patch, including
  `backend/cmd/server/wire_gen.go`, is user-owned and must never be edited or
  staged.
- Do not import or recreate unrelated `cf3577a3` behavior (tool-name
  refactors, request logging, Ops changes, image intent, WS ingress/payload
  work, request body restructures, response-header changes, or frontend work).

## Constraints

- Keep the existing local HTTP retry/failover topology intact. Do not alter
  retry counts for transport, 429, pool, agent-identity, or WebSocket flows.
- Use `gin.Context` only for the request budget; do not store request bodies,
  credentials, or any sensitive upstream payload in context/logs.
- Keep all local safe error mapping, body redaction, accounting, request view,
  response parsing, timeout, and Ops behavior unchanged outside the approved
  non-WS 400 compatibility retry point.
- No dependency install or upgrade; reuse current `gin`, `gjson`, and `sjson`.
- All tests use fake HTTP/stubs only. No actual provider, database, container,
  deployment, or remote Git operation.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -list 'Test(NormalizeOpenAIResponsesRejectedFieldRetryBody|OpenAIResponsesRejectedFieldRetryState|OpenAIGatewayService_Forward.*Rejected.*Field)'
go test ./internal/service -run 'Test(NormalizeOpenAIResponsesRejectedFieldRetryBody|OpenAIResponsesRejectedFieldRetryState|OpenAIGatewayService_Forward.*Rejected.*Field)' -count=10
go test ./internal/service -count=1 -timeout=3m
go test ./cmd/server -run '^$' -count=1
Pop-Location

gofmt -w <all allowed Go production and test files>
git diff --check
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' <all allowed Go files>
git diff --name-only <base>..HEAD
git diff --cached --name-only
git diff --name-only
git ls-files -u
```

## Output

- One business commit limited to the four allowed service source/test paths,
  followed by one result-report commit.
- The result begins exactly `### DONE: upstream-responses-rejected-field-retry-s257`,
  `### FAILED: ...`, or `### BLOCKED: ...`; list changed files, real command
  results, source mapping, scope evidence, risks, and `knowledge_candidates`.
- An independent QA Worker may write only
  `docs/workflow/qa-reports/upstream-responses-rejected-field-retry-s257-qa.md`
  and must use `PASS`, `FAIL`, or `BLOCKED` as the first-line verdict.

## Stop Rules

- Stop if correct behavior requires an excluded handler/WS/Wire/dependency,
  persistence, identity, billing, migration, or frontend path.
- Stop if any retry can loop unbounded, repeat a seen body, change unrelated
  input types, accepts a non-explicit rejection, or changes the existing
  `invalid_encrypted_content` one-time semantics.
- Stop before QA or mainline integration if focused/default-tag/package/server
  gates fail, scope expands, conflict markers/unmerged entries appear, or the
  protected primary worktree changes.

## Budget

- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
