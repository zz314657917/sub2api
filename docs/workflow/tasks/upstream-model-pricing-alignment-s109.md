<!-- codex:pge-contract -->
# S109 上游模型定价与图片输入计费对齐 Contract

## Task ID

`upstream-model-pricing-alignment-s109`

## Role

`Generator`

## Goal

手工迁移 `upstream/main` 中适用于本地架构的模型定价与图片 token 计费修复，覆盖 LiteLLM、渠道定价、OpenAI OAuth 图片 usage、usage log/API/UI 对账，并保留本地已有模型价与账号行为。

## Success Criteria

- LiteLLM `input_cost_per_image_token` 可解析；只有图片价、没有普通 token 价的条目不会参与普通 token 计费。
- 渠道 token 定价支持 `image_input_price`，flat 与 interval 路径均保持未配置时回退文本输入价。
- GLM-5.2 兜底价与官方 GLM-5.1 价一致；Claude 点号/连字符模型名可命中同一渠道价。
- OpenAI usage 解析 `input_tokens_details.image_tokens` / `prompt_tokens_details.image_tokens`；generic hosted tool usage 保留独立图片 token，OAuth 图片在合法 `tool_usage.image_gen` 存在时原子采用该 usage，畸形值回退 `response.usage`。
- 多图 split OAuth 请求累计文本/图片输入和输出 token，不丢失 `ImageInputTokens`。
- `CostBreakdown.InputCost` 只含文本输入，`ImageInputCost` 单独记录，`TotalCost` 与拆分前总额一致；试用 overage 比例缩放包含图片输入费用。
- 账号成本统计的 LiteLLM 与渠道自定义 token 定价按图片输入价拆分，未配置时回退文本输入价。
- migration、repository、DTO/API 和管理员/用户 usage 页面完整暴露图片输入/输出 token 与费用。
- Go 定向与宽范围测试、migration schema、usage API contract、前端 33 项测试、typecheck、production build、lint、格式与 diff 门禁通过。

## Allowed Paths

- `backend/migrations/193_channel_image_input_price.sql`
- `backend/migrations/194_usage_log_image_input_tokens.sql`
- `backend/internal/handler/admin/channel_handler.go`
- `backend/internal/handler/available_channel_handler.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/repository/channel_repo_pricing.go`
- `backend/internal/repository/migrations_schema_integration_test.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/account_stats_pricing.go`
- `backend/internal/service/account_stats_pricing_test.go`
- `backend/internal/service/channel_available.go`
- `backend/internal/service/channel_available_test.go`
- `backend/internal/service/channel.go`
- `backend/internal/service/channel_service.go`
- `backend/internal/service/channel_service_test.go`
- `backend/internal/service/gateway_record_usage_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/model_pricing_resolver_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_images_test.go`
- `backend/internal/service/pricing_service.go`
- `backend/internal/service/pricing_service_test.go`
- `backend/internal/service/usage_log.go`
- `frontend/src/api/admin/channels.ts`
- `frontend/src/api/channels.ts`
- `frontend/src/components/admin/channel/PricingEntryCard.vue`
- `frontend/src/components/admin/channel/types.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `frontend/src/components/channels/SupportedModelChip.vue`
- `frontend/src/i18n/locales/en/admin/availableChannels.ts`
- `frontend/src/i18n/locales/en/admin/channels.ts`
- `frontend/src/i18n/locales/en/availableChannels.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/i18n/locales/zh/admin/availableChannels.ts`
- `frontend/src/i18n/locales/zh/admin/channels.ts`
- `frontend/src/i18n/locales/zh/availableChannels.ts`
- `frontend/src/i18n/locales/zh/usage.ts`
- `frontend/src/types/index.ts`
- `frontend/src/utils/imageUsage.ts`
- `frontend/src/utils/__tests__/imageUsage.spec.ts`
- `frontend/src/views/admin/ChannelsView.vue`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `docs/workflow/**`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- 账号导入、OAuth enrichment、账号套餐类型和调度行为。
- composite 平台/别名完整前置链；本地缺少 `PlatformComposite`、`composite_platform.go` 与对应 gateway billing 架构。
- 本地 GPT-5.6、Grok、DeepSeek、Kimi、MiniMax 独立价格实现。
- OpenAI 长上下文默认策略、alpha/search 按次计费、部署、容器和生产配置。

## Constraints

- 不整体 merge `upstream/main`，只手工移植已证明适用的语义。
- migration 编号使用本地连续编号 `193`、`194`；repository 适配本地 monolithic `usage_log_repo.go`。
- i18n 使用本地拆分的 `en/usage.ts`、`zh/usage.ts`，不回写上游 monolithic dashboard locale。
- 图片输入费用从输入费用中拆分，但不得改变 `TotalCost` / `ActualCost` 总额。
- 主工作树现有用户改动不得修改、暂存或回滚。

## Acceptance Commands

```text
go test ./internal/service -run "TestParseOpenAIImagesSSEUsageBytes|TestBoundedJSONNonNegativeInt|TestMergeOpenAIUsage_CarriesImageInputTokens|TestOpenAIGatewayServiceForwardImages_OAuthPassesNAndReturnsAllImages|TestOpenAIGatewayServiceForwardImages_OAuthStreamingTransformsEvents|TestParsePricingData_ParsesImageInputTokenPrice|TestGetModelPricing_ImageOnlyLiteLLMEntryFallsBackForTokenBilling|TestResolve_WithChannelOverride_TokenWithIntervals" -count=1
go test ./internal/service -run "Test.*Pricing|Test.*Billing|Test.*ImageInput|Test.*OpenAIImages.*Usage|Test.*RecordUsage" -count=1
go test ./internal/service -run "TestDefaultBuild_AccountStats|TestDefaultBuild_ProportionalTokenCountAvoidsIntegerOverflow|TestDefaultBuild_LongContextTotalAvoidsIntegerOverflow|TestDefaultBuild_AvailableChannelPricingPreservesImageInputPrice|TestExtractOpenAIUsageFromJSONBytes_MergesHostedImageGenToolUsage|TestExtractOpenAIUsage_BoundsHostedImageToolTokenFields" -count=1
go test ./internal/repository -run "TestUsageLogRepository|TestPrepareUsageLogInsert|TestScanUsageLog" -count=1
go test -tags=integration ./internal/repository -run "TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate" -count=1
go test -tags=unit ./internal/server -run "TestAPIContracts/GET_/api/v1/usage" -count=1
corepack.cmd pnpm run typecheck
corepack.cmd pnpm exec vitest run src/utils/__tests__/imageUsage.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts
corepack.cmd pnpm run build
corepack.cmd pnpm exec eslint <S109 frontend paths>
gofmt -d <S109 Go paths>
git diff --check
```

## Output

- `docs/workflow/worker-results/upstream-model-pricing-alignment-s109-result.md`
- `docs/workflow/qa-reports/upstream-model-pricing-alignment-s109-qa.md`

## Stop Rules

- composite 修复若需要引入平台、gateway 或约 2,000 行前置架构，停止并移出本 Sprint。
- 任一 migration 与本地编号或现有 schema 冲突，停止合入并重新编号/设计。
- 图片费用拆分改变 `TotalCost`、丢失 token 或触碰 denied paths 时判定 FAIL。
- 完整 API contract 若只失败于与 S109 无关的既有 settings 快照，必须用 usage 子契约和基线对照，不得伪报全表 PASS。
