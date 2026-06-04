# 当前主线

最后更新：2026-06-03

## 当前阶段

Sub2API 近期稳定主线又前移了一次。教程 CMS、登录后跳转保持、共享额度窗口，以及 5 月底那批 chat image workspace / embedded session / 模型目录同步结论仍然成立，但它们已经从“最新主线”退成稳定背景层；当前更靠近主线的是“上游同步批次收口 + OpenAI 路径稳态修复 + 模型广场参考价展示补齐”。

## 当前重点

1. upstream 同步批次与 OpenAI 路径稳态修复，已经成为当前更高频的主线
   - 最近 3 天提交重点不再是前端工作区体验本身，而是围绕 `upstream/main` 差异批次做安全合入，并修复 `oauth refresh credentials`、`401 preserve credentials`、`ws usage dedup conflicts` 这类更靠近网关稳态的路径。
   - 这说明当前默认续做心智应先落在“哪些上游批次已合入、哪些仍待评估、哪些 OpenAI 网关路径已补稳”，而不是继续把主线只理解成聊天生图工作区迁移。

2. 模型广场参考价展示已经进入默认产品面
   - `feat(models): show reference pricing in plaza` 已把 `reference_pricing` 从后端能力补到用户可见入口，且前端已兼容旧后端 `supportedModel.reference_pricing ?? null`。
   - 这类改动虽然不影响真实 billing/倍率/渠道价格，但它会影响模型广场、默认展示口径和后续“模型信息是否齐全”的产品判断，已经值得作为稳定事实记录。

3. 5 月底的 chat image workspace / embedded session / 模型目录同步，现已降为稳定背景层
   - `/chat-images`、`/canvas`、OpenWebUI launch、embedded session recovery、模型目录和 fallback pricing 这些结论仍然有效，但它们已不再代表 6 月初最靠前的默认改动面。
   - 后续继续做工作区或嵌入链路时仍要遵守这些约束，但如果要快速判断“仓库现在主要在做什么”，应优先看 upstream sync、OpenAI 路径和模型广场展示链路。

4. 教程 CMS / 登录跳转 / 共享额度窗口仍是基础层，但已经更远离当前高频主线
   - 这些能力已经从“当前主线”退到“稳定背景约束”。
   - 当前补知识时，更值得优先解释 upstream sync 批次、OpenAI 网关稳态修复和模型广场展示事实，而不是重复教程页的上线背景。

## 已稳定结论

- `knowledge/tasks/current-task.md` 仍适合记录动态交付快照，但 2026-06-03 这一轮快照已经切到“本地分支合并、upstream 同步链核对、模型广场参考价提交”，说明仓库当前稳定主线不该再只停在 2026-05-30 的工作区语境。
- 教程页现在是稳定公共内容链路，不再只是公共前端文案；但如果只看它，会低估最近聊天生图与嵌入工作区改动的影响范围。
- 登录后保留目标跳转仍是默认心智，而且最近又进一步延伸到 embedded session 恢复；后续涉及登录/注册/launch/redeem 和用户页入口时，不应假设“丢 token 后直接重新登录”就是可接受行为。
- 模型广场 `reference_pricing` 当前只用于展示，不写入渠道价格，不影响 billing、倍率和扣费链路；后续如果再扩展模型信息展示，应保持这个边界不被误判。
- 近期 upstream sync 相关提交已经证明：本仓库的“当前高频改动”既包括产品入口，也包括上游合并与网关稳态收口；知识入口必须同时覆盖这两类事实。
- `sub2api` 与 `chatgpt2api` 仍需分开维护知识：前者偏 gateway、公共入口、嵌入式工作区桥接、模型/计费目录；后者偏独立图片工作台、`/canvas` 节点工作区和 ChatGPT Web 能力封装。

## 现在不该误判的点

- 不要再把当前仓库默认理解成“教程页 + 容量池展示”的延长线；最近主线已经明确偏向聊天生图工作区、嵌入链路和模型目录同步。
- 不要再把当前仓库默认理解成“只是在继续 5 月底的工作区迁移”；6 月初的高频事实已经包含 upstream sync 批次、OpenAI 路径修复和模型广场参考价展示。
- 不要把 embedded session 修复当成单纯登录页小修；它实际影响嵌入工作区可用性、cookie 恢复语义和用户从 Sub2API 进入工作区后的连续体验。
- 不要把 `claude-opus-4.8` 这类模型补录当成纯展示文案；它通常牵涉 fallback pricing、默认模型、映射和测试。
- 不要把 `reference_pricing` 当成真实计费字段；当前它只影响展示，不改变扣费链路。
- 不要只看 `docs/ai/current-task.md`；该文件仍是兼容旧入口的说明，当前事实应以 `knowledge/`、`knowledge/tasks/current-task.md` 和时间轴为主。
- 不要把本地 fake 演示账号或 fallback 教程内容当成正式生产数据；它们仍主要服务本地预览、升级兜底或空态保护。
