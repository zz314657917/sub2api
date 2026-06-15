# 当前主线

最后更新：2026-06-14

## 当前阶段

Sub2API 近期稳定主线又前移了一次。教程 CMS、登录后跳转保持、共享额度窗口、5 月底那批 chat image workspace / embedded session / 模型目录同步、6 月初的 upstream sync / `reference_pricing` 收口，以及 6 月 7 日前后的 OpenAI 网关稳态、account capability routing、用户控制台与 `key/base-url` 归一都仍然成立，但它们已经从“最新主线”退成稳定背景层；当前更靠近主线的是 Studio Bridge / 落叶AI生产联调。

## 当前重点

1. Studio Bridge / 落叶AI生产联调，已经成为当前最高频的主线
   - `fe2f80be1 feat: add studio bridge integration` 已把 Sub2API 扩展为落叶AI的账号、充值、余额、默认分组、配置和扣费真源。
   - 这说明当前默认续做心智应先落在“bridge launch/redeem 是否闭环、默认分组和 internal secret 是否配置正确、预扣/确认/退款是否能稳定联调、团队空间 actor/payer 语义是否跑通”，而不是继续把主线只理解成 OpenAI gateway 或旧工作区迁移。
   - 6 月 10 日到 11 日的新稳定事实是：本地 Studio Bridge 配置现在会在 env secret 存在、配置为空或仍是占位值时自动补齐；默认聊天/生图分组会从 active groups 动态选择，不再硬编码旧 group id；继续排查本地 `STUDIO_BRIDGE_DISABLED` / `STUDIO_BRIDGE_GROUP_REQUIRED` 时，先看 env secret、active image group 和占位配置，而不是先怀疑 launch/redeem 本身失效。

2. 当前用户侧入口已经从 OpenWebUI / 旧聊天生图入口，切到落叶AI启动链路
   - `/studio-bridge/launch` 已作为 `/chat-images` 的 alias，避免注册/登录 redirect 到 404。
   - 这会直接影响默认产品入口、登录回跳、用户认知和浏览器验收路径，已经值得作为稳定事实记录，而不该只留在 `current-task` 里。
   - 近期还补齐了侧栏直接启动 Studio Bridge 的路径；因此当前用户入口不只是一张落地页，而是“`/chat-images` alias + sidebar launch + 登录/注册后回跳”的整条链路。

3. session-probe 已从临时调试页进入默认验收面
   - 6 月 10 日新增 `session-probe` iframe 探针后，当前最小 smoke 不再只是“launch 到 `/image` 返回 200”，还要确认 iframe 只请求 `/studio-bridge/session-probe`，并且 CSP / `frame-ancestors` 允许落叶AI宿主域名，而不是错误回退到根路径 iframe 或被浏览器拦截。
   - 这条约束直接影响登录态恢复、余额摘要展示和落叶AI内页是否能稳定读取 Sub2API 会话，不应继续只留在任务时间轴里。

4. OpenAI 网关稳态、账号能力路由和控制台归一，现已降为 Studio Bridge 之前的稳定背景层
   - gateway/auth/session、prompt cache、routed API key capabilities、`key/base-url` 归一仍然有效，但它们已不再代表 6 月 9 日最靠前的默认改动面。
   - 后续继续做模型广场、嵌入工作区、公共入口或上游合成时仍要遵守这些约束，但如果要快速判断“仓库现在主要在做什么”，应优先看 Studio Bridge 配置、真实用户闭环和跨仓库联调。
   - 但最近的 `default API key` / route groups 收口仍是这层背景里必须保住的约束：Studio Bridge 和默认 key 路由改造之后，普通更新路径仍要继续校验分组权限，不能因为补 bridge 入口就放松 API key route group 校验。

