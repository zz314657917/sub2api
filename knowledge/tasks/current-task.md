# 当前任务快照

最后更新：2026-08-14 11:46 +08:00

## 背景

- 用户要求复核上游与本地差异，选择可安全合入的修复，并要求按 P/G/E 使用多个独立智能体审核和验收。
- 上游 OAuth pending exchange 越权绑定路径优先于常规上游合并处理；随后仅审定四项 v0.1.176 correctness 修复。
- 用户随后授权检查、分批提交和清理本地 P/G/E 分支与 worktree。

## 当前目标

- S214 OAuth pending 越权绑定、S213 四项 correctness 修复与 S215 两项可达 Grok 修复均已完成本地合入和验收。
- 已清理全部本地 `pge/*` 引用；唯一剩余的 S215 worktree 目录需在本地命令策略允许时删除。

## 本次已完成

- S214 已合入 `c36ea0fd5`：非终态 OAuth choice session 在 adoption、identity binding、profile mutation、session consume 和 token issuance 前被阻止；`invitation_required` 的既有 decision-only 行为与已认证 `bind_current_user` 保持。
- S213 已按独立提交合入本地 `main`：
  `405614fa1` Responses inconclusive 保持 unknown、
  `489b818ad` 顶层渠道定价冲突归一化、
  `fe32aa3e5` 分组 platform 变更失效 channel cache、
  `9b08c8126` scheduled backup PostgreSQL advisory leader lock。
- 四项合同均经独立 review 与独立 Terra QA；合入后综合默认标签回归、完整 service/server、server compile、Wire、格式和 Git 完整性门禁通过。
- S215 已确认 Grok 未登记文本定价与徽章/增量刷新问题在本地可达；Realtime 音频和分组长上下文两项因前置功能不存在而暂定 `BLOCKED`，不擅自扩大为功能移植或数据库迁移。
- S215 已合入本地 `main`：`3bdeaad66` 为未知 Grok 文本模型加入动态定价未命中后的 4.5 fallback，`69dc79320` 使 Grok 徽章和增量刷新使用 canonical `grok_usage_snapshot`。
- S216 已由多个独立智能体复核祖先关系、stable patch-id 等价性、worktree 状态与精确暂存范围。7 条 `pge/*` 分支均无未入主线的产品补丁，已删除其分支引用；S214/S213 的 5 个 worktree 与 S215 定价 worktree 已删除。

## 已确认事实

- 本地 `main@69dc79320` 比 `origin/main@23b2a1e92` 领先 7 个提交，其中包含 S214、S213 四项和 S215 两项提交；本轮没有 push。
- S213 定价归一化只影响顶层 `Channel.ModelPricing`，不改变 `AccountStatsPricingRules` 或 model mapping 的既有 lower-case-only 语义。
- S213 group cache 在 `groupRepo.Update` 成功后、复制账号前失效；失败、相同或缺省平台不失效，既有 auth cache 失效保留。
- S213 backup 只约束 scheduled job；锁 miss 跳过、获取错误记录 `scheduled backup` 日志并 fail-closed、成功时 defer unlock，manual CreateBackup 不加锁。
- 用户已有的两个前端文件和 `outputs/` 未被暂存、提交、覆盖或格式化。
- `backup/pre-s121-split-4161d254b` 含唯一提交，三条 `backup/*` 引用均保留；不应作为常规分支清理目标。
- `E:/codex-worktrees/sub2api/s215-grok-badge` 已被 Git 解除 worktree 注册且无 `.git`，但其任务副本因本地命令策略拒绝递归删除而残留。它是唯一未完成的清理目标。
- S215 4.5 fallback 保留本地 `grok`/`grok-latest` 的 4.3 默认，先剥离 `xai/`、`x-ai/`、`grok/` 前缀，严格在聚合输入 token 大于 200k 时升阶，并对 media/voice/realtime/search 维持 fail-closed。
- S215 前端保留 `share_display_tier`；Grok tier 以 credential、billing、canonical usage、legacy 等顺序读取，增量刷新比较稳定的 billing/usage 快照 key，canonical tier 存在时忽略 legacy 变动。

## 待验证点

- 预存的 `-tags unit` 全量服务测试仍无法编译，已由 S213 Amendment 1 明确排除；新回归均为默认标签且已执行。
- S216 目录残留需执行精确删除 -> 验证：仅在本地权限允许时删除 `E:/codex-worktrees/sub2api/s215-grok-badge`，再用 `Test-Path`、`git worktree list --porcelain` 和 `git for-each-ref refs/heads/pge` 确认目录、worktree 与 P/G/E refs 均不存在。
- 实际 provider、共享 PostgreSQL 多实例、生产计费路径、部署、容器更新和远程推送仍未验证或执行。

## 当前结论

- `BLOCKED / filesystem-residual`：代码整合、工作流证据提交、P/G/E refs 清理和 6 个 worktree 删除已完成；只剩一个已解除 Git 注册的任务目录因命令策略无法递归删除。S215 代码验收仍保持 `PASS / local-main-integrated`。

## 下一步

1. 删除唯一残留目录 -> 验证：在允许的本地权限下仅删除 `E:/codex-worktrees/sub2api/s215-grok-badge`，确认 `Test-Path` 为 `False`、`git worktree list --porcelain` 只含主工作区且无 `refs/heads/pge/*`。
2. 继续上游差异审查时，从当前 `main` 重新冻结候选 -> 验证：先做祖先、依赖和 `git apply --check`，再进入新 contract。
3. Realtime/长上下文只有在用户单独授权其前置 Voice/Realtime 或分组逐模型定价与前向迁移方案后才建立新 Sprint。

## 验证记录

- S214：exploit regression 移除 guard 后 identity count 为 1，恢复 guard 后通过；完整 handler、routes/server compile、格式、范围、拓扑与独立 QA 通过。
- S213：四个 slice 各自 `go test -list`、聚焦 `-count=10`、完整 service/server、server compile、Wire、格式、allowlist、diff、冲突与 clean worktree 通过。
- 合入后：聚焦 S213 pattern `-count=10` 通过；`go test ./internal/service -count=1`（67.698s）、`go test ./internal/server -count=1`、`go test ./cmd/server -run '^$' -count=0` 和 `go run github.com/google/wire/cmd/wire ./cmd/server` 通过。
- S215 合入后：Go 两项 fallback 测试发现且 `-count=10` 通过；`go test ./internal/service -count=1`（62.042s）、`go test ./internal/server -count=1`、`go test ./cmd/server -run '^$' -count=0` 通过。Vitest 发现强制用例，2 files / 8 tests 连续 3 轮通过，`vue-tsc --noEmit`、目标 ESLint、gofmt、diff、冲突和 index 门禁通过；仅有 Browserslist 数据过期提示。
- S216：7 条 P/G/E 分支的祖先/patch-id 审查通过，所有 worktree 清理前状态为空；`backend/` 中 `go test ./internal/service -count=1`（66.612s）和 `go test ./internal/server -count=1`（0.708s）通过。Git worktree 已仅剩主工作区、P/G/E refs 已清空、未合并索引为空；`git fsck` 无 missing/broken/corrupt/error 输出。根目录非 Go 模块，直接运行 Go 测试会失败，应从 `backend/` 执行。
