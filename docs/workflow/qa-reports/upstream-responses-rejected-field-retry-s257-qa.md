### PASS: upstream-responses-rejected-field-retry-s257

# QA Report

## Task ID

`upstream-responses-rejected-field-retry-s257`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-responses-rejected-field-retry-s257.md`
- Candidate baseline: `903140a48`; business commit: `f8aa4a7b9`.

## Evidence

- Diff reviewed: `yes`. `f8aa4a7b9^..f8aa4a7b9` has exactly the four allowed
  service paths: `openai_gateway_service.go`, the HTTP-forward retry test,
  `openai_responses_rejected_field_retry.go`, and its test. No handler, WS,
  Wire, identity, billing, frontend, dependency/lockfile, Ent, or migration
  path changed.
- Provenance reviewed: local objects `5e4da92d`, `cf3577a3`, and `e440ac48c`
  exist. The selected business diff carries only the rejected-field adaptation;
  it does not import the unrelated 41-file `cf3577a3` topology.
- Retry placement reviewed: WSv2 returns before the added non-WS HTTP loop.
  The added normalizer is called only after an HTTP error response and itself
  requires status `400`; the production `Responses` handler sets HTTP ingress.
- Transformation review: callable `input[n].namespace`, top-level
  `max_output_tokens`/`truncation`, indexed `prompt_cache_breakpoint`, null
  message/reasoning content, and zero-length reasoning-content rewrites are
  guarded by parseable explicit evidence. `param` and parseable message names
  must agree; a missing `param` uses the message fallback. Malformed, non-400,
  non-explicit, mismatched, and unsupported-type cases do not retry.
- State review: the request-context budget permits exactly six transformed
  bodies across reused `gin.Context`; each account state hashes and rejects its
  own duplicate/empty/original body. The existing one-time
  `invalid_encrypted_content` rewrite is remembered before its retry. A later
  failover may apply the same body once under the shared budget, as required.
- Status review: an explicit `input[n].status` rejection reads that item type,
  clears `status` from all same-type objects, preserves other types and fields,
  and falls back to only the named index when type is unusable. The HTTP-forward
  fake-upstream test records two actual outgoing bodies and proves the second
  removes only same-type statuses.

## Commands Run

```text
Push-Location backend; go test ./internal/service -list 'Test(NormalizeOpenAIResponsesRejectedFieldRetryBody|OpenAIResponsesRejectedFieldRetryState|OpenAIGatewayService_Forward.*Rejected.*Field)'
  -> exit 0; discovered 5 target tests.

Push-Location backend; go test ./internal/service -run 'Test(NormalizeOpenAIResponsesRejectedFieldRetryBody|OpenAIResponsesRejectedFieldRetryState|OpenAIGatewayService_Forward.*Rejected.*Field)' -count=10
  -> exit 0; ok github.com/Wei-Shaw/sub2api/internal/service 0.081s.

Push-Location backend; go test ./internal/service -count=1 -timeout=3m
  -> exit 0; ok github.com/Wei-Shaw/sub2api/internal/service 65.939s.

Push-Location backend; go test ./cmd/server -run '^$' -count=1
  -> exit 0; ok github.com/Wei-Shaw/sub2api/cmd/server 5.566s [no tests to run].

gofmt -d <four changed Go files>
  -> exit 0; no output (already formatted; read-only QA probe).
git diff --check 962db2c11..903140a48
  -> exit 0.
rg conflict markers across the four changed Go files
  -> exit 1 (no matches).
git diff --cached --name-only; git ls-files -u
  -> both empty before QA-report staging.
```

## Protected Primary Worktree

- Read-only snapshots at QA start and before report creation both found 106
  porcelain entries and patch-id
  `87e53602732ffd3a5708e02cd0d8537586bc5f2a` (with the standard zero side).
  No natural change was observed; the primary worktree was never written.

## Findings

未发现明确问题。

## Risk / Unverified Boundary

- All runtime verification used default-tag tests, `httptest` context/fake HTTP
  upstream, and server compilation only. No real provider, database, container,
  deployment, or push was accessed.

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

Not required for this passing candidate.

## Knowledge Promotion

`none`
