# 当前任务快照

最后更新：2026-08-17 16:54 +08:00

## 背景

- 用户要求持续比较本地与上游历史，只选择性移植可独立验证的修复，禁止整包合并长期分叉历史。
- S223-S225 已完成；本轮为国产供应商一等支持建立独立 S226 contract，并先按可独立编译、验收的提交批次整理范围。
- 所有新提交仅在本地 `main`；未授权 push、部署、容器、共享/生产数据库或真实 provider 操作。

## 当前目标

- S226 contract 已批准，等待用户明确授权从 S226-A 开始实现。当前不创建 worktree、不调度 worker、不改业务代码。

## 本次已完成

- S223 已本地合入：业务 `7af27c591`、独立 QA `3a3aeb601`，并完成 workflow 收口。
- S224 已本地合入：业务 `69be22fae`、Developer 报告 `7242b824a`、独立 QA `ac3244191`，workflow 收口 `06e0e6ea5`。
- S225 已本地合入：业务 `ba42a434e`、Developer 报告 `b82c9c998`、独立 QA `51b9a47bd`。
- S223、S224 Developer/QA、S225 Developer/QA 共五个 worktree 和五个 `pge/*` 分支已清理；无关 detached `tutorial-nav-20260817` 保留。
- 新增 `docs/workflow/tasks/upstream-cn-providers-s226.md`，将实现拆为 A 平台/账号基础、B 额度余额探测与管理 API、C 多协议网关与冷却、D 前端账户管理、E 集成与独立 QA。
- S226 contract review 已 PASS；workflow 状态停在 `contract-approved`，没有创建 S226 worktree 或调用 Developer/QA。

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
- 用户未提交内容包括两个 account-modal 文件、`TutorialView.vue` 及其测试、`knowledge/00-start-here.md`、`knowledge/05-current-focus.md` 和 `outputs/`。account patch-id 为 `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`，tutorial patch-id 为 `7f5afe57708ae3cc6b5781989c25195eaa6ffda5`，knowledge patch-id 为 `2abee47db90ce1d54e1f9ba7d1a3cc2d633c2374`。

## 待验证点

- S226-A 实现前需由 Controller 记录精确 base `98daf5b8d` 并建立单独 worktree；验证：只允许 A 的 7 个业务/测试路径和报告，focused x10、完整 service、server compile 均 PASS。
- S226-D 仍需验证用户 modal 临时 baseline/最终合入策略；验证：业务 commit 不包含用户 patch，主工作区 account patch-id 在集成前后均为 `5d316e5b...`。
- 前端当前 `node_modules` 缺少 `vitest/vue-tsc/vite` 可执行文件；D/E 开始时必须在任务 worktree 恢复工具链，否则 QA 报 `BLOCKED`，不得跳过。
- 若授权发布：先复核最终 `git status`、主线测试证据和远端差异，再执行普通 `git push origin main`；当前没有发布授权。
- S225 未运行真实 Redis 或上游 provider 集成；合同禁止这些操作，允许范围由 mock cache、完整 service 和 server 编译覆盖。

## 当前结论

- `PASS / S226-contract-approved-only`。
- A-E 分批、上游阻断项处置、本地 owner、allowlist、验收和 stop rules 已冻结；S226 业务尚未实现或运行态验证。

## 下一步

- 用户授权 S226-A -> 验证：从 `98daf5b8d` 创建唯一 A worktree，按 contract 调用 Terra Developer；A Controller PASS 前不开始 B。
- 后续依次 S226-B -> C -> D -> E；每批验证：前一批精确 commit、allowlist、focused/full gates 和报告均 PASS，最终 E 才允许主线集成。
- 发布当前本地提交（需用户授权） -> 验证：push 前后比较 `HEAD`、`origin/main` 和远端 `refs/heads/main`，只允许普通 push。

## 验证记录

- S224 QA：`docs/workflow/qa-reports/upstream-billing-quantize-s224-qa.md`，首行为 `### PASS`。
- S225 QA：`docs/workflow/qa-reports/upstream-fingerprint-user-agent-validation-s225-qa.md`，首行为 `### PASS`。
- S225 Controller：focused x10 `0.092s`、service `60.469s`、server compile PASS；独立 QA：focused x10 `0.077s`、service `60.243s`、server compile PASS。
- S225 集成主线：focused x10 `0.077s`、patch-id/format/provenance/conflict/index 与两组用户 patch-id PASS。
- S226 contract：目标链 ancestry 在 `upstream/main@e330c243a` 可达；直接 apply 检查失败并确认需手工适配；quota 前置 123 文件、threshold 前置 55 文件均已量化并排除。
- S226 保护状态：`main@98daf5b8d`，`origin/main@a865d8b6e`，本地领先 20；account patch-id `5d316e5b...`、tutorial patch-id `7f5afe57...`、knowledge patch-id `2abee47d...`，暂存区为空，`outputs/` 保持未跟踪。
