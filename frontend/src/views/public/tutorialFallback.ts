import type { TutorialPage } from '@/types'

const timestamp = '2026-05-21T00:00:00Z'

export const tutorialFallbackPages: TutorialPage[] = [
  {
    id: 1,
    slug: 'getting-started',
    title: '快速开始',
    description: '准备账号、创建 API Key、确认 Base URL，然后选择要接入的工具。',
    category: '使用指南',
    sort_order: 10,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# 快速开始

第一次接入先按新手最快路线走，只需要准备账号、创建 Codex 密钥、写入配置并启动。Codex App 不需要先安装 Git 或 Node.js；这两个放到 CLI、Claude Code 或代码项目阶段再准备。CC Switch、Cockpit Tools、MinePilotQA 和其他工具都可以后续再看。

[[callout type="tip" title="新手最快路线"]]
1. 登录落叶网络控制台，打开 API 密钥页面。
2. 创建 Codex 类型密钥并复制 Token。
3. 打开 Codex 教程写入 config.toml 和 auth.json。
4. 需要 CLI、Claude Code 或代码项目协作时，再安装 Node.js/npm 和 Git。
[[/callout]]

## 准备工具

- Windows、Linux 或 macOS 电脑。
- 落叶网络 API Token 和 Base URL。
- Codex App：不把 Git、Node.js 作为启动前置条件。
- Codex CLI：需要 Node.js/npm，建议使用当前 LTS 版本；当前 npm 包声明支持 Node.js 16+。
- Claude Code：npm 安装路径需要 Node.js 18+；Windows 原生使用还会依赖 Git Bash 或 WSL。
- Git：不是 Codex App 的硬性要求，但做代码项目、版本回滚和团队协作时强烈建议安装。

[[command title="按需验证开发工具"]]
git --version
node -v
npm -v
[[/command]]

## 创建 API 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型按目标工具选择；Codex 用户请选择 codex。

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
`.trim()
  },
  {
    id: 2,
    slug: 'codex',
    title: 'Codex',
    description: 'Codex App / CLI 配置教程。',
    category: '工具配置',
    sort_order: 20,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# Codex 配置

新手优先使用 Codex App。Codex App 的安装和登录不要求先安装 Git 或 Node.js；Node.js/npm 主要用于 Codex CLI，Git 主要用于代码项目版本管理。先创建 codex 类型 API 密钥，然后写入 config.toml 和 auth.json。

## 安装 Codex App

[[command title="Windows winget 安装"]]
winget install --id OpenAI.Codex -e
[[/command]]

也可以从 Microsoft Store 或 https://chatgpt.com/codex 安装桌面端。

## 写入 config.toml

配置文件位置示例：C:/Users/用户名/.codex/config.toml。

[[command title="config.toml" lang="toml"]]
model = "gpt-5.5"
model_provider = "luoye"

[model_providers.luoye]
name = "luoye"
base_url = "https://ai.3zapi.top/v1"
env_key = "OPENAI_API_KEY"
wire_api = "responses"
[[/command]]

## 写入 auth.json

配置文件位置示例：C:/Users/用户名/.codex/auth.json。

[[command title="auth.json" lang="json"]]
{
  "OPENAI_API_KEY": "输入你的key"
}
[[/command]]

## CLI 用户可选安装

只有使用终端版 Codex CLI 时才需要 Node.js/npm。当前 npm 包声明支持 Node.js 16+，新机器建议直接安装当前 LTS 版本。

[[command title="npm 镜像安装 Codex CLI"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
[[/command]]
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
