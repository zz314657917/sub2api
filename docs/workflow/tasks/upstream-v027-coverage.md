# v0.2.7 剩余行为移植覆盖表

工作树：`E:/codex-worktrees/sub2api/upstream-v027-port`；分支：`codex/upstream-v027-port`。
基线：`d6e6c34718e6eee6388391b346f40e1492e81d57`（先前 Anthropic 根 union schema 修复已推送）。
本轮仅保留未提交的行为适配；不代表已合入 main 或已部署。

## 后端

| 上游提交 | 本地行为 | 当前验收 |
| --- | --- | --- |
| 881ab1b0c / acc05620c | DeepSeek 工具图片转入后续 user 消息，保持并行输出连续与大整数精度 | 独立 QA PASS |
| 0f4d8acaa | Gemini 裸模型按 thinkingConfig 选择变体，保留显式映射 | 已补精确 identity mapping 断言，独立 QA PASS |
| f79b8bf96 | Go/Python GenAI 客户端不发送 SSE 注释心跳 | 独立 QA PASS |
| da74bf13a | Gemini 模型列表读取可调度 Antigravity 账户映射 | 独立 QA PASS；本地不存在的上游 Group allowlist API 不适用 |
| 18bfa4bf2 | 严格 Chat 供应商 developer role 转 system | 独立 QA PASS |
| 2f16e0984 / 31f3003ff | manifest 保留大小写敏感 models 校验并去除重复解析 | 独立 QA PASS |
| d7ee1ab6b | 客户端取消后仍在限定超时内保存响应亲和关系 | 独立 QA PASS |
| db8692d67 | Coding Plan 配额耗尽 403 使用临时冷却并同步进程内调度状态 | QA 发现通知遗漏，已修复，独立复验 PASS |

## 前端

| 上游提交 | 本地行为 | 实施状态 |
| --- | --- | --- |
| 406be7c518 | 清空订阅时恢复 loading | 已实现 |
| b51f0759d4 | 并发付款配置调用等待同一请求 | 已实现 |
| be313735dc | 多对话框标题 ID 唯一 | 已实现 |
| 017e9e98d1 | 余额不足提示比较本次退款额 | 已实现 |
| fe36f4a914 | 注册优惠码显示使用初始公开配置 | 已实现 |
| f8e5a0d94f | 空模型输入框允许 Tab 导航 | 已实现 |
| 0838e0e610 | 管理员平台配额禁止负数 | 不适用：本地不存在该编辑器，独立修订审查 PASS |
| 98321a054e | 非法充值金额恢复已接受的输入文本 | 已实现 |
| ee9ac3e45f | TOTP 验证失败后显示位同步清空 | 已实现 |
| 14636d2f0e | TOTP 展示统一提取的 API 错误 | 已实现 |
| 1a32b91eb2 | 页码跳转兼容数字值 | 已实现 |
| 130ba634ad | 剪贴板回退抛异常时返回失败并清理 | 已实现 |
| 6a4938bdfb | 批量已读保留部分成功结果 | 已实现 |
| 0ed735d3ba | 单项与批量代理测试去重 | 已实现 |
| 21532add46 | 订单筛选重置页码 | 已实现 |

前端独立 QA PASS：14 个定向测试文件、24 个用例通过，vue-tsc 通过。任务私有环境按原锁文件安装真实依赖后生产构建通过（无 stub 或 external 绕过）。
组件浏览器 smoke 在专属 Edge profile 中通过，桌面与 390px 截图已人工查看；session、Vite 和 profile/daemon 归属进程已清理。
浏览器临时非支付 smoke 曾为缺失的无关 Airtracker 依赖使用隔离 stub，此配置已删除；生产构建使用真实依赖，不依赖该 stub。
主控整体验证：apicompat/service/handler 的 TestV027、manifest、响应亲和、CN 403 定向集合通过，go build ./... 通过；路径 allowlist、diff 和冲突检查通过，暂存区为空。
证据目录：`docs/workflow/worker-results/`、`docs/workflow/qa-reports/`、`outputs/upstream-v027-ui/`。
测试使用 localhost/fake API；真实供应商、支付、Redis/数据库及部署未验证。
