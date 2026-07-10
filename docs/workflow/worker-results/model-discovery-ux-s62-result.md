### DONE: model-discovery-ux-s62

## Changed Files
- `frontend/src/views/public/ModelPlazaView.vue`
- `frontend/src/views/public/__tests__/ModelPlazaView.spec.ts`

## Implementation
- 搜索改为行级过滤，并与分类组合；显示型号/规格和分组结果数量。
- 移动端长分组默认预览 6 行，提供展开/收起，移除不可发现的隐藏内滚动。
- 增加账号分组预览与 `✪` 单位说明，默认分组选择接近 1x 而不是最低倍率。
- 提供直达快速开始教程入口；刷新失败时保留上次成功目录并显示非阻断提示。
- Evaluator 退修补齐筛选 `aria-pressed`、搜索 `:focus-within`、交互 `:focus-visible` 和直达路由测试。

## Commands Run
- `npm.cmd run test:run -- src/views/public/__tests__/ModelPlazaView.spec.ts public-pages public-smoke`
  - Worker 首轮 PASS：3 files / 20 tests；退修后模型专测 4/4。
- `npm.cmd run typecheck`
  - PASS。
- scoped `git diff --check`
  - PASS。

## Contract Compliance
- 只修改模型公共页和专用测试；未修改目录 API、计费、账号权限、路由或后端。

## Risks
- 模型比较器明确不在 S62 范围；移动展开真实大目录仍需浏览器几何验收。
