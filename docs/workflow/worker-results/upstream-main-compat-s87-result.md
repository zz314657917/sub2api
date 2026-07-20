### DONE: upstream-main-compat-s87

## Baseline and Scope

- Local baseline: `3418267b324412071b6b8ec1ae378df2e6372f2a`.
- Upstream snapshot: `e625ce3b3b3b955b7c3afc93221f7c5f0ae55aa8` (`v0.1.162`).
- Ported behavior references: `e50276617`, `604b764cf`, `4c73149af`.
- Explicitly deferred: `a05b87321` because the local plan-currency schema/API
  prerequisite is absent and its upstream migration numbering conflicts with
  the local migration history.
- No upstream history was merged; work stayed in the S87 isolated worktree.

## Implemented Behavior

- API-key update DTOs use nullable slice pointers. Omitted and JSON `null` IP
  lists preserve existing values; explicit empty arrays clear one list; valid
  non-empty arrays replace one list; invalid lists fail before repository
  update. Create DTOs remain ordinary `[]string` slices.
- API-key quota exhaustion uses the OpenAI `insufficient_quota` envelope only
  on the three registered Responses roots and their subpaths. Non-Responses
  protocol behavior remains unchanged. The existing model-aware guard is
  reused without adding routes or changing deferred billing.
- Available Channels now mounts the table on `.table-wrapper`, restoring the
  TablePageLayout scroll chain and removing the inner clipping card wrapper.

## Changed Paths

- `backend/internal/handler/api_key_handler.go`
- `backend/internal/handler/api_key_update_s87_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_update_s87_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s87_test.go`
- `backend/internal/server/middleware/middleware.go`
- `frontend/src/components/channels/AvailableChannelsTable.vue`
- `frontend/src/components/channels/__tests__/AvailableChannelsTable.spec.ts`
- S87 workflow contract/status/spec/log files.

## Executed Checks

- `go test ./internal/handler -run '^TestS87APIKeyUpdateJSONPresence$' -count=1`: PASS.
- `go test ./internal/service -run '^TestS87APIKeyIPRestrictions$' -count=1`: PASS.
- `go test ./internal/server/middleware -run '^TestS87APIKeyQuotaError' -count=1`: PASS.
- Default-tag `go test -list '^TestS87'` discovery found all four required test
  names across handler, service, and middleware packages.
- `go test ./internal/handler -run '^TestHandleFailoverError_' -count=1`: PASS.
- `node .../vitest.mjs run src/components/channels/__tests__/AvailableChannelsTable.spec.ts`: PASS, 1 file / 2 tests.
- `gofmt`, `git diff --check`, exact working-tree allowlist, conflict/unmerged
  scan, and S85 path freeze: PASS.

## Not Performed

- No upstream live request, authenticated browser smoke, push, deployment,
  container update, migration, Ent generation, or dependency installation.
- Frontend typecheck was attempted through the existing pnpm store but stopped
  on the unrelated broken `@airwallex/components-sdk` junction in the baseline
  dependency tree. Direct Vite build was also attempted and stopped because the
  baseline dependency tree lacks the `.bin/vue-tsc` wrapper used by the checker.
