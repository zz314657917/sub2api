### DONE: upstream-model-list-read-limit-s256

## Verdict

PASS. The isolated compatibility slice implements the approved bounded model-list
read configuration without modifying denied paths, calling real providers, or
touching the primary worktree.

## Changed files

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/pkg/antigravity/client.go`
- `backend/internal/pkg/antigravity/client_test.go`
- `backend/internal/service/antigravity_quota_fetcher.go`
- `backend/internal/service/models_list_response_limit.go`
- `backend/internal/service/models_list_response_limit_test.go`
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_codex_models_service_test.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/upstream_models_test.go`
- `backend/cmd/server/wire_gen.go`
- `deploy/config.example.yaml`

## Source mapping

- Behavioral source: upstream `847c0c452`.
- Local adaptation: retained the existing local Wire topology and changed only
  the existing Antigravity quota-fetcher constructor call to receive
  `configConfig`.
- Added the local service resolver so nil and zero-value configs return the
  exact `config.DefaultModelsListReadMaxBytes` value (8 MiB).
- Generic `/models`, Codex manifests, Antigravity account-model reads, and
  Antigravity quota reads all pass the resolved limit. Codex reads `limit + 1`
  bytes and returns a retryable bounded upstream error before JSON parsing when
  oversized. The 1 MiB Codex cache-entry cap is unchanged.

## Commands run

All commands ran from this isolated worktree. Focused discovery found the
default-tag tests before execution.

- `go test ./internal/config -list 'Test.*ModelsList.*'` — PASS
- `go test ./internal/pkg/antigravity -list 'Test.*FetchAvailableModels.*(Limit|limit)|Test.*ModelsList.*'` — PASS
- `go test ./internal/service -list 'Test(ResolveModelsListReadLimit|FetchCodexModelsManifest.*Configured.*Limit|FetchUpstreamSupportedModels.*Configured.*Limit|Antigravity.*ModelsList.*Limit)'` — PASS
- `go test ./internal/config -run "Test.*ModelsList.*" -count=10` — PASS
- `go test ./internal/pkg/antigravity -run "Test.*FetchAvailableModels.*(Limit|limit)|Test.*ModelsList.*" -count=10` — PASS
- `go test ./internal/service -run "Test(ResolveModelsListReadLimit|FetchCodexModelsManifest.*Configured.*Limit|FetchUpstreamSupportedModels.*Configured.*Limit|Antigravity.*ModelsList.*Limit)" -count=10` — PASS
- `go test ./internal/config -count=1` — PASS
- `go test ./internal/pkg/antigravity -count=1` — PASS
- `go test ./internal/service -count=1 -timeout=3m` — PASS (65.593s)
- `go test ./cmd/server -run '^$' -count=1` — PASS
- `gofmt -w <all allowed Go production and test files>` — PASS
- `git diff --check` — PASS
- conflict-marker scan — PASS (none)
- `git ls-files -u` — PASS (none)

## Scope evidence

- Business commit: `2f04ba987`.
- `backend/cmd/server/wire.go`, dependencies, migrations, generated Ent,
  primary-worktree files, and all S252/Pixel Cafe/Group paths were untouched.
- The staged business commit contains only the allowed product/test/config/Wire
  paths listed above.

## Risks

- Verification uses local `httptest`/stub upstreams only. No real provider,
  database, container, deployment, or production configuration was contacted.
- Mainline integration still requires Codex to verify that the one-line
  `wire_gen.go` patch applies cleanly against the protected primary-worktree
  S252 changes.

## knowledge_candidates

- None. This is a task-local compatibility implementation; no durable knowledge
  update is proposed.
