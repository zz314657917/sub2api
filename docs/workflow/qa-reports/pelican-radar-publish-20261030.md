# 鹈鹕第三方参考评分与集成预览提交前复核

日期：2026-10-30

## Findings

未发现本次提交范围内的明确阻断问题。

- 正式实现：15 个业务与测试文件，包含第三方参考评分适配、30 分钟进程缓存、并发刷新合并、旧缓存保留、管理员配置、只读接口和原鹈鹕页面嵌入。
- 同一模型的已选档位放在同一张卡内，标题统一标注来源满分；最高原始分数前显示皇冠，并列同样标记。普通分数浅灰，浅色主题最高分纯黑、暗色主题纯白。
- 管理测试计划默认展示列表；仅在显式点击新建或编辑后进入表单。
- 20 个横向画廊预览文件单独提交，包含源码、已生成素材和冻结来源 fixture。不会把演示账号、演示计划或 fixture 接入正式应用。
- 其他任务的依赖、CI、网关、账户使用进度及混合工作流文档不纳入提交。截图、临时 HTML、凭据、二进制和本地配置不纳入提交。

## Executed Checks

### 提交快照

- 基线：`7f1202a2551919b5f7e5bca21df8052e9e224fe3`。
- 使用独立 Git index，从基线仅加入本次精确清单，再导出到独立目录验收；不依赖工作区里其他未提交源码或依赖清单升级。
- 后端沿用基线 `go.mod` / `go.sum`，以本机 Go 1.26.10 执行验证。前端复用本机已安装的依赖，未进行全新锁文件安装。
- 目标清单与 SHA256：`E:/codex-runtime/sub2api/pelican-publish-20261030/manifest.json`。

### 本轮执行

以下命令在独立快照内执行：

```text
npm.cmd run test:run -- src/components/pelican/__tests__ src/stores/__tests__/pelicanMetadata.spec.ts
npm.cmd run typecheck
npx.cmd eslint <本次七个前端文件>
npm.cmd run build
go test ./internal/service ./internal/handler ./internal/server/routes -run TestPelican -count=1 -timeout=180s
go build ./...
```

- 前端：11 个测试文件，100 项通过；类型检查、定向 ESLint 和生产构建通过。
- 后端：service / handler 定向测试通过；routes 包编译通过，无名称匹配的测试；完整构建通过。
- Go 目标文件的 `gofmt -l` 无输出。独立 index 的 `git diff --cached --check` 通过。
- 前端构建存在既有 Browserslist、动态导入与大 chunk 提示，退出码为 0。

### 浏览器与预览

- 从独立快照启动预览 `http://127.0.0.1:62085/pelican-tests`，HTTP 200。预览入口按当前目录生成，支持仓库移动；不依赖原 `F:` 工作目录入口。
- Edge 独立 session：`pelican-publish-20261030`；profile、PID、创建时间均记录在对应本地证据目录。
- 1440px / 390px、明暗主题共四个视口全部通过：5 张模型卡、29 个准确档位；皇冠位于最高分数字前；分数颜色正确；无文字重叠和整页横向溢出。
- 冻结 fixture 的精确最高分逐模型核对通过。显示相同舍入数值不视为并列；并列、零分和不可用项排除另由组件测试覆盖。
- 原横向画廊保留，重复模式按钮已移除；预览请求均为 GET。
- 更早的原页面浏览器交互已验证：侧栏主题切换、搜索、刷新、作品放大，以及计划列表、新建、编辑、空列表和重新打开。
- 本轮已执行 CLI close，`processes-closed.json` 中 `owned_processes=[]`。

本轮完整日志、截图及进程证据：`E:/codex-runtime/sub2api/pelican-publish-20261030/`。
此前合同与独立评审见同目录下的任务、合同评审文件；此前交互证据保存在本地 `pelican-plan-entry-20261030`、`pelican-remove-theme-20261030` 等目录。

## Unverified Risks

- 未更新容器、部署、操作真实数据库或运行真实模型请求。
- 预览使用冻结的第三方来源 fixture，不代表部署后的实时同步；本轮没有重新联网抓取来源。
- 横向画廊、每模型默认 5 张及最多 20 张的调整仍属于布局预览；正式页面保留既有画廊与留存行为。
- 本轮未在全新安装依赖的环境中重复检查，也未重复全仓测试。

## Recommendation

可按精确清单提交并推送：正式功能及计划入口修复一批，集成预览源码一批。保留其他工作区改动，部署与容器更新另行处理。
