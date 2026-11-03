### PASS: upstream-gap-fixes-20261103-precommit-recheck

2026-11-03，独立只读 Reviewer。审查候选 `E:/codex-worktrees/sub2api/upstream-gap-fixes-20261103`，HEAD `7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`。R3 freeze：`E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-candidate-freeze.json`，SHA256 **`0f7078d1f7a9f800d42c1b065c2f549eddf579941a31b870bf3cef7d07c36fd1`**。结论限于该冻结候选；主目录保护性合回、新鲜检查及精确本地提交由 Controller 执行。

## Findings

- 未发现新的阻断问题，最初 **P1/P2 均已闭环**。
- **P1**：`ExtendSubscription` 仅在本方法识别 outerTx；锁读、expiry/status 写入及返回读取使用同一调用方事务，不另开、提交或回滚它。真实 ent commit hook 在下层成功返回后才使 **L1 和 billing subscription 两类缓存失效**；commit 前、rollback 和 commit error 后均保留缓存。无 outerTx 的原 helper、锁读算写、并发累加及错误/边界行为保留。
- **P2**：覆盖目标旧状态前判断响应状态是否实际变化。同 Key 服务端状态不变时保留未保存的 active/inactive 选择，真实表单随后仍提交该选择；实际 quota_exhausted→active 仍同步，响应 quota_used/status 始终用于捕获目标，切换/清空选择不污染其他表单。
- 四路径 R2→R3 diff 已逐项审查，变化对应上述修复及新增回归。原 Go 测试/helper 正文与 Vue 原八例逐字保留，未删断言、降低预期或新增 skip/todo；另外七源严格等于 R2。

## Executed Checks

**本 Reviewer 实际执行**（仅定向检查，无广域重测）：

| 检查 | 结果 |
| --- | --- |
| 原 `caller_transaction_probe_v2_test.go` 原字节 overlay 重放 | default/unit 均通过；begins=1、服务内 commits=0、writes_use_outer=true；outer rollback 后到期日不变 |
| 独立 outer 生命周期 probe | active/expired × success/rollback/commit_failure 六例；真实 ent Tx/commit hooks、生产 ExtendSubscription/BillingCacheService；两类缓存提交前保留、成功提交后失效、回滚/提交失败后保留 |
| 上述两个 Go probe 合计 | default **7/7**、unit **7/7**，fail=0、skip=0；固定 Go1.26.10、`-mod=readonly -count=1 -json`、离线依赖和私有缓存 |
| 原 Key handler probe，仅适配副本的输入/输出目录 | 候选三例通过；保留 HEAD/upstream 对照，原件未改 |
| 独立实际 mounted Vue/jsdom | **9/9**，fail=0、skip=0；含原七例、两个未保存选择反例及真实 `#key-form` submit；Vitest `--no-cache`，私有配置/缓存，未安装依赖 |
| 身份/保护 | 运行前后及封存时 11/11 候选源匹配 R3；主树/候选 HEAD 固定、暂存区为空；主树逻辑 index 匹配 Controller 起点；依赖/lock 各自起点未变，Go 主树/候选仅既有 CRLF 差异 |

**复用但本轮已核实际日志/哈希/源绑定的 Developer 证据**：独立重算必要非缓存证据 **83 文件**，核 **19 份 receipt** 的 argv、exit、stdout/stderr SHA、源 SHA，并从日志重新统计 leaf、fail/skip。原 248 声明回归 default620/unit671、GapFix66/标签、原 GapFix/repo count=3 default198/unit261、quotaReset10＋相邻 KeysView9、typecheck/定向 lint/backend build/gofmt/diff-check 均符合记录。RED 原件保留：原 caller 反例失败、新 outer 七例失败、Vue 原八例通过而新增两例失败；R3 对应 GREEN。未把这些复用检查称为 Reviewer 重跑。Controller 全部7351文件重算另见 `precommit-controller/developer-acceptance.json`，本 Reviewer 不冒称重算了7351文件。

本轮实际证据目录：`E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-qa/`；执行子目录 `run-2026-11-03T05-44-55-201Z/`。

| 证据（相对上述目录） | SHA256 |
| --- | --- |
| `recheck-evidence.json`，绑定实际 receipt、原始日志、私有脚本、四 diff、源身份及保留原件 | `2411d234fc147dbd46196f6241f28da6fe0477a68d418af9ad94bced483afaef` |
| `frozen-input-review.json`，83文件/19 receipt 及原测试正文比对 | `67e1fa924f74b4848d78e2c1066e3c8d40028621e612718c902853e660c4eec9` |
| `run-2026-11-03T05-44-55-201Z/ownership-lifecycle-default.stdout.log` | `ef560091c3c55340536fdef0b18df841fc67c2955ba26c1a3a2acfbf32f9a3b7` |
| `run-2026-11-03T05-44-55-201Z/ownership-lifecycle-unit.stdout.log` | `ef47652657d9579bda82c99aad7f758e7be7eaa877a56e71ad94ec3331d31326` |
| `run-2026-11-03T05-44-55-201Z/key-form-status-comparison.json` | `15e652c300c32033804b648f0e70e18f90b84cc987a109854ed72f221a10acbb` |
| `run-2026-11-03T05-44-55-201Z/key-independent-mounted.vitest.json` | `b63723054b78b0c31061adec7f339749131878dbd8ea9def8edd8a160d615ff0` |
| `controller-status-r2-binding.json` | `66e175c80e5d13984366bd51e1eaddbc03389657bf3f07d0c084fa0599db636d` |

