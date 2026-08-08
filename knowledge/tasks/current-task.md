# 当前任务快照

最后更新：2026-08-08 +08:00

## 背景

- 用户要求检查并选择性合入上游 GPT-5.6 Sol/Terra/Luna 定价更新，并修复复核中发现的动态解析遗漏。
- S205 在隔离 worktree `E:/codex-worktrees/sub2api/gpt56-pricing-metadata-s205` 中按 P/G/E 合同完成，主工作树原有 `outputs/` 未被纳入。
- S205 实现与 QA 链随后已发布到 `origin/main`；仍不包含部署、容器更新、真实 provider 调用或生产操作。

## 当前目标

- S205 已完成本地 `main` 合入、远端发布和发布后复查，当前没有剩余产品代码动作。
- 保持部署、Batch billing 支持、剩余文档同步和 worktree 清理为独立授权事项。

## 本次已完成

- 选择性应用上游 GPT-5.6 定价元数据提交 `78ad91f56`；其 stable patch-id 与候选 `0616a2974` 一致，并保留来源 `60f6dc91cf907841c09b4aa7f9f78874fd08579c`。
- `pricing_service.go` 现保留三项 `long_context_*` 字段以及 Batch/Flex 两项 cache-write 元数据，不修改 `billing_service.go`、tier 选择、handler、前端、数据库或部署路径。
- Sol/Terra/Luna 真实 JSON fixture、standard/priority/flex/cache-write 和 272K 长上下文边界计费矩阵均已覆盖。
- 分支最终提交为 `7ede90e52 test(pricing): verify GPT-5.6 billing metadata`，本地 `main` 已从 `920de6d13` 快进到该提交。
- `498d31072 docs(workflow): record S205 unit baseline risk` 及之前的完整 S205 实现/QA 链已推送到 `origin/main`；`9d2d47bba docs(workflow): record S205 publication` 随后也已同步。
- S205 合同、worker result、QA、workflow status 和 main-log 已记录完整证据。

## 已确认事实

- `origin/main@9d2d47bba9a79bc254cc9d3aec794e060b3fe44e` 已包含 S205 代码、测试、QA、unit baseline 风险和发布记录；本 handoff 文档提交完成后，本地将仅领先该文档提交。
- 主工作树只有用户原有的未跟踪目录 `outputs/`；没有未合并索引或其他跟踪文件改动。
- 合入后从主工作树重跑 `TestDefaultPricingGPT56FeedsBillingTierAndLongContextMatrix` 通过；完整 service、聚焦 service/handler、JSON 语义、gofmt、stable patch-id、allowlist、冲突标记和 Git integrity 门禁均通过。
- OpenAI 官方 pricing/changelog 与当前 Sol/Terra/Luna standard、Batch/Flex、Fast、cache-write 和“超过 272K”为长上下文的元数据一致。
- 本地 shell、Git、Go、`rg` 和 `apply_patch` 可用；Agent Matrix 指定的 `deepseek-v4-pro` worker 在当前协作环境不可用，最终命令验收由主 Evaluator 执行。
- 当前仍注册两个隔离 worktree：S205 worktree，以及 `E:/codex-worktrees/sub2api/upstream-billing-rate-sync-s204`；本轮未删除或修改它们。

## 待验证点

- 真实 provider、远程 pricing refresh、容器、部署、staging 和生产流量均未验证。
- Batch cache-write 字段目前只被保留为元数据；仓库没有 Batch billing selector，本轮未新增。
- S205 产品、QA 和发布记录已远端同步；本 handoff 文档尚未同步，且不能把已发布表述为已部署。
- S180 浏览器资源阻塞与 S205 无关，不应被后续短指令自动当作 S205 的剩余步骤。

## 当前结论

- `PASS / published`：GPT-5.6 Sol/Terra/Luna 定价元数据和解析修复已合入本地 `main` 并发布到 `origin/main`，关键回归与远端祖先检查通过。
- 当前结果是 source/local-regression 加远端 Git 发布证据，不是真实 provider、部署或生产验证。

## 下一步

1. 如用户明确要求同步剩余发布记录和 handoff 文档，执行普通 `git push origin main` -> 验证：`origin/main` 与本地 `main` hash 一致，ahead/behind 为 `0/0`；不得顺带部署。
2. 如用户明确要求清理，先逐个确认两个 worktree 干净且分支已合入 -> 验证：只删除明确归属的 worktree/分支，主工作树 `outputs/` 保持不变。
3. 如需让 Batch/Flex 元数据参与新的运行时选择，另建独立合同 -> 验证：覆盖请求分类、tier 选择、计费和回归，不混入 S205 收尾。

## 验证记录

- `go test ./internal/service -count=1`：PASS，61.175s。
- `go test ./internal/service -run 'TestParsePricingData|TestDefaultPricingIncludesGpt56|Test.*GPT56|Test.*GPT5' -count=1`：PASS。
- `go test ./internal/handler -run 'TestOpenAIResponsesWebSocket|TestOpenAIWS' -count=1`：PASS。
- `go test ./internal/service -run 'TestDefaultPricingGPT56FeedsBillingTierAndLongContextMatrix' -count=1`：PASS；合入后从主工作树复跑仍通过。
- 最终合同门禁：`QA_GATES_PASS changed=9 patch_id=d8c2433637f22f1892040e4e19feb2546c26d26f head=7ede90e52f46`。
- 远端发布复查：`git merge-base --is-ancestor 7ede90e52 origin/main`、`498d31072 origin/main` 均通过；`origin/main` 随后前进到 `9d2d47bba`，提交本 handoff 前 ref parity 为 `0/0`。
