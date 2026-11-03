---
type: task-contract
status: pending-review
task_id: upstream-gap-fixes-20261103-precommit-addendum
base_commit: 7ea675c020ff2c5d5e4124ac5e249c9dc65034c5
last_verified: 2026-11-03
---

# 提交前两项回归修正

## Goal

继续完成用户已授权的六项上游合入，并落实“起个智能体检查一遍没问题提交”。额外独立复查发现 P1/P2，当前 R2 不得提交。先在既有隔离候选修正两个回归，再由原只读 Reviewer 独立验收，Controller 保护性合回并执行本地精确提交；不推送、部署或更新容器。

## Immutable inputs

- HEAD `7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`。
- 原父合同 `upstream-gap-fixes-20261103.md`、测试增补及 R2 freeze 保持原字节。
- R2 freeze SHA256 `d5c30bd3e62d46e983db13614c901d6aab049862d486f4a279facef9090b463b`。
- 本轮失败报告 `docs/workflow/qa-reports/upstream-gap-fixes-20261103-precommit-review.md` 保持原字节。
- Reviewer 反例在 `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-review-agent/`，特别是 v2 调用方事务 overlay 和 Key 状态对照探针；不得改写其原件。
- 主树保护快照 `precommit-controller/before.json` 覆盖 12,357 原文件和暂存区空状态。只允许下述四路径作为 R3 产品/测试变化。

## Findings to close

1. `ExtendSubscription` 无条件另开并提交事务，破坏 `PaymentService.finalizePendingRefundSuccess → applyRefundFinalDeduction` 已传入的退款事务。外层回滚后订阅扣减仍保留，缓存提前失效。
2. `resetQuotaUsed` 在服务端状态没有变化时也覆盖同一 Key 的未保存启用/停用选择。

## Implementation scope

只在候选 `E:/codex-worktrees/sub2api/upstream-gap-fixes-20261103` 修改：

- `backend/internal/service/subscription_service.go`
- `backend/internal/service/upstream_gap_fixes_test.go`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/views/user/__tests__/KeysView.quotaReset.spec.ts`

Developer 可在候选新增专属结果 `docs/workflow/worker-results/upstream-gap-fixes-20261103-precommit-result.md`，辅助证据只写 runtime 下新建 `precommit-repair-developer/`。不得改写旧报告、原测试断言、依赖声明或 lock。

## Required behavior

- 有 `dbent.TxFromContext(ctx)` 时，锁定读取、计算和写入必须使用调用方事务，不另开、提交或回滚调用方事务；返回读取仍使用此事务上下文。
- 无调用方事务时，保留原 R2 的数据库事务和行锁，保留并发两路 +5 累积 +10、错误回滚、到期边界和状态更新。
- L1 和 billing subscription 缓存仅在实际事务所有者提交成功后失效。外层回滚/提交失败不得提前失效；无事务替身路径保持现有兼容行为。优先将事务归属修正限制在 `ExtendSubscription`，避免更改其他调用 helper 的业务路径。
- 重置始终消费实际响应的 `quota_used/status` 并更新捕获目标。覆盖目标旧状态前判断服务端状态是否实际变化；未变化则保留表单未保存的 active/inactive 选择。实际 `quota_exhausted → active` 时仍同步同一目标表单，切换/清空选择不得污染新表单。

## Regression requirements

- 保留原全部断言和测试声明，新增调用方事务成功提交、回滚、commit failure 场景。证明只有一个 begin、无提前 commit、读写使用外层事务；cache 在 commit 前和失败/回滚后保留，成功 commit 后失效。
- 保留原 standalone 并发、rollback、cache 和边界测试；新测试不得用修改旧断言掩盖变化。
- 新增 mounted Vue 两个反例：服务端 active、表单改 inactive；服务端 inactive、表单改 active。重置响应状态不变时，随后保存必须仍提交用户选择。
- 独立 Reviewer 使用自己的原失败探针复测 R3；可另增独立成功/失败生命周期探针，不能让 Developer 编辑其证据。
- 使用固定 Go1.26.10、readonly modules、离线依赖；Vitest `--no-cache`，不安装或改 node_modules/lock。

## Acceptance

1. 独立合同审查 PASS 绑定本文件最终字节后才允许 Developer 写候选。
2. Developer 停止写入后，R3 freeze 绑定十一条原路径。只有本合同四条可与 R2 不同，另外七条须完全相同。
3. 受影响 Go 新测试 default/unit 均通过；原 248 声明回归选择 default620/unit671、原 GapFix 定向选择继续通过，无 skip。新增事务反例应在 R2 失败、R3 通过。
4. quota reset 完整 mounted spec、相邻 multi-group spec、frontend typecheck/定向 lint、backend build 均通过。Vue create-query 既有失败与其他三项 Go 既有失败继续显式保留，不称全套绿色。生产构建仅在出现新疑点时追加，其他未变源码可复用前次证据。
5. 独立角色核 R3 源身份、重放失败探针、审四路径 diff 和相关实际日志；新报告 `docs/workflow/qa-reports/upstream-gap-fixes-20261103-precommit-recheck.md`，保留本轮 FAIL 原件。
6. Controller 重算证据、确认主树四目标仍等于 R2 后精确合入。除四路径和限定的共享任务快照更新外，12,357 原文件字节保持；暂存区仍为空，HEAD 保持。
7. 最终新鲜主目录检查后，精确提交十源码/测试和本批专属合同/结果/评审文档。共享 dirty 文件、Pelican、凭据和其他 untracked 不纳入。

## Roles and stop rules

Controller 为 Planner/Final Evaluator；原独立 Reviewer 审合同并复核，未参与实现。另一个独立 Developer 负责上述候选四路径。遵循已有矩阵及原生角色调用约束，不静默替换模型。

无最终字节合同 PASS、新阻断问题、四路径外必要修改、其他七源漂移、HEAD/index 或无关文件漂移：保留失败原件，交 Controller 判断，不扩范围或降低断言。无真实数据库/provider/browser/负载验收承诺。
