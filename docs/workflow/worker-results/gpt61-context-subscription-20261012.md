### DONE: gpt61-context-subscription-20261012

## Changed Files

仅修改以下文件（均位于 `E:/codex-worktrees/sub2api/gpt61-context-subscription-20261012`）：

- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_codex_models_service_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/account_stats_pricing.go`
- `backend/internal/service/openai_subscription_long_context_test.go`（新增，默认标签）
- `docs/workflow/worker-results/gpt61-context-subscription-20261012.md`（本报告）

## Commands Run

Go 环境：`go version go1.26.3 windows/amd64`。以下测试/构建全部在隔离树 `backend` 执行；没有 overlay，也没有 `-tags unit`。

```text
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestOpenAIGatewayServiceRecordUsage|TestCalculateCostUnified_ModelLongContext' -count=1
```

- 修改前基线：PASS，`0.484s`，`developer-baseline-focused.log`。
- 最终 canonical 验证：PASS，`2.949s`，`developer-canonical-focused.log`。

```text
go test -json -p 1 ./internal/service -run 'TestOpenAISubscription|TestGPT61|TestFetchCodexModelsManifestGPT61' -count=10
go test -p 1 ./internal/pkg/openai -count=1
go build -p 1 ./...
```

- 新增测试十次重测：PASS，15 个顶层测试每个实际执行 10 次，150 个顶层 PASS、3580 个含子测试 PASS 事件、0 FAIL；非 “no tests to run”。
- OpenAI 包：PASS，`0.042s`，`developer-openai-package.log`。
- 完整 build：exit 0，`developer-build.log`（成功时无输出）。
- 十次重测原始事件及逐测试执行次数：`developer-new-tests-count10.jsonl`、`developer-new-tests-count10-summary.json`。

根目录已执行：

```text
git status --porcelain=v1
git rev-parse HEAD
git show 07185812655ec2e0740ca76882879c0b9ae5ee0a -- backend/internal/service/openai_codex_models_service.go
git show 43768680d38ccdbd8fffc2cb91d3c1fc778a61c6 -- backend/internal/service/billing_service.go backend/internal/service/openai_gateway_usage.go
gofmt -w backend/internal/service/openai_codex_models_service.go backend/internal/service/openai_codex_models_service_test.go backend/internal/service/billing_service.go backend/internal/service/openai_gateway_service.go backend/internal/service/account_stats_pricing.go backend/internal/service/openai_subscription_long_context_test.go
gofmt -l backend/internal/service/openai_codex_models_service.go backend/internal/service/openai_codex_models_service_test.go backend/internal/service/billing_service.go backend/internal/service/openai_gateway_service.go backend/internal/service/account_stats_pricing.go backend/internal/service/openai_subscription_long_context_test.go
git diff --check
git diff --name-only
git ls-files --others --exclude-standard
```

初始 worktree 干净；HEAD 始终为 `167ea06927643856ea1354871761d82f721f07bd`。`gofmt -l` 无输出、diff check 通过；tracked/untracked 路径联合 allowlist 检查通过，并单独核对被 `.gitignore:133:docs/*` 忽略的本报告。初始/最终 Go owner SHA256 留在 `developer-initial-hashes.json`、`developer-final-go-hashes.json`；最终源码 hashes 与测试后快照一致。

### 首次失败与修正留证

以下命令均真实执行，失败日志未覆盖：

```text
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifestGPT61|TestOpenAISubscriptionGPT61RecordUsage' -count=1
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifestGPT61|TestOpenAISubscriptionGPT61' -count=1
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifestGPT61|TestOpenAISubscriptionGPT61|TestGPT61DisableModelLongContextPricing' -count=1
```

- 首条 RED：两个 API-key manifest 入口缺失 `context_window`；OAuth 实际 RecordUsage 的 InputCost 为 `1.2`，期望 `.6`。日志：`developer-first-regression-failure.log`。
- 第二条 v1：新增测试误用不存在的 `AccountTypeShadow`，编译失败。按本地 `IsShadow()` 的 `ParentAccountID != nil` 判定修正夹具；覆盖 invalid parent ID、missing parent/repo、lookup error、ineligible/nested parent。日志：`developer-regression-v1.log`。
- 第二条 v2：测试渠道缺少 `StatusActive`，实际统计字段为 nil；补齐活跃渠道夹具。日志：`developer-regression-v2.log`。
- 第三条 v3：未知模型不能靠任意 fallback map 项获得价格；改用内存 `PricingService` 的显式价卡。管理员 interval 未命中实际回退模型基础价、普通 usage Create 的 inserted=false 不拥有去重门禁；依据本地 resolver/settlement 源码修正期望，改用现有 pending ownership fake 验证一次结算。仅修正测试，未扩大业务修改。日志：`developer-regression-v3.log`。
- 第三条 v4：PASS，`2.852s`，`developer-regression-v4.log`；后续补充映射来源、shadow owner、显式 API-key 原字节/ETag 测试后重新执行上述最终 canonical、十次重测及完整 build。
- 探查过程中另有 PowerShell brace/glob、错误相对路径及非数字 `Select-Object -First` 参数诊断；没有成功的越界写入，也未替换测试/生产源码来绕过编译。

所有证据位于 `E:/codex-runtime/gpt61-context-subscription-20261012/`。

## Findings

1. **Context 行为移植**：仅在 API-key manifest adjustment 中对 `IsGPT61SolModelSpelling` 识别的模型补齐 missing/null `context_window=500000`。保留显式 0/272000/900000、`max_context_window`、compaction 字段、无关字段/模型及已有 Responses Lite 调整；覆盖 `models/slug` 和 `data/id` 两条真实 Fetch 入口。没有离线 descriptor builder，不声称新增 offline catalog。
2. **Representation/缓存**：OAuth 的 missing/null/显式 context 均原字节透传，保留上游 ETag/304。改写的 API-key body 使用实际 SHA256 ETag；客户端条件请求命中且无 body，刷新仍发送原上游 validator，304 后保留改写 body/hash。无需改写的显式 API-key representation 同样保持原字节与上游 ETag。
3. **凭据/模型身份**：RecordUsage 复用 `resolveCredentialAccount`；只有 resolved OpenAI OAuth/setup-token 加最终发送 GPT-6.1 Sol 拼写可豁免。`UpstreamModel` trim 后为空才使用 `Model`。BillingModel、requested/channel mapping source、响应模型及客户订阅结算类型不能授予豁免；相似未知名称配置真实价卡作为反例。
4. **实际结算**：新增默认标签测试调用真实 RecordUsage。288 个矩阵子场景覆盖 OAuth/setup-token/API-key、有无 resolver、账号/分组 gates、271999/272000/272001、priority、cache read/write；断言四个分项费用、total/actual、usage flag、统计字段及 fake `DeductBalance` 后钱包余额。客户订阅独立验证真实 `IncrementUsage` 次数/增量；统一结算 fake 与 pending ownership 分别证明一次 monetary application。
5. **Shadow 不换账单主体**：合格母账号提供订阅身份，但 AccountID、AccountRateMultiplier、自定义统计规则匹配及结算仍使用选中账号；母账号缺失、lookup error 或不合格返回可见错误，未静默豁免或扣款。没有放宽母账号仅允许非 shadow OpenAI OAuth 的原 resolver 规则。
6. **策略范围**：`DisableModelLongContextPricing` 独立于 account/group flags，仅关闭模型价卡附加费。保留管理员 flat/interval（含显式零价、未命中回退）、按次/图片/视频、时间/用户倍率、权威 CostOverride 与去重。优先统计自定义规则，再 `ApplyPricingToAccountStats`，最后模型文件；仅最后一层应用新豁免。共享统计方法及 token-cost 方法用可选参数保持原调用兼容。

主要明确数值证据：OAuth 输入 300000/output 1000 的标准 TotalCost 与 AccountStatsCost 为 `.61`，RateMultiplier `1.3` 时实际余额扣款 `.793`；同量上游 API-key 在 enabled gates 下保持 TotalCost `1.215` 和 long-context flag。Priority、缓存桶与阈值的独立算术期望在矩阵中验证。

## Risks

- Developer 完成与本地检查通过不替代 Controller/独立 QA 最终裁决；未编写 QA report。
- 本 Developer report 已实际落盘，但仓库忽略 `docs/*`，不会显示在普通 `git status`；Controller 收集/集成证据时需显式处理，本 worker 未修改 ignore 或暂存它。
- 只使用本地指定 Git 提交、httptest/HTTP stub、内存价卡和 fake repositories；未验证真实供应商 entitlement、真实数据库事务或部署运行。
- 保留既有非豁免语义：无 resolver 的旧 RecordUsage 路径只使用账号 gate；模型文件账号统计原本不读取 account/group billing gates 或请求 tier。本次没有顺带改变它们，矩阵明确断言这些差异。
- 本地有效 interval 未命中时继续使用原模型基础价而非配置 flat token 字段；测试根据原 resolver 语义验证，没有改动管理员规则。
- 未执行全量 tagged suite；canonical 默认标签 focused suite、新增十次测试、OpenAI 包及完整 build 已通过。

## Contract Compliance

- 编辑前已读取本任务 contract，frontmatter 为 `approved/PASS`；已读取共享树指定 independent review，frontmatter/首行均为 PASS。
- 独立 Developer 按要求执行；未调用或委派其他 agents，未做模型替代/全局配置操作。
- 仅四个获准生产 owner、两个获准测试文件及本 Developer report 有工作树改动；外部 supplemental evidence 仅写任务专属 runtime 目录。
- 共享 `F:/mcplugins/sub2api` 仅只读指定评审，没有业务写入；未修改 contract、共享 workflow、knowledge、frontend、定价目录/卡、context 最大值、配置模板、依赖或迁移。
- 无 cherry-pick/merge/rebase、commit/push、真实 provider/database、container、browser 或 dependency install；没有 overlay。
- 最终建议：可交 Controller 和独立 QA 按 approved contract 验收，不声称已上线。
