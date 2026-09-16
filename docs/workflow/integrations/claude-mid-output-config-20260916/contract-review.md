### PASS: claude-mid-output-config

# Contract Review

## Task ID

claude-mid-output-config

## Latest Verdict

`PASS` — 2026-09-16 最新窄复审确认：原四项合同问题已闭环，基线例外现严格限于本报告列出的**四个**具名失败；它只适用于本地 scoped integration，不使完整 suite 变绿，也不构成 release approval。

## Contract Checked

- `docs/workflow/tasks/claude-mid-output-config.md`
- local base `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779`
- upstream behavior commit `d8326fccfce4ba011f2fc1e148a0181912a9dfd1`（只读对照）

## Review History

### Latest PASS — 2026-09-16 contract Stop Rules synchronization

复核通过。`docs/workflow/tasks/claude-mid-output-config.md` 已将先前单项 Stop Rule 替换为与已批准
范围完全相等的四项穷尽列表：四个测试名、源文件行号及三项断言差异均一致；call-ID 项现在记录了
完整 expected/actual hash，属于此前缩写指纹的精确化，并未增加新失败类型。

同步文本仍强制执行原始未过滤的 full suite、保留/report exit `FAIL`、用完整 `go test -json`
`Action=fail` 事件枚举失败集合，并将额外失败或 fingerprint 改变设为阻断。它继续禁止断言弱化、
validator/业务测试修改、full-suite green 或 release claim，也没有豁免 CCH cache 污染或新的 candidate
failure。故 pending re-review 标记现由本段解除，例外范围与审查批准范围无冲突。

### Latest PASS — 2026-09-16 additional exact baseline-failure amendment

`docs/workflow/baseline-validator-exception.md` 新增的三项 proposal 经本独立复审批准，且仅与
前次已批准的 validator 失败共同构成允许集合：

1. `TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation` —
   `claude_code_validator_test.go:52`, `Should be false`。
2. `TestApplyCodexOAuthTransform_PreservesAllowedTools` —
   `openai_codex_transform_test.go:18`, `Should be false`。
3. `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested` —
   `openai_codex_transform_test.go:275`, expected `fc_6336...`, actual `fc_4774...`。
4. `TestParseSSEUsage_SelectiveParsing` —
   `openai_gateway_service_test.go:3259`, expected `9`, actual `1`。

审查证据：新增三项在 clean detached base
`dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779` 中由 controller 与独立 Claude QA 使用精确 anchored
selector 同时复现；当前源文件中的测试声明、行号和断言类型与证据文件一致。它们不在本 task 的
business allowlist，且没有被 candidate implementation 修改。

**强制边界：** 仍须运行原始、未过滤的 `go test ./internal/service -count=1`，保留其实际 exit
status `FAIL`，并用完整输出（优先 `go test -json` 的 `Action=fail` 记录）枚举全部 fail event。
只有失败集合与以上四项完全相等时，才可在其余合同门禁均通过后报
`scoped local integration PASS with documented baseline exceptions`；任何额外失败、缺失/不一致的
断言身份、编译失败、超时或日志截断均为阻断，不能被本例外吸收。

**未豁免项：** CCH cache 污染是本 task 新增测试的隔离缺陷，绝不是基线例外；Developer 必须在
允许的新测试文件中修复并由 QA 重验。不得修改上述四个测试、其业务实现、validator 或其他 denied
path 来消除失败。本修订亦不授权 provider、数据库、容器、部署、提交或 release。

本段取代下一个历史段中“只允许 validator 单项”的当时范围；基线 commit、四项测试身份、断言位置
或任务 allowlist 任一变化，均要求新的独立合同复审。

### Latest PASS — 2026-09-16 baseline-validator exception supplement

- 已只读核对 `docs/workflow/baseline-validator-exception.md` 与 clean detached
  `E:/codex-worktrees/sub2api/integration-baseline-dc851ea3f`：后者 HEAD 为
  `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779`，与 contract 的 `base_commit` 完全相同，且
  没有 candidate implementation。
