---
type: contract-review
status: approved
verdict: PASS
task_id: gpt61-context-subscription-20261012
base_commit: 167ea06927643856ea1354871761d82f721f07bd
reviewer: independent-contract-reviewer
reviewer_model: gpt-6.1-sol
last_verified: 2026-10-12
---
### PASS: gpt61-context-subscription-20261012

## Findings

未发现阻断 contract 批准的范围或可测试性问题。以下是实现与独立 QA 必须遵守的具体边界，不代表功能已经实现：

1. **上下文默认值不是 live entitlement 覆盖。** 上游 `07185812655ec2e0740ca76882879c0b9ae5ee0a` 修改 generated descriptor 的默认值，并明确保留显式上游 metadata。本地没有该 offline builder；contract 正确限定为 API-key manifest 中缺失/null `context_window` 的 500000 兜底。`openai_codex_models_service.go` 的现有转换与 adjustment 入口可以承载该行为，不需要导入 catalog 架构。显式 272000/更大值、`max_context_window`、compaction 字段和无关字段不得随默认值改变；OAuth 必须保留原始字节。
2. **ETag 必须对应实际 representation。** 本地 `fetchCodexModelsManifestUpstream` 已区分客户端 body-hash ETag 与 `upstreamETag`，缓存刷新使用后者。新增 context adaptation 必须发生在 body-hash 判定之前，不能把上游 ETag 原样用于已修改的 body。contract 的“复用现有 ETag/cache”要求可在允许的 service/test owner 内验证。
3. **凭据与发送模型是豁免身份，客户订阅不是。** 本地 `RecordUsage` 的客户订阅判定是 `Subscription != nil && Group.IsSubscriptionType()`，它只决定结算去向。上游 `43768680d38ccdbd8fffc2cb91d3c1fc778a61c6` 则按 OAuth-like 凭据和 `UpstreamModel`（为空时 `Model`）判断豁免。contract 明确区分二者，并排除 `BillingModel`、响应模型及客户别名作为豁免依据。已有 `openai.IsGPT61SolModelSpelling` 可复用，不应改成前缀匹配。
4. **shadow 身份不能猜测。** 本地 `resolveCredentialAccount` 已处理母账号读取、缺失 repo/parent、错误及非合格母账号；合格 shadow 母账号当前仅允许非 shadow 的 OpenAI OAuth。contract 的“applicable shadow parent”和“失败可见”与此相容。直接 setup-token 的豁免不意味着允许 setup-token 母账号或放宽既有 credential resolver。解析出的凭据用于豁免，不得因此替换用量归属、结算账号或既有 account/group 开关语义。
5. **豁免只关闭模型价卡附加费。** 本地 `CostInput`、`calculateTokenCost` 与旧 service-tier 路径允许追加默认关闭的内部策略，不需修改定价目录。保留管理员 flat 覆盖的价格值、有效 intervals 及未命中 interval 时的既有回退；不要为实现豁免删除 interval 或重建其价格。按次/图片/视频路径、分时及用户倍率、显式零价格、`CostOverride`、去重和结算不属于本次重构范围。
6. **账号统计必须有独立策略传播，但不破坏优先级。** `account_stats_pricing.go` 当前顺序为自定义规则、`ApplyPricingToAccountStats` 使用客户 totalCost、模型文件价格、默认回退。只有模型文件定价分支需要本次豁免控制；更高优先级规则必须保留。允许 optional policy 参数且默认保持原行为，可以避免修改 denied 的其他平台调用方。现有 `account_stats_pricing_test.go` 带 `//go:build unit`，包括 `newTestChannelServiceForStats`；新增默认标签测试必须在允许的新测试文件内自备所需 fake，不能照搬上游 unit 测试后依赖不可见 helper。

## Reviewed Scope

