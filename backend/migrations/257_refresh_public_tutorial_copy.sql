-- Refresh only known Markdown fragments from the historical seed and the
-- frozen public tutorial snapshots. Independently authored sections stay intact.
DO $refresh$
DECLARE
    target tutorial_pages%ROWTYPE;
    patch RECORD;
    revised_content TEXT;
    revised_description TEXT;
    match_at INTEGER;
BEGIN
    FOR target IN
        SELECT * FROM tutorial_pages
        WHERE status = 'published'
          AND slug IN ('getting-started', 'codex', 'claude-code', 'cc-switch',
                       'cockpit-tools', 'openclaw', 'hermes-agent', 'faq',
                       'seedream-5-0-pro', 'seedream-5-0-lite')
        ORDER BY slug
    LOOP
        revised_content := target.content_md;
        revised_description := target.description;

        FOR patch IN
            SELECT old_fragment, new_fragment
            FROM (VALUES
                -- cc-switch: live / seed
                ('cc-switch', $before$CC Switch 适合需要多工具、多账号切换的用户。推荐从落叶网络控制台一键导入 Provider，也可以手动添加。$before$, $after$CC Switch 适合需要多工具、多账号切换的用户。可从落叶网络控制台一键导入 Provider，也可以手动添加。

## 从 API 密钥页导入

1. 打开 [API 密钥页面](/keys)，找到已启用且有目标模型权限的密钥，在该行点击“导入到 CCS”。如果按钮未显示，可在 CC Switch 手动添加对应工具的 Provider。
2. 在 CC Switch 中确认导入，并检查目标工具、Base URL 和 Provider；选中落叶网络 Provider 后再使用。

导入链接和导出的配置可能含有完整 API 密钥。只在可信设备上打开，不要转发或上传。$after$),
                -- claude-code: live / seed
                ('claude-code', $before$## 安装

npm 安装路径需要 Node.js 18+。Windows 用户如果不走 WSL，也要准备 Git for Windows / Git Bash；如果使用官方原生安装器，可以先按官方安装器流程执行，再用 `claude doctor` 检查环境。

[[command title="安装 Claude Code"]]
npm install -g @anthropic-ai/claude-code --registry=https://registry.npmmirror.com
claude --version
[[/command]]$before$, $after$## 安装

优先按 Anthropic 官方文档使用原生安装器；Windows 可直接用 PowerShell。Git for Windows 是可选的：未安装时，原生 Windows 版可使用 PowerShell；需要 Git Bash 功能时再安装。

[[link-button href="https://code.claude.com/docs/en/setup" label="查看 Claude Code 官方安装说明"]]

[[command title="Windows PowerShell 原生安装" lang="powershell"]]
irm https://claude.ai/install.ps1 | iex
[[/command]]

[[command title="macOS / Linux / WSL 原生安装" lang="bash"]]
curl -fsSL https://claude.ai/install.sh | bash
[[/command]]

如需改用 npm，先安装 Node.js 22 或更新版本。原生安装不需要 Node.js。

[[command title="npm 备选安装"]]
npm install -g @anthropic-ai/claude-code@latest
[[/command]]

安装后打开新终端，运行 `claude --version`；遇到环境问题可运行 `claude doctor`。$after$),
                -- cockpit-tools: live / seed
                ('cockpit-tools', $before$登录后可以一键添加到 Cockpit Tools，浏览器会唤起 Cockpit Tools。若不能唤起，检查协议注册和浏览器权限。

导入账号后，在 Cockpit Tools 左侧点击 Codex 图标，确认当前账号已经切到落叶网络 Provider。$before$, $after$1. 打开 [API 密钥页面](/keys)，找到已启用且有目标模型权限的 OpenAI 密钥，在该行点击“导入 Cockpit”，允许浏览器唤起 Cockpit Tools。
2. 如果唤起失败，在提示对话框中点击“下载导入 JSON”，再到 Cockpit Tools 的 Codex 页面手动导入该文件。
3. 在 Cockpit Tools 左侧点击 Codex 图标，检查导入账号的 Base URL，确认并选中落叶网络 Provider。

导入链接和 JSON 文件含有 API 密钥，不要分享或上传；导入后妥善保管或删除文件。$after$),
                -- codex: live
                ('codex', $before$## 第一步-安装 Codex App

[[command title="Windows winget 安装"]]
winget install --id OpenAI.Codex -e
[[/command]]$before$, $after$## 第一步：安装官方 ChatGPT

打开官方 ChatGPT 下载页，安装桌面应用。本站 Windows 配置助手是独立工具，用来配置应用中的 Codex 功能，不是官方应用本身。

[[link-button href="https://chatgpt.com/download/" label="下载官方 ChatGPT"]]

[[link-button href="/tutorial" label="打开 Windows 配置助手教程"]]

[[command title="Windows winget 安装"]]
winget install --id OpenAI.Codex -e
[[/command]]$after$),
                -- codex: live
                ('codex', $before$## 第二步-创建 codex 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型请选择 codex。创建后复制 Token，后面写入 auth.json。

[[screenshot src="/tutorial/api-key/sidebar-api-key.png" alt="API 密钥入口" caption="左侧进入 API 密钥页面"]]
[[screenshot src="/tutorial/api-key/create-key-button.png" alt="创建密钥按钮" caption="点击右上角创建密钥"]]
[[screenshot src="/tutorial/api-key/create-key-dialog.png" alt="创建密钥弹窗" caption="密钥类型选择 codex"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]$before$, $after$## 获取 API Key

登录落叶网络控制台，打开 API 密钥页面。优先复用已启用且有目标模型权限的 API Key；如需新建，点“创建密钥”，厂商选 OpenAI，分组选择支持目标模型且有可用额度的分组。复制完整密钥，勿分享给他人。

[[link-button href="/keys" label="打开 API 密钥页面"]]
$after$),
                -- codex: live
                ('codex', $before$把下面内容完整粘进去，然后保存。

[[command title="config.toml" lang="toml"]]
model_provider = "OpenAI"
model = "gpt-5.6-sol"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1

[[/command]]

- env_key = "OPENAI_API_KEY" 要和 auth.json 里的字段名保持一致；CLI 环境变量方案也使用这个名字。$before$, $after$先备份已有的 config.toml 和 auth.json。若 config.toml 已存在，只合并下方相关字段和 Provider 表，不要覆盖其他设置或重复定义同名表。

```toml
model_provider = "OpenAI"
model = "gpt-6.1-sol"
review_model = "gpt-6.1-sol"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1
```

- `model_provider = "OpenAI"` 与 `[model_providers.OpenAI]` 对应；`wire_api = "responses"` 选择 Responses 协议，使用 `https://ai.3zapi.com` 根地址，不加 `/v1`。
- `model` 和 `review_model` 中的 `gpt-6.1-sol` 与快速接入生成器的示例一致，不保证所有分组可用；请改为当前密钥有权访问的模型。
- `requires_openai_auth = true` 与同目录 auth.json 的 `OPENAI_API_KEY` 配套使用。$after$),
                -- codex: live
                ('codex', $before$auth.json 用来保存你的 API Key。这里的 key 就是你在控制台创建的 codex 类型密钥。$before$, $after$auth.json 保存你的 API Key。将 API 密钥页复制的有效密钥填入 `OPENAI_API_KEY`，与上面的 config.toml 配套使用。$after$),
                -- codex: live / seed
                ('codex', $before$[[command title="auth.json" lang="json"]]
{
  "OPENAI_API_KEY": "替换成你的 codex 类型 API Key"
}
[[/command]]

如果 auth.json 已经存在，只需要更新 OPENAI_API_KEY 这一项，不要删除其他仍在使用的配置。$before$, $after$[[command title="auth.json" lang="json"]]
{
  "OPENAI_API_KEY": "替换成你的 API Key"
}
[[/command]]

如果 auth.json 已经存在，请先备份，只更新 `OPENAI_API_KEY`，保留其他配置和有效的 JSON 格式。$after$),
                -- codex: live
                ('codex', $before$[[callout type="warning" title="失败时先查这三项"]]
1. Base URL 是否为 `https://ai.3zapi.com`。
2. API Key 是否复制完整，并且密钥类型是 codex。
3. 当前分组是否有可用套餐和余额。
4.是否是完全退出codex、其他账号切换工具下修改配置。
[[/callout]]$before$, $after$[[callout type="warning" title="连接失败时先检查"]]
1. Codex 的 Base URL 是否为 `https://ai.3zapi.com`，且使用 Responses。
2. auth.json 中的 API Key 是否完整、有效。
3. 当前密钥的分组是否支持所选模型，账户是否有可用余额或额度。
4. 修改后是否完全退出再重启，配置是否被其他账号切换工具覆盖。
[[/callout]]$after$),
                -- codex: live
                ('codex', $before$## 常见问题

- 401 或 unauthorized：优先检查 Token 是否复制完整、密钥类型是否为 codex。
- model not found：当前分组可能不支持配置里的模型，换成控制台可用模型后再试。
- 配置不生效：修改配置后重启 Codex App；如果是 CLI，关闭旧终端后重新打开。$before$, $after$## 常见问题

- 401 或 unauthorized：检查 API Key 是否完整、有效且已启用。
- model not found：核对模型 ID 和分组权限；没有权限时，换用控制台中当前分组可用的模型。
- 配置不生效：完全退出并重新打开 Codex App；CLI 用户可打开新终端重试。

## Base URL 与协议

- Codex（Responses）：使用根地址 `https://ai.3zapi.com`，不加 `/v1`。
- 通用 OpenAI 兼容客户端：如果客户端要求 `/v1`，填写 `https://ai.3zapi.com/v1`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 `https://ai.3zapi.com`，不加 `/v1`，不要选 OpenAI 协议。$after$),
                -- faq: live / seed
                ('faq', $before$## 安装失败

先看是哪类工具安装失败。Codex App 通常不需要 Node.js；Codex CLI 需要 Node.js/npm，建议当前 LTS；Claude Code 的 npm 安装路径需要 Node.js 18+。npm 下载慢时再换镜像源。$before$, $after$## 安装失败

先确认是哪类工具：Codex App 不需要先安装 Node.js；通过 npm 安装 Codex CLI 建议用当前 LTS。Claude Code 优先走 [Anthropic 官方原生安装](https://code.claude.com/docs/en/setup)，Windows 原生安装时 Git for Windows 可选；改用 npm 时要求 Node.js 22+。$after$),
                -- faq: live / seed
                ('faq', $before$## 命令找不到

如果是 Codex App，先确认是否已经安装桌面端并登录；如果是 Codex CLI 或 Claude Code，再确认 Node.js/npm 是否在 PATH 里。代码项目里找不到 git 时，再检查 Git for Windows / Git Bash 安装路径。$before$, $after$## 命令找不到

Codex CLI 或 Claude Code 安装后，先新开终端，再用 `codex --version` 或 `claude --version` 检查命令是否在 PATH 中。只有通过 npm 安装时才需要同时检查 Node.js 和 npm；缺少 git 与应用本身没有装好是两回事。$after$),
                -- faq: live / seed
                ('faq', $before$## 鉴权失败

检查 Token 和 Base URL，注意不要多复制空格。$before$, $after$## 鉴权失败

401 时检查 [API 密钥](/keys) 是否完整、有效且已启用。403 或模型不可用时，检查密钥的厂商、分组是否匹配客户端，确认分组有权访问该模型，模型 ID 拼写正确；同时检查账户余额和套餐额度。不要将真实密钥发给他人。

## Base URL 与协议

- Codex（Responses）：使用根地址 `https://ai.3zapi.com`，不加 `/v1`。
- 通用 OpenAI 兼容客户端：如果客户端要求 `/v1`，填写 `https://ai.3zapi.com/v1`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 `https://ai.3zapi.com`，不加 `/v1`，不要选 OpenAI 协议。$after$),
                -- getting-started: live
                ('getting-started', $before$## 安装 chatgpt App

打开 chatgpt App 安装页，按系统提示安装桌面端。

[[link-button href="https://chatgpt.com/codex" label="打开 chatgpt App"]]$before$, $after$## 安装官方 ChatGPT

打开官方 ChatGPT 下载页，安装桌面应用。本站 Windows 配置助手是独立工具，用来配置应用中的 Codex 功能，不是官方应用本身。

[[link-button href="https://chatgpt.com/download/" label="下载官方 ChatGPT"]]

[[link-button href="/tutorial" label="打开 Windows 配置助手教程"]]$after$),
                -- getting-started: live / seed
                ('getting-started', $before$## 创建 API 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型请选择 `codex`。

[[screenshot src="/tutorial/api-key/sidebar-api-key.png" alt="API 密钥入口" caption="左侧进入 API 密钥页面"]]
[[screenshot src="/tutorial/api-key/create-key-button.png" alt="创建密钥按钮" caption="点击右上角创建密钥"]]
[[screenshot src="/tutorial/api-key/create-key-dialog.png" alt="创建密钥弹窗" caption="选择分组和密钥类型后创建"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]$before$, $after$## 获取 API Key

登录落叶网络控制台，打开 API 密钥页面。优先复用已启用且有目标模型权限的 API Key；如需新建，点“创建密钥”，厂商选 OpenAI，分组选择支持目标模型且有可用额度的分组。复制完整密钥，勿分享给他人。

[[link-button href="/keys" label="打开 API 密钥页面"]]$after$),
                -- getting-started: live
                ('getting-started', $before$## Base URL

OpenAI 兼容工具通常使用：

[[command title="OpenAI 兼容 Base URL"]]
https://ai.3zapi.com/v1
[[/command]]

Claude Code 兼容工具通常使用：

[[command title="Anthropic 兼容 Base URL"]]
https://ai.3zapi.com
[[/command]]$before$, $after$## Base URL 与协议

- Codex（Responses）：使用根地址 `https://ai.3zapi.com`，不加 `/v1`。
- 通用 OpenAI 兼容客户端：如果客户端要求 `/v1`，填写 `https://ai.3zapi.com/v1`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 `https://ai.3zapi.com`，不加 `/v1`，不要选 OpenAI 协议。$after$),
                -- getting-started: live
                ('getting-started', $before$## 下一步：chatgpt 配置

打开 chatgpt 教程，复制配置文件示例并写入你的 Token。

[[link-button href="/tutorial/codex" label="查看 Codex 配置"]]$before$, $after$## 下一步：配置 Codex

Windows 用户可按配置助手教程操作。手动配置时，请先备份已有文件，再按 Codex 教程合并 config.toml 和 auth.json。

[[link-button href="/tutorial/codex" label="查看 Codex 配置"]]$after$),
                -- hermes-agent: live
                ('hermes-agent', $before$参考：https://docs.easyrouter.io/zh/docs/apps/hermes-agent

## 安装

[[command title="安装 Hermes-Agent"]]
curl -fsSL https://raw.githubusercontent.com/terryso/hermes-agent/main/install.sh | bash
[[/command]]

## 配置模型

[[command title="hermes model"]]
hermes model
# Provider: Custom Endpoint
# Base URL: https://ai.3zapi.com/v1
# API Key: 你的API Token
# Model: gpt-5.3-codex
[[/command]]$before$, $after$## 安装

按 Nous Research 官方文档选择适合系统的安装方式。

[[link-button href="https://hermes-agent.nousresearch.com/docs/getting-started/installation" label="查看 Hermes Agent 官方安装说明"]]

## 配置模型

运行 `hermes model`，选择 Custom Endpoint（自定义端点）。按 OpenAI 兼容模式填写 Base URL `https://ai.3zapi.com/v1`，输入你自己的 API Key，并从控制台选择当前密钥和分组有权访问的模型 ID。不要把真实密钥写入共享配置。

[[link-button href="https://hermes-agent.nousresearch.com/docs/integrations/providers#general-setup" label="查看 Hermes Provider 配置文档"]]$after$),
                -- openclaw: live
                ('openclaw', $before$## openclaw.json

[[command title="openclaw.json" lang="json"]]
{
  "models": {
    "providers": {
      "luoye": {
        "api": "openai-completions",
        "name": "luoye",
        "baseURL": "https://ai.3zapi.com/v1",
        "envKey": "LUOYE_API_KEY",
        "models": [
          {
            "id": "gpt-5.3-codex",
            "contextWindow": 400000
          }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": {
        "primary": "luoye/gpt-5.3-codex"
      }
    }
  }
}
[[/command]]$before$, $after$## openclaw.json

先在启动 OpenClaw 的环境中设置 `LUOYE_API_KEY`，不要在配置文件中写入真实密钥。以下以 OpenAI Chat Completions 为例，将各处 `YOUR_MODEL_ID` 替换为当前密钥和分组可用的模型 ID。若选 Responses 模式，请在客户端文档中确认对应设置，并使用与所选协议匹配的服务端端点；不要混用两种模式。

[[link-button href="https://docs.openclaw.ai/concepts/model-providers/custom-providers" label="查看 OpenClaw 自定义 Provider 文档"]]

[[command title="openclaw.json" lang="json"]]
{
  "models": {
    "providers": {
      "luoye": {
        "baseUrl": "https://ai.3zapi.com/v1",
        "apiKey": "${LUOYE_API_KEY}",
        "api": "openai-completions",
        "models": [{ "id": "YOUR_MODEL_ID", "name": "YOUR_MODEL_ID" }]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": { "primary": "luoye/YOUR_MODEL_ID" }
    }
  }
}
[[/command]]$after$),
                -- openclaw: live / seed
                ('openclaw', $before$参考：https://docs.easyrouter.io/zh/docs/apps/openclaw

## 初始化$before$, $after$## 初始化$after$),
                -- seedream-5-0-lite: live / seed
                ('seedream-5-0-lite', $before$```bash
curl https://ai.3zapi.cc/v1/images/generations \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "seedream-5-0-lite",
    "prompt": "一只在花园里玩耍的金毛犬，阳光明媚，高清摄影风格",
    "size": "16:9",
    "resolution": "2K",
    "n": 1,
    "output_format": "png"
  }'
```$before$, $after$```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "seedream-5-0-lite",
    "prompt": "一只在花园里玩耍的金毛犬，阳光明媚，高清摄影风格",
    "size": "16:9",
    "resolution": "2K",
    "n": 1,
    "output_format": "png"
  }'
```$after$),
                -- seedream-5-0-pro: live / seed
                ('seedream-5-0-pro', $before$```bash
curl https://ai.3zapi.cc/v1/images/generations \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "seedream-5-0-pro",
    "prompt": "赛博朋克风格的城市夜景，霓虹灯反射在湿润街道上",
    "size": "16:9",
    "resolution": "2K",
    "n": 1
  }'
```$before$, $after$```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "seedream-5-0-pro",
    "prompt": "赛博朋克风格的城市夜景，霓虹灯反射在湿润街道上",
    "size": "16:9",
    "resolution": "2K",
    "n": 1
  }'
```$after$),
                -- codex: seed
                ('codex', $before$Codex 推荐使用落叶网络的 OpenAI 兼容接口。流程是：安装 Codex App，创建 codex 类型密钥，写入 config.toml 和 auth.json，然后启动验证。$before$, $after$Codex 使用落叶网络的 OpenAI Responses 接口。先安装官方应用，准备有效的 API Key，再配置 config.toml 和 auth.json。$after$),
                -- codex: seed
                ('codex', $before$也可以从 Microsoft Store 或 https://chatgpt.com/codex 安装桌面端。$before$, $after$打开官方 ChatGPT 下载页，安装桌面应用。本站 Windows 配置助手是独立工具，用来配置应用中的 Codex 功能，不是官方应用本身。

[[link-button href="https://chatgpt.com/download/" label="下载官方 ChatGPT"]]

[[link-button href="/tutorial" label="打开 Windows 配置助手教程"]]$after$),
                -- codex: seed
                ('codex', $before$## 创建 codex 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型请选择 codex。创建后复制 Token，后面写入 auth.json。

[[screenshot src="/tutorial/api-key/sidebar-api-key.png" alt="API 密钥入口" caption="左侧进入 API 密钥页面"]]
[[screenshot src="/tutorial/api-key/create-key-button.png" alt="创建密钥按钮" caption="点击右上角创建密钥"]]
[[screenshot src="/tutorial/api-key/create-key-dialog.png" alt="创建密钥弹窗" caption="密钥类型选择 codex"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]$before$, $after$## 获取 API Key

登录落叶网络控制台，打开 API 密钥页面。优先复用已启用且有目标模型权限的 API Key；如需新建，点“创建密钥”，厂商选 OpenAI，分组选择支持目标模型且有可用额度的分组。复制完整密钥，勿分享给他人。

[[link-button href="/keys" label="打开 API 密钥页面"]]$after$),
                -- codex: seed
                ('codex', $before$把下面内容完整粘进去，然后保存。

[[command title="config.toml" lang="toml"]]
model_provider = "OpenAI"
model = "gpt-6-astra"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1
[[/command]]

- model_provider = "OpenAI" 与 [model_providers.OpenAI] 对应，服务商显示名称为 3Z API。
- Codex 的 base_url 使用根地址 `https://ai.3zapi.com`，无需追加 `/v1`。
- requires_openai_auth = true 配合下方 auth.json 中的 OPENAI_API_KEY 使用。
- wire_api = "responses" 表示使用 Responses 协议。$before$, $after$先备份已有的 config.toml 和 auth.json。若 config.toml 已存在，只合并下方相关字段和 Provider 表，不要覆盖其他设置或重复定义同名表。

```toml
model_provider = "OpenAI"
model = "gpt-6.1-sol"
review_model = "gpt-6.1-sol"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1
```

- `model_provider = "OpenAI"` 与 `[model_providers.OpenAI]` 对应；`wire_api = "responses"` 选择 Responses 协议，使用 `https://ai.3zapi.com` 根地址，不加 `/v1`。
- `model` 和 `review_model` 中的 `gpt-6.1-sol` 与快速接入生成器的示例一致，不保证所有分组可用；请改为当前密钥有权访问的模型。
- `requires_openai_auth = true` 与同目录 auth.json 的 `OPENAI_API_KEY` 配套使用。$after$),
                -- codex: seed
                ('codex', $before$auth.json 用来保存你的 API Key。这里的 key 就是你在落叶网络控制台创建的 codex 类型密钥。$before$, $after$auth.json 保存你的 API Key。将 API 密钥页复制的有效密钥填入 `OPENAI_API_KEY`，与上面的 config.toml 配套使用。$after$),
                -- codex: seed
                ('codex', $before$[[callout type="warning" title="失败时先查这三项"]]
1. Base URL 是否为 `https://ai.3zapi.com`。
2. API Key 是否复制完整，并且密钥类型是 codex。
3. 当前分组是否有可用套餐和余额。
[[/callout]]$before$, $after$[[callout type="warning" title="连接失败时先检查"]]
1. Codex 的 Base URL 是否为 `https://ai.3zapi.com`，且使用 Responses。
2. auth.json 中的 API Key 是否完整、有效。
3. 当前密钥的分组是否支持所选模型，账户是否有可用余额或额度。
4. 修改后是否完全退出再重启，配置是否被其他账号切换工具覆盖。
[[/callout]]$after$),
                -- codex: seed
                ('codex', $before$## 常见问题

- 401 或 unauthorized：优先检查 Token 是否复制完整、密钥类型是否为 codex。
- model not found：当前分组可能不支持配置里的模型，换成控制台可用模型后再试。
- Base URL 报错：本页 Codex 配置使用 `https://ai.3zapi.com` 根地址。
- 配置不生效：修改配置后重启 Codex App；如果是 CLI，关闭旧终端后重新打开。$before$, $after$## 常见问题

- 401 或 unauthorized：检查 API Key 是否完整、有效且已启用。
- model not found：核对模型 ID 和分组权限；没有权限时，换用控制台中当前分组可用的模型。
- 配置不生效：完全退出并重新打开 Codex App；CLI 用户可打开新终端重试。

## Base URL 与协议

- Codex（Responses）：使用根地址 `https://ai.3zapi.com`，不加 `/v1`。
- 通用 OpenAI 兼容客户端：如果客户端要求 `/v1`，填写 `https://ai.3zapi.com/v1`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 `https://ai.3zapi.com`，不加 `/v1`，不要选 OpenAI 协议。$after$),
                -- getting-started: seed
                ('getting-started', $before$[[callout type="tip" title="新手最快路线"]]
1. 安装或打开 Codex App。
2. 登录落叶网络控制台，创建 codex 类型密钥并复制 Token。
3. 按 Codex 教程写入 config.toml 和 auth.json。
4. 启动 Codex App，确认模型能正常响应。
[[/callout]]$before$, $after$[[callout type="tip" title="新手最快路线"]]
1. 安装官方 ChatGPT 桌面应用。
2. 在 API 密钥页复制可用的 API Key；没有合适的密钥再创建。
3. 用 Windows 配置助手，或按 Codex 教程手动合并 config.toml 和 auth.json。
4. 重启应用，确认所选模型能正常响应。
[[/callout]]$after$),
                -- getting-started: seed
                ('getting-started', $before$## 安装 Codex App

打开 Codex App 安装页，按系统提示安装桌面端。

[[link-button href="https://chatgpt.com/codex" label="打开 Codex App"]]$before$, $after$## 安装官方 ChatGPT

打开官方 ChatGPT 下载页，安装桌面应用。本站 Windows 配置助手是独立工具，用来配置应用中的 Codex 功能，不是官方应用本身。

[[link-button href="https://chatgpt.com/download/" label="下载官方 ChatGPT"]]

[[link-button href="/tutorial" label="打开 Windows 配置助手教程"]]$after$),
                -- getting-started: seed
                ('getting-started', $before$## Base URL

Codex 使用根地址 https://ai.3zapi.com；其他要求 /v1 的 OpenAI 兼容工具使用：

[[command title="OpenAI 兼容 Base URL"]]
https://ai.3zapi.com/v1
[[/command]]

Claude Code 兼容工具通常使用：

[[command title="Anthropic 兼容 Base URL"]]
https://ai.3zapi.top
[[/command]]$before$, $after$## Base URL 与协议

- Codex（Responses）：使用根地址 `https://ai.3zapi.com`，不加 `/v1`。
- 通用 OpenAI 兼容客户端：如果客户端要求 `/v1`，填写 `https://ai.3zapi.com/v1`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 `https://ai.3zapi.com`，不加 `/v1`，不要选 OpenAI 协议。$after$),
                -- getting-started: seed
                ('getting-started', $before$## 下一步：Codex 配置

打开 Codex 教程，复制配置文件示例并写入你的 Token。

[[link-button href="/tutorial/codex" label="查看 Codex 配置"]]$before$, $after$## 下一步：配置 Codex

Windows 用户可按配置助手教程操作。手动配置时，请先备份已有文件，再按 Codex 教程合并 config.toml 和 auth.json。

[[link-button href="/tutorial/codex" label="查看 Codex 配置"]]$after$),
                -- hermes-agent: seed
                ('hermes-agent', $before$参考：https://docs.easyrouter.io/zh/docs/apps/hermes-agent

## 安装

[[command title="安装 Hermes-Agent"]]
curl -fsSL https://raw.githubusercontent.com/terryso/hermes-agent/main/install.sh | bash
[[/command]]

## 配置模型

[[command title="hermes model"]]
hermes model
# Provider: Custom Endpoint
# Base URL: https://ai.3zapi.top/v1
# API Key: 你的API Token
# Model: gpt-5.3-codex
[[/command]]$before$, $after$## 安装

按 Nous Research 官方文档选择适合系统的安装方式。

[[link-button href="https://hermes-agent.nousresearch.com/docs/getting-started/installation" label="查看 Hermes Agent 官方安装说明"]]

## 配置模型

运行 `hermes model`，选择 Custom Endpoint（自定义端点）。按 OpenAI 兼容模式填写 Base URL `https://ai.3zapi.com/v1`，输入你自己的 API Key，并从控制台选择当前密钥和分组有权访问的模型 ID。不要把真实密钥写入共享配置。

[[link-button href="https://hermes-agent.nousresearch.com/docs/integrations/providers#general-setup" label="查看 Hermes Provider 配置文档"]]$after$),
                -- openclaw: seed
                ('openclaw', $before$## openclaw.json

[[command title="openclaw.json" lang="json"]]
{
  "models": {
    "providers": {
      "luoye": {
        "api": "openai-completions",
        "name": "luoye",
        "baseURL": "https://ai.3zapi.top/v1",
        "envKey": "LUOYE_API_KEY",
        "models": [
          {
            "id": "gpt-5.3-codex",
            "contextWindow": 400000
          }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": {
        "primary": "luoye/gpt-5.3-codex"
      }
    }
  }
}
[[/command]]$before$, $after$## openclaw.json

先在启动 OpenClaw 的环境中设置 `LUOYE_API_KEY`，不要在配置文件中写入真实密钥。以下以 OpenAI Chat Completions 为例，将各处 `YOUR_MODEL_ID` 替换为当前密钥和分组可用的模型 ID。若选 Responses 模式，请在客户端文档中确认对应设置，并使用与所选协议匹配的服务端端点；不要混用两种模式。

[[link-button href="https://docs.openclaw.ai/concepts/model-providers/custom-providers" label="查看 OpenClaw 自定义 Provider 文档"]]

[[command title="openclaw.json" lang="json"]]
{
  "models": {
    "providers": {
      "luoye": {
        "baseUrl": "https://ai.3zapi.com/v1",
        "apiKey": "${LUOYE_API_KEY}",
        "api": "openai-completions",
        "models": [{ "id": "YOUR_MODEL_ID", "name": "YOUR_MODEL_ID" }]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": { "primary": "luoye/YOUR_MODEL_ID" }
    }
  }
}
[[/command]]$after$)
            ) AS patches(slug, old_fragment, new_fragment)
            WHERE patches.slug = target.slug
        LOOP
            -- Check uniqueness in the original (also reject overlapping copies).
            -- Padding with LF checks line boundaries without requiring a particular
            -- document prefix/suffix or consuming an adjacent paragraph.
            match_at := strpos(E'\n' || target.content_md || E'\n', E'\n' || patch.old_fragment || E'\n');
            IF match_at > 0 AND strpos(
                substring(E'\n' || target.content_md || E'\n' FROM match_at + 1),
                E'\n' || patch.old_fragment || E'\n'
            ) = 0 THEN
                match_at := strpos(E'\n' || revised_content || E'\n', E'\n' || patch.old_fragment || E'\n');
                IF match_at > 0 THEN
                    revised_content := overlay(revised_content PLACING patch.new_fragment
                        FROM match_at FOR char_length(patch.old_fragment));
                END IF;
            END IF;
        END LOOP;

        IF target.slug = 'getting-started'
           AND target.description = '安装 Codex App，创建 codex 密钥，写入配置并启动验证。' THEN
            revised_description := '安装官方 ChatGPT，准备有模型权限的 API Key，再用配置助手或手动接入 Codex。';
        END IF;

        IF (revised_content, revised_description)
           IS DISTINCT FROM (target.content_md, target.description) THEN
            UPDATE tutorial_pages
            SET content_md = revised_content, description = revised_description, updated_at = NOW()
            WHERE id = target.id AND status = 'published'
              AND (content_md, description, updated_at) =
                  (target.content_md, target.description, target.updated_at);
        END IF;
    END LOOP;
END
$refresh$;
