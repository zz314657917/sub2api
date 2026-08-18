---
task_id: upstream-cn-provider-responses-drain-s229-d
phase: contract-approved
base: ba6f4ab04d3e0cf34389de7d638876ba1c7dcf15
---

# Task Contract

## Role

Codex Controller/Generator；在独立 worktree 手工适配上游 `10c8b7020` 的
Responses×native-Anthropic 断开排水行为，不直接 cherry-pick 整个上游提交。

## Goal

客户端在 Responses 流式转换过程中断开后，继续读取上游 Anthropic SSE 至自然结束，
推进转换状态机并汇总末尾 `message_delta` usage；仅停止向客户端写出，不改变正常连接、
超时和非流式路径。

## Success Criteria

- 首次下游写失败后，仍继续排水上游事件并返回 `ClientDisconnect=true` 的结果。
- `message_start` 的 `input_tokens` 与末尾 `message_delta` 的 `output_tokens` 均保留；
  不因客户端断开提前返回而把输出计量截断。
- finalize 事件只在客户端仍连接时写出，并继续应用既有工具名反转/客户端工具还原。
- 正常 Responses native-Anthropic 流、上游数据间隔超时和已有客户端断开行为不回归。
- focused drain/finalize tests x10、完整 `internal/service` 回归、server compile、格式、
  scope、provenance、冲突/index 和主工作区保护门禁通过。

## Context

- Repo: `F:/mcplugins/sub2api`
- Base: `main@ba6f4ab04` after S229-C local integration
- Upstream: `upstream/main@938f1868a`；source `10c8b70203feac8fbd744d386af6600aa87c3837`
- 本地 owner 是 `backend/internal/service/openai_gateway_responses_anthropic_native.go`；
  已有 `openai_gateway_anthropic_native_pump_test.go` 的 pipe/failing-writer 夹具和超时回归。
- 上游对应行为位于同一 service owner；上游聚合测试使用 unit tag，但本合同采用默认 tag
  的独立 focused test，避免依赖无关 unit baseline。

## Allowed Paths

- `backend/internal/service/openai_gateway_responses_anthropic_native.go`
- `backend/internal/service/openai_gateway_anthropic_native_pump_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-responses-drain-s229-d-result.md`
- `docs/workflow/qa-reports/upstream-cn-provider-responses-drain-s229-d-qa.md`

## Denied Paths

- `backend/internal/handler/**`、partial-result usage 提交、403、billing、count_tokens、
  `openai_gateway_messages_anthropic_native.go`、其他协议/stream owner
- frontend、migrations、dependencies、provider calls、push、deployment、containers、
  共享/生产数据库
- 所有用户已有 dirty/untracked 文件、`knowledge/**`、`outputs/**`，除本合同明确的
  workflow report 外不得修改

## Constraints

- 保留现有上游 interval timeout 语义；客户端断开不能转化为服务端错误或触发 failover。
- 客户端断开后不得继续向 `c.Writer` 写任何事件；状态机仍需处理所有上游事件。
- finalize 事件必须走与逐事件路径一致的工具名还原逻辑，不能退回旧的裸 `ResponsesEventToSSE`。
- 不新增依赖、网络请求或真实 provider/数据库操作。
- 实现与 QA 必须使用独立 worktree；QA 不得依赖 Controller 自述。

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "TestResponsesStreamingFromNativeAnthropic_ClientDisconnectDrainsUsage|TestResponsesStreamingFromNativeAnthropic_HangTimesOut|TestResponsesStreamingFromNativeAnthropic_HappyPathStillConverts" -count=10
go test ./internal/service -count=1
go test ./cmd/server -run "^$" -count=1
gofmt -d internal/service/openai_gateway_responses_anthropic_native.go internal/service/openai_gateway_anthropic_native_pump_test.go
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor 10c8b70203feac8fbd744d386af6600aa87c3837 upstream/main
Pop-Location
```

## Protected Main Baseline

- `main@ba6f4ab04`，`origin/main@a865d8b6e`，不 push。
- 用户 dirty/untracked 文件和 `outputs/**` 必须保持当前值；`knowledge/tasks/current-task.md`
  与 `docs/workflow/**` 是 Controller 流程文件，不纳入用户 patch 保护值。

## Output

- 一个隔离 worktree implementation commit；Controller result 首行为
  `### DONE: upstream-cn-provider-responses-drain-s229-d`。
- 独立 QA report 首行为 `### PASS: upstream-cn-provider-responses-drain-s229-d`，或明确
  `FAIL/BLOCKED`。

## Stop Rules

- 任何 denied path、依赖、迁移、provider、数据库、容器、部署、远端、冲突或 unmerged-index
  变化立即停止。
- 断开后仍向客户端写事件、末尾 usage 丢失、超时语义变化或正常流回归，判定 FAIL。
