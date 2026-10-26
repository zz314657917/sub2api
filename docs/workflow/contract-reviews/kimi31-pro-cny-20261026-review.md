---
type: contract-review
scope: kimi31-pro-cny-20261026
status: approved
task_id: kimi31-pro-cny-20261026
verdict: PASS
base_commit: 445d5ed3855cb07ae7a8f9761c41e2aa2e11b452
reviewer: independent-contract-reviewer
reviewer_model: gpt-6.1-sol
last_verified: 2026-10-26
contract_sha256: 15684dc0e30bf933c02f20e8648edd3c590130617265056b99e108f6c6e00d3b
approved_contract_sha256: 198edb1e1c64c370af44478177074c64a0cc1f5c2823a81843e2946b21d2a792
---

### PASS: kimi31-pro-cny-20261026

## Contract Checked

- 类型：独立合同评审；不是实现或最终 QA。
- Reviewer：独立 Contract Reviewer，按仓库 Agent Matrix 指定的 gpt-6.1-sol 角色执行。
- Draft：`docs/workflow/tasks/kimi31-pro-cny-20261026.md`
- Draft SHA256：`15684dc0e30bf933c02f20e8648edd3c590130617265056b99e108f6c6e00d3b`
- Base：`445d5ed3855cb07ae7a8f9761c41e2aa2e11b452`
- 日期：2026-10-26。
- 格式门禁补记：原 draft SHA256 仍为上列 `15684d...`，不将后续批准版本冒称原 draft。Controller 仅将 Task ID 标题调整为文档一级标题加 `## Task ID`，并将批准 frontmatter 设为 approved/PASS；这两项不改变合同实质，不改变本 PASS。已批准但尚未修标题的中间合同 SHA256 为 `0a71fd68c2834e2287c5e808629abc9a73d0a056fb34f00235aa39b20f341522`。最终标题格式修正后批准合同 SHA256 为 `198edb1e1c64c370af44478177074c64a0cc1f5c2823a81843e2946b21d2a792`；如有其他实质变化必须重审。

## Findings

未发现阻断合同进入实现的问题。六路径 allowlist 足以实现，禁止修改协议解析、已有测试及依赖的边界明确。

1. 国内标准价以人民币和站点积分数值一致为合同规则；未引入汇率。已读取保存的官方 HTML、结构化摘录、定价和缓存说明：标准模型五项价格为每百万 `13 / 68 / 1.3 / 13 / 26`；高速模型另有独立价格，不在有限别名范围。保存的 HTML SHA256 与结构化证据一致：`05b97e8b8896738c8ef77e773b691e64e5c43e90367a00a4e2b39431b48b0900`。官方页面注明 K3 系列上下文 1,048,576；合同禁止推测输出上限，并限制 max_input 必须符合现有目录语义，边界充分。
2. 有限别名、大小写、已知单层前缀和负例均可枚举；未知保留变体只能阻断默认价格，不能阻断显式管理员价格。这既避免落回旧 K3，也避免为了 fail-closed 破坏渠道/分组覆盖。
3. 合同同时约束目录、PricingService、BillingService 公共 lookup、resolver 和 legacy。陈旧动态卡必须被默认政策纠正且不得原地修改；这要求检查原始卡和连续调用，不可仅验证新增目录数值。
4. 已有 calculateTokenCost 会对所有来源调用模型策略，默认来源再调用 defaultPricingAt；新增策略须在允许的 billing_service.go 内处理来源边界。现有 resolver 自行克隆默认卡并覆盖管理员字段，无需开放 model_pricing_resolver.go。合同保留显式零、间隔、按次/图片模式，已覆盖这一具体风险。
5. 缓存标准字段已有 aggregate、5m、1h 和 normalizeCacheCreationBreakdown，合同要求无重复计费，可在新测试中验证。OpenAIUsage 只有 aggregate cache creation，没有 1h 字段；合同明确禁止将归一化 1h 测试误报为真实 Chat/Responses 1h 集成。
6. 两个真实 RecordUsage 方法及默认标签可见的 new_model_billing_gateway_20261022_test.go 内存仓库夹具均存在；可调用真实结算入口并检查 UsageBillingCommand、日志、扣费、AccountStatsCost 和映射模型归属，不需要 DB/provider 或修改既有测试。部分旧 gateway_record_usage_test.go 夹具有 unit build tag，开发不得让新默认标签测试依赖仅 unit 可见的声明。

