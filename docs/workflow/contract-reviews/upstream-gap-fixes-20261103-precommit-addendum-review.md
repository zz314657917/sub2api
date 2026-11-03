### PASS: upstream-gap-fixes-20261103-precommit-addendum

## Findings

**合同审查通过。未发现阻断条款；四个源文件范围足以关闭原独立评审的 P1/P2，验收要求充分。此结论仅批准按所绑定合同进入 Developer 阶段，不代表修复实现或提交前验收通过。**

- 审查合同：`F:/mcplugins/sub2api/docs/workflow/tasks/upstream-gap-fixes-20261103-precommit-addendum.md`。
- `contract_sha256`：`64c85c29a36f745dc005efb0a0fd75689943fe399507306635ec891e9a14cdfb`。
- 固定 HEAD：`7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`。
- 原 R2 freeze SHA256：`d5c30bd3e62d46e983db13614c901d6aab049862d486f4a279facef9090b463b`。

### P1：范围和验收覆盖事务归属及缓存生命周期

合同明确，有 `outerTx = dbent.TxFromContext(ctx)` 时，锁读、计算、写入及返回读取都沿用调用方事务，不另开事务，不提交或回滚调用方事务；没有外层事务时保留 R2 的行锁、独立事务、并发累加和错误回滚。

合同还明确约束 **L1 和 billing subscription 两类缓存**：只有实际事务所有者成功 commit 后才失效；outerTx commit 前、rollback 后及 commit failure 后均不得失效。该要求同时出现在 Required behavior 和 Regression requirements，不能以“注册了提交钩子”或“函数返回成功”代替真正成功提交后的失效证据。

实现可使用现有事务上下文和 ent 提交钩子，将调整限制在 `subscription_service.go` 的 `ExtendSubscription` 路径，并在 `upstream_gap_fixes_test.go` 增加生命周期验证；不需要据此修改 repository、ent 生成代码、退款调用方或其他 helper 调用路径。当前没有确认必须扩展路径的理由。

验收要求新增 outerTx 成功提交、rollback、commit failure 场景，证明一个 begin、无提前 commit、读写归属外层事务及缓存时序，并保留 standalone 并发、状态、回滚、缓存和边界断言；独立 Reviewer 还须重放原失败探针。该组合覆盖原 P1 反例及修复后需要保留的保护。

后续实施和复核应按合同对两类缓存分别观察调用或状态：失败/回滚场景不能只检查数据库到期日，成功场景不能只检查无提前失效，还须证明 commit 成功后发生失效。

### P2：明确保留同一 Key 的未保存状态选择

合同要求 await 前捕获目标，始终将服务端响应的 `quota_used/status` 应用于该目标，并在覆盖旧状态前判断服务端状态是否实际变化。

**同一 Key 的服务端状态未变化时，必须保留表单未保存的 active/inactive 选择。** 实际 `quota_exhausted → active` 时仍同步同一目标表单；切换或清空选择不得污染新表单。这些条款与原 P2 的修复方向一致。

新增 mounted Vue 两个相反方向的反例，并要求随后保存仍发送用户选择，覆盖了“保留表单显示”到“实际保存参数”的完整回归链。原 quotaReset 测试和独立源函数反例仍须保留、复测。实现及测试可分别限于 `KeysView.vue` 和 `KeysView.quotaReset.spec.ts`。

## Executed Checks

- 本轮实际读取并逐条审查上述增补合同，执行 `Get-FileHash -Algorithm SHA256` 取得最终字节身份；仅聚焦该合同，没有重新扫描背景源码、父合同或整套 QA。
- 核对 Implementation scope：仅四个源文件可成为 R3 产品/测试变化；另有明确单列的 Developer 新结果文档和专属 runtime 新证据目录，不构成扩大产品修改范围。
- 核对不可变输入保护：原父合同、测试增补、R2 freeze、原 FAIL 报告和 Reviewer 原探针不得改写；其他七条冻结源须与 R2 完全一致。
- 核对验收：合同 PASS 在先、Developer 停写后冻结 R3、default/unit 定向和原回归无 skip、原反例 R2 失败/R3 通过、完整额度重置 mounted spec、相邻 multi-group spec、typecheck/定向 lint/backend build，以及独立四路径 diff/日志复核和 fresh-main 门禁。
- 核对角色分离及停止条件：本 Reviewer 不参与实现；另一个 Developer 写候选；出现四路径外必要修改、其他源漂移或新阻断时返回 Controller，不允许降断言或扩大范围。
- 本轮未执行测试、构建、暂存、提交或部署；唯一新增仓库输出为本审查报告。原 FAIL 报告及证据保持原件。

## Unverified Risks

- 尚无修复实现或 R3 执行结果。本合同足以指导关闭两项发现，是否实际关闭须由后续独立验收决定。
- 后续应以成功 commit 的实际返回结果为缓存失效边界，覆盖两个缓存；不能仅凭提交钩子存在推断成功、失败和 rollback 三种时序均正确。
- 复测原失败探针时应使用另建的运行配置指向 R3，保留原探针和旧结果，避免读到 R2 或覆盖原反例。
- 合同明确保留四项旧失败和 live DB/provider/browser/负载未验证边界；本次合同 PASS 不表示全套绿色。
- 本结论仅适用于上述 SHA256 的合同字节；若合同修改，须重新绑定并独立复核。

## Recommendation

**PASS，可由 Controller 调用另一名 Developer，按本合同在四个源文件内实施。** 当前无需扩展产品/测试路径。

Developer 完成并停止写入后，冻结 R3，由本只读独立角色审查四路径 diff、核对另外七源不变、重放原反例并核实新增事务生命周期和 mounted Vue 证据。只有后续独立验收及 Controller fresh-main 门禁通过后，才进入已授权的限定本地提交；不推送或部署。
