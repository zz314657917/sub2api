# 当前任务快照

最后更新：2026-08-16 01:39 +08:00

## 背景

- 用户要求持续核实并选择性合入最新上游版本；本地主线长期分叉，禁止整包合并。
- 上游当前为 `upstream/main@baeac1f3d`，最新 tag 为 `v0.1.177@073e92d17`。
- 主工作区存在用户未提交的账户编辑前端源文件、对应测试和 `outputs/`，所有上游 Sprint 均排除它们。

## 当前目标

- S218 已完成并合入本地主线；下一步正式起草 S219，行为级移植上游 `8219dcfc8` 与测试修正 `4d9fedee2`。
- S219 只处理 Codex `x-codex-turn-state` 的 HTTP streaming/non-streaming/SSE-to-JSON 回传，以及已知跨账号 echo 的剥离；不得把 Claude 兼容桥的注入缓存与原生 Codex provenance 混用。

## 本次已完成

- S218 Developer 提交 `f07518322` 与 R1 修复 `1567b88c8` 已由独立 Terra QA 验收通过；QA 报告提交为 `9b8918182`。
- 主线精确合入为 `2058b69c9`、`32c55f9fe`、`d6c7435bd`；分支重复 Amendment `098b4bd82` 未合入，主线已有等价 `0b2aee26f`。
- 原生 `stream:true + compaction_trigger` 现在保留 `/responses`，补齐 `remote_compaction_v2` beta 特性；legacy compact 与 compact-only mapping 保持隔离。
- API-key Responses unsupported/force-chat 账号会在 native-v2 调度时被排除；直接 `Forward` 也不会 raw-chat 转换并吞掉 trigger。
- compact probe 改用 streaming `/responses`，只有收到真实 compaction output item 才记录支持。

## 已确认事实

- `upstream/main` 仍为 `baeac1f3d`，`v0.1.177` peeled commit 仍为 `073e92d17`；tag 之后只有 VERSION 同步提交。
- S219 两个提交不能直接 apply：本地没有上游拆出的 `openai_gateway_response_handling.go`，响应处理仍在单体 `openai_gateway_service.go`。
- 本地请求白名单已允许 `x-codex-turn-state`，WS handshake 也有相关处理，但 HTTP streaming、non-streaming 与 SSE-to-JSON 没有完整显式回传和原生 Codex 跨账号 provenance 守卫。
- 本地已有 Claude 兼容桥自己的 turn-state 缓存与注入语义；S219 必须保持协议边界，只对原生 Codex echo 做已知异账号剥离，不做服务端注入。
- 指纹 opt-in `fce41e318` 与分组日汇总 migration 222/223 继续排除；后者必须单独取得数据库影响授权。

## 待验证点

- S219 contract review -> 验证：所有 HTTP 响应提交点、first-output staging/failover、passthrough 和请求守卫边界均有明确默认标签测试。
- S219 Developer/QA -> 验证：same-account/unknown provenance 保持透传，known cross-account echo 才剥离；无 session/API-key seed 时不跟踪，TTL 清理有界，WS 与 Claude bridge 不退化。

## 当前结论

- `PASS / S218 local-main-integrated`：remote compaction v2 已通过独立 Terra QA 和主线回归，尚未推送。
- `S219 / intake`：turn-state HTTP relay 与跨账号 echo 保护缺口真实存在，下一合法动作是冻结本地主线基线并起草合同。

## 下一步

1. 提交 S218 workflow/knowledge 收口并删除其 clean worktree/local branch -> 验证：main 不受影响，用户 dirty patch-id 不变。
2. 创建并复审 S219 contract -> 验证：只允许本地 turn-state helper、单体 gateway hook 和 default-tag tests，不引入 migration/fingerprint/Claude bridge 改写。
3. 在新隔离 worktree 调度独立 Terra Developer，再由新的 Terra QA 验收 -> 验证：完整 service/handler/server/compile、allowlist/provenance/index 和本地 fixture 边界通过。
4. S219 收口后重新 fetch upstream/origin，决定是否统一推送本地主线。

## 验证记录

- S218 QA：`docs/workflow/qa-reports/upstream-v0177-remote-compaction-v2-s218-qa.md`，首行 PASS。
- S218 主线：focused handler/service `-count=10`、legacy compact、完整 service 64.519s、handler 59.746s、server 与 compile PASS。
- S218 静态：19 个实现文件 `outside=0`，gofmt、diff、冲突/index、三个上游提交 provenance PASS。
- 用户两处前端 dirty patch 当前仍为 `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`；未 push、部署、更新容器、调用 provider 或触碰 migration。
