---
phase: done
current_sprint: studio-bridge-luoye-production-handoff
total_sprints: 13
pending_action: production-config-and-real-account-verification
project_type: web
qa_mode: runtime
approval_required: false
last_verified: 2026-06-09
---

# Workflow Status

- 当前阶段：`done`
- 当前 Sprint：`studio-bridge-luoye-production-handoff`
- 当前目标：把 Sub2API 从“上游合成 / gateway hardening 默认入口”推进到 `Studio Bridge / 落叶AI` 生产联调默认入口，避免续做时仍误判为 prompt cache 或模型市场跟进。
- 当前结论：提交 `fe2f80be1 feat: add studio bridge integration` 已把 Sub2API 扩展为落叶AI的用户、余额、充值、默认分组和扣费真源；当前默认后续动作不再是继续补 gateway merge，而是补生产配置与真实账号闭环验证。
- 当前已稳定进入默认主线的事实：
  - `/studio-bridge/launch` 已作为 `/chat-images` 的稳定 alias，用户侧入口默认应进入落叶AI启动链路，而不是 OpenWebUI。
  - Sub2API 当前负责维护 bridge internal secret、落叶AI launch URL、充值回跳 URL、默认聊天/生图/视频分组和后台配置。
  - 余额摘要、充值摘要、使用记录摘要，以及 `reserve / commit / refund` 幂等扣费接口，已经成为当前 bridge 联调的核心边界。
  - 团队空间场景下，落叶AI 负责记录 `actor/payer`，Sub2API 仍是扣费和余额唯一真源。
  - 之前的 gateway auth / sticky session / prompt cache 修复已降为稳定背景层；后续若继续上游合成，必须单独开 Sprint，避免覆盖 Studio Bridge 和落叶AI入口。
- 目标验证入口：
  - `knowledge/tasks/current-task.md`
  - `docs/workflow/main-log.md`
  - `frontend/src/__tests__/public-smoke.spec.ts`
  - `backend/internal/service/studio_bridge.go`
  - `backend/internal/handler/studio_bridge_handler.go`
- 下一合法动作：先补生产域名、bridge secret、充值回跳 URL 和默认分组配置，再用真实账号验证注册/登录回跳、充值、创作扣费、使用记录和团队空间最小闭环；如继续做上游合成，需新开独立 Sprint。
- 状态推进规则：先 `spec-approved`，再进入当前 Sprint 的 `contract-draft -> contract-approved -> build -> qa -> fix -> retest -> done`。