5. 首充福利与注册来源信息已经进入当前后台稳定面
   - 6 月 10 日新增首充福利 bonus，说明当前“充值闭环”已不只包含支付成功和余额回写，还包含用户福利/运营奖励规则；后续再改充值、福利、用户余额历史或兑换页时，不应把它当成独立于 Studio Bridge 的边缘功能。
   - `register ip` 已进入管理员用户列表稳定字段，说明当前认证/用户治理也在同步前移；后续排查新用户试用、福利领取和风控限制时，应默认考虑注册来源信息，而不是只看 user id 或 email。

6. 支付套餐配置与用户 IP 画像，已经进入当前后台稳定面
   - 2026-06-13~2026-06-14 的新提交说明，最近高频改动不再只是“首充福利 + Studio Bridge 充值回跳”，而是继续推进到“可配置充值套餐 + 后台支付兑现/恢复 + 用户注册/最近登录 IP 画像”。
   - 这意味着当前支付面已经从单次支付成功与余额回写，进一步前移到“套餐定义、支付恢复、用户支付页展示、福利兑现、后台治理”一整条链路。
   - 同期进入后台稳定面的还有用户 IP 字段；后续做风控、异常支付、OAuth 注册来源或新用户治理时，不应再把 IP 当成一次性排障字段。

7. 教程 CMS / 登录跳转 / 共享额度窗口仍是基础层，但已经更远离当前高频主线
   - 这些能力已经从“当前主线”退到“稳定背景约束”。
   - 当前补知识时，更值得优先解释 OpenAI 网关稳态、账号能力与控制台链路，而不是重复教程页或旧工作区迁移背景。

## 已稳定结论

- `knowledge/tasks/current-task.md` 仍适合记录动态交付快照，但当前稳定主线已经不该只停在 2026-06-08 的 gateway auth / prompt cache 语境；最近默认续做心智已继续推进到 Studio Bridge、真实用户闭环和落叶AI生产联调层。
- 教程页现在是稳定公共内容链路，不再只是公共前端文案；但如果只看它，会低估最近聊天生图与嵌入工作区改动的影响范围。
- 登录后保留目标跳转仍是默认心智，而且最近又进一步延伸到 embedded session 恢复；后续涉及登录/注册/launch/redeem 和用户页入口时，不应假设“丢 token 后直接重新登录”就是可接受行为。
- 模型广场 `reference_pricing` 当前只用于展示，不写入渠道价格，不影响 billing、倍率和扣费链路；后续如果再扩展模型信息展示，应保持这个边界不被误判。
- 近期提交已经证明：本仓库的“当前高频改动”既包括产品入口，也包括上游合并、OpenAI 网关稳态、账号能力路由和控制台状态面；知识入口必须同时覆盖这些事实。
- OpenWebUI 不再是当前默认用户侧“聊天生图”入口；当前默认入口应进入落叶AI启动链路。
- 充值、用户、余额、默认聊天/生图/视频分组和 bridge internal secret 仍应由 Sub2API 管理后台维护。
- 如果存在 `STUDIO_BRIDGE_LUOYE_AI_INTERNAL_SECRET`，且当前本地配置为空、缺 secret/group、禁用或仍是 `example.com` 占位，系统会自动修复本地 launch return URL、充值回跳 URL 和 `127.0.0.1/localhost` allowed domains；正式域名配置不会被这套本地自修复覆盖。
- 默认生图分组不再硬编码旧 id，而是优先选择第一个 active 且 `allow_image_generation=true` 的 image group；默认聊天分组优先 text group，缺失时可复用 image group。
- `reserve / commit / refund` 的预扣/确认/退款语义已经进入当前默认联调面；后续排查任务失败、余额不足或取消回退时，应优先把它当成默认知识，而不是临时接口细节。
- Studio Bridge 当前最小浏览器 smoke 还应确认 session-probe iframe 只打 `/studio-bridge/session-probe`，且 `frame-ancestors` / CSP 对落叶AI父页面放行；如果这里退化，即使 launch token 和 redeem 仍返回 200，用户态也可能表现为“已登录但余额/会话不同步”。
- 首充福利 bonus、注册 IP 和用户福利页已经进入最近稳定产品面；后续改支付、用户列表、福利配置或充值页时，要把它们视为与 Studio Bridge 同期生效的后台默认约束。
- 可配置充值套餐已经进入当前稳定支付面；后续改支付页、后台设置、支付恢复或兑现逻辑时，要默认考虑“套餐来自后台配置”而不是固定面额。
- 管理员用户列表里的注册 IP / 最近登录 IP 已从辅助排障信息升级为稳定后台画像；后续风控、福利发放和异常账户排查应默认读取它们。
- `sub2api` 与 `chatgpt2api` 仍需分开维护知识：前者偏 gateway、公共入口、嵌入式工作区桥接、模型/计费目录；后者偏独立图片工作台、`/canvas` 节点工作区和 ChatGPT Web 能力封装。
- `use key base url` 归一和 routed API key capabilities 已进入当前稳定主线；后续排查用户“为什么这个 key 看不到某能力/为什么 base URL 表现不一致”时，应优先把它当成默认知识，而不是零散提交细节。
- 默认 API key / 默认分组改造之后，普通更新路径仍应执行 route groups 权限校验；如果未来再看到这块编译或逻辑回退，先检查 `validateAPIKeyRouteGroups(..., false)` 一类调用是否被遗漏。

