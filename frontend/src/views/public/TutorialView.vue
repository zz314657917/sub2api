<template>
  <div class="tutorial-page public-page-shell min-h-screen text-white">
    <PublicMatrixBackdrop />

    <PublicTopNav />

    <main class="tutorial-main relative z-10 mx-auto w-full px-4 py-8 sm:px-6 lg:py-10">
      <div class="tutorial-reader">
        <aside class="tutorial-sidebar" aria-label="接入 Agent 工具目录">
          <p class="tutorial-sidebar-title">接入 Agent 工具</p>
          <nav class="tutorial-tabs" aria-label="接入教程目录">
            <a
              v-for="item in sections"
              :key="item.id"
              :href="`#${item.id}`"
              :ref="(element) => setTabLink(item.id, element)"
              :class="{ 'is-active': activeSection === item.id }"
              :aria-current="activeSection === item.id ? 'true' : undefined"
              @click="handleIndexClick(item.id)"
            >
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </a>
          </nav>
        </aside>

        <div class="tutorial-main-column">
          <section id="quick-start" class="tutorial-overview">
            <div class="tutorial-intro">
              <div class="doc-pills" aria-label="教程分类">
                <span>使用指南</span>
                <span>API 文档</span>
                <strong>接入 Agent 工具</strong>
              </div>
              <h1>AI 接入教程</h1>
              <p>
                第一次接入先按“新手最快路线”走，只需要准备账号、创建 Codex 密钥、写入配置并启动。CC Switch、Cockpit Tools 和其他工具都可以先跳过。
              </p>

              <div class="beginner-path" aria-label="新手最快路线">
                <div class="beginner-path-head">
                  <span>新手最快路线</span>
                  <strong>只想先把 Codex 跑起来，看这 4 步</strong>
                </div>
                <div class="beginner-path-grid">
                  <a href="#prepare" class="beginner-step">
                    <span>1</span>
                    <strong>准备账号和环境</strong>
                    <em>安装 Git / Node.js，登录后打开 API 密钥页。</em>
                  </a>
                  <a href="#codex" class="beginner-step">
                    <span>2</span>
                    <strong>创建 Codex 密钥</strong>
                    <em>新建密钥时类型选 codex，复制 API Key。</em>
                  </a>
                  <a href="#codex" class="beginner-step">
                    <span>3</span>
                    <strong>写入两个配置文件</strong>
                    <em>按 Codex 段的 1-8 步写 config.toml 和 auth.json。</em>
                  </a>
                  <a href="#faq" class="beginner-step">
                    <span>4</span>
                    <strong>启动并排查</strong>
                    <em>打开 Codex，失败时先看常见问题。</em>
                  </a>
                </div>
              </div>

              <div class="overview-checklist" aria-label="接入步骤概览">
                <a href="#prepare" class="overview-row">
                  <span>01</span>
                  <strong>准备环境</strong>
                  <em>按系统安装 Git、Node.js，准备 Token 和代理地址。</em>
                </a>
                <a href="#platforms" class="overview-row">
                  <span>02</span>
                  <strong>Linux / macOS</strong>
                  <em>写入 Shell 环境变量，重开终端后生效。</em>
                </a>
                <a href="#cc-switch" class="overview-row">
                  <span>03</span>
                  <strong>CC Switch 账号管理</strong>
                  <em>桌面端统一管理 Codex 和 Claude Code 账号。</em>
                </a>
                <a href="#cockpit-tools" class="overview-row">
                  <span>04</span>
                  <strong>Cockpit Tools 账号管理</strong>
                  <em>网页面板管理 Codex 账号和会话。</em>
                </a>
                <a href="#codex" class="overview-row">
                  <span>05</span>
                  <strong>配置 Codex</strong>
                  <em>按 1-8 步写入配置文件并启动。</em>
                </a>
                <a href="#claude" class="overview-row">
                  <span>06</span>
                  <strong>配置 Claude Code</strong>
                  <em>写入 ANTHROPIC 变量，重启终端后启动。</em>
                </a>
                <a href="#openclaw" class="overview-row">
                  <span>07</span>
                  <strong>配置 OpenClaw</strong>
                  <em>自托管助手平台，使用 OpenAI 兼容 provider 接入。</em>
                </a>
                <a href="#hermes-agent" class="overview-row">
                  <span>08</span>
                  <strong>配置 Hermes-Agent</strong>
                  <em>终端 Agent，通过 Custom Endpoint 接入。</em>
                </a>
              </div>
            </div>

            <aside class="route-map" aria-label="接入路线图">
              <p>接入路线图</p>
              <div class="route-step" v-for="item in routeSteps" :key="item.id">
                <span>{{ item.step }}</span>
                <div>
                  <strong>{{ item.title }}</strong>
                  <em>{{ item.desc }}</em>
                </div>
              </div>
            </aside>
          </section>

          <section class="tutorial-content">
        <section id="prepare" class="guide-panel">
          <div class="guide-heading">
            <span>准备工作</span>
            <h2>先装基础工具</h2>
            <p>所有 AI 工具都需要 Git、Node.js、Token 和代理地址。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card">
              <div class="guide-card-head">
                <span>01</span>
                <PixelIcon name="shield" size="sm" tone="green" />
              </div>
              <h3>你需要准备</h3>
              <ul>
                <li>Windows、Linux 或 macOS 电脑；Claude Code 在 Windows 上建议 Windows 10 1809+。</li>
                <li>Node.js v18 或更高；教程链接为 v24。</li>
                <li>落叶网络 API Token 和 Base URL。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>02</span>
                <PixelIcon name="folder" size="sm" tone="green" />
              </div>
              <h3>Windows 安装 Git / Node.js</h3>
              <p>Windows 选择 64 位安装包：Git 下载 <code>Git-2.53.0-64-bit.exe</code>，Node.js 下载 <code>node-v24.13.1-x64.msi</code>。安装时一路下一步，路径保持默认。Linux 和 macOS 看下一节。</p>
              <ul class="tutorial-link-list">
                <li>
                  <a href="https://github.com/git-for-windows/git/releases/download/v2.53.0.windows.1/Git-2.53.0-64-bit.exe" target="_blank" rel="noopener noreferrer">
                    Git 2.53.0 64 位官方
                  </a>
                </li>
                <li>
                  <a href="https://registry.npmmirror.com/-/binary/git-for-windows/v2.51.0.windows.1/Git-2.51.0-64-bit.exe" target="_blank" rel="noopener noreferrer">
                    Git 2.51.0 64 位镜像
                  </a>
                </li>
                <li>
                  <a href="https://nodejs.org/dist/v24.13.1/node-v24.13.1-x64.msi" target="_blank" rel="noopener noreferrer">
                    Node.js 24.13.1 x64 官方
                  </a>
                </li>
                <li>
                  <a href="https://npmmirror.com/mirrors/node/v24.13.0/node-v24.13.0-x64.msi" target="_blank" rel="noopener noreferrer">
                    Node.js 24.13.0 x64 镜像
                  </a>
                </li>
              </ul>
            </article>

            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>03</span>
                <PixelIcon name="signal" size="sm" tone="green" />
              </div>
              <h3>检查是否安装成功</h3>
              <p>打开 CMD 或 PowerShell 执行命令。能看到版本号就继续。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制验证安装命令"
                  @click="copyCommand('verify-install', commands.verifyInstall)"
                >
                  {{ copiedCommand === 'verify-install' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.verifyInstall }}</code></pre>
              </div>
            </article>

            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>04</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>先拿到 Token / Base URL</h3>
              <ul>
                <li>登录落叶网络控制台。</li>
                <li>打开「API 密钥」。</li>
                <li>点击「新建密钥」。</li>
                <li>复制生成的 Token，并复制页面提供的 Base URL。</li>
                <li>后面所有工具里，Token 填 API Key，Base URL 填代理地址。</li>
              </ul>
              <router-link :to="apiKeysLink" class="guide-action-link">
                <PixelIcon name="key" size="xs" />
                {{ authStore.isAuthenticated ? '打开 API 密钥页面' : '登录后打开 API 密钥页面' }}
              </router-link>
            </article>
          </div>
        </section>

        <section id="platforms" class="guide-panel">
          <div class="guide-heading">
            <span>Linux / macOS</span>
            <h2>Linux 和 macOS 环境配置</h2>
            <p>Linux 默认示例使用 Bash，macOS 默认示例使用 Zsh。Token 换成控制台生成的值即可。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card">
              <div class="guide-card-head">
                <span>L1</span>
                <PixelIcon name="cube" size="sm" tone="green" />
              </div>
              <h3>Linux 基础环境</h3>
              <p>Ubuntu / Debian 可用以下命令准备 Git、Node.js，并安装 Codex 与 Claude Code。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Linux 基础环境命令"
                  @click="copyCommand('linux-install', commands.linuxInstall)"
                >
                  {{ copiedCommand === 'linux-install' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.linuxInstall }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>L2</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>Linux Shell 配置</h3>
              <p>写入 <code>~/.bashrc</code> 后执行 <code>source ~/.bashrc</code>，新终端也会自动生效。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Linux Shell 配置命令"
                  @click="copyCommand('linux-env', commands.linuxEnv)"
                >
                  {{ copiedCommand === 'linux-env' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.linuxEnv }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>M1</span>
                <PixelIcon name="folder" size="sm" tone="green" />
              </div>
              <h3>macOS 基础环境</h3>
              <p>如果已经安装 Homebrew，直接用 Brew 准备 Git、Node.js，再安装两个命令行工具。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 macOS 基础环境命令"
                  @click="copyCommand('mac-install', commands.macInstall)"
                >
                  {{ copiedCommand === 'mac-install' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.macInstall }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>M2</span>
                <PixelIcon name="settings" size="sm" tone="green" />
              </div>
              <h3>macOS Shell 配置</h3>
              <p>macOS 默认使用 Zsh，写入 <code>~/.zshrc</code> 后重新打开终端。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 macOS Shell 配置命令"
                  @click="copyCommand('mac-env', commands.macEnv)"
                >
                  {{ copiedCommand === 'mac-env' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.macEnv }}</code></pre>
              </div>
            </article>
          </div>
        </section>

        <section id="cc-switch" class="guide-panel">
          <div class="guide-heading">
            <span>CC Switch</span>
            <h2>CC Switch 账号管理</h2>
            <p>推荐从落叶网络控制台一键导入 Provider。手动配置保留为兜底方案。</p>
          </div>

          <div id="cc-switch-features" class="doc-feature-stack" aria-label="CC Switch 核心特性">
            <article class="doc-feature">
              <h3>Provider 管理</h3>
              <ul>
                <li>在 Claude Code、Codex、Gemini CLI 的 Provider 配置之间切换。</li>
                <li>每个 Provider 可维护多个端点，适合多账号和多模型入口。</li>
                <li>模型配置可区分主模型、轻量模型、均衡模型和高能力模型。</li>
              </ul>
            </article>

            <article class="doc-feature">
              <h3>MCP 服务器管理</h3>
              <ul>
                <li>统一管理 Claude、Codex、Gemini 三端可用的 MCP 服务。</li>
                <li>适合把本地工具、HTTP 服务和 SSE 服务集中维护。</li>
                <li>多端同步后，不需要在每个工具里重复编辑配置。</li>
              </ul>
            </article>

            <article class="doc-feature">
              <h3>Prompts 管理</h3>
              <ul>
                <li>集中管理系统提示词和项目提示词。</li>
                <li>覆盖 Claude 的 <code>CLAUDE.md</code>、Codex 的 <code>AGENTS.md</code>、Gemini 的 <code>GEMINI.md</code>。</li>
                <li>适合多套工作流提示词快速切换。</li>
              </ul>
            </article>

            <article class="doc-feature">
              <h3>多平台支持</h3>
              <ul>
                <li>Windows、macOS、Linux 都有桌面安装方案。</li>
                <li>无头服务器或 SSH 环境可使用 Web 版本。</li>
                <li>熟悉终端的用户也可以选择 CLI 方式管理。</li>
              </ul>
            </article>
          </div>

          <div id="cc-switch-import" class="doc-subsection">
            <h3>落叶网络接入方法</h3>
            <p>CC Switch 支持 <code>ccswitch://</code> Deep Link。落叶网络控制台的 API 密钥页已经提供导入入口，适合新手优先使用。</p>
            <ol class="doc-steps">
              <li>进入「API 密钥」页面，找到要使用的密钥。</li>
              <li>点击该密钥右侧的「导入 CC Switch」。</li>
              <li>系统会唤起 CC Switch，并弹出 Provider 配置窗口。</li>
              <li>在弹窗里确认应用类型、名称和模型选择，然后打开 CC Switch 完成导入。</li>
            </ol>
            <router-link :to="apiKeysLink" class="guide-action-link">
              <PixelIcon name="key" size="xs" />
              {{ authStore.isAuthenticated ? '打开 API 密钥页面' : '登录后打开 API 密钥页面' }}
            </router-link>
          </div>

          <div id="cc-switch-install" class="doc-subsection">
            <h3>安装方式</h3>
            <div class="install-grid">
              <article class="guide-card">
                <div class="guide-card-head">
                  <span>M1</span>
                  <PixelIcon name="settings" size="sm" tone="green" />
                </div>
                <h4>macOS 推荐 Homebrew</h4>
                <div class="command-block">
                  <button
                    type="button"
                    class="copy-command-button"
                    aria-label="复制 CC Switch macOS 安装命令"
                    @click="copyCommand('cc-switch-mac-install', commands.ccSwitchMacInstall)"
                  >
                    {{ copiedCommand === 'cc-switch-mac-install' ? '已复制' : '复制' }}
                  </button>
                  <pre><code>{{ commands.ccSwitchMacInstall }}</code></pre>
                </div>
              </article>

              <article class="guide-card">
                <div class="guide-card-head">
                  <span>W1</span>
                  <PixelIcon name="folder" size="sm" tone="green" />
                </div>
                <h4>Windows</h4>
                <p>下载 <code>.msi</code> 安装包，或下载 Portable / <code>.zip</code> 免安装版。</p>
                <ul class="tutorial-link-list">
                  <li>
                    <a href="https://ccswitch.ai/" target="_blank" rel="noopener noreferrer">
                      官方介绍
                    </a>
                  </li>
                  <li>
                    <a href="https://github.com/farion1231/cc-switch/releases" target="_blank" rel="noopener noreferrer">
                      Releases 下载
                    </a>
                  </li>
                </ul>
              </article>

              <article class="guide-card">
                <div class="guide-card-head">
                  <span>L1</span>
                  <PixelIcon name="cube" size="sm" tone="green" />
                </div>
                <h4>Linux</h4>
                <p>从 Releases 下载 <code>.deb</code> 或 <code>.AppImage</code>。ArchLinux 可用以下命令。</p>
                <div class="command-block">
                  <button
                    type="button"
                    class="copy-command-button"
                    aria-label="复制 CC Switch ArchLinux 安装命令"
                    @click="copyCommand('cc-switch-arch-install', commands.ccSwitchArchInstall)"
                  >
                    {{ copiedCommand === 'cc-switch-arch-install' ? '已复制' : '复制' }}
                  </button>
                  <pre><code>{{ commands.ccSwitchArchInstall }}</code></pre>
                </div>
              </article>

              <article class="guide-card">
                <div class="guide-card-head">
                  <span>WEB</span>
                  <PixelIcon name="panel" size="sm" tone="green" />
                </div>
                <h4>Web 版本</h4>
                <p>适合无头服务器、SSH 远程环境，启动后默认访问 <code>http://localhost:17666</code>。</p>
                <div class="command-block">
                  <button
                    type="button"
                    class="copy-command-button"
                    aria-label="复制 CC Switch Web 版本启动命令"
                    @click="copyCommand('cc-switch-web-install', commands.ccSwitchWebInstall)"
                  >
                    {{ copiedCommand === 'cc-switch-web-install' ? '已复制' : '复制' }}
                  </button>
                  <pre><code>{{ commands.ccSwitchWebInstall }}</code></pre>
                </div>
              </article>
            </div>
          </div>

          <div id="cc-switch-fallback" class="doc-subsection">
            <h3>手动配置兜底</h3>
            <p>如果浏览器没有唤起 CC Switch，再手动添加 Provider。</p>
            <ul>
              <li>打开 CC Switch，点击 <code>Add Provider</code>。</li>
              <li>应用选择 <code>Codex</code>、<code>Claude Code</code> 或 <code>Gemini CLI</code>。</li>
              <li>API Key 填落叶网络控制台生成的 Token。</li>
              <li>Base URL 填 <code>https://ai.3zapi.top</code>。</li>
              <li>保存后点击 <code>Enable</code>，重启对应终端或工具。</li>
            </ul>
          </div>
        </section>

        <section id="cockpit-tools" class="guide-panel">
          <div class="guide-heading">
            <span>Cockpit Tools</span>
            <h2>Cockpit Tools 账号管理</h2>
            <p>用于在网页面板里管理 Codex 账号和会话。部署或下载后，添加落叶网络账号即可启动。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>T1</span>
                <PixelIcon name="folder" size="sm" tone="green" />
              </div>
              <h3>部署或下载 Cockpit Tools</h3>
              <p>可以按仓库说明部署，也可以直接下载 release 里的安装包。</p>
              <ul class="tutorial-link-list">
                <li>
                  <a href="https://github.com/jlcodes99/cockpit-tools" target="_blank" rel="noopener noreferrer">
                    项目仓库
                  </a>
                </li>
                <li>
                  <a href="https://github.com/jlcodes99/cockpit-tools/releases/tag/v0.23.2" target="_blank" rel="noopener noreferrer">
                    v0.23.2 下载
                  </a>
                </li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>T2</span>
                <PixelIcon name="settings" size="sm" tone="green" />
              </div>
              <h3>添加 Codex 账号</h3>
              <ul>
                <li>打开 Cockpit Tools 面板。</li>
                <li>选择 <code>Codex</code>。</li>
                <li>点击“添加账号”。</li>
                <li><code>APIKEY</code> 填入你的落叶网络秘钥。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>T3</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>填写接口地址</h3>
              <p>域名填写落叶网络代理地址，秘钥填写控制台生成的 API Key。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Cockpit Tools 接口地址"
                  @click="copyCommand('cockpit-base-url', commands.cockpitBaseUrl)"
                >
                  {{ copiedCommand === 'cockpit-base-url' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.cockpitBaseUrl }}</code></pre>
              </div>
              <ul>
                <li>保存账号配置。</li>
                <li>点击启动。</li>
                <li>能正常打开 Codex 会话，即接入成功。</li>
              </ul>
            </article>

            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>T4</span>
                <PixelIcon name="book" size="sm" tone="green" />
              </div>
              <h3>恢复 Codex 会话</h3>
              <ul>
                <li>打开 Cockpit Tools。</li>
                <li>点击 <code>Codex</code>。</li>
                <li>进入“会话管理”。</li>
                <li>点击“修复可见性”，即可恢复会话显示。</li>
              </ul>
            </article>
          </div>
        </section>

        <section id="codex" class="guide-panel">
          <div class="guide-heading">
            <span>Codex</span>
            <h2>Codex 配置</h2>
            <p>参考 Codex App 配置教程：新手优先使用桌面端，配置文件和 VSCode 插件走同一套 <code>.codex</code> 目录。</p>
          </div>

          <div class="guide-grid guide-grid--single">
            <article class="guide-card">
              <div class="guide-card-head">
                <span>1</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>先创建 Codex 类型 API 密钥</h3>
              <ul>
                <li>登录落叶AI 控制台，进入「API 密钥」。</li>
                <li>点击右上角「创建密钥」。</li>
                <li>名称可以自定义，密钥类型选择 <code>codex</code>。</li>
                <li>额度、分组、速率限制等保持默认即可。</li>
                <li>创建后复制 API Key，后面写入 <code>auth.json</code>。</li>
              </ul>
              <router-link :to="apiKeysLink" class="guide-action-link">
                <PixelIcon name="key" size="xs" />
                {{ authStore.isAuthenticated ? '打开 API 密钥页面' : '登录后打开 API 密钥页面' }}
              </router-link>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>2</span>
                <PixelIcon name="folder" size="sm" tone="green" />
              </div>
              <h3>安装 Codex App</h3>
              <p>Windows 用户推荐从 Microsoft Store 安装；打不开商店时可用 <code>winget</code>，也可以访问 Codex 官网下载。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex App 安装命令"
                  @click="copyCommand('codex-app-install', commands.codexAppInstall)"
                >
                  {{ copiedCommand === 'codex-app-install' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexAppInstall }}</code></pre>
              </div>
              <ul class="tutorial-link-list">
                <li>
                  <a href="https://chatgpt.com/codex" target="_blank" rel="noopener noreferrer">
                    Codex 官网
                  </a>
                </li>
                <li>
                  <a href="https://apps.microsoft.com/detail/9ntx2k95jp4w" target="_blank" rel="noopener noreferrer">
                    Microsoft Store
                  </a>
                </li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>3</span>
                <PixelIcon name="settings" size="sm" tone="green" />
              </div>
              <h3>打开 Codex 配置文件</h3>
              <ul>
                <li>首次启动 Codex App 时选择 API 方式进入主界面。</li>
                <li>打开设置，进入配置区域，点击打开 <code>config.toml</code>。</li>
                <li>也可以直接编辑 <code>C:\Users\用户名\.codex\config.toml</code>。</li>
                <li>CLI 和 VSCode 插件也会读取同一份配置。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>4</span>
                <PixelIcon name="settings" size="sm" tone="green" />
              </div>
              <h3>写入 config.toml</h3>
              <p>把下面内容复制到 <code>config.toml</code>。模型名可按控制台可用模型调整，代理地址固定使用落叶网络 Base URL。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex config.toml 配置"
                  @click="copyCommand('codex-config-toml', commands.codexConfigToml)"
                >
                  {{ copiedCommand === 'codex-config-toml' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexConfigToml }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>5</span>
                <PixelIcon name="shield" size="sm" tone="green" />
              </div>
              <h3>写入 auth.json</h3>
              <p>编辑 <code>C:\Users\用户名\.codex\auth.json</code>，把「输入你的key」替换为刚刚创建的 Codex 类型 API Key。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex auth.json 配置"
                  @click="copyCommand('codex-auth-json', commands.codexAuthJson)"
                >
                  {{ copiedCommand === 'codex-auth-json' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexAuthJson }}</code></pre>
              </div>
              <ul>
                <li>如果文件不存在，就新建 <code>.codex</code> 文件夹和 <code>auth.json</code>。</li>
                <li>不要把真实 API Key 发给别人，也不要提交到仓库。</li>
                <li>保存后重启 Codex App 或重新打开 VSCode。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>6</span>
                <PixelIcon name="cursor" size="sm" tone="green" />
              </div>
              <h3>启动 Codex</h3>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex 启动命令"
                  @click="copyCommand('codex-start', commands.codexStart)"
                >
                  {{ copiedCommand === 'codex-start' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexStart }}</code></pre>
              </div>
              <p>桌面端重新打开应用即可使用；命令行用户可以在项目目录运行 <code>codex</code>。</p>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>7</span>
                <PixelIcon name="book" size="sm" tone="green" />
              </div>
              <h3>VSCode 集成 Codex</h3>
              <ul>
                <li>在 VSCode 扩展市场安装 <code>Codex - OpenAI's coding agent</code>。</li>
                <li>确认插件读取的是同一个 <code>.codex</code> 目录。</li>
                <li>如果 Codex App 闪退或系统暂不兼容，可先使用 VSCode 插件。</li>
                <li>修改配置后重启 VSCode，再打开 Codex 面板测试。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>8</span>
                <PixelIcon name="cursor" size="sm" tone="green" />
              </div>
              <h3>CLI 用户可选安装</h3>
              <p>只想用命令行时，再安装 <code>@openai/codex</code>。配置仍然读取上面的 <code>config.toml</code> 和 <code>auth.json</code>。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex CLI 镜像安装命令"
                  @click="copyCommand('codex-install-mirror', commands.codexInstallMirror)"
                >
                  {{ copiedCommand === 'codex-install-mirror' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexInstallMirror }}</code></pre>
              </div>
            </article>
          </div>
        </section>

        <section id="claude" class="guide-panel">
          <div class="guide-heading">
            <span>Claude Code</span>
            <h2>Claude Code 配置</h2>
            <p>安装后写入两个 ANTHROPIC 环境变量，然后重启终端。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card">
              <div class="guide-card-head">
                <span>A1</span>
                <PixelIcon name="cube" size="sm" tone="green" />
              </div>
              <h3>安装 Claude Code</h3>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Claude Code 常规安装命令"
                  @click="copyCommand('claude-install', commands.claudeInstall)"
                >
                  {{ copiedCommand === 'claude-install' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.claudeInstall }}</code></pre>
              </div>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Claude Code 镜像安装命令"
                  @click="copyCommand('claude-install-mirror', commands.claudeInstallMirror)"
                >
                  {{ copiedCommand === 'claude-install-mirror' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.claudeInstallMirror }}</code></pre>
              </div>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Claude Code 版本验证命令"
                  @click="copyCommand('claude-version', commands.claudeVersion)"
                >
                  {{ copiedCommand === 'claude-version' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.claudeVersion }}</code></pre>
              </div>
            </article>

            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>A2</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>推荐：PowerShell 配置</h3>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Claude Code PowerShell 配置命令"
                  @click="copyCommand('claude-env', commands.claudeEnv)"
                >
                  {{ copiedCommand === 'claude-env' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.claudeEnv }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>A3</span>
                <PixelIcon name="panel" size="sm" tone="green" />
              </div>
              <h3>备用：Windows 手动配置</h3>
              <ul>
                <li><code>Win + R</code> 输入 <code>sysdm.cpl</code>。</li>
                <li>打开“高级” → “环境变量”。</li>
                <li>新增 <code>ANTHROPIC_AUTH_TOKEN</code> 和 <code>ANTHROPIC_BASE_URL</code>。</li>
                <li>保存后重新打开终端。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>A4</span>
                <PixelIcon name="book" size="sm" tone="green" />
              </div>
              <h3>兜底配置文件</h3>
              <p>如果环境变量无效，再写入 <code>%userprofile%\.claude\settings.json</code>。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Claude Code settings.json 配置"
                  @click="copyCommand('claude-settings', commands.claudeSettings)"
                >
                  {{ copiedCommand === 'claude-settings' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.claudeSettings }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>A5</span>
                <PixelIcon name="cursor" size="sm" tone="green" />
              </div>
              <h3>启动与 IDE 接入</h3>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Claude Code 启动命令"
                  @click="copyCommand('claude-start', commands.claudeStart)"
                >
                  {{ copiedCommand === 'claude-start' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.claudeStart }}</code></pre>
              </div>
              <p>按提示回车进入对话。VS Code、Cursor、Trae 可通过插件连接本地 Claude Code。</p>
            </article>
          </div>
        </section>

        <section id="openclaw" class="guide-panel">
          <div class="guide-heading">
            <span>OpenClaw</span>
            <h2>OpenClaw 接入</h2>
            <p>OpenClaw 通过 <code>models.providers</code> 接入 OpenAI 兼容网关。把落叶网络作为 provider 写入 <code>openclaw.json</code> 后即可使用。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card">
              <div class="guide-card-head">
                <span>O1</span>
                <PixelIcon name="folder" size="sm" tone="green" />
              </div>
              <h3>安装或启动 OpenClaw</h3>
              <p>按 OpenClaw 官方文档完成安装，先跑完 onboard，并确认 Gateway 与 Control UI 可以打开。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 OpenClaw 初始化命令"
                  @click="copyCommand('openclaw-setup', commands.openclawSetup)"
                >
                  {{ copiedCommand === 'openclaw-setup' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.openclawSetup }}</code></pre>
              </div>
              <ul class="tutorial-link-list">
                <li>
                  <a href="https://docs.easyrouter.io/zh/docs/apps/openclaw" target="_blank" rel="noopener noreferrer">
                    OpenClaw 接入文档
                  </a>
                </li>
              </ul>
            </article>

            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>O2</span>
                <PixelIcon name="settings" size="sm" tone="green" />
              </div>
              <h3>配置 OpenAI 兼容 provider</h3>
              <p>在 <code>~/.openclaw/openclaw.json</code> 中新增 provider。API Key 推荐用环境变量注入，Base URL 需要带 <code>/v1</code>。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 OpenClaw provider 配置"
                  @click="copyCommand('openclaw-config', commands.openclawConfig)"
                >
                  {{ copiedCommand === 'openclaw-config' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.openclawConfig }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>O3</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>选择模型</h3>
              <ul>
                <li><code>models.providers.luoye.models</code> 里列出你准备使用的模型。</li>
                <li><code>agents.defaults.model.primary</code> 使用 <code>luoye/模型ID</code> 格式。</li>
                <li>模型名建议先选 <code>gpt-5.3-codex</code> 或控制台模型广场里可用的 OpenAI 兼容模型。</li>
                <li>如果模型返回不可用，先到模型广场确认当前分组是否支持该模型。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>O4</span>
                <PixelIcon name="signal" size="sm" tone="green" />
              </div>
              <h3>验证连接</h3>
              <p>保存配置后重启 OpenClaw，打开 dashboard 或列出模型。能看到 <code>luoye/</code> 前缀模型并收到回复即代表接入成功。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 OpenClaw 验证命令"
                  @click="copyCommand('openclaw-check', commands.openclawCheck)"
                >
                  {{ copiedCommand === 'openclaw-check' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.openclawCheck }}</code></pre>
              </div>
            </article>
          </div>
        </section>

        <section id="hermes-agent" class="guide-panel">
          <div class="guide-heading">
            <span>Hermes-Agent</span>
            <h2>Hermes-Agent 接入</h2>
            <p>Hermes-Agent 通过 <code>hermes model</code> 交互式配置模型，选择 Custom Endpoint 后填入落叶网络地址和 Token。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card">
              <div class="guide-card-head">
                <span>H1</span>
                <PixelIcon name="cube" size="sm" tone="green" />
              </div>
              <h3>安装 Hermes-Agent</h3>
              <p>按官方脚本安装 Hermes-Agent，确认 <code>hermes</code> 命令可用。Windows 原生环境不支持，建议先进入 WSL2。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Hermes-Agent 安装命令"
                  @click="copyCommand('hermes-install', commands.hermesInstall)"
                >
                  {{ copiedCommand === 'hermes-install' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.hermesInstall }}</code></pre>
              </div>
              <ul class="tutorial-link-list">
                <li>
                  <a href="https://docs.easyrouter.io/zh/docs/apps/hermes-agent" target="_blank" rel="noopener noreferrer">
                    Hermes-Agent 接入文档
                  </a>
                </li>
              </ul>
            </article>

            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>H2</span>
                <PixelIcon name="settings" size="sm" tone="green" />
              </div>
              <h3>启动模型配置向导</h3>
              <p>运行 <code>hermes model</code>，在交互中选择自定义 OpenAI 兼容接口。</p>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Hermes-Agent 模型配置命令"
                  @click="copyCommand('hermes-model', commands.hermesModel)"
                >
                  {{ copiedCommand === 'hermes-model' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.hermesModel }}</code></pre>
              </div>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>H3</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>填写 Custom Endpoint</h3>
              <ul>
                <li>Provider 选择 <code>Custom Endpoint</code>。</li>
                <li>Base URL 填 <code>https://ai.3zapi.top</code>。</li>
                <li>API Key 填落叶网络控制台生成的 Token。</li>
                <li>Model 填 <code>gpt-5.3-codex</code> 或模型广场中可用的模型名。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>H4</span>
                <PixelIcon name="cursor" size="sm" tone="green" />
              </div>
              <h3>启动并验证</h3>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Hermes-Agent 启动命令"
                  @click="copyCommand('hermes-start', commands.hermesStart)"
                >
                  {{ copiedCommand === 'hermes-start' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.hermesStart }}</code></pre>
              </div>
              <p>发起一次简单任务，能收到模型回复即接入成功。</p>
            </article>
          </div>
        </section>

        <section id="faq" class="guide-panel">
          <div class="guide-heading">
            <span>FAQ</span>
            <h2>常见问题先看这里</h2>
            <p>大多数失败都集中在 Node 版本、终端未重启、Token 或 Base URL 填错。</p>
          </div>

          <div class="faq-grid">
            <article v-for="item in faqItems" :key="item.title" class="faq-card">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </section>
          </section>
        </div>

      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import PixelIcon from '@/components/icons/PixelIcon.vue'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, type ComponentPublicInstance } from 'vue'
import { useAuthStore } from '@/stores'
import PublicMatrixBackdrop from './components/PublicMatrixBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'

const authStore = useAuthStore()
const apiKeysLink = computed(() =>
  authStore.isAuthenticated
    ? '/keys'
    : { path: '/login', query: { redirect: '/keys' } }
)

const sections = [
  { id: 'quick-start', title: '最快路线', desc: '新手先看' },
  { id: 'prepare', title: '准备工作', desc: '安装必备软件' },
  { id: 'platforms', title: 'Linux / macOS', desc: '跨平台配置' },
  { id: 'cc-switch', title: 'CC Switch', desc: '可选账号管理' },
  { id: 'cockpit-tools', title: 'Cockpit Tools', desc: '可选网页面板' },
  { id: 'codex', title: 'Codex', desc: 'App / 插件配置' },
  { id: 'claude', title: 'Claude', desc: '命令行使用' },
  { id: 'openclaw', title: 'OpenClaw', desc: '自托管助手' },
  { id: 'hermes-agent', title: 'Hermes', desc: '终端 Agent' },
  { id: 'faq', title: '常见问题', desc: '快速排查' },
]

const routeSteps = [
  { id: 'quick-start', step: '01', title: '先走最快路线', desc: '新手只按 4 步把 Codex 跑起来。' },
  { id: 'prepare', step: '02', title: '装好基础软件', desc: 'Git、Node.js、Token 和代理地址准备好。' },
  { id: 'codex', step: '03', title: '使用 Codex', desc: '创建 codex 密钥，写入 config.toml 和 auth.json。' },
  { id: 'cc-switch', step: '04', title: '可选：CC Switch', desc: '需要多工具切换账号时再用。' },
  { id: 'cockpit-tools', step: '05', title: '可选：Cockpit Tools', desc: '需要网页面板管理会话时再用。' },
  { id: 'claude', step: '06', title: '可选：Claude', desc: '写入 ANTHROPIC 变量后重启终端。' },
  { id: 'openclaw', step: '07', title: '可选：OpenClaw', desc: '写入 OpenAI 兼容 provider。' },
  { id: 'hermes-agent', step: '08', title: '可选：Hermes-Agent', desc: '选择 Custom Endpoint 后填入地址和 Token。' },
]

const faqItems = [
  { title: '安装失败', desc: '换镜像源，再确认 Node.js 版本是 v18 或更高。' },
  { title: '配置不生效', desc: '保存后关闭终端重开，仍不行就重启电脑。' },
  { title: '命令找不到', desc: 'Git 和 Node.js 建议默认安装，不要手动改路径。' },
  { title: '鉴权失败', desc: '检查 Token 和 Base URL，注意不要多复制空格。' },
  { title: '网络异常', desc: 'npm 安装用镜像源，运行时确认代理地址可访问。' },
  { title: '查看额度', desc: '到落叶网络控制台查看用量、记录和剩余额度。' },
]

const commands = {
  verifyInstall: `git --version
node -v
npm -v`,
  codexInstallMirror: 'npm install -g @openai/codex --registry=https://registry.npmmirror.com',
  codexAppInstall: 'winget install --id OpenAI.Codex -e',
  codexConfigToml: `model = "gpt-5.5"
model_provider = "luoye"

[model_providers.luoye]
name = "luoye"
base_url = "https://ai.3zapi.top/v1"
env_key = "OPENAI_API_KEY"
wire_api = "responses"`,
  codexAuthJson: `{
  "OPENAI_API_KEY": "输入你的key"
}`,
  cockpitBaseUrl: 'https://ai.3zapi.top',
  linuxInstall: `sudo apt update
sudo apt install -y git nodejs npm
npm install -g @openai/codex @anthropic-ai/claude-code --registry=https://registry.npmmirror.com
git --version
node -v
npm -v`,
  linuxEnv: `cat >> ~/.bashrc <<'EOF'
export CODEX_TOKEN="你的API Token"
export CODEX_BASE_URL="https://ai.3zapi.top"
export ANTHROPIC_AUTH_TOKEN="你的令牌"
export ANTHROPIC_BASE_URL="https://ai.3zapi.top"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
EOF
source ~/.bashrc`,
  macInstall: `brew install git node
npm install -g @openai/codex @anthropic-ai/claude-code --registry=https://registry.npmmirror.com
git --version
node -v
npm -v`,
  macEnv: `cat >> ~/.zshrc <<'EOF'
export CODEX_TOKEN="你的API Token"
export CODEX_BASE_URL="https://ai.3zapi.top"
export ANTHROPIC_AUTH_TOKEN="你的令牌"
export ANTHROPIC_BASE_URL="https://ai.3zapi.top"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
EOF
source ~/.zshrc`,
  ccSwitchMacInstall: `brew tap farion1231/ccswitch
brew install --cask cc-switch`,
  ccSwitchArchInstall: 'paru -S cc-switch-bin',
  ccSwitchWebInstall: `wget https://github.com/farion1231/cc-switch/releases/latest/download/cc-switch-web-linux-x64.tar.gz
tar -xzf cc-switch-web-linux-x64.tar.gz
cd cc-switch-web/
./cc-switch-web
# open http://localhost:17666`,
  codexStart: 'codex',
  codexCommon: `codex chat
codex run
codex clear
codex update`,
  claudeInstall: 'npm install -g @anthropic-ai/claude-code',
  claudeInstallMirror: 'npm install -g @anthropic-ai/claude-code --registry=https://registry.npmmirror.com',
  claudeVersion: 'claude --version',
  claudeEnv: `$env:ANTHROPIC_AUTH_TOKEN="你的令牌"
$env:ANTHROPIC_BASE_URL="https://ai.3zapi.top"

[Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN","你的令牌","User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL","https://ai.3zapi.top","User")`,
  claudeSettings: `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "你的令牌",
    "ANTHROPIC_BASE_URL": "https://ai.3zapi.top",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}`,
  claudeStart: 'claude',
  openclawSetup: `openclaw onboard
openclaw dashboard`,
  openclawConfig: `{
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
}`,
  openclawCheck: `openclaw models
openclaw dashboard`,
  hermesInstall: 'curl -fsSL https://raw.githubusercontent.com/terryso/hermes-agent/main/install.sh | bash',
  hermesModel: `hermes model
# Provider: Custom Endpoint
# Base URL: https://ai.3zapi.top/v1
# API Key: 你的API Token
# Model: gpt-5.3-codex`,
  hermesStart: 'hermes',
}

const activeSection = ref(sections[0].id)
const copiedCommand = ref('')
const tabLinks = new Map<string, HTMLElement>()
let observer: IntersectionObserver | null = null
let copiedTimer: number | undefined

function handleIndexClick(id: string) {
  activeSection.value = id
}

function setTabLink(id: string, element: Element | ComponentPublicInstance | null) {
  if (element instanceof HTMLElement) {
    tabLinks.set(id, element)
    return
  }

  tabLinks.delete(id)
}

async function copyCommand(id: string, command: string) {
  async function fallbackCopy() {
    if (typeof document === 'undefined') return

    const textarea = document.createElement('textarea')
    textarea.value = command
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  }

  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(command)
    } else {
      await fallbackCopy()
    }
  } catch {
    await fallbackCopy()
  }

  copiedCommand.value = id
  window.clearTimeout(copiedTimer)
  copiedTimer = window.setTimeout(() => {
    if (copiedCommand.value === id) {
      copiedCommand.value = ''
    }
  }, 1600)
}

onMounted(() => {
  if (!('IntersectionObserver' in window)) {
    return
  }

  void nextTick(() => {
    observer = new IntersectionObserver(
      (entries) => {
        const visibleEntry = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0]

        if (visibleEntry?.target.id) {
          activeSection.value = visibleEntry.target.id
        }
      },
      {
        rootMargin: '-26% 0px -56% 0px',
        threshold: [0.01, 0.18, 0.42],
      }
    )

    sections.forEach((section) => {
      const element = document.getElementById(section.id)
      if (element) {
        observer?.observe(element)
      }
    })

  })
})

watch(activeSection, (id) => {
  void nextTick(() => {
    tabLinks.get(id)?.scrollIntoView({
      behavior: 'smooth',
      block: 'nearest',
      inline: 'center',
    })
  })
})

onBeforeUnmount(() => {
  observer?.disconnect()
  window.clearTimeout(copiedTimer)
})
</script>

<style scoped>
@import './public-page.css';

.tutorial-page {
  position: relative;
  overflow: visible;
  background:
    radial-gradient(circle at 50% 0%, rgba(69, 125, 255, 0.1) 0, transparent 24rem),
    linear-gradient(180deg, #eef1f4 0%, #e9ecef 46%, #e5e8eb 100%);
  color: #20242a;
}

.tutorial-page::before {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0 0 auto;
  z-index: 0;
  height: 18rem;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.54), rgba(255, 255, 255, 0));
}

.tutorial-page :deep(.public-matrix-rain) {
  display: block;
  opacity: 0.08;
  mix-blend-mode: multiply;
}

.tutorial-page :deep(.public-blur-field) {
  display: block;
  opacity: 0.22;
  filter: blur(74px);
  mix-blend-mode: multiply;
}

.tutorial-page :deep(.public-noise) {
  display: block;
  opacity: 0.08;
  mix-blend-mode: multiply;
}

.tutorial-page :deep(.public-top-shell) {
  border-bottom-color: rgba(17, 24, 39, 0.1);
  background: rgba(236, 239, 242, 0.86);
  box-shadow: 0 10px 24px rgba(24, 34, 48, 0.08);
}

.tutorial-page :deep(.public-brand),
.tutorial-page :deep(.public-brand span),
.tutorial-page :deep(.public-nav-button) {
  color: #111827;
}

.tutorial-page :deep(.public-nav-pill) {
  color: #46505d;
}

.tutorial-page :deep(.public-nav-pill.router-link-active),
.tutorial-page :deep(.public-nav-pill:hover) {
  color: #0755dd;
}

.tutorial-page :deep(.public-nav-center),
.tutorial-page :deep(.public-icon-button),
.tutorial-page :deep(.public-nav-button),
.tutorial-page :deep(.public-brand-logo) {
  border-color: rgba(15, 23, 42, 0.12);
  background: rgba(255, 255, 255, 0.54);
  box-shadow: 0 8px 20px rgba(24, 34, 48, 0.08);
}

.doc-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  align-items: center;
}

.doc-pills span,
.doc-pills strong {
  display: inline-flex;
  min-height: 2.15rem;
  align-items: center;
  border-radius: 999px;
  padding: 0.35rem 1rem;
  font-size: 0.9rem;
  font-weight: 850;
}

.doc-pills span {
  background: rgba(17, 24, 39, 0.07);
  color: #343a42;
}

.doc-pills strong {
  background: linear-gradient(135deg, #1e75ff, #1748d4);
  color: white;
  box-shadow: 0 8px 20px rgba(23, 72, 212, 0.22);
}

.beginner-path {
  margin-top: 1.2rem;
  border: 1px solid rgba(30, 117, 255, 0.14);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.46);
  padding: 1rem;
  box-shadow: 0 10px 22px rgba(24, 34, 48, 0.06);
}

.beginner-path-head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.75rem;
  align-items: baseline;
  margin-bottom: 0.85rem;
}