- 基线命令、测试名和断言位置均被精确锚定：
  `go test ./internal/service -run '^TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation$' -count=1`
  返回 exit 1，失败为 `claude_code_validator_test.go:52: Should be false`。这不是将一类
  validator 问题泛化为允许失败的豁免。
- 合同 Stop Rules 仍要求运行未过滤的 `go test ./internal/service -count=1`，并明确完整 suite
  不能宣称 green。只允许该**同名既有**失败留存；任何额外 test failure、编译错误、超时或
  未能取得完整 suite 输出均为阻断，不能借例外 PASS。
- 例外不授权修改 `claude_code_validator_test.go`、validator 实现或任何 denied path；它也不影响
  新 `MidConversation` selector、四 builder、CCH、`go build ./...`、格式和范围门禁。QA 必须原样
  报告 full-suite 的 exit/status 与所有失败测试名，最终报告只能称“scoped local integration PASS
  with documented baseline exception”（若且唯若失败集合正好为上述一项）。

该例外在本独立复审通过前不生效；本段即为批准记录。它只适用于此 task 的当前
`dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779` 基线，基线、命令、失败身份或范围任一变化时必须重新审查。

### Historical FAIL — 2026-09-16 initial review

以下四项为初审发现，均已由当前 contract 修订解决：

1. **P1 — 新回归测试是否会被验收命令执行未定义。** 上游的
   `gateway_mid_conversation_output_config_test.go` 首行是 `//go:build unit`；本地
   `gateway_context_management_test.go` 也带该 tag。然而合同两条 service test 命令均
   未带 `-tags unit`。若实现者沿用上游文件 tag，新增 selector 会匹配零个已编译测试，
   `go test` 仍可能成功，违反 Success Criterion 5 的“new tests cannot pass acceptance”意图。
   **修订：** 明确新
   `gateway_mid_conversation_output_config_test.go` 不得带 build constraint，或把两条
   service 命令均改为 `go test -tags unit ...`；并把 focused command 收紧为实际、非空的
   新测试名，例如：
   `go test ./internal/service -run '^(TestSanitizeAnthropicBodyForBetaTokens_MidConversationOutputConfig_(MultiMessage|EmptySystemVariants|SystemWithBodyKept|ByteNoopWhenBetaPresent|NoFieldByteNoop|Idempotent)|TestBuildUpstreamRequestOAuthMimic_MidConversationOutputConfig|TestBuildUpstreamRequestAnthropicAPIKeyPassthrough_MidConversationOutputConfigConsistentWithClientHeader|TestBuildCountTokensRequest.*MidConversationOutputConfig)$' -count=1`。
   若选择 unit tag，命令需写作 `go test -tags unit ...`，并在结果中逐项报告匹配的测试名。

2. **P1 — “APIKey and count_tokens”没有绑定到具体 builder/情形，无法证明 header/body
   symmetry。** 本地有四条不同的相关 builder：
   `buildUpstreamRequest`、`buildUpstreamRequestAnthropicAPIKeyPassthrough`、
   `buildCountTokensRequest`、`buildCountTokensRequestAnthropicAPIKeyPassthrough`。前两条和
   后两条的 beta 计算、header 透传及 body 清理顺序不同；另有 native count_tokens 直建
   request，当前不调用 sanitizer。当前文字无法判断验收要求的是标准 Anthropic 两条
   count_tokens builder，还是也要求 native 路径。
   **修订：** 将 Success Criterion 4 改成可判定矩阵，至少要求：
   (a) OAuth mimic `buildUpstreamRequest` 默认保留、policy drop 后删除；
   (b) `buildUpstreamRequestAnthropicAPIKeyPassthrough` 客户端 beta 带/不带时分别保留/删除；
   (c) `buildCountTokensRequest`（OAuth mimic 及 drop）和
   `buildCountTokensRequestAnthropicAPIKeyPassthrough`（客户端带/不带）各验证一次；每例均
   断言 outgoing `anthropic-beta` 的精确 token 状态、messages 长度/顺序和顶层
   `output_config.effort`。明确 native count_tokens 是否 denied；如 in scope，必须把其
   sanitizer 缺失列为允许修改路径，否则明确不对其作“all count_tokens”声明。

