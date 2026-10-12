### PASS: upstream-v0215-protocol-20261009

## Findings
- 最终未发现本 Protocol 合同范围内仍未解决的实现缺陷或必需行为覆盖缺口。PASS 仅指 scoped local QA，不指全库、真实 provider 或部署验收。
- 原 `TestStreamingReasoning` 失败不能归为 baseline blocker。已核对 approved child 合同及独立 PASS revision review；Developer 只更新相关 SSE expected，明确要求 `"signature":""`。QA 独立重跑原场景及完整两包，均 PASS；未跳过、overlay 或弱化该测试。
- QA 发现 signature 边界缺少空值及 CacheControl 保留断言，已向原 Developer 报告，由 Developer 补到自有测试；QA 未改库内测试或生产实现。
- 补覆盖中曾出现一次真实 FAIL：QA 初始要求 Antigravity 空 text 必须带 `text:""`，但其 immutable baseline 的 Text 本就 `omitempty`。纠正了过宽的 QA 要求，由 Developer 将该新断言改为原 wire shape `{"type":"text"}`，未改生产语义。apicompat 原本强制保留空 text，最终仍严格要求 `text:""`，同时验证 CacheControl/TTL 不丢失。失败日志保留，不以“baseline 可绕过”代替重跑。
- service 无 overlay 的 fresh compile probe exit 1，0 tests 执行；这是已知既有 fixture 漂移，不能称该次检查 PASS。最终 service 结论明确依赖六文件 test-only overlay。
- `anthropic_responses_test.go` 的 gofmt 非规范来自 copied baseline 的 ThinkingDisabled hunk；baseline 与 candidate 均被 `gofmt -l` 列出。QA 不格式化旧测试，逐文本证明 candidate 相对 copied baseline 仅允许的一行断言增量。

## Contract / Independence / Scope
- 唯一源码工作目录：`E:/codex-worktrees/sub2api/upstream-v0215-20261009`；HEAD/合同基准为 `ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b`。
- 已读取 approved QA 合同、独立 PASS QA review、Protocol child 合同及其 narrow-exception PASS review、Developer result、batch spec、isolated status 和 matrix。本批 current-session-default 模型例外来自 spec/显式派发；没有修改矩阵。
- 本 QA 独立于 Developer；自行读断言、审 baseline delta、运行测试、统计 JSON events 和生产构建，不采用 Developer 自述作为通过证据。
- 本 QA 仅写本报告以及 `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/` 证据/外置构建输出。未写业务源码、库内测试、依赖、overlay、其他 QA 报告、F 主树、知识库；无安装、commit、push、部署、数据库或容器操作。
- 向原 Developer 的 coverage 消息成功送达。向主控 native ancestor 的消息工具明确拒绝，本 QA 在过程更新及本报告同步 Findings，主控可直接接收最终交付；没有声称工具送达。

## Exact Baseline Delta Review
- QA 独立验证 10 个既有 Protocol owner 文件的外置 baseline SHA256 均等于 `baseline.json`，并生成完整相对 copied baseline 的 diff；另有 4 个新增 unit 测试文件。
- 九个生产 owner 的实际增量仅涉及 instruction role、named tool_choice、legacy pairing、refusal、signature、四处 index lower bound 和 ping 写出。不是相对 HEAD 的混合 dirty diff，也未整文件替换上游。
- copied `chatcompletions_to_responses.go` 三处 `Type: "message"` 保留；copied ThinkingDisabled/OutputConfig/effort none/Summary hunks 逐文本保留。
- `anthropic_responses_test.go` 满足严格等式：candidate = immutable baseline 仅替换一次旧 content_block expected，加入空 signature；其余字符不变。
- 全部 baseline 中 apicompat/antigravity 既有文件逐哈希扫描，无 allowlist 外新增 delta；service 仅审本 Protocol 四个明确 owner，其他并行 owner 不冒认本 QA 已验收。
- 最终测试/build 前后 14 个 Protocol 源码/测试路径哈希完全一致。共享 checkout 的其他 worker 变动及 F 主树整合冲突检查仍归主控。
- 原 baseline 包测试/production build PASS 来自主控外置 `baseline-checks.md`；本 QA 没有把该历史记录表述为本轮亲自重跑 baseline。

## Behavior / Boundary Assertions
下表每项均有非零实际测试执行；测试名与全部子测试事件见独立 JSON 日志。