.beginner-path-head span {
  color: #0755dd;
  font-size: 0.76rem;
  font-weight: 950;
  letter-spacing: 0.08em;
}

.beginner-path-head strong {
  color: #151922;
  font-size: 1rem;
  font-weight: 950;
}

.beginner-path-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.7rem;
}

.beginner-step {
  display: grid;
  gap: 0.45rem;
  align-content: start;
  min-height: 8.2rem;
  border: 1px solid rgba(17, 24, 39, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.5);
  padding: 0.75rem;
  color: inherit;
  transition: border-color 160ms ease, background 160ms ease, transform 160ms ease;
}

.beginner-step:hover,
.beginner-step:focus-visible {
  border-color: rgba(30, 117, 255, 0.38);
  background: rgba(255, 255, 255, 0.7);
  transform: translateY(-1px);
}

.beginner-step:focus-visible {
  outline: 2px solid rgba(30, 117, 255, 0.35);
  outline-offset: 3px;
}

.beginner-step span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.85rem;
  height: 1.85rem;
  border-radius: 6px;
  background: linear-gradient(135deg, #1e75ff, #14b8a6);
  color: white;
  font-size: 0.78rem;
  font-weight: 950;
}

.beginner-step strong {
  color: #151922;
  font-size: 0.92rem;
  font-weight: 950;
}

