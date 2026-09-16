### PASS: wire-baseline-repair

# Wire 基线修复合同独立审查

## 审查范围

- 任务合同：`docs/workflow/tasks/wire-baseline-repair.md`
- 已提交基线证据：`docs/workflow/integrations/composite-p1-audit-20260916/qa.md`
- 当前源码：`backend/internal/service/wire.go`、`content_moderation.go`、`usage_billing_settlement_service.go`、`welfare_service.go`、`cmd/server/wire.go`
- 工作树：`E:/codex-worktrees/sub2api/wire-baseline-repair`
- 审查性质：合同与静态源码审查；未运行生成器、测试、构建或服务。

## Findings

- 未发现阻止该隔离修复开始的合同缺口。Composite P1 独立 QA 已将 `go generate ./cmd/server` 的失败精确归因为两个缺失提供者：`[]ProxyRepository`（经 `NewContentModerationService` 的变参）与未导出的 `newUserTrialConsumer`（经 `ProvideUsageBillingSettlementService`）。本任务只修复这两项，与用户已确认的范围一致。
- 行为保真约束充分：`NewContentModerationService` 保持原签名和实现不变；固定参数 adapter 必须把全部既有依赖、以及恰好一个 `ProxyRepository` 转发给它。源码表明该构造器只在 `settingRepo != nil && repo != nil` 时启动 moderation worker 和 cleanup worker，因此 adapter 测试可传入 `nil` 的这两项并使用安全的依赖哨兵，避免后台 worker、DB、provider 或网络副作用。
- `*WelfareService` 已实现 `ConsumeNewUserTrial`，其签名与 `newUserTrialConsumer` 一致；在 `ProviderSet` 增加 `wire.Bind(new(newUserTrialConsumer), new(*WelfareService))` 可满足 Wire，同时不改变 `ProvideWelfareService`、结算构造器、`Start` 条件、清理生命周期或 billing 行为。合同明确禁止为生成通过而改写这些语义。
- 测试边界可执行：新增测试限定在 `wire_provider_adapters_test.go` 且以 `TestWireProvider` 开头；应断言 adapter 完整转发（包括单个 proxy）及 `*WelfareService` 的编译期接口满足。测试不得调用 `ProvideUsageBillingSettlementService`，因为它会调用 `Start()`；不得用非 nil moderation setting/repository 触发 worker。
- 生成结果尚不存在，因而本次不能对生成 diff 作“已审查”陈述。合同已要求开发者连续两次执行 `go generate ./cmd/server`，记录两份 SHA256，并完整审查第一次 diff；仅允许预期的 provider 调用/名称改动或可证明等价的 generator 排序。任何其他路径、额外 Wire 缺口、依赖变更或行为漂移均须停止并回交 Planner，禁止手改或 reset 生成文件。
- 共享脏改动受到保护：审查时 `git status --short` 显示 `docs/workflow/main-log.md` 与 `docs/workflow/status.md` 均为已有修改，均不在任务允许路径中。本任务不得覆盖、暂存、吸收或还原它们；所有前后 inventory、`git diff --check`、冲突检查和最终完整 diff 审查必须将它们当作基线并单独记录。

## Executed Checks

- 阅读任务合同、已提交 Composite P1 QA 证据及上述 Wire/构造器/生命周期源码。
- 静态核对：`NewContentModerationService` 的七项固定依赖加单个变参 proxy、`WelfareService.ConsumeNewUserTrial` 的接口签名、`ProvideUsageBillingSettlementService` 的启动副作用，以及当前 `ProviderSet` 中的两处缺口。
- `git status --short`：仅见上述两个共享 workflow 基线修改；未对其修改。
- `git diff --check`：退出码 `0`（仅现有文件 CRLF 警告）。
- `git ls-files -u`：退出码 `0`，无未合并条目。

## Unverified Risks

- 尚未生成，故尚未证明两次生成可复现、首次完整 diff 合规，或生成后服务器可编译。
- 尚未运行任何聚焦测试、Composite 原选择器、`go build ./...` 或格式检查；这些是开发和独立 QA 的后续门禁，不能由本合同 PASS 代替。
- 无真实 DB、Redis、provider、管理员 HTTP、运行中服务或 P2/P3 运行态验收；即使后续本地门禁通过，也不得据此宣称生产或完整运行态就绪。

## Recommendation

