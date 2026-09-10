### DONE: upstream-v024-image25-oauth-s299

# Worker Result

## Task ID
upstream-v024-image25-oauth-s299

## Status
`done`

## Summary
- `ported`: 默认目录、OAuth 管理测试选择器与前端白名单加入 Flare/Sunburst；显式账号映射继续限制可见图片模型。
- `ported`: OAuth Images Responses driver 默认改为 `gpt-5.6-luna`，支持 `SUB2API_IMAGES_MAIN_MODEL` 覆盖，且与 tool model 独立。
- `ported`: 精准识别 Codex 400 plan-gated model error；主控模型被拒绝时直接返回上游错误且不写冷却，requested 图片模型被拒绝时写入 30 分钟模型级 cooldown 并触发 failover。
- `ported`: Image 2.5 两个型号及日期版本的静态资源价与缺失远端价格 fallback，并保留显式价格优先。
- `ported`: 四份 Compose 及 `.env.example` 传递同一主控模型默认值。
- `equivalent`: 本地 `ImageInputTokens` 解析与计费字段已存在，本轮未重复修改。
- `skipped`: README 发布说明、真实 provider、容器启动/更新、部署、提交和推送。

## Changed Files
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

## Commands Run
```text
go test ./internal/pkg/openai -run 'TestDefaultModelsIncludeGPTImage25' -count=1 -> PASS
go test ./internal/service -run 'Test(OpenAIImagesResponsesDriverAndImageModels|OpenAIImagesRejectedDriverDoesNotCoolImageModel|GPTImage25.*|FetchOpenAIAccountModelsOAuth.*|AccountTestService_OpenAIImageOAuth.*)' -count=1 -> PASS
go build ./... -> PASS
npm.cmd exec vitest run src/composables/__tests__/useModelWhitelist.spec.ts -> PASS (15/15)
npm.cmd run typecheck -> PASS
docker compose -f deploy/docker-compose*.yml config --quiet -> initial standalone required env values; rerun with static placeholder POSTGRES_PASSWORD/DATABASE_HOST/DATABASE_PASSWORD/REDIS_HOST -> PASS for all four files
PowerShell ConvertFrom-Json model_prices_and_context_window.json -> PASS
git diff --check -- <Allowed Paths except worker report> -> PASS
QA fix focused service test (driver 0 cooldown; image model 1 cooldown) -> PASS
QA fix go build ./... -> PASS
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/pkg/openai
ok github.com/Wei-Shaw/sub2api/internal/service
Test Files 1 passed; Tests 15 passed
COMPOSE_CONFIG_PASS
JSON_PARSE_PASS
```

## Risks
- 未访问真实 OpenAI provider，主控模型与 Image 2.5 实际账号可用性留给独立 runtime QA/生产前验证。
- 未启动容器或部署；Compose 证据仅证明静态插值和语法有效。
- 图片缓存输入单价写入静态价格资源；本地 `LiteLLMModelPricing` 当前没有独立图片缓存字段，未越界扩展计费结构。

## Knowledge Candidates
- OAuth Codex manifest 只列 Responses driver 时，管理选择器可在 service 返回层追加本地支持的图片 tool model，同时用 `IsModelSupported` 保留账号映射限制。

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- 无。
