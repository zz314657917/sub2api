### DONE: upstream-v0215-protocol-20261009

## Verdict / Findings
- 七类实现完成。2026-10-10 合同 narrow exception 经独立 revision review PASS 后，仅更新 `TestStreamingReasoning` 第 776 行的 SSE expected，明确要求空 `"signature":""`。
- 修订后 full apicompat/antigravity、service focused、生产构建均重新执行且 exit 0；无失败测试。DONE 是本 Developer scope 的真实完成结果，不替代主控/独立 QA 最终裁决。
- 相对 immutable copied baseline，该既有测试文件只有上述一行增量；逐文本核对保留所有 copied ThinkingDisabled hunks。未弱化断言、删除测试或 overlay 该包测试。
- 原 BLOCKED 原因是相关 wire-contract expected 过时，不是 baseline 原失败或 compile drift；旧失败日志保留为历史证据，当前结论由修订后新日志支撑。
- 后续 QA coverage 反馈已处理：只补两个 allowed v0215 测试，新增空 text、CacheControl、thinking/signature 双空的完整 JSONEq 边界。QA 随后纠正 Antigravity 空 text 预期应保持 baseline `{"type":"text"}`，本 worker 未扩展生产语义。纠正后再次 full package exit 0，381 顶层 + 136 子测试 PASS。

## Contract / Scope
- 唯一源码工作区：`E:/codex-worktrees/sub2api/upstream-v0215-20261009`。
- 已读 approved/PASS 合同、独立 PASS review、batch spec、isolated status 和 matrix；按本轮授权 current-session-default 执行。isolated status 仍是旧 contract-draft 快照，未覆盖主控文件；以已批准合同和用户 dispatch 为实施授权。
- 通过 git show 阅读七个指定 commit，只适配选定行为和最小生产 hunks，无整文件替换、cherry-pick、commit、push、deploy、Docker 操作、付费 provider 或数据库操作。
- 未修改 F 主树、其他 worker owner、QA 报告、旧 service 编译 fixture、依赖、配置、知识库；既有包测试仅按新增 narrow exception 修改这一行。
- 所有新增测试文件均有 `//go:build unit`，顶层测试名含 `V0215Protocol`。
- 九个生产路径的外置 baseline 文件哈希均与 baseline.json 一致；已生成相对 copied baseline 的实际增量。三处既有 `Type: "message"` hunks 全部保留。

## Changed Paths (absolute)
生产源码：
1. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
2. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/apicompat/responses_to_chatcompletions.go`
3. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/apicompat/types.go`
4. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/antigravity/claude_types.go`
5. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/antigravity/stream_transformer.go`
6. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/gateway_forward_as_chat_completions.go`
7. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/gateway_forward_as_responses.go`
8. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_gateway_chat_completions_anthropic_native.go`
9. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_gateway_responses_anthropic_native.go`

新增测试：
10. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/apicompat/v0215_protocol_test.go`
11. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/antigravity/v0215_protocol_test.go`
12. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/v0215_protocol_negative_index_test.go`
13. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/v0215_protocol_ping_test.go`