## Executed Checks

- 只读检查主树 AGENTS.md、status.md、spec.md、Agent Matrix 和合同；当前 phase 为 contract-draft。
- 读取 P/G/E、review-and-verification 技能，并以仓库明确 Agent Matrix 覆盖全局技能中的默认模型提示。
- 检查候选 HEAD 等于合同 base；候选 git status 目前仅有受保护 go.mod/go.sum prerequisite overlay。
- 核对 Go 工具路径存在；核对两份 prerequisite 实际 SHA256 与 candidate-prerequisites.json 一致：go.mod `12bb2cd38041cc56a6e5ad6c631128fb87d96019a57cb9792d15b0686ba1c7b2`，go.sum `d878dad929cbd2cd570387fd27e9896c10dba52bf86aaa5ca07833a99ef144a5`。
- 只读追踪 PricingService.GetModelPricing、BillingService.GetModelPricing/calculateTokenCost/legacy、缓存拆分、resolver 分组/渠道覆盖、两个 RecordUsage 和既有内存结算夹具。
- 读取保存的官方原文并核对 HTML 哈希；浏览器抓取工具本次未提供可用正文，价格结论来自已保存并校验的官方响应。
- 本轮没有产品代码 diff；未运行产品测试或构建，未启动 Developer，未改变合同或 workflow 状态。

## Acceptance Interpretation

以下是原 Success Criteria 的可执行数值检查示例，不增加实现范围：

- 每项 1,000 normalized tokens：input `0.013`、output `0.068`、read `0.0013`、aggregate write `0.013`，总计 `0.0953`；0.25x actual cost `0.023825`。
- aggregate write 1,000，其中 5m 600、1h 400：write `0.0182`，其他上述项目相同，总计 `0.1005`；0.25x actual cost `0.025125`。明细与 aggregate 不得累加两次。
- 单独 1h 1,000 为 `0.026`；无明细 aggregate 1,000 为默认 5m `0.013`；混合字段、矛盾总量均应遵循既有归一化规则。
- 各有限别名需在 fallback/bundle/stale 三来源及 public lookup/resolver/legacy 对齐；stale 卡应同时包含错误价格和长上下文元数据，验证无隐式附加收费且原始共享卡不变。
- 自定义完整/部分字段、显式零、interval 和非 token 模式走真实来源，保留现有缺省继承语义；未知变体的管理员覆盖也应可用。
- 映射结算同时验证用户计费模型与最终 account 模型的默认成本各用自己的模型归属；管理员价不能意外污染默认账户成本。执行实际 Apply 并核对量化后的 debit，不能只比较 helper 输出。
- 原始宽命令必须完整执行，不能用筛掉失败的选择器代替。Controller 补充的基线执行结果为 default 186 run/pass，unit 250 run/249 pass，唯一失败 TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost（billing_service_test.go:718，pricing not found for model long-image-model）；此为转述，独立 QA 须读取原始日志并独立复现后方可认定旧失败。任何新失败、漏跑、skip 或改变既有断言均不能据此豁免。

## Gate Checks

- success_criteria_testable: yes
- allowed_paths_explicit: yes
- denied_paths_explicit: yes
- acceptance_commands_executable: yes（命令结构、工具路径及测试入口已检查；真实执行留给 Developer/QA）
- worker_model_confirmed: yes（仓库 Agent Matrix 与合同一致）
- base_commit_confirmed: yes
- openspec_traceable: not-applicable

## Unverified Risks

实现、完整测试、基线失败复现、最终文件 seal 和真实 in-memory settlement 均尚未执行；本 PASS 只批准合同可执行性。官方原始网关 1h usage 解析和真实 provider/DB 仍为明确排除项。

## Recommendation

允许 Controller 仅更新批准所需 frontmatter 并进入独立 Developer；独立 QA 门禁仍必须单独执行。若修改合同实质内容，应重新提交合同评审，不能复用此 draft seal。
