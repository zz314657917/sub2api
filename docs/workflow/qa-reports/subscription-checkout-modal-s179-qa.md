### PASS: subscription-checkout-modal-s179

## Findings

- 未发现 S179 范围内明确问题。套餐选择后由现有 `BaseDialog` 打开居中确认弹窗，原文档流确认卡已移除。
- `selectedPlan.price`、`order_type=subscription` 和 `plan_id` 请求语义保持不变；没有后端、数据库、支付 provider 或生产资源改动。
- 提交期间关闭函数拒绝清空选择，Esc、遮罩和取消入口也由 `submitting` 状态禁用或守卫。

## Executed Checks

- `PaymentView.spec.ts`：`9/9 PASS`，覆盖 BaseDialog 接入、付款阶段显示条件、关闭函数、旧确认卡移除和订阅金额不受余额充值倍率影响。
- ESLint：`PaymentView.vue` 与 `PaymentView.spec.ts` 均为 `PASS`。
- Edge mocked-browser：实际 `/purchase` 路由点击套餐后出现 `role=dialog`；桌面为居中双栏，`390x844` 为单栏；文档和弹窗均无横向溢出。
- 关闭路径：Esc、取消按钮、遮罩点击均使 dialog count 变为 0，并移除 `body.modal-open`。
- mocked `POST /api/v1/payment/orders` 请求体为 `amount=10`、`payment_type=alipay`、`order_type=subscription`、`plan_id=301`；返回二维码等待态后 dialog 消失，实际 `PaymentStatusPanel` 显示“支付宝扫码支付”。
- 截图：`frontend/output/playwright/subscription-checkout-modal-s179/desktop-dialog.png`、`mobile-dialog.png`。
- `git diff --check`、未合并索引检查和后台进程/Playwright session 清理均通过。

## Unverified Risks

- `npm.cmd run typecheck` 与 `npm.cmd run build` 被当前仓库依赖状态阻断：`AirwallexPaymentView.vue:103` 无法解析 `@airwallex/components-sdk`，pnpm 缓存目录没有该包本体。该依赖问题不在 S179 改动或 Allowed Paths 中，未修改锁文件或依赖。
- 浏览器支付成功场景使用本地 fixture，没有真实商户、真实二维码、回调、结算、部署或生产验证。
- 浏览器 console 中的 Airwallex checker 提示、历史 i18n fallback warning 和 fixture 外围 store warning 不属于 S179 弹窗回归；弹窗交互和请求证据不依赖这些 warning。

## Contract Compliance

- Product source changes only: `frontend/src/views/user/PaymentView.vue`、`frontend/src/views/user/__tests__/PaymentView.spec.ts`。
- 未修改后端、共享 `BaseDialog`、全局 CSS、数据库、配置、部署或生产资源。
- 工作树已有 S177 文档改动被保留；S179 只向共享 main log 追加自己的记录。

## Recommendation

`可继续提测 S179；发布前另行恢复 @airwallex/components-sdk 依赖并补跑全量 typecheck/build。`
