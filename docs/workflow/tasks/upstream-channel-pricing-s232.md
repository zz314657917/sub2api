# Upstream Channel Pricing S232

## Task ID

`upstream-channel-pricing-s232`

## Role

Controller/Generator: Codex. Independent QA is performed from a separate
worktree before main integration.

## Goal

Behaviorally port upstream `8f6f45983` from the latest upstream line
(`upstream/main@49504adc9`, v0.1.178) so channel pricing model synchronization
and the admin channel editor recognize the first-class CN platforms `kimi`,
`zhipu`, and `deepseek`. Preserve composite-group semantics and the existing
local S226/S228 CN account behavior.

## Success Criteria

- Backend pricing sync maps `gemini` to the local LiteLLM provider key and
  supports `grok`, `kimi`, `zhipu`, and `deepseek` without changing unrelated
  platform validation.
- Admin channel platform ordering, composite-group filtering, and model tag
  colors expose the three CN platforms while composite groups remain limited to
  the five established composite platforms.
- Focused backend tests, frontend focused tests, frontend typecheck/build (when
  available), formatting, exact scope, provenance, conflict/index, and dirty
  worktree protection pass.

## Frozen Base

`main@91e7b4f820` (working tree user changes are protected and are not part of
this task).

## Allowed Paths

- `backend/internal/handler/admin/channel_handler.go`
- `backend/internal/handler/admin/channel_handler_test.go`
- `frontend/src/components/admin/channel/types.ts`
- `frontend/src/views/admin/ChannelsView.vue`
- `docs/workflow/results/upstream-channel-pricing-s232-result.md`

## Denied Paths

All other product files, including backend migrations/schema, quota-monitor
features, Codex identity/fingerprint files, gateway/provider logic, dependency
files, user-owned dirty files, `knowledge/*`, `outputs/*`, deployment,
containers, databases, provider traffic, and remote refs.

## Constraints

- Do not merge `main..upstream/main`; history is divergent.
- Adapt only the scoped behavior; preserve local naming and existing CN
  platform types.
- Do not overwrite, stage, or commit the user's dirty/untracked files.
- No push, deployment, container, database, or real provider operation.

## Acceptance Commands

From `backend/`:

```powershell
go test ./internal/handler/admin -run 'TestSyncPricingModels_ValidPlatform_EmptyService' -count=10
go test ./internal/handler/admin -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -l internal/handler/admin/channel_handler.go internal/handler/admin/channel_handler_test.go
```

From `frontend/`:

```powershell
pnpm exec vitest run src/components/admin/channel --run
pnpm run typecheck
pnpm run build
```

Also verify exact allowlist, clean index/conflict markers, upstream ancestry,
patch-id/provenance, and preservation of the frozen user dirty patch IDs and
untracked tutorial files.

## Output

One business implementation commit plus
`docs/workflow/results/upstream-channel-pricing-s232-result.md` with a first
line verdict (`### PASS`, `### FAIL`, or `### BLOCKED`), changed files,
commands/evidence, risks, and contract compliance.

## Stop Rules

- Stop if the implementation needs schema, quota-monitor, fingerprint, or
  provider changes.
- Stop if any denied path changes or the user's dirty/untracked state changes.
- Stop if focused behavior cannot be independently tested or if frontend
  dependencies are unavailable; report the verification blocker instead of
  claiming product failure.
