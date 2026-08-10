# 当前任务快照

最后更新：2026-08-10 18:19 +08:00

## 背景

- 用户要求继续选择性合入上游 `v0.1.173`，并明确包括 `6e34fb09c`。
- 本地与上游长期分叉，S207 只做行为级移植，不整体 merge release。
- 隔离 worktree：`E:/codex-worktrees/sub2api/upstream-v0173-s207`；冻结基线：`main@ebc3438e4`。

## 当前目标

- S207 实现、Evaluator QA 和本地 `main` fast-forward 已完成，结论为 `PASS / local-regression`。
- 合入后聚焦 service 与 server compile smoke 已通过。
- 不 push、不部署、不更新容器，不触碰主工作树用户原有 `outputs/`。

## 本次已完成

- Gemini 原生和 Claude-compatible 转发按实际有效 `inlineData` / `inline_data` 图片数计费；流式累计内容取单 payload 最大值，保留原模型和 mapped model 的一图 fallback。
- Gemini 三条本地路径保留错误策略结果，池模式 429、`ErrorPolicySkipped`、`ErrorPolicyTempUnscheduled` 不再落账号级限流；自定义错误码命中、OAuth 和非池账号原行为保留。
- Web Search setting 缺失作为正常禁用空配置并使用正常 TTL；`BaseDialog` 重开只重置 body 滚动位置。
- Grok OAuth client 缺失返回 `GROK_OAUTH_CLIENT_NOT_CONFIGURED`，Exchange session 在配置检查前不被消耗。
- Antigravity 模型 fallback 成功后，`ForwardResult.UpstreamModel` 记录实际 fallback model。
- 上游 `6e34fb09c` 的前置 `db0bff82c` 及其 76 文件 schema/migration/admin usage-audit 链明确排除。

## 已确认事实

- 两组 contract 聚焦测试、三条 Gemini 策略路径测试、完整 `internal/service`、`cmd/server` compile 均通过。
- BaseDialog Vitest `1/1`、frontend typecheck、production build 通过。
- 14 个实现/流程路径均在 allowlist；`git diff --check`、冲突标记、unmerged-index 均通过。
- 五个源提交均属于 `v0.1.173` release commit；`db0bff82c` 不是 S207 HEAD 的祖先。
- 临时 `frontend/node_modules` junction 已删除，主工作树依赖目录仍存在。
- 本地 `main` 已从 `ebc3438e4` fast-forward 到 `f50183f83`；合入后聚焦 service 与 server compile smoke 通过。
- `origin/main` 未变化，本地 `main` 当前领先远端 9 个提交；用户原有 `outputs/` 仍为未跟踪且未触碰。

## 待验证点

- 未进行真实 Gemini/Grok/Antigravity provider、共享 PostgreSQL/Redis、容器、部署、staging、生产或性能验证。
- 未执行远端 push；本地 source merge 不等于发布或部署。

## 当前结论

- `PASS / local-regression`：S207 已选择性合入本地 `main`，未发现阻断 finding。
- review-and-verification 建议：保留当前本地结果；远端发布和运行环境验证保持独立授权。

## 下一步

1. 如用户明确要求发布，再单独执行普通 push 并验证远端 ref parity；不得顺带部署。
2. 如需真实 provider 或生产验收，另建隔离 runtime 合同，不扩展 S207 source merge。

## 验证记录

- `go test ./internal/service -run '<S207 focused regex>' -count=1`：PASS。
- `go test ./internal/service -run 'TestAntigravityGatewayService_ForwardGemini|TestGeminiMessagesCompatService|TestWebSearch' -count=1`：PASS。
- `go test ./internal/service -run 'TestGemini(ChatCompletions|MessagesCompat|Native)_(TempUnscheduled429|PoolMode429)DoesNotSetAccountRateLimit' -count=1`：PASS。
- `go test ./internal/service -count=1`：PASS，最新运行 61.428s。
- `go test ./cmd/server -run '^$' -count=0`：PASS。
- BaseDialog Vitest：`1 file / 1 test PASS`；frontend typecheck/build：PASS。
- allowlist、provenance、`db0bff82c` exclusion、format、conflict、unmerged-index：PASS。
- 合入后 `go test ./internal/service -run '<S207 focused + policy paths regex>' -count=1`：PASS，5.472s。
- 合入后 `go test ./cmd/server -run '^$' -count=0`：PASS，10.469s。