.beginner-step em {
  color: #626a76;
  font-size: 0.82rem;
  font-style: normal;
  line-height: 1.55;
}

.tutorial-main {
  max-width: min(128rem, calc(100vw - 1rem));
}

.tutorial-overview {
  position: relative;
  display: grid;
  gap: 1.5rem;
  align-items: start;
  border-bottom: 1px solid rgba(17, 24, 39, 0.12);
  padding: 0 0 1.4rem;
}

.tutorial-overview::after {
  content: none;
}

.tutorial-intro,
.route-map {
  position: relative;
  z-index: 1;
}

.tutorial-intro h1 {
  margin-top: 1rem;
  max-width: 42rem;
  color: #151922;
  font-size: clamp(2.5rem, 5.8vw, 4rem);
  font-weight: 950;
  line-height: 1.03;
  letter-spacing: 0;
}

.tutorial-intro > p {
  margin-top: 1rem;
  max-width: 50rem;
  color: #5f6673;
  font-size: 1.02rem;
  line-height: 1.8;
}

.overview-checklist {
  display: none;
}

.overview-row {
  display: grid;
  grid-template-columns: 2.3rem minmax(10rem, 13rem) minmax(0, 1fr);
  gap: 1rem;
  align-items: center;
  min-height: 4.2rem;
  border-bottom: 1px solid rgba(17, 24, 39, 0.12);
  color: inherit;
}

