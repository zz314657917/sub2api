# 上游合成 Sprint：`upstream-main-ops-alert-temp-unscheduled-s14`

## Summary

- 目标：在独立分支 `codex/upstream-main-ops-alert-temp-unscheduled-s14` 上合成 Ops 告警指标补丁，不直接 merge `upstream/main`。
- 基线：本地 `main=074dc565a`，上游 `upstream/main=be0174456`。
- 范围：只移植 `account_temp_unscheduled_count` 告警指标，让临时不可调度账号进入 Ops 告警规则。
- 不纳入：代理有效期/失败回退、README/sponsors/VERSION、Ent/migration、邮件模板、DingTalk、上游模型同步、用户×平台配额、Channel Monitor API mode、前端 UI 大改。

## Key Changes

- 候选提交按顺序移植或手工等价移植：
  - `f20e6bf76`：`feat(ops): 新增 account_temp_unscheduled_count 告警指标`。
- 允许路径：
  - `backend/internal/handler/admin/ops_alerts_handler.go`
  - `backend/internal/service/ops_alert_evaluator_service.go`
  - `backend/internal/service/ops_alert_evaluator_service_test.go`
  - `frontend/src/api/admin/ops.ts`
  - `frontend/src/views/admin/ops/components/OpsAlertRulesCard.vue`
  - `frontend/src/i18n/locales/en/admin/ops.ts`
  - `frontend/src/i18n/locales/zh/admin/ops.ts`
  - `docs/workflow/tasks/upstream-main-ops-alert-temp-unscheduled-s14.md`
  - `docs/workflow/worker-results/upstream-main-ops-alert-temp-unscheduled-s14-result.md`
  - `docs/workflow/qa-reports/upstream-main-ops-alert-temp-unscheduled-s14-qa.md`
  - `docs/workflow/main-log.md`
- 兼容路径：
  - 上游原提交修改 `frontend/src/i18n/locales/en.ts`、`frontend/src/i18n/locales/zh.ts`；本地项目使用 modular i18n，落地时应改为 `frontend/src/i18n/locales/*/admin/ops.ts`。
- 禁止路径：
  - `backend/ent/**`
  - `backend/migrations/**`
  - `.github/**`
  - `deploy/**`
  - `assets/**`
  - `README*`
  - `skills/**`
  - `knowledge/**`
  - `docs/workflow/status.md`
  - `docs/workflow/spec.md`

## Public APIs / Interfaces

- 不新增数据库字段、migration、Ent schema 或配置项。
- 不新增对外 HTTP endpoint。
- 行为变化：
  - Ops 告警规则允许新 metric type：`account_temp_unscheduled_count`。
  - 该指标统计当前 `TempUnschedulableUntil` 未过期的账号数。
  - 前端告警规则编辑器展示该账号级指标和中英文文案。

## Test Plan

- 基础检查：
  - `git status --short --branch`
  - `git diff --check`
  - 路径审计：确认无 `backend/ent/`、`backend/migrations/`、`.github/`、`deploy/`、`assets/`、`README*`、`skills/`、`knowledge/`、`docs/workflow/status.md`、`docs/workflow/spec.md` 改动。
- 后端定向测试：
  - `go test ./internal/service -run "OpsAlert|TempUnscheduled|AccountTemp|RuleMetric" -count=1`
  - `go test ./internal/handler/admin -run "OpsAlert|Metric" -count=1`
- 前端检查：
  - `corepack.cmd pnpm --dir frontend run typecheck`
- 回归测试：
  - `go test ./internal/service ./internal/handler/admin -count=1`

## Assumptions

- 优先 `git cherry-pick -x f20e6bf76`；若冲突扩大到禁止路径，停止并标记 `DEFERRED`。
- 保留本地模型市场、APIMart 计费、工单、Canvas、Chat/Image Studio、OpenWebUI 和 workflow 文档。
- 本 Sprint 不处理 `af19d4432` 代理有效期/失败回退，也不处理 README/sponsors/docs-only 提交。
