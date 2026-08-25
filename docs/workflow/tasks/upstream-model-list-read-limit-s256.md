# Task Contract

## Task ID

upstream-model-list-read-limit-s256

## Role

Developer Worker (`gpt-5.6-terra`) implements this approved, isolated
compatibility slice. Codex reviews the diff before a separate QA Worker
(`gpt-5.6-terra`) reruns every gate in a fresh worktree.

## Goal

Behaviorally adapt upstream `847c0c452` so operators may configure a bounded
read size for upstream model-list responses, while the zero/default behavior
remains the existing 8 MiB limit.

## Success Criteria

- Add `gateway.models_list_read_max_bytes`, defaulting to exactly 8 MiB and
  rejecting non-positive configured values.
- Use one service-level resolver with a safe 8 MiB fallback when a service is
  constructed with a nil or zero-valued config.
- Apply the resolver consistently to generic upstream `/models` reads, Codex
  manifest reads, Antigravity account-model reads, and Antigravity quota reads.
- Codex manifest reads exactly `limit + 1` bytes and return a bounded upstream
  error when the response exceeds the configured limit; they must not parse a
  truncated body as malformed JSON.
- Keep the 1 MiB Codex manifest cache entry cap, request timeouts, response
  parsing, auth/routing, billing, model mappings, and all non-model-list body
  limits unchanged.
- Add default-tag tests proving the default/zero fallback, validation,
  configured acceptance above 8 MiB, and configured oversized rejection on
  every affected read family.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-model-list-read-limit-s256`
- Base: `main@536c4bde3`
- Upstream source: `847c0c452` from merge `7075ae0d8`.
- Direct application fails because the local config, model-service, and Wire
  topology have since diverged. Preserve the source behavior; do not replay
  unrelated upstream structure.
- The primary worktree has user-owned S252 edits, including
  `backend/cmd/server/wire_gen.go`. The isolated branch is clean. Mainline
  integration is forbidden until Codex verifies the final Wire patch can apply
  without modifying those user-owned changes.

## Allowed Paths

- `backend/cmd/server/wire_gen.go`
- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/pkg/antigravity/client.go`
- `backend/internal/pkg/antigravity/client_test.go`
- `backend/internal/service/antigravity_quota_fetcher.go`
- `backend/internal/service/antigravity_quota_fetcher_test.go`
- `backend/internal/service/models_list_response_limit.go`
- `backend/internal/service/models_list_response_limit_test.go`
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_codex_models_service_test.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/upstream_models_test.go`
- `deploy/config.example.yaml`
- `docs/workflow/worker-results/upstream-model-list-read-limit-s256-result.md`

## Denied Paths

- `backend/cmd/server/wire.go`, `frontend/**`, `knowledge/**`, `outputs/**`
- schema, migrations, Ent output, dependencies/lockfiles, authentication,
  account scheduling, billing, model pricing/mappings, provider credentials,
  containers, deployment, real provider/database traffic, push, and every path
  not explicitly Allowed.
- All primary-worktree S252/Pixel Cafe/Group changes, including their dirty
  files, are owned by the user and must never be edited or staged.

## Constraints

- Keep the change configuration-only: no implicit increase and no unlimited
  reader. An explicit positive configured limit is the only behavior change.
- Use the local service/config topology. `wire_gen.go` may change only as the
  generated constructor call needed to inject `configConfig` into the existing
  Antigravity quota fetcher; do not hand-port unrelated generator changes.
- Keep generic/Codex/Antigravity error shapes compatible except the required
  explicit bounded Codex oversized-response error.
- Do not regenerate Ent, change `wire.go`, or install/upgrade dependencies.
- Do not use real provider, shared/production database, container, deployment,
  push, or primary-worktree operations.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/config -run "Test.*ModelsList.*" -count=10
go test ./internal/pkg/antigravity -run "Test.*FetchAvailableModels.*(Limit|limit)|Test.*ModelsList.*" -count=10
go test ./internal/service -run "Test(ResolveModelsListReadLimit|FetchCodexModelsManifest.*Configured.*Limit|FetchUpstreamSupportedModels.*Configured.*Limit|Antigravity.*ModelsList.*Limit)" -count=10
go test ./internal/config -count=1
go test ./internal/pkg/antigravity -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
Pop-Location

gofmt -w <all allowed Go production and test files>
git diff --check
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" <all allowed source/test files>
git diff --name-only <base>..HEAD
git diff --cached --name-only
git diff --name-only
git ls-files -u
```

## Output

- One business commit limited to the listed product/test/config/example/Wire
  paths, then one Developer result commit.
- The result must begin exactly `### DONE: upstream-model-list-read-limit-s256`,
  `### FAILED: ...`, or `### BLOCKED: ...`, and list changed files, real command
  results, source mapping, scope evidence, risks, and `knowledge_candidates`.
- An independent QA Worker may write only
  `docs/workflow/qa-reports/upstream-model-list-read-limit-s256-qa.md` and must
  use the same verdict format with `PASS`, `FAIL`, or `BLOCKED`.

## Stop Rules

- Stop if a compatible solution requires `wire.go`, a migration, Ent output,
  dependencies, a provider request, or any denied path.
- Stop if the default remains anything other than 8 MiB, a non-positive value
  is accepted, any affected family remains fixed at 8 MiB, or Codex parses an
  oversized response rather than rejecting it.
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
