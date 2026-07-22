# Task Contract: upstream-main-redis-acl-s97

## Task ID

`upstream-main-redis-acl-s97`

## Status

`done`

## Role

Codex Planner/Generator/Evaluator. This is a behavior-level port of the
upstream Redis ACL username support; do not cherry-pick or merge upstream
history.

## Goal

Allow deployments using Redis ACL users to configure an optional Redis
username through the server config, environment variables, setup wizard, and
Compose examples while preserving the existing default-user behavior.

## Success Criteria

- `RedisConfig.Username` is loaded from `redis.username` and
  `REDIS_USERNAME`, defaulting to an empty string.
- Runtime Redis clients and setup-time Redis connection tests pass the username
  to `redis.Options`.
- Setup API validation trims the username and rejects values longer than 128
  characters without changing existing password, DB, TLS, or host validation.
- Setup wizard API types, form state, labels, placeholders, and review output
  expose the optional username in both English and Chinese.
- README, config example, `.env.example`, and all built-in Compose topologies
  document and forward `REDIS_USERNAME`.
- Existing Redis configuration without a username remains byte- and behavior-
  compatible.
- Focused backend tests, frontend typecheck/build, and diff/path gates pass.

## Verification

- `go test ./internal/config ./internal/repository ./internal/setup -count=1`
  passed, including empty/non-empty username, trim, and 128/129-character
  boundary coverage.
- `npm.cmd run typecheck` and `npm.cmd run build` passed; Vite transformed 1089
  modules.
- `git diff --check`, `git ls-files -u`, conflict-marker scan, and allowlist
  audit passed. Existing `KeysView` changes remained outside the S97 staging
  boundary.
- No Redis runtime, container, deployment, or production data operation was
  performed.

## Context

- Repo: `F:/mcplugins/sub2api`
- Local baseline: `4e6feae88ebc6c35a602642dbe0bdc2b7d16676f`
- Upstream reference: `49200d47473216e58904d01043ca802d5117215b`
- Upstream main is structurally divergent; implementation must be adapted to
  the local setup/config topology.

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/repository/redis.go`
- `backend/internal/repository/redis_test.go`
- `backend/internal/setup/handler.go`
- `backend/internal/setup/setup.go`
- `backend/internal/setup/setup_test.go`
- `frontend/src/api/setup.ts`
- `frontend/src/i18n/locales/en/setup.ts`
- `frontend/src/i18n/locales/zh/setup.ts`
- `frontend/src/views/setup/SetupWizardView.vue`
- `README.md`
- `deploy/.env.example`
- `deploy/config.example.yaml`
- `deploy/docker-compose.dev.yml`
- `deploy/docker-compose.local.yml`
- `deploy/docker-compose.standalone.yml`
- `deploy/docker-compose.yml`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-redis-acl-s97.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, generated code, lockfiles, and
  `VERSION`
- Runtime Redis/container operations, deployment, and production data changes
- Redis URL redesign, password/TLS/DB semantic changes, or unrelated setup
  validation changes
- API-key routing, billing, scheduler, account selection, payment, security
  audit, and unrelated frontend views
- `knowledge/**`, global memories, handoff, and timeline files

## Constraints

- Keep the username optional and preserve the default Redis user when empty.
- Trim setup input before validation and forwarding; do not log credentials.
- Follow the local Vue i18n setup namespace and existing form layout.
- Do not cherry-pick `49200d474` because its paths do not apply cleanly to the
  local branch.
- Do not update containers or push until final QA passes and the user has
  explicitly authorized publication.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/config -run 'TestLoadRedisUsernameFromEnvironment' -count=1
go test ./internal/repository -run 'TestBuildRedisOptions' -count=1
go test ./internal/setup -run 'TestWriteConfigFileIncludesRedisUsername' -count=1
Pop-Location

Push-Location frontend
npm.cmd run typecheck
npm.cmd run build
Pop-Location

git diff --check
git ls-files -u
```

Evaluator must audit changed paths against the allowlist, verify both locale
keys and the setup review step, and confirm no runtime/container operation was
performed.

## Output

- Implementation commit on the current feature branch after QA PASS.
- `docs/workflow/status.md` and `docs/workflow/main-log.md` updated with the
  final verdict.
- No deployment or container refresh.

## Stop Rules

- Stop if the change requires a migration, generated code, Redis URL redesign,
  or runtime/container operation.
- Stop if an empty username changes existing default-user behavior.
- Stop if setup input is not trimmed or can exceed the 128-character bound.
- Stop if unrelated upstream features or frontend views enter the diff.
- Stop after two failed focused test attempts and return to Planner.