授权修订的既有测试（仅一行）：
14. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/pkg/apicompat/anthropic_responses_test.go`

自身报告：
15. `E:/codex-worktrees/sub2api/upstream-v0215-20261009/docs/workflow/worker-results/upstream-v0215-protocol-20261009-result.md`

## Behavior / Upstream Evidence
| 类别 | git show 来源 | 实现及回归 |
| --- | --- | --- |
| instruction role | 9a8ab4f88 | developer/system 原角色保留；字符串、text parts 保留内容，message type baseline 不丢失 |
| tool_choice | f3d752693 | 仅 named Chat function choice 扁平化；字符串、Responses shape、其他 tool type、空名称原样保留 |
| legacy pairing | 7a767d5ce | legacy assistant call 转 function_call；同名结果 FIFO、全历史现代 ID 预留、确定性 ID、不改输入历史、现代调用优先、orphan 维持原 fallback |
| refusal | 775c1a3c1 | 非流式 message.refusal、流式 delta.refusal、buffer 重建；done 不重复、普通文本独立、已有 terminal output 不覆盖 |
| signature | 412718edf | apicompat/Antigravity thinking JSON 始终带 thinking/signature；两个 Antigravity start 入口带空 signature；普通 text/tool_use 不强加字段 |
| negative index | 15f469c86 | 四个 buffered owner 统一 lower/upper bounds；fake SSE 包含负 text/thinking/input_json index 和正越界 index，正常文本继续输出 |
| SSE heartbeat | 9f84204c2 | ping 输出 `: ping\n\n`，立即 Flush、不计 first token、不泄漏 Anthropic event；写失败立即停止 downstream，正常文本/usage/DONE 保留 |

本地结构适配：apicompat baseline 尚无 AnthropicContentBlock.Signature，故在 allowlist types.go 补字段；本地 streaming handler 有 includeUsage 参数、新 Responses handler 无上游 tool mapping 参数，新增测试按真实本地签名适配，不修改旧 fixture。

## Executed Commands / Results
以下 Go 命令 cwd 均为 `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend`；JSON 输出外置保存。计数为顶层 Test 函数，子测试单独统计，不以 no-tests-to-run 算 PASS。

| 真实命令 | exit | 最终结果 |
| --- | --- | --- |
| `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity -run V0215Protocol -count=1 -json` | 0 | 边界补充前的 focused 历史：apicompat 6 顶层 + 14 子测试 PASS；Antigravity 2 顶层 PASS |
| `go test -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json ./internal/service -run V0215Protocol -count=1 -json` | 0 | revision 后重跑：4 顶层 + 6 子测试 PASS，真实执行四个 buffered owner 和三个 heartbeat 场景 |
| `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity -count=1 -json` | 0 | 最新 coverage baseline-corrected 重跑：apicompat 246 顶层 + 75 子测试 PASS；Antigravity 135 顶层 + 61 子测试 PASS；总计 381 顶层 + 136 子测试，无失败 |
| `go build -o E:/codex-runtime/upstream-v0215-20261009/protocol-server.exe ./cmd/server` | 0 | revision 后重跑：产物 149214208 bytes，时间 2026-10-10 00:25:34 +08:00；仅构建，未启动/部署 |
| `git diff --check` (worktree root) | 0 | 无 whitespace error；Git 提示共享 status.md 的 LF/CRLF，不修改它 |
| `gofmt -l <上述13个Go文件>` (worktree root) | 0 | 输出为空，自有源码/测试格式合规 |

当前新增共 12 个顶层 Test + 34 个子测试：最新 full package 实际执行其中 8 顶层 + 28 子测试并 PASS；未改动的 service 上次实际执行 4 顶层 + 6 子测试并 PASS。首次自有测试编译/运行暴露本地 Signature 字段、构造函数/stream handler 参数差异和 reasoning-start fixture 使用错误；已只在自有 allowlist 修正。revision 之前 full package exit 1 仅剩旧 TestStreamingReasoning；批准单行 expected 修订后已全包重测通过，早期失败日志不作为最终通过证据。

## QA Coverage Follow-up — 2026-10-10
- 修改范围仅 `backend/internal/pkg/apicompat/v0215_protocol_test.go`、`backend/internal/pkg/antigravity/v0215_protocol_test.go` 和自身报告。对上一轮 14 个自有 Go 路径逐一 SHA256 比对，只有上述两个测试变更；9 个生产源码、service 测试和窄范围旧 expected 均未变。
- 两个 TestV0215ProtocolThinkingSignature 使用有名称的完整 JSONEq 子测试；apicompat 8 cases、Antigravity 6 cases，普通文本/tool_use 必须保留原有字段且不得多出 thinking/signature。
- apicompat：空 text 必须含 `"text":""`；空 text + cache_control 保留 Type/TTL `"1h"`；双空 thinking/signature + cache_control 保留 ephemeral 且空 TTL 不额外输出。
- 双空 thinking/signature：两边完整预期均为 `{"type":"thinking","thinking":"","signature":""}`。
- 第一次按 QA 初始过宽要求给 Antigravity 空 text 也要求 `"text":""`，focused exit 1；随后完整包重跑 exit 1，失败为 `TestV0215ProtocolThinkingSignature/text_empty_required`。这是真实已执行失败，不隐瞒、不作为 PASS。
- immutable baseline 的 ClaudeContentItem.Text 已为 `json:"text,omitempty"`，所以空 text 本就输出 `{"type":"text"}`。QA 明确回撤过宽要求后，该 case 改名 text_empty_baseline，完整 JSONEq 要求原 wire shape；这是 baseline 归因后的正确预期，不是生产修复或本批 regression。没有为通过测试扩展 Antigravity text 语义。
- 随后执行上述最新 full package 命令 exit 0。仅测试改变，未再次运行 service/build；它们的最近真实结果仍来自上一轮 revision 日志。独立 QA 将另行重跑，本 worker 不编辑 QA 报告。

## Baseline Fixture Overlay Review
- 原始旧 fixture 错误、精确生成逻辑：`E:/codex-runtime/upstream-v0215-20261009/baseline-checks.md`、`E:/codex-runtime/upstream-v0215-20261009/create-baseline-overlay.py`。
- 原错误包括重复 stringPtr、旧 billing 参数、旧 buildCountTokensRequest 返回数量、已删 Proxy 字段、email-bind helper/interface 漂移。
- 使用主控现成 `E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json`，本 worker 未修改 overlay 或任何旧 fixture。
- 逐一比对原/替换测试名称及数量：ops_health_score 5、billing_service 71、billing_service_unified 9、gateway_context_management 43、proxy_update_probe_invalidation 1、auth_service_email_bind 13，总计 142/142 保留。
- Replace 的六个入口全部为旧 `_test.go`，没有生产代码或本轮 v0215_protocol 测试入口；所有完整替换文件在该 fixture-overlay 目录中。旧 fixture 本身的运行正确性不由此次 focused run 背书。

## Evidence Artifacts (outside repo)
- 最新边界 baseline-corrected 全包：`E:/codex-runtime/upstream-v0215-20261009/protocol-coverage-baseline-corrected-pkg-all.jsonl`
- 初始过宽预期的 focused 真实失败：`E:/codex-runtime/upstream-v0215-20261009/protocol-coverage-signature.jsonl`
- 初始过宽预期的 full package 真实失败：`E:/codex-runtime/upstream-v0215-20261009/protocol-coverage-pkg-all.jsonl`
- 边界补充前 revision 全包 PASS：`E:/codex-runtime/upstream-v0215-20261009/protocol-revision-pkg-all.jsonl`
- 当前修订后 service：`E:/codex-runtime/upstream-v0215-20261009/protocol-revision-service.jsonl`
- 当前修订后 build：`E:/codex-runtime/upstream-v0215-20261009/protocol-revision-build.log`
- 单行 expected 相对 copied baseline 的 diff：`E:/codex-runtime/upstream-v0215-20261009/protocol-revision-test-baseline.diff`
- 下列 pre-revision 日志保留用于执行历史，其中 pkg-all-final 的旧失败已由上述新全包日志取代：
- `E:/codex-runtime/upstream-v0215-20261009/protocol-pkg-focused-final.jsonl`
- `E:/codex-runtime/upstream-v0215-20261009/protocol-service-final.jsonl`
- `E:/codex-runtime/upstream-v0215-20261009/protocol-pkg-all-final.jsonl`
- `E:/codex-runtime/upstream-v0215-20261009/protocol-build-final.log`
- `E:/codex-runtime/upstream-v0215-20261009/protocol-own-baseline.diff`
- `E:/codex-runtime/upstream-v0215-20261009/protocol-owned-hashes.json`
- `E:/codex-runtime/upstream-v0215-20261009/protocol-overlay-review.json`

## Remaining Risks / Recommendation
- 原 signature expected 阻塞已解除；可继续交给主控/独立 QA 核验，本 worker 未修改 QA 报告或代替 QA 裁决。
- service 使用精确 compile-only overlay；未跑完整 service suite、未验证真实 provider、未做部署/runtime server acceptance。fake SSE 测试确实调用生产 handler，但不等于真实端到端 provider。
- 构建发生在并行共享的 isolated checkout，是当时共同源码状态的生产构建，不是其他 worker 的验收结论。
- 只交付本 worker scope；后续整合、独立 QA、F 树冲突/哈希检查归主控。无 knowledge_candidates。
