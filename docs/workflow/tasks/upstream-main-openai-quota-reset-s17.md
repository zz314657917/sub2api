## Task ID
upstream-main-openai-quota-reset-s17

## Role
你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal
把上游 `b81694929 feat(openai-quota): query + reset rate-limit credits for OpenAI accounts` 作为独立 S17 小步迁移到本地。目标是在管理员账号用量列中，为 OpenAI OAuth 账号提供上游 ChatGPT/Codex `/wham/usage` 查询和 rate-limit reset credit 消费入口，同时保留本地已有用量窗口、账号 quota 和产品定制。

## Success Criteria
- 后端新增 OpenAI quota query/reset service，复用现有 OpenAI token provider、privacy client factory 和账号代理解析，不新增数据库迁移。
- 管理端路由新增：
  - `GET /api/v1/admin/openai/accounts/:id/quota`
  - `POST /api/v1/admin/openai/accounts/:id/reset-quota`
- Handler 仅允许管理员 OpenAI OAuth 账号调用；非 OpenAI、非 OAuth、缺少 `chatgpt_account_id` / `organization_id` 时返回稳定错误。
- 前端 `AccountUsageCell` 对 OpenAI OAuth 账号展示一行相关动作：本地主动查询、上游 reset credit 次数查询、上游重置按钮；不重复渲染 5h/7d 用量条。
- 本地 modular i18n 路径补齐 `admin.accounts.openaiQuotaReset` 中英文文案，不照搬上游根 `en.ts/zh.ts`。
- 定向后端和前端测试覆盖核心行为，`git diff --check` 通过，范围审计确认未修改 denied paths。

## Context
- Repo: `F:/mcplugins/sub2api`
- Current branch: `codex/upstream-v0137-safe-patches`
- Upstream commit: `b816949291f972586cd3c37138ca741869b8a3f0`
- Read first:
  - `docs/workflow/status.md`
  - `docs/workflow/spec.md`
  - `knowledge/tasks/current-task.md`
- Related local files:
  - `backend/internal/service/openai_token_provider.go`
  - `backend/internal/service/openai_privacy_service.go`
  - `backend/internal/service/account_usage_service.go`
  - `backend/internal/handler/admin/openai_oauth_handler.go`
  - `backend/internal/server/routes/admin.go`
  - `frontend/src/components/account/AccountUsageCell.vue`
  - `frontend/src/i18n/locales/{zh,en}/admin/accounts.ts`

## Allowed Paths
- `backend/internal/service/openai_quota_service.go`
- `backend/internal/service/openai_quota_service_test.go`
- `backend/internal/service/openai_quota_unit_exports.go`
- `backend/internal/service/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/admin/openai_oauth_handler.go`
- `backend/internal/handler/admin/openai_oauth_handler_quota_test.go`
- `backend/internal/server/routes/admin.go`
- `frontend/src/api/admin/accounts.ts`
- `frontend/src/components/account/AccountUsageCell.vue`
- `frontend/src/components/account/OpenAIQuotaResetCell.vue`
- `frontend/src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts`
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
- `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/VERSION`
- `frontend/src/views/public/**`
- `frontend/src/views/payment/**`
- `frontend/src/views/canvas/**`
- `frontend/src/components/studio/**`
- `frontend/src/views/admin/ModelMarket*.vue`
- `frontend/src/views/admin/Payment*.vue`
- 未在 Allowed Paths 中列出的生产配置、数据库 schema、支付、Studio Bridge、模型市场、Canvas、公共页和架构入口。

## Constraints
- 不直接 merge/rebase `upstream/main`；仅等价迁移 `b81694929` 需要的最小实现。
- 保持本地 S15/S16 未提交改动，不回滚、不格式化无关文件。
- 后端外部请求必须使用现有 privacy client factory，尊重账号代理；不得引入新依赖。
- 前端必须适配本地 modular i18n；避免上游根 locale 文件改法。
- `/admin/accounts/:id/reset-quota` 是本地账号 quota 语义，不得复用或改变；S17 使用 `/admin/openai/accounts/:id/reset-quota`。
- 若发现需要迁移 Ent/migration、改变 token refresh 策略或修改支付/Studio/Canvas 产品面，立即停止并回 Codex 裁决。

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test -tags=unit ./internal/service -run "TestOpenAIQuota" -count=1
go test -tags=unit ./internal/handler/admin -run "TestOpenAIOAuthHandler.*Quota" -count=1

cd F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts"
git diff --check
```

Additional audit:
```powershell
git diff --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/VERSION|frontend/src/views/public/|frontend/src/views/payment/|frontend/src/views/canvas/|frontend/src/components/studio/|frontend/src/views/admin/ModelMarket|frontend/src/views/admin/Payment)" || echo NO_DENIED_PATHS
```

## Output
- 写入 `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`。
- Worker report 第一行必须为 `### DONE: upstream-main-openai-quota-reset-s17`、`### FAILED: upstream-main-openai-quota-reset-s17` 或 `### BLOCKED: upstream-main-openai-quota-reset-s17`。
- 必须列出 changed files、commands run、test output、risks、knowledge_candidates。
- 不允许直接写长期知识库；只提交候选结论。

## Stop Rules
- 需要修改 Denied Paths、生产配置、数据库迁移、安全边界或未授权架构入口时，停止并报告 blocked reason。
- 上游 API 行为无法在本地单测中稳定模拟时，保留 mock/fake 单测，不发真实 chatgpt.com 请求。
- 前端全量 Vitest 若因非本轮 Studio/Canvas/支付等旧问题失败，只记录残余风险，不扩大本 Sprint。

## Budget
- worker_mode: `main-codex-direct`
- qa_worker_mode: `main-codex-direct`
- worker_model: `gpt-5.5`
- max_budget_usd: `n/a`
- worktree_root: `F:/mcplugins/sub2api`

## Worker Output
- 兼容旧脚本字段；内容同 `Output`。
