### PASS: upstream-codex-fingerprint-convergence-s233

## Result

- Business commit: `775d4f7ed`.
- Upstream source: `a34123959`, reachable from `upstream/main@49504adc9`.
- The locally applicable credential-face and version convergence behavior was
  adapted to the existing S225/S230 Codex identity owners. Upstream-only PAT
  service and resolver topology were not imported.

## Changed Files

- `backend/internal/repository/openai_oauth_service.go`
- `backend/internal/repository/openai_oauth_service_test.go`
- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/openai_codex_identity.go`
- `backend/internal/service/openai_codex_identity_test.go`
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_codex_models_service_test.go`
- `backend/internal/service/openai_codex_version_consistency_test.go`

## Evidence

- `go test ./internal/service -run 'TestCodex|TestFetchCodexModelsManifest|TestOpenAICodex' -count=10`: PASS (`9.258s`).
- `go test ./internal/repository -run 'TestOpenAIOAuthServiceSuite' -count=10`: PASS (`5.530s`).
- `go test ./internal/service -count=1`: PASS (`65.653s`).
- `go test ./internal/repository -count=1`: PASS (`1.656s`).
- `go test ./cmd/server -run '^$' -count=1`: PASS (`5.608s`, no tests to run).
- `gofmt -l` over all eight allowed product/test files: PASS, no output.
- `git diff --check`, exact eight-file allowlist, conflict-marker and unmerged-index checks: PASS.
- OAuth exchange/refresh tests assert the canonical paired `User-Agent` and
  `originator`; the credential-face helper removes an existing inference-only
  `version` header.
- Manifest regression proves stale `client_version=0.125.0` remains in the
  query while the outbound `Version` header falls back to local
  `codexCLIVersion`.
- `openAICodexProbeVersion` has no remaining production reference; the quota
  probe and manifest fallback use `codexCLIVersion`.
- Protected main remained `main@de08968f1`, ahead of `origin/main` by 102
  commits, with an empty staged index and the pre-existing user dirty/untracked
  paths unchanged during Controller validation.

## Contract Compliance

- No upstream PAT service, identity resolver, frontend, migration, dependency,
  gateway route, knowledge, output, user-owned dirty path, or remote ref was
  changed.
- No provider, database, container, deployment, or push operation was run.

## Risks

- Validation uses unit/httptest coverage and package compilation; no real
  OpenAI credential endpoint or provider traffic was authorized.
- Independent QA from a separate clean worktree is still required before main
  integration.
