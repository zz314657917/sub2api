---
status: approved
review_verdict: PASS
task_id: upstream-key-limits-20261003
worker_model: gpt-6.1-sol
base_commit: af190154bc62e8d2a66db8f19018ae6a7727a9d6
spec_ref: docs/workflow/tasks/upstream-key-limits-20261003.md
---

## Task ID
upstream-key-limits-20261003

## Role
Independent Developer, followed by independent QA per repository Agent Matrix.

## Goal
Adapt upstream 9ecb34082 and 017bbcb90 API Key creation safeguards to local owners. Keep ordinary routing, default keys, Cafe managed keys and existing error-attempt counter intact.

## Success Criteria
- Config api_key_create.max_active_per_user defaults 200, max_per_user_per_hour defaults 60; zero disables each independently, negative config rejected.
- Existing CountByUserID counts all nondeleted keys including inactive/expired/managed keys. This upstream-compatible count-before-create check is deliberately a soft guard, NOT a transactional hard count cap under concurrent creation. Do not add schema/DB locking or claim hard-cap guarantees.
- Hourly attempt counter is separate from existing failed-custom-key counter. Atomic increment with fixed one-hour TTL; never extend on subsequent attempts. Counts above limit denied. Concurrent ordinary calls cannot get more than limit successful counter admissions while Redis is healthy. Later persistence failure consumes its slot; update/delete never refund new hourly counter. Redis unavailable/nil cache retains upstream fail-open behavior; report that limitation.
- Preserve existing failed-custom-key counter methods and all current reset semantics, route/group/Cafe validation.
- Limits are checked after validation/key generation and before Create persistence. Custom and generated keys share quota. EnsureInitialKey calls Create and is counted; repeated bootstrap when keys exist consumes nothing; current auth error-handling unchanged.
- Cafe managed keys created directly by repository remain exempt from hourly gate. Existing managed lifecycle code untouched. Static call path review and existing targeted default/Cafe regression evidence required; do not claim new limits cover every repository creator.
- No new dependencies/migrations/wire/frontend changes.

## Allowed Paths
- `backend/internal/config/config.go`
- `backend/internal/config/api_key_create_config_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_create_limits.go`
- `backend/internal/service/api_key_create_limits_test.go`
- `backend/internal/repository/api_key_cache.go`
- `backend/internal/repository/api_key_create_count_test.go`
- `deploy/config.example.yaml`
- `docs/workflow/worker-results/upstream-key-limits-20261003-result.md`

## Denied Paths
- `frontend/**`
- `backend/migrations/**`
- `backend/ent/**`
- `knowledge/**`
- Other product paths and shared F:/mcplugins/sub2api writes. No worker commits/push/deploy.

## Constraints
Add narrow optional create-counter interface if needed so existing APIKeyCache doubles remain source-compatible. Production cache must implement it and tests must prove constructor exposes it. Missing optional capability follows nil-cache fail-open policy and must not be hidden in reports. Preserve current guard order and all local fields. Work in E:/codex-worktrees/sub2api/upstream-integration-20261003 after controller dispatch only.

## Acceptance Commands
In backend: config tests for defaults/zero/negative, service actual Create and EnsureInitialKey tests for custom/generated/count/hourly/no-refund/failure/nil/error paths, repository miniredis or task-owned Redis tests for fixed TTL/per-user isolation/concurrent increments. Use existing miniredis dependency if present, never connect shared Redis. Run `go build ./...` and `git diff --check`. If full unit package is baseline compile-blocked, document exact errors and use explicit production GoFiles plus focused test files (no production stubs/deletion); require nonzero executed tests. Baseline GroupID task must pass isolated QA first; baseline source must not be edited by this worker.

## Output
Worker report starts DONE/BLOCKED/FAILED and lists precise changes, commands, executed tests, limits and scope. Independent QA produces separate report before main integration.

## Stop Rules
Stop on new unrelated baseline errors, required denied path edits, missing dependencies or concurrent task changes. Never silently expand into hard-cap DB concurrency, account routing, payment or security redesign.
