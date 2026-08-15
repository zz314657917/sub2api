# 当前任务快照

最后更新：2026-08-16 01:39 +08:00

## 背景

- 用户要求持续核实并选择性合入最新上游版本；本地主线长期分叉，禁止整包合并。
- 上游当前为 `upstream/main@baeac1f3d`，最新 tag 为 `v0.1.177@073e92d17`。
- 主工作区存在用户未提交的账户编辑前端源文件、对应测试和 `outputs/`，所有上游 Sprint 均排除它们。

## 当前目标

- S218 已完成并合入本地主线；下一步正式起草 S219，行为级移植上游 `8219dcfc8` 与测试修正 `4d9fedee2`。
- S219 只处理 Codex `x-codex-turn-state` 的 HTTP streaming/non-streaming/SSE-to-JSON 回传，以及已知跨账号 echo 的剥离；不得把 Claude 兼容桥的注入缓存与原生 Codex provenance 混用。
- `8219dcfc8` 只定义守卫，实际 normal/passthrough builder 调用位于 `fce41e318`；S219 只取这两个守卫挂点，继续排除 fingerprint 默认值、收敛、client metadata 和 frontend。

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
- `fce41e318` 的 fingerprint opt-in/default、收敛、client metadata 和 frontend 继续排除，只允许复用其中两个 turn-state guard 调用位置；分组日汇总 migration 222/223 必须单独取得数据库影响授权。

## 待验证点

- S219 contract review -> 验证：所有 HTTP 响应提交点、first-output staging/failover、passthrough 和请求守卫边界均有明确默认标签测试。
- S219 Developer/QA -> 验证：same-account/unknown provenance 保持透传，known cross-account echo 才剥离；无 session/API-key seed 时不跟踪，TTL 清理有界，WS 与 Claude bridge 不退化。

## 当前结论

- `PASS / S218 local-main-integrated`：remote compaction v2 已通过独立 Terra QA 和主线回归，尚未推送。
- `S219 / contract-approved`：Evaluator 已确认 commit-boundary、正 API-key/session seed、stale clear、normal/passthrough guard 和现有兼容测试，下一合法动作是创建隔离 worktree 并调度独立 Terra Developer。

## 下一步

1. 创建 `E:/codex-worktrees/sub2api/s219-turn-state` 并调度独立 Terra Developer -> 验证：基线为 contract approval commit，只改 6 个 allowlisted 路径。
2. 主控审 diff 并复跑 focused/compatibility 门禁 -> 验证：实际 flush provenance、stale clear 和 guard 行为满足合同。
3. 新建独立 Terra QA 验收 -> 验证：完整 service/handler/server/compile、allowlist/provenance/index 和本地 fixture 边界通过。
4. S219 收口后重新 fetch upstream/origin，决定是否统一推送本地主线。

## 验证记录

- S218 QA：`docs/workflow/qa-reports/upstream-v0177-remote-compaction-v2-s218-qa.md`，首行 PASS。
- S218 主线：focused handler/service `-count=10`、legacy compact、完整 service 64.519s、handler 59.746s、server 与 compile PASS。
- S218 静态：19 个实现文件 `outside=0`，gofmt、diff、冲突/index、三个上游提交 provenance PASS。
- 用户两处前端 dirty patch 当前仍为 `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`；未 push、部署、更新容器、调用 provider 或触碰 migration。