.overview-row:focus-visible,
.tutorial-tabs a:focus-visible,
.tutorial-link-list a:focus-visible,
.copy-command-button:focus-visible {
  outline: 2px solid rgba(119, 255, 173, 0.86);
  outline-offset: 3px;
}

.overview-row span,
.route-step > span,
.guide-card-head span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 1.8rem;
  min-width: 1.8rem;
  border-radius: 6px;
  background: linear-gradient(135deg, #0f9f90, #5638d7);
  color: white;
  font-size: 0.75rem;
  font-weight: 950;
}

.overview-row strong {
  color: #151922;
  font-size: 0.94rem;
  font-weight: 950;
}

.overview-row em {
  color: #6f7682;
  font-size: 0.86rem;
  font-style: normal;
  line-height: 1.6;
}

.route-step em {
  color: #6f7682;
  font-size: 0.86rem;
  font-style: normal;
  line-height: 1.6;
}

.route-map {
  display: none;
}

.route-map p {
  margin: 0 0 0.35rem;
  color: #2dd4bf;
  font-size: 0.78rem;
  font-weight: 950;
  letter-spacing: 0.08em;
}

.route-step {
  display: grid;
  grid-template-columns: 2.2rem minmax(0, 1fr);
  gap: 0.85rem;
  align-items: center;
  min-height: 5.55rem;
  border: 1px solid rgba(226, 232, 240, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.055);
  padding: 0.9rem;
}

