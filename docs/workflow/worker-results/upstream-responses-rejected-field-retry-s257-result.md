### DONE: upstream-responses-rejected-field-retry-s257

## Changed files

- `backend/internal/service/openai_responses_rejected_field_retry.go`
- `backend/internal/service/openai_responses_rejected_field_retry_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_rejected_field_retry_test.go`

## Source mapping

- `5e4da92d`: adapted the bounded rejected-field retry guard and callable
  `input[n].namespace` compatibility behavior to the local monolithic HTTP
  forwarding loop.
- `cf3577a3`: retained only the reusable helper hardening required here:
  request-scoped shared six-body budget, per-account body dedupe, top-level
  `max_output_tokens`/`truncation`, indexed `prompt_cache_breakpoint`, and the
  null/maximum-zero reasoning-content repairs. The unrelated tool-schema and
  41-file refactor behavior was not imported.
- `e440ac48c`: adapted same-type `input[n].status` removal. A named rejected
  item clears `status` for every input item of the same exact type; without a
  usable type it clears only the named index.

## Executed checks

- `gofmt -w backend/internal/service/openai_responses_rejected_field_retry.go backend/internal/service/openai_responses_rejected_field_retry_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_rejected_field_retry_test.go` — passed.
- `Push-Location backend; go test ./internal/service -list 'Test(NormalizeOpenAIResponsesRejectedFieldRetryBody|OpenAIResponsesRejectedFieldRetryState|OpenAIGatewayService_Forward.*Rejected.*Field)'` — discovered five focused default-tag tests.
- `go test ./internal/service -run 'Test(NormalizeOpenAIResponsesRejectedFieldRetryBody|OpenAIResponsesRejectedFieldRetryState|OpenAIGatewayService_Forward.*Rejected.*Field)' -count=10` — passed (`0.080s` on the final run).
- `go test ./internal/service -count=1 -timeout=3m` — passed (`68.185s`).
- `go test ./cmd/server -run '^$' -count=1` — passed (`0.063s`, no tests to run).
- `git diff --check` — passed.
- Conflict-marker search across all four service files — no matches.
- Before business commit, `git diff --cached --name-only` was empty; the staged business commit contained exactly the four listed service files. `git ls-files -u` was empty.

## Scope evidence

- Business commit: `f8aa4a7b9 fix(openai): retry rejected Responses fields`.
- The retry state is created only in the non-WS HTTP `OpenAIGatewayService.Forward` loop. Existing WebSocket, handler, Wire, identity, billing, frontend, provider, database, container, deployment, and dependency paths remain unchanged.
- The existing one-time `invalid_encrypted_content` retry remains one-time; its serialized retry body is recorded by the rejected-field state before continuing.

## Risks

- Verification uses only `httptest` contexts and the existing fake HTTP upstream recorder; no real provider, database, container, deployment, or push was used.
- Compatibility rewrites intentionally require HTTP 400 plus explicit, non-conflicting rejected-field evidence. Other upstream errors retain the existing error/failover path.

## knowledge_candidates

- A reusable compatibility pattern: share a small transformed-body budget in request context while keeping body hashes per account attempt, so failover is bounded without blocking the same rewrite on a later account.
