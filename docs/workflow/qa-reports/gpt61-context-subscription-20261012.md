### PASS: gpt61-context-subscription-20261012

## Findings

未发现本次 scoped diff 的明确缺陷或阻断 approved contract 验收的问题。本结论来自独立源码审查和本 QA 的 fresh execution，不沿用 Developer 的 PASS。

- 验收日期：2026-10-13，Asia/Shanghai。
- 独立角色：QA，未调用或委派其他 agent，未修改业务源码、业务测试或契约。
- 要求模型：`gpt-6.1-sol`。已只读核验本 QA 所属 `codex.exe`（PID 38612）的模型启动参数为 `gpt-6.1-sol`；证据为 `qa-launch-model-evidence.json`。该证据证明启动配置，不额外声称已核验服务端内部路由。
- 隔离工作树：`E:/codex-worktrees/sub2api/gpt61-context-subscription-20261012`。
- HEAD：`167ea06927643856ea1354871761d82f721f07bd`，验收前后不变。

### 两个上游行为移植及本地兼容性

1. **`07185812655ec2e0740ca76882879c0b9ae5ee0a`：500k 缺省值，而非 entitlement 覆盖。**
   - 已直接读取指定本地 Git 对象的生产 diff 和测试 diff；上游改变 generated descriptor 默认值，同时要求显式 metadata 优先。
   - 本地没有该 offline descriptor builder；本次在已有 API-key manifest adjustment 入口适配，没有引入 catalog/split-file 架构。
   - `openai_codex_models_service.go:479-533` 只对 `openai.IsGPT61SolModelSpelling` 识别的模型补齐 missing/null `context_window=500000`。显式 `0`、`272000`、`900000`、`max_context_window`、compaction 和其他模型字段保留；未知相似拼写不获兜底。既有 Responses Lite 调整保持原目标集合。
   - 新测试经过真实 `FetchCodexModelsManifest`，覆盖 `models/slug` 和 `data/id` 两个入口；后者复用既有转换，而非重写转换行为。
   - OAuth 缺失/null/较小/较大窗口的 representation 按原字节透传，没有调用 API-key adjustment；测试比较完整字符串而非 JSON 语义等价。
   - `openai_codex_models_service.go:437-463` 在 adaptation 之后判断 representation 是否变化，再生成实际 body 的 SHA256 ETag；`351-369` 缓存刷新仍使用独立的上游 validator。fresh tests 断言改写 body/hash、客户端条件命中无 body、上游 304 后缓存保留 adapted body/hash，以及无需改写的 API-key body/ETag 原样保留。

2. **`43768680d38ccdbd8fffc2cb91d3c1fc778a61c6`：按有效凭据和最终发送模型豁免。**
   - 已直接读取上游 billing/usage diff 及新增测试，核对其 credential/model 身份、旧路径、统一定价和统计传播。
   - `openai_gateway_service.go:7772-7778,8363-8374` 使用 `resolveCredentialAccount` 的有效凭据。仅 OpenAI OAuth/setup-token 加识别的 GPT-6.1 Sol 发送模型可豁免；优先 trimmed `UpstreamModel`，空时使用 `Model`，不读取客户 BillingModel、response model 或 Sub2API 客户订阅类型。
   - 默认标签真实 RecordUsage tests 覆盖映射到目标/离开目标、仅 billing/response 为目标、请求/渠道/上游三种计费来源、recognized alias、未知 suffix/prefix、API-key 和其他平台反例。未知模型夹具有真实内存价卡，不靠 pricing-unavailable/零计费掩盖错误。
   - 同时只读追踪本地普通 Responses、compact、passthrough、WebSocket 的 model 写入/返回逻辑：`UpstreamModel` 来源于发送模型，而 response model 保存在独立字段。没有为本次策略改动这些转发路径。
   - shadow 复用本地 resolver 的严格资格：仅非 shadow 的 OpenAI OAuth 母账号有效。missing repo、invalid/missing parent、lookup error、API-key/setup-token/其他平台母账号及 nested shadow 返回可见错误；测试断言无 usage 写入、余额扣款或客户订阅增量。没有扩大母账号资格或静默豁免。
   - shadow 的 usage AccountID、account multiplier、自定义统计规则匹配与结算主体仍是选中账号，母账号只提供凭据身份。

