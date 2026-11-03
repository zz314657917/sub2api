### PASS: upstream-gap-fixes-20261103-test-addendum

---
status: approved
verdict: PASS
task_id: upstream-gap-fixes-20261103-test-addendum
base_commit: 7ea675c020ff2c5d5e4124ac5e249c9dc65034c5
contract_sha256: a51b65f646f67a0a38dfda32bbdd63b2a1729c07b17ab5f20618cb79c8a9bb5c
parent_contract_sha256: b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0
review_date: 2026-11-03
reviewer_role: independent-contract-reviewer
review_execution: native-environment
---
### PASS: upstream-gap-fixes-20261103-test-addendum

## Findings

未发现阻断增补合同批准的问题。此次授权是两个旧测试对父合同已批准行为的定向对齐；生产实现、原 R1 九路径及三个新增测试均冻结。当前证据支持这 17 个失败属于旧语义冲突，未发现本增补授权掩盖真实缺陷。

独立计算合同最终原始字节 SHA256 为 `a51b65f646f67a0a38dfda32bbdd63b2a1729c07b17ab5f20618cb79c8a9bb5c`；父合同仍为 `b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0`。本报告批准的文件是 `F:/mcplugins/sub2api/docs/workflow/tasks/upstream-gap-fixes-20261103-test-addendum.md`，不是对任何后续变更自动批准。

### 两个测试修改的合理性

1. **Scheduler stale snapshot 测试。** 原文件 `openai_account_scheduler_test.go:917-947` 明确把“三小时前快照 + 一小时后 reset”定义成无条件放行，要求选择 35901；候选实际选择 35902，原始日志在第 947 行失败。这与父合同第 4 项“有效未来 reset 期间维持暂停”直接相反。增补要求保留原顶层名称和 issue 背景，将一个旧场景拆成“无已知未来 reset 时选 35901”和“有未来 reset 时选 35902”，并保留实际选中账号 ID、优先级差和独立 fixture。它保留自愈的正向保护，也增加已耗尽账号不得提前回池的反向保护，未用删除场景或放宽断言换绿。
2. **Settlement matrix 账号统计。** 原文件 `openai_subscription_long_context_test.go:188-194` 在 APIKey 且 n>272000 时始终给账号统计乘长上下文倍率，忽略 accountGate。候选新增开关传递符合父合同第 5 项，故 account=false 的这组预期需要对齐。原始日志的 16 个失败全部落在 `assertSettlement` 第 92 行的 `AccountStatsCost` 精确容差比较，全部是 APIKey/account=false/n=272001；它们恰好覆盖 resolver、group、tier、cache 各两个值的 16 种组合。增补仅允许统计倍率谓词增加 accountGate 和修改对应说明，保留完整循环、case 名、用户收费、凭据、倍率、缓存 token、账户归属和结算断言。该调整不会把 group gate 或请求 tier 混入模型文件账号成本，也不改变 OAuth 豁免。

已静态核对候选 `openai_gateway_service.go`：reset 逻辑只在无有效 reset 时允许 stale 自愈；账号统计只额外合并账号长上下文关闭条件，仍调用原有统计优先级链。`account_stats_pricing.go` 与 R1 冻结内容一致。上述生产行为在本次增补中不允许再改。

### 范围、冻结及停止门禁

- Developer 仅能修改候选中的两个指明测试文件，并向原 worker report 追加 R2 历史；没有对生产、helper、其他测试或主仓库业务文件的写权限。
- 已独立重算 R1 九路径 SHA256，全部与 `candidate-freeze.json` 相符，包括未改变的 account_stats_pricing.go 和三个新测试。新增授权不能改变这九个哈希。
- 两个旧测试当前在 main 与候选中均字节一致，仍是尚未对齐的原文。本报告没有提前修改或接受未来 R2 diff。
- R2 绑定十一条路径、其中十条实际变化的计算成立：原九路径中一条为不变的 account_stats_pricing.go，另增两条旧测试。
- 禁删、禁 skip/xfail、禁改名绕开选择、禁减少循环、禁扩大容差，以及新失败/身份漂移/越界即停的规则明确。原失败日志和 R1 freeze 必须保留，R2 证据另起名称；Author 与独立 QA 分离。
- 父合同的图片价格语义暂缓、无关改动保护、无网络/真实 DB/provider、无提交发布部署等约束继续有效。

## Executed Checks

### 原始日志独立重算

直接解析四份 `.stdout.log` 的 Go JSON 事件，按完整测试名去重，排除有子测试的父节点后统计叶子，并核对对应 receipt 退出码；没有依赖汇总 JSON 的数字作最终计算。

