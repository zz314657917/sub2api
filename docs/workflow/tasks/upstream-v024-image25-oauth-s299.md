---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v024-image25-oauth-s299
worker_model: gpt-5.6-sol
base_commit: a6993d902ff11ab7356021a1a603b43408714ce0
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-10
---

# Task Contract: Upstream v0.2.4 Image 2.5 OAuth S299

## Task ID
upstream-v024-image25-oauth-s299

## Role
你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal
按本地拓扑手工适配上游 `7ccc8a6f5` 的 GPT Image 2.5 与 OAuth Images 主控模型修复，使 Flare/Sunburst 可发现、可选择、可正确计费，并避免主控模型被拒绝时错误冷却图片模型。

## Success Criteria
- 默认模型目录、OAuth 管理测试选择器和前端 OpenAI 白名单包含 `gpt-image-2.5-flare`、`gpt-image-2.5-sunburst`，显式账号模型映射仍保持限制语义。
- OAuth Images 的 Responses 主控模型默认使用 `gpt-5.6-luna`，可由 `SUB2API_IMAGES_MAIN_MODEL` 覆盖；图片工具模型与主控模型保持独立。
- 上游拒绝主控模型时直接暴露可操作错误，不把该错误误记为图片模型限流；真正拒绝图片模型时保留既有有界 failover/冷却行为。
- Flare/Sunburst 及日期版本在远端价格缺失时使用官方当前单价：文本输入/缓存 5e-6/1.25e-6，图片输入/缓存 8e-6/2e-6，图片输出 30e-6；显式价格条目优先。
- Compose 变体传递同一环境变量默认值，不执行容器更新或部署。
- 本地已经存在的图片输入 Token 解析视为等价实现，不重复改写。

## Context
- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`
- Upstream reference: `7ccc8a6f5` from `upstream/main` v0.2.4.
- Official docs: `https://developers.openai.com/api/docs/pricing#image-generation`, `https://developers.openai.com/api/docs/guides/image-generation`.

## Allowed Paths
- `backend/internal/pkg/openai/constants.go`
- `backend/internal/pkg/openai/constants_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_openai_image_test.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/upstream_models_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_images_model_test.go`
- `backend/internal/service/pricing_service.go`
- `backend/resources/model-pricing/model_prices_and_context_window.json`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/composables/__tests__/useModelWhitelist.spec.ts`
- `deploy/.env.example`
- `deploy/docker-compose.yml`
- `deploy/docker-compose.dev.yml`
- `deploy/docker-compose.local.yml`
- `deploy/docker-compose.standalone.yml`
- `docs/workflow/worker-results/upstream-v024-image25-oauth-s299-result.md`

## Denied Paths
- `knowledge/**`
- `C:/Users/Administrator/.codex/memories/**`
- `backend/migrations/**`, `backend/ent/**`, `backend/cmd/server/wire_gen.go`
- `frontend/pnpm-lock.yaml`, `VERSION`, `README_CN.md`, `outputs/**`
- Payment, redeem, proxy, Claude, Grok, Pixel Cafe and every path not listed in Allowed Paths.

## Constraints
- 保持最小改动，不做无关重构或格式化；不覆盖当前工作树其他脏改。
- 不复制已经存在的 `ImageInputTokens` 解析或计费字段。
- 价格必须匹配本 contract 中已核对的官方值，显式静态/远端价格必须优先于 fallback。
- 不访问真实 provider，不启动或更新容器，不部署、不推送。

## Acceptance Commands
```powershell
Set-Location backend
go test ./internal/pkg/openai -run 'TestDefaultModelsIncludeGPTImage25' -count=1
go test ./internal/service -run 'Test(OpenAIImagesResponsesDriverAndImageModels|OpenAIImagesRejectedDriverDoesNotCoolImageModel|GPTImage25.*|FetchOpenAIAccountModelsOAuth.*|AccountTestService_OpenAIImageOAuth.*)' -count=1
go build ./...
Set-Location ../frontend
npm.cmd exec vitest run src/composables/__tests__/useModelWhitelist.spec.ts
npm.cmd run typecheck
Set-Location ..
docker compose -f deploy/docker-compose.yml config --quiet
docker compose -f deploy/docker-compose.dev.yml config --quiet
docker compose -f deploy/docker-compose.local.yml config --quiet
docker compose -f deploy/docker-compose.standalone.yml config --quiet
git diff --check -- <all Allowed Paths except worker report>
```

## Output
- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 worker report，首行必须为 `### DONE: upstream-v024-image25-oauth-s299`、`### BLOCKED: ...` 或 `### FAILED: ...`。
- 将上游行为逐项标为 `ported`、`equivalent` 或 `skipped`，列出 changed files、commands、risks 与 knowledge_candidates。

## Stop Rules
- 需要修改 migration、Ent、wire、支付、代理、现有脏文件或未核对的计费语义时停止并请求 Codex 裁决。
- 真实 provider、容器或部署成为完成前提时停止并报告未验证风险。

## Approved Topology Amendment
- Local admin handlers call `AccountTestService.FetchUpstreamSupportedModels`
  from `upstream_models.go`; the upstream commit's `FetchOpenAIAccountModels`
  owner does not exist in this topology. Implement OAuth picker augmentation
  in `upstream_models.go` and prove it in `upstream_models_test.go`.
- This amendment replaces the nonexistent local
  `account_test_models_test.go` path. Handler routes, API-key discovery and
  shared upstream response bodies remain unchanged.

## Budget
- worker_mode: `codex-agent-gpt-5.6-sol`
- qa_worker_mode: `codex-agent-gpt-5.6-sol`
- worker_model: `gpt-5.6-sol`
- qa_worker_model: `gpt-5.6-sol`
- max_budget_usd: `0.15`
- worktree_root: `shared-primary-worktree-by-user-override`
