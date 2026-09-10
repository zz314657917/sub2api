---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v024-selective-reliability-s298
worker_model: gpt-5.6-terra
base_commit: f013405bc38531826b4776772599907909b3de89
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-10
---

# Task Contract: Upstream v0.2.4 Selective Reliability S298

## Task ID
upstream-v024-selective-reliability-s298

## Role
你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal
按本地拓扑手工适配上游 `dc6b318c3` 与 `6aabbdf54` 的两个独立可靠性修复，不 merge、rebase 或 cherry-pick 上游历史。

## Success Criteria
- 用户 Usage 页面加载 API Key 筛选项时，按每页 100 条继续加载至 API 返回的最后一页；空页可安全停止，既有错误处理保留。
- OpenAI Responses 的原生和 passthrough 流在已向客户端输出语义内容后收到 `response.failed` 时，保留终止事件且记录可查询的上游诊断（账号、上游 request ID、失败负载/信息）；不得触发重放或改变既有 failover 决策。
- 为两项行为补充或调整聚焦回归测试。

## Context
- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`
- Upstream references: `dc6b318c3`, `6aabbdf54` from refreshed `upstream/main`.
- Local stream owner is consolidated in `backend/internal/service/openai_gateway_service.go`.

## Allowed Paths
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `docs/workflow/worker-results/upstream-v024-selective-reliability-s298-result.md`

## Denied Paths
- `knowledge/**`
- `C:/Users/Administrator/.codex/memories/**`
- `backend/migrations/**`, `backend/ent/**`, `backend/cmd/server/wire_gen.go`
- `deploy/**`, `VERSION`, `frontend/pnpm-lock.yaml`, `outputs/**`
- Payment, pricing, model catalog, proxy, account, Claude, Grok, Pixel Cafe and all files not listed in Allowed Paths.

## Constraints
- 保持最小改动，不做无关重构或格式化。
- 不回滚、覆盖或吸收当前工作树中已有改动；现有脏路径全部在本 contract 外。
- 将上游两份 Responses stream owner 适配到本地单一 owner，不复制上游文件拓扑。
- 不访问真实 provider、数据库、容器或部署环境，不提交、不推送。

## Acceptance Commands
```powershell
Set-Location backend
go test ./internal/service -run 'Test.*(Response|Stream|Failed|Usage)' -count=1
go build ./...
Set-Location ../frontend
npm.cmd exec vitest run src/views/user/__tests__/UsageView.spec.ts
npm.cmd run typecheck
Set-Location ..
git diff --check -- frontend/src/views/user/UsageView.vue frontend/src/views/user/__tests__/UsageView.spec.ts backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_test.go
```

## Output
- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 worker report，首行为 `### DONE: upstream-v024-selective-reliability-s298`、`### BLOCKED: ...` 或 `### FAILED: ...`。
- 列出 changed files、commands run、关键结果、risks 和 knowledge_candidates。

## Stop Rules
- 需要修改 Denied Paths、数据库迁移、模型/定价、支付、部署或外部运行资源时停止并报告。
- 无法保持已输出流不重放的边界时停止并请求 Codex 裁决。

## Budget
- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
