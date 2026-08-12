# 当前主线

## 2026-08-12 之后

- 默认续做心智已从 `Usage S135-S138` / `Pixel Cafe S139+` 继续前移到 `S211 标准分组时段倍率` 与 `S212 账号时段可用性`；较早的 Pixel Cafe、Usage、`S111/S112`、`group-buy` 和更早的 Studio Bridge 语境只保留为背景。
- 继续接手时，先看 `docs/workflow/status.md` 和 `knowledge/tasks/current-task.md`，再用 `knowledge/tasks/timeline.md` 回收最近阶段历史。

最后更新：2026-08-12

## 当前阶段

Sub2API 近期稳定主线再次前移。截至 2026-08-12，默认续做心智应先落在 `S211 标准分组时段倍率 + S212 账号时段可用性 + phase=done + 当前 workflow/status 为准`，而不是继续停在 2026-08-03 的 `Usage S135-S138 + Pixel Cafe S139+` 语境。

## 当前重点

1. `S211 标准分组时段倍率` 已成为当前最近一层分组 / 计费主线
   - 标准分组现在也支持既有 `peak_rate_*` 语义的同日时段倍率；窗口外返回 `1.0`，不改字段名、不引入 schema/migration。
   - 继续接手 group、rate、usage、billing 或调度相关工作时，要先把 `S211` 视为当前默认产品/工程基线，而不是先回退到更早的 Usage 统计语境。

2. `S212 账号时段可用性` 已成为当前最近一层账号 / 调度主线
   - 单账号可配置 server-timezone 的每日可用窗口；窗口外账号会被排除出新请求调度，但不会改持久化 `status`、`schedulable`、分组绑定或 API Key 绑定。
   - 这条主线直接覆盖账号调度、Antigravity gate、account dialog、管理员配置与局部视觉验收边界，已经值得进入入口知识，而不应只留在 workflow/current-task。

3. 当前工作树边界已经切到 `S211/S212` 的 account/group/admin 实现面
   - 主工作树当前高频未提交改动集中在 `backend/internal/service/**`、`backend/internal/handler/**`、`frontend/src/components/account/**`、`frontend/src/views/admin/GroupsView.vue` 与 `docs/workflow/**`。
   - 这意味着后续做任何上游小步迁移、调度兼容、管理端 UI 或知识回写时，都要先确认本轮范围，不能再沿用 8 月初那批 Pixel Cafe / group-buy dirt 的边界判断。

4. `S210`、`S209`、`S208`、`S207` 已退成前一层稳定工程背景，但不能丢失
   - `S210` 的 streaming terminal audit / WebSocket audit dedupe、`S209` 的 API key 输入校验、`S208` 的 streaming route cooldown 传递、`S207` 的 availability / fallback 小步上游适配都已经进入当前网关稳态。
   - 继续做分组倍率、账号可用性、网关调度或管理端配置时，不应把这些能力当成“另一个旧 Sprint”；它们是当前主线默认继承的下层基线。

5. `Usage S135-S138`、`Pixel Cafe S139+`、`S65-S70`、暖白前端统一、共享账号渠道状态可见性和首充 only 语义仍成立，但都已退成更早的稳定背景层
   - `d6ff6a158`、`640b9341d`、`7a457f25d`、`71dad20f9` 这些能力仍要保住，但它们不再代表 2026-07-19 时最近的默认续做入口。
   - 当前补知识或恢复上下文时，如果入口还把“8 月 3 日 Usage / Pixel Cafe 收口”写成最近主线，会明显低估 `S207-S212` 对调度、倍率、账号可用性和管理端的基线前移。

6. 排行榜 / 数据台继续前移，已经不再只是 6 月 24 日那轮模型榜补齐
   - 2026-07-08 的 `feat(leaderboard): show rank movement` 和 `feat(leaderboard): show new rank and cached refresh state` 说明当前 leaderboard 稳定面已继续扩展到“排名变化 + 新晋标记 + 缓存刷新状态”。
   - 这意味着当前用户数据台默认不只包含模型榜、Token 占比和增长百分比，还包含榜单状态反馈与更强的周期对比语义；后续若再看 dashboard / leaderboard，不应继续按 6 月 24 日的旧卡片结构理解。

