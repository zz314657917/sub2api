# 聊天生图与嵌入工作区链路

最后更新：2026-05-30

## 适用范围

- `sub2api` 用户从公共页、用户控制台或聊天生图入口进入图片工作区。
- OpenWebUI / ChatImage launch、redeem、嵌入式登录态恢复、模型目录与定价同步。
- 需要判断“为什么一个登录/模型/图片任务改动会同时影响聊天生图工作区”时。

## 一句话心智

当前 `sub2api` 不能只按“教程 CMS + 容量池 + 控制台”理解；更高频的默认产品链路已经变成“公共入口/控制台 -> launch token -> 嵌入工作区登录态 -> 聊天生图或 Canvas 工作区 -> 图片任务/模型目录/计费映射”。

## 默认产品链路

1. 用户从 `sub2api` 公共页或控制台进入聊天生图/相关工作区入口。
2. 前端通过 OpenWebUI / ChatImage launch 链路生成一次性跳转能力。
3. 目标工作区用 redeem 或等价桥接方式换取本地登录态，而不是要求用户重新维护一套独立账号。
4. 登录态建立后，用户继续在嵌入式或被跳转的图片工作区里发起图片相关任务。
5. 图片任务、模型白名单、定价展示和计费映射继续复用现有 gateway / billing 体系，而不是在工作区内各自维护一套逻辑。

## 这条链路为什么值得单独记

- 它跨越公共页、认证、launch/redeem、用户工作区、模型目录、计费和图片任务，不是单页面知识。
- 近期高频改动已经证明，改登录恢复、改模型定价、改工作区导航，都会反向影响这条链路的连续体验。
- 如果只看 `current-task.md` 或单个提交，很容易知道“改了什么”，但不知道“它属于哪条稳定产品链路”。

## 稳定约束

### 1. launch / redeem 不是附属小功能

- 它们已经是工作区进入方式的一部分，不是纯实验入口。
- 继续改聊天生图、嵌入式工作区或用户侧 launch 行为时，要同时考虑 token 生命周期、cookie/session 写入和恢复语义。

### 2. embedded session recovery 是默认可用性要求

- 近期稳定修复已经把“嵌入模式下 stale token 或 cookie 失配后如何恢复会话”提升为默认约束。
- 后续如果只按普通登录页思路处理匿名态回退，容易破坏从 `sub2api` 跳入工作区后的连续体验。

### 3. 模型目录与定价同步是一条完整链路

- 新增或调整模型时，不能只看前端展示卡片。
- 至少要一起理解默认模型列表、fallback pricing、billing family、第三方映射、模型广场排序和对应测试。
- `claude-opus-4.8` 这类新增模型就是近期稳定样例。

### 4. 图片任务仍应视为平台链路的一部分

- 无论用户看到的是聊天生图、嵌入工作区还是图片管理界面，底层仍会落回图片任务、计费和用量归属体系。
- 不要把“工作区 UI”和“图片任务/计费逻辑”拆成两套完全独立的认知。

## 常见误判

- 不要把当前主线误判成还停在教程 CMS / 共享额度窗口收口。
- 不要把 launch / redeem 当成可随意调整的桥接胶水；它已经影响用户登录连续性。
- 不要把 embedded session recovery 当成单纯 cookie 小修；它实际影响工作区可用性。
- 不要把模型补录当成纯前端文案更新；它通常牵涉后端 pricing 和映射。

## 推荐补读路径

- 仓库级入口：
  - `knowledge/00-start-here.md`
  - `knowledge/05-current-focus.md`
  - `docs/ai/current-task.md`
- 任务快照：
  - `knowledge/tasks/current-task.md`
  - `knowledge/tasks/timeline.md`
- 相关代码面：
  - `frontend/src/api/openWebUI.ts`
  - `frontend/src/router/index.ts`
  - `frontend/src/views/user/`
  - `backend/internal/service/openai_images.go`
  - `backend/internal/service/open_webui_launch_service.go`

## 最小验证建议

- 改 launch / redeem / 登录恢复：至少补前端构建、相关测试和一次工作区入口人工回读。
- 改模型目录/定价：至少补展示侧测试、pricing 数据校验和后端映射/计费测试。
- 改聊天生图或嵌入工作区：不要只验 `sub2api` 单仓库；涉及双仓工作区时，要同步回看目标工作区仓库的最小验证面。
