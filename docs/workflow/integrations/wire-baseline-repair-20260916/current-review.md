### PASS: wire-current-integration

# 当前主线 Wire 集成 QA 合同独立审查

## Findings

- 合同的 QA 范围明确且适配当前主线：`HEAD` 为 `ce316421cf4384a1cdcdd3e7de33f39b5bc37a7a`，`d69da80b9` 是其祖先；中间已提交 Pelican（`82de91105`、`c44ac29c3`）及 first-response（`ce316421c`）变更。旧开发者仅在 `d69da80b9` 上产生审查过的修复源差异，当前 QA 以 `ce316421c` 为唯一生成和 diff 基线，禁止复制旧 `wire_gen.go`，可避免以陈旧生成物覆盖主线集成。
- “七个 source/test 差异”的可追溯范围由已批准的 `wire-baseline-repair` 合同精确给出：`backend/internal/service/wire.go`、`wire_provider_adapters.go`、`wire_provider_adapters_test.go`、`backend/internal/handler/wire.go`、`wire_provider_adapters_test.go`、`backend/internal/handler/admin/wire_provider_adapters.go`、`wire_provider_adapters_test.go`。当前 QA 不拥有这些文件的手改权限；controller 只能在 Developer `DONE` 后移植该七项已审查差异。允许的第八个 Go 文件仅为本树自动生成的 `backend/cmd/server/wire_gen.go`。
- 旧开发者报告保留的 `BLOCKED` 证据证明首次生成曾删除 `SetOpsService` 与 Cafe 可选 quota-reset hook；amendment 已将它们限定为两个 provider 封装，而不是手工恢复生成文件。新合同正确要求从当前基线重新生成、两次 SHA256 一致、并审查完整首次 diff，不能将旧基线的生成结果或“首次成功”作为本树证据。
- 生成完整 diff 的验收标准充分：除七项移植源和自动生成的 `wire_gen.go` 外，任何路径 drift 均停止；审查必须逐项核对构造器全部参数、`Start()` 相关调用及 cleanup。特别是必须证明 OpenAI handler 的 coordinator 和 `OpsService` setter 都保留，Cafe adapter 保持三项构造器依赖和仅在 `AdminService` 实现 `CafeQuotaResetService` 时的可选 setter；不能只以文本顺序或 Wire 成功作为行为等价证明。
- 当前 `ce316421c` 相对 `d69da80b9` 已在 backend 中改动包括 `cmd/server/wire.go`、`internal/handler/wire.go`、`internal/service/wire.go` 等路径，且承载并行 Pelican/first-response 工作。合同要求 pre/post inventory 及完整 diff 审查，足以防止移植覆盖这些主线改动；QA 报告必须明确它们是当前基线而非本轮改动。
- 原 Wire/Composite 门禁与新增 handler/admin provider 测试仍全部需要执行；额外 `PelicanReview` 和 `KeyRouteFirstResponse` smoke 以本地测试形式覆盖并行变更。合同明确禁止 DB、Redis、外部 provider、应用或容器启动；即使测试的 fake HTTP server 被使用，也不构成真实 provider 或运行态验收。

## Executed Checks

- 阅读 `wire-current-integration` 合同、已批准的旧合同及 amendment review、旧 Developer `BLOCKED` 报告。
- 验证当前 `HEAD` 等于声明的 `ce316421c`，并验证 `d69da80b9` 和 `ce316421c` 均为当前树祖先。
- 静态检查旧基线与当前主线的目标路径差异，确认当前主线已有 Pelican/first-response 的 Wire 相关基线改动，不能导入旧 generated output。
- 本树审查时 `git status --short` 无输出；`git diff --check` 退出码 `0`，`git ls-files -u` 退出码 `0`。未执行测试、生成器、编译、构建或服务启动。

## Unverified Risks

- Developer 尚未在本合同规定的前置条件下提供 `### DONE: wire-baseline-repair`；因此本 PASS 只批准后续独立 QA 合同，不允许现在运行 QA 或生成器。
- 未证明 controller 的七项移植与 Developer 最终输出逐字/语义等价，也未证明当前基线两次生成一致、完整 generated diff 无漂移、编译/构建及全部焦点选择器通过。这些均必须由后续 QA 重新取得实际证据。
- 本合同不覆盖真实 DB、管理员 HTTP、Redis、provider、容器或生产运行态，Composite P2/P3 亦仍未验证。

## Recommendation

`PASS`：在 Developer 报告变为 `DONE`、controller 完成且仅完成七个允许 source/test 差异的移植后，批准独立 Terra QA 在 `E:/codex-worktrees/sub2api/wire-current-integration` 重跑全部原/amendment 门禁、两次当前基线生成和两项附加 smoke。以移植前 inventory 为基线；任何额外 provider/hook、无关输出、Start/cleanup 不可证明的排序变化、失败门禁或未允许路径，立即 `BLOCKED`、保留证据且不修复。

---

## Current-base test-fixture amendment independent review — PASS

### Findings

- 已静态确认 `ce316421c` 的 production `provideCleanup` 在 `scheduledTestRunner` 与 `backupSvc` 之间新增 `pelicanTests *service.PelicanTestService`，并在 cleanup 并行步骤中仅于非 nil 时调用 `pelicanTests.Stop(ctx)`。这是当前主线 Pelican 生命周期行为，不能删除、重排或豁免。
- `backend/cmd/server/wire_gen_test.go` 的 `TestProvideCleanup_WithMinimalDependencies_NoPanic` 仍在 `scheduledTestRunner` 后立刻传入 `nil // backupSvc`，导致 positional argument 缺失并使 `cmd/server` 无法编译。该测试随后只执行返回的 cleanup closure，并以 `require.NotPanics` 断言；该失败是测试 fixture 与既有 production 签名失同步，不是 Wire repair、generator 或业务语义问题。
- amendment 仅授权 Developer 在此精确槽位插入 `nil, // pelicanTests`，不改生产签名、cleanup 步骤、断言、其他测试或任何 source/test 路径。nil 会走现有 guard，不调用 `PelicanTestService.Stop`，因而保持该测试的最小依赖和无启动/DB/provider 副作用边界。
- amendment 没有降低门禁：Developer 必须运行完整 `cmd/server` 测试包并写独立 `DONE` 证据；新鲜 QA 必须重跑此前失败的 server compile、完整 `cmd/server` 测试、所有原/修订 Wire 门禁、两次生成/hash、build、九个 Go 文件的 gofmt 与完整 diff 审查。任何额外 fixture 失败或 production/generated drift 仍为 stop rule，不能以本修复豁免。

### Recommendation

`PASS`：批准这一个测试 fixture 槽位同步。在 Developer `DONE` 和 controller 的完整 diff inventory 前，QA 不得执行后续验收；不得将 fixture 编译恢复表述为 Pelican cleanup、Wire 生成、完整 `cmd/server` 测试或运行态已通过。
