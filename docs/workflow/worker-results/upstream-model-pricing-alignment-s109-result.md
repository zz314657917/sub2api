### DONE: upstream-model-pricing-alignment-s109

# Worker Result

## Task ID

`upstream-model-pricing-alignment-s109`

## Status

`done`

## Summary

- 完成 LiteLLM 图片输入价、渠道 `image_input_price`、Claude 名称归一、GLM-5.2 兜底价、OpenAI OAuth 图片 tool usage、usage log/API/UI 图片输入输出对账的本地适配。
- 修复 split OAuth 非流式路径仍读取 `response.usage` 以及多图累加器漏计 `ImageInputTokens` 的问题。
- 修复账号成本统计与可用渠道 fallback 漏计图片输入价、generic hosted tool usage 语义偏差，以及长上下文极值整数溢出。
- composite 别名计费经审计缺少本地前置链，按 contract stop rule 移出 S109，未引入无关架构。

## Changed Files

- `backend/migrations/193_channel_image_input_price.sql`
- `backend/migrations/194_usage_log_image_input_tokens.sql`
- `backend/internal/handler/{admin/channel_handler.go,available_channel_handler.go}`
- `backend/internal/handler/dto/{mappers.go,types.go}`
- `backend/internal/repository/{channel_repo_pricing.go,usage_log_repo.go,usage_log_repo_request_type_test.go,migrations_schema_integration_test.go}`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/service/{account_stats_pricing.go,billing_service.go,channel_available.go,channel.go,channel_service.go,gateway_service.go,model_pricing_resolver.go,openai_gateway_service.go,openai_images.go,openai_images_responses.go,pricing_service.go,usage_log.go}`
- `backend/internal/service/*_test.go` 中 S109 对应定向测试
- `frontend/src/api/{admin/channels.ts,channels.ts}`
- `frontend/src/components/admin/channel/{PricingEntryCard.vue,types.ts}`
- `frontend/src/components/admin/usage/{UsageTable.vue,__tests__/UsageTable.spec.ts}`
- `frontend/src/components/channels/SupportedModelChip.vue`
- `frontend/src/i18n/locales/{en,zh}` 下 S109 渠道和 usage 文案
- `frontend/src/types/index.ts`
- `frontend/src/utils/{imageUsage.ts,__tests__/imageUsage.spec.ts}`
- `frontend/src/views/admin/{ChannelsView.vue,GroupsView.vue}`
- `frontend/src/views/user/{UsageView.vue,__tests__/UsageView.spec.ts}`

## Commands Run

```text
focused OpenAI image/pricing service tests -> PASS
broad pricing/billing/image-input/record-usage service tests -> PASS
account stats, available-channel fallback, hosted tool usage and max-int regressions -> PASS
usage repository focused tests -> PASS
migration schema integration -> PASS (38.498s)
targeted /api/v1/usage contract -> PASS
frontend typecheck -> PASS
focused Vitest -> PASS (3 files / 33 tests)
frontend production build -> PASS (1089 modules)
scoped ESLint -> PASS
gofmt -d -> no output
git diff --check -> PASS
conflict marker scan -> no matches
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/repository
ok github.com/Wei-Shaw/sub2api/internal/server
Test Files 3 passed (3)
Tests 33 passed (33)
vite: 1089 modules transformed; built successfully
```

## Risks

- 未对真实 OpenAI OAuth 图片上游发请求，tool usage 语义由 recorder/SSE fixture 覆盖。
- 未做登录态浏览器 smoke、部署或容器更新。
- 全表 `TestAPIContracts` 有与 S109 无关的既有 settings 快照漂移；本轮 `/usage` 子契约已独立通过。
- `go test -tags=unit ./internal/service` 被 `origin/main` 既有的重复 helper、旧 billing 测试签名和缺失 Grok runtime helper 阻断，不列为 PASS。
- `unit` build-tag 的 service 聚合仍受既有 `stringPtr`、billing 旧签名和 Grok helper 编译漂移阻断；GLM-5.2 与长上下文图片费用已补默认 build-tag 测试并实际通过，未把 `[no tests to run]` 或编译失败写成 PASS。
- 最终 QA 修复了显式 token 图片行误判、OpenAI usage 非法数值边界和长上下文测试倍率语义，最终三路只读复审无阻断 finding。

## Knowledge Candidates

- 本地 monolithic usage log 新增两列后固定数据参数为 53，timezone 参数为 `$54`；上游拆分文件的参数编号不能机械照抄。
- OAuth 图片 split 聚合除替换 parser 外，还必须累计 `ImageInputTokens`。

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `yes: composite prerequisite chain absent, correctly moved out of scope`

## Blocked Reason

- 无。
