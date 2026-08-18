---
task_id: upstream-cn-provider-partial-usage-s229-e
phase: contract-approved
base: ddddcab70c4174c52ab64c9255bca4fe42c2b02d
---

# Task Contract

## Role

Codex Controller/Generator；在独立 worktree 手工适配上游 `10c8b7020` 的
partial-result usage 提交，不直接 cherry-pick 整个上游提交。

## Goal

当 OpenAI Chat Completions、Responses 或 Anthropic Messages handler 的 forward
返回非 failover 错误但携带部分 `OpenAIForwardResult` 时，仍提交一次 usage 记录；
客户端断开形成的 Messages partial result 也必须入账。Failover 错误必须保持
`result=nil` 语义，不能重复计费或阻止账号切换。

## Success Criteria

- Chat Completions 非 failover error + non-nil result 会提交 usage；Responses 同样。
- Messages 非 failover error + non-nil result 会提交 usage；`ClientDisconnect` 分支
  在返回前也提交 usage。
- Failover error 不提交 partial usage，原有 retry/switch/exhausted 行为不变。
- usage 任务继续使用当前 `RequestStartedAt`、渠道映射、quota platform、session 和
  trial session/release 语义，不引入重复 release。
- helper focused tests x10、完整 `internal/handler` 回归、server compile、格式、
  scope、provenance、冲突/index 和主工作区保护门禁通过。

## Context

- Repo: `F:/mcplugins/sub2api`
- Base: `main@ddddcab70` after S229-D local integration
- Upstream: `upstream/main@938f1868a`；source `10c8b70203feac8fbd744d386af6600aa87c3837`
- Local owners: `backend/internal/handler/openai_chat_completions.go` and
  `backend/internal/handler/openai_gateway_handler.go` (Responses + Messages).
- Existing success paths already build the correct `OpenAIRecordUsageInput`; the slice
  must reuse those fields for error-path partial results.

## Allowed Paths

- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_partial_usage_contract_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-partial-usage-s229-e-result.md`
- `docs/workflow/qa-reports/upstream-cn-provider-partial-usage-s229-e-qa.md`

## Denied Paths

- `backend/internal/service/**`、Responses native-Anthropic drain/finalize owner,
  billing, 403, count_tokens, stream conversion, composite routes
- frontend、migrations、dependencies、provider calls、push、deployment、containers、
  共享/生产数据库
- 所有用户已有 dirty/untracked 文件、`knowledge/**`、`outputs/**`，除本合同明确的
  workflow report 外不得修改

## Constraints

- 只在 non-failover error path 提交 partial usage；不得在 failover retry/switch 分支提交。
- `result == nil` 不提交；image-specific error branch 保留现有专用处理语义。
- 复用现有 `submitOpenAIUsageRecordTask`、`OpenAIRecordUsageInput` 字段和 trial release
  保护，不复制 pricing 或计费逻辑。
- 不新增依赖、网络请求或真实 provider/数据库操作。
- 实现与 QA 必须使用独立 worktree；QA 不得依赖 Controller 自述。

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/handler -run "TestShouldSubmitOpenAIPartialUsage|TestOpenAIRecordUsageInputsCarryQuotaPlatform" -count=10
go test ./internal/handler -count=1
go test ./cmd/server -run "^$" -count=1
gofmt -d internal/handler/openai_chat_completions.go internal/handler/openai_gateway_handler.go internal/handler/openai_partial_usage_contract_test.go
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor 10c8b7020 upstream/main
Pop-Location
```

## Protected Main Baseline

- `main@ddddcab70`，`origin/main@a865d8b6e`，不 push。
- 用户 dirty/untracked 文件和 `outputs/**` 必须保持当前值；`knowledge/tasks/current-task.md`
  与 `docs/workflow/**` 是 Controller 流程文件，不纳入用户 patch 保护值。

## Output

- 一个隔离 worktree implementation commit；Controller result 首行为
  `### DONE: upstream-cn-provider-partial-usage-s229-e`。
- 独立 QA report 首行为 `### PASS: upstream-cn-provider-partial-usage-s229-e`，或明确
  `FAIL/BLOCKED`。

## Stop Rules

- 任何 denied path、依赖、迁移、provider、数据库、容器、部署、远端、冲突或 unmerged-index
  变化立即停止。
- failover 发生重复 usage、trial release 重复/丢失、partial result 仍漏记或现有 handler
  回归，判定 FAIL。