.route-step strong {
  display: block;
  margin-bottom: 0.25rem;
  color: rgba(255, 255, 255, 0.96);
  font-size: 0.92rem;
  font-weight: 950;
}

.tutorial-reader {
  --tutorial-index-sticky-top: 4.15rem;

  display: grid;
  grid-template-columns: minmax(13rem, 16rem) minmax(0, 72rem);
  gap: 2rem;
  align-items: start;
}

.tutorial-sidebar {
  position: sticky;
  top: var(--tutorial-index-sticky-top);
  z-index: 28;
  border-right: 1px solid rgba(17, 24, 39, 0.12);
  padding: 4.8rem 1rem 1rem 0.5rem;
}

.tutorial-main-column {
  display: grid;
  gap: 1.2rem;
  min-width: 0;
  padding-top: 1.2rem;
}

.tutorial-sidebar-title {
  margin: 0 0 0.58rem;
  color: #1f2937;
  font-size: 0.95rem;
  font-weight: 850;
  letter-spacing: 0;
}

.tutorial-tabs {
  display: grid;
  gap: 0.28rem;
}

.tutorial-tabs a {
  display: grid;
  gap: 0.18rem;
  min-height: 2.75rem;
  align-content: center;
  border-left: 3px solid transparent;
  border-radius: 6px;
  padding: 0.42rem 0.7rem;
  color: #3f4652;
  transition: background 160ms ease, border-color 160ms ease, color 160ms ease;
}

