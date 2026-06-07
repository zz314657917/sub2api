# 上游合成 Sprint：upstream-main-usage-cache-stats-s10

## Summary

- Task ID: upstream-main-usage-cache-stats-s10
- Role: Developer Worker
- Branch: codex/upstream-main-usage-cache-stats-s10
- Worktree: E:/codex-worktrees/sub2api/upstream-main-usage-cache-stats-s10
- Baseline: local main=c6fefc8c6, upstream/main=f868f7cb4
- Goal: 合成 usage stats 缓存创建/命中 token 拆分字段，不直接 merge upstream/main。

## Scope

本轮只移植 `/api/v1/usage/stats` 聚合统计的 cache creation/read token 拆分字段，以及对应 API contract 覆盖。

不纳入 frontend i18n、`skills/sub2api-admin`、Codex skill 文档、Ent/migration、deploy、knowledge、旧聚合分支残留内容、历史 deferred 大改动。

## Candidate Commits

按顺序 cherry-pick -x 或手工等价移植：

1. `029b6d61a` feat(usage): 聚合统计拆分缓存创建与命中 token
2. `7386f38cf` test(usage): API契约测试补充缓存创建/命中token字段

明确跳过：

- `0760cda92`: `DEFERRED`. frontend i18n，暂不纳入本轮。
- `9ecfc4e92` / `cb4f0015f`: `DEFERRED`. 新增 Codex admin skill 资产，单独评估，不混入后端合成。
- merge commits `8ec448a8f` / `f868f7cb4`: `SKIPPED`. 不直接 merge upstream/main。

## Allowed Paths

- backend/internal/pkg/usagestats/usage_log_types.go
- backend/internal/repository/usage_log_repo.go
- backend/internal/service/usage_service.go
- backend/internal/server/api_contract_test.go
- docs/workflow/tasks/upstream-main-usage-cache-stats-s10.md
- docs/workflow/worker-results/upstream-main-usage-cache-stats-s10-result.md
- docs/workflow/qa-reports/upstream-main-usage-cache-stats-s10-qa.md
- docs/workflow/main-log.md

## Denied Paths

- frontend/**
- skills/**
- backend/ent/**
- backend/migrations/**
- deploy/**
- knowledge/**
- .github/**
- assets/**
- README*
- docs/workflow/status.md
- docs/workflow/spec.md

## Constraints

- 不新增数据库字段、migration、Ent schema 或配置项。
- 不直接 merge upstream/main。
- 允许新增向后兼容 JSON 响应字段，但不得改变既有字段语义。
- 若 cherry-pick 触及 frontend、skills、migration、Ent schema、knowledge 或 deploy，停止该提交并记录 DEFERRED。
- 不从旧未合并聚合分支恢复内容；只基于 upstream/main 的明确候选提交做等价移植。

## Public APIs / Interfaces

- 允许 `/api/v1/usage/stats` 新增响应字段：
  - `total_cache_creation_tokens`
  - `total_cache_read_tokens`
- 保留现有字段：
  - `total_cache_tokens`
  - `total_tokens`
- `total_tokens` 仍按 `input + output + cache_creation + cache_read` 计算。

## Acceptance Commands

在 `backend/` 目录执行：

```powershell
go test ./internal/service ./internal/repository ./internal/server -run "Usage|Stats|Cache|Contract" -count=1
go test ./internal/service ./internal/repository ./internal/server -count=1
```

基础检查：

```powershell
git status --short --branch
git diff --check
```

路径审计必须确认无 denied paths 改动。

如果 repository 测试依赖外部数据库导致失败，必须记录失败原因，并至少补跑：

```powershell
go test ./internal/service ./internal/server -run "Usage|Stats|Cache|Contract" -count=1
go test ./internal/service ./internal/server -count=1
```

## Output

- docs/workflow/worker-results/upstream-main-usage-cache-stats-s10-result.md
- docs/workflow/qa-reports/upstream-main-usage-cache-stats-s10-qa.md
- docs/workflow/main-log.md 追加 S10 记录

## Stop Rules

- 任何候选要求修改 denied paths，停止该候选并记录 DEFERRED。
- 任何候选要求新增数据库字段或改变 existing contract 字段语义，停止并重新起计划。
- 测试失败必须先归因；不能在未解释失败原因时合回 main。
