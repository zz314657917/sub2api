---
phase: done
current_sprint: upstream-main-v0137-postfixes-s19
total_sprints: 5
pending_action: none
project_type: web
qa_mode: runtime
approval_required: false
last_verified: 2026-06-18 19:25 +08:00
---

# Workflow Status

- 当前阶段：`done`
- 当前 Sprint：`upstream-main-v0137-postfixes-s19`
- 当前目标：继续把上游 `v0.1.137` 中少量后端小补丁按 contract 小步迁入，只覆盖 OpenAI failover body、Anthropic cooldown、account repo 参数/refresh candidates 修复。
- 当前结论：S19 已实现并通过定向 QA；S18 APIMart task webhook 仍保留为已起草的产品候选，但本轮没有实现 S18。
- 当前已确认事实：
  - 本地 `main` 与 `upstream/main` 严重分叉，直接 merge 会冲突大量 Ent、wire、网关、设置页和前端文件。
  - 本地当前主线包含 Studio Bridge / 落叶AI、支付套餐、模型市场、Canvas、工单和公共页定制；上游小步迁移 Sprint 不允许覆盖这些产品面，产品合并批次则必须单独列出真实触达范围和验证。
  - 上游 `v0.1.137` 低风险候选包括前端依赖安全、token refresh 不可重试、zstd、非 JSON/SSE 错误保留、计费兜底、thinking 协议兼容、Responses sticky hash、Haiku 探测、OpenAI responses tool probe 和 ACL 拒绝信息。
  - `b81694929` 是完整功能链，不是安全/兼容小补丁；适合独立 S17，且不需要 Ent/migration/VERSION。
  - S17 新增的上游 quota/reset 能力只挂在管理员 OpenAI OAuth 账号路径，不改变本地账号 quota reset 语义。
  - 当前 `origin/main..HEAD` 已实际触达 `backend/cmd/server/wire_gen.go`、`backend/internal/repository/studio_bridge_repo.go`、`frontend/src/views/public/ModelPlazaView.vue`、`frontend/src/components/layout/AppHeader.vue`、`frontend/src/views/user/KeysView.vue`、`frontend/src/views/admin/SettingsView.vue`，以及统一 Key、APIMart 图片网关、quota reset 相关文件。
  - 真实 UI smoke、真实 OpenAI OAuth 上游和真实 APIMart 上游仍未在本地完成；当前证据以代码级定向测试、typecheck/build 和审查为主。
  - APIMart task webhook 适合补强视频/长任务可靠结算；Sub2API 仍是 Studio Bridge / 落叶AI余额和扣费真源，`chatgpt2api` 不应绕过 Sub2API 决定扣费。
  - 当前 APIMart 图片异步模型仍通过 `openai_images.go` 内部轮询后同步返回，S18 不改变普通 `/v1/images/generations` 兼容行为。
  - S19 已明确跳过 OpenAI image failover、token refresh retry amplification、OAuth promo signup、scheduler outbox dedup/cleanup、cyber policy、channel monitor jitter、Claude OAuth system prompt blocks 和 migration-heavy 链路。
- 目标验证入口：
  - `docs/workflow/tasks/apimart-task-webhook-s18.md`
  - `docs/workflow/tasks/upstream-main-v0137-postfixes-s19.md`
  - `docs/workflow/tasks/upstream-main-v0137-safe-patches-s15.md`
  - `docs/workflow/worker-results/upstream-main-v0137-safe-patches-s15-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-safe-patches-s15-qa.md`
  - `docs/workflow/tasks/upstream-main-v0137-small-compat-s16.md`
  - `docs/workflow/worker-results/upstream-main-v0137-small-compat-s16-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-small-compat-s16-qa.md`
  - `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
  - `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
  - `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`
  - `docs/workflow/worker-results/upstream-main-v0137-postfixes-s19-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-postfixes-s19-qa.md`