7. 共享账号渠道状态可见性，已经进入新的稳定权限边界
   - `7a457f25d feat: add channel status visibility setting for shared accounts` 说明共享账号的 `channel status` 是否对用户侧可见，已经进入后台设置与公共载荷边界，而不再是前端临时展示细节。
   - 这条约束会直接影响共享账号可见字段、用户监控页、容量池解释和客服排障口径；后续涉及 shared account 展示时，应默认先看后台 visibility setting，而不是先改用户页文案。

8. 首充福利 bonus 已收口为“仅首次”语义，不能再按宽松赠送理解
   - `71dad20f9 fix(payment): make recharge package bonus first-time only` 把充值套餐 bonus 收口到首次充值语义，说明当前支付/福利默认边界已从“存在福利”进一步前移到“福利触发条件必须稳定一致”。
   - 后续再改支付页、福利页、套餐说明或后台设置时，不能只记得有首充奖励，还要明确“首充 only”已经是稳定产品约束。

9. Studio Bridge / 落叶AI生产联调，现已降为当前前端统一主线之前的稳定背景层
   - `fe2f80be1 feat: add studio bridge integration` 已把 Sub2API 扩展为落叶AI的账号、充值、余额、默认分组、配置和扣费真源。
   - 这说明当前默认续做心智应先落在“bridge launch/redeem 是否闭环、默认分组和 internal secret 是否配置正确、预扣/确认/退款是否能稳定联调、团队空间 actor/payer 语义是否跑通”，而不是继续把主线只理解成 OpenAI gateway 或旧工作区迁移。
   - 6 月 10 日到 11 日的新稳定事实是：本地 Studio Bridge 配置现在会在 env secret 存在、配置为空或仍是占位值时自动补齐；默认聊天/生图分组会从 active groups 动态选择，不再硬编码旧 group id；继续排查本地 `STUDIO_BRIDGE_DISABLED` / `STUDIO_BRIDGE_GROUP_REQUIRED` 时，先看 env secret、active image group 和占位配置，而不是先怀疑 launch/redeem 本身失效。

10. 当前用户侧入口已经从 OpenWebUI / 旧聊天生图入口，切到落叶AI启动链路
   - `/studio-bridge/launch` 已作为 `/chat-images` 的 alias，避免注册/登录 redirect 到 404。
   - 这会直接影响默认产品入口、登录回跳、用户认知和浏览器验收路径，已经值得作为稳定事实记录，而不该只留在 `current-task` 里。
   - 近期还补齐了侧栏直接启动 Studio Bridge 的路径；因此当前用户入口不只是一张落地页，而是“`/chat-images` alias + sidebar launch + 登录/注册后回跳”的整条链路。

11. session-probe 已从临时调试页进入默认验收面
   - 6 月 10 日新增 `session-probe` iframe 探针后，当前最小 smoke 不再只是“launch 到 `/image` 返回 200”，还要确认 iframe 只请求 `/studio-bridge/session-probe`，并且 CSP / `frame-ancestors` 允许落叶AI宿主域名，而不是错误回退到根路径 iframe 或被浏览器拦截。
   - 这条约束直接影响登录态恢复、余额摘要展示和落叶AI内页是否能稳定读取 Sub2API 会话，不应继续只留在任务时间轴里。

12. OpenAI 网关稳态、账号能力路由和控制台归一，现已降为 Studio Bridge 之前的稳定背景层
   - gateway/auth/session、prompt cache、routed API key capabilities、`key/base-url` 归一仍然有效，但它们已不再代表 6 月 9 日最靠前的默认改动面。
   - 后续继续做模型广场、嵌入工作区、公共入口或上游合成时仍要遵守这些约束，但如果要快速判断“仓库现在主要在做什么”，应优先看 Studio Bridge 配置、真实用户闭环和跨仓库联调。
   - 但最近的 `default API key` / route groups 收口仍是这层背景里必须保住的约束：Studio Bridge 和默认 key 路由改造之后，普通更新路径仍要继续校验分组权限，不能因为补 bridge 入口就放松 API key route group 校验。

