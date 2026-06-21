# 当前任务快照

最后更新：2026-06-21 11:02 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 本轮任务是普通账号级图片输入 URL 化：不按平台名或 APIMart 名称判断，只按上游账号 `extra` 的图片输入能力配置决定是否把本地图片转对象存储 URL。
- 问题来源是部分上游账号返回 `Part exceeded maximum size of 1024KB`，这属于上游 1MB multipart part 限制；Sub2API 本地默认并不是整体 1MB 限制。

## 当前目标

- 对有 1MB 上传限制的普通上游账号，通过账号 `extra` 启用对象存储 URL 化。
- 普通账号默认行为保持不变；failover 切账号时必须按新账号能力重新处理输入，不能复用上一账号的改写结果。

## 本次已完成

- 新增账号 `extra` 通用字段识别：
  - `image_input_transport: "object_url"`：启用对象存储 URL 化。
  - `image_upload_limit_bytes`：可选，超过限制时触发 URL 化。
  - `image_url_fields_supported`：普通 OpenAI-compatible 上游显式声明支持 `image_urls` / `mask_url` 后才走 JSON URL 字段改写。
- 接入现有 S3/COS 对象存储和 presigned URL，图片输入临时对象默认 URL 有效期为 `2h`。
- APIMart/兼容 `image_urls` 路径在启用 object URL 策略后会把 multipart 图片、mask、JSON data URL 上传到对象存储，再提交 `image_urls` / `mask_url`。
- 已经是 `http/https` URL 的输入原样透传，不重复上传。
- OAuth 图片 Responses 路径也会在账号策略触发时基于克隆后的 parsed request 转 URL，split 请求使用同一份账号级重算结果。
- APIMart object URL 准备阶段如果中途失败，会清理已创建的临时对象；对象 key 带 UUID，避免相同图片并发请求互相删除临时对象。
- 保留上游 `Part exceeded maximum size of 1024KB` 错误归一，客户会看到这是上游 1MB 限制。
- 后台账号编辑弹窗已给 OpenAI API Key 账号新增“图片输入 URL 化”配置区，可直接写入上述三个 `extra` 字段，不再需要手工改数据库。

## 已确认事实

- 未配置 `image_input_transport` / `image_url_fields_supported` 的普通 OpenAI-compatible 账号不改写 multipart，避免破坏未知兼容上游。
- APIMart async image 路径仍保留默认 `/v1/uploads/images` 上传行为；只有账号策略启用 object URL 时才跳过上游 multipart 上传。
- 服务启动注入了 `BackupObjectStoreFactory`，图片输入临时对象复用现有对象存储配置。

## 待验证点

- 生产或 staging 上目标上游必须能公网访问对象存储 presigned URL。
- 真实受限账号需要在后台账号编辑页开启“图片输入 URL 化”，普通兼容上游还需要勾选“上游支持 image_urls / mask_url”。
- 若仅配置 `image_upload_limit_bytes=1048576`，只有本地输入超过该阈值时才触发 URL 化；是否要强制所有本地图片都 URL 化，应按账号能力再确认。

## 当前结论

- 这轮已按“账号能力”而不是“平台名”实现，APIMart 只是已知支持 `image_urls` 的路径之一。
- 对没有显式 URL 字段能力的普通上游，仍保持原 multipart 行为；上游 1MB 报错会继续归一提示客户这是上游限制。

## 下一步

- 配置受 1MB 限制的普通账号：动作 -> 后台账号管理编辑该 OpenAI API Key 账号，开启“图片输入 URL 化”，上传限制字节数填 `1048576`，确认支持 URL 字段后勾选 `image_urls / mask_url`；验证 -> 用大于 1MB 的本地图请求确认上游 body 走 `image_urls`。
- 部署前确认对象存储公网可达：动作 -> 用生成的 presigned URL 从上游可访问网络发起 HEAD/GET；验证 -> 返回 200 且不过期。

## 验证记录

- `cd F:/mcplugins/sub2api/backend && go test ./internal/service -run "TestOpenAIGatewayServiceForwardImages_APIMart|TestOpenAIGatewayServiceForwardImages_CompatibleObjectURLTransport|TestPrepareOpenAIImagesObjectURLInputs|TestPrepareAPIMartImageInputsObjectURLTransportCleansObjectsOnError|TestOpenAIImagesUpstreamImageTooLarge" -count=1` 通过。
- `cd F:/mcplugins/sub2api/backend && go test ./...` 通过。
- `cd F:/mcplugins/sub2api/frontend && npm.cmd run test:run -- src/components/account/__tests__/EditAccountModal.spec.ts` 通过。
- `cd F:/mcplugins/sub2api/frontend && npm.cmd run typecheck -- --pretty false` 通过。
- `cd F:/mcplugins/sub2api/frontend && npm.cmd run build` 通过，仅有既有 chunk、Browserslist 和 Node deprecation 警告。
- `cd F:/mcplugins/sub2api/frontend && npm.cmd run lint` 通过。
- `cd F:/mcplugins/sub2api && git diff --check` 通过。