3. **P2 — CCH/signing 保证不可由当前合同的具体测试复核。** 合同要求“existing signing
   order”，但未要求开启 CCH 的真实 builder 场景，也未列出应复用的已有
   `TestSanitizeMustBeBeforeCCHSigning_HashConsistency`（该测试同样为 `unit` tag）。只测
   常规 builder 不能证明新增 messages 重建发生在签名之前。
   **修订：** 明确新增一个启用 CCH 的 builder test，或把已有 CCH hash-consistency test
   纳入带 `-tags unit` 的接受命令；测试须使用缺 beta、会删除 message-level
   `output_config` 的 body，并验证 CCH 基于最终 body。若本 Sprint 不修改 signing，范围可
   保持 `gateway_service.go` denied，但该 test 必须在允许的新测试文件中完成。

4. **P2 — “保留 messages byte-for-byte”需限定为本增量未触发重建的场景。** 现有
   sanitizer 先独立清理 `context_management`、`thinking.block_binding`、fallback 字段；它们
   在 beta 缺失时本来就可改变完整 body。上游实现能保证的是：目标 beta 存在时其新增 helper
   为 byte no-op，且没有任一 message-level `output_config` 时该 helper byte no-op，不能泛称
   任意输入 body 的全部 messages 都原样。
   **修订：** 将 criterion 表述为“在不触发既有 sanitizer 字段的 fixture 中，target beta
   存在或 messages 中无 `output_config` 时，比较原始 `[]byte` 与返回值完全相等；target beta
   缺失时只允许 `messages[].output_config` 删除及无正文 system 控制消息删除”。同时要求 malformed
   JSON fixture 返回原 bytes/`changed=false`，避免 fast-path 与不完整 JSON 的边界回归。

## Latest Re-review Evidence

- 新的测试文件被合同明确要求不得带 build constraint，且所有新增测试名必须包含
  `MidConversation`；`go test ./internal/service -list 'Test.*MidConversation'` 是先于执行的
  非空门禁，随后同一 selector 执行。这样不会复现上游 unit-tag 测试在默认 suite 中被静默跳过的问题。
- 四条实际 builder 和各自 default/drop 或 client-token-present/absent 分支均被逐项列出；每例必须
  检查 outgoing beta token、message 顺序/数量和顶层 effort。native count_tokens 已明确 excluded，
  因而不产生错误的全路径覆盖声明。
- 合同要求在允许的新测试文件中启用 CCH 的真实 builder case：缺 target beta 且 message-level
  output_config 被删除，并验证签名基于最终 sanitized body。既有 signing 顺序不需要改动
  `gateway_service.go`。
- byte-noop fixture 已限制为不触发既有 context/thinking/fallback 清理的输入；malformed JSON 明确
  要求原 bytes 与 `changed=false`，与保守失败行为相符。新 untagged compatibility cases 同时覆盖既有
  context、thinking 和 fallback 规则，防止新 helper 替换旧 sanitizer 分支。
- `worker_model: gpt-5.6-terra` 与主树 Agent Matrix 的 Developer/QA 指定一致；base commit 仍为当前
  HEAD `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779`。

## Gate Checks — Latest

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes`（Developer/QA 均为 `gpt-5.6-terra`，与 Agent Matrix 一致）
- base_commit_confirmed: `yes`（当前 HEAD 为 `dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779`）
- openspec_traceable: `not-applicable`（spec_ref 指向本 contract；这是行为移植的独立 sprint）

## Positive Scope Notes

- 业务改动限制在常量表和 `gateway_request.go`，不允许改动 `gateway_service.go`，与现有调用点
  在 final beta 计算之后、CCH 签名之前执行 sanitizer 的架构相容。
- 上游 `d8326fccf` 的最小行为（beta 常量加入 mimic list、message-level 字段选择性清理、空
  system 控制消息删除）与本地已有 `sanitizeAnthropicBodyForBetaTokens`/fallback 规则可叠加；
  contract 必须要求保留既有 context/thinking/fallback 分支的 focused regression。

## Approval

合同审查 gate 已通过。业务实现、真实 provider、容器/数据库、部署、提交和推送仍不属于本审查结论，且均未执行。
