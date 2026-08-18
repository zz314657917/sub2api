---
task_id: upstream-cn-provider-403-s229-c
phase: contract-approved
base: 44fa471245618b58ddc1277624f90fdb9253a277
---

# Task Contract

## Role

Codex Controller/Generator；使用独立实现 worktree 和独立 QA worktree，手工适配上游行为，不直接 cherry-pick `10c8b7020`。

## Goal

从上游 `10c8b7020` 独立移植 CN 供应商 403 处置：Kimi/Zhipu/DeepSeek 与 OpenAI 共用 HTML 403 豁免、累计结构化 403 临时冷却及阈值永久禁用策略；非 CN 平台行为保持不变。

## Success Criteria

- CN 账号收到 HTML 403 时不递增 403 计数、不写永久错误、不写临时不可调度，并返回可继续 failover 的结果。
- CN 账号收到结构化 403 首次命中时递增计数并临时不可调度，错误原因含 `(1/3)`；达到阈值时永久 `SetError`，不再写临时冷却。
- OpenAI 既有 HTML/结构化 403 测试继续通过；Anthropic 等非 CN HTML 403 仍保持既有永久处罚行为。
- focused CN 403 tests x10、完整 `internal/service` 回归、server compile、格式、scope、provenance、冲突/index 和主工作区保护门禁通过。

## Context

- Repo: `F:/mcplugins/sub2api`
- Base: `main@44fa47124` after S229-B local integration
- Upstream: `upstream/main@938f1868a`；source `10c8b70203feac8fbd744d386af6600aa87c3837`
- 本地 `handleOpenAI403` 已包含 HTML 识别、403 counter、3 次阈值和临时冷却；本片只调整 `handle403` 分派。
- 上游 `openai_gateway_cn_fixes_test.go` 是 `unit` 专用聚合测试；本合同使用默认 tag 的独立 focused 文件，避免依赖 `unit` helper。

## Allowed Paths

- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/ratelimit_service_cn_403_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-403-s229-c-result.md`
- `docs/workflow/qa-reports/upstream-cn-provider-403-s229-c-qa.md`

## Denied Paths

- `backend/internal/service/openai_gateway_service.go` 及所有 billing/pricing owner
- `backend/internal/service/openai_gateway_responses_anthropic_native.go`
- `backend/internal/service/openai_gateway_messages_anthropic_native.go`
- `backend/internal/handler/**`、partial-result usage、disconnect drain/finalize、stream、composite routes
- frontend、migrations、dependencies、provider calls、push、deployment、containers、共享/生产数据库
- 所有用户已有 dirty/untracked 文件、`knowledge/**`、`outputs/**`，除本合同明确的 workflow report 外不得修改

## Constraints

- 复用现有 `handleOpenAI403`，不得复制或分叉计数器、阈值、冷却时间和错误构造逻辑。
- 只增加 `IsCNProvider(account.Platform)` 的分派资格；Antigravity 和其他平台分支顺序/行为保持不变。
- 不新增依赖、Redis/数据库 schema、网络请求或真实 provider 操作。
- 不把 billing、partial-result、disconnect 或 stream 修复混入 S229-C。

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "TestHandleUpstreamError_CNProviderHTML403SkipsAccountPenalty|TestHandleUpstreamError_CNProviderStructured403TempUnschedulable|TestHandleUpstreamError_CNProviderStructured403ThresholdDisables" -count=10
go test ./internal/service -count=1
go test ./cmd/server -run "^$" -count=1
gofmt -d internal/service/ratelimit_service.go internal/service/ratelimit_service_cn_403_test.go
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor 10c8b70203feac8fbd744d386af6600aa87c3837 upstream/main
Pop-Location
```

## Protected Main Baseline

- `main@44fa47124`，`origin/main@a865d8b6e`，不 push。
- 用户 dirty patch IDs：backend `d665008e...`；account modal `5d316e5b...`；tutorial `a07a7c33...`；knowledge files `2abee47d...`。`knowledge/tasks/current-task.md`、`docs/workflow/**` 为 Controller 流程文件，会随门禁更新。
- 六个未跟踪 tutorial migration/test 文件 SHA256 保持上一合同记录值；`outputs/**` 必须保持未跟踪且内容不变。

## Output

- 一个隔离 worktree implementation commit；Controller result 首行为 `### DONE: upstream-cn-provider-403-s229-c`。
- 独立 QA report 首行为 `### PASS: upstream-cn-provider-403-s229-c`，或明确 `FAIL/BLOCKED`。

## Stop Rules

- 任何 denied path、依赖、迁移、provider、数据库、容器、部署、远端、冲突或 unmerged-index 变化立即停止。
- 若 CN HTML 403 仍处罚账号、CN 结构化 403 未累计冷却、OpenAI/非 CN 行为回归，判定 FAIL。
