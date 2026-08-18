# 当前任务快照

最后更新：2026-08-18 19:20 +08:00

## 背景

- 用户要求持续比较本地与上游历史，只选择性移植可独立验证的修复，禁止整包合并长期分叉历史。
- S223-S225 已完成；本轮为国产供应商一等支持建立独立 S226 contract，并先按可独立编译、验收的提交批次整理范围。
- 所有新提交仅在本地 `main`；未授权 push、部署、容器、共享/生产数据库或真实 provider 操作。

## 当前目标

- S226、S228、S229-A、S229-B 与 S229-C 已按批准边界集成 `main`；断开排水/partial-result usage 仍保持独立评估，未执行 push、部署或真实 provider 操作。

## 本次已完成

- S223 已本地合入：业务 `7af27c591`、独立 QA `3a3aeb601`，并完成 workflow 收口。
- S224 已本地合入：业务 `69be22fae`、Developer 报告 `7242b824a`、独立 QA `ac3244191`，workflow 收口 `06e0e6ea5`。
- S225 已本地合入：业务 `ba42a434e`、Developer 报告 `b82c9c998`、独立 QA `51b9a47bd`。
- S223、S224 Developer/QA、S225 Developer/QA 共五个 worktree 和五个 `pge/*` 分支已清理；无关 detached `tutorial-nav-20260817` 保留。
- 新增 `docs/workflow/tasks/upstream-cn-providers-s226.md`，将实现拆为 A 平台/账号基础、B 额度余额探测与管理 API、C 多协议网关与冷却、D 前端账户管理、E 集成与独立 QA。
- S226 contract review 已 PASS；批准时停在 `contract-approved`，当时尚未创建 S226 worktree 或调用 Developer/QA。
- S226-A worktree `E:/codex-worktrees/sub2api/upstream-cn-providers-s226-a` 从冻结 base `98daf5b8d` 创建；Developer 实现 `ba7c00c78`、报告 `3ed89c995`，未进入 B 或 QA。
- S226-A Controller review PASS：7 个业务/测试文件加报告严格 allowlist，8/8 focused 可发现且 x10 PASS，完整 service、server compile、格式、Git/provenance 与主工作区保护门禁均 PASS。
- S226-B worktree `E:/codex-worktrees/sub2api/upstream-cn-providers-s226-b` 已从 A 报告 commit `3ed89c995` 创建，分支为 `pge/upstream-cn-providers-s226-b`；仅授权 20 个 B 业务/测试路径和结果报告。
- S226-B Controller review PASS：初版候选在读取异常和无效 payload 时会覆盖旧快照并可误触发余额暂停，已退回原 Developer 修复。最终业务 `316fa46c6`、报告 `f6b380e21` 保持两提交边界，18 个业务/测试文件及报告均严格 allowlist；17/17 合同测试可发现，20 个 focused/鲁棒性测试 x10、完整 config/service/routes/cmd-server、B4 零出站、Wire、格式、diff、provenance、冲突/index 和主工作区保护门禁均 PASS。
- S226-C 实际调用 `gpt-5.6-terra` Worker CLI 时在推理前返回 API `404`；零 token、无业务文件或报告生成，C 工作树仍为精确基线 `f6b380e21` 且洁净。
- 用户明确允许替代模型；可用性探测确认 `sonnet` 解析为 `claude-sonnet-4-6` 并成功返回，已在 S226 contract 中指定为 Developer 和独立 QA 的具名替代模型。
- 两次具名 Worker 尝试均未形成有效报告或业务提交：首次无报告退出，第二次返回 `Content block not found`；只读探针显示 CLI 把 C 路径映射到其他环境。低成本 Worker 循环已停止，Controller 接管实现，范围和验收门禁不变。
- S226-C Controller review PASS：业务 `24873abf1`、报告 `5bb985cb6` 保持两提交边界，且仅含 C allowlist 和报告；16/16 合同测试及新增凭证/WebSocket 回归共 17 项均可发现并 x10 PASS，完整 `service`/`handler`/`routes`、`cmd/server` 编译、gofmt、diff、allowlist、冲突/index、三项 provenance 与 C0 主工作区保护门禁全部 PASS。业务 patch-id 为 `d6ee6e8e161ad9343b86f8092e55a4be9e2fbe88`。
- S226-D Controller review PASS：业务 `a559956f7`、报告 `c539d1f01`，D 工作树以用户 modal baseline `d7158e916` 隔离；7 个 focused 文件共 87 项、typecheck/build、allowlist、provenance 和保护门禁通过，业务 patch-id 为 `04fc586c994a0264280db52a88c6398d83e29ebe`。
- S226-E 独立 QA PASS：QA report `5ca12b78b`；A-C focused 40 项均可发现并 x10 PASS，完整 backend service/handler/routes、server compile、前端 focused/typecheck/build、scope/provenance/conflict/index 和保护 patch/hash 通过。浏览器 session `s226-e-qa-20260818-final` 检查公共首页非空并完成清理；无登录态导致后台账号页真实操作未覆盖，已记录为残余风险。
- S226 已本地集成：A-C 业务/报告、无 baseline 的 D 业务提交 `501c3830a`、D 报告和 QA 报告依序进入 `main`，最终 HEAD `6ca47c2f8`，相对 `origin/main` ahead 45；D 集成 patch-id 保持 `04fc586c...`。
- S228 已完成独立实现、Controller review 和独立 QA：业务 `df43f3876`/`26a5dec9d`，Controller report `b0a7a6e8b`，QA report `9e4beddc2`；按顺序集成到 `main` 为 `22b04fa0d`/`cc1630bd7`/`2cbe98f0b`/`ff241be81`。
- S229-A 已完成独立实现、Controller review 和独立 QA：业务 `ce0ffdb65`，Controller report `fb391fd08`，QA report `fe11096aa`；按顺序集成到 `main` 为 `2422b9b15`/`65ac54145`/`de62dd8d6`。
- S229-B billing-only contract 已批准：基线 `main@de62dd8d6`，上游 source `10c8b7020`，范围限定 CN 计费候选过滤、显式定价放行与空候选 zero-cost usage；403 和断开排水切片继续分离。
- S229-B 已完成隔离实现、Controller review、独立 QA 和主线集成：业务 `c3b0ed259`，Controller report `upstream-cn-provider-billing-s229-b-result.md`，QA report `upstream-cn-provider-billing-s229-b-qa.md`，主线仍未 push。
- S229-C 403-only contract 已批准：基线 `main@44fa47124`，上游 source `10c8b7020`，范围限定 CN 403 分派复用 OpenAI 策略；断开排水切片继续分离。
- S229-C 已完成隔离实现、Controller review、独立 QA 和主线集成：业务 `7911a0ef2`（候选 `2d60e8cd0`）、Controller report `upstream-cn-provider-403-s229-c-result.md`、QA report `upstream-cn-provider-403-s229-c-qa.md`（提交 `9e5050aac`）。主线 focused CN 403 x10、完整 service、server compile、scope/provenance/conflict/index 和保护检查均通过；未 push。

