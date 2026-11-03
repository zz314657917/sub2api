# 六项上游修复：额外独立提交前代码评审

## Findings

**结论：需修复，当前冻结版本不建议提交。发现 1 项 P1 和 1 项 P2。**

### P1：订阅调整脱离调用方退款事务，外层回滚仍保留扣减

- 变更位置：`backend/internal/service/subscription_service.go:576-578`；事务实现：同文件 `302-321`。
- 现有调用链：`PaymentService.finalizePendingRefundSuccess` 在 `payment_refund.go:496-505` 创建退款事务，`518` 将其 `txCtx` 传给 `applyRefundFinalDeduction`，后者在 `579` 调用 `ExtendSubscription`。因此该方法确实存在带外层事务的生产调用方。
- 本次修改无条件进入 `withSubscriptionUpdateTx`；该 helper 使用根 `entClient.Tx(ctx)` 新建事务，并以新事务覆盖上下文。订阅读取与写入在这个内层事务提交，随后才回到退款收口。旧代码的 repository 操作通过 `clientFromContext` 使用调用方事务，没有这次独立提交。
- 当后续 `markRefundOkTx` 或退款事务提交失败时，外层回滚无法撤回已经提交的订阅天数扣减。退款重试还可能再次扣减；订阅缓存也在外层退款尚未提交时被失效。
- **本轮定向实测**：真实候选方法及真实 ent 事务/commit hooks，配以私有事务 driver/repository 替身。调用方先开事务，调整 `-5` 天，再回滚调用方事务。候选记录 `begins=2, commits=1, writes_use_outer=false`，到期日从 `2026-11-23T12:00:00Z` 留在 `2026-11-18T12:00:00Z`，断言失败；相同探针仅 overlay 回固定 HEAD 的原 `subscription_service.go`，记录 `begins=1, commits=0, writes_use_outer=true`，回滚后保留原到期日，断言通过。
- 修复方向：有调用方事务时，让锁读、计算和写入加入该事务，避免擅自提交；缓存失效跟随实际事务所有者成功提交。无外层事务时保留本次新增的锁读和事务原子性。修复后覆盖外层回滚和成功提交，保留已有并发 `+5/+5`、失败回滚及缓存生命周期断言。
- 证据：专属目录中的 `caller_transaction_probe_v2_test.go`、`v2-candidate-overlay.json`、`v2-baseline-overlay.json`、`v2-candidate.stdout.log`、`v2-baseline.stdout.log` 及各自 receipt。

### P2：重置额度覆盖同一 Key 尚未保存的启用/停用选择

- 变更位置：`frontend/src/views/user/KeysView.vue:2641-2644`。
- 状态选择器在 `1005-1011` 绑定 `formData.status`；用户可以先修改启用/停用，再点击编辑框中的额度重置。当前实现只检查仍选中同一 ID，就无条件用响应状态覆盖表单。
- 可复现例：服务端 Key 仍为 `active`，用户在表单选了 `inactive` 但尚未保存；额度重置返回原有 `active`，表单被改回 `active`，后续保存无法完成用户刚选择的停用。反向 `inactive → 表单 active` 同样被清除。
- **本轮定向实测**：提取实际 `resetQuotaUsed` 原函数，仅移除 catch 的 TypeScript 类型注解，在隔离 Node VM 中执行。候选两种未保存选择均被覆盖；固定 HEAD 的原函数保留这两种选择；上游 `a65d476f25ddcd5ef961f18b80fc640cb7341ecd` 的修复也保留两者，同时通过 `quota_exhausted → active` 控制场景。
- 原因是本地适配省略了上游“响应状态与原 Key 状态确实不同才同步表单”的保护。应在覆盖 `targetKey.status` 前保存或比较旧状态；始终消费响应的 `quota_used/status`，只有实际发生服务端状态转换且仍选中目标 ID 时同步表单，保留不同 Key 的防污染逻辑。
- 证据：`key-form-status-probe.mjs`、`key-form-status-comparison.json`。这是实际源函数的定向执行，**不是 mounted Vue 或浏览器验收**。既有新增 8 项 mounted 测试未覆盖“同一 Key、服务端状态不变、表单已有未保存选择”。

其余四项生产修复、本次两份旧测试语义对齐和新增测试中，未发现其他明确阻断问题。上述发现不要求修改既有不可变报告，也不等同于重新执行合同或整套 QA。

## Executed Checks

### 本轮实际执行

