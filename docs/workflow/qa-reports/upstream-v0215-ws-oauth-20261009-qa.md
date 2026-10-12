### PASS: upstream-v0215-ws-oauth-20261009

## Verdict / scope

- 独立 QA，仅验收 approved WS/OAuth child contract 的本地候选，不是全库、真实 provider、数据库、部署或最终 integration PASS。
- 已读取 approved QA contract、独立 PASS review、WS/OAuth child contract/review、batch spec，以及 Harvey 的 `DONE` 报告。先只读审查，等待 Developer 完成后才运行最终候选。
- 唯一代码目录：`E:/codex-worktrees/sub2api/upstream-v0215-20261009`；HEAD 为 `ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b`。
- QA 未修改生产代码、repository tests、旧 fixture、overlay、依赖、其他报告或 F 主树；仅新增本报告与 `E:/codex-runtime/upstream-v0215-20261009/qa-ws-oauth*` 证据/cache/tmp/binary。未提交、推送、启动 server 或访问 provider/共享 DB/Redis/container。
- 命令在 2026-10-10（Asia/Shanghai）实际运行，使用 Go `go1.26.3`、`-tags unit`、`-count=1`/显式重复次数；禁用 dependency download，GOCACHE/GOTMPDIR/TEMP/TMP 均外置到本 QA 路径。

## Findings

### 本合同增量

- 未发现本合同新增实现缺陷；12 个新增顶层测试、WS v2 全包、精确 32 个保护回归、外置生产 build 均独立通过。
- 结论不只来自 helper：已审实际 `AfterTurn → async task → RecordUsage → PricingInput.Group` 接线，运行真实 usage worker、现有 RecordUsage 定价/settlement stub 回归，以及 loopback HTTP/WS wire 测试。

### 扩展旧回归实际失败，不能记为 PASS

扩展 regex 实际执行 182 个顶层测试：179 PASS、3 FAIL；含子测试为 207 PASS、3 FAIL、0 SKIP，命令 exit 1。原始失败日志完整保留。

1. `TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost`
   - `billing_service_test.go:717`：`pricing not found for model: long-image-model`。
   - pricing fixture 与 `billing_service.go` 均与 immutable baseline SHA256 一致；overlay 未修改此测试，仅适配该文件另外两处旧函数签名和 unused import。
2. `TestApplyCodexOAuthTransform_PreservesAllowedTools`
   - `openai_codex_transform_test.go:18`：旧断言 `require.False(result.Modified)` 失败。
   - 请求没有 `input`/`web_search_call`；原有 store/stream/instructions 处理本来会标记 Modified。本轮 history hook 返回 false，不改这个请求。
3. `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested`
   - `openai_codex_transform_test.go:275`：旧断言期望 `fc_6336c46cf80eccddb18637bafbdab3b1de6137977fd8f941d2ccc9763b32d`，实际为 `fc_477489beb020a16331ccd2917b0310e7bb315418ee7f92d3781a5e28eb761`。
   - 原逻辑 hash normalized candidate，断言 hash original ID；请求只有 function_call/function_call_output，没有 web_search_call，新 hook 不产生 mutation。

归因方法是只读 baseline/source/test/control-flow 比较：transform 文件去掉唯一新增 option 和 history hook 后与 baseline **逐字一致**，两个旧 transform 测试文件也与 baseline hash 一致。**未替换生产源码重跑 baseline**，不把静态归因写成 baseline 运行证据。上述 3 项为明确记录的既有限制，不通过删/改测试或 overlay 生产实现掩盖；不阻断此 scoped 增量 PASS。证据：`E:/codex-runtime/upstream-v0215-20261009/qa-ws-oauth-evidence/expanded-regression-baseline-attribution.json`。

## Actual wiring / criterion review

### Per-turn billing snapshot

- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/handler/openai_gateway_handler.go:33-143`：same key ID、GroupID、group object ID、platform、subscription type 才采纳新 pricing；nil/lookup error/身份变化回连接 pricing。lookup 使用现存 `APIKeyService.GetByKey` auth cache，未新增重授权/撤销策略。
- `openai_gateway_handler.go:2229-2266`：连接初价先冻结；后续 BeforeRequest 和 BeforeTurn 均调用 idempotent begin。特别核对 passthrough 的真实生命周期：后续回合也没有 BeforeTurn，靠实际 BeforeRequest hook 刷新，而非只给未调用的 helper 接线。
- `openai_gateway_handler.go:2301,2334-2360`：AfterTurn 按 turn take 独立 key，闭包在提交之前捕获，`RecordUsage.APIKey = turnBillingAPIKey`；不是继续传连接 apiKey。
- User、Subscription、trial session/release、QuotaPlatform、SessionID、request IDs/hash/time、ChannelUsageFields、调度/route/security/privacy 输入保持连接来源；只复制 billing fields，保留 group routing 和 RequirePrivacySet 等原配置。ModelPricing、Models/Intervals 和全部价格/MaxTokens 指针独立复制。
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_gateway_service.go:7730-7895,8300-8413`：RecordUsage 从 input.APIKey.Group 获取 group/default/peak/image multiplier，并把同一 group 传入 `PricingInput`/`CostInput`；ModelPricingResolver 优先 snapshot group config，再沿 GroupID 查 channel，最后 catalog/fallback。没有被连接 Gin group 或新查询 group 覆盖。
- 新测试覆盖首轮无 BeforeTurn、身份变化/lookup error/missing metadata、多个新 Before 与旧 After 交错、延迟使用及 nested pricing mutation。真实 UsageRecordWorkerPool 被阻塞后，再发生下一 turn 更新和 parent cancel，旧任务仍保留价格 2、ledger/client IDs、trial 和 detached context。
- 这里真实 worker 测试不等于完整 authenticated handler+DB settlement；实际 handler 接线为代码审查，RecordUsage pricing/ledger/trial/subscription 为现有 stub 测试和未改源码保护证据，边界分开表述。

### Image accounting / repeated terminal

- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_ws_v2/passthrough_relay.go:880-954`：Responses/Chat image details、顶层 usage、managed image_gen fallback，usage 明细优先；负值 clamp，invalid totals 不计入；turn/session 正确累加 input/output/image/cache 明细。
- `passthrough_relay.go:654-668`：explicit response ID 的 repeated terminal guard 位于 usage parse/累加之前。重复事件返回 `terminal=false`，因此 `emitTurnComplete` 不会再次调用；wire 仍由 relay 转发。
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_ws_v2_passthrough_adapter.go:752-760,847-856`：turn 和最终 session 两个 adapter 映射都携带 ImageInputTokens，未只改 relay struct。
- fake-frame Relay 测试完成两个不同 response，再收到旧 response 的 done/retransmission：2 次 callback，image input 合计 60、image output 85、input 220、cached 12、cache-write 3，四个 wire frames 仍转发。真实 loopback 两回合分别为 image input 20/40、output 55，重复终态不产生第三次 AfterTurn。
- repeated ID guard 保留整个 relay 生命周期的 map；按完成回合数增长的内存风险未做长连接压力验收，仍明确保留。

### OAuth tools / HTTP signature

