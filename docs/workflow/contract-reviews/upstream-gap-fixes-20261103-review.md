---
status: approved
verdict: PASS
task_id: upstream-gap-fixes-20261103
base_commit: 7ea675c020ff2c5d5e4124ac5e249c9dc65034c5
contract_sha256: b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0
review_date: 2026-11-03
reviewer_role: independent-contract-reviewer
review_execution: native-environment
---
### PASS: upstream-gap-fixes-20261103

## Findings

未发现阻断合同批准的问题。本结论仅批准下述最终合同字节对应的开发范围和验收要求，不代表实现、QA、集成或运行环境已经通过。

- 合同：`F:/mcplugins/sub2api/docs/workflow/tasks/upstream-gap-fixes-20261103.md`。
- 使用 `Get-FileHash -Algorithm SHA256` 多次独立计算原始文件哈希，结果均为 `b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0`。本报告不批准旧哈希 `a1a1ef02abce353087a09d5474572349f4a365a0c65bef9794cbaeb9c4e8564a`。
- 当前工作流为 `contract-draft`；合同中预置的批准元数据明确以独立、精确哈希绑定的 PASS 为生效条件，因此不存在单凭元数据跳过审查的授权。
- 当前矩阵将 Contract Reviewer、Developer、QA 指定为 `gpt-6.1-sol`，要求开发与 QA 独立。最终合同声明使用 native independent dispatch，且未改变矩阵模型要求；外部 CLI 认证失败不应被当作代码失败，也不授权静默换模型。本轮没有重试认证、读取凭据或发起模型调用。

### 范围与兼容性

| 项目 | 独立核对依据与判断 |
| --- | --- |
| 删除上游账号后的结算 | `usage_billing_repo.go:220-269` 中用户扣费、客户端 Key 计数、账号计数和 share ledger 位于同一效果链。合同只允许忽略账号计数中的 `ErrAccountNotFound`，真实 SQL 错误仍回滚，且要求余额、订阅、活动账号、重复结算和已有客户端 Key 行为回归，范围适当。不能把错误容忍扩大到所有错误或跳过 share ledger。 |
| 非流式图片张数 | `openai_images.go:1460-1525` 当前从 `parsed.N` 初始化，只在实际张数大于零时覆盖，确实保留了请求 n。审计的空、一、二张探针及本地保存的 `1487fa511...patch` 支持此缺口。合同明确保留适用的单张兜底，并保护流式及 gpt-image-2 拆分；未授权空响应免费或图片价格语义迁移。 |
| 管理员订阅调整 | `subscription_service.go:574-637` 当前普通读取后写绝对到期时间；已有 `withSubscriptionUpdateTx:302-321` 和仓储 `GetByIDForUpdate` 可复用。合同要求在同一事务中锁定、计算、更新到期时间与状态，提交成功后才失效缓存和回读，同时保护边界及兑换续期路径，满足本地架构。 |
| 未来 reset 期间维持暂停 | `openai_gateway_service.go:1850-1911` 先判断 reset 已到，再无条件按旧快照自愈，确有未来 reset 被忽略的缺口。合同要求绝对时间优先、相等即到期、相对时间保持原始锚点，以及 5h/7d、缺失/非法/过期 reset 和更新时间行为回归，覆盖关键边界。 |
| 账号成本长上下文开关 | `openai_gateway_service.go:8547-8552` 仅把模型豁免传入统计链；`account_stats_pricing.go:19-85,260-281` 已支持可选禁用参数，且先走自定义规则、再走 ApplyPricingToAccountStats、最后走模型价。可在允许文件中最小适配账号开关，保留现有调用兼容、OAuth 特定模型豁免及优先级。合同将成本统计与分组售价明确分开，未宣称当前缺口导致用户多扣费。 |
| Key 额度重置 UI | `KeysView.vue:2634-2646` 丢弃接口返回值，并在 await 后读当前 selection。合同要求捕获请求目标、使用返回的 quota_used/status、按当前目标 ID 决定表单同步，且列出选择切换、清空、失败、不激活响应、非零额度和后续保存的 mounted 场景，足以约束该异步状态修复。 |

审计第七项图片价格留空规则被明确推迟；resolver/billing-service 对应语义、旧断言、依赖、迁移、配置、Vertex 扩展、余额预留和 SQL 优化均未进入本次授权范围。

### 白名单与停止条件

