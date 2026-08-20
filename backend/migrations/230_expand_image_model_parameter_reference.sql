WITH tutorial_references (slug, reference_md) AS (
    VALUES
        ('gpt-image-2', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `gpt-image-2`。
- `prompt`: 必填，非空字符串；建议写明主体、场景、构图、风格、光线和禁止项。
- `size`: 可选，使用 `auto`、`1:1`、`3:2`、`2:3`、`4:3`、`3:4`、`5:4`、`4:5`、`16:9`、`9:16`、`2:1`、`1:2`、`3:1`、`1:3`、`21:9`、`9:21`，或 `宽x高` 像素串；`size` 只表示画幅比例。
- `resolution`: 可选，仅使用小写 `1k`、`2k`、`4k`；未传或不识别时本地网关按 `1k` 转发，先用 `1k` 调试，再提高到 `2k` 或 `4k`。
- `n`: 可选，正整数；推荐传 `1`。当传入大于 `1` 时，网关可能拆成多次上游提交，应按多份结果和多次计费处理，不要假设一次请求只产生一项任务。
- `image_urls`: 可选，字符串数组，用于图生图；元素可为服务端可访问的图片 URL 或完整 `data:image/...;base64,...` 数据 URI。先传 1 张并保持图片稳定可访问；参考图数量、体积和格式限制由实际可用渠道返回。

### 成功返回

同步完成时通常返回 OpenAI 兼容结构：`data` 是数组，每个元素的 `url` 是生成结果。消费端应先检查 HTTP 为 2xx，再遍历 `data`，不要只读取第一项。部分渠道会返回任务对象；发现 `data[0].task_id` 时改用 `GET /v1/tasks/{task_id}` 轮询，成功后从 `data.result.images[0].url[0]` 取得图片。

### 错误状态与处理

错误主体通常为 `{"error":{"type":"...","message":"..."}}`；上游已经返回错误时，状态码、`type` 和 `message` 可能会原样透传。`400 invalid_request_error` 表示 JSON、模型名、`n` 类型或字段组合不正确；`401 authentication_error` 表示 API Key 无效；`402 payment_required` 表示余额不足；`403 permission_error` 表示该分组未开启图像能力；`413 invalid_request_error` 表示请求体或上传图片过大；`429 rate_limit_error` 表示并发或频率受限；`502 upstream_error`、`503 api_error` 表示暂时没有可用上游。对 `429`、`502`、`503` 使用指数退避重试；其他 4xx 先修正请求，不要盲目重试。
$reference$),
        ('gpt-image-2-official', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `gpt-image-2-official`。
- `prompt`: 必填，非空字符串；局部编辑时同时说明保留部分和要替换的蒙版区域。
- `size`: 可选，使用 `auto`、`1:1`、`3:2`、`2:3`、`4:3`、`3:4`、`5:4`、`4:5`、`16:9`、`9:16`、`2:1`、`1:2`、`3:1`、`1:3`、`21:9`、`9:21`，或 `宽x高` 像素串。
- `resolution`: 可选，仅使用小写 `1k`、`2k`、`4k`；`4k` 加 `quality: "high"` 的耗时和费用最高。
- `n`: 可选，整数 `1` 到 `4`；结果在 `data` 数组中逐项返回。
- `quality`: 可选，`auto`、`low`、`medium`、`high`；先选 `auto` 或 `low` 验证提示词，交付图再选 `medium` 或 `high`。
- `background`: 可选，`auto` 或 `opaque`；本模型不把透明背景作为稳定能力，传 `transparent` 不应作为业务依赖。
- `moderation`: 可选，`auto` 或 `low`；它是审核强度，不是绕过审核的开关。
- `output_format`: 可选，`png`、`jpeg`、`webp`；需要无损或后续编辑时用 `png`，网页展示可用 `jpeg` 或 `webp`。
- `output_compression`: 可选，整数 `0` 到 `100`；只在 `output_format` 为 `jpeg` 或 `webp` 时传入。
- `image_urls`: 可选，参考图 URL 或完整数据 URI 数组；引用图应可访问，建议先从 1 张开始验证。
- `mask_url`: 可选，蒙版图片 URL；必须配合 `image_urls` 使用，并与首张参考图尺寸一致、带 Alpha 通道。

### 成功返回

同步完成时读取 `data[].url`。若渠道返回 `task_id`，轮询 `GET /v1/tasks/{task_id}`，只在 `completed`、`succeeded` 或 `success` 状态读取 `data.result.images[].url[]`，结果链接应尽快下载或转存。

### 错误状态与处理

错误主体通常为 `{"error":{"type":"...","message":"..."}}`。`400 invalid_request_error` 常见于 `size`、`resolution`、压缩值、格式或蒙版组合错误；先移除最近新增的可选字段并逐项加回。`401 authentication_error` 检查 API Key；`402 payment_required` 检查余额；`403 permission_error` 检查图像权限；`413 invalid_request_error` 压缩或改用 URL 引用图片；`429 rate_limit_error` 降低并发后退避重试；`502 upstream_error`、`503 api_error` 稍后重试。安全审核拒绝属于 4xx，必须修改提示词或素材，不能自动重试。
$reference$),
        ('gemini-3-pro-image-preview', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `gemini-3-pro-image-preview`。
- `prompt`: 必填，非空字符串；多参考图时写清每张图负责主体、配色、材质还是构图。
- `size`: 可选，使用 `auto`、`1:1`、`2:3`、`3:2`、`3:4`、`4:3`、`4:5`、`5:4`、`9:16`、`16:9`、`21:9`，或受支持的 `宽x高` 像素串。
- `resolution`: 可选，仅使用小写 `1k`、`2k`、`4k`；先以 `1k` 或 `2k` 确认参考图关系。
- `n`: 必须显式传 `1`；需要多个方案时使用不同提示词分别提交。
- `image_urls`: 可选，最多 `14` 个 URL 或完整数据 URI；数组顺序就是参考优先级，数量超过 14 会直接产生参数错误。

### 成功返回

本页模型以同步最终图为主，从 `data[0].url` 读取结果；程序仍应按数组处理。若返回任务对象，按 `task_id` 轮询而不是重复提交，终态失败时读取任务错误信息。

### 错误状态与处理

错误主体通常为 `{"error":{"type":"...","message":"..."}}`。`400 invalid_request_error` 多数是模型 ID、`n`、分辨率或参考图数量不合法；`401 authentication_error` 是密钥错误；`402 payment_required` 是余额不足；`403 permission_error` 是图像能力未授权；`429 rate_limit_error` 降低并发后重试；`502 upstream_error`、`503 api_error` 为上游暂不可用。画面偏离不是 API 错误，应减少互相冲突的参考图并在 `prompt` 中明确优先级。
$reference$),
        ('gemini-3-pro-image-preview-official', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `gemini-3-pro-image-preview-official`。
- `prompt`: 必填，非空字符串；先写交付用途和构图，再写参考图中必须保留的特征。
- `size`: 可选，使用 `auto`、`1:1`、`2:3`、`3:2`、`3:4`、`4:3`、`4:5`、`5:4`、`9:16`、`16:9`、`21:9`，或受支持的 `宽x高` 像素串。
- `resolution`: 可选，仅使用小写 `1k`、`2k`、`4k`；不要将 `2k` 或 `4k` 写入 `size`。
- `n`: 必须显式传 `1`；该兼容路径不支持通过增大 `n` 批量获取候选。
- `image_urls`: 可选，最多 `14` 个 URL 或完整数据 URI；把最关键的产品或主体图放在前面。

### 成功返回

成功时从 `data[0].url` 取图；若响应为任务对象，保存 `task_id` 并轮询。不要把 HTTP 2xx 的“已提交”误当成图片已经可下载。

### 错误状态与处理

`400 invalid_request_error` 表示模型、`n`、尺寸、清晰度或参考图输入有误；`401 authentication_error` 检查 API Key；`402 payment_required` 检查余额；`403 permission_error` 检查分组图像权限；`429 rate_limit_error` 退避重试；`502 upstream_error`、`503 api_error` 稍后重试。错误响应通常是 `{"error":{"type":"...","message":"..."}}`，记录 `type` 和请求 ID（如有），但不要记录密钥或完整数据 URI。
$reference$),
        ('gemini-3-1-flash-image-preview', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `gemini-3.1-flash-image-preview`。
- `prompt`: 必填，非空字符串；开启搜索后仍要描述最终画面，不能只写搜索关键词。
- `size`: 可选，使用 `auto`、`1:1`、`3:2`、`2:3`、`4:3`、`3:4`、`5:4`、`4:5`、`16:9`、`9:16`、`21:9`、`1:4`、`4:1`、`1:8`、`8:1`，或受支持的 `宽x高` 像素串。
- `resolution`: 可选，仅使用小写 `1k`、`2k`、`4k`；`0.5k` 不是本地兼容输入，未识别取值会按 `1k` 转发。
- `n`: 必须显式传 `1`。
- `image_urls`: 可选，最多 `14` 个 URL 或完整 `data:image/...;base64,...` 数据 URI；URL 和数据 URI 可以混用。
- `google_search`: 可选，布尔值 `true` 或 `false`；仅在图像内容确实依赖最新事实时使用。
- `google_image_search`: 可选，布尔值 `true` 或 `false`；只在已启用 `google_search` 且需要外部视觉线索时使用。

### 成功返回

成功时从 `data[0].url` 读取最终图；如果返回 `task_id`，以 3 秒左右间隔轮询 `GET /v1/tasks/{task_id}`，并在终态后读取图片链接。

### 错误状态与处理

`400 invalid_request_error` 常见于布尔值写成字符串、`n` 不为 `1`、`0.5k`、参考图超过 14 张或图片格式无效；`401 authentication_error`、`402 payment_required`、`403 permission_error` 分别对应密钥、余额和图像权限；`429 rate_limit_error` 需要退避；`502 upstream_error`、`503 api_error` 需要稍后重试。搜索导致内容不稳定不是服务错误，先关闭一个搜索开关，再检查参考图与提示词优先级。
$reference$),
        ('gemini-3-1-flash-image-preview-official', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `gemini-3.1-flash-image-preview-official`。
- `prompt`: 必填，非空字符串；明确搜索只决定事实内容，参考图只决定视觉要素，避免两类来源互相冲突。
- `size`: 可选，使用 `auto`、`1:1`、`3:2`、`2:3`、`4:3`、`3:4`、`5:4`、`4:5`、`16:9`、`9:16`、`21:9`、`1:4`、`4:1`、`1:8`、`8:1`，或受支持的 `宽x高` 像素串。
- `resolution`: 可选，仅使用小写 `1k`、`2k`、`4k`；`0.5k` 不属于本地兼容取值。
- `n`: 必须显式传 `1`。
- `image_urls`: 可选，最多 `14` 个 URL 或完整数据 URI；传入前检查每个链接可访问。
- `google_search`: 可选，布尔值 `true` 或 `false`；默认应省略或传 `false`。
- `google_image_search`: 可选，布尔值 `true` 或 `false`；仅在 `google_search: true` 时考虑开启。

### 成功返回

成功响应的图片在 `data[0].url`；若是异步任务，保存 `task_id`，重复查询而不是重复生成。任务状态为 `failed`、`cancelled` 或 `canceled` 时停止轮询并记录其错误信息。

### 错误状态与处理

`400 invalid_request_error` 多由字段类型、`n`、`resolution` 或参考图数量引起；`401 authentication_error` 检查密钥；`402 payment_required` 检查余额；`403 permission_error` 检查权限；`429 rate_limit_error` 减少并发；`502 upstream_error`、`503 api_error` 延迟重试。错误主体通常包含 `error.type` 和 `error.message`，程序应该按状态码分类处理，而不能只匹配中文或英文消息。
$reference$),
        ('midjourney', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `midjourney`；请求路径必须是 `POST /v1/midjourney/generations`。
- `prompt`: 必填，非空字符串；可使用原生提示词参数，但与结构化字段冲突时以结构化字段为准。
- `size`: 可选，比例字符串，例如 `1:1`、`16:9`、`9:16`、`3:2`、`2:3`、`4:3`、`3:4`；等价于提示词中的 `--ar`。
- `version`: 可选，推荐 `5`、`5.1`、`5.2`、`6`、`6.1`；本地 `stop` 只会在这些版本上转发。其他版本是否可用由实际渠道决定。
- `speed`: 可选，`relax`、`fast`、`turbo`；不传等同默认 `relax`。
- `quality`: 可选，`0.25`、`0.5`、`1`、`2`；需与所选 `version` 兼容。
- `stylize`: 可选，整数 `0` 到 `1000`；值越高越偏向模型审美。
- `chaos`: 可选，整数 `0` 到 `100`；值越高候选差异越大。
- `weird`: 可选，整数 `0` 到 `3000`；仅适合实验性创作。
- `stop`: 可选，整数 `10` 到 `100`；仅用于支持它的 `version`，否则网关不会转发。
- `niji`: 可选，布尔值；动漫风格时传 `true` 并配合兼容版本。
- `raw`: 可选，布尔值；用于减少默认风格化干预，建议在支持的版本上使用。
- `tile`: 可选，布尔值；仅用于无缝平铺纹理。
- `image_urls`: 可选，最多 `4` 个 URL 或完整数据 URI；用于垫图或融合，超过 4 个会产生参数错误。

### 成功返回

生成结果通常位于 `data[].url`；如返回 `task_id`，轮询 `GET /v1/tasks/{task_id}` 直到终态。生成任务失败时，状态对象的错误消息比提交请求的 HTTP 2xx 更有诊断价值。

### 错误状态与处理

`400 invalid_request_error` 表示 `prompt` 缺失、参数类型不对、参考图超过 4 张或版本与控制项不兼容；`401 authentication_error` 是密钥错误；`402 payment_required` 是余额不足；`403 permission_error` 是权限不足；`404 not_found_error` 多为不存在或无权访问的任务；`429 rate_limit_error` 需要降低并发；`502 upstream_error`、`503 api_error` 表示通道不可用。任务 `failed` 时读取 `error.message` 或失败原因；违禁提示词需改写，任务超时或没有可用通道可在退避后重新提交。
$reference$),
        ('doubao-seedance-4-0', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `doubao-seedance-4-0`。
- `prompt`: 必填，非空字符串；写清要保留的参考图特征和需要生成的场景。
- `size`: 可选，使用画幅比例或受支持的 `宽x高` 像素串；先从 `1:1`、`4:3`、`16:9`、`9:16` 这类常用比例开始。
- `resolution`: 可选，仅使用小写 `1k`、`2k`、`4k`；先用 `1k` 验证构图和输入素材。
- `n`: 可选，正整数；`n` 加上 `image_urls` 的数量最多为 `15`。
- `image_urls`: 可选，URL 或完整数据 URI 数组；其数量与 `n` 合计不得超过 `15`，否则请求在提交前返回参数错误。

### 成功返回

这是异步流程。提交成功后保存 `data[0].task_id`，以约 3 秒间隔调用 `GET /v1/tasks/{task_id}`。只有状态为 `completed`、`succeeded` 或 `success` 才读取 `data.result.images[0].url[0]`；链接可能会过期，应尽快下载或转存。

### 错误状态与处理

`400 invalid_request_error` 常见于模型、提示词、`n` 或“参考图数加输出数超过 15”；`401 authentication_error`、`402 payment_required`、`403 permission_error` 分别对应密钥、余额和图像权限；`429 rate_limit_error` 需退避；`502 upstream_error`、`503 api_error` 稍后重试。轮询时 `failed`、`cancelled`、`canceled` 都是终态，读取任务 `error.message` 后停止轮询；不要因未完成就重复提交同一任务。
$reference$),
        ('doubao-seedance-4-5', $reference$
## 字段取值、返回和错误处理

### 可用字段与取值

- `model`: 必填，固定为 `doubao-seedance-4-5`。
- `prompt`: 必填，非空字符串；高分辨率任务应把材质、镜头、光线和不可改变的品牌元素写成可观察的要求。
- `size`: 可选，使用画幅比例或受支持的 `宽x高` 像素串；画幅和清晰度由 `size`、`resolution` 分别控制。
- `resolution`: 可选，建议只使用小写 `2k` 或 `4k`；先用 `2k` 定稿构图，再用 `4k` 生成最终交付图。
- `n`: 可选，正整数；与 `image_urls` 的数量合计最多 `15`。
- `image_urls`: 可选，URL 或完整数据 URI 数组；传不可替代的产品或品牌参考图，并使数量加 `n` 不超过 `15`。

### 成功返回

这是异步流程。提交后从 `data[0].task_id` 保存任务 ID，间隔约 3 秒查询 `GET /v1/tasks/{task_id}`。应用重启后应恢复已有任务的轮询；当状态为 `completed`、`succeeded` 或 `success` 时，从 `data.result.images[0].url[0]` 取得结果。

### 错误状态与处理

`400 invalid_request_error` 表示参数、清晰度或输入/输出总数错误；`401 authentication_error` 检查 API Key；`402 payment_required` 检查余额；`403 permission_error` 检查图像权限；`429 rate_limit_error` 降低并发并退避；`502 upstream_error`、`503 api_error` 稍后重试。任务为 `failed`、`cancelled` 或 `canceled` 时，保留 `task_id` 与 `error.message` 用于排查，但不要无限轮询或重复扣费式地反复提交。
$reference$)
)
UPDATE tutorial_pages AS page
SET content_md = regexp_replace(
        page.content_md,
        E'\r?\n\r?\n## cURL',
        E'\n\n' || btrim(tutorial_references.reference_md, E'\r\n') || E'\n\n## cURL',
        's'
    ),
    updated_at = NOW()
FROM tutorial_references
WHERE page.slug = tutorial_references.slug
  AND page.content_md ~ E'## AI 调用清单'
  AND page.content_md ~ E'## cURL';