- normal tools、Lite input.additional_tools（位于 trailing compaction_trigger 前）、existing declarations、existing function tools、required/object tool_choice、malformed/empty/string input、large integer exactness、idempotence 均有断言。
- real Forward loopback HTTP 覆盖 OAuth normal/passthrough、Lite、compact、API-key；real coder/websocket loopback 覆盖 OAuth/API-key/Lite 后续 turn 的实际 upstream payload。Compact/API-key 不注入 history tools，existing choices 不被强制改写。
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_gateway_service.go:3739-3759`：HTTP 400 + exact `thinking_signature_invalid` + 三段 verification/decrypt 文本联立；使用原有 once flag/trim/replay rules，未扩大到所有 thinking errors。
- fake HTTP wire 覆盖 exact recovery、再次拒绝最多两次请求、无 encrypted content 不重试、缺 verify/decrypt 任一条件不重试、other code/status 不重试、legacy invalid_encrypted_content 保持；重试剥 encrypted/previous_response_id，保留 summary。现有 WS replay/output-boundary 相关旧回归也实际执行。

## Executed Checks

以下 Go 命令均从 `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend` 执行，全部命令数组、exit、时间、counts 在：

`E:/codex-runtime/upstream-v0215-20261009/qa-ws-oauth-evidence/checks.json`

| Check | 实际结果 |
| --- | --- |
| `go test -tags unit ./internal/service -run '^$' -count=1`，无 overlay | exit 1，旧 fixture compile drift；不是测试 PASS，也不是本轮新缺陷 |
| `go test -json -tags unit ./internal/service/openai_ws_v2 -count=1`，无 overlay | exit 0；39 top-level / 55 含子测试，0 fail/skip，其中 2 个新增 V0215 |
| `go test -json -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json ./internal/handler ./internal/service -run '^TestV0215(WS\|OAuth)' -count=1` | exit 0；4 handler + 6 service = 10 top-level / 56 含子测试，0 fail/skip；不计并行 protocol 测试 |
| 同上精确新测试 `-count=5` | exit 0；10 distinct top-level × 5 = 50 PASS events / 280 含子测试，0 fail/skip |
| 精确 protection regex（下列 P） | exit 0；32 top-level / 43 含子测试，0 fail/skip |
| 扩展 existing regex（下列 E） | **exit 1**；179 top-level PASS、3 FAIL；207 含子测试 PASS、3 FAIL，完整见 Findings |
| `go build -o E:/codex-runtime/upstream-v0215-20261009/qa-ws-oauth-server.exe ./cmd/server` | exit 0，无 overlay；149214208 bytes，未启动 |
| `git diff --check -- <12 owned Go paths>` / `gofmt -l <same paths>` | exit 0 / exit 0，gofmt 输出为空；未格式化/写源码 |
| baseline/source/test/overlay SHA256 | own 12 paths 在初跑、后续 repeat/build 前后全部相同；800 个 protected baseline files 无差异 |

P（以 `./internal/service -run 'P' -count=1` 执行）：

```text
Test(OpenAITokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated|GatewayTokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated|CalculateImageCost_|GetImageUnitPrice_|DefaultBuild_.*Image|ExtractOpenAIUsage_CapturesImageInputTokens|MergeOpenAIUsage_CarriesImageInputTokens|BuildOpenAIUsageJSON_CarriesImageInputTokens|Resolve_GroupPricingOverridesChannel|OpenAIGatewayServiceRecordUsage_(UsesUserSpecificGroupRate|PeakRateAffectsTokenModeImageOutputTokens|DuplicateUsageLogSkipsBilling|DuplicateBillingKeySkipsBillingWithRepo|BillingUsesDetachedContext|BillingRepoUsesDetachedContext|ImageCacheRead)|OpenAIGatewayService_Forward_HTTPIngressRetries.*InvalidEncryptedContentOnce)
```

E（以 `./internal/handler ./internal/service -run 'E' -count=1` 执行，没有略去失败用例）：

```text
^(TestOpenAIGatewayServiceRecordUsage_|TestOpenAIGatewayHandlerSubmit(OpenAI|Mandatory)?UsageRecordTask_|TestCalculateTokenCostForRequest_GroupIDOnlyUsesChannelPricing$|TestResolve_|TestCalculateImageCost_|TestGetImageUnitPrice_|TestMergeOpenAIUsage_CarriesImageInputTokens$|TestExtractOpenAIUsage_CapturesImageInputTokens$|TestBuildOpenAIUsageJSON_CarriesImageInputTokens$|TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost$|TestApplyCodexOAuthTransform|TestFilterCodexInput|TestTrimOpenAIEncryptedReasoningItems|TestOpenAIGatewayService_Forward_(HTTPIngress|WSv2.*(Encrypted|PreviousResponse|ResponseAlreadyWritten))|TestOpenAIGatewayTrialSettlement_)
```

新增 12 个顶层测试实际执行：

```text
TestV0215WSBillingRefreshIdentity
TestV0215WSBillingActualDelayedWorkerContext
TestV0215WSBillingInterleavedDelayedUsage
TestV0215WSBillingMissingConnectionMetadata
TestV0215OAuthHistoryMapAndRaw
TestV0215OAuthHistoryMalformedAndScope
TestV0215OAuthHistoryTransformLiteAndCompact
TestV0215WSHTTPThinkingSignatureRecovery
TestV0215OAuthHistoryFakeLocalHTTP
TestV0215WSLocalOAuthHistoryAndImageAdapter
TestV0215WSImageUsageForms
TestV0215WSImageMultiTurnDuplicateTerminal
```

## Exact delta / overlay

- 相对 `E:/codex-runtime/upstream-v0215-20261009/baseline.json` 和 immutable baseline 文件逐项审 7 production owners + 5 新 test files，不只比较 HEAD diff；复制进来的原有 dirty hunks没有冒充本轮增量。
- 6 个旧 production owners 的差异和新 helper/test 完整内容见 `E:/codex-runtime/upstream-v0215-20261009/qa-ws-oauth-evidence/own-baseline.diff`；实际被测 SHA256 见 `owned-hashes.json`、`before-check-hashes.json`、`after-check-hashes.json`、`final-check-hashes.json`。
- 800 个保护文件包括 Ent、migrations、go.mod/go.sum，以及 pricing/billing/model_pricing/usage_billing/channel/group/welfare/api_key/new_user_trial/openai_privacy 前缀；与 baseline 0 mismatch。没有把其他 Developer 的并行 delta 算作本 QA 修改或验收。
- 使用原批准的 `E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json`，只映射以下旧 fixtures，test names 和顺序逐一相同：
  - ops_health_score：5；billing_service：71；billing_service_unified：9；
  - gateway_context_management：43；proxy_update_probe_invalidation：1；auth_service_email_bind：13。
  - 合计 **142 个原 Test 函数保留**，不等于这 142 个全都运行。
- 已实际复现无 overlay 的 stringPtr redeclaration、旧 computeTokenBreakdown/calculateCostInternal/buildCountTokensRequest 签名和 Proxy stale fields 编译错误。编译器达到 too-many-errors 后未显示的 email-bind 二次错误来自主控 baseline 记录，不冒充本 QA 独立复现。
- 逐份审 overlay 完整 replacement 与 diff：只有旧签名/import/helper/stale fields compile adaptation；不改本轮测试或生产源码，不删除断言。auth fixture 的 interface embedding/helper 不构成 auth runtime 验收。
- 每份完整 replacement 和 diff 均复制为本 QA 证据 `E:/codex-runtime/upstream-v0215-20261009/qa-ws-oauth-evidence/overlay-*`，精确 source/replacement hash 及保留计数在 `overlay-review.json`。原共享 overlay 未改写。
- 二进制 SHA256：`2b6f43fb23e9654be6672484e7bb60546ba75391b7cfbf2215f08edcb083bbe5`；见 `production-binary.json`。

## Unverified Risks / Recommendation

- **Scoped local PASS only**。无 overlay 的 service package 编译仍失败；扩展回归存在上述 3 个实际 FAIL；不能声称完整 service suite 或全库 PASS。
- repeated-terminal map 每个 explicit completed response 保留一项，长连接内存随回合数增长；未做长时间/资源压力测试，也未运行 race detector。无 ID 的终态仍沿原 correlation/fallback 行为，不能据此宣称任意损坏 provider 帧都有新的去重保证。
- 未验证真实 provider、live auth-cache invalidation、完整 authenticated WS handler、共享 DB/Redis ledger settlement、生产调度、容器、浏览器或部署。Go build 仅证明编译，外置 binary 没有启动。
- 用户/主控提供的“batch 64-file scopecheck PASS、F 树 4646 protected hashes unchanged”属于主控证据；本 QA 未读取/重测 F 主树，也未接管整个 batch 的 integration gate。
- 可以交给主控进入本 slice 的 integration/conflict gate，保持这份限制。若未来解决上述既有测试断言或扩大长连接验收，另行授权原 owner；本 QA 不修实现、不修改 tests/依赖，不以本 PASS 扩大为整批或生产发布授权。