| 修复 | 实际断言 / final execution |
| --- | --- |
| instruction role | `TestV0215ProtocolInstructionRoles`：system/developer × string/text-parts 四个子测试，角色、message type、内容 JSONEq；4/4 PASS。 |
| tool_choice | `TestV0215ProtocolToolChoice`：named function trim/flatten；auto/none/required、Responses flat shape、allowed_tools、空名、custom 均不变；8 个子测试 PASS，畸形 JSON passthrough 断言亦执行。 |
| legacy pairing | `TestV0215ProtocolLegacyPairing`：同名 FIFO、交错不同名称、后续现代 call_legacy_1 及 result-only call_legacy_2 预留，因此新 ID 为 3/4/5；确定性、modern 优先、orphan fallback、空 arguments、输入历史不变均断言；1 个顶层 PASS。 |
| refusal | Nonstream 顶层 PASS；StreamAndBuffer 的 refusal-only/mixed 两个子测试 PASS。stream marshal/unmarshal 的 delta.refusal 拼接、done 不重复、empty delta、HasContent、buffer 重建、普通 text 独立、已有 terminal output 不覆盖均断言。 |
| signature | apicompat 8、Antigravity 6 个 JSONEq 子测试 PASS，含 thinking 双空/非空、普通 text、tool_use 无新增 thinking/signature 字段；apicompat 空 text/cache_control/ttl 及 thinking cache_control 完整保留；Antigravity 空 text 保持 baseline omission。Responses reasoning content_block_start 及 Antigravity 两个 start 入口均验证空 signature。 |
| negative index | 四个 buffered owner 全部通过生产 handler fake SSE 调用：负 text/thinking/input_json 和正越界 index 均被忽略；每个 owner NotPanics/NoError、正常 Hello 保留、IGNORED 不泄漏；4/4 子测试 PASS。 |
| ping | PingFlushesImmediately、PingDoesNotSetFirstToken、PingWriteFailureStopsDownstream 三个顶层 PASS，后者含 message_start 前/后两子测试。io.Pipe 证明等待下一事件/EOF 前已 Flush；四个 `: ping\n\n` 注释、不泄漏 Anthropic event、ping-only FirstTokenMs nil、正文/usage/单次 DONE 保留；写失败后零写入/零 flush、无 DONE、后续 usage 未处理。 |

## ExecutedChecks / Counts / Exits
所有 Go 命令 cwd：`E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend`。Go：`go1.26.3 windows/amd64`。次数均为 fresh `-count=1`；顶层和子测试分开计数，不重复求和。所有 final 日志及 `.exit` 均在上述 QA 专用证据目录。

| 命令 | exit | 实际结果 / evidence filename |
| --- | --- | --- |
| `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity -count=1 -json` | 0 | `pkg-all-final.jsonl`：apicompat 246 顶层 +75 子测试、Antigravity 135 顶层 +61 子测试，全部 PASS，0 FAIL/skip。 |
| `go test -tags unit ./internal/pkg/apicompat -run '^TestStreamingReasoning$' -count=1 -json` | 0 | `original-failure-retest-final.jsonl`：原失败场景 1 顶层 PASS。 |
| `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity -run V0215Protocol -count=1 -json` | 0 | `pkg-focused-final.jsonl`：apicompat 6 顶层 +22 子测试；Antigravity 2 顶层 +6 子测试，全部 PASS。 |
| `go test -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json ./internal/service -run V0215Protocol -count=1 -json` | 0 | `service-focused.jsonl`：4 顶层 +6 子测试 PASS；覆盖四个 owner 与三项 ping 场景。 |
| `go test -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json ./internal/service -run '^(TestV0215Protocol\|TestHandleCCBufferedFromAnthropic_\|TestHandleCCStreamingFromAnthropic_\|TestHandleResponsesBufferedStreamingResponse_\|TestHandleResponsesStreamingResponse_\|TestExtractCCReasoningEffortFromBody$\|TestExtractResponsesReasoningEffortFromBody$\|TestToolArgumentsAreValidJSON$\|TestAppendRawJSON_EmptyObjectPlaceholder$)' -count=1 -json` | 0 | `service-regression-final.jsonl`：15 顶层 +9 子测试 PASS（含上述 4+6 和 11 个既有顶层回归）；表内 `\|` 仅 Markdown 转义，实际命令为正则 `|`。 |
| `go build -o E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/server-final.exe ./cmd/server` | 0 | `build-final.stdout.log`/`build-final.stderr.log` 空；149214208 bytes，SHA256 `2B6F43FB23E9654BE6672484E7BB60546BA75391B7CFBF2215F08EDCB083BBE5`；无 overlay 的真实生产 build，未启动。 |
| scoped `git diff --check -- <Protocol package dirs / four service owners / two new service tests / own report>` | 0 | `diff-check-final.log` 及 `report-tracked-diff-check.log`；无 whitespace errors。新增 untracked 报告另用 no-index 检查，见下一项。 |
| `git diff --no-index --check -- NUL docs/workflow/qa-reports/upstream-v0215-protocol-20261009-qa.md` | 1 | `report-diff-check.log`：新文件相对 NUL 有差异导致 no-index exit 1；只输出 LF/CRLF 提示，无 whitespace errors，不将其伪称 exit 0。 |
| `python E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/audit.py` | 0 | `audit-final.log`：baseline/one-line assertion/copied hunks/overlay names/counts 全部严格断言通过；`final-state.json` 证明 final checks 期间 14 owner 哈希稳定。 |
| 14 owner `gofmt -l`（仅检查，不改） | 0 | `gofmt.log` 只列既有 `anthropic_responses_test.go`；对应 immutable baseline 同样被列出，见 `gofmt-baseline.log`。其他 13 owner 输出为空。 |
| `go test -tags unit ./internal/service -run V0215Protocol -count=1 -json`（无 overlay） | 1 | `service-without-overlay.jsonl`：既有 fixture compile FAIL，实际 0 tests，明确不算通过。 |
| 补充边界初轮 focused（历史失败） | 1 | `pkg-boundary-retest.jsonl`：apicompat 6 顶层 PASS；Antigravity 1 PASS/1 FAIL，原因是 QA 过宽空 text expected；由原 Developer 按 immutable baseline 纠正新断言后，final focused/full 两次均 PASS。 |

