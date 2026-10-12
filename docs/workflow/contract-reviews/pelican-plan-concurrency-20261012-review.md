### PASS: pelican-plan-concurrency-20261012

## 独立 Contract Review

- 日期：2026-10-12
- Reviewer：独立 Contract Reviewer，按 `docs/workflow/agent-matrix.md`
- Contract：`docs/workflow/tasks/pelican-plan-concurrency-20261012.md`
- 检查工作树：`E:/codex-worktrees/sub2api/pelican-plan-concurrency-20261012`
- 实际 HEAD：`a62b9246397f1caf41f0011db353698164e848f7`，与 contract 一致；检查时工作树无未提交修改。
- Verdict：PASS。范围明确、验收可执行，未发现必须修订 contract 的阻断项。该结论仅批准实现范围，不代表实现或运行验收已通过。

## Findings

1. **四个不同套餐的有界并发**：当前 `pelican_test_service.go:314` 创建容量为 1 的 `planGate`。`RunNow` 与 scheduler 均进入 `claimAndRun`（约 715–764、1013 行），因此允许在同一服务实例上把容量调整为 4，并保持共同入口。容量检查发生在 settings 和 Claim 之前；保持这一顺序即可使第五个启动请求不触发 Claim、配额预留或计划时间更新。调度失败后仍可在下一 tick 重试到期计划。Contract 对此要求明确。
2. **同套餐仓储保护**：`pelican_test_repo.go:82–142` 的 Claim 使用事务和 `SELECT ... FOR UPDATE`，检查未过期的 `running_until` 后才增加 `run_generation`。`Finalize`、`ReserveAttempt`、`SaveResult` 仍按 generation 校验。仓储属于 denied paths；不得用放宽同套餐 Claim 或删除 generation fence 来实现并发。新测试中的 fake repository 必须自行模拟同套餐原子 Claim 冲突，不能把当前始终返回 ID=1 的 `pelicanRunnerRepo` 当作此项证据。
3. **停机与槽释放**：`claimAndRun:719–733` 在 Stop 共用的 mutex 内检查 stopped/root context、取得槽并执行 `wg.Add`；`Stop:1026–1044` 先在该锁内标记 stopped 并 cancel，再等待 WaitGroup。这个顺序已有可用基础，应保持，防止 Add 与 Wait 竞态。Claim/settings 在 Stop 时被取消，所有失败路径释放槽；run 在 Finalize 完成后释放槽，worker 随后 Done。Contract 已要求失败、完成、取消释放以及禁止停机后新启动。
4. **已有测试必须同步调整**：允许修改的 `pelican_test_runner_test.go` 中 `TestPelicanRunnerRejectsOverlappingManualRun` 当前要求不同 ID=2 被拒绝，与新行为冲突。开发需改为验证四个不同计划同时被接受、第五个拒绝，或由新并发测试替代其旧断言；不得仅删除覆盖。该文件已在 allowlist 中。
5. **内部模型与账号限制**：`run:771–789` 顺序遍历同套餐 models，`runModel` 仍使用 selector lease、每模型 timeout 与 ReserveAttempt。容量调整无需扩大到这些业务规则；contract 的保留约束充分。

## 实现与 QA 必须提供的已有验收要求

- 用 channel 屏障证明四个不同 ID 已同时进入执行，第五个 manual/scheduled 入口均拒绝且没有 Claim/ReserveAttempt 副作用。
- 证明一个完成/取消的 worker 释放槽后新的不同套餐可启动；settings 读取错误、feature disabled、Claim 错误与同套餐冲突也不漏槽。
- 在非满容量下尝试同套餐重复 Claim，证明仓储冲突仍生效，冲突释放的本地槽可复用。
- 以阻塞 Claim 和阻塞执行分别覆盖 Stop：取消可观察、Stop 等待所有已注册工作退出、之后的新启动被拒绝。有限 Stop context 超时应保留现有返回语义，不误报全部退出。
- 覆盖 `manual=false` 共用路径并核对容量拒绝后 due 时间不变；不必为测试等待真实 30 秒 scheduler tick。
- 执行 contract 中 Pelican service 测试与 service build；报告明确列出并发测试正则并以 `-count=20` 重复。若使用 external overlay，保留首次失败、overlay 映射及被屏蔽的无关测试路径，不能把 overlay 成功说成原始整包通过。

## Gate Checks

- success_criteria_testable：yes
- allowed_paths_explicit：yes
- denied_paths_explicit：yes
- acceptance_commands_executable：yes（运行结果由 Developer/QA 提供；本评审未执行测试）
- worker_model：contract/matrix 均指定 `gpt-6.1-sol`；实际调用身份与可用性由 Controller 确认
- base_commit_confirmed：yes
- openspec_traceable：not-applicable（独立 task contract）
- required_amendments：none

## 边界

本次只读检查 contract、service、仓储实现及已有测试，并写入本评审。未修改业务代码，未调用付费 provider、数据库、容器，未 commit/push/deploy。真实 PostgreSQL 多实例证明不属于本次 fake concurrency contract 的执行范围；仓储 fence 的保留结论来自当前源代码核对。