1. 读取仓库 `AGENTS.md`、指定 `review-and-verification/SKILL.md`、父合同、已批准测试增补、专属任务报告及既有独立 QA 证据；按限定十个变化路径逐行审查。必要调用链包括实际订阅 repository 的 `ForUpdate`/事务 client 选择、退款收口、ent 事务实现、账号统计定价优先级、图片流式/拆分返回、Key API 返回值和编辑保存。
2. 只读 Git 均设置 `GIT_OPTIONAL_LOCKS=0`。起止 HEAD 均为 `7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`；最终 `git diff --cached --name-only` 为空，目标 `git diff --check HEAD -- …` exit 0。七个 tracked 差异和三个新增测试合计十个变化路径；冻结额外绑定的 `account_stats_pricing.go` 未改变。依赖声明及 lock 的限定 diff 为空。
3. 两个既有测试只作已批准的冲突语义调整：scheduler 保留原名称及真实选择断言，拆出有/无未来 reset；结算矩阵只为账号成本乘数谓词增加 `accountGate`，客户售价、循环、case 名、精度和其余断言保留。未发现删断言、skip 或放宽容差。
4. 独立重算十一源文件 SHA256，开始和结束均与 R2 freeze 一致；重算七份既有任务专属文档，全部匹配最终结果中的原哈希。
5. 实际重算既有证据映射：`qa-recovery-r2-evidence.json` 的 **128/128**、`controller-main-r2-review.json` 的 **27/27**、`controller-final-result.json` 的 **16/16** 引用哈希一致，映射之间存在交叉；阅读实际命令 receipts、Go JSON 结果、Vue 结果及失败边界。此操作是复核证据，不是重新运行其中的测试。
6. 针对 P1 新疑点执行：

   ```text
   node E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-review-agent/run-probe-v2.mjs
   ```

   runner 实际使用 Go `1.26.10`、`-mod=readonly -tags unit -overlay <私有overlay> ./internal/service -run ^TestPrecommitExtensionRespectsCallerTransaction$ -count=1 -json`，候选 **exit 1 / 1 个断言失败**；原 HEAD 对照 **exit 0 / 1 个通过**。runner 自身完成不代表候选通过，以两份 receipt 为准。`GOCACHE/GOTMPDIR` 均位于本 reviewer 专属目录，`GOPROXY/GOSUMDB=off`，未安装依赖或修改共享生成缓存。首版私有探针使用了不匹配当前 ent 的 Exec/Query 参数签名，两次构建失败日志完整保留；修正只写新 `v2` 文件，结论来自成功编译后的 v2。
7. 针对 P2 新疑点执行：

   ```text
   node E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-review-agent/key-form-status-probe.mjs
   ```

   实际 exit 1：候选有 2 个回归、1 个 exhausted→active 控制通过；原 HEAD 保留两种未保存选择，但仍有本次原定修复的 exhausted 状态缺口；上游修复三场景均符合预期。初次尝试加载 TypeScript 模块未得到可用导出，未执行探针；最终脚本直接擦除原函数唯一的 catch 类型注解，未重写处理逻辑。
8. 最终执行专属 `final-check.mjs`：十一源 **11/11** 匹配，HEAD 一致、暂存区空、限定 diff-check 与依赖/未改统计文件检查均 exit 0。初次 Node REPL Git 包装因无 `process` 全局未能执行，保留其记录；最终检查由独立 Node 脚本完成。没有执行暂存、提交、推送、部署、浏览器或外部服务调用，没有读取私密未跟踪配置。

### 复用且已核验身份的既有执行证据

| 既有证据 | 当时实际结果；本轮未重跑 |
| --- | --- |
| fresh main 相关 Go 回归 | unit 671、default 620；fail/skip 0 |
| fresh main 新 Go 场景重复三次 | 231 PASS |
| 独立 QA | service unit 73/default 65、repository 12、regression unit 671/default 620、mounted Vue 7；fail/skip 0 |
| frontend 三份相关 specs | 37 PASS、1 项已归因旧 FAIL；新增 quotaReset 8/8 |
| 构建及静态检查 | backend/frontend build、frontend typecheck、定向 lint、gofmt、限定 diff-check 已通过 |
| 旧测试对照及覆盖 | 原 production overlay 双标签各 17 个预期失败；候选两份对齐测试 290 叶通过；248 个顶层声明保留，scheduler 拆分使相关回归叶各增加 1 |

这些已有成功结果没有覆盖本轮新增的“带外层退款事务”和“表单已有未保存状态选择”，不能覆盖上述新反例。

### 审查身份

