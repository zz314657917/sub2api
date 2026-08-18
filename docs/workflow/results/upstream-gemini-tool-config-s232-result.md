### PASS: upstream-gemini-tool-config-s232

## Changed Files

- `backend/internal/pkg/antigravity/gemini_types.go`
- `backend/internal/pkg/antigravity/request_transformer.go`
- `backend/internal/pkg/antigravity/request_transformer_test.go`

## Implementation

- Applied upstream `3c3bb2fa1` and `1ba92449c` in order as adapted commits
  `3dc5e7acf` and `6a5d4d95a`.
- Added the optional typed `includeServerSideToolInvocations` field.
- Enabled it only for mixed Google Search plus function declarations; pure
  function and pure web-search requests remain unset.

## Evidence

- Focused mixed/function-only/web-search-only tests, `-count=10`: PASS.
- Complete `go test ./internal/pkg/antigravity -count=1`: PASS.
- Complete `go test ./internal/service -count=1`: PASS (72.054s).
- `go test ./cmd/server -run '^$' -count=1`: PASS (10.592s).
- `gofmt -l` on all three allowed Go files: no output.
- `git diff --check 91e7b4f820..HEAD`: PASS.
- Exact product scope: three allowed business/test files only.
- Upstream provenance: both source commits are reachable and no later
  `upstream/main` edits touch these three owners.
- Worktree index is clean after the two business commits; no conflict markers
  or unmerged entries.

## Risks

- No live Gemini/provider request was made; validation uses deterministic typed
  transform tests and package/service regression.
- Main integration and independent QA remain pending.

## Contract Compliance

PASS. No gateway, schema, frontend, fingerprint, dependency, provider,
database, deployment, container, knowledge, or user-owned path changed.
