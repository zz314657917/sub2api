INSERT INTO tutorial_pages (slug, title, description, category, sort_order, status, content_md, published_at)
VALUES
('gpt-image-2', 'GPT Image 2 图像生成', '使用 GPT Image 2 生成图像。', '图像模型', 2240, 'published', $md$
# GPT Image 2 图像生成

模型 ID：`gpt-image-2`。请求 `POST https://ai.3zapi.top/v1/images/generations`，并设置 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `gpt-image-2`
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`；tier 使用 `resolution`，例如 `1k`、`2k`、`4k`。
- `n`: 生成数量。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"gpt-image-2","prompt":"一只红色风筝","size":"1024x1024","resolution":"1k","n":1}'
```

## Python

```python
import requests
response = requests.post("https://ai.3zapi.top/v1/images/generations", headers={"Authorization": "Bearer YOUR_API_KEY"}, json={"model": "gpt-image-2", "prompt": "一只红色风筝", "size": "1024x1024", "resolution": "1k", "n": 1})
print(response.json()["data"][0]["url"])
```

## JavaScript

```javascript
const response = await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers: {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"}, body: JSON.stringify({model: "gpt-image-2", prompt: "一只红色风筝", size: "1024x1024", resolution: "1k", n: 1})});
console.log((await response.json()).data[0].url);
```

## 响应与排错

网关会等待任务完成，直接返回 OpenAI 兼容的最终响应：`{"created": 0, "data": [{"url": "..."}]}`；无需轮询提交请求返回的上游任务。收到 `401` 时检查 Bearer 密钥，收到参数错误时检查 `model`、`prompt`、`size` 和 `n`。
$md$, NOW()),
('gpt-image-2-official', 'GPT Image 2 官方图像生成', '使用 GPT Image 2 官方路径生成或编辑图像。', '图像模型', 2241, 'published', $md$
# GPT Image 2 官方图像生成

模型 ID：`gpt-image-2-official`。请求 `POST https://ai.3zapi.top/v1/images/generations`，认证头为 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `gpt-image-2-official`；`prompt` 为图像描述，`size` 使用像素尺寸或比例，tier 使用 `resolution`。
- `n`: 生成数量。
- 可选：`quality`、`background`、`moderation`、`output_format`、`output_compression`、`mask_url`。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"gpt-image-2-official","prompt":"一张干净的产品图","size":"1024x1024","resolution":"1k","n":1,"quality":"high","background":"auto","output_format":"png"}'
```

## Python

```python
import requests
payload = {"model": "gpt-image-2-official", "prompt": "一张干净的产品图", "size": "1024x1024", "resolution": "1k", "n": 1, "quality": "high"}
response = requests.post("https://ai.3zapi.top/v1/images/generations", headers={"Authorization": "Bearer YOUR_API_KEY"}, json=payload)
print(response.json()["data"][0]["url"])
```

## JavaScript

```javascript
const payload = {model: "gpt-image-2-official", prompt: "一张干净的产品图", size: "1024x1024", resolution: "1k", n: 1, output_format: "png"};
const response = await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers: {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"}, body: JSON.stringify(payload)});
console.log((await response.json()).data[0].url);
```

## 响应与排错

网关等待完成后返回 `{"created": 0, "data": [{"url": "..."}]}`，不要对提交响应轮询上游任务。`401` 请检查 Bearer 密钥；可选项报错时先移除不支持的组合，再确认 `mask_url` 是可访问的图像 URL。
$md$, NOW()),
('gemini-3-pro-image-preview', 'Gemini 3 Pro Image Preview 图像生成', '使用 Gemini 3 Pro Image Preview 生成图像。', '图像模型', 2242, 'published', $md$
# Gemini 3 Pro Image Preview 图像生成

模型 ID：`gemini-3-pro-image-preview`。使用 `POST https://ai.3zapi.top/v1/images/generations` 和 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `gemini-3-pro-image-preview`；`prompt` 为描述，`size` 使用像素尺寸或比例，`resolution` 支持 `1k`、`2k`、`4k`。
- `image_urls` 可传参考图，最多 14 张；兼容流程中 `n` 固定为 `1`。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"gemini-3-pro-image-preview","prompt":"水彩山谷","size":"1024x1024","resolution":"2k","n":1}'
```

## Python

```python
import requests
payload = {"model": "gemini-3-pro-image-preview", "prompt": "水彩山谷", "size": "1024x1024", "resolution": "2k", "n": 1}
response = requests.post("https://ai.3zapi.top/v1/images/generations", headers={"Authorization": "Bearer YOUR_API_KEY"}, json=payload)
print(response.json()["data"][0]["url"])
```

## JavaScript

```javascript
const response = await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers: {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"}, body: JSON.stringify({model: "gemini-3-pro-image-preview", prompt: "水彩山谷", size: "1024x1024", resolution: "2k", n: 1})});
console.log((await response.json()).data[0].url);
```

## 响应与排错

请求会等待完成并返回 `{"created": 0, "data": [{"url": "..."}]}`，不需要轮询提交调用。`401` 表示 Bearer 密钥无效；参考图超过 14 张或 `resolution` 不是 `1k`、`2k`、`4k` 时请缩减或修正参数。
$md$, NOW()),
('gemini-3-pro-image-preview-official', 'Gemini 3 Pro Image Preview 官方图像生成', '使用 Gemini 3 Pro Image Preview 官方路径生成图像。', '图像模型', 2243, 'published', $md$
# Gemini 3 Pro Image Preview 官方图像生成

模型 ID：`gemini-3-pro-image-preview-official`。请求地址为 `POST https://ai.3zapi.top/v1/images/generations`，使用 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `gemini-3-pro-image-preview-official`；必填 `prompt`。
- `size`: 像素尺寸或比例；`resolution` 使用 `1k`、`2k` 或 `4k`；兼容流程中 `n` 固定为 `1`。
- `image_urls`: 可选参考图，最多 14 张。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"gemini-3-pro-image-preview-official","prompt":"未来城市夜景","size":"1024x1024","resolution":"4k","n":1}'
```

## Python

```python
import requests
response = requests.post("https://ai.3zapi.top/v1/images/generations", headers={"Authorization": "Bearer YOUR_API_KEY"}, json={"model": "gemini-3-pro-image-preview-official", "prompt": "未来城市夜景", "size": "1024x1024", "resolution": "4k", "n": 1})
print(response.json()["data"][0]["url"])
```

## JavaScript

```javascript
const body = {model: "gemini-3-pro-image-preview-official", prompt: "未来城市夜景", size: "1024x1024", resolution: "4k", n: 1};
const response = await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers: {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"}, body: JSON.stringify(body)});
console.log((await response.json()).data[0].url);
```

## 响应与排错

网关返回最终 `{"created": 0, "data": [{"url": "..."}]}`，不应轮询上游任务。遇到 `401` 请检查 Bearer 密钥；遇到请求错误请确认模型 ID、分辨率和不超过 14 张的参考图。
$md$, NOW()),
('gemini-3.1-flash-image-preview', 'Gemini 3.1 Flash Image Preview 图像生成', '使用 Gemini 3.1 Flash Image Preview 生成图像。', '图像模型', 2244, 'published', $md$
# Gemini 3.1 Flash Image Preview 图像生成

模型 ID：`gemini-3.1-flash-image-preview`。使用 `POST https://ai.3zapi.top/v1/images/generations`，认证为 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `gemini-3.1-flash-image-preview`；`prompt` 为描述，兼容流程中 `n` 固定为 `1`。
- `size` 使用像素尺寸或比例；`resolution` 支持本地 `1k`、`2k`、`4k`，不要使用 `0.5k`。
- `image_urls` 最多 14 张，且可选 `google_search`、`google_image_search`。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"gemini-3.1-flash-image-preview","prompt":"复古海报","size":"1024x1024","resolution":"1k","n":1,"google_search":true}'
```

## Python

```python
import requests
payload = {"model": "gemini-3.1-flash-image-preview", "prompt": "复古海报", "size": "1024x1024", "resolution": "1k", "n": 1, "google_image_search": True}
response = requests.post("https://ai.3zapi.top/v1/images/generations", headers={"Authorization": "Bearer YOUR_API_KEY"}, json=payload)
print(response.json()["data"][0]["url"])
```

## JavaScript

```javascript
const response = await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers: {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"}, body: JSON.stringify({model: "gemini-3.1-flash-image-preview", prompt: "复古海报", size: "1024x1024", resolution: "1k", n: 1, google_search: true})});
console.log((await response.json()).data[0].url);
```

## 响应与排错

网关同步等待并返回 `{"created": 0, "data": [{"url": "..."}]}`；提交后不必轮询上游任务。`401` 时检查 Bearer 密钥；`0.5k`、超过 14 张参考图或无效搜索布尔值会导致请求错误。
$md$, NOW()),
('gemini-3.1-flash-image-preview-official', 'Gemini 3.1 Flash Image Preview 官方图像生成', '使用 Gemini 3.1 Flash Image Preview 官方路径生成图像。', '图像模型', 2245, 'published', $md$
# Gemini 3.1 Flash Image Preview 官方图像生成

模型 ID：`gemini-3.1-flash-image-preview-official`。请求 `POST https://ai.3zapi.top/v1/images/generations`，认证头是 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `gemini-3.1-flash-image-preview-official` 和 `prompt`；兼容流程中 `n` 固定为 `1`。
- `size` 使用像素尺寸或比例；`resolution` 支持本地 `1k`、`2k`、`4k`，不使用 `0.5k`。
- `image_urls` 最多 14 张；可选 `google_search` 和 `google_image_search`。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"gemini-3.1-flash-image-preview-official","prompt":"玻璃质感图标","size":"1024x1024","resolution":"2k","n":1,"google_image_search":true}'
```

## Python

```python
import requests
response = requests.post("https://ai.3zapi.top/v1/images/generations", headers={"Authorization": "Bearer YOUR_API_KEY"}, json={"model": "gemini-3.1-flash-image-preview-official", "prompt": "玻璃质感图标", "size": "1024x1024", "resolution": "2k", "n": 1, "google_search": True})
print(response.json()["data"][0]["url"])
```

## JavaScript

```javascript
const body = {model: "gemini-3.1-flash-image-preview-official", prompt: "玻璃质感图标", size: "1024x1024", resolution: "2k", n: 1, google_image_search: true};
const response = await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers: {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"}, body: JSON.stringify(body)});
console.log((await response.json()).data[0].url);
```

## 响应与排错

网关返回最终 `{"created": 0, "data": [{"url": "..."}]}`，无需轮询提交请求。Bearer 鉴权失败请更换有效密钥；请只传 `1k`、`2k`、`4k`，并将参考图限制在 14 张内。
$md$, NOW()),
('midjourney', 'Midjourney 图像生成', '使用 Midjourney 图像生成接口。', '图像模型', 2246, 'published', $md$
# Midjourney 图像生成

模型 ID：`midjourney`。请求 `POST https://ai.3zapi.top/v1/midjourney/generations`，认证头为 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `midjourney`；`prompt` 为描述。
- 本地解析字段只有 `size`、`version`、`speed`、`quality`、`stylize`、`chaos`、`weird`、`stop`、`niji`、`raw`、`tile`、`image_urls`。

## cURL

```bash
curl https://ai.3zapi.top/v1/midjourney/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"midjourney","prompt":"雨后的霓虹街道","size":"1024x1024","version":"6","quality":"1","stylize":100,"chaos":10,"raw":true}'
```

## Python

```python
import requests
payload = {"model": "midjourney", "prompt": "雨后的霓虹街道", "size": "1024x1024", "version": "6", "speed": "fast", "stylize": 100, "image_urls": []}
response = requests.post("https://ai.3zapi.top/v1/midjourney/generations", headers={"Authorization": "Bearer YOUR_API_KEY"}, json=payload)
print(response.json()["data"][0]["url"])
```

## JavaScript

```javascript
const body = {model: "midjourney", prompt: "雨后的霓虹街道", size: "1024x1024", version: "6", quality: "1", niji: false, tile: false};
const response = await fetch("https://ai.3zapi.top/v1/midjourney/generations", {method: "POST", headers: {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"}, body: JSON.stringify(body)});
console.log((await response.json()).data[0].url);
```

## 响应与排错

网关等待完成后直接返回 `{"created": 0, "data": [{"url": "..."}]}`，不要对提交调用轮询上游任务。`401` 时检查 Bearer 密钥；参数错误时仅使用上列本地解析字段，并检查 `stop` 是否与 `version` 兼容。
$md$, NOW()),
('doubao-seedance-4-0', '豆包 Seedance 4.0 图像生成', '使用豆包 Seedance 4.0 异步图像生成。', '图像模型', 2247, 'published', $md$
# 豆包 Seedance 4.0 图像生成

模型 ID：`doubao-seedance-4-0`。提交到 `POST https://ai.3zapi.top/v1/images/generations`，并使用 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `doubao-seedance-4-0`；`prompt`、像素尺寸或比例形式的 `size`、`n` 控制生成；tier 填入 `resolution`。
- `image_urls` 可传参考图，但参考图数量加 `n` 不得超过 15。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"doubao-seedance-4-0","prompt":"晨雾森林","size":"1024x1024","resolution":"2k","n":1}'
```

## Python

```python
import requests
import time
headers = {"Authorization": "Bearer YOUR_API_KEY"}
created = requests.post("https://ai.3zapi.top/v1/images/generations", headers=headers, json={"model": "doubao-seedance-4-0", "prompt": "晨雾森林", "size": "1024x1024", "resolution": "2k", "n": 1}).json()
task_id = created["data"][0]["task_id"]
while True:
    task = requests.get(f"https://ai.3zapi.top/v1/tasks/{task_id}", headers=headers).json()
    status = task.get("data", {}).get("status", task.get("status", "")).lower()
    if status in {"completed", "succeeded", "success"}:
        print(task["data"]["result"]["images"][0]["url"][0])
        break
    if status in {"failed", "cancelled", "canceled"}:
        raise RuntimeError(f"Seedream task failed: {task}")
    time.sleep(3)
