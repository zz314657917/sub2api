---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: usage-billing-multiplier-breakdown-s275
worker_model: gpt-5.6-terra
base_commit: 3b8a710a176af1a1b26d6ea01ddf747d30e4e7a4
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-08-31
---

# Usage Billing Multiplier Breakdown S275

## Task ID

`usage-billing-multiplier-breakdown-s275`

## Role

你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

修复 APIMart 图片成本换算被误展示为分组/用户倍率的问题。在不改变任何扣费、余额、历史账本或现有 `rate_multiplier` 存储语义的前提下，为 usage API 提供“计价倍率”和“余额换算倍率”拆分，并让用户端、管理端及导出使用清晰口径。

## Success Criteria

- APIMart 图片 usage 的现有 `total_cost`、`actual_cost`、`rate_multiplier` 和实际余额扣减保持不变；`gpt-image-2-official` 与 APIMart OpenAI API Key 账号图片请求仍使用现有 `7 * 1.2` 换算。
- 用户和管理员 usage DTO 新增向后兼容的 `pricing_rate_multiplier` 与 `balance_conversion_multiplier`：普通记录分别等于现有倍率和 `1`；命中 APIMart 图片换算时，前者为现有综合倍率除以 `8.4`，后者为 `8.4`。
- 拆分判定与现有扣费触发边界一致：仅 `image_count > 0`，且实际账号为 `api.apimart.ai` OpenAI API Key，或 usage 的请求/计费/上游模型候选包含 `gpt-image-2-official` 时拆分；非图片记录不因模型名称误拆。
- 用户端和管理端的分组徽章、tooltip、详情抽屉及导出将 `pricing_rate_multiplier` 作为可见计价倍率，并在 `balance_conversion_multiplier != 1` 时单独显示/导出“余额换算”；旧后端未返回新字段时继续回退现有 `rate_multiplier`。
- APIMart `1x -> 8.4x` 与 `2x -> 16.8x`、APIMart 账号任意图片模型、普通 OpenAI 图片、非图片 official 模型、缺少账号关联的 official 历史记录均有回归覆盖。

## Context

- Repo: `F:\mcplugins\sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, `docs/workflow/agent-matrix.md`
- Current composite rule: `backend/internal/service/apimart_gpt_image2_pricing.go`
- Current usage projection: `backend/internal/service/usage_log.go`, `backend/internal/handler/dto/mappers.go`
- Current UI consumers: `frontend/src/views/user/UsageView.vue`, `frontend/src/components/admin/usage/UsageTable.vue`
- Protected existing changes: API-key route breaker/auth files, `backend/internal/service/admin_service.go`, Pixel Cafe admin view, and `outputs/**`.

## Allowed Paths

- `backend/internal/service/usage_log.go`
- `backend/internal/service/usage_log_multiplier_breakdown_test.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/mappers_usage_test.go`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/zh/usage.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `frontend/src/views/admin/UsageView.vue`
- `docs/workflow/worker-results/usage-billing-multiplier-breakdown-s275-result.md`

## Denied Paths

- `backend/internal/service/apimart_gpt_image2_pricing.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/**`
- `backend/config/**`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、生产配置、容器、部署或数据文件

## Constraints

- 保持最小改动：这是只读展示投影，不重写历史 usage，不新增 schema/migration，不改变余额扣减或成本计算路径。
- 保留 `rate_multiplier` 字段和原值，避免破坏旧客户端、统计、审计公式和 CSV/XLSX 消费者；新字段是附加的明确口径。
- 后端拆分逻辑必须集中在 service 层并由 DTO mapper 复用，不在 Vue 中根据模型字符串硬编码 `8.4`。
- APIMart 账号触发的历史记录只能使用 usage 查询时关联到的账号配置；测试和代码注释必须承认账号配置变化后的历史投影限制，不伪装为不可变账本快照。
- UI 不显示内部上游品牌说明；可见文案使用“计价倍率”“余额换算”。不增加功能说明卡片或破坏现有紧凑 tooltip/详情布局。
- 不回滚或覆盖任何既有改动；不得执行 commit、push、数据库、容器、部署、真实 provider 请求或网络抓取。

## Acceptance Commands

从 `F:\mcplugins\sub2api\backend` 执行：

```powershell
go test ./internal/service -run 'TestUsageRateMultiplierBreakdown' -count=10
go test ./internal/handler/dto -run 'TestUsageLogFromService.*MultiplierBreakdown' -count=10
go test ./internal/service ./internal/handler/dto
go test ./cmd/server -run '^$'
gofmt -w internal/service/usage_log.go internal/service/usage_log_multiplier_breakdown_test.go internal/handler/dto/types.go internal/handler/dto/mappers.go internal/handler/dto/mappers_usage_test.go
```

从 `F:\mcplugins\sub2api\frontend` 执行：

```powershell
cmd.exe /d /s /c "corepack.cmd pnpm exec vitest run src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm run typecheck"
```

从仓库根目录执行：

```powershell
git diff --check -- <allowed paths>
git diff --name-only --diff-filter=U
```

还必须人工核对：APIMart `total_cost * pricing_rate_multiplier * balance_conversion_multiplier == actual_cost`；普通记录不出现余额换算行；用户/管理员导出同时保留可解释的计价、换算和综合倍率口径；changed paths 精确属于 allowlist。

## Output

- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 `docs/workflow/worker-results/usage-billing-multiplier-breakdown-s275-result.md`。
- 报告第一行必须是 `### DONE: usage-billing-multiplier-breakdown-s275`、`### BLOCKED: usage-billing-multiplier-breakdown-s275` 或 `### FAILED: usage-billing-multiplier-breakdown-s275`。
- 报告列出 changed files、commands run、关键测试摘要、scope/provenance 检查、未验证风险和 `knowledge_candidates`。

## Stop Rules

- 如果准确拆分必须新增数据库字段、修改扣费路径、重写历史记录，立即 `BLOCKED`，交回 Codex 裁决，不得自行扩大范围。
- 如果必须修改 denied path、已有 API 的 `rate_multiplier` 值、生产配置或外部状态，立即停止。
- 任何测试失败记录真实命令与归因；不得删除测试、放宽断言或吸收用户已有改动。
- 不提交产品 commit，不 push，不更新容器或共享数据。

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