### 计价、统计和实际结算

- `billing_service.go:892-958` 新策略只控制模型价卡 long-context surcharge；输入、输出、cache read/write、priority 仍由原 `computeTokenBreakdown` 计算。有效 administrator intervals 的选价/价格没有被删除、重建或改变。
- 新默认标签矩阵有 **288 个 RecordUsage 子场景**：OAuth/setup-token/API-key × resolver 有无 × account/group gate 四组合 × `271999/272000/272001` × standard/priority × cache 有无。普通输入低于阈值、缓存补足总输入越界的场景同样在矩阵内。
- 每个矩阵场景独立按算术期望核对四个费用桶、TotalCost、ActualCost、LongContextBillingApplied、usage token 桶、实际 fake DeductBalance 后钱包余额及 AccountStatsCost。
- 标准 OAuth/setup-token 输入 `300000`、输出 `1000`：TotalCost/模型文件统计成本为 `.61`，倍率 `1.3` 后实际扣款 `.793`。同量 API-key、enabled gates：TotalCost 为 `1.215`，实际扣款 `1.5795`，long-context flag 保持 true。
- 上游 OAuth + 客户余额结算，以及上游 API-key + 客户 Sub2API 订阅结算均真实调用 RecordUsage；分别断言 DeductBalance/IncrementUsage 的调用次数、金额及结算去向。客户订阅不会授予供应商订阅豁免。
- administrator flat/interval、显式零价、interval 未命中 fallback、非 token 的 per-request/image/video、分时倍率、结果 CostOverride 优先级及两种结算去重均有 fresh tests。原模型价卡 long-context threshold/multipliers 和 administrator interval 数据还被明确断言未变。
- `account_stats_pricing.go:18-90,258-277` 仅模型文件分支收到豁免。自定义规则优先于 `ApplyPricingToAccountStats`，后者优先于模型文件；新测试覆盖两层优先级及 shadow 的选中账号规则。custom-rule/interval 匹配与计算函数没有改动。
- 共享统计和 token-cost helpers 的 optional 参数默认保持旧行为；fresh tests 直接验证未传参数的模型文件/统计成本仍为 `1.215`，完整生产 build 也证明其他平台调用方保持源码兼容。
- **保留本地原有语义，不误判为新一致性要求：** legacy 无 resolver RecordUsage 使用账号 gate；模型文件统计不使用客户 account/group gates 或请求 priority tier。priority/倍率后的客户费用不要求等于统计原始标准成本。矩阵按各自既有计价基准核验，新豁免同时作用于两条模型价卡路径。

## Executed Checks

以下均为本 QA 在当前隔离树重新执行。Go：`go1.26.3 windows/amd64`。没有使用 `-tags unit`、外部 overlay、替换 fixture 或 Developer 日志作为通过依据。

### 1. 原样 canonical acceptance

在 `backend`：

```text
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestOpenAIGatewayServiceRecordUsage|TestCalculateCostUnified_ModelLongContext' -count=1
```

**PASS**，exit 0，`0.570s`。证据：`qa-canonical-focused.log`、`qa-canonical-focused-result.json`。

这是在指定 committed baseline 加当前 scoped diff 上的首次 canonical execution；未出现共享 dirty tree 的 `gateway_beta_test.go:138` 错误。没有以共享树 drift 为由跳过 canonical 测试，也没有需要保留的 canonical 失败或使用 overlay 的理由。

### 2. 新增测试十次重测

```text
go test -json -p 1 ./internal/service -run 'TestOpenAISubscription|TestGPT61|TestFetchCodexModelsManifestGPT61' -count=10
```

**PASS**，exit 0。与两份当前测试源码的新增顶层名称逐项比对：**15 个顶层测试各执行 10 次，150 个顶层 PASS，3580 个含子测试 PASS，0 FAIL，0 skip，没有 “no tests to run”。**

证据：`qa-new-tests-count10.jsonl`、`qa-new-tests-count10-result.json`、`qa-fresh-evidence-summary.json`；summary 保留每个测试的实际执行次数。

### 3. OpenAI 包与生产完整构建

