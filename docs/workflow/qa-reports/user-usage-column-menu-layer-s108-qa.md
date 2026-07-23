### PASS: user-usage-column-menu-layer-s108

# S108 User Usage Column Menu Layer QA

## Findings

- 未发现 S108 实现缺陷。用户端筛选卡仅在 `showColumnMenu` 为真时应用
  `relative z-[221]`，高于 `DataTable` 固定表头的最大 `z-index: 220`；菜单关闭后
  临时层级移除。
- 列设置菜单仍使用原有定位、滚动、选项切换和持久化逻辑；共享
  `DataTable`、`TablePageLayout` 与后端均未修改。
- 1366x768 Chromium smoke 中，菜单与表头实际重叠，重叠点最上层元素为菜单项。
  DOM 采样为 `filterZ=221`、`maxHeaderZ=220`、`menuOwnsOverlapPoint=true`，
  最终判定 `pass=true`。

## Executed Checks

- `node_modules/.bin/vitest.cmd run src/views/user/__tests__/UsageView.spec.ts`：
  PASS，1 个文件、20 个测试通过。
- `npm.cmd run typecheck`：PASS。
- `npm.cmd run build`：PASS，Vite 生产构建完成 1090 modules，产物包含 S108
  用户端 `UsageView`。
- `node_modules/.bin/eslint.cmd src/views/user/UsageView.vue src/views/user/__tests__/UsageView.spec.ts`：PASS。
- `git diff --check`：PASS。
- Playwright CLI headless Chromium：PASS。使用合成登录态、分组、密钥、用量和
  统计 API 数据打开 `http://127.0.0.1:62108/usage`，点击列设置后完成快照、
  DOM 层级/重叠点检查和截图；干净会话控制台为 0 error / 0 warning。
- Screenshot: `output/playwright/user-usage-s108/user-usage-column-menu-s108-clean.png`。

## Unverified Risks

- `--headed` Chrome 首轮启动在创建临时自动化配置后退出，因此最终浏览器证据来自
  headless Chromium；已实际渲染页面并做截图、像素可见性和 DOM 命中检查。
- 未使用真实用户登录态或真实后端数据；API 仅为稳定复现 UI 的合成数据。真实数据
  不改变本次纯 CSS 层叠关系。
- 生产构建保留既有 browserslist 过期、动态/静态 import 和大 chunk warning，
  均未导致构建失败，也不由 S108 引入。

## Contract Compliance

- Success Criteria：满足。动态父层级高于表头，菜单开关前后由定向测试锁定，
  浏览器重叠场景通过。
- Allowed/Denied Paths：满足。业务实现只修改用户 `UsageView` 及其专用测试；
  未修改共享表格/布局、backend、部署或容器。
- Constraints/Stop Rules：满足。未引入 Teleport，未改变全局表格层级、菜单定位、
  关闭逻辑、列显隐或 API 行为。

## Recommendation

- Final Evaluator 可判定 `PASS / source-only`。本轮不提交、不推送、不部署、
  不更新容器。
