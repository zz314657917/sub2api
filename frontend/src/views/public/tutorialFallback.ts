import type { TutorialPage } from '@/types'

const timestamp = '2026-05-21T00:00:00Z'

export const tutorialFallbackPages: TutorialPage[] = [
  {
    id: 1,
    slug: 'getting-started',
    title: '快速开始',
    description: '安装 Codex App，创建 codex 密钥，写入配置并启动验证。',
    category: '使用指南',
    sort_order: 10,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# 快速开始

按下面 4 步完成第一次接入；完整配置见 Codex 教程。

[[callout type="tip" title="新手最快路线"]]
1. 安装或打开 Codex App。
2. 登录落叶网络控制台，创建 codex 类型密钥并复制 Token。
3. 按 Codex 教程写入 config.toml 和 auth.json。
4. 启动 Codex App，确认模型能正常响应。
[[/callout]]

## 安装 Codex App
打开 Codex App 安装页，按系统提示安装桌面端。

[[link-button href="https://chatgpt.com/codex" label="打开 Codex App"]]

## 创建 API 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型请选择 codex。

[[screenshot src="/tutorial/api-key/sidebar-api-key.png" alt="API 密钥入口" caption="左侧进入 API 密钥页面"]]
[[screenshot src="/tutorial/api-key/create-key-button.png" alt="创建密钥按钮" caption="点击右上角创建密钥"]]
[[screenshot src="/tutorial/api-key/create-key-dialog.png" alt="创建密钥弹窗" caption="选择分组和密钥类型后创建"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]

## Base URL

OpenAI 兼容工具通常使用：

[[command title="OpenAI 兼容 Base URL"]]
https://ai.3zapi.top/v1
[[/command]]

Claude Code 兼容工具通常使用：

[[command title="Anthropic 兼容 Base URL"]]
https://ai.3zapi.top
[[/command]]

## 下一步：Codex 配置

打开 Codex 教程，复制配置文件示例并写入你的 Token。

[[link-button href="/tutorial/codex" label="查看 Codex 配置"]]
`.trim()
  },
  {
    id: 2,
    slug: 'codex',
    title: 'Codex',
    description: 'Codex App / CLI 详细配置教程。',
    category: '工具配置',
    sort_order: 20,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# Codex 配置

Codex 推荐使用落叶网络的 OpenAI 兼容接口。流程是：安装 Codex App，创建 codex 类型密钥，写入 config.toml 和 auth.json，然后启动验证。

[[callout type="tip" title="推荐路线"]]
新手先走 Codex App。Codex App 不要求先安装 Git 或 Node.js；Node.js/npm 只用于 Codex CLI，Git 主要用于代码项目版本管理和团队协作。
[[/callout]]

## 适用场景

- Codex App：推荐新手使用，先安装桌面端再写配置。
- Linux / macOS：优先使用 Codex CLI，需要 Node.js/npm；也可以下载 GitHub Release 二进制包。
- VS Code 或其他 IDE 插件：优先复用同一套 Base URL 和 API Key。

## 安装 Codex App

[[command title="Windows winget 安装"]]
winget install --id OpenAI.Codex -e
[[/command]]

也可以从 Microsoft Store 或 https://chatgpt.com/codex 安装桌面端。

## Linux / macOS 安装 Codex CLI

Linux 和 macOS 通常使用 Codex CLI。先确认已经安装 Node.js/npm，然后全局安装 Codex。

[[command title="Linux / macOS npm 安装"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
codex --version
[[/command]]

不想走 npm 时，可以到官方 GitHub Release 下载对应架构的二进制包；Linux 常见架构是 x86_64 或 arm64。下载后把 codex 可执行文件放到 PATH 目录里，再执行 codex --version 验证。

[[link-button href="https://github.com/openai/codex/releases" label="打开 Codex Releases"]]

## 创建 codex 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型请选择 codex。创建后复制 Token，后面写入 auth.json。

[[screenshot src="/tutorial/api-key/sidebar-api-key.png" alt="API 密钥入口" caption="左侧进入 API 密钥页面"]]
[[screenshot src="/tutorial/api-key/create-key-button.png" alt="创建密钥按钮" caption="点击右上角创建密钥"]]
[[screenshot src="/tutorial/api-key/create-key-dialog.png" alt="创建密钥弹窗" caption="密钥类型选择 codex"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]

## 第一步：找到 Codex 配置文件夹

Codex 会读取一个叫 .codex 的文件夹。这个文件夹就放在你当前电脑账号的用户目录里，不是项目目录，也不是浏览器下载目录。

- Windows 示例：C:/Users/你的用户名/.codex/
- macOS / Linux 示例：/home/你的用户名/.codex/ 或 /Users/你的用户名/.codex/

不知道用户名也没关系，直接复制下面命令执行即可，它会自动进入你的用户目录并创建 .codex 文件夹。

[[command title="Windows PowerShell 创建配置文件夹"]]
mkdir $env:USERPROFILE\.codex -Force
explorer $env:USERPROFILE\.codex
[[/command]]

[[command title="Linux / macOS 创建配置文件夹"]]
mkdir -p ~/.codex
cd ~/.codex
pwd
[[/command]]

## 第二步：创建 config.toml

config.toml 是 Codex 的主配置文件，用来告诉 Codex 使用哪个模型、哪个服务商、哪个 Base URL。

Windows 用户：在刚打开的 .codex 文件夹里新建文本文件，命名为 config.toml。注意不要变成 config.toml.txt。

Linux / macOS 用户：在终端执行下面命令打开编辑器；如果文件不存在会自动创建。

[[command title="Linux / macOS 打开 config.toml"]]
nano ~/.codex/config.toml
[[/command]]

把下面内容完整粘进去，然后保存。

[[command title="config.toml" lang="toml"]]
model = "gpt-5.5"
model_provider = "luoye"

[model_providers.luoye]
name = "luoye"
base_url = "https://ai.3zapi.top/v1"
env_key = "OPENAI_API_KEY"
wire_api = "responses"
[[/command]]

- model_provider = "luoye" 表示 Codex 使用下面这个自定义服务商。
- base_url 必须写成 \`https://ai.3zapi.top/v1\`，结尾带 \`/v1\`。
- env_key = "OPENAI_API_KEY" 要和 auth.json 里的字段名保持一致；CLI 环境变量方案也使用这个名字。
- wire_api = "responses" 表示使用 Responses 协议。

## 第三步：创建 auth.json

auth.json 用来保存你的 API Key。这里的 key 就是你在落叶网络控制台创建的 codex 类型密钥。

Windows 用户：在同一个 .codex 文件夹里新建 auth.json，再用记事本打开。

Linux / macOS 用户：

[[command title="Linux / macOS 打开 auth.json"]]
nano ~/.codex/auth.json
[[/command]]

把下面内容粘进去，并把中文提示替换成你的真实 API Key。

[[command title="auth.json" lang="json"]]
{
  "OPENAI_API_KEY": "替换成你的 codex 类型 API Key"
}
[[/command]]

如果 auth.json 已经存在，只需要更新 OPENAI_API_KEY 这一项，不要删除其他仍在使用的配置。

## 启动验证

关闭并重新打开 Codex App，进入一个项目后发送简单问题，确认模型能正常响应。

[[callout type="warning" title="失败时先查这三项"]]
1. Base URL 是否为 \`https://ai.3zapi.top/v1\`。
2. API Key 是否复制完整，并且密钥类型是 codex。
3. 当前分组是否有可用套餐和余额。
[[/callout]]

## CLI 用户可选安装

只有使用终端版 Codex CLI 时才需要 Node.js/npm。Linux/macOS 用户通常走这一条；新机器建议直接安装当前 LTS 版本。

[[command title="npm 镜像安装 Codex CLI"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
codex --version
[[/command]]

安装后在终端执行 codex，CLI 会读取同一目录下的 config.toml 和 auth.json。

## 常见问题

- 401 或 unauthorized：优先检查 Token 是否复制完整、密钥类型是否为 codex。
- model not found：当前分组可能不支持配置里的模型，换成控制台可用模型后再试。
- Base URL 报错：OpenAI 兼容地址必须写 \`https://ai.3zapi.top/v1\`，不要漏掉 \`/v1\`。
- 配置不生效：修改配置后重启 Codex App；如果是 CLI，关闭旧终端后重新打开。
`.trim()
  },
  {
    id: 3,
    slug: 'claude-code',
    title: 'Claude Code',
    description: 'Claude Code 命令行接入教程。',
    category: '工具配置',
    sort_order: 30,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# Claude Code 配置

Claude Code 使用 Anthropic 兼容环境变量。确认 API Token 后写入 ANTHROPIC_AUTH_TOKEN 和 ANTHROPIC_BASE_URL。

## 安装

npm 安装路径需要 Node.js 18+。Windows 用户如果不走 WSL，也要准备 Git for Windows / Git Bash；如果使用官方原生安装器，可以先按官方安装器流程执行，再用 claude doctor 检查环境。

[[command title="安装 Claude Code"]]
npm install -g @anthropic-ai/claude-code --registry=https://registry.npmmirror.com
claude --version
[[/command]]

## Windows PowerShell 环境变量

[[command title="PowerShell"]]
$env:ANTHROPIC_AUTH_TOKEN="你的令牌"
$env:ANTHROPIC_BASE_URL="https://ai.3zapi.top"

[Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN","你的令牌","User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL","https://ai.3zapi.top","User")
[[/command]]

## 兜底配置文件

[[command title="settings.json" lang="json"]]
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "你的令牌",
    "ANTHROPIC_BASE_URL": "https://ai.3zapi.top",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
[[/command]]
`.trim()
  },
  {
    id: 4,
    slug: 'cc-switch',
    title: 'CC Switch',
    description: '桌面端统一管理 Codex 和 Claude Code 账号。',
    category: '工具配置',
    sort_order: 40,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# CC Switch 账号管理

CC Switch 适合需要多工具、多账号切换的用户。推荐从落叶网络控制台一键导入 Provider，也可以手动添加。

## 功能点

- Provider 管理。
- MCP 服务器管理。
- Prompts 管理。
- 多平台支持。
- Deep Link 导入：ccswitch://。

## macOS 安装

[[command title="Homebrew"]]
brew tap farion1231/ccswitch
brew install --cask cc-switch
[[/command]]

## Web 版本

[[command title="Linux Web"]]
wget https://github.com/farion1231/cc-switch/releases/latest/download/cc-switch-web-linux-x64.tar.gz
tar -xzf cc-switch-web-linux-x64.tar.gz
cd cc-switch-web/
./cc-switch-web
# open http://localhost:17666
[[/command]]
`.trim()
  },
  {
    id: 5,
    slug: 'cockpit-tools',
    title: 'Cockpit Tools',
    description: '网页面板管理 Codex 账号和会话。',
    category: '工具配置',
    sort_order: 50,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# Cockpit Tools 账号管理

Cockpit Tools 适合用网页面板管理 Codex 账号和会话。项目地址：https://github.com/jlcodes99/cockpit-tools

## 导入落叶网络

[[command title="Cockpit Base URL"]]
https://ai.3zapi.top
[[/command]]

[[callout type="tip" title="恢复 Codex 会话"]]
导入账号后，如果 Codex 会话不可见，优先检查账号可见性和当前 Provider 是否选中。
[[/callout]]
`.trim()
  },
  {
    id: 6,
    slug: 'minepilotqa',
    title: 'MinePilotQA',
    description: '在 MinePilotQA 中添加 3zapi 服务商并测试连接。',
    category: '工具配置',
    sort_order: 60,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# MinePilotQA 接入

在 AI 配置里添加 3zapi 服务商，测试连接返回 HTTP 200 即成功。

[[command title="MinePilotQA Base URL"]]
https://ai.3zapi.top/
[[/command]]

- Core 正常。
- HTTP 200。
- Provider models endpoint reachable。
- models endpoint reachable。

[[screenshot src="/tutorial/minepilotqa/provider-list.png" alt="服务商列表" caption="进入 Provider 管理"]]
[[screenshot src="/tutorial/minepilotqa/add-provider.png" alt="添加服务商" caption="添加 3zapi 服务商"]]
[[screenshot src="/tutorial/minepilotqa/connection-success.png" alt="连接成功" caption="连接测试成功"]]
`.trim()
  },
  {
    id: 7,
    slug: 'openclaw',
    title: 'OpenClaw',
    description: '自托管助手平台，使用 OpenAI 兼容 provider 接入。',
    category: '工具配置',
    sort_order: 70,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# OpenClaw 接入

OpenClaw 使用 OpenAI 兼容 provider 接入落叶网络。

[[command title="openclaw.json" lang="json"]]
{
  "models": {
    "providers": {
      "luoye": {
        "type": "openai-completions",
        "baseUrl": "https://ai.3zapi.top/v1",
        "apiKeyEnv": "LUOYE_API_KEY"
      }
    }
  }
}
[[/command]]

[[command title="启动检查"]]
openclaw onboard
openclaw dashboard
[[/command]]
`.trim()
  },
  {
    id: 8,
    slug: 'hermes-agent',
    title: 'Hermes-Agent',
    description: '多代理开发工作流，配置自定义模型端点。',
    category: '工具配置',
    sort_order: 80,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# Hermes-Agent 接入

Hermes-Agent 可通过 Custom Endpoint 使用落叶网络模型。

[[command title="安装"]]
curl -fsSL https://raw.githubusercontent.com/terryso/hermes-agent/main/install.sh | bash
[[/command]]

[[command title="模型配置"]]
hermes model
# Provider: Custom Endpoint
# Base URL: https://ai.3zapi.top/v1
# Model: gpt-5.3-codex
[[/command]]
`.trim()
  },
  {
    id: 9,
    slug: 'faq',
    title: '常见问题',
    description: '密钥、Base URL、权限和工具连接的常见排查。',
    category: '排查',
    sort_order: 90,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# 常见问题

## 401 或未授权

检查密钥是否复制完整，是否选择了目标工具需要的密钥类型。

## Base URL 应该写哪个

- OpenAI 兼容工具通常写 https://ai.3zapi.top/v1。
- Claude Code 兼容工具通常写 https://ai.3zapi.top。

## 工具能启动但模型不可用

确认账号分组有可用套餐，模型名和工具协议匹配。
`.trim()
  }
]