## 已确认事实

- S224 在生成/保留原始请求指纹后，用 decimal 八位量化六个金额字段，包含本地 `PrepaidBalanceCost`。Developer、Controller、独立 QA 与集成主线 focused 均 PASS。
- S225 保留本地 `claude-cli/2.1.92`、Stainless `0.70.0/v24.13.0` 等默认值；创建和升级共用 UA 校验，污染缓存两种自愈均保留 `ClientID`。独立 QA 未发现实现缺陷。
- S225 集成主线 11/11 focused 测试 x10 PASS（0.077s）；候选与主线业务 patch-id 均为 `3c649274094273e6c75c14859669eed1b6c8e753`。
- `origin/main` 仍为 `a865d8b6e`，本轮没有 push。`upstream/main` 已 fetch 到 `e330c243a`。
- 上游目标链为 `901a0439f -> 4b667ccd4 -> e72854538`，最终约 78 文件、6195 新增/125 删除；直接 apply 在 Wire、config、gateway 和缺失 schema 处失败，必须按本地拓扑手工移植。
- `4b667ccd4` 的 B1 根 `docker-compose.yml` 排除；B2 `user_platform_quotas` 迁移在本地 N/A，不改号为 226。该产品前置 `6b39b344d` 本身为 123 文件/14220 行，并有后续 flusher，不属于国产供应商探测或网关的必要前置。
- 上游可配置调度阈值又依赖本地缺失的 `7c62382d0`（55 文件/3542 行）。S226 保留额度快照和响应式 429 重置点冷却，但不暗中引入通用阈值产品或前端设置面板。
- B3 四个 Anthropic-native 读循环的 interval timeout 属于 S226-C；B4 探测 URL allowlist 与拒绝时零出站属于 S226-B。
- 上游多个拆分 gateway 文件本地不存在；contract 已将其改写到 `gateway_service.go`、`openai_gateway_service.go`、`openai_gateway_chat_completions_raw.go` 和 `openai_ws_forwarder.go` 等本地 owner。
- S226-A 保留 `IsOpenAICompatible` 的 openai/grok 语义，未提前开放 CN 路由；也未扩展 `AllowedQuotaPlatforms` 或 `AllowedSchedulingThresholdPlatforms`。业务 patch-id 为 `b0ec5bd95a5e00fffd8e06000f2f96dfbe552680`。
- 用户未提交内容包括两个 backend 教程测试、两个 account-modal 文件、`TutorialView.vue` 及其测试、`knowledge/00-start-here.md`、`knowledge/05-current-focus.md`、六个未跟踪教程 migration/test 文件和 `outputs/`。C0 patch-id 为 backend `a81fbffb...`、account `5d316e5b...`、tutorial `a07a7c33...`、knowledge `2abee47d...`；六个文件 SHA256 已记录到 contract，均必须保持原样。
- S228 集成后 `main@ff241be81` 相对 `origin/main@a865d8b6e` ahead 55；`upstream/main@8869775ed` 未变化。S228 业务 patch-id 为 `8b0caf6e...` 和 `cae02fc1...`，与候选实现一致；精确 allowlist、无冲突索引、三项上游 ancestry 均通过。
- S229-A 集成后 `main@de62dd8d6` 相对 `origin/main@a865d8b6e` ahead 58；业务 patch-id 为 `ad03cda9...`，与候选 `ce0ffdb65` 一致。三个 focused 测试 x10、完整 handler/service、server compile、scope/provenance/conflict/index 和保护门禁均 PASS。

