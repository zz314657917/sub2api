### PASS: recent-selfcheck-20261030

本轮审查输入为 2026-10-27 至 2026-10-30 的 15 次近期提交，以及冻结工作区的未提交改动。重点审查网关流式计费与收尾、重试屏障、模型冷却与调度，以及 OAuth 非流式、缓存、渠道配置、管理员保护和近期前端逻辑。独立 Reviewer 发现两项可复现 P2；隔离修复、独立 QA 及真实主目录定向验证完成后，六个源码/测试文件精确合回本地。未提交、推送或部署。

Findings：

- Anthropic 转 Chat/Responses 已交付响应后收到 in-band error，原 helper 返回 nil + failover，导致 handler 提前退出并漏掉已观察 usage。修后在写错误前判断真实提交状态；优先使用 `Committed()`，暂存前奏仍可重试且不计费，已提交或断开保留 usage/cache 拆分并返回不可重放的原错误；无 usage 不制造计费结果，错误写失败正确标记断开。
- OpenAI 普通/透传流在输出前收到合法 bare error，原实现可能等到 EOF 丢失 overload/Retry-After 分类，或在后接 DONE 时误报成功。修后复用已有分类器与失败构造器立即收敛，保留原请求级重试预算、间隔和账号策略；输出后及断开禁止重放。
- 独立 QA R1 另拦下修复中的截断帧回归：含 overload code 的无效 JSON 被提前升级成 failover。before 两个 tag 的八个控制全过，候选每 tag 各三次失败。最终仅在两个即时重试条件前增加有效 JSON object 校验，保留原外置负面断言；完整重新冻结后独立验收通过。原分类器、输出屏障及 usage 解析函数保持原字节。

Executed Checks：

| 检查 | default | unit / 其他 |
| --- | ---: | ---: |
| 独立 QA：原 270/273 个适用声明 + 新六个声明 | 2184 PASS | 2189 PASS |
| 主目录 fresh：同一完整定向集合 | 2184 PASS | 2189 PASS |
| 独立外置实际 owner/首响应 writer 控制，155 个叶子各重复三次 | 465 叶子 PASS | 465 叶子 PASS |
| 独立 QA：新增六声明，各重复三次 | 786 PASS | 786 PASS |
| 独立扩展 stream/retry/nonstream/compact/first-response/Anthropic | 981 PASS + 一条既有 FAIL | 990 PASS + 同一既有 FAIL |
| backend build、六文件 gofmt | PASS | PASS |
| 前端原冻结快照定向测试 | 17 文件、201 测试 PASS | typecheck/build PASS |

Go 数值除明确标注“叶子”外为 JSON 测试终态 PASS 事件，包含父项/子项，不等于独立输入数。主目录逐声明核对 default 276、unit 279，missing/duplicate/fail/skip 均为空。三个仅 unit 适用的协议 ping 测试由 `go list` 的真实 TestGoFiles/XTestGoFiles 确认，unit 均执行；评分用例显式使用固定原始 fixture SHA `200a64a9625fdc75cbe1fc36b1369cb56d2ba5da55486d55902a23496d0cade7`。

Go 使用本地 1.26.10、固定 GOROOT、GOTOOLCHAIN=local、GOFLAGS=-mod=readonly、GOMAXPROCS=4。前端复用既有 node_modules，未重新安装 frozen-lock 依赖。另从纯 HEAD 源码执行 backend build 与 frontend typecheck，并检查 tracked 前端未依赖 untracked 源码；均通过。修复未改变前端源码，因此未重复前端整组。

旧断言迁移均先经独立 addendum 审查：普通 writer 已实际提交的 Anthropic connected/overload 分支改为保留已观察 usage；B2 同一流首个 generic temporary error 立即终结后，不再由后续 overload 覆盖其分类。原 payload、其它 kind/ping/safe、关闭及输出屏障断言保留；未通过笼统删除旧测试获得绿灯。

既有失败与验收边界：

`TestParseSSEUsage_SelectiveParsing` 仍在原测试第 3259 行报告 expected 9 / actual 1。before/candidate × default/unit 四份原日志一致，独立扩展也复现；测试全文及两解析方法字节不变。只读 Git show/blame 确认相关非 terminal merge 来自 2026-09-12 的 `4ad1d455e4fe5cd7b6e9473be39c03d9a9d66213`，早于本次审查范围。保留原 FAIL，未缩扩展选择、改该断言或把扩展组称为全绿。本报告结论限定本次两项修复及其边界无新增回归。

原始 Reviewer RED、Developer RED、首轮 QA FAIL、首次新正控中误把脱敏 ResponseBody 当原 payload 的夹具失败，以及工具路径/GOROOT/差异检查/冻结准备失败均保留；最终结论来自新的实际执行证据。冻结准备首次在发现 canonical QA 报告已经存在时于复制前停止，随后确认其与保存的 R1 FAIL 完全同字节；未因此修改产品或覆盖旧版本报告。

冻结与保护：

最终 freeze SHA `168e10a8439ad87544a3a126ab748f45c81955c8d30c7b5b11adf21900663779`；独立 QA seal SHA `9181ca55fafb05bca4c085f0f70117de0cd22a75fba9bdc696c54b859489d684`，主控核对其中 99 项证据摘要及完整 4874 文件 QA 集合。集成脚本经独立审查，持 Windows 目标写锁保护五个既有目标，以 CREATE_NEW 创建新增测试，使用已核验的同一份 sealed bytes 写入；原文件备份和恢复清单位于外置证据根。

主目录原冻结 4873 项输入中，仅批准的五个既有源码/测试发生任务变化，新增一个专属测试；另只插入本任务报告、current-task 快照和 status 并行条目。两份交接文档保留插入段之外的原始字节及混合换行。最终 HEAD `e6318739d05b85f235ab76638ccd32b174222c2c`、index 及其余受保护输入保持不变。任务 worker 已关闭。

Unverified Risks：

未访问真实 provider、生产数据库或执行实际扣费持久化验收；未启动部署容器、浏览器或 race instrumentation。近期下载 EXE、全部预览产物及仓库所有源码/测试未作全量审计；本报告不宣称全仓库、依赖重装或生产环境已通过。

Recommendation：

本轮本地自检修复完成；已确认的旧 SelectiveParsing 失败保留为后续独立处理项。未提交、推送、更新容器或部署。

证据根：`E:/codex-runtime/sub2api/recent-selfcheck-20261030/`。关键入口为 `reviewer/report.md`、`fix-approved-contract.md`、`fix-approved-test-addendum.md`、`fix/qa-report-r1.md`、`fix/qa-failure-seal-r1.json`、`fix/malformed-fix-plan-review.md`、`fix/freeze-r2.json`、`fix/freeze-r3.json`、`fix/qa-report.md`、`fix/qa-seal.json`、`fix/controller-r3-review.json`、`fix/integration.json`、`fix/main-final/summary.json`、`fix/final-audit/result.json`。