- 固定 HEAD：`7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`。
- 固定 upstream：`4a4a83ee5ec86b19d3a2dd9b6471de5128963bec`；上游参考只读取本地 Git 对象，未改变引用或联网更新。
- R2 freeze：`candidate-freeze-r2.json`，SHA256 `d5c30bd3e62d46e983db13614c901d6aab049862d486f4a279facef9090b463b`。
- 父合同 SHA256：`b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0`；测试增补：`a51b65f646f67a0a38dfda32bbdd63b2a1729c07b17ab5f20618cb79c8a9bb5c`。

| 审查源路径 | 起止均匹配的 SHA256 |
| --- | --- |
| backend/internal/repository/usage_billing_repo.go | `4a01690416403c25713e02d638bce8897c02ff6160071c1d895357c8c215b111` |
| backend/internal/repository/usage_billing_deleted_account_test.go | `595ea8de66352ce633b77ec7c42ab0f24a954125c3af3e4e633453722725de8e` |
| backend/internal/service/openai_images.go | `384b054c35991ad90ae7acaaf34f118681d8339e3158626bb93f881251768d93` |
| backend/internal/service/subscription_service.go | `bf305852650a40ce783a4fc722e5a6d3ff4e934a07220fea3f0db1a575af996d` |
| backend/internal/service/openai_gateway_service.go | `1d325b359d3931d8e8b30d41eb76354fc1896f8b6276b911bd5bbad3f7fe20d9` |
| backend/internal/service/upstream_gap_fixes_test.go | `5352fdb1350c96fe5aa0bc68df39c47781909ddf8099f1b44f21da0d802b1293` |
| frontend/src/views/user/KeysView.vue | `27489944f4a59349261866f9e690d5f7fabaa45344fff856464cad9dc58bd0a6` |
| frontend/src/views/user/__tests__/KeysView.quotaReset.spec.ts | `2f2d3029c640ff603fdd8294f50e2c2f27d2bf2c74228c49bd28af2fd6219c5b` |
| backend/internal/service/openai_account_scheduler_test.go | `b6fa14d686afe4702a5fcb5365b1903d54774633b255a4d7a8cf1004b9a9ca5c` |
| backend/internal/service/openai_subscription_long_context_test.go | `2302fb4c3c39d166c59f554822e633e163de3602774de8db0964b92461e1e51c` |
| backend/internal/service/account_stats_pricing.go（绑定但未改变） | `4eabb069d8a736ce825892c9e69b2d499f63d0e40ca59d3ff97230ec4d8df4d4` |

既有证据根：`E:/codex-runtime/sub2api/upstream-gap-fixes-20261103`。本轮新增证据仅位于其 `precommit-review-agent/` 子目录；最终索引 `precommit-review-evidence-final.json` 的 SHA256 为 `156bc58a4dcf79c671b3338aa4c941eb827a1bc22d0dd4446c17c4f5e7e19fd1`。其中绑定本轮 probe、结果日志和 receipt 的哈希；最初失败原件均保留。

## Unverified Risks

- 未执行真实数据库、provider、浏览器、线上负载或容器验收。P1 使用真实候选方法和 ent 事务生命周期，driver/repository 为替身；未在真实 PostgreSQL 中执行整条退款失败注入。P2 是实际原函数的 Node VM 定向检查，未补 mounted Vue/browser 测试。两项均已有明确源码可达路径和对照反例。
- 以下四项既有广域失败仍然保留，本轮未修复、未重跑：`TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost`；`TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy`；`TestOpenAIQuotaServiceNotConfigured`（nil panic）；`KeysView create query > does not overwrite a touched default group when rates load after the dialog opens`。不能称全套绿色。
- 第七项“图片空价继承目录价”仍暂缓，未混入本次审查目标。
- Controller 提供的全树文件保护和待提交十八路径范围是其并行职责，本 reviewer 未重新冻结或暂存全树。最终只读观察的暂存区为空；原始 index 历史曾有字节变化的既有记录保持原样，不据此声称全阶段 index 字节恒定。
- 旧合同及旧 QA 报告中的“未提交”是各自生成时的阶段快照。用户本轮后续已授权 Controller 作限定本地提交；本 reviewer 没有暂存、提交、推送或部署授权，也没有改写任何不可变历史报告。该后续授权不消除本轮发现的代码阻断。

## Recommendation

**需修复；暂不提交当前 R2 冻结候选。** 先解决 P1 的调用方事务归属及提交后缓存失效，再解决 P2 对未保存状态选择的覆盖。修复角色补相应定向回归、重新冻结变更源哈希，并由独立角色复核受影响路径；不要求无意义重跑全部既有验收。

通过后由 Controller 在用户已授权范围内执行限定本地提交。既有七份任务文档保持历史原件，当前结论由本报告补充。当前报告是提交前代码评审结论，不是发布、部署或线上验收批准。