`PASS`：批准 Terra Developer 仅在本隔离 worktree 按合同允许路径实施两个 Wire 依赖修复并执行所列本地门禁。必须先完成开发者报告，再由新的独立 Terra QA 复跑全部验收命令、核验两次生成哈希与完整首次 diff。若出现第三个 provider、allowlist 外生成变动、依赖突变、被拒绝路径需求或无法隔离共享脏改动，立即停止并交回 Planner；本独立审查不授权直接修改合同以外的实现。

---

## Planner amendment independent review — PASS

### Amendment evidence and findings

- 首次生成后的 `wire_gen.go` 完整 diff 显示两处既有、仅存在于生成文件的 hook 被删除：`openAIGatewayHandler.SetOpsService(opsService)` 与 `adminService.(service.CafeQuotaResetService)` 成功后的 `cafeRoomHandler.SetQuotaResetService(quotaReset)`。开发者在此停止且保留输出，符合原 stop rule；这不是可接受的“仅 provider 调用/名称变更”。
- 两个 hook 都可在生成前基线 `HEAD:backend/cmd/server/wire_gen.go` 中定位，且当前 hand-written `handler/wire.go` 已有保留 coordinator 的 `ProvideOpenAIGatewayHandler` provider。amendment 仅授权将原有 setter 接入该 provider：新增 `*service.OpsService` 参数、调用现有 `SetOpsService`，并继续对同一 handler 设置 coordinator。不得改变 `NewOpenAIGatewayHandler` 签名、handler 逻辑或 coordinator 注入。
- 基线中 `opsService := service.ProvideOpsService(...)` 位于 OpenAI gateway handler 构造之前；`ProvideOpsService` 不调用 `Start()`，但会在存在 `SettingService` 时同步 warm quota-auto-pause 设置。新增 provider 依赖不应使此已存在的先后关系反转。最终完整 generated diff 必须逐项证明所有会 `Start()` 的 provider、cleanup 参数和调用仍保持可证明等价的依赖顺序；仅凭 Wire 编译成功不足以证明这一点。
- Cafe amendment 的封装边界充分：`admin.ProvideCafeRoomHandler` 只用当前三个构造器依赖调用未变的 `NewCafeRoomHandlerWithActivation`，然后对传入的 `service.AdminService` 做原样可选 `CafeQuotaResetService` 断言；仅 `ok` 时调用既有 setter。它不执行 reset、DB、网络或 worker 操作，且替换 `handler.ProviderSet` 中原构造器是必要的生成源修复。
- 新增 handler/admin 四项 allowlist 与所述 provider/test 文件相匹配；它们不会放宽 service 构造器、业务路径或生成文件的手工编辑禁令。现有 service worktree 修改、生成输出以及 `docs/workflow/main-log.md`、`docs/workflow/status.md` 仍属于当前实现/共享基线，审查者和后续 worker 均不得覆盖、暂存、回滚或吸收无关部分。

### Required test and diff gates

- `handler` provider 测试必须断言返回 handler 同时保留传入的 `*service.OpsService` 与 coordinator 的同一对象身份；测试只直接调用 provider，不启动应用或后台服务。
- `admin` provider 测试必须覆盖：支持 `CafeQuotaResetService` 时注入同一 reset capability；不支持时 `quotaReset` 仍为 nil；`AdminService(nil)` 时也保持 nil。三项原构造器依赖必须原样保留，测试不得调用 reset 方法、DB 或网络。
- 在 amendment 后，开发者和新鲜独立 QA 均须重新执行原有及新增全部命令。每方都必须保留首次完整 generated diff、第一次和第二次生成后 `wire_gen.go` 的 SHA256，以及两次字节相同的证据；完整 diff 必须明确对照全部构造器参数、所有带 `Start()` 的 provider、cleanup 调用和这两个 setter hook。
- 任一额外 hook 丢失、额外 provider 错误、allowlist 外路径、构造器/业务语义修改、不可证明的有副作用排序漂移，或无法隔离共享脏改动，均立即 `BLOCKED` 并回交 Planner；不得通过手改 `wire_gen.go`、reset 或扩大合同来继续。

### Amendment recommendation

`PASS`：批准仅按 amendment 实施两个 handler provider 封装及对应无副作用测试，然后重新完成全量本地和生成可复现门禁。此 PASS 只授权修复已证实被生成删除的两个基线 hook；不构成对当前首次 generated diff、运行态、DB/provider/admin HTTP 或生产行为的验收。