## 现在不该误判的点

- 不要再把当前仓库默认理解成“教程页 + 容量池展示”的延长线；最近主线已经明确偏向聊天生图工作区、嵌入链路和模型目录同步。
- 不要再把当前仓库默认理解成“只是在继续 5 月底的工作区迁移”或“只是在收尾 6 月初的 upstream sync”；最近高频事实已经继续推进到账号能力路由、OpenAI 使用量/锁语义和控制台状态面。
- 不要再把当前仓库默认理解成“只是在继续 gateway auth / prompt cache follow-up”；最近高频事实已经继续推进到 Studio Bridge、真实充值/扣费闭环和团队空间联调。
- 不要把 embedded session 修复当成单纯登录页小修；它实际影响嵌入工作区可用性、cookie 恢复语义和用户从 Sub2API 进入工作区后的连续体验。
- 不要把 `claude-opus-4.8` 这类模型补录当成纯展示文案；它通常牵涉 fallback pricing、默认模型、映射和测试。
- 不要把 `reference_pricing` 当成真实计费字段；当前它只影响展示，不改变扣费链路。
- 不要把 routed API key capabilities 或 `key/base-url` 归一当成纯后台小修；它们会影响用户面板、能力暴露、默认接入体验和后续排障路径。
- 不要把 `/studio-bridge/launch` 当成单纯路由 alias；它实际承担当前用户侧启动入口，直接影响注册/登录回跳和浏览器验收。
- 不要把本地 Studio Bridge 自修复当成生产配置机制；它只服务本地空配置/占位配置恢复，正式域名、正式 group 和正式 secret 仍应由管理员明确配置。
- 不要把 session-probe 当成可删的调试页；当前它已经进入默认会话恢复和 iframe 安全边界。
- 不要把首充福利 bonus 或注册 IP 当成纯运营字段；它们已经进入支付兑现、新用户治理和后台用户视图的稳定数据面。
- 不要把可配置充值套餐当成单纯设置页小功能；它实际影响用户支付入口、恢复逻辑、福利兑现和后续验收基线。
- 不要把用户 IP 字段当成临时调试字段；它已经进入用户治理、风控和运营判断的默认后台视图。
- 不要把落叶AI团队空间联调误判成 chatgpt2api 单仓库任务；当前团队空间的 actor/payer、余额和扣费真源仍在 Sub2API。
- 不要只看 `docs/ai/current-task.md`；该文件仍是兼容旧入口的说明，当前事实应以 `knowledge/`、`knowledge/tasks/current-task.md` 和时间轴为主。
- 不要把本地 fake 演示账号或 fallback 教程内容当成正式生产数据；它们仍主要服务本地预览、升级兜底或空态保护。