```

## JavaScript

```javascript
const headers = {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"};
const created = await (await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers, body: JSON.stringify({model: "doubao-seedance-4-0", prompt: "晨雾森林", size: "1024x1024", resolution: "2k", n: 1})})).json();
const taskId = created.data[0].task_id;
for (;;) {
  const task = await (await fetch(`https://ai.3zapi.top/v1/tasks/${taskId}`, {headers})).json();
  const status = (task.data?.status ?? task.status ?? "").toLowerCase();
  if (["completed", "succeeded", "success"].includes(status)) { console.log(task.data.result.images[0].url[0]); break; }
  if (["failed", "cancelled", "canceled"].includes(status)) throw new Error(`Seedream task failed: ${JSON.stringify(task)}`);
  await new Promise(resolve => setTimeout(resolve, 3000));
}
```

## 响应与排错

提交后会从 `data[0].task_id` 得到任务标识；请轮询 `GET /v1/tasks/{task_id}` 直到完成，并从 `data.result.images[0].url[0]` 读取图像 URL。结果 URL 应尽快下载。`401` 请检查 Bearer 密钥；任务失败时检查 `model`、提示词和“参考图 + n <= 15”。
$md$, NOW()),
('doubao-seedance-4-5', '豆包 Seedance 4.5 图像生成', '使用豆包 Seedance 4.5 异步图像生成。', '图像模型', 2248, 'published', $md$
# 豆包 Seedance 4.5 图像生成

模型 ID：`doubao-seedance-4-5`。提交请求为 `POST https://ai.3zapi.top/v1/images/generations`，认证为 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: `doubao-seedance-4-5`；传入 `prompt`、像素尺寸或比例形式的 `size`、`n`。
- 推荐 `resolution` 仅使用 `2k` 或 `4k`。`image_urls` 参考图数量加 `n` 不得超过 15。

