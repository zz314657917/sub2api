### PASS: upstream-gap-fixes-20261103-precommit-repair-developer

Developer DONE。四个产品/测试文件已停止修改；本结果写入后全部停止写入，交 Controller 冻结和独立 Reviewer 验收。

## Findings

- P1 已修复：仅在 ExtendSubscription 内识别 outerTx，锁读/计算/expiry/status 写/最终读全沿用它，不新开或结束调用方事务；OnCommit 下层成功后才失效 L1 和 billing。无 outer 的原 helper、行锁、原子性、+5/+5 累计 +10、回滚与 nil-client 兼容保持。
- P2 已修复：始终消费响应 quota_used/status，覆盖目标前比较旧 status；仅实际状态改变且仍选中同 ID 才同步表单，两方向未保存选择及最终 handleSubmit 参数均通过。
- 新增 TestGapFixAdminExtendCallerTransaction：success、rollback、commit_error、read/expiry/status/final_read_error 共 7 场景；SQLmock 只允许 owner 的单 begin 与后续 commit/rollback，校验全部读写 outer 身份及两缓存提交前/成功后/失败后时序。另增 standalone billing success/rollback/commit_error 3 场景。旧 Go 正文与 Vue 8 项均逐字还原匹配 R2，未改旧断言。

## Executed Checks

- 写前合同 SHA256：64c85c29a36f745dc005efb0a0fd75689943fe399507306635ec891e9a14cdfb，与独立 PASS 绑定一致。R2 freeze SHA256：d5c30bd3e62d46e983db13614c901d6aab049862d486f4a279facef9090b463b，初始 11/11 源匹配。
- 候选根 AGENTS.md 不存在；读取主树等效规则并按用户限定执行，未新增 AGENTS。
- RED 在实现前保留：Vue 原 8 通过、新 2 失败；新增 outer 7 子场景失败。原 Reviewer v2 probe 仅复制后运行，RED exit 1（begins=2 / commits=1 / writes_use_outer=false，rollback 扣减残留），GREEN exit 0（begins=1 / commits=0 / writes_use_outer=true，rollback 保留到期日）。原件未改。
- 原 248 声明回归 default 620、unit 671 leaf 全过，无 skip；原 GapFix+repo 正则 default/unit 各 count=3 全过。quotaReset 10/10、相邻 mounted KeysView 9/9；typecheck/lint/build/gofmt/diff-check 通过。gofmt -l 输出为空。

| 实际检查 | exit | 结果 |
| --- | ---: | --- |
| green-backend-build | 0 | 见同名 receipt/stdout/stderr |
| green-copied-probe-unit | 0 | 1 leaf pass / 0 fail / 0 skip |
| green-diff-check | 0 | 见同名 receipt/stdout/stderr |
| green-existing-default | 0 | 620 leaf pass / 0 fail / 0 skip |
| green-existing-unit | 0 | 671 leaf pass / 0 fail / 0 skip |
| green-gap-default | 0 | 66 leaf pass / 0 fail / 0 skip |
| green-gap-unit | 0 | 66 leaf pass / 0 fail / 0 skip |
| green-gofmt | 0 | 见同名 receipt/stdout/stderr |
| green-lint | 0 | 见同名 receipt/stdout/stderr |
| green-original-gap-repo-default | 0 | 198 leaf pass / 0 fail / 0 skip |
| green-original-gap-repo-unit | 0 | 261 leaf pass / 0 fail / 0 skip |
| green-typecheck | 0 | 见同名 receipt/stdout/stderr |
| green-vue | 0 | quotaReset 10 + KeysView 9 pass |
| red-copied-probe-unit | 1 | 0 leaf pass / 1 fail / 0 skip |
| red-new-default | 1 | 0 leaf pass / 7 fail / 0 skip |
| red-vue | 1 | 8 pass / 2 fail |
| test-gofmt | 0 | 见同名 receipt/stdout/stderr |

完整实际 argv/cwd/env/PID/起止时间/十一源 SHA/日志 SHA 见 runtime 同名 *.receipt.json。命令入口为：

- D:/nobe/node.exe E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-developer/run-checks.mjs red
- D:/nobe/node.exe E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-developer/run-checks.mjs green
- D:/nobe/node.exe E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-developer/finish-checks.mjs
- Go test/build 全部 -mod=readonly，固定 E:/tools/go1.26.10/go/bin/go.exe；process GOROOT=E:/tools/go1.26.10/go、GOTOOLCHAIN=local、GOPROXY=off、GOSUMDB=off，GOCACHE/GOTMPDIR/TEMP/TMP 仅本轮 runtime。
- Vitest run --no-cache --config E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-developer/vitest.config.mjs --reporter=json --outputFile=<本轮runtime>；vue-tsc --noEmit；定向 eslint 无 --fix 且使用指定 NODE_PATH。未安装依赖/改 lock/node_modules。
- 原正则读取 E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/existing-regression-regex.txt；GapFix+repo 复用原 Controller receipt 正则。两个 git diff --no-index 的 exit 1 是预期差异，非验证失败。

最终四源 SHA256：

| 路径 | SHA256 |
| --- | --- |
| backend/internal/service/subscription_service.go | 020be45e7d9305a25f6f6483de59426cf3ed18a0917f043fc42ea570b921d59f |
| backend/internal/service/upstream_gap_fixes_test.go | 8c635fa38f03093e05895c001898c3bf27ffe199ad5a74fc821c883e1dfc2898 |
| frontend/src/views/user/KeysView.vue | 3b9bcfbb8286099ae0df795ec3ed31d7eaeaea47b4174d39f9cb7c610c28ef6a |
| frontend/src/views/user/__tests__/KeysView.quotaReset.spec.ts | c2b34b9b528c61c54b681fc310350e3095790ae052a7b3de320c29d6c37c09f9 |

保护结果：七个非授权源与 R2 原 hash 一致；候选起点覆盖 3,881 文件，仅四源变化。Controller 主树 12,357 原文件核对仅其已声明并行 docs/workflow/status.md 变化，十一源仍同 R2。候选和主树 HEAD/index 原始字节均匹配本轮起点；不可变合同/原失败报告/原探针全部匹配。Controller 封印 R2 引用 27/27 哈希一致，仅复用其他七源旧证据。

证据目录：E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-developer/。verification.json 及 developer-final-source-hashes.json 为最终检查/十一源身份。首次 finalize.mjs 在最终保护核对 PASS 后因报告插值局部变量未定义 exit 1，未写结果文档；本脚本只读取已完成验证和源身份写简洁结果，不重跑测试。

## Unverified Risks

以下四项既有失败未修复、未重跑，不能称广域全绿：

- TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost
- TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy
- TestOpenAIQuotaServiceNotConfigured（nil panic）
- KeysView create query > does not overwrite a touched default group when rates load after the dialog opens

真实 PostgreSQL/provider/浏览器/负载未验收；测试为真实 ent hooks 配合 SQLmock/driver，以及 mounted Vue/jsdom。未启动其他 agent、应用或容器，未 stage/commit/push/deploy。

## Recommendation

仅 Developer 本地 PASS，独立 Reviewer 尚待执行。产品/测试写入仅四个表中路径；唯一新增仓库文档为本文件。全部停止写入，Controller 可冻结并独立复查。
