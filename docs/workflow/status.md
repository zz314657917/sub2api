---
phase: done
current_sprint: smart-routing-composite-optimal-v1
total_sprints: 1
pending_action: none
project_type: web
qa_mode: runtime
approval_required: false
last_verified: 2026-06-13
---

# Workflow Status

- 当前阶段：`done`
- 当前 Sprint：`smart-routing-composite-optimal-v1`
- 当前目标：把当前“多分组路由预设 + OpenAI 实验账号调度”推进为 v1 “综合最优”策略，按价格、成功率、速度和实时拥堵综合评分，自动选择当前更合适的分组与账号。
- 当前结论：Sprint 已实现并通过目标 QA。v1 已新增账号级价格评分和前端“综合最优”预设；分组级动态综合评分留作后续增强。
- 当前已确认事实：
  - 用户侧 API Key 创建/编辑页已有“价格优先 / 速度优先 / 成功率优先 / 手动配置”等多分组路由预设。
  - OpenAI 实验调度器默认关闭，由 `openai_advanced_scheduler_enabled` 控制。
  - 实验调度器已有 `priority / load / queue / error_rate / ttft` 评分，本 Sprint 已补入账号级 `price` 权重。
  - `price` 因子当前读取 `account.BillingRateMultiplier()`；未知倍率按既有账号方法回退。
  - 前端新增 `optimal` / “综合最优”预设，默认创建 API Key 时生成同优先级候选池。
  - 关闭实验调度时，应继续走旧的粘性会话、优先级、负载和 LRU 路径。
- 目标验证入口：
  - `docs/workflow/spec.md`
  - `docs/workflow/tasks/smart-routing-composite-optimal-v1.md`
  - `docs/workflow/worker-results/smart-routing-composite-optimal-v1-result.md`
  - `docs/workflow/qa-reports/smart-routing-composite-optimal-v1-qa.md`
  - `backend/internal/service/openai_account_scheduler.go`
  - `frontend/src/views/user/KeysView.vue`
- 已执行验证：
  - `go test ./internal/service -run "TestOpenAI.*Scheduler|TestAPIKey|Test.*Routing" -count=1` 通过。
  - `go test ./internal/config -count=1` 通过。
  - `npm.cmd run test:run -- src/views/user/__tests__/KeysView.createQuery.spec.ts` 通过，12 tests；仅有既有 Browserslist 数据过期提示。
  - `git diff --check` 通过；仅有 Markdown LF/CRLF 工作区提示。
- 下一合法动作：如需继续增强，另开 Sprint 做分组级动态综合评分、运行态可解释 score breakdown 或管理后台权重 UI。
- 状态推进规则：`contract-draft -> contract-approved -> build -> qa -> fix -> retest -> done`。