.tutorial-tabs a:hover,
.tutorial-tabs a.is-active {
  border-left-color: #1e75ff;
  background: rgba(30, 117, 255, 0.1);
  color: #0755dd;
}

.tutorial-tabs strong {
  font-size: 0.9rem;
  font-weight: 850;
}

.tutorial-tabs span {
  color: #747b86;
  font-size: 0.78rem;
  line-height: 1.35;
}

.tutorial-tabs a.is-active span,
.tutorial-tabs a:hover span {
  color: #3869b9;
}

.tutorial-content {
  display: grid;
  gap: 1.4rem;
  min-width: 0;
}

.guide-panel {
  scroll-margin-top: calc(var(--tutorial-index-sticky-top) + 1rem);
  border-bottom: 1px solid rgba(17, 24, 39, 0.1);
  padding: 1.2rem 0 1.8rem;
}

.guide-heading {
  max-width: 46rem;
}

.guide-heading span {
  color: #0755dd;
  font-size: 0.76rem;
  font-weight: 950;
  letter-spacing: 0.08em;
}

.guide-heading h2 {
  margin-top: 0.28rem;
  color: #151922;
  font-size: clamp(1.65rem, 3vw, 2.35rem);
  font-weight: 950;
}

.guide-heading p {
  margin-top: 0.45rem;
  color: #626a76;
  line-height: 1.75;
}

