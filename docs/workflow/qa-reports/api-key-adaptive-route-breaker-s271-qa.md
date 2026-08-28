### PASS: api-key-adaptive-route-breaker-s271

# QA Report

## Task ID
api-key-adaptive-route-breaker-s271

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/api-key-adaptive-route-breaker-s271.md`

## Evidence
- diff reviewed: `yes` (candidate worktree `E:/codex-worktrees/sub2api/api-key-adaptive-route-breaker-s271`)
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run from `E:/codex-worktrees/sub2api/api-key-adaptive-route-breaker-s271/backend`:
```text
go test ./internal/repository -run "APIKeyRouteBreaker" -count=10 -> PASS
go test ./internal/service -run "APIKey(RouteBreaker|RouteCooldown|Routing)" -count=10 -> PASS
go test ./internal/server/middleware -run "APIKeyRoute(Cooldown|Breaker)|ShouldCooldownAPIKeyRoute" -count=10 -> PASS
go test ./internal/service -count=1 -timeout 120s -> PASS (65.190s)
go test ./internal/server/middleware -count=1 -> PASS
go test ./cmd/server -run '^$' -count=1 -> PASS
go test ./internal/repository ./internal/service ./internal/server/middleware -count=1 -> FAIL only at account_repo_upstream_billing_probe_update_test.go:559, expected 32 actual 34 (contract-dispositioned baseline)
gofmt -d <S271 Go files> -> PASS (no output)
git diff --check -> PASS
git diff --name-only -> S271 allowed modified files only
git diff --cached --name-only -> empty
git ls-files -u -> empty
git status --short -> S271 allowed modified/untracked files only
```

## Manual Checks
- Redis key uses group ID, normalized routing scope, and SHA-256 digest of exact normalized requested model; scope/model isolation tests pass.
- Lua scripts use Redis `TIME` and one `KEYS[1]`; lazy first-failure state, 3-failure threshold, 30s/2m/10m/30m backoff, 24h retention, and 30m inactive streak are covered by miniredis tests.
- Half-open probe lease is single-use; generation/token checks reject stale tail requests; concurrent failure test opens once. Request-local skipper memoizes decisions and releases unselected probes.
- Healthy acquire returns generation 0 without creating Redis state; Redis errors fail open. Success clears existing state and does not create absent state.
- Middleware distinguishes original upstream status from mapped response status, excludes ordinary business 4xx, preserves streaming classification, and lets final success clear state. Existing per-Key cooldown remains separate.
- Pinned-account and no-default-bypass behavior are covered by routing tests.
- Integration precondition: S271 worktree predates S272 and still contains the old middleware route-count fast path at `api_key_auth.go:386`. S271 QA does not treat this as a failure; integration owner must hand-port S271 while preserving S272's nil-only guard and run `TestS272` x10.

## Findings
- 未发现明确的 S271 实现问题。
- 联合 repository 命令的唯一失败是 contract 已 disposition 的范围外 sqlmock fixture 漂移：`account_repo_upstream_billing_probe_update_test.go:559`，精确错误 `expected 32, actual 34`。

## Unverified Risks
- 未运行真实 Redis、provider traffic、数据库、容器或部署，均被 contract 排除；Lua 状态机由 miniredis 覆盖。
- S272 integration guard 尚未在该旧 worktree 验证，需由 integration owner 在主线手工合并后验收。

## Bug Owner Recommendation
`integration-owner`

## Root Cause
- `none`（S271 QA）；联合包失败为范围外 `test-bug`/fixture drift，非本 Sprint 修复项。

## Retest Scope
- 集成到当前 main 后：保留 S272 nil-only authorization guard，执行 `TestS272` x10；随后重跑本报告全部 focused、完整 service/middleware、server compile 和 scope/index gates。

## Knowledge Promotion
- `none`
