### DONE: public-docs-layout-toolbar-s63

## Changed Files
- `frontend/src/views/public/TutorialView.vue`
- `frontend/src/views/public/ModelPlazaView.vue`
- `frontend/src/views/public/__tests__/TutorialView.spec.ts`
- `frontend/src/views/public/__tests__/ModelPlazaView.spec.ts`
- `frontend/src/__tests__/public-pages.spec.ts`

## Result
- 教程详情的左教程目录、中间正文、右本页目录已提升为三个同级列；宽屏正文上限为 `50rem`，文章整体不再使用卡片背景、边框和阴影。
- `1101-1360px` 隐藏左目录的次要分类标签，避免中等桌面宽度出现截断；`<=1100px` 保留移动目录与移动 TOC。
- 模型搜索框独占第一行；分类 tabs 与结果数量位于同一条扁平 filter row；toolbar 和 tabs 容器无面板背景、边框、阴影或 blur。

## Worker Checks
- Tutorial worker：`TutorialView.spec.ts` 5/5 PASS，`npm.cmd run typecheck` PASS，`git diff --check` PASS。
- Model worker：`ModelPlazaView.spec.ts` 5/5 PASS，定向 lint 与 `git diff --check` PASS。

## Contract Compliance
- 未修改 API、路由、CMS 数据、模型目录或计价逻辑。
- 未回滚 S58-S62 或其他 mixed dirty tree 内容。