13. 首充福利与注册来源信息已经进入当前后台稳定面
   - 6 月 10 日新增首充福利 bonus，说明当前“充值闭环”已不只包含支付成功和余额回写，还包含用户福利/运营奖励规则；后续再改充值、福利、用户余额历史或兑换页时，不应把它当成独立于 Studio Bridge 的边缘功能。
   - `register ip` 已进入管理员用户列表稳定字段，说明当前认证/用户治理也在同步前移；后续排查新用户试用、福利领取和风控限制时，应默认考虑注册来源信息，而不是只看 user id 或 email。

14. 支付套餐配置与用户 IP 画像，已经进入当前后台稳定面
   - 2026-06-13~2026-06-14 的新提交说明，最近高频改动不再只是“首充福利 + Studio Bridge 充值回跳”，而是继续推进到“可配置充值套餐 + 后台支付兑现/恢复 + 用户注册/最近登录 IP 画像”。
   - 这意味着当前支付面已经从单次支付成功与余额回写，进一步前移到“套餐定义、支付恢复、用户支付页展示、福利兑现、后台治理”一整条链路。
   - 同期进入后台稳定面的还有用户 IP 字段；后续做风控、异常支付、OAuth 注册来源或新用户治理时，不应再把 IP 当成一次性排障字段。

15. 账号级图片输入 URL 化，已经进入当前稳定兼容面
   - 2026-06-21 的图片链路改动说明，当前图片输入能力不应再只按“平台名”或“APIMart 特例”理解，而要按上游账号 `extra` 显式声明的能力决定是否把本地图片改写成对象存储 URL。
   - 当前稳定边界包括：`image_input_transport=object_url`、可选的 `image_upload_limit_bytes`，以及普通 OpenAI-compatible 上游只有在声明支持 `image_urls` / `mask_url` 时才走 JSON URL 字段改写。
   - 这条能力会直接影响账号编辑页、上游兼容性排查、multipart 失败归因与 failover 后的再次请求；后续如果再看到 “Part exceeded maximum size of 1024KB” 一类问题，不应先假设是 Sub2API 全局上传限制。

16. 排行榜 / 模型榜已进入新的当前用户台面
   - 2026-06-24 的主线不再只停在 Studio Bridge、支付治理或账号能力兼容；用户台 `leaderboard` 已新增模型榜、Token 占比、增长百分比和排名变化，并有模型商图标语义。
   - 这说明当前默认产品面已经包含“用户可见数据台”的持续演进，后续如果再看 dashboard/leaderboard，不应继续按旧的单一 Token 榜理解。
   - 模型榜当前稳定边界包括：后端 `model_ranking` 聚合、上一周期对比的 `growth_percent` / `rank_change`、前端榜单卡片内嵌切换，以及移动端右侧指标区响应式堆叠。

17. 教程 CMS / 登录跳转 / 共享额度窗口仍是基础层，但已经更远离当前高频主线
   - 这些能力已经从“当前主线”退到“稳定背景约束”。
   - 当前补知识时，更值得优先解释 OpenAI 网关稳态、账号能力与控制台链路，而不是重复教程页或旧工作区迁移背景。

