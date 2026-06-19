# Task Contract

## Task ID
apimart-task-webhook-s18

## Role
你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal
为 APIMart 异步任务接入任务完成 webhook，让 Sub2API 在任务进入 `completed` / `failed` 等终态时主动完成状态落库、视频任务结算和失败退款。首轮只覆盖 OpenAI-compatible 视频/长任务链路与 Studio Bridge 可复用的账本语义，保留普通图片同步代理现有行为。

## Success Criteria
- 新增一个 APIMart task callback 接收入口，建议路径为 `POST /api/v1/gateway/apimart/tasks/callback`；该入口不依赖用户 API key 鉴权，但必须校验配置型 secret。
- 新增或复用配置项控制 webhook 注入与校验，至少包含：启用开关、外部可访问 base URL、callback secret。配置为空时保持旧行为，不向上游注入 `webhook`。
- 提交 APIMart 视频异步任务时，在满足以下条件时向请求体注入 `webhook`：
  - account 上游 host 是 `api.apimart.ai`；
  - 本地已记录或即将记录 `task_id -> api_key/user/account/group/request_hash/reserved_cost`；
  - incoming body 未显式携带客户自己的 `webhook` 字段；
  - 配置了公网 HTTPS callback base URL 和 secret。
- 不覆盖客户已传入的 `webhook`；遇到客户自带 webhook 时继续依赖现有 `/v1/tasks/:task_id` 查询结算兜底。
- webhook payload 按 APIMart task status payload 解析 `task_id/id`、`status` 和原始 JSON；payload 结构应兼容 `status`、`data.status`、`task_id`、`id` 等现有解析风格。
- `completed/succeeded/success` 等成功终态只触发一次 usage 结算；重复回调只刷新状态，不重复扣费。
- `failed/failure/error/canceled/cancelled/timeout/expired` 等失败终态只退款一次；重复回调不重复退款。
- webhook 结算复用或收敛到现有 `SettleOpenAIVideoTaskIfTerminal` / `openai_video_tasks` 语义；不得创建另一套并行账本。
- webhook 无法解析、secret 错误、未知 task、非终态状态都返回稳定响应并写最小日志；不得泄露内部用户、API key、账号或余额信息。
- 现有 `GET /v1/tasks/:task_id` 查询触发结算逻辑保留为兜底，webhook 失败或 APIMart 重试耗尽时仍可通过查询完成结算。
- 定向单测覆盖：secret 校验、成功幂等、失败退款幂等、非终态只更新状态、未知 task 安全 ack/拒绝策略、客户自带 webhook 不被覆盖、配置缺失不注入 webhook。

## Context
- Repo: `F:/mcplugins/sub2api`
- Current branch/worktree may already contain unrelated dirty changes; do not revert or reformat them.
- Read first:
  - `docs/workflow/status.md`
  - `docs/workflow/spec.md`
  - `knowledge/tasks/current-task.md`
  - `knowledge/studio-bridge-luoye.md`
- External docs:
  - `https://docs.apimart.ai/cn/api-reference/tasks/webhook`
  - `https://docs.apimart.ai/cn/api-reference/tasks/status`
- Local facts:
  - Sub2API 是 Studio Bridge / 落叶AI余额和扣费真源。
  - `chatgpt2api` / 落叶创作台负责任务体验，但不应绕过 Sub2API 直接决定扣费。
  - APIMart 图片异步模型当前在 `openai_images.go` 内部轮询到完成后同步返回 OpenAI 图片响应，本 Sprint 不改变该兼容行为。
  - 视频任务当前已有 `openai_video_tasks`、预扣、`/v1/tasks/:task_id` 终态结算和失败退款逻辑。
- Related files:
  - `backend/internal/service/openai_videos.go`
  - `backend/internal/service/openai_videos_test.go`
  - `backend/internal/handler/openai_videos.go`
  - `backend/internal/server/routes/gateway.go`
  - `backend/internal/repository/openai_video_task_repo.go`
  - `backend/migrations/166_openai_video_tasks.sql` (read-only reference)

