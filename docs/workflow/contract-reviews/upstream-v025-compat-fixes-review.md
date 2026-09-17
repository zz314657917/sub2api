### PASS: upstream-v025-compat-fixes

# Contract Review

## Contract Checked

- `docs/workflow/tasks/upstream-v025-compat-fixes.md` at isolated-worktree base `a022d9a3457457463ed06dd025370ad57936e668`.
- Upstream references: `bc8da7815`, `62635532a`, `8447bdd36`, `196c15b5c`, `e5272c130`, and `087987070`.
- Workflow state: `phase: contract-review`, `current_sprint: upstream-v025-compat-fixes`.

## Verdict

`PASS`。此前六项阻断均已形成可执行范围和验收约束，可进入实现派发；不得扩大 allowlist。

## Review Findings Resolved

1. `spec_ref` 已修正为 `docs/workflow/spec.md`，隔离 worktree 的 workflow phase/current sprint 也已切换到本 contract 的独立审查门禁。
2. Redis 运行验收已给出 task-owned Docker 生命周期、PONG/端口检查、环境变量、生产 cache+invalidator 定向测试和精确容器清理。实际 task-owned Redis 已启动并返回 PONG，地址为 `127.0.0.1:60229`；不得改用共享 Redis。
3. 序号验收已明确覆盖自定义 `wireBase()` 的首帧零值、默认 terminal 序列化、compact 递增、handler 失败、WS/HTTP fallback 和非零上游值保留。`responses_stream_event_wire.go` 的 `wireBase()` 当前已无条件写出 `sequence_number`，无需为本任务改动该生产文件，测试必须证明该行为。
4. 本地 SSE 实际 owner `handleGeminiStreamingResponse` 已被锁定为 `antigravity_gateway_service.go`；验收要求 LF/CRLF 精确帧边界、非 data 行保留及 heartbeat/cancel/disconnect。
5. OAuth 验收已绑定 map transform、HTTP raw-body 和 `normalizeOpenAIResponsesWebSocketCompatibilityBody`。该 WS 归一化函数确实位于本地 `openai_responses_compatibility.go`，并由 `openai_ws_v2_passthrough_adapter.go` 调用；该调用点保持只读。测试必须证明仅 OAuth 的 input 数组对象顶层字段被移除，且 API-key 路径、嵌套/文本同名内容与非对象元素保持不变。
6. 已允许在 tagged fixture 基线漂移时使用实际生产源码的定向测试，同时禁止以重实现或行为 stub 规避验证。当前无 tag 生产包可编译；无关基线失败仅为 `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested` 的固定 hash expected/actual 不一致。报告须记录精确命令、相关测试集合和该未改动的基线证据；不得修复或掩盖它。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes` (`gpt-5.6-terra`)
- base_commit_confirmed: `yes`
- openspec_traceable: `yes` (governing `docs/workflow/spec.md`)

## Approval Boundaries

- 只允许 contract 的 allowlist；禁止上游 merge/rebase/cherry-pick、主工作树修改、迁移、前端、真实 provider、部署、推送或共享 Redis。
- Redis、tagged fixture 或生产源码定向测试任一必需验证无法完成时，worker 必须报告 `BLOCKED`，不得宣称完成。
- 本 review 仅批准进入 Generator 实现；独立 QA 和 controller 最终裁决仍是后续必经门禁。
