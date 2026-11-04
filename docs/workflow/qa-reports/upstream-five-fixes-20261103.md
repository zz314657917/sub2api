# 上游五项修复本地适配验收

日期：2026-11-04。当前阶段：精确18文件已合回，本地适配验收完成。

## Findings

本批适配 PR #8471、#8477、#8432、#7761、#8436：

- 非法 UTF-8 Key 在缓存和仓储前拒绝；原始请求取消返回499；
  请求或仓储超时、数据库故障保持500，两个认证封装一致。
- Anthropic API Key 初次透传按映射模型过滤无效历史 thinking，
  保留合法签名、文本及现有 passback 协议行为。
- 图片账号探测不再发送 response_format，生产图片计费和转发不变。
- 调度缓存投影保留 openai_capabilities；topK前过滤和完整账号复查保留。
- 用户 CSV 每页1000条；固定初始筛选、排序、日期、总页数和文件名；
  仅429按 Retry-After 重试，每页最多3次重试，等待1至60秒，
  缺失或非法值默认60秒；失败不下载部分数据，导出锁正常释放。

11项既有 routing 失败在原始 baseline、原候选和组合候选保持相同断言；
未修改旧断言或把扩展回归表述为全绿。两次并行模型广场输入漂移触发
保护失败，原FAIL保留；精确R1/R2输入经独立审查，只作为非目标输入，
不纳入本批18文件交付。

## Executed Checks

- 完整合同SHA `3ecd5d04` 独立PASS，固定五份上游diff及PR head。
- 原始12388路径冻结 `a8d2bcc9`；交付18路径 `2f7aa4b5`；
  原候选12394输入 `26b1690d`，组合12394输入 `b5d6d3bf`。
- 独立正式QA seal `8b0f8ba9`、报告 `0fd83014` 绑定390证据，
  主控逐项SHA复核PASS。独立Go33叶子、前端40项通过；新增Go默认/unit
  各55通过事件（48叶子/10顶层），middleware303事件（270叶子）。
- 原前端59项保留，新增后两份117项通过；typecheck、定向lint、
  前后端构建通过。所有构建输出在外部目录，未覆盖后端静态资源。
- 语言fixture首次缺生产JIT配置导致2FAIL，按真实runtime-only alias
  和JIT define重测两语言通过；原日志保留，不修改依赖或全局i18n。
- 组合候选已执行新增Go双标签、middleware、受影响526声明及panic后
  精确补跑、backend build；11旧失败相同，0missing/skip。
- ModelMarket unit17顶层29事件通过；默认原命令因buildtag零测试保留，
  仓库外仅移除首行约束的原字节overlay实际运行同17顶层29事件通过。
- 主控组合首尾12394输入、source18/archive及原390证据均一致；
  前端1048输入不变，原封印前端证据可复用。
- 原独立QA已实际执行组合后端检查，断流前未另写组合seal；主控逐项
  验证原始日志、526声明覆盖和11原失败，补做默认ModelMarket及完整
  首尾身份检查后，以Final Evaluator接受组合版本。未返回的追加审签
  不计独立PASS，详见外部controller-final-input-decision.md。
- 精确18已合回主目录，integration-before/after保护PASS；
  原12388路径中的非目标输入、HEAD和逻辑index保留。
- fresh主目录新增Go默认/unit各55通过事件、middleware303、
  ModelMarket29、前端两份117项均通过，无新增失败或跳过；
  后端build、前端typecheck/定向lint/build、gofmt及18路径diff检查通过。
  启动时核对全部backend/frontend输入等于组合清单，结束核对交付18
  字节不变；实际命令、日志和摘要见外部`main-final/summary.json`。
- 最终完整性记录为外部`final-protection.json`与`controller-final.json`，
  绑定交付清单、组合清单、原正式QA seal和fresh验收记录。
- 变更Go gofmt及定向diff检查通过，完整差异已逐行审查。

## Unverified Risks

未访问真实provider、数据库、Redis、浏览器或生产环境。
httptest、出站recorder和mounted Vue只证明本地运行行为。
广域回归仍保留上述11项既有routing失败，不能据此宣称整个测试套件全绿。

## Recommendation

五项修复可作为已完成的本地适配交付。所有任务角色已关闭；
实际provider/DB链路仍需对应环境验收，未更新容器或部署。

## Publication Scope

用户于2026-11-04追加授权“整理下 提交 推送”。本次提交仅包含上述18个
源码/测试文件及本报告；先前三项适配、Pelican、模型广场和共享流程文档
继续保留在工作区。推送目标为origin/main，并包含该分支既有的
`77b338b71929966688ddc5528783a902572346b3`用量展示提交。

发布前以HEAD加精确19路径构造独立Git快照，使用该快照的依赖声明执行
定向测试、类型检查和前后端构建；不得以混有其他未提交源码的工作区结果
代替提交树验收。实际命令、提交SHA和远端回读见本任务证据目录中的
`publication-20261104/`，本地适配阶段的证据和既有失败记录保持不变。

证据：`E:/codex-runtime/sub2api/upstream-five-fixes-20261103/`。
候选：`E:/codex-worktrees/sub2api/upstream-five-fixes-20261103/`。