18. 2026-06-17 的上游小步合成结果，已经进入当前稳定背景层
   - `v0.1.137` 的 S15/S16/S17 不是新的默认产品主线，但已经形成新的稳定工程边界：安全与兼容补丁、计费兜底、thinking 协议过滤、Responses probe 能力校验、API Key ACL IP 拒绝信息，以及 OpenAI OAuth 上游 quota/reset 入口都已落盘。
   - 这些结论之所以值得进入当前焦点，而不是只留在 task 快照里，是因为它们会直接影响后续继续合上游、排查 OpenAI/Anthropic/国产模型兼容、做管理员账户运维或解释为什么某些 patch 可以继续小步迁、某些 migration-heavy 变更仍应跳过。
   - 当前默认心智应是：Studio Bridge / 支付治理仍是产品主链，上游合成则进入“低风险小步、保护本地定制、不 merge 大链路”的稳定工程主线。

19. 2026-06-26 的 S21 / S22 follow-up safe patches，已经把“最近默认续做入口”从 leaderboard 小任务前移到新一轮上游收口
   - S21 已稳定落地 Spark `image_generation` tool strip、OpenAI weekly reset 二次确认、usage cache token 明细展示和邮箱绑定后缀白名单。
   - 当前默认续做不应再停在 6/24 的 leaderboard 视觉/交互语境；更接近事实的是“在不覆盖本地 Studio Bridge、支付和公共页定制的前提下，继续小步吸上游安全/兼容修复”。
   - S22 仍是候选评估，不应误写成“已完成主线”；支付/订阅/余额预扣、order currency、Antigravity fallback、GPT-5.5 instructions fallback、ops chart UI、Claude terminal template、payment supported-types 继续属于跳过或待独立 Sprint 的范围。

20. 2026-07-28 的 `S124 upstream-v0166-config-usage-ui` 已经成为新的上游稳定背景层
   - `CONFIG_FILE` source helper、exact `request_id` list-only filter、route user-label hydration、allowed paths 和 source-only QA 都已落盘。
   - 后续再看 configuration / usage 查询兼容性时，应先按 `S124` 的 exact list predicate 和 config-source fallback 理解，不要再回退到旧的模糊过滤或默认配置假设。

## 已稳定结论

