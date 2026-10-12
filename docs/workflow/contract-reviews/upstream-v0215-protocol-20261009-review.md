---
status: approved
verdict: PASS
task_id: upstream-v0215-protocol-20261009
base_commit: ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b
reviewer: current-session-default
last_verified: 2026-10-10
---
### PASS: upstream-v0215-protocol-20261009

## Findings
- 合同范围、Success Criteria 与本地结构兼容，未发现必须修订的合同问题。PASS 仅指合同审查，不指实现或验收通过。
- 已核对七个上游 commit 的路径；本地转换器、类型、Antigravity 序列化和四个 buffered owner 均存在，覆盖 developer/tool_choice/legacy pairing/refusal/signature/negative index/ping 的必要改动入口。
- 与 frontend、ws-oauth 合同的实际业务 owner 和约定测试前缀无重叠；不得为混合职责测试命名而同时匹配其他合同的测试通配符。
- copied dirty baseline 的 `chatcompletions_to_responses.go` 已有三处 `Type: "message"` 改动，必须保留，不能用原 upstream 文件替换。

## Executed Checks
- 在唯一工作目录 `E:/codex-worktrees/sub2api/upstream-v0215-20261009` 读取 spec、三合同、isolated status、matrix，执行 `git rev-parse HEAD`、`git status --short`、针对性 `git diff`、`git show --stat` 和必要上游 diff。
- HEAD 与合同 base_commit 一致；9 个现存精确业务路径的 SHA256 均等于外置 `baseline.json`。
- 静态检查确认四个 buffered 路径目前均有未排除负值的 `idx < len(...)`；refusal 和 ping 入口在 allowlist 内。
- baseline 包测试及生产构建 PASS 是主控提供的证据；本 reviewer 未重跑测试、构建或完整审计。

## Gates / Unverified Risks
- 主控须在本报告 PASS 后将合同标为 `status: approved`、`review_verdict: PASS` 才能 dispatch；当前 draft/pending 是正常前置状态。
- 用户已明确授权本轮 current-session-default native agents；以 spec 中本轮例外为准，不改全局矩阵、不因历史 matrix 模型阻塞。
- 新测试名称必须包含 `V0215Protocol`，报告实际执行数量；无匹配测试不能算通过。逐项覆盖 Success Criteria，既有现代 tool call、普通文本和非 thinking block 不退化。
- service 的已知 fixture 编译漂移单独记录。允许外置、精确、compile-only overlay：保留原错误、完整替换内容及命令，仅修已证实的既有签名/fixture 漂移；不能屏蔽本轮测试、替换生产实现或删除整批测试绕过失败。
- 后续生产构建建议使用 `go build -o E:/codex-runtime/upstream-v0215-20261009/protocol-server.exe ./cmd/server`，避免在 allowlist 外生成 `server.exe`；构建、日志、overlay 均外置。
- 验收差异相对 copied baseline 而非仅相对 HEAD；检查自身最小增量、copied hunks、其他 worker owner 和非 allowlist 文件未被覆盖。真实 provider、部署和生产运行未验证且不在授权范围。

## Recommendation
合同可批准并进入实现；独立 QA、production build、完整逐项测试、scope/baseline preservation 仍为后续门禁。本轮只新增本报告，不改合同、源代码、依赖、知识库，不提交、推送或部署。

## Revision Review — 2026-10-10
### PASS: upstream-v0215-protocol-20261009 — narrow TestStreamingReasoning exception
- 已只读核对本次合同扩展：allowlist 新增 `backend/internal/pkg/apicompat/anthropic_responses_test.go`，Narrow relevant-test exception 明确仅改 `TestStreamingReasoning` 的 content_block SSE expected，要求新增 wire contract 的空 `signature`；禁止其他测试变动、删除测试或覆盖 copied unrelated hunks。未扩大产品范围或其他 worker owner。
- 静态证据：该测试第 776 行仍要求 `"content_block":{"thinking":"","type":"thinking"}`；当前 thinking MarshalJSON 明确序列化 `signature`，含空字符串。因此主控报告的本轮失败是相关 wire 契约变化引起的旧 expected 不兼容，不得归类为既有 baseline 失败或用 compile overlay 屏蔽。
- 此文件当前 SHA256 与 immutable copied baseline 一致；相对 HEAD 的既有改动仅在 `TestAnthropicToResponses_ThinkingDisabled`。其 OutputConfig、effort `"none"` 和 Summary 断言等 copied hunks 必须原样保留。后续相对 copied baseline 的增量只允许上述一个 SSE expectation。
- **全包必须重新跑且通过：** 从 backend 执行 `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity -count=1`，不得只跑新 tests、只跑 TestStreamingReasoning 或仅以主控所报 32 cases PASS 替代全包。记录实际测试数、退出码和日志；独立 QA 重新核对新 expected 确实要求空 signature，不能弱化为只检查 thinking/type。
- service focused tests（按已批准外置精确 overlay）、生产 build、scope/diff 检查和独立 QA 仍保留。此 revision PASS 是合同扩展批准，不是测试修复或全包验收 PASS。
- 本次仅追加本 review；未修改 test、合同、其他报告或业务源码，未重跑测试。