- Contract：`F:/mcplugins/sub2api/docs/workflow/tasks/gpt61-context-subscription-20261012.md`，当前 `draft/pending`。
- 当前共享 `docs/workflow/status.md` 与 `agent-matrix.md`：`contract-draft`；Reviewer、Developer、QA 均要求 `gpt-6.1-sol`，独立角色，不允许静默替代。
- 两个指定上游提交的原始 diff；context 提交的 generated/live metadata 边界，以及 subscription 提交的 billing、usage、account-stat 传播和相关测试。
- 本地 owners：`openai_codex_models_service.go`、`billing_service.go`、`openai_gateway_service.go`、`account_stats_pricing.go` 及对应 manifest、RecordUsage、long-context 测试。
- 只读依赖核对：`openai_agent_identity.go` 的 credential resolver、`account.go` 的 shadow/type/gate 语义、`internal/pkg/openai/constants.go` 的模型拼写识别，以及 `model_pricing_resolver.go` 的 override/interval 行为。
- 隔离 worktree：`E:/codex-worktrees/sub2api/gpt61-context-subscription-20261012`。开始时 `git status --porcelain=v1` 为空；`git rev-parse HEAD` 与指定 base 完全一致。
- allowlist 足以承载行为及默认标签 fake tests；不要求新增 schema、catalog、account helper、前端或部署改动。

## Acceptance

### 已执行的基线检查

在隔离树 `backend` 下，Go `go1.26.3 windows/amd64`：

```text
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestOpenAIGatewayServiceRecordUsage|TestCalculateCostUnified_ModelLongContext' -count=1
PASS: ok github.com/Wei-Shaw/sub2api/internal/service 0.475s

go test -p 1 ./internal/pkg/openai -count=1
PASS: ok github.com/Wei-Shaw/sub2api/internal/pkg/openai 0.038s

go build -p 1 ./...
PASS: exit 0
```

根目录 `git diff --check`：exit 0。固定已提交基线未复现 contract 所述共享树 `gateway_beta_test.go:138` 编译问题；没有使用或创建 overlay。

### 实现后独立 QA 的可验收断言

- 通过真实 `FetchCodexModelsManifest` + httptest/stub 覆盖 API-key missing/null/显式较小及较大值、识别 alias 与拒绝相似未知名称、无关字段/模型、OAuth 原始字节及 ETag；改写 body 的客户端 ETag 必须匹配实际 body，并检查条件请求/缓存保留上游 validator 的行为。覆盖原始 OpenAI `data/id` 转换入口与已有 `models/slug` 入口。
- 默认标签新增测试必须被 contract 的 regex 实际选中；报告实际测试名称/执行数量，不能用“no tests to run”作为新行为 PASS。按 contract 重复新测试十次。
- 真实 `RecordUsage` + fake repos 验证 OAuth/setup-token 与 API-key、其他平台/模型、group/account gate 组合、resolver 有无、272000 以下/等于/以上、cache token 参与边界及 priority；断言分项 cost、`LongContextBillingApplied`、总额、实际扣款/订阅增量与 account-stat cost。
- 客户余额结算 + 上游 OAuth 应豁免；客户 Sub2API 订阅结算 + 上游 API-key 不应因客户订阅获豁免。分别验证结算去向与金额，防止把两种 subscription 混为一谈。
- 发送模型映射到/离开目标、仅 BillingModel/响应模型为目标、shadow 合格母账号、缺失/错误/不合格母账号均有明确断言；失败解析不能静默获豁免或错误完成结算。
- 保留管理员 flat/interval（包括显式零价格、interval 未命中回退）、按次/图片/视频、倍率、cost override 和去重行为；同样验证 account-stat 自定义规则及 `ApplyPricingToAccountStats` 优先于模型文件分支。非 OpenAI 调用方保持签名兼容与默认行为。
- 实现后重新执行 contract 的 canonical tests、十次新测试、OpenAI 包测试、完整 build、所有 changed Go files 的 gofmt 与 diff/allowlist 检查，再交独立 QA；本 review 不替代该门禁。

## Limits

- 此 PASS 仅批准 contract 的范围与可验收性，不是实现、上线或真实模型服务的验收。新增行为尚未实现，基线测试 PASS 不证明新豁免有效。
- 所有上游行为判断依据指定的本地 Git 提交，不声称这些提交反映真实供应商当前 entitlement、可用性或官方计费政策。
- 未执行新增行为十次重测、全量 tagged unit suite、真实数据库结算、供应商调用、浏览器或容器验证；这些也不在本次 Contract Reviewer 范围。
- 本次只写该 review 文件；未修改源码、contract、共享 workflow、knowledge 或配置，未 commit/push。Controller 仍需单独更新 contract/流程，之后才能调度 Developer 与独立 QA。
