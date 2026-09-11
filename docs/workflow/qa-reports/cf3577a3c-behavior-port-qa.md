---
type: qa-report
scope: repository
status: completed
task_id: cf3577a3c-behavior-port
verdict: PASS
base_commit: c1833f66c5ca5777ae0d11c0645d5e433ba1502e
reviewer: evaluator
last_verified: 2026-09-12
---

### PASS: cf3577a3c-behavior-port

# Independent QA Report

## Findings

- 未发现阻断本地行为级集成的明确问题。
- 审查中发现并在最终复跑前完成修复：tool parameter `"type": null` 的空白格式不能绕过 sanitizer；WebSocket 跨 turn 的 `python`、原生 `python__sub2api` 与 `Python` 归一化冲突会以 policy violation 拒绝；reverse-map 合并采用 copy-on-write，避免与并发 response restore 共享可变 map。
- 41 条上游路径映射已完成：39 条 `implemented/equivalent`，alpha/search 的 2 条因本地没有对应 gateway route 标为 `not-applicable`，没有导入其前置依赖链；禁止的 upstream split gateway/WS/alpha 文件均未出现。

## Executed Checks

- 静态全量 diff 审查：实现限于合同精确 allowlist；`NO_DENIED_PATHS`；无 conflict marker；`git diff --check` 通过。
- `backend/`: `go build ./...` 通过。
- `backend/`: 按合同以实际 service 生产 `GoFiles` 加三份 `cfport_*_test.go` 执行 `go test ... -count=1 -v`，通过。覆盖 reserved `allowed_tools` 别名、Astra `reasoning.mode=pro` 保留、prompt/null/schema、native call/item ID、拒绝字段受限重试、HTTP stream/nonstream/SSE-to-JSON/usage、真实 httptest WebSocket 双 turn、跨 turn alias 冲突拒绝及 buffered restore。
- `backend/`: 按合同以实际 handler 生产 `GoFiles` 加 `cfport_observability_test.go` 执行 `go test ... -count=1 -v`，4/4 通过；覆盖 request-body 脱敏诊断、terminal SSE capture 与失败状态归类。
- `backend/`: `go test ./internal/util/responseheaders -count=1 -v`，3/3 通过；覆盖 `reasoning.included` 头透传及过滤边界。
- 最终 Go 格式检查通过（无 `gofmt -d` 输出）。

## Unverified Risks

- 未使用真实 OpenAI OAuth/API-key/provider、真实 WebSocket 上游或生产凭据；HTTP/WS 仅由 fake transport 与 httptest server 验证。
- 未验证真实数据库、容器启动、部署、性能/长连接并发负载、推送或生产发布。
- frontend、migration、Ent/schema 与依赖锁文件按合同未触碰，也未作为本任务验收对象。

## Recommendation

`PASS`：可作为本地行为级上游 port 合入候选进入主控的后续提交/发布决策。该结论不代表真实 provider、数据库、容器或生产环境验收通过。
