# Task Contract

## Task ID
home-reveal-background-cover-s59

## Goal
修复首页 relief reveal 背景在超宽屏无法覆盖宽度、窄屏因固定宽高产生上下留白或错位的问题。

## Allowed Paths
- `frontend/src/views/public/components/PublicRevealBackdrop.vue`
- `frontend/src/__tests__/home-theme.spec.ts`
- `docs/workflow/tasks/home-reveal-background-cover-s59.md`
- `docs/workflow/qa-reports/home-reveal-background-cover-s59-qa.md`
- `docs/workflow/status.md`
- `knowledge/tasks/current-task.md`

## Denied Paths
- 首页业务逻辑、认证、路由、后端、数据库、部署和背景图片文件本身。

## Success Criteria
- hero 背景使用容器级 `cover`，不再受 `2048px` 最大宽度限制。
- hero 背景高度始终跟随容器，不再使用移动端固定 `18rem` 高度或 `48rem` 背景宽度。
- 桌面精细指针继续使用 canvas reveal mask。
- 触屏、窄屏和 reduced-motion 场景禁用动态 mask，但保留低对比度静态背景。
- 超宽、常规桌面和移动视口均无背景空带或横向溢出。

## Acceptance
```powershell
cd frontend
npm.cmd run test:run -- home-theme public-pages
npm.cmd run build
cd ..
git diff --check
```