## 待验证点

- `10c8b7020` 剩余 disconnect drain/partial-result usage slices -> 验证：继续读取本地 `openai_gateway_service.go`、stream owners 和 usage logging 路径，拆出可独立测试的 contract；不要与已完成 S229-C 混合。
- 若授权发布：先复核最终 `git status`、主线测试证据和远端差异，再执行普通 `git push origin main`；当前没有发布授权。
- S225/S226/S228 均未运行真实 Redis 或上游 provider 集成；合同禁止这些操作，当前证据来自 mock/httptest、包回归、server 编译和前端构建。

## 当前结论

- `PASS / S226-E independent QA`：独立 QA 报告 `5ca12b78b`，静态、运行态、构建、scope/provenance 和保护边界均通过；UI 登录态限制已显式记录。
- `PASS / S226 main integration`：A-D 业务与证据提交已按顺序集成 `main@6ca47c2f8`，用户 dirty patch IDs、未跟踪教程文件和 `outputs/` 均保持原值。
- `PASS / S228 independent QA`：QA 报告 `docs/workflow/qa-reports/upstream-cn-group-entry-s228-qa.md` 首行为 `### PASS`，后端 binding x10、前端 focused 7 项、admin 117 项、typecheck、scope/provenance/conflict/index 均通过。
- `PASS / S228 main integration`：业务与证据提交已按序集成 `main@ff241be81`，用户 dirty patch IDs、六个未跟踪教程文件及 `outputs/` 均保持原值；未 push。
- `PASS / S229-A independent QA`：QA 报告 `docs/workflow/qa-reports/upstream-cn-provider-correctness-s229-a-qa.md` 首行为 `### PASS`，gate/dispatch/count_tokens focused x10、完整 handler/service、server compile、scope/provenance/conflict/index 均通过。
- `PASS / S229-A main integration`：业务与证据提交已按序集成 `main@de62dd8d6`，用户 dirty patch IDs、六个未跟踪教程文件及 `outputs/` 均保持原值；未 push。

