# 当前主线

最后更新：2026-06-07

## 当前阶段

Sub2API 近期稳定主线又前移了一次。教程 CMS、登录后跳转保持、共享额度窗口、5 月底那批 chat image workspace / embedded session / 模型目录同步，以及 6 月初的 upstream sync / `reference_pricing` 收口结论仍然成立，但它们已经从“最新主线”退成稳定背景层；当前更靠近主线的是 OpenAI 网关稳态收口、account capability routing、用户控制台与 `key/base-url` 归一。

## 当前重点

1. OpenAI 网关稳态与账号能力路由，已经成为当前更高频的主线
   - 最近几天提交重点继续偏向更靠近网关与能力判定的路径，例如 `support routed api key capabilities`、`self-heal stale Codex used% snapshots + lock semantics`、`handle missing messages stream terminal`、`recognize claude code clients via billing block`。
   - 这说明当前默认续做心智应先落在“哪些账号能力判断已稳定、哪些 OpenAI 路径已补稳、哪些使用量/锁语义已经修正”，而不是继续把主线只理解成工作区迁移或模型广场展示。

2. 用户控制台和 `key/base-url` 归一，已经进入默认产品面
   - 最近提交已经把 `user channel status dashboard`、默认 API key bootstrap、`normalize use key base url` 这类能力推进到当前主线。
   - 这些改动会直接影响用户控制台、能力感知、默认接入体验和后续排障入口，已经值得作为稳定事实记录，而不该只留在提交历史里。

3. 6 月初的 upstream sync / `reference_pricing` 收口，现已降为稳定背景层
   - upstream 批次合流、OpenAI 路径第一轮 hardening、模型广场 `reference_pricing` 展示边界仍然有效，但它们已不再代表 6 月 7 日最靠前的默认改动面。
   - 后续继续做模型广场、嵌入工作区或公共入口时仍要遵守这些约束，但如果要快速判断“仓库现在主要在做什么”，应优先看账号能力路由、OpenAI 稳态与控制台入口。

4. 教程 CMS / 登录跳转 / 共享额度窗口仍是基础层，但已经更远离当前高频主线
   - 这些能力已经从“当前主线”退到“稳定背景约束”。
   - 当前补知识时，更值得优先解释 OpenAI 网关稳态、账号能力与控制台链路，而不是重复教程页或旧工作区迁移背景。

## 已稳定结论

- `knowledge/tasks/current-task.md` 仍适合记录动态交付快照，但当前稳定主线已经不该只停在 2026-06-03 的 upstream sync / 模型广场参考价提交语境；最近默认续做心智已继续推进到账号能力与 OpenAI 稳态层。
- 教程页现在是稳定公共内容链路，不再只是公共前端文案；但如果只看它，会低估最近聊天生图与嵌入工作区改动的影响范围。
- 登录后保留目标跳转仍是默认心智，而且最近又进一步延伸到 embedded session 恢复；后续涉及登录/注册/launch/redeem 和用户页入口时，不应假设“丢 token 后直接重新登录”就是可接受行为。
- 模型广场 `reference_pricing` 当前只用于展示，不写入渠道价格，不影响 billing、倍率和扣费链路；后续如果再扩展模型信息展示，应保持这个边界不被误判。
- 近期提交已经证明：本仓库的“当前高频改动”既包括产品入口，也包括上游合并、OpenAI 网关稳态、账号能力路由和控制台状态面；知识入口必须同时覆盖这些事实。
- `sub2api` 与 `chatgpt2api` 仍需分开维护知识：前者偏 gateway、公共入口、嵌入式工作区桥接、模型/计费目录；后者偏独立图片工作台、`/canvas` 节点工作区和 ChatGPT Web 能力封装。
- `use key base url` 归一和 routed API key capabilities 已进入当前稳定主线；后续排查用户“为什么这个 key 看不到某能力/为什么 base URL 表现不一致”时，应优先把它当成默认知识，而不是零散提交细节。

## 现在不该误判的点

- 不要再把当前仓库默认理解成“教程页 + 容量池展示”的延长线；最近主线已经明确偏向聊天生图工作区、嵌入链路和模型目录同步。
- 不要再把当前仓库默认理解成“只是在继续 5 月底的工作区迁移”或“只是在收尾 6 月初的 upstream sync”；最近高频事实已经继续推进到账号能力路由、OpenAI 使用量/锁语义和控制台状态面。
- 不要把 embedded session 修复当成单纯登录页小修；它实际影响嵌入工作区可用性、cookie 恢复语义和用户从 Sub2API 进入工作区后的连续体验。
- 不要把 `claude-opus-4.8` 这类模型补录当成纯展示文案；它通常牵涉 fallback pricing、默认模型、映射和测试。
- 不要把 `reference_pricing` 当成真实计费字段；当前它只影响展示，不改变扣费链路。
- 不要把 routed API key capabilities 或 `key/base-url` 归一当成纯后台小修；它们会影响用户面板、能力暴露、默认接入体验和后续排障路径。
- 不要只看 `docs/ai/current-task.md`；该文件仍是兼容旧入口的说明，当前事实应以 `knowledge/`、`knowledge/tasks/current-task.md` 和时间轴为主。
- 不要把本地 fake 演示账号或 fallback 教程内容当成正式生产数据；它们仍主要服务本地预览、升级兜底或空态保护。
