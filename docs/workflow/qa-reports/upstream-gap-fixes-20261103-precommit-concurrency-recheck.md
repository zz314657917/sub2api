### PASS: upstream-gap-fixes-20261103-precommit-concurrency-recheck

日期：2026-11-03。范围：最终发布/验证脚本的静态门禁复核；不重新审查产品实现。

## Findings

无新增阻断。原 concurrency FAIL 指出的验证结果与 staged tree/snapshot 缺少绑定的问题，在本次最终脚本中已关闭：

- `verify-repair.mjs` 的 `controller-exact-staged-r3` 分支要求实际 cwd 等于 snapshot 路径、snapshot.tree 等于 staged.tree；从该 tree 读取全部 blob/path/mode，测试前后逐项计算隔离快照的 Git blob SHA1，与 tree 原值核对，并保存完整清单的数量/摘要及两份输入记录 SHA256。
- `publish.mjs` 的 commit 门禁拒绝缺失或不匹配的 exactBinding；核对 tree、snapshot、报告 cwd、两份记录 SHA256、全 tree 前后摘要，以及每条 receipt 的 cwd/exit；提交前再次核对当前 index tree 和父提交。
- stage 要求新 PASS 报告包含最终观察及两脚本的实际完整 SHA256。路径集合严格为原 22 + 原 concurrency FAIL + 本报告，共 24；gateway 使用冻结 R3 blob 写入 index，保留主树并行版本。原工作树保护和提交后的父/tree/路径/字节核对仍在。

## Executed Checks

实际执行：只读重算以下 SHA256，读取两脚本并静态复核上述分支；确认原清单 22 路径、HEAD 为 `7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`、当前暂存区为空。所有只读 Git 命令设置 `GIT_OPTIONAL_LOCKS=0`。

- `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-controller/concurrent-observation-r3.json`：`0f1f435cc33a1443849fcfdb4cfdcd75149f5224b334fe79654d95b6f338bfc1`
- `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-controller/publish.mjs`：`61868d3d4c207faa0188e7eb788928124bceb0c875d3b49ae3234db17d1d2296`
- `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-controller/verify-repair.mjs`：`9cab52374cb304d57d36e3bb41b85b373121e529bde2c674b63d61864415240f`

以上三项均与用户指定最终字节一致。原 concurrency FAIL 报告 SHA256 仍为 `a64301fcc146a65a0859e0a230682ce29d6787427ef34e846368ccd9945ef59b`，未修改。

复用此前 R3 观察与产品复核背景；本轮未重新验证 12,369 路径、产品反例或既有测试。实际测试数为 0，未执行两脚本，未暂存、导出快照或提交。

## Unverified Risks

当前尚无本轮实际 staged tree/snapshot 验证结果。本 PASS 仅确认静态门禁修正充分，不能作为实际暂存树已经测试通过或提交完成的证明。Controller 仍须执行精确快照验证并通过 commit 门禁。

既有产品 P1/P2 复核结论不变；旧 4 项基线失败及主目录 source-drift FAIL 证据保留，不称全套绿色。未增加 live DB/provider/browser 验证。旧合同/QA 报告中的“未提交”是当时阶段快照；用户后续本地提交授权不改写历史报告。

## Recommendation

静态门禁评审 PASS，允许 Controller 按上述最终字节继续精确 24 路径暂存、隔离快照验证；实际验证 PASS 且 tree 未变后方可本地提交。旧 FAIL 报告与本报告一并保留。