- 已执行验证：
  - `go test -tags=unit ./internal/service -run "TestGetFallbackPricing_FamilyMatching|TestGetModelPricing_DoubaoEmbeddingVisionImageInputRate|TestCalculateCost_DoubaoEmbeddingVisionDifferentialInput|TestHandleNonStreamingResponse|TestHandleStreamingResponse_SSEErrorEvent|TestIsNonRetryableRefreshError|TestResolveThinkingProtocol|TestThinkingFilters|TestNormalizeChineseLLMThinking|TestApplyThinkingEnabledFallback|TestExtractOpenAIReasoningEffortFromBody" -count=1`
  - `go test ./internal/service -run "TestOpenAIGatewayServiceRecordUsage_MissingPricingRecordsZeroCostUsageLog|TestExtractOpenAIReasoningEffortFromBody|TestIsNonRetryableRefreshError|TestResolveThinkingProtocol|TestThinkingFilters|TestNormalizeChineseLLMThinking|TestApplyThinkingEnabledFallback" -count=1`
  - `go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback" -count=1`
  - `go test ./internal/repository -run "Test.*Decompress|Test.*HTTPUpstream" -count=1`
  - `go test ./internal/pkg/apicompat -count=1`
  - `go test -tags=unit ./internal/service -run "TestParseGatewayRequest_ResponsesInput|TestGenerateSessionHash_ResponsesInputProducesHash|TestDecideResponsesProbeSupportRequiresFunctionCallOn2xx|TestOpenAIResponsesProbePayloadForcesFunctionCall|TestSelectResponsesProbeModelUsesMappedUpstreamModel|TestProbeOpenAIAPIKeyResponsesSupportPersistsToolCapability" -count=1`
  - `go test -tags=unit ./internal/handler -run "TestDetectInterceptType_MaxTokensOneHaiku|TestSendMockInterceptResponse_MaxTokensOneHaiku" -count=1`
  - `go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionUsesGenericMessageForBlacklistDenial" -count=1`
  - `go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback|Test.*GenerateSessionHash|TestParseGatewayRequest" -count=1`
  - `go test -tags=unit ./internal/service -run "TestOpenAIQuota" -count=1`
  - `go test -tags=unit ./internal/handler/admin -run "TestOpenAIOAuthHandler.*Quota" -count=1`
  - `go test ./internal/service -run "^$" -count=1`
  - `go test ./internal/handler/admin -run "^$" -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts"`
  - `git diff --check`
  - S15-S17 当时的 denied-path audit returned `NO_DENIED_PATHS`；当前 `origin/main..HEAD` 已包含后续产品合并，不再适用该结论。
  - lockfile scan confirmed no `form-data@4.0.5` / `form-data: 4.0.5` remains.
  - `go test ./internal/service -run "Test.*Failover.*Body|Test.*Cached.*Body|Test.*Anthropic.*Window|Test.*Cooldown|TestOpenAI.*Images" -count=1`
  - `go test ./internal/repository -run "Test.*Account.*List|Test.*Refresh.*Candidate|Test.*Temp.*Unscheduled|TestAccountsToService" -count=1`
  - `go test ./internal/server -run "Test.*APIContract" -count=1`
  - `go test -tags=unit ./internal/service -run "TestOpenAIGatewayService_HandleFailoverSideEffects_DoesNotRereadResponseBody|TestOpenAIGatewayService_Forward_FailoverReparsesCachedBodyForNextAccount|TestHandleUpstreamError_AnthropicWindowLimitPreemptsTempUnschedRule" -count=1`
  - `go test -tags=unit ./internal/repository -run "TestAccountsToService_LargeActiveAccountSetDoesNotExceedPostgresParameterLimit" -count=1`
  - S19 denied-path audit returned `NO_DENIED_PATHS`.
- 下一合法动作：若回到产品主线，审查 S18 APIMart webhook contract；若继续上游合成，另开 Sprint。
- 状态推进规则：`contract-draft -> contract-approved -> build -> qa -> fix -> retest -> done`。