| 原始运行 | 顶层终态数 | 通过叶子 | 失败叶子 | skip | 退出码 |
| --- | ---: | ---: | ---: | ---: | ---: |
| main-original-regression-default | 207 | 619 | 0 | 0 | 0 |
| main-original-regression-unit | 248 | 670 | 0 | 0 | 0 |
| controller-candidate-existing-default | 207 | 602 | 17 | 0 | 1 |
| controller-candidate-existing-unit | 248 | 653 | 17 | 0 | 1 |

四份日志均可逐行解析，无非 JSON 行。候选两标签失败集合完全相同：scheduler 1 项、上述矩阵统计 16 项，未在此选择内发现其他失败。示例矩阵错误为期望 `1.1030039999999999`、实际 `0.554002`、容差仍为 `1e-10`，不是客户扣费断言失败。

### 验收可执行性

- 对 `existing-regression-selection.json` 的 20 文件逐一读取 main/候选原文：全部字节一致，原 248 个顶层声明均存在，未发现清单缺失、额外声明或正则漏选。20 文件中包含一个没有 Test 声明的 benchmark 文件，不能把文件数当测试数。
- 默认标签实际执行 207 个顶层测试，unit 标签为 248；这是已记录基线的标签差异，不能要求 default 也出现 248 个终态，更不能把缺测或 skip 当作绿色。原 248 声明须保留并由 unit 清单验证。
- 单个 scheduler 叶子拆成两个，在其他 case 不变时，验收目标 default 620 / unit 671 正确。须同时核对原声明/叶子身份、零 skip 和测试进程退出码，不仅检查通过总数。
- 合同两条定向命令准确命中两个顶层 owner；完整选择仍使用原 `existing-regression-regex.txt`，不能以定向两条命令替代整份选择。固定工具链、readonly modules 与本地依赖路径沿用父合同，已有双标签日志证明当前选择可收集和执行。
- 已检查现有私有 `developer-baseline-overlay.json` 及其实际运行 receipt：Go `-overlay` 已用于本候选的原生产源码验证；其五个映射目标原文字节与当前原 main 对应文件一致。由此确认合同要求的原生产 overlay 在当前环境有可用基础，不需要改主仓库或访问外部服务。
- **R2 RED 必须继续加载对齐后的测试，只替换回原生产源码。** 两种意图都应独立暴露：scheduler 的 future-reset 子例仍选原 35901 而失败、account-off 的长上下文统计仍按旧倍率而失败。编译失败、仅一个 owner 失败、映射回旧测试或无测试可跑都不能满足该 RED 门禁。正向自愈、account-on、OAuth 等控制项应保留可归因结果。
- R2 GREEN 必须在不使用原生产 overlay 的冻结候选上双标签通过，之后仍由原独立 QA 审精确两测试 diff、重跑既有 probes 和完整选择；生产冻结相同不能替代这些验收。

本轮执行的是源文件与原始证据读取、哈希、清单/事件重算和静态路径审查，没有新运行测试或生成对齐测试。未改合同、源码、既有日志或 freeze；未调度 agent，未排查 CLI 网关，未读取凭据。

## Unverified Risks

- 增补尚未实现；本轮不存在 R2 GREEN 或对齐测试对原生产源码的新增 RED 执行结果。本报告不能替代开发后精确 diff 及独立 QA。
- 原四个已归因限制继续公开：长上下文图片输入成本旧失败、两个 quota API unit 失败（含 nil dereference panic）、Vue create-query 旧失败。本轮未重新裁定其完整归因；增补未授权修改其 owner 或豁免新失败。20 文件选择全绿不等于父合同广域 Go/Vue 命令全绿。
- 未验证真实 DB/provider、线上并发或浏览器；本次仅为测试对齐合同门禁。

## Review Evidence

外部证据：`E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/addendum-review-evidence.json`。

证据 SHA256：`14c4cb457f2a715b53780ccc4daa3ffd4d9c9885b8e54d441ed2cb7b72c39802`。

其中记录了最终合同与父合同哈希、九路径重新计算、20 文件清单及原文哈希、四份日志独立叶子统计、17 个失败的原始断言位置、overlay 原文核对，以及所读关键证据的 SHA256。

## Recommendation

PASS。允许原 Developer 按本增补的两测试精确权限生成 R2；保留 R1 身份和失败证据，以原生产 overlay 的两个预期 RED、候选双标签完整选择 GREEN、九路径哈希不变和原独立 QA 作为后续必需门禁。任何额外失败、源码漂移或超出精确两测试语义的改动都须停止并回 Controller。
