---
repo: sub2api
project_type: generic
qa_mode: runtime
last_verified: pending
---

# Product Spec

## 一句话需求
- 实现截图风格完整工单系统：用户可在 `/tickets` 与客服会话留言，管理员可在 `/admin/tickets` 管理、回复、关闭、重开工单，并可从用户列表主动给用户发起留言。

## 目标与非目标
- 目标：
  - 用户侧提供双栏工单会话页，支持创建工单、查看消息、回复、标记已读、关闭和刷新。
  - 管理员侧提供工单管理页，支持状态/未读/用户搜索筛选，查看详情、回复、标记已读、关闭、重开。
  - 管理员用户列表增加“留言/工单”入口，可对指定用户主动创建工单消息。
  - 后端持久化工单和消息，并保证用户只能访问自己的工单、管理员接口仅管理员可用。
  - 未读计数和状态随用户/管理员发言稳定变化。
- 非目标：
  - v1 不支持附件。
  - v1 不支持 WebSocket、SSE、邮件通知或自动轮询。
  - v1 不实现客服分配、优先级、标签、内部备注和 SLA。

## 后端方案
- 新增表 `support_tickets` 和 `support_ticket_messages`，通过 Ent schema、migration、repository、service、handler 和 routes 接入。
- 工单状态：
  - `open`：新建或重开。
  - `pending_admin`：用户最后发言，等待管理员。
  - `pending_user`：管理员最后发言，等待用户。
  - `closed`：已关闭。
- 用户 API：
  - `GET /api/v1/user/tickets`
  - `POST /api/v1/user/tickets`
  - `GET /api/v1/user/tickets/:id`
  - `POST /api/v1/user/tickets/:id/messages`
  - `POST /api/v1/user/tickets/:id/read`
  - `POST /api/v1/user/tickets/:id/close`
- 管理员 API：
  - `GET /api/v1/admin/tickets`
  - `GET /api/v1/admin/tickets/:id`
  - `POST /api/v1/admin/tickets/:id/messages`
  - `POST /api/v1/admin/tickets/:id/read`
  - `POST /api/v1/admin/tickets/:id/close`
  - `POST /api/v1/admin/tickets/:id/reopen`
  - `POST /api/v1/admin/users/:id/tickets`

## 前端方案
- 用户侧新增 `frontend/src/views/user/TicketsView.vue`，路由 `/tickets`，侧边栏显示“工单服务”。
- 后台新增 `frontend/src/views/admin/TicketsView.vue`，路由 `/admin/tickets`，侧边栏显示“工单管理”。
- 新增 API/type：
  - `frontend/src/api/tickets.ts`
  - `frontend/src/api/admin/tickets.ts`
  - 更新 `frontend/src/api/index.ts`
  - 更新 `frontend/src/api/admin/index.ts`
  - 更新 `frontend/src/types/index.ts`
- 用户列表主动留言：
  - `frontend/src/components/admin/user/UserTicketMessageModal.vue`
  - `frontend/src/views/admin/UsersView.vue` 只增加入口和弹窗挂载。
- i18n 增加 `nav.tickets`、`nav.ticketManagement`、`tickets.*`、`admin.tickets.*`。

## 验收标准
- 用户 A 不能读或回复用户 B 的工单。
- 管理员能查看、回复、关闭、重开任意工单。
- 空标题、空内容、超长内容返回稳定 4xx。
- 未读计数随用户/管理员消息正确变化。
- 工单关闭后用户回复被拒绝，管理员重开后可继续回复。
- 前端列表、详情、发送、关闭、重开、空态、错误态可用。

## Sprint 计划
- Sprint 1：完整工单系统 v1，后端持久化/API 与前端用户/后台页面同步完成，QA 以 API 权限、状态流转和页面 smoke 为主。