- Developer 白名单为六个现有生产文件、三个新增测试文件和一份 worker report；足以覆盖上述六项修复。旧测试不在可写范围内，既有断言不得削弱。
- Controller 专属流程和证据文件与 Developer 范围明确分离；对 `knowledge/**` 的例外仅按合同中 Controller 自有、限定的任务追加理解，不允许 Developer 写入。
- 隔离候选、独立冻结 QA、精确 source SHA256、证据哈希和集成前目标重核构成独立门禁。合同禁止 Developer 改主仓库源码及自行 add/commit/push/merge/reset/clean。
- 最终哈希没有独立 PASS、模型/工具不可用、目标漂移、需要越界文件或连续两轮开发失败时，都有明确停止/升级规则。不得用本报告豁免后续漂移或验收失败。

### 验收充分性及执行时必须落实的细节

合同的行为标准、默认与 unit 标签测试、后端构建、mounted Vue、类型检查、lint、独立输出目录构建和 diff 检查足以进入开发，以下是这些既有条款的执行解释，不是额外扩展范围：

1. 预列 Go 正则不是全部兼容性验收。例如 `TestResolveAccountStatsCost_*`、部分图片及协议测试不被该正则完整覆盖。必须执行合同紧随其后的“识别并运行受影响既有测试”要求，记录实际测试名、叶子结果及退出码；无匹配不能作为 PASS。
2. 并发两个 +5 必须证明累积 +10，也要验证事务上下文传递、失败回滚以及 commit 后缓存失效/回读。现有 `subscription_renewal_lock_test.go` 的两个测试无 entClient，且其中累加测试顺序调用，仅重跑它们不能证明管理员入口的事务并发修复。应在允许的新测试文件中用本地事务替身/SQLmock 等提供所需证据，不依赖在线数据库，也不能以新增全局锁替代约定的数据库事务。
3. 图片空响应断言必须能拒绝请求 n=4 的回退，并与合同要求的适用单张兜底相容；不能趁此批次改免费策略。流式和 gpt-image-2 的既有拆分回归仍须独立覆盖。
4. 成本统计必须覆盖账号开关与分组售价开关的独立组合、OAuth 模型豁免、自定义规则和 ApplyPricingToAccountStats 优先级；仅测试一个 helper 的账号关案例不足以满足合同。
5. 额度重置测试必须挂载组件，确认后续保存请求中的新 status，而非仅检查源代码字符串或重写一个脱离组件的 handler。
6. 全部验证保持离线，使用已安装依赖和本地 fixtures；依赖缓存不足就按 Stop Rules 阻断，不以下载或访问 provider/DB 补齐。

## Executed Checks

- 已读取仓库 `AGENTS.md`、当前 status、Agent Matrix、spec 中本任务范围、最终完整合同和指定 gap-report。
- 已独立核对上述本地生产路径、共享价格统计调用方、仓储行锁实现、已有订阅锁测试、已删除客户端 Key 的 unit 测试和 mounted KeysView 测试基础设施。
- 已读取审计原始图片张数探针日志、结算探针日志尾部及保存的图片计数上游 patch；将其作为历史审计证据，没有把原始 FAIL 改称修复 PASS。
- 主仓库及候选 `E:/codex-worktrees/sub2api/upstream-gap-fixes-20261103` 的 HEAD 均为 `7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`。候选 tracked 状态只显示 `docs/workflow/agent-matrix.md` 修改，未显示产品文件修改。
- 已逐文件计算 main-before.json 所列 4,885 个原始文件的当前 SHA256：4,883 个一致，仅 `docs/workflow/spec.md`、`docs/workflow/status.md` 与快照不同；主仓库 index SHA256 与快照一致。流程文件漂移已显式记录，后续集成仍须精确审查并保护原有内容，不得声称全清单完全未变。
- 固定 Go 可执行文件及已安装 vitest、vue-tsc、eslint、vite 的 CLI 文件均存在；未运行安装、构建或测试，不据此保证实际验收命令成功。
- 本轮仅执行只读检查并写本报告，没有修改合同或业务源码，没有创建 agent、提交、推送、访问在线数据库/provider/network 或读取凭据。读取命令中曾有 Windows glob/猜测测试文件名不匹配，随后用实际文件定位完成相关检查；这些查找失败不构成测试执行证据。

## Unverified Risks

- 未运行本次修复的测试、编译、浏览器或真实数据库并发验证；本轮为开发前合同评审。
- 无 entClient 的测试兼容性与生产事务正确性需要分别举证，不能混为同一项通过。
- 快照只证明其中记录的文件和 index；本轮未声称扫描或保护了所有未跟踪文件，后续仍须遵守全部无关改动保护要求。
- 本报告绑定的合同任何字节变化都需要重新评审；候选冻结、QA、集成重核属于后续独立门禁。

## Recommendation

PASS。可按本合同在指定隔离候选内开展六项开发，再由独立 QA 对冻结候选验收。不得将本报告用于第七项图片价格语义变更、发布或部署授权。