## cURL

```bash
curl https://ai.3zapi.top/v1/images/generations -H "Authorization: Bearer YOUR_API_KEY" -H "Content-Type: application/json" -d '{"model":"doubao-seedance-4-5","prompt":"极简建筑渲染","size":"1024x1024","resolution":"4k","n":1}'
```

## Python

```python
import requests
import time
headers = {"Authorization": "Bearer YOUR_API_KEY"}
created = requests.post("https://ai.3zapi.top/v1/images/generations", headers=headers, json={"model": "doubao-seedance-4-5", "prompt": "极简建筑渲染", "size": "1024x1024", "resolution": "4k", "n": 1}).json()
task_id = created["data"][0]["task_id"]
while True:
    task = requests.get(f"https://ai.3zapi.top/v1/tasks/{task_id}", headers=headers).json()
    status = task.get("data", {}).get("status", task.get("status", "")).lower()
    if status in {"completed", "succeeded", "success"}:
        print(task["data"]["result"]["images"][0]["url"][0])
        break
    if status in {"failed", "cancelled", "canceled"}:
        raise RuntimeError(f"Seedream task failed: {task}")
    time.sleep(3)
```

## JavaScript

```javascript
const headers = {Authorization: "Bearer YOUR_API_KEY", "Content-Type": "application/json"};
const created = await (await fetch("https://ai.3zapi.top/v1/images/generations", {method: "POST", headers, body: JSON.stringify({model: "doubao-seedance-4-5", prompt: "极简建筑渲染", size: "1024x1024", resolution: "4k", n: 1})})).json();
const taskId = created.data[0].task_id;
for (;;) {
  const task = await (await fetch(`https://ai.3zapi.top/v1/tasks/${taskId}`, {headers})).json();
  const status = (task.data?.status ?? task.status ?? "").toLowerCase();
  if (["completed", "succeeded", "success"].includes(status)) { console.log(task.data.result.images[0].url[0]); break; }
  if (["failed", "cancelled", "canceled"].includes(status)) throw new Error(`Seedream task failed: ${JSON.stringify(task)}`);
  await new Promise(resolve => setTimeout(resolve, 3000));
}
```

## 响应与排错

提交后会从 `data[0].task_id` 得到任务标识，轮询 `GET /v1/tasks/{task_id}` 至完成，再从 `data.result.images[0].url[0]` 读取图像 URL；结果 URL 应尽快下载。`401` 时检查 Bearer 密钥，超过“参考图 + n <= 15”或未按建议使用 `resolution` 的 `2k`、`4k` 时请调整参数。
$md$, NOW())
ON CONFLICT (slug) DO NOTHING;
