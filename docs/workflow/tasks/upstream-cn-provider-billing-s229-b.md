---
task_id: upstream-cn-provider-billing-s229-b
phase: contract-approved
base: de62dd8d6bfcbbe88f94873e23f8756fb4aaafee
---

# Task Contract

## Role

Codex Controller/Generator；使用独立实现 worktree 和独立 QA worktree，手工适配上游行为，不直接 cherry-pick `10c8b7020`。

## Goal

从上游 `10c8b7020` 独立移植 CN 计费候选链修复：CN 账号的客户端 `claude-*` 候选只有在分组或渠道显式配置定价时才可参与计费；全部候选被过滤后仍写入 zero-cost usage log，并以 `ErrModelPricingUnavailable` 归类告警。非 CN 行为保持不变。

## Success Criteria

- Kimi/Zhipu/DeepSeek 账号无显式 Group/Channel 定价时过滤所有 `claude-*` 候选，保留非 Claude 候选及其顺序。
- CN 账号存在显式 Group 或 Channel 定价时保留对应 `claude-*` 候选；过滤只针对 CN 账号，OpenAI/Grok/Anthropic 行为不变。
- 候选列表为空或没有非空候选时，`calculateOpenAIRecordUsageCost` 返回可由 `isUsagePricingUnavailableError` 识别的错误。
- `RecordUsage` 在 CN 候选全被过滤时仍写入一条 zero-cost usage log，不丢 usage，不向上游发起请求。
- focused billing tests x10、完整 `internal/service` 回归、server compile、格式、scope、provenance、冲突/index 和主工作区保护门禁通过。

## Context

- Repo: `F:/mcplugins/sub2api`
- Base: `main@de62dd8d6` after S229-A local integration
- Upstream: `upstream/main@938f1868a`；source `10c8b70203feac8fbd744d386af6600aa87c3837`
- 上游 `backend/internal/service/openai_gateway_usage.go` 在本地不存在；本地计费用量 owner 是 monolithic `backend/internal/service/openai_gateway_service.go`。
- 本地已有 pricing-unavailable zero-cost 记录分支，但空候选错误尚未包装为 `ErrModelPricingUnavailable`。

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_cn_billing_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-billing-s229-b-result.md`
- `docs/workflow/qa-reports/upstream-cn-provider-billing-s229-b-qa.md`

## Denied Paths

- `backend/internal/service/ratelimit_service.go` 及所有 403 policy owner
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/service/openai_gateway_responses_anthropic_native.go`
- `backend/internal/service/openai_gateway_messages_anthropic_native.go`
- partial-result usage 提交、client-disconnect drain/finalize、stream handler、composite routes、frontend、migrations、dependencies、provider calls、push、deployment、containers、共享/生产数据库
- 所有用户已有 dirty/untracked 文件、`knowledge/**`、`outputs/**`，除本合同明确的 workflow report 外不得修改

## Constraints

- 过滤条件必须复用现有 `resolveOpenAIChannelPricing`，显式 Group/Channel 来源均可放行；不得按模型字符串硬编码定价金额。
- 只过滤 CN 账号的 `claude` 子串候选；空白候选继续跳过；非 CN 候选顺序和值保持原样。
- 不能改变正常已有的定价解析、倍率、图片计费或 usage log 字段语义。
- 不新增依赖、迁移、网络请求或真实 provider/数据库操作。
- Implementation 必须先在隔离 worktree 完成并经 Controller review，再由独立 QA worktree 验收；QA 不得依赖 Controller 自述。

## Acceptance Commands

```powershell
go test ./internal/service -run "TestFilterCNProviderBillingModelCandidates|TestCalculateOpenAIRecordUsageCost_EmptyCandidatesIsPricingUnavailable|TestOpenAIGatewayServiceRecordUsage_CNFilteredCandidatesWriteZeroCostLog" -count=10
go test ./internal/service -count=1
go test ./cmd/server -run "^$" -count=1
gofmt -d internal/service/openai_gateway_service.go internal/service/openai_gateway_cn_billing_test.go
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor 10c8b70203feac8fbd744d386af6600aa87c3837 upstream/main
```

## Protected Main Baseline

- `main@de62dd8d6`，`origin/main@a865d8b6e`，不 push。
- 用户 dirty patch IDs：backend `d665008e...`；account modal `5d316e5b...`；tutorial `a07a7c33...`；knowledge files `2abee47d...`（`knowledge/00-start-here.md` + `knowledge/05-current-focus.md`，dispatch 前重新计算并记录）。`knowledge/tasks/current-task.md`、`docs/workflow/**` 是 Controller 流程文件，会随门禁更新，不纳入用户 patch 保护值。
- 六个未跟踪 tutorial migration/test 文件 SHA256：
  - `226_update_image_model_tutorial_domains_to_cc.sql`: `D7EDF11F2D7F5A1BCE0D6D10CE7BF50C6FEC35D8F01AD46E93DE8526DC4DB839`
  - `227_format_image_model_tutorial_parameters.sql`: `A426D11E76E029D4CF6A6BD1606E2894FB59029425888B6AB40E59F73CF61`
  - `228_expand_image_model_tutorial_details.sql`: `854BBC7BCEDC47682FB78FD315EBB88678AF7488E0B5E1D9EA8C64CF82C5`
  - `229_format_image_tutorial_curl_examples.sql`: `C9676B553D0526142311AB5CFD90317937F3031AD91F72BEF302646608F488D1`
  - `image_model_tutorial_curl_format_test.go`: `84C47AB0226587D7D098A3D98786081CE8A98860AECBE8EBC2FCC4BCF0D85C27`
  - `image_model_tutorial_details_test.go`: `6D07FA3C646BDD7CB90E3B7039FE933F815163F70050CEC8EA3324A89442DC54`
- `outputs/**` 必须保持未跟踪且内容不变。

## Output

- 一个隔离 worktree implementation commit；Controller result 首行为 `### DONE: upstream-cn-provider-billing-s229-b`。
- 独立 QA report 首行为 `### PASS: upstream-cn-provider-billing-s229-b`，或明确 `FAIL/BLOCKED`。

## Stop Rules

- 任何 denied path、依赖、迁移、provider、数据库、容器、部署、远端、冲突或 unmerged-index 变化立即停止。
- 不得将 403、断开排水、响应终态修复或其他 `10c8b7020` slice 混入 S229-B。
- 若显式 Group/Channel 定价被过滤、非 CN 行为变化或 zero-cost usage 未写入，判定 FAIL。
