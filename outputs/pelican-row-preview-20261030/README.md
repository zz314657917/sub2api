# 鹈鹕测试右侧排版预览

在已安装前端依赖的仓库根目录启动：

```powershell
$env:PELICAN_PREVIEW_PORT='62081'
node outputs/pelican-row-preview-20261030/build-integrated.cjs
node outputs/pelican-row-preview-20261030/serve-integrated.mjs
```

打开 http://127.0.0.1:62081/pelican-tests。默认使用本仓库 `frontend`，不指定端口时使用 62080。需要预览其他隔离候选时，可在启动前指定其绝对路径：

```powershell
$env:PELICAN_PREVIEW_FRONTEND='<候选工作区>/frontend'
```

服务绑定本机 `127.0.0.1`。缓存默认位于本仓库 `.cache`，可通过 `PELICAN_PREVIEW_CACHE` 指定其他目录。已提交 SVG 素材可直接使用，运行预览不依赖素材生成脚本。

预览直接复用现有 `PelicanTestsView.vue`、`AppLayout.vue`、`AppSidebar.vue` 和 `AppHeader.vue`。独立 Vite 服务仅在内存中替换原视图右侧的画廊部分，原侧栏、顶部栏、设置与计划入口、底部分页继续由原组件渲染。

画廊使用三组、六个模型、30 张演示记录，每个模型默认保留 5 张，单模型最多展示 20 张。计划编辑器在预览中也显示默认 5、上限 20。卡片将状态放在模型名称最右侧，删除独立的状态与计划行，保留时间和耗时。支持分组内模型筛选、横向浏览、作品放大、搜索和深浅色。所有 API 调用在预览入口中拦截，写操作直接拒绝，不连接真实后端。

横向画廊和留存数调整只在本预览目录实现。雷达参考评分另有正式前后端实现；此预览通过 `radar-fixture.json` 提供模拟接口，不代表真实后端运行或部署。默认选中原来的五个模型及其全部来源档位，共 29 个组合；按精确模型与档位从冻结来源计算各自分数，来源日期和同步时间保留 fixture 原记录，不冒充实时刷新。生产设置保持管理员保存的选择，不自动扩充。旧 `index.html` 保留为独立草图，当前服务首页与 `/pelican-tests` 均进入原控制台中的集成预览。

雷达异常预览参数：`?radar=off|stale|unavailable|missing|unsupported`；普通用户预览：`?role=user`。写操作仍拒绝，不连接真实数据库或模型服务。本任务不更新容器。

测试计划入口先展示列表，默认提供两个停用的演示计划，可点击编辑或显式新建；`?plans=empty` 检查空列表状态。演示计划没有写入真实数据库，保存和运行继续由预览拒绝。

评分区每个模型一张卡，卡内各档位的强度与分数同行；最高分数字前显示皇冠，并列最高均标记。普通分数浅灰，最高分浅色主题纯黑、暗色主题纯白。满分统一在区域标题标注。集成画廊不提供重复的主题按钮，使用原侧栏切换。

可提交源码、素材与冻结 fixture 保存在本目录；截图、历史验证 JSON 和进程记录保留在本地证据目录，不纳入版本库。提交前检查和运行态证据摘要见 `docs/workflow/qa-reports/pelican-radar-publish-20261030.md`。
