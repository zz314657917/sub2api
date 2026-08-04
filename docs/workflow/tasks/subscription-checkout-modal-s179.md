# Subscription Checkout Modal S179

## Task ID

`subscription-checkout-modal-s179`

## Role

Planner、Generator 与 Evaluator 均由当前 Codex 会话按独立小修流程执行；不调用 worker，不覆盖当前 S176 状态事实源。

## Goal

把用户定价页中位于套餐卡片列表下方的订阅确认区改为居中付款弹窗，同时保持现有套餐金额、支付方式和创建订单逻辑不变。

## Success Criteria

- 点击任一套餐的“立即订阅/立即续费”后，确认付款内容通过 `BaseDialog` 居中展示，不再插入套餐列表下方。
- 弹窗支持关闭按钮、取消按钮、遮罩点击和 `Escape` 关闭；提交期间不得误关闭。
- 创建订单成功并进入 `paying` 阶段后，订阅确认弹窗不再遮挡 `PaymentStatusPanel`；创建失败时仍可在弹窗中重试。
- 桌面宽度下使用双栏信息布局，移动端变为单栏且无横向溢出。
- 订阅金额继续直接使用 `selectedPlan.price`，不引入余额充值倍率或改变支付 API payload。

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`
- Related files: `frontend/src/views/user/PaymentView.vue`, `frontend/src/components/common/BaseDialog.vue`, `frontend/src/views/user/__tests__/PaymentView.spec.ts`
- Current workflow note: `docs/workflow/status.md` remains on S176 browser QA and is not reassigned by this isolated UI fix.

## Allowed Paths

- `frontend/src/views/user/PaymentView.vue`
- `frontend/src/views/user/__tests__/PaymentView.spec.ts`
- `docs/workflow/tasks/subscription-checkout-modal-s179.md`
- `docs/workflow/qa-reports/subscription-checkout-modal-s179-qa.md`
- append-only S179 entries in `docs/workflow/main-log.md`
- `frontend/output/playwright/subscription-checkout-modal-s179/**`

## Denied Paths

- `backend/**`
- `frontend/src/components/common/BaseDialog.vue`
- `frontend/src/style.css`
- `knowledge/**`
- `docs/workflow/status.md`
- existing S176/S177 task and QA artifacts
- `C:/Users/Administrator/.codex/memories/**`
- database migrations, payment provider configuration, deployment and production resources

## Constraints

- 保持最小改动，不抽取新组件，不修改订阅订单请求和价格计算。
- 复用现有 `BaseDialog` 的 Teleport、焦点恢复、滚动锁定与键盘关闭能力。
- 保留工作树内现有用户改动，不做无关格式化、提交、推送或部署。
- 用户可见文本继续复用现有 i18n/fallback 文案。

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/frontend
npm.cmd run test:run -- src/views/user/__tests__/PaymentView.spec.ts
npm.cmd run typecheck
npm.cmd run build
cd F:/mcplugins/sub2api
git diff --check
```

人工浏览器验收：在桌面和 `390x844` 移动端打开用户支付页，点击套餐后检查弹窗居中、关闭路径、布局与横向溢出。

## Output

- QA 报告首行必须为 `### PASS: subscription-checkout-modal-s179`、`### FAIL: subscription-checkout-modal-s179` 或 `### BLOCKED: subscription-checkout-modal-s179`。
- 报告列出 changed files、executed checks、browser evidence、risks 与 contract compliance。
- 不写长期知识库；只有产生稳定且可复用的新结论时才另行评估 knowledge candidate。

## Stop Rules

- 若必须修改后端、支付 payload、全局模态框组件/样式或数据库，停止并重新规划。
- 若现有支付页文件出现未识别的并发改动，停止编辑并先重新检查 diff。
- 浏览器无法访问本地页面时，自动化验证可判通过，但最终结论必须标明浏览器证据为 `BLOCKED`，不得伪称完整视觉 PASS。

## Evaluator Contract Review

`PASS / contract-approved`：现有 `BaseDialog` 已提供所需模态交互边界，`PaymentView` 的 `selectedPlan` 和 `paymentPhase` 足以控制打开与支付后退出；验收命令和写入边界均可执行，不需要后端或共享组件变更。

## Evaluator Acceptance Amendment

- `npm.cmd run typecheck` 与 `npm.cmd run build` 在本轮实现前后的当前依赖状态中均被
  `AirwallexPaymentView.vue:103` 无法解析 `@airwallex/components-sdk` 阻断；该模块的 pnpm 缓存目录缺少包本体，且不在本任务 Allowed Paths。
- 不修改锁文件或联网修复依赖。本任务的范围内构建证据改为：聚焦 Vitest、两份改动文件 ESLint、
  Vite dev transform、Edge 桌面/移动端真实路由与交互、支付请求体检查、`git diff --check`。
- 完整 typecheck/build 保留为明确的仓库依赖基线风险，不作为 S179 弹窗行为 FAIL；不得据此声称全仓构建通过。
