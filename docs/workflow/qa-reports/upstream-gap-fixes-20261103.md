### PASS: upstream-gap-fixes-20261103

## Findings

六项上游缺口已按本地架构适配并精确合回主目录，独立 QA、fresh-main 及最终源码保护通过。没有混入第七项“渠道图片价格留空继承目录价”语义迁移。未提交、推送、部署或更新容器。

| 修复 | 本地行为 |
| --- | --- |
| 请求期间删除上游账号 | 账号配额计数仅忽略 ErrAccountNotFound，保留用户余额/订阅结算、share ledger、幂等和真实数据库错误回滚 |
| 非流式图片张数 | 使用实际返回张数，无法计数时保留单张兜底，不再按请求 n 兜底；保留流式与拆分路径 |
| 管理员调整订阅时长 | 在既有事务中锁定读取、计算并更新到期时间/状态，提交后失效缓存；两路 +5 累积 +10 |
| 未来重置前配额暂停 | 有效未来 reset 优先于旧快照自愈；无有效 reset 时保留自愈，相对 reset 固定更新时间锚点 |
| 账号成本长上下文开关 | 账号关闭时模型文件成本不加长上下文价，保持分组售价、OAuth 豁免及自定义统计优先级 |
| Key 额度重置状态 | await 前捕获目标，消费服务端 quota_used/status，切换选中项时保持新编辑表单 |

实际改动为五个生产文件、三个新增测试及两个既有测试，共十个路径。冻结还绑定未修改的 account_stats_pricing.go，共十一条路径。

## Executed Checks

- 父合同 SHA256 `b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0` 独立 PASS；测试增补 `a51b65f646f67a0a38dfda32bbdd63b2a1729c07b17ab5f20618cb79c8a9bb5c` 独立 PASS。R2 freeze `d5c30bd3e62d46e983db13614c901d6aab049862d486f4a279facef9090b463b` 保留原九个 R1 哈希。
- 20 个相关旧测试文件原有 248 个声明保持完整；原快照 unit 670/default 619 全过。R1 的 17 个失败由两处旧语义引起，原日志完整保留。独立批准后仅对齐 scheduler 的有/无未来 reset 两场景及结算矩阵账号成本谓词；客户扣费、循环、case 名、精度和所有其他断言保留。
- 同一最终断言在原生产 overlay 中双标签均为 273 PASS/17 预期 FAIL；候选双标签均为 290 PASS。完整相关回归候选 unit 671/default 620 全过，只有原 scheduler 一叶拆为两叶，无删除/skip。
- Controller 候选新 Go 场景 77 项各重复三次共 231 PASS；backend build、frontend typecheck、定向 lint、生产构建、gofmt、scoped diff-check 通过。完整三份 Vue specs 为 37 PASS/1 个下述既有 FAIL，新增额度重置 8/8。
- 独立 QA 在另一副本实际运行 service unit 73/default 65、repository 12、mounted Vue 7 及完整回归 unit 671/default 620，均 PASS，无跳过。复用的 v2 原生产独立探针准确复现 30 个不同失败；原夹具问题及修正历史保留。
- 原 QA 准备阶段停止返回进度后，Controller 关闭并恢复同角色仍无结果，再由同矩阵 gpt-6.1-sol 的只读合同审查角色独立执行最终 QA。该角色未编写任何生产或测试；此为按仓库矩阵分配的角色恢复，不是用户另行指定模型。最终报告见 `upstream-gap-fixes-20261103-independent.md`。
- Controller 重算独立 QA 全部 128 项证据哈希、十一源、六探针和六组实际日志，全部一致。依赖保护核对 70,740 包代码文件及 3,187 链接，对变化项/关键运行包重哈希 1,124 项，无代码漂移。曾有早期 QA 默认 Vitest 写入共享生成缓存，已如实记录；之后禁用缓存，没有回滚未知用户缓存。
- 保护性集成检查 4,885 原 tracked 文件、目标 preimage、HEAD/index；复制后目标字节与冻结一致，其他 4,877 原路径在该复制动作前后不变。原 Pelican 两文件及无关文档 dirty 保留，未 stage。
- 主目录 fresh 检查：相关 unit 671/default 620、新 Go 重复三次 231 PASS；backend build、frontend typecheck、定向 lint、生产构建、gofmt/diff-check 全部通过，十一源前后哈希不变。Vue 三份 specs 为同样的 37 PASS/1 个既有 FAIL。汇总与 27 项证据哈希见 `controller-main-r2-review.json`，不是广域测试全绿。
- 最后一次原始 index 哈希保护如实 FAIL，保留为 `main-final-r2-protection.json`。发生于 fresh 检查及 scoped Git diff 之后；Controller 随后只读核对 4,885 条 path/mode/blob 与原 HEAD 完全相同、所有当前标志正常、仍无暂存改动，观测 index 前后字节稳定，没有恢复或覆盖 index。精确记录为 `index-reconciliation-r2.json`，最终保护要求此观测哈希及逻辑条目继续匹配；只能声明暂存内容不变，不能声明最终原始 index 字节不变。

证据根目录：`E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/`。关键文件为 `candidate-freeze-r2.json`、`controller-candidate-r2-review.json`、`controller-r2-regression-coverage.json`、`qa-recovery-r2-report.md`、`qa-recovery-r2-evidence.json`、`qa-recovery-r2-seal.json`、`controller-independent-qa-review-r2.json`、`integration-r2-result.json`、`controller-main-r2-review.json`、`main-final-r2-protection-reconciled.json`。每次命令的实际参数、cwd、时间和退出码均在对应 receipt/log，旧失败没有覆盖。

## Unverified Risks

以下四项已有失败本批未修，不能称广域 Go/Vue 全套绿色：

1. `TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost`
2. `TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy`
3. `TestOpenAIQuotaServiceNotConfigured`：旧 quota-service API nil dereference panic
4. `KeysView create query > does not overwrite a touched default group when rates load after the dialog opens`

Vue 旧失败使用原组件 overlay 和相同依赖复现。没有真实数据库、provider、浏览器、线上并发负载或容器验收；本地 SQLmock/事务替身与 mounted Vue 的结论限于所测路径。第七项图片价格语义仍暂缓。

## Recommendation

本批六项本地集成 PASS；已知基线失败单独保留，第七项语义迁移暂缓。提交、推送与部署均未执行；真实运行环境另行验收。
