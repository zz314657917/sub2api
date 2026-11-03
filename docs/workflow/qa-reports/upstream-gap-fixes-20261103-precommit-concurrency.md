### FAIL: upstream-gap-fixes-20261103-precommit-concurrency

2026-11-03，最终字节只读评审。**最终观察 PASS，发布脚本仍有一项阻断；原 R3 代码 PASS及P1/P2闭环保持。**

绑定最终输入：

- `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-controller/concurrent-observation-r3.json` SHA256 **`0f1f435cc33a1443849fcfdb4cfdcd75149f5224b334fe79654d95b6f338bfc1`**
- `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-controller/publish.mjs` SHA256 **`88ec18dad52d1c998ff4f8a124be99fa98c9c11dc4e688d495f559b4d3bf259d`**

## Findings

**[P2，发布阻断] commit门禁未绑定验证所对应的tree/snapshot。**`publish.mjs:92–98` 仍只检查 `checked.verdict === 'PASS'` 与十一源 `checked.files`。未读取 `staged-snapshot.json`，也未要求验证报告的tree、snapshot路径与本次导出记录一致。随后确认index tree等于staged.tree，只能证明暂存树未变，无法证明PASS验证的是该树；相同十一源、不同其余文件的验证结果仍可被接受。这是未来提交门禁的静态缺口，**并非要求当前执行尚未授权的staged-tree检查**。须在commit前强制绑定验证报告、staged-result与导出snapshot的同一tree及命令cwd身份，缺失或不匹配即停止。

此前status观察过时项已由最终R3观察关闭，不再作为当前阻断。

## Executed Checks

- 实际重算上述两个最终SHA；比对脚本确认相对首版仅观察文件名变化，commit门禁缺口仍在。
- 逐项比较最终与先前12,369路径观察记录，差异仅 `status.md`、`spec.md`；37产品及12新增文件记录不变。最终31项原路径变化为group-fast25＋本批4修正＋两个限定共享段。此前37内容/preimage、十一源、22预期路径核验复用，不重扫实现。
- 实际读取当前status/spec并核最终观察SHA：status仅换回初始精确CRLF段、spec仅删除明确新增四行后，各整文件SHA均匹配初始快照；没有放宽其他字节。
- 23路径方案范围可接受：gateway仅从冻结R3候选写Git blob/index，主树并行gateway及其余37路径保留；两个共享文档均不暂存。方案范围正确不替代上述验证身份门禁。
- HEAD实读 `7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`，暂存区为空，逻辑index与观察一致。未执行测试、publish模式、暂存、快照导出或提交。
- 最终核验记录 `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/precommit-repair-qa/concurrency-final-r3-review.json`，SHA256 `48ced80e54112d2a6a6c0afb8d8f9f16f5cd03d2ea3372f057a7c96449705019`。上一版并行评审FAIL已逐字保存为同目录 `precommit-concurrency-initial-fail.md`，SHA256 `6ce240b65d0939c22b1995303f3cb6ead2ee55c4558ea4fd17c6d8ce3526a589`；原代码PASS/FAIL、旧观察及失败证据不改。

## Unverified Risks

当前未stage，尚无实际staged tree或精确快照验证结论。原main source-drift FAIL、四项基线失败及未做live验收边界保持。

## Recommendation

**需修复发布脚本的验证身份门禁后再取得最终字节评审；当前不可据此提交。**无需重新开发或重测R3产品。后续仍由Controller执行精确暂存、绑定tree的快照验收及提交前后保护。
