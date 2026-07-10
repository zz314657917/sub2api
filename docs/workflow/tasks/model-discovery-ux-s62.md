# Task Contract

## Task ID
model-discovery-ux-s62

## Role
Developer Worker，负责公共模型发现体验实现；主 Codex 负责 diff review 与最终裁决。

## Goal
提升模型广场搜索、移动长表和价格理解，让用户能快速找到具体型号、理解当前价格口径并进入接入教程。

## Allowed Paths
- `frontend/src/views/public/ModelPlazaView.vue`
- `frontend/src/views/public/__tests__/ModelPlazaView.spec.ts`
- `docs/workflow/worker-results/model-discovery-ux-s62-result.md`

## Denied Paths
- 模型目录 API、价格计算后端、账号权限、路由、数据库、后端和部署。

## Required Behavior
- 搜索具体型号或规格时只渲染命中行，而不是保留整个命中分组。
- 分类与搜索组合过滤，清空后恢复全部数据，并显示结果数量。
- 移动端不再把超长表塞进无提示的隐藏内滚动区域；使用有限预览和“展开/收起”。
- 当前账号分组价格口径和 `✪` 单位有明确说明，不宣称匿名用户一定拥有最低倍率分组。
- 提供进入接入教程的明确入口。
- 移动端刷新按钮降为次要操作；已有数据刷新失败时保留旧目录并显示非阻断错误。

## Success Criteria
- 搜索具体型号/规格只显示命中行，分类组合与结果数量正确。
- `390x844` 下超长分组没有不可发现的隐藏内滚动，展开/收起可用。
- 价格口径、教程入口、刷新错误状态满足 Required Behavior。

## Constraints
- 不新增依赖，不改变 API/计费逻辑，只在 Allowed Paths 内写入并兼容 mixed dirty tree。

## Output
- 代码与定向测试；最终报告首行必须为 `### DONE: model-discovery-ux-s62`、`### BLOCKED: model-discovery-ux-s62` 或 `### FAILED: model-discovery-ux-s62`。

## Acceptance
```powershell
cd frontend
npm.cmd run test:run -- src/views/public/__tests__/ModelPlazaView.spec.ts public-pages public-smoke
npm.cmd run typecheck
```

## Stop Rules
- 如需修改计费计算、API 合约或后端价格数据，返回 `BLOCKED`。
- 不实现模型比较器，不覆盖当前工作区其他未提交改动。