.guide-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
  margin-top: 1rem;
}

.guide-grid--single {
  grid-template-columns: 1fr;
}

.guide-card {
  position: relative;
  display: grid;
  align-content: start;
  gap: 0.65rem;
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgba(17, 24, 39, 0.12);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.36);
  padding: 1rem;
  box-shadow: 0 10px 22px rgba(24, 34, 48, 0.06);
}

.guide-card::before {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: linear-gradient(180deg, #1e75ff, #14b8a6);
  opacity: 0.58;
}

.guide-card--wide {
  grid-column: span 2;
}

.guide-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
}

.guide-card h3 {
  color: #151922;
  font-size: 1.05rem;
  font-weight: 950;
}

.guide-card h4 {
  color: #151922;
  font-size: 1rem;
  font-weight: 900;
}

.guide-card p,
.guide-card li,
.faq-card p {
  color: #626a76;
  line-height: 1.72;
}

.guide-card ul {
  margin: 0;
  padding-left: 1.15rem;
}

.guide-card code,
.faq-card code {
  color: #0755dd;
  font-size: 0.9em;
}

.doc-feature-stack {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
  margin-top: 1rem;
}

.doc-feature,
.doc-subsection {
  scroll-margin-top: calc(var(--tutorial-index-sticky-top) + 1rem);
}

