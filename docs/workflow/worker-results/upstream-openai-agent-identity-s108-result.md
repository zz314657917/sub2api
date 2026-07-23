### DONE: upstream-openai-agent-identity-s108

## Summary

- Manually ported the approved eleven-commit OpenAI Agent Identity backend
  behavior to the local monolithic gateway and WebSocket topology.
- Feature commit: `6b87a2d2b`.
- No upstream merge, schema change, production credential, deployment, or
  container update was used.

## Changed Files

- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/admin/account_codex_import.go`
- `backend/internal/handler/admin/account_codex_agent_identity_import_test.go`
- `backend/internal/handler/dto/credentials_redact_test.go`
- `backend/internal/service/account_credentials_redact.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/openai_agent_identity.go`
- `backend/internal/service/openai_agent_identity_test.go`
- `backend/internal/service/openai_agent_identity_compat_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_count_tokens.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_quota_service.go`
- `backend/internal/service/openai_quota_service_test.go`
- `backend/internal/service/openai_ws_client.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_pool.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/wire.go`

## Implementation

- Added PKCS#8 Ed25519 assertion signing, plaintext/encrypted task
  registration, persisted task recovery, shared per-account locking, and
  credential-safe errors.
- Added JSON import normalization for snake/camel Agent Identity shapes,
  Team/user isolation, 14-user synthetic coverage, and `A -> B -> A` batch
  deduplication.
- Added no-token Agent Identity authentication to account tests, usage,
  quota/reset-credit, HTTP compatibility routes, images, and supported WS
  paths while preserving OAuth, PAT, and API-key bearer behavior.
- Added dial-time WS assertions, stale prewarm invalidation, one-shot invalid
  task recovery, FedRAMP account headers, and Wire integration.
- Added `agent_private_key` snake/camel/lowercase redaction and import-extra
  protection.

## Commands Run

- Admin discovery `12`, DTO discovery `5`, service discovery `16`.
- Exact concurrency discovery `3`; all three passed with `-count=10`.
- Broader OpenAI/Codex service discovery `380`; all passed.
- Full affected handler packages passed.
- `go test ./internal/service -skip PeakMultiplier -count=1` passed.
- S108 and clean-baseline `PeakMultiplier` runs both passed.
- `go test ./cmd/server -run '^$' -count=1` passed after Wire generation.
- `gofmt`, `git diff --check`, allowlist, conflict-marker, unmerged-index,
  and credential-literal scans passed.

## Security And Contract Compliance

- Tests generate ephemeral Ed25519 keys and use loopback `httptest` servers.
- No supplied private key, token, runtime/task/account/user/Team identifier,
  or external credential file was copied into source, fixtures, reports, or
  commits.
- All 23 business/test paths are in the approved allowlist; denied schema,
  migration, repository, frontend, billing, deployment, container, VERSION,
  and production configuration paths were not changed.
- The absent Codex models manifest and independent import UI remain skipped.

## Unverified Risks

- The Windows toolchain has `CGO_ENABLED=0` and no available C compiler, so
  the race detector was not run; the contract's exact concurrency tests passed
  ten times instead.
- The repository's pre-existing `unit` aggregate test set does not compile due
  unrelated `stringPtr`, billing signature, and Grok runtime-block drift.
- No live K12 account, external OpenAI endpoint, authenticated browser,
  deployment, container, or running-service smoke was used.

## Knowledge Candidates

- None. S108-specific topology and evidence remain in repository workflow
  documents rather than global memory.

## Publication Note

- Feature commit `6b87a2d2b` was later published as an ancestor of the current
  `main`; this result file is retained as historical implementation evidence.
