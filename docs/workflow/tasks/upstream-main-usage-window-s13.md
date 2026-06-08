# 上游合成 Sprint：`upstream-main-usage-window-s13`

## Summary

- 目标：在独立分支 `codex/upstream-main-usage-window-s13` 上合成 5h 使用窗口 `ResetsAt` 同步修复，不直接 merge `upstream/main`。
- 基线：本地 `main=b905c03a2`，上游 `upstream/main=b7cfe2462`。
- 范围：只移植 active usage poll 写回 `SessionWindowEnd`、过期窗口归零、前端 UsageProgressBar “待刷新/现在”语义修复。
- 不纳入：ops 告警指标、Select 下拉高度、代理有效期/失败回退、README/VERSION/assets/skills 改动。

## Key Changes

- 候选提交按顺序移植或手工等价移植：
  - `16bc87693`：`fix(usage): sync 5h ResetsAt to SessionWindowEnd and zero expired window`。
- 允许路径：
  - `backend/internal/repository/account_repo.go`
  - `backend/internal/server/api_contract_test.go`
  - `backend/internal/service/account_service.go`
  - `backend/internal/service/account_service_delete_test.go`
  - `backend/internal/service/account_usage_service.go`
  - `backend/internal/service/account_usage_session_window_test.go`
  - `backend/internal/service/gateway_multiplatform_test.go`
  - `backend/internal/service/gemini_multiplatform_test.go`
  - `backend/internal/service/ratelimit_session_window_test.go`
  - `frontend/src/components/account/UsageProgressBar.vue`
  - `frontend/src/components/account/__tests__/UsageProgressBar.spec.ts`
  - `frontend/src/i18n/locales/en.ts`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en/usage.ts`
  - `frontend/src/i18n/locales/zh/usage.ts`
  - `docs/workflow/tasks/upstream-main-usage-window-s13.md`
  - `docs/workflow/worker-results/upstream-main-usage-window-s13-result.md`
  - `docs/workflow/qa-reports/upstream-main-usage-window-s13-qa.md`
  - `docs/workflow/main-log.md`
- 禁止路径：
  - `backend/ent/**`
  - `backend/migrations/**`
  - `skills/**`
  - `.github/**`
  - `deploy/**`
  - `assets/**`
  - `README*`
  - `knowledge/**`
  - `docs/workflow/status.md`
  - `docs/workflow/spec.md`

## Public APIs / Interfaces

- 不新增数据库字段、migration、Ent schema 或配置项。
- 后端 repository interface 增加内部方法 `UpdateSessionWindowEnd`，仅用于写回现有 `accounts.session_window_end` 字段。
- 行为变化：
  - active usage poll 拿到 5h `ResetsAt` 后同步写回 `SessionWindowEnd` column。
  - 过期的 5h 窗口在估算用量展示时归零，避免 reset 时间已经过期但 utilization 仍显示非零。
  - 前端区分 5h usage “待刷新”和“现在”语义，降低 UI 误导。

## Test Plan

- 基础检查：
  - `git status --short --branch`
  - `git diff --check`
  - 路径审计：确认无 `backend/ent/`、`backend/migrations/`、`skills/`、`assets/`、`README*`、`knowledge/`、`docs/workflow/status.md`、`docs/workflow/spec.md` 改动。
- 后端定向测试：
  - `go test ./internal/service -run "SessionWindow|Usage|ResetsAt|RateLimit|Gateway|Gemini|Delete" -count=1`
  - `go test ./internal/repository ./internal/server -run "SessionWindow|Usage|Contract|Account" -count=1`
- 前端定向测试：
  - `corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/UsageProgressBar.spec.ts`
- 回归测试：
  - `go test ./internal/service ./internal/repository ./internal/server -count=1`
  - `corepack.cmd pnpm --dir frontend run typecheck`

## Assumptions

- 优先 `git cherry-pick -x 16bc87693`；若冲突要求 Ent/migration 或无关前端改动，停止并标记 `DEFERRED`。
- 保留本地模型市场、APIMart 计费、工单、Canvas、Chat/Image Studio、OpenWebUI 和 workflow 文档。
- 本 Sprint 不处理 `f20e6bf76` 或 `f5cecea5b`；它们作为后续独立 Sprint 评估。
