# v0.2.15 定向提交复核（2026-10-12）

## 范围

- 用户同意整理并提交推送上一轮已验收适配；不合数字商店，不部署。
- 发布基线 `acaaf8b9c54519931cd2ba0c82d1da34ae6dfbe7`，保留此前输出 TPS 提交。
- 从上一轮 `integration-manifest.json` 的 64 个业务/测试路径提取 reviewed before→after 差异，不直接暂存混合脏文件。
- 一并保留 19 份原任务专属合同、审查、worker 和 QA 文档；不纳入共享 status/spec/main-log/agent-matrix 或混合知识文件。
- 必要前置修复一并保留：Responses 消息显式 `Type:message` 和两个 group modal 卸载监听器/定时器清理。原专项测试直接覆盖这些行为。
- UsageTable 及其测试与新 TPS 提交三方对齐，保留 TPS 展示及回归；未覆盖已有提交。

## 新执行验证

验证工作树 `E:/codex-worktrees/sub2api/upstream-v0215-publish-20261012`，由本次精确暂存补丁从上述基线生成。

- `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity ./internal/service/openai_ws_v2 -count=1` PASS。
- `go test -tags unit -overlay E:/codex-runtime/upstream-v0215-publish-20261012/fixture-overlay.json ./internal/handler ./internal/service -run '^TestV0215' -count=1` PASS。
- 后端生产构建 PASS，无生产 overlay。
- 原 28 文件相关前端回归 PASS：129 tests，包含新 TPS 既有回归。
- 精确暂存差异 `git diff --cached --check` PASS；仅提交 blob 规范化 CRLF→LF，不改主树文件。

## 边界

- 前端 full typecheck 再次失败：`AirwallexPaymentView.vue:103` 缺 `@airwallex/components-sdk`；不宣称完整前端构建通过。
- Handler/service 定向测试仍使用外部六个既有 fixture 编译修正；不纳入仓库，不修改生产代码。
- 未执行全库测试、浏览器视觉、真实渠道/数据库或部署验收。
- 数字商店留在 `codex/digital-store-isolated-20261010`；其他未验收业务改动及共享脏文档留主工作区。
- 精确补丁、暂存 scope、保护哈希位于 `E:/codex-runtime/upstream-v0215-publish-20261012`。

## 结论

本次精确提交快照的专项测试和后端构建通过，可按用户授权提交推送；上述环境与运行态未验证项仍保留。