- `knowledge/tasks/current-task.md` 仍适合记录动态交付快照，但当前稳定主线已经不该只停在 2026-06-08 的 gateway auth / prompt cache 语境；最近默认续做心智已先推进到 Studio Bridge、真实用户闭环和落叶AI生产联调层，再进一步前移到 2026-07-08 的前端统一风格、首页认证入口和排行榜/共享账号展示收口层。
- 教程页现在是稳定公共内容链路，不再只是公共前端文案；但如果只看它，会低估最近聊天生图与嵌入工作区改动的影响范围。
- 登录后保留目标跳转仍是默认心智，而且最近又进一步延伸到 embedded session 恢复；后续涉及登录/注册/launch/redeem 和用户页入口时，不应假设“丢 token 后直接重新登录”就是可接受行为。
- 模型广场 `reference_pricing` 当前只用于展示，不写入渠道价格，不影响 billing、倍率和扣费链路；后续如果再扩展模型信息展示，应保持这个边界不被误判。
- 近期提交已经证明：本仓库的“当前高频改动”既包括产品入口，也包括上游合并、OpenAI 网关稳态、账号能力路由和控制台状态面；知识入口必须同时覆盖这些事实。
- OpenWebUI 不再是当前默认用户侧“聊天生图”入口；当前默认入口应进入落叶AI启动链路。
- 充值、用户、余额、默认聊天/生图/视频分组和 bridge internal secret 仍应由 Sub2API 管理后台维护。
- 如果存在 `STUDIO_BRIDGE_LUOYE_AI_INTERNAL_SECRET`，且当前本地配置为空、缺 secret/group、禁用或仍是 `example.com` 占位，系统会自动修复本地 launch return URL、充值回跳 URL 和 `127.0.0.1/localhost` allowed domains；正式域名配置不会被这套本地自修复覆盖。
- 默认生图分组不再硬编码旧 id，而是优先选择第一个 active 且 `allow_image_generation=true` 的 image group；默认聊天分组优先 text group，缺失时可复用 image group。
- `reserve / commit / refund` 的预扣/确认/退款语义已经进入当前默认联调面；后续排查任务失败、余额不足或取消回退时，应优先把它当成默认知识，而不是临时接口细节。
- 首页右侧内嵌认证卡片已经成为当前默认用户入口；后续再改登录/注册、公共首页、回跳或 OAuth 展示时，不应继续按“独立登录页才是主入口”的旧心智理解。
- 暖白/陶土/黑灰主题当前不只是首页样式，而是公共页、认证页、控制台和多数用户/后台弹窗的共享视觉基线；后续 UI 验收不应只盯单页是否能用，还要检查是否回退到旧绿/旧蓝主色体系。
- 共享账号 `channel status` 可见性设置已经进入稳定后台设置面；后续 shared account 用户视图与容量状态解释必须默认考虑这个开关。
- 充值套餐 bonus 当前稳定语义是“首充 only”，不是任意充值包都重复赠送。
- Studio Bridge 当前最小浏览器 smoke 还应确认 session-probe iframe 只打 `/studio-bridge/session-probe`，且 `frame-ancestors` / CSP 对落叶AI父页面放行；如果这里退化，即使 launch token 和 redeem 仍返回 200，用户态也可能表现为“已登录但余额/会话不同步”。
- 首充福利 bonus、注册 IP 和用户福利页已经进入最近稳定产品面；后续改支付、用户列表、福利配置或充值页时，要把它们视为与 Studio Bridge 同期生效的后台默认约束。
- 可配置充值套餐已经进入当前稳定支付面；后续改支付页、后台设置、支付恢复或兑现逻辑时，要默认考虑“套餐来自后台配置”而不是固定面额。
- 管理员用户列表里的注册 IP / 最近登录 IP 已从辅助排障信息升级为稳定后台画像；后续风控、福利发放和异常账户排查应默认读取它们。
- 账号级图片输入 URL 化已经进入当前稳定兼容边界：普通 OpenAI-compatible 上游默认仍走原 multipart；只有账号能力显式要求 object URL 或超过声明的上传阈值时，才应改写输入。
- 上游 `Part exceeded maximum size of 1024KB` 当前应优先理解为“目标上游 multipart part 只有 1MB 限制”，不是 Sub2API 整体图片输入统一只有 1MB；后续排障应先查账号 `extra`、对象存储可达性和 `image_urls / mask_url` 能力声明。
- 后台账号编辑页已经进入这条稳定链路：图片输入 URL 化不再要求手工改数据库，而是可以直接在 OpenAI API Key 账号编辑弹窗配置。
- 2026-06-24 的用户榜单稳定面已经不是“只看 Token 消耗榜”：
  - 模型榜当前已包含模型商图标、Token 占比、增长百分比和排名变化。
  - `Token 消耗榜 / 模型榜` 切换已收进同一排行榜卡片，不应再把它理解成页面外层临时控件。
- 2026-06-26 的默认工程入口也已经不是“只看 leaderboard 是否完成”：
  - Spark image tool strip 必须先于本地图片权限 gate 执行，否则 Codex CLI 默认携带的 tool 会被误判成图片生成意图。
  - OpenAI quota/reset 已不仅是 S17 时的“后台运维面存在”，而是继续收口到 weekly reset 二次确认和 usage cache token 细项展示。
