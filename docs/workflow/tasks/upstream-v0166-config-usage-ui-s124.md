---
task_id: upstream-v0166-config-usage-ui-s124
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Manually adapt three independent upstream v0.1.166 compatibility behaviors to
the local topology: explicit `CONFIG_FILE` loading, administrator usage-log
`request_id` filtering, and administrator usage-page route user-label
hydration.

## Success Criteria

- A non-blank `CONFIG_FILE` loads the named YAML file for both normal startup
  and bootstrap address resolution; omitted `CONFIG_FILE` preserves the
  existing search paths and environment overrides.
- `GET /api/v1/admin/usage?request_id=<id>` trims the value and filters the
  paginated usage list by exact request ID. An omitted value changes no query
  behavior.
- `/admin/usage?user_id=<id>` displays the selected user's email in the
  existing user filter. A user typing in that filter while the lookup is
  pending keeps their own search text.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-v0166-config-usage-ui-s124`
- Upstream references: `5c471485a`, `1850e0095`, and `d11b83870`.
- The primary `F:/mcplugins/sub2api` worktree is user-dirty and must remain
  unchanged until this isolated branch has passed its source-level gates.

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/handler/admin/usage_handler.go`
- `backend/internal/handler/admin/usage_handler_request_type_test.go`
- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `frontend/src/components/admin/usage/UsageFilters.vue`
- `frontend/src/components/admin/usage/__tests__/UsageFilters.spec.ts`
- `frontend/src/views/admin/UsageView.vue`
- `frontend/src/views/admin/__tests__/UsageView.spec.ts`
- `docs/workflow/**`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/pricing_service.go`
- `backend/internal/service/openai_gateway_grok*.go`
- `backend/internal/pkg/claude/constants.go`
- `backend/internal/domain/constants.go`
- `backend/resources/model-pricing/**`
- `frontend/src/components/account/**`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/views/admin/group-buy/**`
- `deploy/**`
- `Dockerfile*`
- `knowledge/**`
- `outputs/**`

## Constraints

- Adapt behavior manually; do not cherry-pick or merge upstream history.
- Keep the usage predicate exact and list-only. Do not alter usage statistics,
  exports, persistence, billing, or query ordering.
- Keep `CONFIG_FILE` optional and avoid logging configuration contents.
- The isolated worktree currently has no frontend dependencies. Dependency
  preparation may run only as `corepack.cmd pnpm install --frozen-lockfile`
  under its `frontend` directory; it must not modify manifests or lockfiles.
- Do not deploy, update containers, push, clean the primary worktree, or
  modify unrelated branches/worktrees.

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-v0166-config-usage-ui-s124/backend
go test ./internal/config -run "ConfigFile|GetServerAddress" -count=1
go test ./internal/handler/admin -run "Usage.*Request" -count=1
go test ./internal/repository -run "UsageLogRepository.*Request" -count=1
go test ./... -run "^$"
cd ../frontend
corepack.cmd pnpm exec vitest run src/components/admin/usage/__tests__/UsageFilters.spec.ts src/views/admin/__tests__/UsageView.spec.ts
corepack.cmd pnpm run typecheck
corepack.cmd pnpm run build
cd ..
git diff --check
```

## Output

- A focused code change, an S124 QA report, and split local commits only after
  all source-level acceptance gates pass.
- No worker is invoked for this bounded task; Codex performs implementation and
  final evaluation directly.

## Stop Rules

- Stop if the behavior requires a migration, deployment change, billing,
  protocol/routing rewrite, or a denied path.
- Stop if exact request-ID filtering cannot be applied without changing
  statistics or exports.
- Stop before primary-worktree integration if an S124 file overlaps a user
  dirty file or if a test exposes unrelated baseline failure.
