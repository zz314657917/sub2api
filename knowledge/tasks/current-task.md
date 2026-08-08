# 当前任务快照

最后更新：2026-08-08 17:20 +08:00

## 背景

- 用户要求检查上游最新版本并合入；上游 `main` 已刷新到 `cc67b1aca`。
- 本地与上游长期分叉，S206 只选择性移植最新 OAuth routing-hint 三提交链和 `nanoid` 安全升级，不整体 merge 上游历史。
- 工作在隔离 worktree `E:/codex-worktrees/sub2api/upstream-openai-routing-hints-s206` 完成，主工作树用户原有 `outputs/` 未被纳入。

## 当前目标

- S206 源码、依赖、测试和 P/G/E 证据已完成，最终结论为 `PASS / local-regression`。
- 完整 S206 提交链已以 `ff-only` 合入本地 `main`；不 push、不部署、不更新容器。

## 本次已完成

- OAuth HTTP normal/passthrough 在最终模型和 tier 策略完成后生成 `x-codex-routing-hint`，并移除旧 `OpenAI-Beta: responses=experimental` 行为。
- API-key 路径和所有大小写 header 变体 fail closed；hint 只允许模型加 canonical `priority`/`flex` tier，日志不记录原始 header、hint 或 token。
- WS 首轮和后续 `response.create` 更新 hint；pool 优先匹配/替换闲置 mismatch，容量繁忙时允许兼容连接软回退，不把 hint 作为续链硬条件。
- 直接拨号增加 generation guard，prewarm 校验最新 URL/proxy/routing target，避免清池或新请求后陈旧连接回灌。
- 上游 `8ad0a5ff5` 已精确 cherry-pick 为 `e6120ec69`，stable patch-id 为 `39eab1acf608c09d5492b0615eec3d8250427184`。
- 实现提交为 `dda605f62 fix(openai): port OAuth routing hints`；合同、worker result 和 QA 证据已记录。
- 本地 `main` 已从 `3cec8bb90` 快进到 `e8cfdead6`；合入后聚焦 service smoke 和 `cmd/server` compile 均通过。

## 已确认事实

- 聚焦 routing/WS/图片测试、完整 `internal/service`、`cmd/server` compile、gofmt、依赖 provenance、allowlist、冲突标记和 unmerged-index 门禁均通过。
- 初次完整 service 仅失败于旧图片测试仍断言 legacy beta header；该文件被窄范围加入合同 allowlist，断言按上游 `915cc7e7b` 同步后完整重跑通过。
- 本地 pool 没有上游后期的 `changedCh` 拓扑等待架构；S206 没有为本次 hint 功能引入该无关架构。
- 主工作树当前位于 `main@e8cfdead6`，只有用户原有未跟踪 `outputs/`；S206 不触碰该目录。

## 待验证点

- `go test -race` 因当前 Windows Go 为 `CGO_ENABLED=0` 且不存在 `gcc`/`clang` 未执行；默认构建并发回归已通过，但不声称 race-detector PASS。
- 真实 OpenAI OAuth HTTP/WS provider、代理、网络重试、多进程流量、容器、部署、staging 和生产均未验证。
- 未执行远端 push；本地 source merge 不等于发布或部署。

## 当前结论

- `PASS / local-regression`：S206 已达到本地 `main` fast-forward 合入门槛，无已知阻断 finding。
- review-and-verification 的建议为可继续本地合入；远端发布和运行环境验证保持独立授权。

## 下一步

1. 如用户明确要求发布，再单独执行普通 push 并验证远端 ref parity；不得顺带部署。
2. 如需生产验收，另建 provider/WS runtime 合同，使用测试凭据与隔离环境，不在 S206 source merge 中扩展。

## 验证记录

- `go test ./internal/service -run '<S206 routing/WS/image regex>' -count=1`：PASS，提交态 1.368s。
- `go test ./internal/service -count=1`：PASS，61.833s。
- `go test ./cmd/server -run '^$' -count=0`：PASS，提交态 0.056s。
- 实现门禁：`QA_GATES_PASS changed=13 patch_id=39eab1acf608c09d5492b0615eec3d8250427184 head=dda605f62b7f`。
- `go test -race ...`：未执行，环境阻塞为 `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`。
- 合入后：`go test ./internal/service -run '<S206 routing/WS/image regex>' -count=1` PASS，`go test ./cmd/server -run '^$' -count=0` PASS。
- 远端复查：`origin/main` 尚未包含 `e8cfdead6`，`git rev-list --left-right --count main...origin/main` 为 `5 0`；未执行 push。
