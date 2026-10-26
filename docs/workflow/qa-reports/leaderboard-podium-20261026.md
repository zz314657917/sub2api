# 排行榜前三领奖台卡片：提交前复核

日期：2026-10-26。新增提交范围为排行榜组件、其中英文词条、现有组件测试和本报告。

## Findings

- 未发现本批改动的明确阻断问题。前三按第二、第一、第三名展示；第一名居中抬高，使用皇冠和金色边框，第二、第三名分别使用银色和铜色头像框。
- 卡片显示头像、昵称、Token、请求数和当前周期总 Token 占比，保留称号、排名变化、当前用户标记以及鼠标/键盘可访问的 Token 明细。第四名起保留横条；不足三人时按实际名次定位。
- 本批没有新增依赖、后端接口、迁移或配置变更。共享工作区其他任务的源码、依赖清单和混合任务文档不纳入本次新增提交。

## Executed Checks

- 提交前逐行核对四个源码/测试文件的完整 diff，定向 `git diff --check` 通过。
- `npm.cmd run test:run -- src/views/user/__tests__/LeaderboardView.spec.ts src/__tests__/leaderboard-theme.spec.ts`：40 项通过。
- 复核开始时本地已有提交 `d2cd2835e`（用量记录默认今天），补跑 `npm.cmd run test:run -- src/views/user/__tests__/UsageView.spec.ts`：33 项通过。提交前并行任务已将该提交及监控样式提交同步到远端，排行榜目标文件哈希未变，本次新增提交基于同步后的 `4cbe4b025`。
- `npm.cmd run typecheck` 和四个目标源码/测试文件的定向 ESLint 通过。
- 实施阶段执行 `npm.cmd run build -- --outDir E:/codex-runtime/leaderboard-podium-20261026/dist --emptyOutDir` 通过；输出位于独立目录，未覆盖后端静态资源。保留既有 Browserslist 数据提示与大分块提示。
- 真实 Vue 页面配模拟 API 检查深浅两主题，320、390、768、1024、1440、1920px 共 12 组：前三顺序、冠军高度和头像尺寸、第四名后列表数量正确，无页面横向溢出。
- 日/周/月/总榜切换、键盘明细提示、0/1/2 人场景通过，无页面异常。本次独立浏览器 profile、CLI daemon 和预览进程已关闭。

实施证据：`E:/codex-runtime/leaderboard-podium-20261026`。
提交前复核证据：`E:/codex-runtime/leaderboard-podium-publish-20261026`。
截图：`output/playwright/leaderboard-podium-cards-dark-1920.png`、`output/playwright/leaderboard-podium-cards-dark-390.png`。

## Unverified Risks

- 浏览器使用模拟榜单数据，未联调真实后端、真实用户数据或线上服务。
- 构建基于本地现有依赖环境；工作区其他任务的依赖更新不属于本批发布范围。
- 提交和推送不会更新正在运行的服务、容器或线上静态资源。

## Recommendation

可按精确文件范围提交并正常推送；部署和线上验收留待单独授权。
