### PASS: group-fast-20261103

2026-11-03，选择性移植上游 PR #6443，已合入本地主目录；未提交、
推送、更新容器、运行迁移或部署。

## Findings

本任务增量未发现明确缺陷。分组默认关闭，创建/编辑均可启用
`force_openai_fast`；只对可信 OpenAI/Composite 分组的实际 OpenAI
上游请求强制 `service_tier=priority`，再执行原全局过滤/拦截策略。
HTTP 五入口、两种 WebSocket 首帧/后续帧均覆盖，复制、鉴权缓存主分组/
路由分组及旧版本拒绝均已接通。计费和价格代码保持原样，未加入 Free Fast。

四项既有回归失败在独立 baseline 和 candidate 上复现相同结果，保持原断言：

- `TestAdminService_CreateGroup_AllowsStandardPeakRate`
- `TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost`
- `TestForwardAsChatCompletions_UnknownModelDoesNotUseDefaultMappedModel`
- `TestOpenAIGatewayService_ResponsesUnknownModelDoesNotFallbackToGPT54`

## Executed Checks

- 独立合同评审 PASS：最终合同 SHA256
  `d030d97471765c44c50a9d92652af3605729da086b50e76adc4bc584799c724f`。
  首版两条上游文案路径已改为本地 `admin/groups.ts` 并独立重审。
- 两个隔离 Developer，组合37文件 manifest
  `e0e9e24aa625e2517899990cd175521880a5b0894ea4eb4688401deb95ebdaa6`；
  独立 QA seal `d42051bed730cc829f31a3235853f90a5ef6135a194345f5b769f2657be92aa0`。
- 独立 QA 新增 Go default160/unit172、额外边界探针各33通过；
  既有广回归 default801通过/2既有失败、unit980通过/4既有失败。
  Messages 各19通过，计数可能与广回归重叠，不加总宣称完整套件全绿。
- 独立前端48/48（新增26、既有22）、类型检查、7文件 ESLint、Vite
  实际构建通过。mounted 测试实际点击两张表单的保存按钮。
- 主控精确集成后，actual-main Go 定向 default980/unit1171、
  前端48、类型检查、目标 ESLint、后端及前端构建通过。
  使用已安装 `eslint.cmd`，直接 Node 启动遗漏 pnpm 解析路径的首次退出2
  保留在日志，未改依赖。
- 并行任务在合入前更新订阅事务/Key状态及两个对应测试；四项非本任务
  变化保持，逐项原始 hash/diff/时间记录在 `concurrent-drift.json`。
  actual-main 另补 Go `TestGapFix` 两标签各72和 Key 重置10项通过。
- 精确集成瞬间保护12350原路径，最终4886个冻结原输入按原字节/已记录
  并行增量复核通过；HEAD及原始index SHA不变、暂存区为空。
  14项隔离 checkout 的纯换行差异已在冻结前恢复为原始 main 字节并留证。
- 任务差异/Ent纯字段增量/gofmt检查通过；无价格、依赖或旧断言任务增量。

## Unverified Risks

- `254_group_force_openai_fast.sql` 仅为源迁移，未在数据库执行；需后续
  服务更新使新字段和界面上线。未验收真实 provider、跨节点缓存和真实
  付费结算；本地 HTTP/WS 使用 localhost/受控 transport。
- 已建立 WebSocket 会话需重连。没有真实浏览器视觉验收或部署声明。
- 四项既有失败未在本任务修复，不能将完整仓库测试描述为全部通过。

## Recommendation

PASS：本地选择性合入完成，可进入另行授权的提交或部署流程。
所有独立角色已结束，未保留任务服务或浏览器。

完整命令、日志、冻结及精确集成证据：
`E:/codex-runtime/sub2api/group-fast-20261103/qa/qa-report.md`、
`main-acceptance.json`、`integration.json`、`final-preservation.json`。
