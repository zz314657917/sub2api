### PASS: upstream-a-fixes-20261028

2026-10-28，第一批上游 A1–A7 已按本地架构适配并精确合回工作区。独立合同复审、三组开发、独立组合 QA、主控源码复核和主树真实 dirty 组合复验均完成。没有提交、推送或部署。

## Findings

未发现剩余明确问题，无失败豁免。生产修改仅 11 个文件，新增 13 个测试文件、32 个测试声明；原有测试断言和无关 dirty 文件保持原状。

| 项目 | 上游来源 | 本地结果 |
| --- | --- | --- |
| A1 国产频率型 429 | de908ed6d | 仅明确频控且有效额度窗口均低于 90% 时使用既有短冷却；余额、额度和坏快照保守处理 |
| A2 DeepSeek 缓存读取 | 3e18e78ab | prompt_cache_hit_tokens 加入末级兜底，保留嵌套显式 0、数值范围和重复请求幂等 |
| A3 Claude 指纹 | 7bce2cfb4 | 按 UTF-16 码元 4/7/20 取样，拼接后编码，保持 ASCII 和真实请求 header 一致 |
| A4 分页溢出 | f96528048、9aa16c8c6 | 安全解析 native int、限制完整页 offset；严格接口返回 400，保留本地 repository 调用链 |
| A5 multipart 上传 | cde6ddfd8 | 每 part 读取 50 MiB+1 探测超限并返回 400，合法大文件保持完整 |
| A6 Gemini 图片计数 | 917dbdc70 | 支持 inline/file、camel/snake 表示的非空字符串，每 part 最多一次，累计帧取最大值 |
| A7 Antigravity 非流式错误 | 14a5fe305 | direct/wrapped 及无 parts 错误末帧返回 502 malformed_function_call，不产生成功结果 |

## Executed Checks

固定 baseCommit：`b3091c1c9a28884cf4b00fe97e87fc2f295b0512`；上游 pin：`f178e5ae32659e8901644a9f2378a0439ffa36ae`。

Go：`E:/tools/go1.26.10/go/bin/go.exe`，GOROOT 同工具根目录，GOTOOLCHAIN=local，GOMAXPROCS=4，test/build 使用 -p=2。下表数量为通过事件，包含子测试。

| 实际验收 | 隔离候选 | 独立 QA | 主树真实 dirty 组合 |
| --- | --- | --- | --- |
| canonical service default / unit | 1261 / 1498 | 1261 / 1498 | 1261 / 1498 |
| 其他相关包 | 104 | 104 | 104 |
| 完整 backend build | exit 0 | exit 0 | exit 0 |
| 新增测试 default / unit，count=3 | 912 / 912 | 912 / 912 | 912 / 912 |
| 32 声明逐项执行审计 | 每标签各 3 次 | 每标签各 3 次 | 每标签各 3 次 |
| 完整 pagination/response/antigravity 包 | 211 | 211 | 211 |
| 三个既有 Gemini 测试 default / unit | 12 / 12 | 12 / 12 | 12 / 12 |

以上全部 exit 0、无失败、无 Test skip。主树原基线 default1045/unit1282/packages16 的每个匹配测试都仍执行，无遗漏。

独立 QA 另写 17 个顶层测试，default/unit 各 260 个通过事件，覆盖真实调度写入、RecordUsage 幂等、Node.js UTF-16 oracle、big.Int 分页 oracle、严格 handler、实际切片消费者、URI 图片 ForwardNative/reset 和非流 collector。真实 localhost Images HTTP 探针确认 image/mask/prompt 的 50 MiB+1 返回 400，恰好 50 MiB 和 20 MiB+1 通过解析；正常组权限哨兵为 403。

24 文件 gofmt、限定范围 diff --check、未合并路径检查均通过。主控逐行检查全部 11 个生产 diff；无依赖、schema、目录价格、frontend 或额外功能变更。

## Freeze and Protection

- 独立合同 r2 PASS 绑定原 draft SHA256 `dddec9dece08f0efcedfb66fe679743be0e569c73abc274dbaf35668dc225c54`。原 r1 FAIL 保留，修订只消除了本地不存在的 API Key 内存分页移植要求。
- 固定 run-checks.mjs SHA256：`b77edead5be4802d27ed3986c25cf6676fbb76c6b6b21bfa8cb1a36f6820ac74`，开发和 QA 未改。
- 候选 freeze SHA256：`ebd4262de3bec577ac4eb6967c35d21bbb74f5c3db2d83784c47e33416b58bfd`。
- 独立 QA seal SHA256：`21e95377ed33dca7bfbbbe80e214e5f29b9c64ff99757e75dfb92a53587d14db`，绑定全部 24 源文件及 122 份证据。主控合回前后均验证全部封印。
- 主树原 12212 文件中，12197 个原文件字节完全不变；变化仅为本批 11 个生产文件和 4 个既有流程/任务文档。仅新增 13 个目标 Go 测试；HEAD、index、原 dirty 依赖及其他无关改动保留，无冲突。
- QA 临时探针已从独立 checkout 清理，原始源码和日志保留在外置证据目录。

## Evidence and Corrections

证据根：`E:/codex-runtime/sub2api/upstream-a-fixes-20261028`。

- 独立原始报告和封印：`qa/qa-report.md`、`qa/qa-seal.json`。
- 主树新鲜运行：`main-after/`、`main-after-new/`、`main-after-extra/`、`main-after-declarations.json`。
- 最终主控复核：`controller-final-verification.json`、`main-final-scope.json`、`integration.json`。
- 原实现 red、开发中间态失败、独立 QA 常量/UTF16 草稿/成本 oracle 修正历史全部保留；没有修改既有断言或用产品代码迁就错误探针。A2 QA 成本 fixture 原期望错误，最终固定请求时刻精确断言 0.0001209 和只扣费一次，双标签通过。
- 外置冻结检查第一次误将 CRLF 检出与 LF Git blob 差异当漂移，保留 attempt-1，改为分别封印原字节并核对规范化内容；依赖未改。最终主控检查第一次因 Git 换行 warning 被误判失败，保留 attempt-1，仅在只读冲突检查使用命令级 core.safecrlf=false 后通过，未修改仓库配置或源码。

## Unverified Risks

未连接真实 provider、数据库或容器，未做部署、生产流量或真实供应商额度快照验收。验证限于内存 repository 端口、真实生产函数与 localhost HTTP；本报告不代表线上行为已更新。

## Recommendation

PASS：本批可保留在本地继续提测。提交、推送、部署及其他候选功能等待独立授权。
