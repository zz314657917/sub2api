---
task_id: upstream-async-image-tasks-s123
phase: contract-approved
qa_mode: runtime
source_commits:
  - 1fb942dd7
  - 0eb6e21aa
  - 37db8d031
  - b08cab91a
  - 0343dba91
---

# Task Contract: Async Image Task API

## Task ID

upstream-async-image-tasks-s123

## Role

Generator implements only this contract after Evaluator approval. The
implementation is a selective adaptation, not a merge of `upstream/main`.

## Goal

Add an opt-in OpenAI-compatible asynchronous image task API that returns a
pollable task immediately and stores compact, object-storage-backed results.
Existing console `image-creator` tasks and synchronous Images API requests must
keep their current behavior.

## Success Criteria

- `POST /v1/images/generations/async` and `POST /v1/images/edits/async`
  accept the same non-streaming payloads as their synchronous equivalents and
  return `202 Accepted`, `Location`, `Retry-After`, and a poll URL.
- `GET /v1/images/tasks/:task_id` is readable only by the submitting user and
  API key, reports `processing`, `completed`, or `failed`, and remains
  pollable after the feature is switched off.
- The asynchronous execution reuses the current image gateway. It preserves
  authentication, group/capability checks, moderation, concurrency, routing,
  billing, and failover behavior rather than duplicating them.
- The feature is disabled unless a complete S3-compatible image-storage
  configuration resolves. It must never persist raw image base64 to Redis;
  successful results expose object-storage URLs and an offload failure produces
  a failed task.
- File configuration, environment overrides, and the admin hot-setting path
  work without exposing or overwriting stored secrets. Admin settings use the
  existing encryption and step-up protections.
- `image_creator.*`, synchronous `/v1/images/generations` and
  `/v1/images/edits`, schema/migrations, deployments, and containers remain
  unchanged.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-async-image-tasks-s123`
- Base: local `main` commit `4672daa0c`; the primary worktree is dirty and is
  not an implementation target.
- Upstream reference: commits `1fb942dd7`, `0eb6e21aa`, `37db8d031`,
  `b08cab91a`, and `0343dba91` from `upstream/main`.
- Existing boundary: local `ImageCreatorService` is a user-console DB task
  queue with its own storage configuration. This task adds a separate gateway
  API task service and must not merge the two models.

## Allowed Paths

- `backend/cmd/server/wire_gen.go`
- `backend/go.mod`
- `backend/go.sum`
- `backend/internal/config/config.go`
- `backend/internal/config/*image_storage*_test.go`
- `backend/internal/handler/endpoint.go`
- `backend/internal/handler/endpoint_test.go`
- `backend/internal/handler/handler.go`
- `backend/internal/handler/image_task*.go`
- `backend/internal/handler/wire.go`
- `backend/internal/handler/admin/backup_handler.go`
- `backend/internal/repository/backup_s3_store.go`
- `backend/internal/repository/image_storage_s3.go`
- `backend/internal/repository/image_task_store*.go`
- `backend/internal/repository/s3_client.go`
- `backend/internal/repository/wire.go`
- `backend/internal/server/middleware/api_key_auth*.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_test.go`
- `backend/internal/service/image_storage*.go`
- `backend/internal/service/image_task*.go`
- `backend/internal/service/wire.go`
- `deploy/config.example.yaml`
- `docs/ASYNC_IMAGE_TASKS.md`
- `README.md`
- `frontend/src/api/admin/backup.ts`
- `frontend/src/i18n/locales/en/admin/backup.ts`
- `frontend/src/i18n/locales/zh/admin/backup.ts`
- `frontend/src/views/admin/BackupView.vue`
- `frontend/src/views/admin/__tests__/BackupView*.spec.ts`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-async-image-tasks-s123.md`
- `docs/workflow/worker-results/upstream-async-image-tasks-s123-result.md`
- `docs/workflow/qa-reports/upstream-async-image-tasks-s123-qa.md`

## Denied Paths

- `backend/internal/service/image_creator*.go`
- `backend/internal/handler/image_creator*.go`
- `frontend/src/views/user/ImageCreatorView.vue`
- `backend/migrations/**`
- `deploy/docker-compose*.yml`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`
- Any path not listed in Allowed Paths.

## Constraints

- Do not merge or cherry-pick `upstream/main` as a branch. Port only required
  behavior and adapt it to local signatures.
- Keep the feature opt-in and fail closed when storage credentials are absent.
- Use bounded background contexts; do not retain the live Gin request after the
  submission response returns.
- Re-hosted upstream image URLs must have bounded size, timeout, redirect, and
  private-network behavior reviewed before release. Do not weaken existing
  request or egress safeguards.
- Do not run Docker, change live configuration, contact a real object-storage
  bucket, deploy, push, or publish in this task.
- Preserve unrelated changes and make no broad formatting/refactoring pass.

## Acceptance Commands

```powershell
Set-Location backend
go test ./internal/config -run 'Test.*ImageStorage' -count=1
go test ./internal/service -run 'TestImageTask|TestImageStorage|TestImageStorageSetting' -count=1
go test ./internal/handler -run 'TestAsyncImageHandler|TestImageTaskAdminToggle' -count=1
go test ./internal/repository -run 'Test.*Image(Task|Storage)' -count=1
go test ./internal/server/middleware -run 'TestAPIKeyAuth.*ImageTask' -count=1
go test ./internal/server/routes -run 'TestGatewayRoutes' -count=1
go test ./... -run '^$'
gofmt -d <changed-go-files>

Set-Location ..
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
git diff --check
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend frontend deploy docs
```

## Output

- Write `docs/workflow/worker-results/upstream-async-image-tasks-s123-result.md`
  with changed files, compatibility decisions, command results, risks, and
  evidence gaps.
- Write `docs/workflow/qa-reports/upstream-async-image-tasks-s123-qa.md` with
  first-line `### PASS`, `### FAIL`, or `### BLOCKED` verdict.
- Keep workflow status and main-log aligned with each gate.

## Stop Rules

- Stop for a decision if completing the feature would modify denied paths,
  require a schema migration, weaken auth/ownership/secret handling, or demand
  a production object-storage configuration.
- Stop if upstream dependencies cannot be adapted without a broad gateway or
  backup-storage refactor.
- Stop before a release claim when Redis/S3/API-key runtime smoke is absent.
