### PASS: security-prompt-normalization

# Contract Review

## Task ID
security-prompt-normalization

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/security-prompt-normalization.md`
- Upstream candidate `bef505942df0f0f68b7f204e930dcd1efcec3d76`, limited to its `prompt_snapshot.go` / `prompt_snapshot_test.go` normalization hunk.

## Findings
- 可独立移植。候选仅在 `buildPrioritizedScanText` 对已提取的文本段执行 NFKC、指定零宽字符剔除、指定控制字符空格替换、外层 trim 与空段过滤，再由既有快照字段生成 scan text、metadata、hash、preview 和 full prompt。当前基线的 `ExtractPromptSnapshot` 仅反序列化 `Request.Body`，该路径不写回请求字节；异步入队与阻断 Guard 均消费 `Snapshot.ScanText`，因此无需触及 provider body 或网关转发路径。
- `bef505942` 的其余行为是 `prompt_service.go` 的进程内 session-risk 计数/阻断及新增 `prompt_session_risk*.go`；均不在本合同 allowed paths 内。本批不得导入这些内容，也不得声称已合入整个 source branch；Composite gateway ownership 同样仍待独立审查。
- 候选源提交父节点 `07619d491...` 不是本地基线 `c30bb535...` 的祖先，故只能按合同逐段行为移植，不能 cherry-pick 或整体合并。`golang.org/x/text v0.39.0` 已在当前 `backend/go.mod` 模块图中，`norm.NFKC` 不需要扩大依赖/allowlist。
- 全部段在规范化后为空时，候选的既有语义是：因规范化前 `segments` 非空而返回 snapshot；`ScanText`、metadata、`FullPrompt`、preview 为空，`PromptHash` 为空串 SHA-256，`PromptLength` 为 0，`MessageCount` 保持规范化前的段数。`buildPrioritizedScanText` 对零个 retained segment 不索引数组，故无 panic。开发和 QA 必须以专门断言锁定该源行为，不能擅自把它改成 `ErrNoPromptText` 或重定义 `MessageCount`。
- 合同已要求开发阶段补足候选上游测试未覆盖的跨协议、空段、控制字符、优先级、原始 `Request.Body` 不变、redaction 与有效 scanner 路径测试；其中必须同时覆盖上述全空快照语义及一个非空规范化绕过文本实际传入 scanner 的断言。这些是 Generator/QA gate 的必需证据，不是当前合同审查的通过声明。

## Gate Checks
- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes` (`go test ./internal/securityaudit -count=1` 在基线通过；其余命令留待构建/QA gate 执行)
- worker_model_confirmed: `yes` (`gpt-5.6-terra`)
- base_commit_confirmed: `yes` (`c30bb5350006490eaf0dac314e6dc66746818930`)
- openspec_traceable: `not-applicable`

## Approval
- 合同限定的文本规范化可进入 Generator build。实现只能修改允许的 snapshot 生产/测试文件；不得修改 `Request.Body`、provider payload、角色选择、Guard policy/action、持久化/redaction gate、session-risk 或 Composite。
