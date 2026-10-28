### PASS

# upstream-b2-retry-20261028 合同复审 R2

- 日期：2026-10-28。
- 合同：`docs/workflow/tasks/upstream-b2-retry-20261028.md`，仍为 draft/pending；Reviewer 未修改其状态。
- 最终 draft SHA256：`e2ed9f680ac554f3fa2368133c90a46f0c5ab66a7ebf58222c1e967b91b8aeba`。
- 本地基准：`b3091c1c9a28884cf4b00fe97e87fc2f295b0512` 加冻结 dirty 前置；上游：`c591412c42ab195dacdede9eb19f9cbfee221144`。

## 阻塞项复核

| R1 发现 | R2 结论及依据 |
| --- | --- |
| CR-B2-001：allowlist 缺口 | CLOSED。合同、工具及 targets-before 的15路径完全一致；新增五个实际 HTTP owner 已纳入。目标与原 main-before 一致，现存主树/候选目标字节未变。调用点抽查确认当前范围足够，无需新增路径。 |
| CR-B2-002：资格及副作用边界不明 | CLOSED。规范表固定账号、实际 upstream、HTTP状态、transport及原始权威字段冲突优先级；排除 WS/bridge/native Anthropic/CN。共享默认处理保留，HTTP wrapper显式opt-in，flag与副作用共用谓词，真实429/auth保持原处理。 |
| CR-B2-003：handler/usage验收歧义 | CLOSED。逐branch矩阵固定重试序列、slot/body释放、健康采样、Images partial不可重放；Responses保持nil,error，rawChat单独验partial。同步提交或drain后断言精确usage/charge次数及token。 |

未发现剩余合同阻塞。只放开proven-empty known事件；tool completion、工具公告、unknown/malformed与既有keepalive事件保持保守边界。NaN/Inf/overflow/日期/8s及header优先级仍须按合同取得测试证据。

## 已执行检查

- 只读复核修订合同、真实owner/共享调用点和验收工具；未开发、未重跑基线或测试。
- 15路径及原冻结快照比较、主树HEAD/index核对通过；固定selectors与已落盘baseline命令逐字一致。
- 工具SHA256：`f59b6622277d093ec8a620f9dec9345cf84d5e9ca123696d9fcc3698d0c55878`；QA scope使用自身qa-before，integration保留候选全树非目标保护。
- R1外部副本与覆盖前原review哈希一致，SHA256：`440a18988bd78df93077d70503e5e8af98bf74bfe3169c94fc0b80a3777fb7a3`；副本未改。

## 剩余风险与门禁

- 本PASS仅为合同门禁，尚无实现/独立QA通过证据。
- 原基线失败仍须独立QA对冻结前置精确复现；不得修改旧断言、改变selectors或新增豁免。新增测试须default/unit各三次，无skip/missing。
- 不扩展账号认证、WS、其他协议或15路径之外的实现；遇到必要越界须重新审合同。

结论：三项原阻塞全部关闭，主控可办理合同批准并按合同调度。本轮唯一写入为本review文件。
