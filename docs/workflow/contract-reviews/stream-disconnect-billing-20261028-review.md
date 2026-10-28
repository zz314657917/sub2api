### PASS: stream-disconnect-billing-20261028

---
type: contract-review
status: approved
verdict: PASS
review_round: R3
task_id: stream-disconnect-billing-20261028
base_commit: 1536c4849533460ab8c6ce81cf9e406c808e1cc4
reviewer: independent-contract-reviewer
requested_model: gpt-6.1-sol
last_verified: 2026-10-28
reviewed_draft_sha256: eac19b8cbcd0de8be858d13388d20ea6ef47ffedf457903f8a93187fadb2e1d4
contract_sha256: e1db724374b6176a72813e465a848759b41a6b8d19d82a5310d48177741b1407
task_tools_sha256: 9ce7044e6fbb04b5280d37ec8af12e3c794b829b19c7c0f0f9b1ac3afa30d77f
config_sha256: 3f1af0fb67794bfcc04b92a37af0559f0a7f92048a8c1146ade52e10aa5c9fbb
targets_before_sha256: 966533a005f64bb32b577ccfdab01bc5e8b38650eb2af364e22a56d125ff4371
verify_results_sha256: 012e8d0863ccdfc8b9f3681e3147e2009c6a1f0a63153a017704ef6c991c51c1
r3_report_path: E:/codex-runtime/sub2api/stream-disconnect-billing-20261028/contract-reviews/review-r3.md
---

## Findings

未发现 R1 两个阻塞点的未闭合项。本轮仅审 R2 对 CR-01、CR-02 的修订，不扩展为实现或最终验收。

- CR-01 已闭合：compare 核对完整 canonical 检查名称集合、重复/缺失检查及完整命令；new 要求 default/unit 各唯一检查及各自完整命令，以 package + test 核对每种 tag 三次通过，保留失败/skip 拒绝。合同 Acceptance Commands 已加入 compare/new 审核器调用。本轮在内存重新执行 selfcheck：原命令 PASS；替换 command、移除 unit tag、重复 check 均 FAIL；未写入伪造 Go 验收文件。
- CR-02 已闭合：合同固定独立 compat pump 位于 Anthropic 允许的 gateway_forward_compat_stream.go，明确禁止修改现有共享 native pump 及其调用者，不依赖本地不存在的上游接口。Anthropic native/passthrough/compat、OpenAI Responses/passthrough 的允许 owner 均受明确排水约束：connected 时保留既有正 gap，零配置禁用；检测首次上游字节前的客户端取消；断开后 idle gap 为配置正值或 180 秒，另有 180 秒总上限，先到者关闭本次 body。终止、流内错误、读错误出口均要求取消 reader、关闭 body，并测试 goroutine 退出。上述约束可在既有四个写集内实施，无需扩入共享 pump owner。
- R1 的 Gemini delivery 验收补充已写入合同：零字节失败、部分 Write、此前成功图片后失败、未完成断开时禁用默认一张回退；不扩大 observer owner。
- canonical review 路径已与本文件一致。task_tools.py/config.json 哈希保持 R1 值，写集未变化。

## Executed Checks

- 读取 R1 归档与 R2 合同/verify_results 差异，核对新增 Acceptance Commands 和 ABI/排水约束。
- 静态核对新版 new 分支的 tag 唯一性、完整命令及 package + test 计数。
- 阅读 verify_command_selfcheck.py/command-identity-selfcheck.json，并在内存重新执行四个 selfcheck 场景；移除其最终文件输出，不修改现有证据。
- 核对四文件 SHA256。批准仅把合同 frontmatter 的 status 改为 approved、review_verdict 改为 PASS；reviewed_draft_sha256 绑定审阅输入，contract_sha256 绑定批准后的合同字节。

## Unverified Risks

本轮未运行 Go 宽回归、业务实现、断流运行态或最终 QA。合同 PASS 只批准实施范围和验收要求，不构成代码、计费正确性或集成 PASS。后续仍须独立 QA、冻结封印及脏工作树保护门禁。

## Recommendation

PASS，关闭 CR-01、CR-02。仅更新本 canonical review 与合同批准 frontmatter；不启动 Developer/QA，不修改业务代码，不推进其他工作流文件，不执行集成、提交、推送或部署。R1 原件保留在外部 contract-reviews 归档。


## 工具补充审查：PASS（2026-10-28）

仅批准 verify_results.py 中 assertions() 的四行补充；业务合同、四个写集和验收命令保持不变。保留上文 R2 结论及归档原件，本补充不构成业务实现或最终 QA 通过。

