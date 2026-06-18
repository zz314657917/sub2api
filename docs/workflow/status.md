---
phase: done
current_sprint: upstream-main-openai-quota-reset-s17
total_sprints: 3
pending_action: none
project_type: web
qa_mode: runtime
approval_required: false
last_verified: 2026-06-18 13:49 +08:00
---

# Workflow Status

- 当前阶段：`done`
- 当前 Sprint：`upstream-main-openai-quota-reset-s17`；本轮另行修复最近 `origin/main..HEAD` 合并后审查发现的问题。
- 当前目标：保持 S15-S17 上游小步迁移证据清晰，同时把最近主线合并的 workflow/knowledge 证据口径同步到真实触达范围。
- 当前结论：S15-S17 仍是已完成的上游小步迁移批次；但当前 `main` 之后又合入统一 API Key、APIMart 图片网关、公共页/导航/设置页等产品改动，不能再把旧的 `NO_DENIED_PATHS` 当作当前 HEAD 证据。
- 当前已确认事实：
  - 本地 `main` 与 `upstream/main` 严重分叉，直接 merge 会冲突大量 Ent、wire、网关、设置页和前端文件。
  - 本地当前主线包含 Studio Bridge / 落叶AI、支付套餐、模型市场、Canvas、工单和公共页定制；上游小步迁移 Sprint 不允许覆盖这些产品面，产品合并批次则必须单独列出真实触达范围和验证。
  - 上游 `v0.1.137` 低风险候选包括前端依赖安全、token refresh 不可重试、zstd、非 JSON/SSE 错误保留、计费兜底、thinking 协议兼容、Responses sticky hash、Haiku 探测、OpenAI responses tool probe 和 ACL 拒绝信息。
  - `b81694929` 是完整功能链，不是安全/兼容小补丁；适合独立 S17，且不需要 Ent/migration/VERSION。
  - S17 新增的上游 quota/reset 能力只挂在管理员 OpenAI OAuth 账号路径，不改变本地账号 quota reset 语义。
  - 当前 `origin/main..HEAD` 已实际触达 `backend/cmd/server/wire_gen.go`、`backend/internal/repository/studio_bridge_repo.go`、`frontend/src/views/public/ModelPlazaView.vue`、`frontend/src/components/layout/AppHeader.vue`、`frontend/src/views/user/KeysView.vue`、`frontend/src/views/admin/SettingsView.vue`，以及统一 Key、APIMart 图片网关、quota reset 相关文件。
  - 真实 UI smoke、真实 OpenAI OAuth 上游和真实 APIMart 上游仍未在本地完成；当前证据以代码级定向测试、typecheck/build 和审查为主。
- 目标验证入口：
  - `docs/workflow/tasks/upstream-main-v0137-safe-patches-s15.md`
  - `docs/workflow/worker-results/upstream-main-v0137-safe-patches-s15-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-safe-patches-s15-qa.md`
  - `docs/workflow/tasks/upstream-main-v0137-small-compat-s16.md`
  - `docs/workflow/worker-results/upstream-main-v0137-small-compat-s16-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-small-compat-s16-qa.md`
  - `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
  - `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
  - `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`
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
- 下一合法动作：完成本轮审查修复和重测；后续若继续评估上游候选，应另开 Sprint，不要把当前产品合并批次和 S17 证据混用。
- 状态推进规则：`contract-draft -> contract-approved -> build -> qa -> fix -> retest -> done`。
