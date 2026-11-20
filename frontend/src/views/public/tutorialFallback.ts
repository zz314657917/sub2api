import type { TutorialPage } from '@/types'
import { codexTutorialConfig } from './tutorialQuickstart'

const timestamp = '2026-05-21T00:00:00Z'

export const tutorialFallbackPages: TutorialPage[] = [
  {
    id: 1,
    slug: 'getting-started',
    title: '快速开始',
    description: '安装官方 ChatGPT，准备有模型权限的 API Key，再用配置助手或手动接入 Codex。',
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
1. 安装官方 ChatGPT 桌面应用。
2. 在 API 密钥页复制可用的 API Key；没有合适的密钥再创建。
3. 用 Windows 配置助手，或按 Codex 教程手动合并 config.toml 和 auth.json。
4. 重启应用，确认所选模型能正常响应。
[[/callout]]

## 安装官方 ChatGPT

打开官方 ChatGPT 下载页，安装桌面应用。本站 Windows 配置助手是独立工具，用来配置应用中的 Codex 功能，不是官方应用本身。

[[link-button href="https://chatgpt.com/download/" label="下载官方 ChatGPT"]]

[[link-button href="/tutorial" label="打开 Windows 配置助手教程"]]

## 获取 API Key

登录落叶网络控制台，打开 API 密钥页面。优先复用已启用且有目标模型权限的 API Key；如需新建，点“创建密钥”，厂商选 OpenAI，分组选择支持目标模型且有可用额度的分组。复制完整密钥，勿分享给他人。

[[link-button href="/keys" label="打开 API 密钥页面"]]

## Base URL 与协议

- Codex（Responses）：使用根地址 \`https://ai.3zapi.com\`，不加 \`/v1\`。
- 通用 OpenAI 兼容客户端：如果客户端要求 \`/v1\`，填写 \`https://ai.3zapi.com/v1\`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 \`https://ai.3zapi.com\`，不加 \`/v1\`，不要选 OpenAI 协议。

## 下一步：配置 Codex

Windows 用户可按配置助手教程操作。手动配置时，请先备份已有文件，再按 Codex 教程合并 config.toml 和 auth.json。

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

Codex 使用落叶网络的 OpenAI Responses 接口。先安装官方应用，准备有效的 API Key，再配置 config.toml 和 auth.json。

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

打开官方 ChatGPT 下载页，安装桌面应用。本站 Windows 配置助手是独立工具，用来配置应用中的 Codex 功能，不是官方应用本身。

[[link-button href="https://chatgpt.com/download/" label="下载官方 ChatGPT"]]

[[link-button href="/tutorial" label="打开 Windows 配置助手教程"]]

## Linux / macOS 安装 Codex CLI

Linux 和 macOS 通常使用 Codex CLI。先确认已经安装 Node.js/npm，然后全局安装 Codex。

[[command title="Linux / macOS npm 安装"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
codex --version
[[/command]]

不想走 npm 时，可以到官方 GitHub Release 下载对应架构的二进制包；Linux 常见架构是 x86_64 或 arm64。下载后把 codex 可执行文件放到 PATH 目录里，再执行 codex --version 验证。

[[link-button href="https://github.com/openai/codex/releases" label="打开 Codex Releases"]]

## 获取 API Key

登录落叶网络控制台，打开 API 密钥页面。优先复用已启用且有目标模型权限的 API Key；如需新建，点“创建密钥”，厂商选 OpenAI，分组选择支持目标模型且有可用额度的分组。复制完整密钥，勿分享给他人。

[[link-button href="/keys" label="打开 API 密钥页面"]]

## 第一步：找到 Codex 配置文件夹

Codex 会读取一个叫 .codex 的文件夹。这个文件夹就放在你当前电脑账号的用户目录里，不是项目目录，也不是浏览器下载目录。

- Windows 示例：C:/Users/你的用户名/.codex/
- macOS / Linux 示例：/home/你的用户名/.codex/ 或 /Users/你的用户名/.codex/

不知道用户名也没关系，直接复制下面命令执行即可，它会自动进入你的用户目录并创建 .codex 文件夹。

[[command title="Windows PowerShell 创建配置文件夹"]]
mkdir $env:USERPROFILE\\.codex -Force
explorer $env:USERPROFILE\\.codex
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

先备份已有的 config.toml 和 auth.json。若 config.toml 已存在，只合并下方相关字段和 Provider 表，不要覆盖其他设置或重复定义同名表。

\`\`\`toml
${codexTutorialConfig()}
\`\`\`

- \`model_provider = "OpenAI"\` 与 \`[model_providers.OpenAI]\` 对应；\`wire_api = "responses"\` 选择 Responses 协议，使用 \`https://ai.3zapi.com\` 根地址，不加 \`/v1\`。
- \`model\` 和 \`review_model\` 中的 \`gpt-6.1-sol\` 与快速接入生成器的示例一致，不保证所有分组可用；请改为当前密钥有权访问的模型。
- \`requires_openai_auth = true\` 与同目录 auth.json 的 \`OPENAI_API_KEY\` 配套使用。

## 第三步：创建 auth.json

auth.json 保存你的 API Key。将 API 密钥页复制的有效密钥填入 \`OPENAI_API_KEY\`，与上面的 config.toml 配套使用。

Windows 用户：在同一个 .codex 文件夹里新建 auth.json，再用记事本打开。

Linux / macOS 用户：

[[command title="Linux / macOS 打开 auth.json"]]
nano ~/.codex/auth.json
[[/command]]

把下面内容粘进去，并把中文提示替换成你的真实 API Key。

[[command title="auth.json" lang="json"]]
{
  "OPENAI_API_KEY": "替换成你的 API Key"
}
[[/command]]

如果 auth.json 已经存在，请先备份，只更新 \`OPENAI_API_KEY\`，保留其他配置和有效的 JSON 格式。

## 启动验证

关闭并重新打开 Codex App，进入一个项目后发送简单问题，确认模型能正常响应。

[[callout type="warning" title="连接失败时先检查"]]
1. Codex 的 Base URL 是否为 \`https://ai.3zapi.com\`，且使用 Responses。
2. auth.json 中的 API Key 是否完整、有效。
3. 当前密钥的分组是否支持所选模型，账户是否有可用余额或额度。
4. 修改后是否完全退出再重启，配置是否被其他账号切换工具覆盖。
[[/callout]]

## CLI 用户可选安装

只有使用终端版 Codex CLI 时才需要 Node.js/npm。Linux/macOS 用户通常走这一条；新机器建议直接安装当前 LTS 版本。

[[command title="npm 镜像安装 Codex CLI"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
codex --version
[[/command]]

安装后在终端执行 codex，CLI 会读取同一目录下的 config.toml 和 auth.json。

## 常见问题

- 401 或 unauthorized：检查 API Key 是否完整、有效且已启用。
- model not found：核对模型 ID 和分组权限；没有权限时，换用控制台中当前分组可用的模型。
- 配置不生效：完全退出并重新打开 Codex App；CLI 用户可打开新终端重试。

## Base URL 与协议

- Codex（Responses）：使用根地址 \`https://ai.3zapi.com\`，不加 \`/v1\`。
- 通用 OpenAI 兼容客户端：如果客户端要求 \`/v1\`，填写 \`https://ai.3zapi.com/v1\`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 \`https://ai.3zapi.com\`，不加 \`/v1\`，不要选 OpenAI 协议。
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

安装后打开新终端，运行 \`claude --version\`；遇到环境问题可运行 \`claude doctor\`。

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

CC Switch 适合需要多工具、多账号切换的用户。可从落叶网络控制台一键导入 Provider，也可以手动添加。

## 从 API 密钥页导入

1. 打开 [API 密钥页面](/keys)，找到已启用且有目标模型权限的密钥，在该行点击“导入到 CCS”。如果按钮未显示，可在 CC Switch 手动添加对应工具的 Provider。
2. 在 CC Switch 中确认导入，并检查目标工具、Base URL 和 Provider；选中落叶网络 Provider 后再使用。

导入链接和导出的配置可能含有完整 API 密钥。只在可信设备上打开，不要转发或上传。

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

1. 打开 [API 密钥页面](/keys)，找到已启用且有目标模型权限的 OpenAI 密钥，在该行点击“导入 Cockpit”，允许浏览器唤起 Cockpit Tools。
2. 如果唤起失败，在提示对话框中点击“下载导入 JSON”，再到 Cockpit Tools 的 Codex 页面手动导入该文件。
3. 在 Cockpit Tools 左侧点击 Codex 图标，检查导入账号的 Base URL，确认并选中落叶网络 Provider。

导入链接和 JSON 文件含有 API 密钥，不要分享或上传；导入后妥善保管或删除文件。

[[callout type="tip" title="恢复 Codex 会话"]]
导入账号后，如果 Codex 会话不可见，先确认当前 Provider 已选中落叶网络；具体操作见恢复历史会话教程。
[[/callout]]

[[link-button href="/tutorial/recover-codex-session" label="查看恢复历史会话"]]
`.trim()
  },
  {
    id: 6,
    slug: 'recover-codex-session',
    title: '恢复历史会话',
    description: '使用 Cockpit Tools 修复 Codex 历史会话可见性。',
    category: '排查',
    sort_order: 55,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# 恢复历史会话

如果 Codex App 已经登录但历史对话没有显示，可以用 Cockpit Tools 修复可见性。

## 操作步骤

1. 打开 Cockpit Tools。
2. 点击左侧 Codex 图标。

[[screenshot src="/tutorial/cockpit-tools/codex-sidebar.png" alt="Cockpit Tools 左侧 Codex 图标" caption="点击左侧 Codex 图标进入 Codex 页面"]]

3. 点击顶部的会话管理。
4. 点击修复可见性。

[[screenshot src="/tutorial/cockpit-tools/session-visibility.png" alt="Cockpit Tools 会话管理修复可见性" caption="进入会话管理后点击修复可见性"]]

5. 回到 Codex App，刷新或重新打开会话列表。

[[callout type="tip" title="看不到会话时请等待一下"]]
历史会话加载可能有延迟，修复后请等待一会儿，再刷新或重新打开会话列表。
[[/callout]]
`.trim()
  },
  {
    id: 7,
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
    id: 8,
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

先在启动 OpenClaw 的环境中设置 \`LUOYE_API_KEY\`，不要在配置文件中写入真实密钥。以下以 OpenAI Chat Completions 为例，将各处 \`YOUR_MODEL_ID\` 替换为当前密钥和分组可用的模型 ID。若选 Responses 模式，请在客户端文档中确认对应设置，并使用与所选协议匹配的服务端端点；不要混用两种模式。

[[link-button href="https://docs.openclaw.ai/concepts/model-providers/custom-providers" label="查看 OpenClaw 自定义 Provider 文档"]]

[[command title="openclaw.json" lang="json"]]
{
  "models": {
    "providers": {
      "luoye": {
        "baseUrl": "https://ai.3zapi.com/v1",
        "apiKey": "\${LUOYE_API_KEY}",
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
[[/command]]

[[command title="启动检查"]]
openclaw onboard
openclaw dashboard
[[/command]]
`.trim()
  },
  {
    id: 9,
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

## 安装

按 Nous Research 官方文档选择适合系统的安装方式。

[[link-button href="https://hermes-agent.nousresearch.com/docs/getting-started/installation" label="查看 Hermes Agent 官方安装说明"]]

## 配置模型

运行 \`hermes model\`，选择 Custom Endpoint（自定义端点）。按 OpenAI 兼容模式填写 Base URL \`https://ai.3zapi.com/v1\`，输入你自己的 API Key，并从控制台选择当前密钥和分组有权访问的模型 ID。不要把真实密钥写入共享配置。

[[link-button href="https://hermes-agent.nousresearch.com/docs/integrations/providers#general-setup" label="查看 Hermes Provider 配置文档"]]
`.trim()
  },
  {
    id: 10,
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

## 安装失败

先确认是哪类工具：Codex App 不需要先安装 Node.js；通过 npm 安装 Codex CLI 建议用当前 LTS。Claude Code 优先走 [Anthropic 官方原生安装](https://code.claude.com/docs/en/setup)，Windows 原生安装时 Git for Windows 可选；改用 npm 时要求 Node.js 22+。

## 命令找不到

Codex CLI 或 Claude Code 安装后，先新开终端，再用 \`codex --version\` 或 \`claude --version\` 检查命令是否在 PATH 中。只有通过 npm 安装时才需要同时检查 Node.js 和 npm；缺少 git 与应用本身没有装好是两回事。

## 鉴权失败

401 时检查 [API 密钥](/keys) 是否完整、有效且已启用。403 或模型不可用时，检查密钥的厂商、分组是否匹配客户端，确认分组有权访问该模型，模型 ID 拼写正确；同时检查账户余额和套餐额度。不要将真实密钥发给他人。

## Base URL 与协议

- Codex（Responses）：使用根地址 \`https://ai.3zapi.com\`，不加 \`/v1\`。
- 通用 OpenAI 兼容客户端：如果客户端要求 \`/v1\`，填写 \`https://ai.3zapi.com/v1\`，并让协议模式与所选端点匹配。
- Claude Code（Anthropic Messages）：填写 \`https://ai.3zapi.com\`，不加 \`/v1\`，不要选 OpenAI 协议。
`.trim()
  }
]