- R2 verifier 原 SHA256：`b5fe904dd6d0b1e2818c5d4b057994d528b834b171d3852206580c4aa76e5f79`。
- 本次 verifier SHA256：`012e8d0863ccdfc8b9f3681e3147e2009c6a1f0a63153a017704ef6c991c51c1`；当前 frontmatter 的 verify_results_sha256 绑定此版本。
- 变更在保存 Testify 的 Test: 行之后设置 active=None，结束该断言块；后续 Error Trace/源位置仍可重新开启下一块。未过滤 expected、actual、Diff 或 trace。
- 已对照 baseline/service-unit.json 与 qa-baseline/service-unit.json 中 TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy 的原始 failure_output：两段断言完全相同；R2 误纳入其后的应用 JSON 日志，日志 time/loopback port 不同导致漂移。新版保留两段完整断言并正确比较为相同。
- 本 reviewer 另在内存改动第二段 actual，新版仍检测为漂移，确认 Test: 后重新开启的下一段断言没有被忽略。
- 内存重跑 command selfcheck：原命令 PASS；replace-command、remove-unit-tag、duplicate-check、change-assertion 均 FAIL，符合预期；未写入伪造 Go 验收文件。
- 原 qa-baseline-comparison-r2-fail.json 与 verify_results-r2.py 继续保留。未重跑 Go 测试，未独立复核其余基线用例；controller 所述全基线比较 PASS 与本次单条原始断言核验范围分开。

结论：本工具补充 PASS，可使用新哈希绑定的 verifier 继续既定门禁；本次仅修改 canonical review，未修改合同或业务文件。


## R3 ping 测试边界补充审查：PASS（2026-10-28）

本节是当前 R3 结论；上文 R2 及工具补充文字保留为历史结论。独立 reviewer 只审 criterion 10、一个既有测试路径例外和对应工具/保护输入变化，没有重启完整 R2、启动 Developer/QA 或运行 Go 测试。批准的是测试迁移合同，原 freeze23 的 canonical 比较仍为 FAIL。

### Findings

未发现本次窄补充的合同阻塞。亲读 main/immutable baseline 的原测试、candidate 当前 owner 和共享转换状态机，并逐条核对 candidate/service-unit.json 的原始 failure_output，三个旧 case 的矛盾成立：

| 原 case | 原始失败及源码依据 | 允许的最小迁移 |
| --- | --- | --- |
| TestV0215ProtocolHandleCCStreamingFromAnthropic_PingDoesNotSetFirstToken | 原测试第 166 行要求 ping-only EOF 无错误；candidate 返回 `stream usage incomplete: EOF`。compat reader 仅在完整终止事件后成功，failure helper 在无 observed tokens 时返回 nil result。这符合 criterion 3，不能为保留旧 NoError 而伪造成功。 | 断言 incomplete error、无 billable result/成功 terminal；保留 ping comment、即时 flush 和 ping 不构成首 token 的保护。 |
| TestV0215ProtocolHandleCCStreamingFromAnthropic_PingWriteFailureStopsDownstream（两个子项及父项共 3 events） | 原测试第 196 行要求 OutputTokens 为 0，实际为 7；此前的 NoError、writer.failed、失败后零 writes/flushes、无 DONE 断言均已执行通过。reader 在调用 downstream processor 前合并 usage，写失败后标记 disconnect 并继续有界 drain，所以后续 output7 和已有 input10 可保留。旧输出为 0 的要求会造成计费遗漏。 | 保留所有 downstream 禁写断言，改为验证 ClientDisconnect、output7、input10（有 message_start 时）/input0（无时），并保持正常终止的结果语义。不得放宽失败写入保护。 |
| TestV0215ProtocolHandleCCStreamingFromAnthropic_PingFlushesImmediately | 原测试第 107 行已发送含 stop_reason=end_turn/output7 的完整 message_delta；第 108 行继续发送 message_stop，遇到 `io: read/write on closed pipe`。当前 reader 将该带 stop_reason 的完整 delta 作为终止，先处理 usage/转换，再 finalize；pump.stop 关闭 body 并等待 reader 退出。旧 fixture 随后还要求 post-terminal ping、DONE 等到 EOF，和及时终止/关闭的边界冲突。失败发生在后续发送，不能据此判定之前的 ping 未即时 flush。 | fixture 停在实际 terminal，直接等待返回且不依赖 EOF；保留 terminal 前 ping 的即时 flush/comment/TTFT 保护，以及 hello world、input10/output7/total17、finish_reason=stop 和唯一 DONE 断言。terminal 后 ping 不再作为必须消费的输入。 |

这一例外不允许生产代码为 nil Request 测试分支特殊处理，也不允许修改原已失败的旧断言、其他旧测试、filter/tag/count、skip 行为或 verifier。保留上述行为断言并重新执行原命令，是后续迁移必须满足的门禁；本 review 没有批准任何实际测试 diff 或生产实现的最终正确性。

### Executed Checks