.doc-feature,
.doc-subsection:not(#cc-switch-install) {
  position: relative;
  display: grid;
  align-content: start;
  gap: 0.65rem;
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgba(17, 24, 39, 0.12);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.36);
  padding: 1rem;
  box-shadow: 0 10px 22px rgba(24, 34, 48, 0.06);
}

.doc-feature::before,
.doc-subsection:not(#cc-switch-install)::before {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: linear-gradient(180deg, #1e75ff, #14b8a6);
  opacity: 0.58;
}

.doc-feature h3,
.doc-subsection h3 {
  color: #151922;
  font-size: 1.05rem;
  font-weight: 950;
}

.doc-feature ul,
.doc-subsection ul,
.doc-steps {
  margin: 0;
  padding-left: 1.28rem;
}

.doc-feature li,
.doc-subsection li,
.doc-subsection p,
.doc-steps li {
  color: #3f4652;
  line-height: 1.9;
}

.doc-subsection {
  margin-top: 1rem;
}

.doc-subsection p {
  margin-top: 0;
}

.install-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
  margin-top: 1rem;
}

.command-block {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(119, 255, 173, 0.18);
  border-radius: 8px;
  background: #101827;
}

.command-block::before {
  content: '';
  display: block;
  height: 0.42rem;
  background: linear-gradient(90deg, #14b8a6, #5638d7);
  opacity: 0.62;
}

.copy-command-button {
  position: absolute;
  top: 0.72rem;
  right: 0.62rem;
  z-index: 2;
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 6px;
  background: rgba(15, 23, 42, 0.82);
  padding: 0.28rem 0.52rem;
  color: rgba(220, 252, 231, 0.92);
  font-size: 0.72rem;
  font-weight: 900;
  transition: border-color 160ms ease, background 160ms ease, color 160ms ease;
}

.copy-command-button:hover {
  border-color: rgba(45, 212, 191, 0.52);
  background: rgba(15, 118, 110, 0.82);
  color: white;
}

.guide-card pre {
  max-width: 100%;
  margin: 0;
  overflow-x: auto;
  border: 0;
  background: transparent;
  padding: 0.82rem 4.2rem 0.82rem 0.82rem;
  color: #f8fafc;
  font-size: 0.78rem;
  line-height: 1.65;
}

.command-block code {
  color: #f8fafc;
}

.tutorial-link-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.45rem;
  padding-left: 0 !important;
  list-style: none;
}

.tutorial-link-list a {
  display: block;
  border: 1px solid rgba(30, 117, 255, 0.16);
  border-radius: 6px;
  background: rgba(30, 117, 255, 0.08);
  padding: 0.55rem 0.65rem;
  color: #0755dd;
  font-size: 0.86rem;
  font-weight: 900;
}

.tutorial-link-list a:hover {
  border-color: rgba(30, 117, 255, 0.42);
  color: #003fbd;
}

.guide-action-link {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  gap: 0.42rem;
  margin-top: 1rem;
  border: 1px solid rgba(119, 255, 173, 0.34);
  border-radius: 6px;
  background:
    linear-gradient(180deg, rgba(30, 117, 255, 0.18), rgba(20, 184, 166, 0.08)),
    rgba(255, 255, 255, 0.68);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    0 10px 22px rgba(0, 0, 0, 0.18);
  padding: 0.58rem 0.75rem;
  color: #0755dd;
  font-size: 0.84rem;
  font-weight: 950;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.guide-action-link:hover {
  border-color: rgba(30, 117, 255, 0.52);
  background:
    linear-gradient(180deg, rgba(30, 117, 255, 0.2), rgba(20, 184, 166, 0.1)),
    rgba(255, 255, 255, 0.84);
  color: #003fbd;
  transform: translateY(-1px);
}

.guide-action-link .pixel-glyph {
  --pixel-glyph-on: rgba(217, 255, 230, 0.95);
  --pixel-glyph-accent: rgba(119, 255, 173, 0.78);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.faq-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.8rem;
  margin-top: 1rem;
}

.faq-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(17, 24, 39, 0.12);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.36);
  padding: 1rem;
  box-shadow: var(--public-shadow-soft);
}

.faq-card::before {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: linear-gradient(180deg, #1e75ff, #14b8a6);
  opacity: 0.5;
}

.faq-card strong {
  color: #151922;
  font-size: 0.98rem;
  font-weight: 950;
}

.faq-card p {
  margin-top: 0.45rem;
}

@media (max-width: 1180px) {
  .tutorial-main {
    max-width: 100%;
  }

  .tutorial-reader {
    grid-template-columns: minmax(12rem, 14rem) minmax(0, 1fr);
  }
}

@media (max-width: 980px) {
  .tutorial-overview {
    grid-template-columns: 1fr;
  }

  .beginner-path-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .tutorial-reader {
    --tutorial-index-sticky-top: 6.55rem;

    display: block;
  }

  .tutorial-sidebar {
    top: var(--tutorial-index-sticky-top);
    margin: 1.15rem 0.5rem 0;
    padding: 0.35rem;
    border-right: 0;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.62);
    box-shadow: 0 10px 22px rgba(24, 34, 48, 0.08);
    backdrop-filter: blur(18px);
  }

  .tutorial-sidebar-title {
    display: none;
  }

  .tutorial-tabs {
    overflow-x: auto;
    grid-template-columns: repeat(10, minmax(8.5rem, 1fr));
  }

  .tutorial-content {
    margin-top: 1.4rem;
  }

  .tutorial-main-column {
    padding-top: 1rem;
  }

  .faq-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .doc-feature-stack {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .tutorial-page {
    overflow: hidden;
  }

  .tutorial-overview {
    padding: 0.95rem;
  }

  .tutorial-intro h1 {
    margin-top: 0.72rem;
    font-size: clamp(2.15rem, 12vw, 3.05rem);
  }

  .tutorial-intro > p {
    margin-top: 0.8rem;
    font-size: 0.96rem;
    line-height: 1.65;
  }

  .overview-checklist {
    margin-top: 1rem;
  }

  .overview-row {
    grid-template-columns: 2rem minmax(0, 1fr);
    gap: 0.7rem;
    min-height: 3.55rem;
    padding: 0.58rem 0;
  }

  .overview-row em {
    grid-column: 2;
    font-size: 0.82rem;
    line-height: 1.45;
  }

  .route-step {
    min-height: auto;
  }

  .route-map {
    display: none;
  }

  .tutorial-reader {
    --tutorial-index-sticky-top: 6.35rem;
  }

  .tutorial-sidebar {
    margin-inline: 0.25rem;
    padding: 0.28rem;
  }

  .tutorial-tabs {
    grid-template-columns: repeat(10, minmax(8rem, 1fr));
  }

  .tutorial-tabs a {
    min-height: 3rem;
    padding: 0.44rem 0.58rem;
  }

  .guide-grid,
  .faq-grid,
  .doc-feature-stack,
  .install-grid,
  .beginner-path-grid {
    grid-template-columns: 1fr;
  }

  .guide-card--wide {
    grid-column: auto;
  }

  .tutorial-link-list {
    grid-template-columns: 1fr;
  }

  .copy-command-button {
    position: static;
    float: right;
    margin: 0.55rem 0.55rem 0 0;
  }

  .guide-card pre {
    clear: both;
    padding-right: 0.82rem;
  }
}
</style>