## 下一步

- 评估 `10c8b7020` 剩余 slices -> 验证：先读取本地 `openai_gateway_service.go`、`ratelimit_service.go` 和 handler/stream owners，确认每片的可独立测试边界。
- 保留当前本地提交和用户 dirty 内容，等待明确发布授权。
- 发布当前本地提交（需用户授权） -> 验证：push 前后比较 `HEAD`、`origin/main` 和远端 `refs/heads/main`，只允许普通 push。

## 验证记录

- S224 QA：`docs/workflow/qa-reports/upstream-billing-quantize-s224-qa.md`，首行为 `### PASS`。
- S225 QA：`docs/workflow/qa-reports/upstream-fingerprint-user-agent-validation-s225-qa.md`，首行为 `### PASS`。
- S225 Controller：focused x10 `0.092s`、service `60.469s`、server compile PASS；独立 QA：focused x10 `0.077s`、service `60.243s`、server compile PASS。
- S225 集成主线：focused x10 `0.077s`、patch-id/format/provenance/conflict/index 与两组用户 patch-id PASS。
- S226 contract：目标链 ancestry 在 `upstream/main@e330c243a` 可达；直接 apply 检查失败并确认需手工适配；quota 前置 123 文件、threshold 前置 55 文件均已量化并排除。
- S226-A Controller：focused 8/8 可发现，x10 `0.079s`；service `60.255s`；server compile `0.071s`；gofmt、diff、allowlist、冲突/index、三项 upstream provenance 与 batch boundary PASS。
- S226-A 保护状态：`main@6ebabe92b`，`origin/main@a865d8b6e`，本地领先 22；backend/account/tutorial/knowledge patch-id 与两个 migration SHA256 均和 dispatch 快照一致，暂存区为空，`outputs/` 保持未跟踪。
- S226-B Controller：先修复读取失败/无效响应不能覆盖旧快照或暂停账号的缺陷；17/17 可发现，20 项 focused/鲁棒性 x10 `0.090s`，config `0.710s`、service `60.407s`、routes `1.491s`、cmd/server `0.086s`，B4 零出站、owned-pause、多币种、proxy、Wire、gofmt、diff、allowlist、冲突/index、三项 provenance 和主工作区保护均 PASS。业务 patch-id `a8c91f5789b96a93ffb6c8d99969519726906e03`；主工作区为 `main@73cf6aa21`，`origin/main@a865d8b6e`，领先 24，用户 patch/hash 和 `outputs/` 状态不变。
- B PASS 后的 live protection check：TutorialView patch 已由用户更新为 `ce6749a8c5d0256cfa1a986f3e4d8d7377df6753`；B worktree、报告和 Controller workflow 提交均未包含该路径。保留该用户变化；不要用 B gate 的旧 `9e0894bc...` 值继续派发 C。
- C0 amendment：当前 TutorialView patch 再次变为 `a07a7c33f09d9fa0e308a1bddf6bf0ee9d7cf671`，并新增四个教程 migration/test 未跟踪文件；C+ 使用这组当前 patch 和六个教程文件 SHA256 作为保护基线，历史 A/B 值只保留审计用途。
- S226-C worktree `E:/codex-worktrees/sub2api/upstream-cn-providers-s226-c` 已从 B report `f6b380e21` 创建，分支为 `pge/upstream-cn-providers-s226-c`，工作树洁净；Developer 仅可改 C allowlist 和结果报告。
- S226-C Worker retry 证据：`claude.cmd --bare -p` 携带配置模型 `gpt-5.6-terra` 返回 `api_error_status: 404`、`input_tokens: 0`、`total_cost_usd: 0`；未生成 worker report，未改动 C 工作树或主工作区保护文件。
- 具名替代模型探测：用户授权后，以 `--model sonnet` 运行的最小无写入请求成功，CLI 报告实际模型为 `claude-sonnet-4-6`；此探测未读取或修改业务文件。
- Worker 升级证据：Sonnet 重试没有业务 diff；一次无报告退出，另一次返回 `Content block not found`，只读探针在 `$0.05` 上限内反复解析错误路径后停止。Controller 接管，不再重复低成本 Worker 调度。
- S226-C Controller：17 项 focused 回归（16 合同项加凭证/WebSocket）均可发现并 `-count=10` PASS；完整 `go test ./internal/service ./internal/handler ./internal/server/routes -count=1` 与 `go test ./cmd/server -run '^$' -count=1` PASS。gofmt、diff、allowlist、冲突/index、三项 provenance 和 C0 保护均 PASS；业务/报告提交为 `24873abf1` / `5bb985cb6`。
- S226-D dispatch：工作树从 `5bb985cb6` 创建，用户 account modal patch 以 `d7158e916` 作为不合入 baseline；主工作区 patch-id 仍为 `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`，D 仅允许 baseline 之后的前端 allowlist 差异。
- S226-D Developer attempt 1：Sonnet 因 contract 路径解析到 D 工作树外并达到 `$0.10` 预算而退出，实际 `$0.1079`；无业务 diff、报告或依赖清单变化，D 仍停在 `d7158e916`，允许一次绝对路径受控重试。
- S226-D Developer attempt 2：绝对路径重试立即返回 `Content block not found`，零 token、零文件和报告变化；按连续失败规则停止 Worker loop，Controller 接管 D 实现，范围、baseline 与 E 独立 QA 门禁不变。
- S226-D Controller review PASS：业务 `a559956f7`、报告 `c539d1f01` 保持两提交边界；D 工作树从 `5bb985cb6` 加用户 modal 临时 baseline `d7158e916` 开始，业务 diff 仅含 D allowlist。7 个 focused Vitest 文件共 87 项、typecheck、build、diff、allowlist、冲突/index、三项 provenance 和主工作区保护门禁均 PASS；业务 patch-id 为 `04fc586c994a0264280db52a88c6398d83e29ebe`。独立 QA 尚未开始，D 未集成 main。
- S226-E QA：`5ca12b78b`；QA worktree 在 `c539d1f01` 上无产品 diff，A-C 40 项 focused x10、完整 backend service/handler/routes、server compile、D 87 项 focused、typecheck/build、provenance、scope、冲突/index 和保护 patch/hash 均 PASS；浏览器 session `s226-e-qa-20260818-final` 首页检查非空，后台账号页因无登录态未操作，session/profile/daemon/server 均已清理。
- S226 主线集成：`1219d5352`/`cb34acf28`、`0a6990ea4`/`f32deb2ee`、`974793cd4`/`b93324820`、无 baseline 的 D `501c3830a`、`b97272adc`、`6ca47c2f8` 按序进入 `main`；A/B/C/D patch-id 分别为 `b0ec5bd95...`、`a8c91f578...`、`d6ee6e8e1...`、`04fc586c...`，最终主线前端/后端 fresh verification 全部通过。
- S228 主线集成：后端 `go test -tags=unit ./internal/handler/admin -run "TestGroupPlatformBinding" -count=10`、前端 3 个 focused 文件共 7 项和 `pnpm run typecheck` 均 PASS；业务 patch-id `8b0caf6e...`/`cae02fc1...` 与候选一致，主线 ahead 55，远端 refs 未变。
- S229-A 主线集成：三个 focused Go 回归均 `-count=10` PASS，完整 `handler`/`service` 已由独立 QA PASS，`cmd/server` compile PASS；业务 patch-id `ad03cda9...`，主线 ahead 58，远端 refs 未变。
- S229-B 主线集成：focused billing 三项测试 x10、完整 `internal/service`、`cmd/server` compile、scope/provenance/conflict/index 和保护 patch/hash 均由独立 QA PASS；业务 patch-id `10ef0f42...`，主线 `main@f4e7f45d8` ahead `origin/main` 62，未 push。
- S229-C 主线集成：focused `TestHandleUpstreamError_CNProviderHTML403SkipsAccountPenalty|TestHandleUpstreamError_CNProviderStructured403TempUnschedulable|TestHandleUpstreamError_CNProviderStructured403ThresholdDisables` x10、完整 `internal/service`、`cmd/server` compile、gofmt、diff、scope/provenance/conflict/index 与保护 patch/hash PASS；业务 patch-id `914820cfc3804f2e40a8f58f64ad6f266926b2a2`，主线 `main@9e5050aac` ahead `origin/main` 64，未 push。
