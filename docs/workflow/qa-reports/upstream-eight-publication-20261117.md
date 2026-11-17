# 上游八项适配：独立自检与发布验收（2026-11-17）

### PASS: upstream-eight-publication-20261117

## Findings

独立智能体首次审查发现账号额度竞态：保留表格行组件后，旧被动/主动请求的返回与 finally 可覆盖新额度或提前结束加载。已修复：共用请求序号与账号/API范围归属检查，保护结果、缓存、错误和 finally；捕获请求开始时的账号，防止排队后读取变化的 props；主动查询防重入，卸载后的返回忽略。

新增10个mounted回归场景；旧实现9失败/23通过，修复后32个UsageCell测试全部通过。同一独立智能体复核修复后PASS。其余七项未发现本批阻塞问题。

本批包含原A1-A6协议修复、B1账号刷新、B2 Key独立用量加载与上述B1竞态修复。无数据库迁移、计费、新模型、依赖升级或部署变更。

## Executed Checks

发布候选从干净HEAD `a1a1982dede6c3d7b41b6ff95e8a2fef4507fd86` 建立，只应用本批增量。未将同文件其他任务的dirty内容打包。

- 完整apicompat：299通过；新增handler/service default及unit各41通过；相关Go回归192通过。数字含父/子测试事件，重复模式不累计。
- Vue：50通过（UsageCell32、Keys17、Accounts1）；typecheck、限定6文件ESLint、前端build通过。
- Go `build -mod=readonly ./...` 通过；限定diff检查通过。
- 新竞态回归原版本red / 修复green真实运行。之前本地集成的测试数字较多，因包含其他dirty任务用例；此次仅验证实际提交树，没有删除既有断言。
- 智能体实际阅读17路径及生产调用链，并执行受控Promise竞态探针；完整构建/测试由主控执行。不能把智能体审查称为真实浏览器/provider验收。

## Frozen Scope

- 源码/测试17路径，加本报告，共18路径。
- 修复后17路径source tree：`e25a6e00c570e1b9463d778803190fec71b01592`（包含基线其他未变文件，不包含本报告）。
- 逐文件SHA、原始日志和独立结论：`E:/codex-runtime/sub2api/upstream-eight-20261116/publication-20261117`。
- 主工作区保留既有未提交内容，只有2个已审查竞态修复文件增量写入；提交使用干净发布候选blob，不对主目录文件整文件git add。
- 用户已明确授权“起个智能体自检一下 没问题 提交推送”。本报告记录提交前验收；提交与推送后的SHA及远端回读保存在上述目录的receipt.json。

## Unverified Risks

未做真实provider、生产数据库/Redis、浏览器或全仓所有测试。已验证本批相关回归、生产入口loopback及前后端构建；Browserslist和资源体积提示不作为失败，没有因此变更依赖。

## Recommendation

独立自检与主控发布验证通过，可按18路径提交并正常推送origin/main；不包含容器更新或部署。

## Independent Review

PASS：原 FAIL 已修复，本轮未发现新增阻塞问题。主控完整验证通过后，可将这两份修复暂存并随本批提交推送。

独立智能体 Lorentz（01a1e7f6-8358-7201-817f-3bb44a414608）复核：

- AccountUsageCell 的共享请求序号统一保护 active/passive 的成功、失败、缓存写入和 finally；旧请求不能再覆盖新额度或提前解除 loading。
- 独立实际执行发布候选 UsageCell mounted 测试：32/32通过。
- 对候选实际函数执行21组内存探针，全部通过：覆盖乱序成功/失败、缓存污染、旧finally、卸载、账号/scope变化、主动查询防重入及缓存命中接管。
- 两文件未暂存差异审查及 git diff --check 通过。
- 结论绑定产品 SHA256 d9c54721…b749213，测试 d0381cc5…d4c5fa1；完整SHA在发布manifest中记录。
- 未编辑源码、修改索引、提交或推送；修复复核未重复后端验证。探针不构成浏览器或真实provider验收。

首次复核为FAIL，真实复现旧请求覆盖新额度及旧finally提前清loading，要求修复后才放行；本PASS为修复后的第二次独立结论。
