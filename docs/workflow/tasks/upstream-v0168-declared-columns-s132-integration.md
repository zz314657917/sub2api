---
task_id: upstream-v0168-declared-columns-s132-integration
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Port only the independently revertible declared-column persistence behavior
from S132 onto the current local mainline. User and API-key mutations must
write exactly their intended columns so a stale entity cannot overwrite an
unrelated concurrent update.

## Success Criteria

- `UserRepository` and `APIKeyRepository` expose an explicit update-field
  boundary which rejects an empty field set and writes only declared fields.
- Existing user, API-key, administrator, authentication, quota, rate-limit,
  billing, group-binding, and moderation call sites declare their exact
  persisted fields.
- The administrator promo-code save path preserves concurrent atomic
  `used_count` redemption increments.
- Focused regressions demonstrate that stale user/API-key snapshots do not
  replay unrelated fields, and the local backend still compiles.

## Context

- Repo: `F:/mcplugins/sub2api`
- Integration worktree: `E:/codex-worktrees/sub2api-s132-concurrency-integration-20260804`
- Baseline: `main@9099db0c4`
- Source behavior: `e6c5a37ca fix(concurrency): scope user and api key updates`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`, and
  `docs/workflow/qa-reports/upstream-v0168-declared-columns-s132-qa.md` from
  the source commit.

## Allowed Paths

- `backend/internal/repository/user_repo.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/repository/promo_code_repo.go`
- `backend/internal/repository/*lost_update*_test.go`
- `backend/internal/repository/*user*_test.go`
- `backend/internal/repository/*api_key*_test.go`
- `backend/internal/service/user_service.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/auth_service.go`
- `backend/internal/service/auth_email_binding.go`
- `backend/internal/service/auth_email_oauth_auto.go`
- `backend/internal/service/content_moderation.go`
- `backend/internal/service/*user*_test.go`
- `backend/internal/service/*api_key*_test.go`
- `backend/internal/service/*admin*_test.go`
- `backend/internal/service/*auth*_test.go`
- `backend/internal/service/*moderation*_test.go`
- `backend/internal/service/cafe_room_activation_service_test.go`
- `backend/internal/service/group_buy_test.go`
- `backend/internal/service/leaderboard_participation_exclusion_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/handler/*api_key*_test.go`
- `backend/internal/handler/auth_*_test.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/server/middleware/*api_key*_test.go`
- `backend/internal/server/middleware/admin_auth_test.go`
- `docs/workflow/tasks/upstream-v0168-declared-columns-s132-integration.md`
- `docs/workflow/qa-reports/upstream-v0168-declared-columns-s132-integration-qa.md`

## Denied Paths

- `backend/migrations/**`
- `backend/go.mod`
- `backend/go.sum`
- `backend/internal/**/passkey*`
- `frontend/**`
- `deploy/**`
- `Dockerfile*`
- `docker-compose*.yml`
- `outputs/**`
- `output/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `C:/Users/Administrator/.codex/memories/**`
- Any local container, database, provider account, deployment, or production state.

## Constraints

- Adapt to the current repository/service APIs; do not cherry-pick the whole
  S132 commit and do not introduce a schema or authentication change.
- Keep the source change restricted to concurrency lost-update protection.
- Preserve validation, audit, cache invalidation, and public API behavior at
  each existing call site.
- A post-baseline Cafe test stub may receive only the mechanical repository
  method-signature update required to compile this narrowed interface change.
- Do not touch the primary worktree's uncommitted `main-log.md` or `outputs/`.

## Acceptance Commands

Run from `backend`:

```powershell
gofmt -w <changed Go files>
go test ./internal/repository ./internal/service ./internal/handler ./internal/server/... -run "Test(User|APIKey|UpdateQuotaUsed|UpdateProfile|ChangePassword|UpdateStatus)" -count=1
go test ./... -run "^$"
go build ./...
```

Run from the worktree root:

```powershell
git diff --check
git ls-files -u
git diff --name-only main...HEAD
```

## Output

- One independently revertible integration commit with this contract and an
  evaluator QA report.
- The QA report must start with `### PASS:`, `### FAIL:`, or `### BLOCKED:` and
  distinguish focused test evidence from unverified PostgreSQL concurrency.

## Stop Rules

- Stop if a migration, Passkey/authentication behavior, frontend, deployment,
  container, database, or production change is needed.
- Stop if a call site cannot declare precise fields without changing its public
  semantics or if a focused test failure is unrelated to this slice.
- Stop before push, remote branch deletion, Docker operation, deployment, or
  production write.

## Contract Review

`PASS / contract-approved`: the source diff and current mainline seams confirm
that declared-column writes, atomic balance adjustments, and the promo-code
counter guard are executable entirely inside the allowed repository/service
boundary. The only post-baseline dependency is a Cafe test stub receiving the
same API-key repository method signature; its scope is limited to compilation
compatibility and is explicitly allowlisted. Migration, Passkey, frontend,
container, database, and deployment paths remain unnecessary and denied.