新增 Protocol 测试最终合计 **12 个顶层 +34 个子测试 PASS**；完整两包 **381 个顶层 +136 个子测试 PASS**。focused 与 full 为重复执行，不相加冒充更多覆盖。

## Fixture Overlay / Original Limits
- QA 未改 `E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json` 或替换内容；逐映射验证 source SHA 等于 immutable baseline，所有入口均为六个既有 service `_test.go`，不包含生产源码、新 Protocol tests 或原 signature 失败测试。
- 原/替换 Test 函数名称及顺序逐项完全相同：ops_health_score 5、billing_service 71、billing_service_unified 9、gateway_context_management 43、proxy_update_probe_invalidation 1、auth_service_email_bind 13；**142/142 保留**。这不是 142 tests 已执行的声明。
- 精确变动：health 重复 helper 改名；billing 旧签名适配/删除 unused context import；countTokens 三返回值改二；proxy 删除已不存在字段；email-bind 添加缺 helper、Update 参数适配并嵌入接口补未用方法。原 assertions 不删除；接口嵌入可能在未经验证的调用路径 panic，不能据此背书 email-bind runtime。
- fresh 无 overlay probe 已复现 stringPtr redeclare、billing 多参数、countTokens assignment mismatch、proxy unknown fields/undefined FallbackModeNone；编译器达到 too-many-errors 后停止。email-bind 后续错误来源于原始主控基线记录及 overlay 精确内容，不伪称本次无 overlay probe 同时输出了全部后续错误。
- 保存全六份实际替换内容为 QA 证据 `overlay-content-<fixture>_test.go.txt`；完整逐文本 diff 为 `overlay-exact.diff`；映射、source/replacement hashes、142 个函数名及 overlay JSON hash 在 `overlay-review.json`。
- 它只使 focused service fixtures 可编译；**带 overlay 的局部检查不是全库验收，也未运行完整 service suite**。

## Evidence
- `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/test-counts-and-names.json`
- `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/owned-baseline.diff`
- `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/owned-hashes.json`
- `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/package-scope-review.json`
- `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/overlay-review.json`
- `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/overlay-exact.diff`
- `E:/codex-runtime/upstream-v0215-20261009/qa-protocol-20261010/final-state.json`

## UnverifiedRisks / Recommendation
- 全仓库 unit suite、全 service suite、独立 `internal/handler` suite、真实 provider、HTTP server 启动、部署/容器/数据库和 F 主树整合后状态均未验证、不在本次 PASS 范围。
- fake SSE 确实调用生产转换/buffer/stream handler，但不等价于真实 upstream 网络端到端验收。
- production build 是并行 isolated checkout 在本次时间点的共同源码产物；不替其他 slice 作 QA 结论。
- 原相关 signature FAIL 已经由允许的精确断言更新及独立原场景/full-pkg 重跑解除；不再作为绕过项。建议主控继续 scoped integration gate，并单独完成其他 slice QA 与 F 主树 source-hash conflict check。
