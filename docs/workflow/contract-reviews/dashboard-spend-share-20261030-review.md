---
type: contract-review
scope: dashboard-spend-share
status: approved
task_id: dashboard-spend-share-20261030
verdict: PASS
base_commit: 8f202b73282465a9b4e9d5bcfeb1c267362e72c4
reviewer: independent-contract-reviewer
last_verified: 2026-10-30
reviewed_contract_sha256: 85ace2481c435d7fd95895a6cfcea23615be8844b3a293024d8f039ccaa04828
---

### PASS: dashboard-spend-share-20261030

## Findings

已核对修订后 Success Criteria 2：正常行逐笔以 total_cost 分摊；total<=0、缺失/负reference或sum>total整笔归other；严格舍入降级、图片/按次/视频剩余和subscription口径均已明确。先前唯一精化阻塞已解除，其它范围、入口和兼容策略未发现明确问题。

## Confirmed Algorithm / Implementation Guidance

下列为对已修订条款的执行口径说明，不增加业务范围：

> 对每笔 applied 且 actual_cost 为有限非负数的记录，以该笔存储的 total_cost 为唯一 reference 分母。四项 reference 为 input_cost/output_cost/cache_creation_cost/cache_read_cost。仅当 total_cost>0、四项均非 NULL/有限/非负且四项之和<=total_cost 时，分别计算 actual_cost*component/total_cost；other 取该笔 actual_cost 减去四类合计。total_cost<=0/NULL/非有限、任一 reference 无效/为负/缺失、四项之和>total_cost 时，四类全零，该笔 actual_cost 全部归 other。严格比较数据库 numeric，不设宽容阈值、不改分母、不按四类重归一；仅因 decimal(20,10) 存储舍入而超界的行也保守归 other。pending/failed 五项全零。不得以 max(total_cost,sum(components)) 修正异常行，也不得按 token 数量推断费用。actual_cost 自身非有限/负数属于异常数据，不能伪造成正常非负分类；遇到此情况按 Stop Rules 向 Planner 报告。

> SQL 用 NULLIF/CASE 避免零除；尽量保留 numeric 精度至聚合完成，只在扫描到 DTO 时转 float。前端以期间 actual_cost 合计作分母，五项完整有限非负且合计与该分母在明确小误差内一致才显示份额；建议一致性容差 max(1e-8, abs(totalActual)*1e-9)，只用于判断浮点表示误差，不用于改分母或重新分配实际金额。任何桶缺少/部分缺少 nested 明细时显示不可用，不对有明细子集归一；所有桶有效且实际消费为0时显示0%。显示舍入只影响文案，不回写计算值；微小正份额可显示“<0.1%”。

该政策有意选择严格保守的 Other 降级；它不会将不一致数据包装成确定分类。前端精度容差值为建议，可选其它有测试依据的小容差，但不得掩盖金额不一致。这里的分类金额是依据已保存 reference 权重对已结算实际扣费的分摊，并不是现存的逐类独立扣款账本。

## Confirmed Semantics / Evidence

- `backend/internal/repository/usage_log_repo.go:111-122`：用户 actual_cost 只累计 applied；pending/failed 的 measured actual_cost 不等于实际已扣款。
- `backend/internal/service/billing_service.go:1104-1170`：图片输入/输出 token 已从文本输入/输出成本拆出，TotalCost 包含图片与四类成本。图片费用必须留在 other；不得重加图片token或把图片费用并入文本类。
- `billing_service.go:1217-1248`、`:1597-1624`：按次、按图及其它无四类明细费用落 other。
- `backend/internal/service/gateway_service.go:10192-10199`、`openai_gateway_service.go:8421-8428`：分类标准成本与 actual_cost 已独立存储，读取扩展无需schema/计费写入变更。
- 沿用 existing actual_cost 口径，包含 subscription 额度扣减。不得擅自改成仅余额支出或修改 subscription 结算语义。
- `usage_log_repo.go:4879` 与 `:5173` 两条用户 raw 入口均须扩展；后一条只在 userID>0 的用户范围返回明细。管理员 raw/预聚合维持原9列扫描和原统计语义；`scanTrendRows` 仍能服务管理员路径。
- `backend/internal/pkg/usagestats/usage_log_types.go:92` 可选指针 nested + omitempty 合适；现有service直接传递DTO，snapshot/trend不应另造一套分类计算。
- 当前 `UserDashboardCharts.vue:192-203` 的 tokenMax 归一不可保留为消费比例算法；保留token计数本身无问题。
- `backend/ent/schema/usage_log.go:91-108` 成本字段为 decimal(20,10)，支持上述严格numeric策略。

## Gate Checks

- success_criteria_testable: yes。
- allowed_paths_explicit: yes。
- denied_paths_explicit: yes。
- acceptance_commands_executable: yes，package.json 有相应test/build命令；独立QA仍需实跑，当前评审不声称测试通过。
- worker_model_confirmed: yes，Agent Matrix/contract均为gpt-6.1-sol。
- base_commit_confirmed: yes，本轮 git rev-parse HEAD 与 git cat-file -t 确认该hash存在且匹配当前HEAD。
- openspec_traceable: not-applicable。
- dirty_preservation: 已确认baseline-files.json存在；最终集成仍须重新核对SHA。

## Approval

独立contract review PASS，绑定上述SHA256的修订版draft。Planner可以仅更新approval frontmatter后推进contract-approved；任何业务范围或算法变化必须重新审阅。该PASS仅覆盖contract，不替代Developer构建、独立QA或最终集成验收。

只修改本review文档，未改业务文件、contract、计费/schema或其它workflow；未联网、未访问真实DB/provider。
