# Upstream OpenAI Sticky System Prefix S245

## Task ID

`upstream-openai-sticky-system-prefix-s245`

## Role

Developer Worker and independent QA Worker both use `gpt-5.6-terra` in
separate executions and evidence paths. Codex is Planner and Final Evaluator.
The implementation must follow this approved contract without widening scope.

## Goal

Behaviorally adapt upstream source `e45490a36` (merged by `2ddda6735`) so the
content-derived Chat Completions sticky seed includes only the leading,
contiguous system/developer prefix. System or developer messages injected after
conversation history begins must not change account affinity.

## Success Criteria

- Consecutive leading `system` and `developer` messages remain part of the
  sticky seed, so changing that initial prefix still changes the seed.
- The prefix closes at the first non-system/developer message. Any later
  system/developer message is ignored for sticky identity, including messages
  inserted after a user, assistant, tool, or other role.
- The first user message is still captured after the prefix closes, and later
  user/assistant history remains excluded as before.
- Responses API `input` handling, `instructions`, model, tools/functions,
  canonical JSON normalization, explicit session hints, and hash formatting
  remain unchanged.
- Focused regressions pass repeatedly; the complete service package and server
  compilation remain green.
- No primary-worktree user dirty or untracked file is staged, overwritten, or
  committed.

## Frozen Base And Provenance

- Frozen product base: local
  `main@4ddfb0dc53561ebbdfc6c84f793c567bd204552c`.
- Upstream source: `e45490a36031474f19888b14cbcb55d18945a801`.
- Upstream merge: `2ddda67354e656aca72b5eaa4071de766a8b578f`.
- Upstream audit tip:
  `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54`.
- Both source and merge are ancestors of the audit tip. The two-file
  first-parent patch does not apply directly because upstream first introduced
  the unrelated single-scan implementation in `86800a8cd`; adapt only the
  `systemPrefixOpen` behavior and focused regressions to the local owners.
- Upstream `219368ec6` is not a prerequisite. Its Composite video fix remains
  deferred because the local tree lacks the upstream Composite Resolver and
  `GrokVideoGeneration` owner; a route-only port would still fail in the local
  video handler.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`,
  `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`, and this contract.
- Product owner:
  `backend/internal/service/openai_content_session_seed.go`.
- Test owner:
  `backend/internal/service/openai_content_session_seed_test.go`.

## Allowed Paths

- `backend/internal/service/openai_content_session_seed.go`
- `backend/internal/service/openai_content_session_seed_test.go`
- `docs/workflow/worker-results/upstream-openai-sticky-system-prefix-s245-result.md`
- `docs/workflow/qa-reports/upstream-openai-sticky-system-prefix-s245-qa.md`

## Denied Paths

- All other backend, frontend, generated, schema, migration, dependency,
  lockfile, configuration, deployment, container, and workflow product paths.
- `docs/workflow/status.md`, `docs/workflow/spec.md`,
  `docs/workflow/main-log.md`, this contract, and all `knowledge/**` paths are
  Controller-owned and denied to workers.
- All user-owned dirty and untracked paths in the primary worktree, including
  the eleven current Pixel Cafe paths and `outputs/`.
- Remote writes, push, force operations, history rewrites, real provider
  traffic, shared/production data, and browser automation.

## Constraints

- Keep the local direct `gjson` scan topology; do not import upstream's
  unrelated single-scan refactor or rename constants/helpers.
- Add the minimal prefix-state guard and focused regression cases. Do not
  redesign sticky-session selection, cache keys, TTLs, scheduler behavior,
  request parsing, or hashing.
- Preserve the exact current behavior for Responses API input arrays, including
  their existing system/developer handling; this source fix is Chat messages
  only.
- Do not install or update dependencies, call a real provider, or touch shared
  services.
- Do not stage, overwrite, revert, or format unrelated work.

## Acceptance Commands

From `backend/` in the isolated worktree:

```powershell
go test ./internal/service -run '^TestDeriveOpenAIContentSessionSeed_ChatCompletions_(IgnoresLaterSystemMessages|UsesLeadingSystemDeveloperPrefix)$' -count=10
go test ./internal/service -run '^TestDeriveOpenAIContentSessionSeed_' -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -l internal/service/openai_content_session_seed.go internal/service/openai_content_session_seed_test.go
```

From the worktree root:

```powershell
git diff --check
git diff --cached --name-only
git ls-files -u
git merge-base --is-ancestor e45490a36031474f19888b14cbcb55d18945a801 upstream/main
git merge-base --is-ancestor 2ddda67354e656aca72b5eaa4071de766a8b578f upstream/main
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend/internal/service/openai_content_session_seed.go backend/internal/service/openai_content_session_seed_test.go
```

The Controller must additionally verify exact business/evidence commit
allowlists, source/merge first-parent scope, empty index/conflict state, and
preservation of the primary worktree's protected snapshot.

The protected primary-worktree patch ID is scoped to these eleven user-owned
paths only:

- `backend/internal/service/cafe_public.go`
- `backend/internal/service/cafe_public_test.go`
- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/features/pixelCafe/components/CafeScene.vue`
- `frontend/src/features/pixelCafe/components/SceneFallback.vue`
- `frontend/src/features/pixelCafe/components/__tests__/CafeScene.spec.ts`
- `frontend/src/features/pixelCafe/renderer/assetManifest.ts`
- `frontend/src/features/pixelCafe/renderer/createCafeRenderer.ts`
- `frontend/src/features/pixelCafe/renderer/sceneLayout.ts`
- `frontend/src/types/pixelCafe.ts`

Their combined stable patch ID must remain
`370ac77de0e2f530ab652b99fb3eb35e809f4c84`. The primary staged/unmerged index
must remain empty, and `outputs/` must retain its two pre-existing untracked
files.

## Output

- Developer produces one business commit containing only the two product/test
  paths and one separate evidence commit containing only
  `docs/workflow/worker-results/upstream-openai-sticky-system-prefix-s245-result.md`.
- The Developer report first line must be exactly
  `### DONE: upstream-openai-sticky-system-prefix-s245`,
  `### BLOCKED: upstream-openai-sticky-system-prefix-s245`, or
  `### FAILED: upstream-openai-sticky-system-prefix-s245`.
- Independent QA may modify only
  `docs/workflow/qa-reports/upstream-openai-sticky-system-prefix-s245-qa.md`;
  its first line must be exactly
  `### PASS: upstream-openai-sticky-system-prefix-s245`,
  `### FAIL: upstream-openai-sticky-system-prefix-s245`, or
  `### BLOCKED: upstream-openai-sticky-system-prefix-s245`.
- Reports list changed files, commands run, key output, risks, contract
  compliance, and `knowledge_candidates` without unrelated long logs.

## Stop Rules

- Stop if `gpt-5.6-terra` is unavailable; do not silently replace the model.
- Stop if implementation requires any path outside the allowlist, the upstream
  single-scan refactor, dependency changes, gateway/scheduler redesign,
  frontend/schema changes, browser automation, or real external state.
- Stop if the focused selector discovers no tests, a baseline failure is owned
  outside this contract, or any protected-primary path changes unexpectedly.
- Stop rather than weakening the leading-prefix rule or applying it to
  Responses `input` without a separate reviewed contract.

## Budget

- worker_mode: native `gpt-5.6-terra`
- qa_worker_mode: native `gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- developer_max_budget_usd: `0.10`
- qa_max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Status

`contract-approved`

## Worker Output

Same requirements as `Output`; this compatibility heading is retained for the
worker dispatcher.
