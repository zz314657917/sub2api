# 当前任务快照

最后更新：2026-08-15 00:05 +08:00

## 背景

- 用户要求核实并合入上游最新版的 GPT/Codex 额度异常修复。
- 当前本地主线长期分叉，禁止整包合并上游；每项必须按行为移植、独立回归和 P/G/E 门禁执行。
- 主工作区已有用户未提交的账户编辑前端源文件、对应测试和 `outputs/`，本任务始终排除。

## 当前目标

- S217 仅移植上游 `v0.1.176` 中本地确实缺失的三组 GPT/Codex 额度行为：
  `358e4a89a` 个人订阅到期日、`12abb5470` HTML 403 账号保护，以及
  `54a2bcfd1` 尚未覆盖的 reset-credit API/UI 一致性。
- 保留既有本地实现：`99b31067f` / `3d3aee2e` 的实际 OpenAI 自动暂停逻辑已等价，S188 已实现额度券消费后的恢复优先顺序。

## 本次已完成

- 已 fetch 并冻结上游 `upstream/main@fbfdcef81`；确认五项候选都已包含在 `v0.1.176`。
- 三个独立只读审计已完成：订阅/403 可适配，调度阈值是既有等价实现，reset-credit 后端高风险恢复链已在 S188 落地但前端/API 一致性仍缺失。
- S217 合同及规格已由 `45ec2ed13` 提交，独立合同审查要求的默认标签 403 测试、禁止 reset 后自动二次刷新、POST/GET 路由契约、精确订阅回归名称均已写入并批准。
- 已创建干净隔离 worktree：`E:/codex-worktrees/sub2api/s217-gpt-quota`，分支为 `codex/upstream-v0176-gpt-quota-s217`，基线 `45ec2ed13`。
- Terra Generator 启动前即因模型 404 失败，未消费输入 token、未写文件。该 BLOCKED 状态已由 `b260a8501` 记录。
- 用户随后明确授权 `gpt-5.6-sol`，但该模型同样在输入前返回 404，隔离 worktree 仍无文件变更。

## 已确认事实

- 本地 `main@b260a8501` 未包含三项待移植行为；`358e4a89a` 与 `54a2bcfd1` 不能直接 apply，`12abb5470` 可以适配但不应原样复制其缺失的上游测试桩。
- 本地 `openai_gateway_service.go` 的真实资格判定已正确保留 0-100 Codex 百分比，并跳过已 reset/超过两小时的陈旧快照；针对这些行为的 focused service regression 已通过。
- S188 已使用独立 8 秒 context、先恢复账号状态再刷新缓存；S217 不得改变该顺序或消费前错误语义。
- 当前根工作区仅有用户未提交的：
  `frontend/src/components/account/EditAccountModal.vue`、
  `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`、
  `outputs/`。它们从未被暂存、提交或覆盖。
- 未执行 push、部署、容器更新、真实 provider 调用、生产流量或数据库迁移。

## 待验证点

- `BLOCKED / worker model`：`gpt-5.6-terra` 与用户明确授权的 `gpt-5.6-sol` 均由 Claude CLI 返回 selected-model 404。必须由用户明确授权一个实际 CLI 模型别名，或授权停止 P/G/E worker mode 后由主控实现；不得静默降级。
- 许可替代模型后 -> 在既有 S217 worktree 运行 Generator -> 验证：标准 worker report、每个切片的 focused regression、完整 service/server、server compile、前端可执行命令或准确环境 BLOCKED 证据。
- Generator 完成后 -> 独立 Terra QA -> 验证：报告首行 verdict、allowlist/diff/冲突/index/provenance 门禁；再由 Codex 决定是否集成本地主线。

## 当前结论

- `BLOCKED / worker-model-unavailable`：S217 已完成审计与合同门禁，但两次指定 Developer Worker 模型均不可用；没有业务代码、测试、数据库或前端行为被修改，因此尚未合并 GPT 额度修复。
- 隔离 worktree 保留且干净，等待用户授权一个可用的 Developer Worker 模型后可直接恢复。

## 下一步

1. 用户指定允许的实际 CLI Developer Worker 模型别名，或授权主控直接实现 -> 验证：重新启动 S217 Generator 或正式停止 worker mode，确认不会触发已知模型 404。
2. Generator 实现三项合同切片 -> 验证：合同中所有后端/前端/静态检查及标准 worker report。
3. 运行独立 QA 并审 diff -> 验证：QA PASS、无 denied-path、主工作区用户改动仍未暂存。
4. 仅在 QA PASS 后集成到 local `main` -> 验证：主线回归、Git 完整性和精确暂存范围；推送仍需用户另行授权。

## 验证记录

- `git fetch upstream main` 成功；`v0.1.176` 已验证包含 `358e4a89a`、`12abb5470`、`99b31067f`、`3d3aee2e`、`54a2bcfd1`。
- `git apply --check`：`99b31067f` / `3d3aee2e` 缺少上游阈值框架文件；`358e4a89a` / `54a2bcfd1` 因本地分叉不适用；`12abb5470` 可应用但测试支撑需本地化。
- 阈值审计已执行：
  `go test ./internal/service -run 'Test(OpenAIGatewayService_SelectAccountForModelWithExclusions_(AutoPauseBy5hThreshold|AutoPauseBy7dThreshold|AllowsBelow5hThreshold|StaleUsageWindowResetSkipsPause|FreshUsageWindowStillPauses|StaleUsageSnapshotSkipsPause_Issue2994|FreshExhaustedSnapshotStillPauses_Issue2994)|OpenAIGatewayService_SelectAccountByPreviousResponseID_QuotaAutoPausedMiss|BuildCodexUsageExtraUpdates_FreshAccountUsedPercentNotInverted_Issue2994)$' -count=1` 通过（0.069s）。
- S217 worktree 与根工作区状态已复核；Claude CLI `2.1.191` 可执行，但对 `gpt-5.6-terra` 和用户授权的 `gpt-5.6-sol` 请求均在输入前返回 404。
