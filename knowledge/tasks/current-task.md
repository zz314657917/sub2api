# 当前任务快照

最后更新：2026-08-16 00:18 +08:00

## 背景

- 用户要求持续核实并选择性合入最新上游版本；本地主线长期分叉，禁止整包合并。
- 上游当前为 `upstream/main@baeac1f3d`，最新 tag 为 `v0.1.177@073e92d17`。
- 主工作区存在用户未提交的账户编辑前端源文件、对应测试和 `outputs/`，所有上游 Sprint 均排除它们。

## 当前目标

- S218 行为级移植上游 remote compaction v2 三提交：`9662cff2e`、`a8b9ea22b`、`8ae6d8f67`。
- 原生 streaming `compaction_trigger` 保持 `/responses`，补齐会话 beta 头并把账号 compact probe 改为原生 v2；保留本地 legacy compact、调度、计费、failover、WS 和身份语义。

## 本次已完成

- S217 已由独立 Terra Developer 实现并由不同角色的 Terra QA 验收通过，最终合入 `main@56d86521b`。
- S217 覆盖个人订阅到期日、OpenAI HTML 403 保护、reset-credit API/UI 一致性和显式 POST quota refresh。
- 首轮 QA 发现 route contract 位于 unit-tag 文件而默认命令未执行；QA-1 `e8fea2d57` 移到 default-tag 后，默认/unit 重测均通过。
- 已 fetch 并确认 v0.1.177 未继续前进；remote compaction v2 三提交直接 apply 均因本地单文件网关拓扑分叉失败。
- S218 contract 已复审批准：修正 Developer build-base diff、独立 QA 写入边界、SSE probe 头、provenance 和 default-tag channel restriction 回归。

## 已确认事实

- 本地当前会把所有 bare `/responses` 中的 `compaction_trigger` 提升为 `/responses/compact`；这会把原生 v2 发往已下线的旧端点。
- 本地 compact account probe 仍调用 `/responses/compact`，并仅以 HTTP 2xx 判定成功，缺少实际 compaction item 校验。
- v0.1.177 的 turn-state (`8219dcfc8`/`4d9fedee2`)、指纹 opt-in (`fce41e318`) 与分组日汇总 migration 222/223 不属于 S218。
- S217 完整 service/server、server compile、21 个前端聚焦测试通过；S217 与主线均只剩同一既有 Airwallex TS2307。
- 用户两处前端 dirty patch 的 patch-id 在 S217 fast-forward 前后相同；未 push、部署、更新容器、调用 provider 或触碰 migration。

## 待验证点

- Terra Developer 实现 -> 验证：原生 v2 不改 path/body、不吃 legacy compact mapping；Responses-only selection 和 probe item 判定可重复通过。
- 独立 Terra QA -> 验证：完整 service/handler/server、server compile、allowlist/diff/index/provenance 通过，legacy compact 回归无退化。

## 当前结论

- `PASS / S217 local-main-integrated`：额度修复已合入本地主线，尚未推送。
- `S218 / contract-approved`：remote compaction v2 缺口真实存在，合同门禁通过，下一合法动作是创建隔离 worktree 并调度 Terra Developer。

## 下一步

1. 创建 `E:/codex-worktrees/sub2api/s218-remote-compaction-v2` 并调用独立 Terra Developer -> 验证：基线为 contract approval commit，只改 allowlist 且提交标准 worker result。
2. 主控审 diff 并补跑关键门禁 -> 验证：default-tag 测试可发现，native/legacy 语义和 no-provider 边界满足 contract。
3. 独立 Terra QA 后决定是否合入 main -> 验证：QA 首行 PASS、主工作区用户改动 patch-id 不变。
4. S218 收口后再评估 S219 turn-state；migration-heavy 分组日汇总必须单独授权数据库影响。

## 验证记录

- S217 QA：`docs/workflow/qa-reports/upstream-v0176-gpt-quota-s217-qa.md`，首行 PASS。
- S217 后端：focused `-count=10`、完整 `internal/service` 63.288s、完整 `internal/server`、server compile PASS。
- S217 前端：`OpenAIQuotaResetCell.spec.ts` + `AccountUsageCell.spec.ts`，2 files / 21 tests PASS；type/build 与 main 同为单一 Airwallex TS2307。
- 上游：`git fetch upstream main --tags --prune` 后 `upstream/main=baeac1f3d`、`v0.1.177=073e92d17`。
- S218 三个源提交逐个 `git apply --check` 均失败，原因是上游拆分文件与本地 monolithic gateway 不匹配。