- 合同逐行比较 pre-ping-addendum/contract.md：除待批准字段外，仅新增 criterion 10、Anthropic 写集中的单一路径和该路径的保护说明；Acceptance Commands、ABI、其他成功标准与写集不变。
- task_tools.py 字节比较：删除 GROUPS.anthropic 新增的唯一一行后，与已批准原件完全相同。config.json 仅在该组和 allowlist 中增加同一路径，四组仍互斥，共 24 路径；没有新增生产 owner。
- targets-before.json 前 23 条记录及顺序与原件完全相同，第 24 条是 `backend/internal/service/v0215_protocol_ping_test.go`。其 SHA256 `b4498bb644860a7d36933e39110937816cd5da2e73536480da3220651be559c9` 同时匹配 main-before/baseline-before/candidate-before/qa-before 的原始记录及 main、immutable baseline、candidate、pre-migration QA 的当前文件字节；业务测试尚未迁移。
- 原批准的九份输入在 pre-ping-addendum 完整保存。旧 contract/tools/config/verifier 的哈希与旧 review 绑定一致；controller_audit.py、candidate-freeze.json、qa-before.json、verify_results.py 当前字节仍等于归档。原 freeze23 的 23 个 candidate 文件全部匹配冻结哈希。
- 独立重建五条 canonical 命令，核对 baseline、qa-baseline、candidate 的检查名称集合、无重复及完整命令相同；原 ping 三 case/两个子项均在 baseline 和 qa-baseline 执行且未失败。固定 broad unit 命令继续覆盖这些既有测试，不靠改名/过滤/计数排除。
- 在内存重跑 compare baseline candidate，得到与保存的 candidate-comparison.json 完全相同的 FAIL，只有 service-unit 五个新增 ping fail events，无 missing/skip/assertion drift；未覆盖原比较产物。
- 在内存重跑 selfcheck：unchanged PASS；replace-command、remove-unit-tag、duplicate-check、change-assertion 均 FAIL。仅拦截最终证据写入，未伪造或重跑 Go 验收。

### 精确绑定与原件

当前五项输入由本文件 frontmatter 精确绑定；contract_sha256 是仅批准 status/review_verdict 后的字节，reviewed_draft_sha256 是收到的 draft/pending 原件。R3 原件为 `E:/codex-runtime/sub2api/stream-disconnect-billing-20261028/contract-reviews/review-r3.md`，保存本 review 的完整同字节副本。

| 保留输入（pre-ping-addendum/） | SHA256 |
| --- | --- |
| contract.md | 087bf6a54cac40b5606a78e9d45db67dc0be26f5ebf220bcbe2c6188ac237404 |
| review.md | 108770a4f59b4a8ea9389782d2c3555c8a1ae873cdb5d025236a81eeb7450e9c |
| task_tools.py | 81f6bb3c66ed31fd80f17fa08bd8a367469916d1416b007e88c73160ab308c05 |
| config.json | 987151245e58a0d2ee0c77327f5caa4f9457e57631632c2c31cd27e6fe8df264 |
| targets-before.json | bb5726b88a95ed26d664a80682e3cfb7e1f7b0073b03e8af04915fa80d984198 |
| controller_audit.py | bcb972b790ef282a45b79f37592fa51a57dd92c69409fa29a3c1110c5a2d11f5 |
| candidate-freeze.json | 0477fa96488e6f8f9e4a34bbbef9128b61cd07d28b8d65ba63dcc29f57495f43 |
| qa-before.json | 977be609d449a4894e1b0a4c54643ea73fcd3d8d542a8ef504e27f9068b2466e |
| verify_results.py | 012e8d0863ccdfc8b9f3681e3147e2009c6a1f0a63153a017704ef6c991c51c1 |

R2 reviewed_draft_sha256 仍为 `9444ac72b638d31eee387ffc3651cd194f9088c7b597da88623df76ab1aa4566`，其原 review 完整保存在上述 review.md，R2 verifier 的更早旧 hash 继续保留在工具补充段。

| 本轮证据（相对 evidence root） | SHA256 |
| --- | --- |
| contract-reviews/ping-addendum.md | 9ce40b411e9614d63d22d2594339e026201f7d476db6291df1b2c432472130a3 |
| contract-reviews/ping-addendum-inputs.json | 021a0b5d661f74eaace6a946106a634234db25c66644023f12b1f6e7780c019f |
| main-before.json | 2984582db067af5a35080c4bf9982b645819df3d37f6df4d01f9929362911477 |
| candidate-comparison.json（保留 FAIL） | 3b33868fb4cc46fb2f3418d3a68a3bbeeee7ff7f72dffe42ae12cefe8684894f |
| candidate/service-unit.json | 956736bb9a65e03b7a4fefcadb24455e7ecaf4a6adc10bd4e27b647d4d760736 |
| baseline/service-unit.json | 35f20b9a6aa0da46fab9c2f1768efb2f039000d487a8b4204089448454383d37 |
| qa-baseline/service-unit.json | 2dfbdbbc8c19957e0baeb59916b28a7582687db261d0c6d354a58ffd069ed2d3 |

### Unverified Risks / Recommendation

PASS，仅批准上述三个 case 的合同边界迁移及新增测试路径。本轮写入仅为 canonical review、合同两个批准字段和外部 R3 review 原件；没有修改业务/测试/工具、移动 QA worktree、更新状态、调度或合入主树。

下一阶段仍须保留 freeze23/旧 QA/原 FAIL，不覆盖旧标签；迁移后生成新的 freeze24 与 fresh QA，并由独立 QA 使用原完整命令、原 verifier 和额外回归断言验收。任何额外生产改动、放宽其他旧断言、缺失原测试或证据漂移都不在本次 PASS 内。
