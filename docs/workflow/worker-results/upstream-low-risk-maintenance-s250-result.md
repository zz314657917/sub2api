### DONE: upstream-low-risk-maintenance-s250

## Scope

- Candidate branch: `codex/upstream-low-risk-maintenance-s250`
- Candidate business commits: `ed2002f57`, `94b8370ee`, `16ea417a3`
- Upstream behavior sources: `4a1da2950`, `cd05772e9`, `5dfad32b8`

## Changed Files

- DOMPurify security slice: `frontend/package.json`, `frontend/pnpm-lock.yaml`
- Ops memory slice: `backend/internal/service/ops_metrics_collector.go`,
  `backend/internal/service/ops_metrics_collector_memory_test.go`
- User concurrency slice: `frontend/src/components/admin/user/UserEditModal.vue`,
  `frontend/src/components/admin/user/__tests__/UserEditModal.spec.ts`,
  `frontend/src/i18n/locales/en/admin/users.ts`,
  `frontend/src/i18n/locales/zh/admin/users.ts`

## Verification

- `corepack pnpm --dir frontend install --frozen-lockfile --ignore-scripts`: PASS;
  lockfile was accepted without regeneration.
- `corepack pnpm --dir frontend why dompurify`: PASS; direct, Mermaid, and
  `@types/dompurify` paths resolve to `3.4.14`.
- `go test ./internal/service -run "TestResolveMemoryStats" -count=10`: PASS.
- `go test ./internal/service -count=1`: PASS (`64.615s`).
- `go test ./cmd/server -run '^$' -count=1`: PASS (compile-only).
- `corepack pnpm --dir frontend exec vitest run
  src/components/admin/user/__tests__/UserEditModal.spec.ts`: PASS (3 tests).
- `corepack pnpm --dir frontend run typecheck`: PASS.
- `corepack pnpm --dir frontend run build`: PASS (`vite` built in `21.31s`).
- `gofmt`, `git diff --check`, conflict-marker scan, and unmerged-index check:
  PASS.

## Risks

- The production build reports pre-existing Browserslist age, dynamic-import,
  and chunk-size warnings; it exits successfully and this task does not change
  the affected build topology.
- No browser session, real provider, container, deployment, shared database,
  or push was used.

## Knowledge Candidates

- None. The upstream sources and local topology are task-specific evidence.
