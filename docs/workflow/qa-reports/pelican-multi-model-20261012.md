### PASS — 完整时间 tooltip 已定向修复并通过独立复验；本轮合同验收通过。

日期：2026-10-12。独立 QA：`gpt-6.1-sol`。隔离树：`E:/codex-worktrees/sub2api/pelican-multi-model-20261007`。

## Findings

**已修复：生成时间 tooltip 完整时间。** 原 `title` 复用短格式的 P2 问题，Developer 改为独立 `formatTimestamp` 输出 ISO 秒/毫秒和 `(UTC)`。QA 真实 Chrome DOM 读到 `2026-10-07T02:00:00.000Z (UTC)`，紧凑 display 保持 `2026/10/7 10:00`。失败模型仍引用旧成功作品 10:00（UTC 02:00），本轮失败时间为 11:00；未误用失败时间。当前未发现未修复合同缺陷。本轮 QA 未修改业务源码。

## Executed Checks

- 真实 Chrome session `pelican-multi-model-qa-20261012`，profile `E:/codex-playwright-profiles/pelican-multi-model-qa-20261012`，浏览器 PID `31472`；独立 fixture Vite `62081` PID `43292`。
- 多模型编辑：原计划 #2 两模型正确选中；切到 group 4 清空旧模型，只有 `claude-sonnet-4` 可选、已选 0；切回 group 3 不恢复旧选择。按 sol→astra 顺序选中保存，请求 `model_ids=["gpt-6.1-sol","gpt-6-astra"]`、首项 alias 正确；重新编辑两项保留。
- 多模型创建：按 astra→sol 选择，预估每轮 2 次、每日 48 次；真实 UI 点击创建，请求 `model_ids=["gpt-6-astra","gpt-6.1-sol"]`、`enabled=false`。证据 `outputs/pelican-multi-model-20261007/qa-20261012-browser-create-edit.json.txt`。
- Replay：真实点击后 iframe DOM 被替换，sandbox 仍为 `allow-scripts`。
- Enlarge：真实点击后作品 iframe 出现在 `.enlarged-art`，弹窗明确对应 gpt-6-astra / plan #2；Escape 关闭恢复焦点。
- History：真实打开 gpt-6.1-sol / plan #2；在弹窗中触发列表刷新，gallery 调用累计 8 次后弹窗仍为同一模型，history 两次请求都带相同 `plan_id=2, model_id=gpt-6.1-sol`。自动刷新设置 15 秒，弹窗暂停轮询。证据 `qa-20261012-history-identity.txt`。
- 组卡：plan #2/#12 同属 group 3 合为一组卡；同模型不同计划保持独立，未知模型有 fallback；旧成功作品生成于 10:00，与失败测试 11:00 分开。
- 真实桌面截图 `qa-20261012-desktop.png`；390px dark class 截图 `qa-20261012-mobile-dark.png`，实际检查 `scrollWidth=clientWidth=390`；长预览 580px、短预览 246px。
- Canonical Go 专项：`GOMAXPROCS=2 GOMEMLIMIT=1GiB go test -p 1 ./internal/service ./internal/repository -run PelicanMultiModel -count=1 -v` exit 0。PG 真实连接任务库 `127.0.0.1:55487`，通过独立 migration/CRUD、history/retention、pagination/authorization、quota/concurrency/stale-write 检查；无 PG skip。日志 `qa-20261012-multimodel.log`。
- 前端：`NODE_OPTIONS=--max-old-space-size=4096 node node_modules/vitest/vitest.mjs run src/components/pelican/__tests__/PelicanTestsView.spec.ts src/components/pelican/__tests__/PelicanPlansPanel.spec.ts --maxWorkers=1 --minWorkers=1` exit 0，2 files / 29 tests。日志 `qa-20261012-vitest.log`。
- Scoped ESLint：设置 `NODE_PATH=<frontend/node_modules/.pnpm/node_modules>` 后执行现有 `.eslintrc.cjs`，Pelican API/view/panel 与两测试文件 exit 0，无输出。只补已安装解析器搜索路径，未安装依赖或修改锁文件。日志 `qa-20261012-eslint.log`（成功时为空）。
- `git diff --check` exit 0。
- 2026-10-07 的后端完整 Pelican suite/build、前端完整 typecheck/build 结果作为既有证据，本轮未重复；canonical tests 的新增不影响旧构建证据。
- 清理：session `close` 成功；精确 Vite PID 命令行确认后停止；PG PID `22296` 使用独占数据目录 `pg_ctl -m fast stop`。最终本任务 Chrome/Vite/cliDaemon 查询无输出，62081/55487 无监听；未停止共享容器、用户预览或其他浏览器。
- 修复后定向复验：新 session `pelican-tooltip-qa-20261012`，独立同名 profile，Chrome PID `31780`、Vite PID `9948`。上述完整 title 与 fallback 时间来源真实浏览器通过，证据 `qa-20261012-tooltip-browser.txt`。
- 新增测试只运行 `complete generation timestamps`：1 PASS / 17 intentionally skipped，exit 0；覆盖非零秒、时区、fallback 来源和无效时间。日志 `qa-20261012-tooltip-vitest.log`。第一次从 repo cwd 执行缺 alias 配置，改为正确 frontend cwd 后通过；不属于业务问题。
- 复验 session 已 close，精确 Vite PID 已停止；最终本任务 Chrome/Vite/cliDaemon 查询无输出，62081 无监听。未重启 PG。

## Unverified Risks

真实 provider、生产库、部署未执行，合同范围无需付费调用。浏览器使用真实组件加模拟 API，后端行为由独立 PG/Go 验证。

## Recommendation

主控可执行此前授权的本地主线白名单整合，保留 canonical tests 与本轮报告。本 PASS 不授权 push/deploy，也不代表真实 provider 或生产环境验收。其他已通过项无需重复全量重跑。