```text
go test -p 1 ./internal/pkg/openai -count=1
go build -p 1 ./...
```

均 **PASS / exit 0**。OpenAI 包 `0.038s`；build 成功无输出。证据：`qa-openai-package.log`、`qa-production-build.log`、`qa-package-build-results.json`。

### 4. 补充 fresh execution：ETag 匹配与逐测试事件

```text
go test -json -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestOpenAIGatewayServiceRecordUsage|TestCalculateCostUnified_ModelLongContext|TestCodexModelsManifestETagMatchesWeakAndMultipleValues' -count=1
```

**PASS**，exit 0。实际执行 **100 个 canonical 顶层测试 + 1 个既有 weak/multiple/wildcard ETag 测试**，共 474 个含子测试 PASS，0 FAIL。证据：`qa-focused-with-etag.jsonl`、`qa-focused-with-etag-result.json`。

### 5. 格式、精确差异、范围及 owner 漂移

根目录：

```text
git rev-parse HEAD
git status --porcelain=v1
git diff --name-only
git ls-files --others --exclude-standard
git diff --binary
git diff --check
gofmt -l backend/internal/service/account_stats_pricing.go backend/internal/service/billing_service.go backend/internal/service/openai_codex_models_service.go backend/internal/service/openai_codex_models_service_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_subscription_long_context_test.go
node E:/codex-runtime/gpt61-context-subscription-20261012/qa-summarize-evidence.mjs
```

- `git diff --check` exit 0；六个 changed Go files（包含新增未跟踪文件）的 `gofmt -l` exit 0、无输出。开始测试后与最终阶段均重新检查。
- 每一处生产 diff 可追溯到 context 缺省或 subscription 策略及其传播，没有额外生产路径、配置、定价卡、依赖、迁移或前端改动。
- 六个 owner 的初始/最终 SHA256 全部一致；tracked binary diff、Git status 与 HEAD 均保持 QA 开始状态。tracked/untracked 联合 allowlist 检查无未批准路径。
- 只读依赖 resolver、Account/type/gate、model-spelling、统计旧测试等保持基线，没有 QA 写入。
- `docs/*` 被 `.gitignore:133` 忽略，报告确实落盘，但不会显示在普通 Git status；未为此修改 ignore、暂存或提交。
- 证据：`qa-initial-snapshot.json`、`qa-scoped-initial.diff`、`qa-format-scope-results.json`、`qa-fresh-evidence-summary.json`。脚本 exit 0，明确检查每项 hash/scope/test-count 条件。

## Unverified Risks

- 本 PASS 限于 approved contract 的本地行为移植、httptest/HTTP stub、内存价卡和 fake repository 结算证据；没有证明真实供应商 entitlement、实际收费或当前模型 availability。指定上游提交是固定 Git 审查对象，不是供应商实时政策依据。
- 未执行 live DB/真实事务、容器、部署、浏览器、真实客户端或付费 provider 验收；未执行全量 tagged unit suite、全量服务测试、race 或性能测试。生产 build 不替代这些运行态证据。
- 新矩阵主要使用本地 fallback card；未知拼写反例使用显式动态内存 card。没有在本 QA 中刷新远端价格目录。
- 选中账号统计的 custom-rule interval 具体算术函数按未改动源码审查；本次新增运行断言覆盖 custom-rule 优先级、per-request 规则及选中账号归属，不声称重跑了带 unit 标签的全部既有统计规则测试。
- Controller 后续若集成到共享 dirty checkout，仍须核对被保护 owner 哈希、批准路径和目标树结果；本隔离树 PASS 不授权覆盖其他改动、提交、推送或部署。

## Recommendation

**可交 Controller 执行该 scoped diff 的最终评估及批准范围内集成。** 无需本 QA 修复。只接收本次已验收的六个 Go owner；集成前后继续保护共享树不相关改动，并重新核对 scoped diff/哈希。

本 QA 主动写入仅此报告及 `E:/codex-runtime/gpt61-context-subscription-20261012/qa-*` evidence；没有编辑业务源码、测试、contract、共享 workflow/checkout、config/dependencies 或 knowledge，没有 commit/push、provider/live DB/container/browser/global config 操作。