## Allowed Paths
- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `deploy/config.example.yaml`
- `deploy/.env.example`
- `backend/internal/service/openai_videos.go`
- `backend/internal/service/openai_videos_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/handler/openai_videos.go`
- `backend/internal/handler/openai_videos_webhook_test.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_test.go`
- `backend/internal/repository/openai_video_task_repo.go`
- `backend/internal/repository/openai_video_task_repo_test.go`
- `docs/workflow/tasks/apimart-task-webhook-s18.md`
- `docs/workflow/worker-results/apimart-task-webhook-s18-result.md`
- `docs/workflow/qa-reports/apimart-task-webhook-s18-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
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
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_test.go`
- 未在 Allowed Paths 中列出的生产配置、数据库 schema、支付、Studio Bridge、模型市场、Canvas、公共页和架构入口。

## Constraints
- 保持最小改动；不做无关重构，不格式化无关文件，不回滚当前脏工作区里的既有改动。
- 不新增数据库迁移；首轮必须复用 `openai_video_tasks` 已有字段完成幂等结算。
- 不把真实 secret、公网域名或私有 APIMart key 写入仓库；示例配置只能使用占位符。
- webhook callback base URL 必须校验为公网 HTTP(S) URL；生产建议 HTTPS。禁止自动生成或注入 `127.0.0.1`、`localhost`、内网地址作为上游 webhook。
- 不信任 webhook payload 中的金额作为扣费真源；扣费以本地提交任务时的预估/预扣和现有 billing 规则为准。上游返回 cost 只能作为可选审计字段或后续候选。
- 回调 endpoint 必须限制读取 body 大小，避免任意大 payload；日志必须脱敏，不记录 token、secret、完整 API key。
- 如果发现 webhook 结算无法从 `task_id` 找回 API key/user/account/group 上下文，停止并回 Codex 裁决，不要落一个只能更新状态但不能扣退的半实现。
- 如果为了支持客户自带 webhook 需要实现 fan-out/relay，本 Sprint 停止并另开任务；首轮只保证不覆盖客户 webhook。
- 如果需要改 Image Creator、普通图片同步响应、Studio Bridge 协议或 chatgpt2api 协议，停止并回 Codex 重拆。

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/config -run "Test.*APIMart.*Webhook|TestLoad.*Config|TestValidate" -count=1
go test ./internal/service -run "TestOpenAIGatewayServiceVideoTaskSettlement|Test.*APIMart.*Webhook|Test.*OpenAIVideo.*Webhook" -count=1
go test ./internal/handler -run "Test.*APIMart.*Webhook|Test.*VideoTask" -count=1
go test ./internal/server -run "Test.*Gateway.*Tasks|Test.*APIMart.*Webhook" -count=1

cd F:/mcplugins/sub2api
git diff --check
```

Additional audit:
```powershell
git diff --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/VERSION|frontend/src/views/public/|frontend/src/views/payment/|frontend/src/views/canvas/|frontend/src/components/studio/|frontend/src/views/admin/ModelMarket|frontend/src/views/admin/Payment|backend/internal/service/openai_images)" || echo NO_DENIED_PATHS
```

Manual review checklist:
- Confirm APIMart webhook injection is absent when config is missing, base URL is private, or request body already has `webhook`.
- Confirm callback endpoint returns stable 2xx for safe duplicate terminal callbacks and does not leak task owner details for unknown tasks.
- Confirm existing `/v1/tasks/:task_id` still settles terminal status without webhook.

## Output
- 写入 `docs/workflow/worker-results/apimart-task-webhook-s18-result.md`。
- Worker report 第一行必须为 `### DONE: apimart-task-webhook-s18`、`### FAILED: apimart-task-webhook-s18` 或 `### BLOCKED: apimart-task-webhook-s18`。
- 必须列出 changed files、commands run、test output、risks、knowledge_candidates。
- 不允许直接写长期知识库；只提交候选结论。

## Stop Rules
- 需要修改 Denied Paths、生产密钥、数据库迁移、安全边界或未授权架构入口时，停止并报告 blocked reason。
- webhook secret / base URL 规则无法从官方文档确认，或 APIMart payload 与本地解析预期冲突时，停止并要求 Codex 复核。
- 定向测试无法覆盖幂等扣费/退款时，停止，不进入 QA。
- 同一实现连续两轮失败时，停止 worker loop，回 Codex 深审 contract 或拆分任务。

## Budget
- worker_mode: `claude-bare-deepseek-v4-pro`
- qa_worker_mode: `claude-bare-deepseek-v4-pro`
- worker_model: `deepseek-v4-pro`
- max_budget_usd: `0.20`
- worktree_root: `E:/codex-worktrees`

## Worker Output
- 兼容旧脚本字段；内容同 `Output`。
