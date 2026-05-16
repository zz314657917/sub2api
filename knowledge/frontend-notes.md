# 前端开发笔记

最后更新：2026-05-15

## 技术入口

- Vue 3 + TypeScript + Vite。
- 路由在 `frontend/src/router/index.ts`。
- API client 在 `frontend/src/api/`。
- Pinia store 在 `frontend/src/stores/`。
- i18n 在 `frontend/src/i18n/locales/zh.ts` 和 `frontend/src/i18n/locales/en.ts`。
- 通用样式在 `frontend/src/style.css` 和 `frontend/src/styles/console-ui.css`。

## 公共页

入口：

- `/home` -> `frontend/src/views/HomeView.vue`
- `/tutorial` -> `frontend/src/views/public/TutorialView.vue`
- `/models` -> `frontend/src/views/public/ModelPlazaView.vue`
- `/legal/:documentId` -> `frontend/src/views/public/LegalDocumentView.vue`

共享组件：

- `frontend/src/views/public/components/PublicTopNav.vue`
- `frontend/src/views/public/components/PublicMatrixBackdrop.vue`
- `frontend/src/views/public/public-page.css`

当前设计方向：

- 公共页倾向统一到深色 Matrix / 绿紫点缀风格。
- 教程页强调可跟读的新手接入文档，Codex 内容顺序优先于 Claude。
- 模型广场需要关注真实数据下筛选区、分组下拉、倍率开关和卡片拥挤度。

常用测试：

```powershell
pnpm.cmd exec vitest run src/__tests__/public-pages.spec.ts
```

## 控制台布局

关键文件：

- `frontend/src/components/layout/AppHeader.vue`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/components/layout/AppLayout.vue`
- `frontend/src/components/layout/TablePageLayout.vue`
- `frontend/src/styles/console-ui.css`

注意：

- 侧栏、顶部栏、客服入口、教程/模型广场快捷入口会影响用户和管理端共同体验。
- 侧栏/i18n 改动通常需要同步测试 `AppSidebar.spec.ts` 和中英文文案。
- 表格页不要随意引入大面积视觉重构，优先保持可扫描、密度适中和状态清晰。

## 客服弹窗

关键文件：

- `frontend/src/components/common/SupportPopup.vue`
- `frontend/src/utils/supportContent.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/api/admin/settings.ts`

当前能力：

- 支持多个二维码卡片。
- 支持每个二维码下方说明。
- 支持覆盖角标。
- 只有纯文本联系方式时也应显示兜底卡片。

## 用户侧新能力

近期活跃页面：

- `frontend/src/views/user/DashboardView.vue`
- `frontend/src/components/user/dashboard/UserDashboardAccountUsage.vue`
- `frontend/src/views/user/ChatImageStudioView.vue`
- `frontend/src/views/user/ChatStudioView.vue`
- `frontend/src/views/user/ImageCreatorView.vue`
- `frontend/src/views/user/PaymentView.vue`

注意：

- 用户侧 API 变更要同步 `frontend/src/api/user.ts`、`frontend/src/types/index.ts` 和页面测试。
- ChatStudio / ImageCreator 属于交互复杂页面，改动后至少跑对应 Vitest。
- ChatImageStudio 的会话是 `localStorage` 本地状态，key 为 `sub2api:chat-image-studio:v1`；删除会话不会删除或取消服务端图片任务，详见 `knowledge/chat-image-studio.md`。
- Payment 页面历史上出现过并行改动导致 typecheck 噪声，遇到时先查 `git diff` 确认责任范围。

## 前端验证建议

- 小组件：跑对应组件测试。
- 页面文案/i18n：跑页面测试和相关 store/API 测试。
- 公共页视觉：跑 public-pages 测试，并用浏览器检查桌面/移动端。
- 构建前：`pnpm.cmd run typecheck` 和 `pnpm.cmd run build`，若失败需区分既有并行改动和本轮改动。