R3 绑定十一源完整 SHA256：

| 路径 | SHA256 |
| --- | --- |
| `backend/internal/repository/usage_billing_repo.go` | `4a01690416403c25713e02d638bce8897c02ff6160071c1d895357c8c215b111` |
| `backend/internal/repository/usage_billing_deleted_account_test.go` | `595ea8de66352ce633b77ec7c42ab0f24a954125c3af3e4e633453722725de8e` |
| `backend/internal/service/openai_images.go` | `384b054c35991ad90ae7acaaf34f118681d8339e3158626bb93f881251768d93` |
| `backend/internal/service/subscription_service.go` | `020be45e7d9305a25f6f6483de59426cf3ed18a0917f043fc42ea570b921d59f` |
| `backend/internal/service/openai_gateway_service.go` | `1d325b359d3931d8e8b30d41eb76354fc1896f8b6276b911bd5bbad3f7fe20d9` |
| `backend/internal/service/account_stats_pricing.go`（绑定，未变） | `4eabb069d8a736ce825892c9e69b2d499f63d0e40ca59d3ff97230ec4d8df4d4` |
| `backend/internal/service/upstream_gap_fixes_test.go` | `8c635fa38f03093e05895c001898c3bf27ffe199ad5a74fc821c883e1dfc2898` |
| `frontend/src/views/user/KeysView.vue` | `3b9bcfbb8286099ae0df795ec3ed31d7eaeaea47b4174d39f9cb7c610c28ef6a` |
| `frontend/src/views/user/__tests__/KeysView.quotaReset.spec.ts` | `c2b34b9b528c61c54b681fc310350e3095790ae052a7b3de320c29d6c37c09f9` |
| `backend/internal/service/openai_account_scheduler_test.go` | `b6fa14d686afe4702a5fcb5365b1903d54774633b255a4d7a8cf1004b9a9ca5c` |
| `backend/internal/service/openai_subscription_long_context_test.go` | `2302fb4c3c39d166c59f554822e633e163de3602774de8db0964b92461e1e51c` |

共享保护例外已单独核实：`status.md` 的 group-fast 段 contract-draft→build→qa。最新观察 SHA256 `1bccbec6959de32ed392cd8c35f6edcdfe53c70082142d9726d7e59a2885259b`；仅在内存换回初始精确 CRLF 段，整文件 SHA 复原 `e8d85b66c6edc8fbfe366b328f90914c1250d0086ea4bbf4b377e811b065fb2d`。Controller `precommit-controller/status-reconciliation-r2.json` SHA `5733f239fa41638638f5fd5fed7f5d842e9dc664d8bf9139729c3921abd3ecfc`，同名 `-observed.md` 与独立观察逐字一致。保留最新共享段且不暂存；不声称12,357原文件字节全程恒定。首次严格 status guard 及原 seal 脚本失败留存；后续按精确段重构核验，未放宽其他字节。

## Unverified Risks

原四项基线失败继续保留，本轮未修复或重跑，不能称全套绿色：

- `TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost`
- `TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy`
- `TestOpenAIQuotaServiceNotConfigured`（nil panic）
- `KeysView create query > does not overwrite a touched default group when rates load after the dialog opens`

无真实 PostgreSQL/provider/浏览器/负载验收；事务测试使用真实 ent hooks 配合 driver/SQLmock，mounted 为 jsdom。第七项图片空价继承仍暂缓。本报告不替代 Controller 的主目录合回及 fresh-main 检查。

## Recommendation

**可继续保护性合回；Controller 完成其既定新鲜主目录检查及精确路径保护后，可按用户后续授权执行本地提交。**本 Reviewer 未修改产品/测试/合同/旧报告、未 stage/commit/push/deploy。旧合同/QA报告中的“未提交”是当时阶段快照，本轮后续授权本地提交不改变其历史事实。

原 FAIL 报告 SHA256 `9c9f2330334abd6f916da5c0100ad42621127a63ed350652d9ed80267dc7d8a3`、补充合同 `64c85c29a36f745dc005efb0a0fd75689943fe399507306635ec891e9a14cdfb`、合同审报告 `58fb6d4e4e9b515c541f198857916c1e35730f5a75f0a6700287b32221b4949c` 均保持原字节；原失败探针/日志保留。最终提交继续排除共享 dirty 文档、Pelican、凭据及其他非目标文件。
