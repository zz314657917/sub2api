### PASS: pelican-plan-accounts-20261024

本地实施与模拟浏览器验收通过，未部署。

日期：2026-10-24。目标为计划编辑弹窗和按选择顺序设置优先测试账号。

## Findings

未发现待修复的产品缺陷。独立合同 r3、两名 Developer 自检、独立 QA 和主控最终组合验证均已完成。测试账号优先顺序只作用于该计划，不改变账号全局调度优先级。

- 编辑/新建在嵌套弹窗中进行；取消不保存，保存期间禁止关闭，失败保留草稿并在弹窗显示错误，过期响应不会覆盖新的编辑会话。
- 下拉选择账号后按选择顺序追加；先选优先，可移除，编辑回显原顺序，切组清空，失效账号需显式移除。
- `preferred_account_ids` 由新迁移252持久化；空列表保持自动调度。优先账号按序检查和获取并发名额，不可用时顺延，耗尽后有限自动回退；provider 尝试后不换号重试。
- 保留分组、平台、模型、状态、额度、运行代际及并发限制，拒绝的租约及时释放，基础设施错误终止选择。

## Executed Checks

- 独立 QA：service 85个顶层测试、repository 8个顶层测试、handler 8个顶层测试通过；另13个独立后端边界场景通过；后端完整构建通过。
- 前端8个文件/72项测试、类型检查、6路径 ESLint、原项目 `vue-tsc -b` 和生产构建通过。sqlmock覆盖 create/update/list/claim/resume 顺序、清空和 nil 非 NULL；4份自建数据库 fixture 均追加252。
- 真实 Chrome、实际 AppLayout、模拟 API：29项检查通过、页面错误0；验证顺序 `[102,101]` 的显示及 PUT、取消/Escape/焦点返回、双层滚动锁、保存失败/重复提交/关闭保护、失效账号、切组与刷新失败。1440px和390px均无横向溢出，长列表滚动位置4717→4717。
- 全19个冻结文件与主工作区整合后的 SHA256 一致；4701个非目标原有受跟踪文件及39份受保护脏文件在整合前检查通过，另3份既有流程文件按本轮任务更新，HEAD/index不变。随后仅添加本轮流程与交接记录。
- 主工作区重新执行：三包 Pelican 定向测试、后端完整构建、前端72项测试、类型检查，全部 exit0。
- 本次浏览器/profile/daemon进程归零；预览PID25900及其esbuild子43368已精确清理，62084监听0。

## Evidence

- 合同及独立审查：`docs/workflow/tasks/pelican-plan-accounts-20261024.md`、`docs/workflow/contract-reviews/pelican-plan-accounts-20261024-review.md`。
- 外部证据根：`E:/codex-runtime/pelican-plan-accounts-20261024`。
- 独立 QA：`qa/qa-report.md`、`qa/qa-seal.json`、`qa/commands.json`、`qa/browser-smoke-r3.json`、`qa/browser-scroll.json`。
- 冻结与整合：`combined-freeze.json`、`integration-scope.json`、`main-checks.json`、`preview-cleanup.json`。
- 预览截图：`qa/r3/desktop-editor.png`、`qa/r3/mobile390-editor.png`。

原合同 r1 FAIL、浏览器接管初态超时、共享引用快照及外置 TypeScript 配置失败证据保留。后两项由 QA 修正记录/命令，不修改产品或原验收条件。pnpm shim 和 ESLint 顶层 parser 解析问题使用已存在包入口解决，未安装或修改依赖。

## Unverified Risks and Recommendation

10项真实数据库测试因禁用测试 DSN 跳过，不计作数据库通过；迁移252未执行，真实 PostgreSQL和provider均未验收。浏览器接口为模拟数据，不是前后端真实联调。

本地代码交付完成。没有提交、推送、容器更新或部署；后续部署需包含迁移252并另做真实数据库/账号联调。