- `sub2api` 与 `chatgpt2api` 仍需分开维护知识：前者偏 gateway、公共入口、嵌入式工作区桥接、模型/计费目录；后者偏独立图片工作台、`/canvas` 节点工作区和 ChatGPT Web 能力封装。
- `use key base url` 归一和 routed API key capabilities 已进入当前稳定主线；后续排查用户“为什么这个 key 看不到某能力/为什么 base URL 表现不一致”时，应优先把它当成默认知识，而不是零散提交细节。
- 默认 API key / 默认分组改造之后，普通更新路径仍应执行 route groups 权限校验；如果未来再看到这块编译或逻辑回退，先检查 `validateAPIKeyRouteGroups(..., false)` 一类调用是否被遗漏。
- 2026-06-17 的上游小步合成已确认以下边界进入稳定知识面：
  - 前端 `form-data` 锁定到 `4.0.6`。
  - token refresh 新增不可重试错误分类。
  - 上游响应支持 zstd；非流式 2xx 非 JSON 和 SSE `event:error` 会进入 failover 并保留原始错误体。
  - tool strict 缺省补 `false`；国产模型 fallback pricing 与图像输入 token 计费补齐。
  - DeepSeek `reasoning_effort=max` 归一到 `xhigh`；Anthropic thinking block 过滤按 mapped upstream model 分流。
  - Responses sticky hash 以 `input` 兜底，Claude Code `max_tokens=1` Haiku 流式探测会被拦截，OpenAI APIKey `/responses` probe 会校验工具能力。
  - OpenAI OAuth usage cell 已支持上游 WHAM quota 查询与 reset credits 操作，但这一块仍属于“小步迁移完成的后台运维面”，不是新产品主线。
- 2026-06-17 的三轮上游 Sprint 本身都明确没有 merge/rebase `upstream/main`，也没有触碰 Ent/migrations/VERSION、Studio Bridge、Canvas、支付页、公共页或模型市场；这条“保护本地定制”的边界只适用于 S15-S17。
- 后续 `main` 已合入统一 API Key、APIMart 图片模型、公共页/模型市场显示和系统设置导航等产品批次，`origin/main..HEAD` 已实际触达 `wire_gen.go`、Studio Bridge repo、公共页、模型市场、`KeysView`、`SettingsView` 等路径；做最近提交复核时必须按当前 diff 重新列证据，不要沿用 S15-S17 的 `NO_DENIED_PATHS`。

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
- 不要把 2026-07-08 的前端统一主题误判成纯视觉换肤；它已经改变了首页认证入口、公共页与控制台的默认心智和验收路径。
- 不要把共享账号渠道状态可见性误判成单页展示逻辑；它已经进入后台设置、用户可见字段和客服解释口径的稳定权限边界。
- 不要把首充福利 bonus 继续理解成宽松运营文案；当前稳定语义已经收口为首次充值才触发。
- 不要把用户 IP 字段当成临时调试字段；它已经进入用户治理、风控和运营判断的默认后台视图。
- 不要把图片输入 URL 化误判成 APIMart 专属分支；当前它是账号能力驱动的兼容层，普通 OpenAI-compatible 上游也可能需要。
- 不要把上游 `Part exceeded maximum size of 1024KB` 误判成 Sub2API 本地统一上传上限；当前更多是在提示目标上游 multipart part 限制与账号能力配置不匹配。
- 不要把 2026-06-24 的 leaderboard 改动误判成纯视觉小修；它已经引入新的用户侧稳定数据语义，包括模型榜、增长、排名变化和榜单内嵌切换。
- 也不要把 2026-06-24 的 leaderboard 改动继续误判成“当前默认续做入口”；截至 2026-06-26，更接近主线的是 S21 已完成、S22 待评估的上游 follow-up safe patches。
- 不要把落叶AI团队空间联调误判成 chatgpt2api 单仓库任务；当前团队空间的 actor/payer、余额和扣费真源仍在 Sub2API。
- 不要把 2026-06-17 的上游 patch 误判成“已经可以整体跟上游合并”；当前稳定策略仍是按 Sprint 做低风险小步迁移，并显式避开 migration-heavy、合规门禁或会覆盖本地定制的大链路。
- 不要把 OpenAI quota/reset、thinking filter、Responses probe 或 zstd 支持当成零散实现细节；它们已经影响后续排障基线和上游 patch 取舍。
- 不要只看 `docs/ai/current-task.md`；该文件仍是兼容旧入口的说明，当前事实应以 `knowledge/`、`knowledge/tasks/current-task.md` 和时间轴为主。
- 不要把本地 fake 演示账号或 fallback 教程内容当成正式生产数据；它们仍主要服务本地预览、升级兜底或空态保护。
