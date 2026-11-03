### PASS: pelican-layout-choice-20261103

## Contract Checked

- task_id: `pelican-layout-choice-20261103`
- contract: `docs/workflow/tasks/pelican-layout-choice-20261103.md`
- contract_sha256: `ffd1be85520e6cc68e78cd016e55735e7f0b1129b6aab9907558d9a6aea663b0`
- base_commit: `582ad5ab092441068c61c4f96eec148d8d022106`
- reviewer_role: Independent Contract Reviewer
- last_verified: `2026-11-03`
- review_round: `R2`
- 哈希针对磁盘原始字节，包含 `status: approved` / `review_verdict: PASS` 元数据。本次独立 PASS 仅对上述精确字节生效；合同任何字节变化均须重新审查。

## R1 Record

- R1 verdict: `FAIL`
- R1 contract_sha256: `605fe4f769f33db9dcf53fc6f15e1b5fcc062b5690c17a6a7563a40541cebcf7`
- R1 原字节报告：`E:/codex-runtime/sub2api/pelican-layout-choice-20261103/contract-review-r1.md`
- R1 report_sha256: `8215d8ee026f1b7b60d73f7c51804753ebea9ca57ba2820f23219adfe08722da`
- R2 更新前，指定 review 报告与该外部 R1 存档 SHA256 完全相同；R1 FAIL 和原始发现已保留，不以本轮 PASS 覆盖存档。

## Findings

未发现须进一步修正的合同问题。R1 两项均已关闭：

1. **构建证据目录隔离：已修正。** Developer、独立 QA 和 fresh main 分别使用 `developer-dist`、`qa-dist`、`main-dist`；重测明确要求新版本目录（如 `qa-dist-r2`），保留旧构建和日志。阶段之间及重测之间不再共用清空目录。
2. **真实 diff 路径与工作目录：已修正。** 命令明确在相应 candidate/main 仓库根执行，逐项列出四个 `frontend/src/...` 真实路径；不再使用占位 pathspec。

## Executed Checks

### R2 本轮执行

- 完整重读最终合同、当前 review 和 Agent Matrix；只读核验 R1 存档哈希及当前 HEAD，HEAD 仍与合同 base_commit 一致。
- 将本轮合同 Acceptance Commands 段仅在内存还原为 R1 原文，按原换行编码为 UTF-8 后计算 SHA256，得到 R1 合同的 `605fe4f769f33db9dcf53fc6f15e1b5fcc062b5690c17a6a7563a40541cebcf7`。由此核验本轮仅 Acceptance Commands 调整，其他内容及批准元数据未变。
- 核对两项命令修正及各阶段/重测输出目录约束，R1 必须修正项全部关闭。
- scope、可验收性、dirty 保留、单份 DOM、浏览器 cleanup、现有测试保护和四文件边界沿用下列 R1 静态审查，未增加要求。
- 本轮未运行任何产品检查、浏览器或服务；未修改合同、业务、测试或其他 docs。

### R1 已执行检查（历史证据，非本轮重跑）

- 只读读取 `AGENTS.md`、workflow status、Agent Matrix、spec 本任务段、完整合同，以及 P/G/E 和 review-and-verification 技能及合同审查模板。
- `git rev-parse HEAD` 与合同 base_commit 一致；`git diff --cached --quiet` 退出 0，当前索引无已暂存差异。`git status --short` 显示大量并行 dirty/untracked；本轮未操作这些文件。
- `Get-FileHash ... -Algorithm SHA256` 两次读取 R1 合同得到同一 R1 哈希，包含批准元数据。
- 只读检查四个允许文件，以及 `usePelicanArtwork.ts`、其现有测试声明、`package.json`、Vite/Vitest 配置和历史 `6d3f34e9d` 视图。四个 `.cmd` 验收工具入口在当前 main 存在。
- 使用明确的仓库根相对四路径执行 `git diff --check`，退出 0；这仅证明当前 dirty 输入无该检查所报告的空白错误，不代表未来实现通过。
- 当前四路径中 view/test 已有 dirty 差异，另外两个未显示差异。合同要求冻结当前字节、隔离开发、停止写入后冻结、独立 QA、集成前检查漂移并保护非目标原始字节，符合保留既有修改的边界。主控并行制作的外部快照/candidate 不属于本轮已验证成果。
- scope：四个产品/测试路径足以承载个人布局偏好、共享模板及预览调度；服务端设置、依赖、sandbox/sanitizer 和其他界面均明确禁止修改。spec 本任务段与合同目标一致。
- 可验收性：默认纵向、有效偏好恢复、非法/受阻存储降级、切换不请求 list/settings、多轮身份及过滤/放大/重播保留均有明确结果要求；浏览器桌面/移动、明暗和确定性滚动突发指标有独立验收要求，不以 fixture 宣称生产 FPS。
- 单份 DOM：第 45–52 行禁止同时挂载两个隐藏画廊，要求保留当前多轮 keys、页面级 artwork cache 并清理旧消费者/重新观察布局；未要求还原历史单轮数据或历史对话框。
- 浏览器 cleanup：第 97–98 行要求专用 session/profile、PID、关闭和所属浏览器/daemon 归零；结合 AGENTS.md，失败/超时同样必须清理，仅处理本任务精确 PID，不得关闭其他会话。
- 现有测试：view 已覆盖多轮稳定节点、旧轮精确放大、模型/状态过滤、running 独立重播、异步横向导航、分页/自动刷新和权限变化；LazyArtwork 已覆盖 240px 预取/实际可见 iframe、隐藏文档、失败重试、权限失效；store 测试覆盖三并发、取消传输仍占槽、去重/身份隔离、60 条/6MiB LRU 和暂停。合同要求在允许测试文件内新增布局/存储/帧合并/卸载覆盖，保留现有语义断言。
- 当前横向专属测试及导航测试需显式选择横向模式后继续验证，不能因改为默认纵向而删除或弱化多轮、权限和预览断言；这属于已授权测试适配，无需扩大文件范围。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- dirty_preservation_defined: `yes`
- single_gallery_dom_required: `yes`
- browser_cleanup_required: `yes`
- existing_tests_reviewed: `yes`
- acceptance_commands_executable: `yes`，命令/路径已静态核验，未执行产品检查
- worker_model_configuration_matches_matrix: `yes`，Developer / QA 为 `gpt-6.1-sol`
- base_commit_confirmed: `yes`
- openspec_traceable: `not-applicable`

## Unverified Risks

- R2 仅审最终合同及 R1 存档；R1 代码检查为历史静态证据。未执行测试、typecheck、lint、build、浏览器验收，未启动服务或子 agent，也未检测/调度模型可用性。
- 主控报告 candidate 和 12376 主文件基线已建好，本轮未核验其内容；快照、冻结及保护脚本仍须在后续门禁实证检查，本轮不为这些产物背书。
- 当前同步调用几何更新的 LazyArtwork 测试在引入 requestAnimationFrame 后需要确定性推进帧；不得用放宽可见性/权限断言消除失败。
- 本报告不是实现 QA、集成许可、提交许可或部署许可。

## Recommendation

`PASS`。合同门禁通过，绑定本报告的 R2 原始字节 SHA256。主控可按 Agent Matrix 推进后续已授权流程；独立 QA、精确冻结/集成保护及最终验收仍为后续门禁。本 Reviewer 不调度 Developer，不验证实现，不扩大授权范围。

本轮唯一写入为本报告；未修改合同、业务、测试、其他 docs，未提交。
