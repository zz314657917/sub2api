<template>
  <div class="tutorial-page public-page-shell min-h-screen text-white">
    <PublicMatrixBackdrop />

    <PublicTopNav />

    <main class="relative z-10 mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:py-10">
      <section id="quick-start" class="tutorial-overview">
        <div class="tutorial-intro">
          <div class="public-kicker">
            <span></span>
            QUICK START
          </div>
          <h1>AI 接入教程</h1>
          <p>
            先准备 Token 和代理地址。账号管理工具可选，实际使用入口是 Codex 或 Claude Code。
          </p>

          <div class="overview-checklist" aria-label="接入步骤概览">
            <a href="#prepare" class="overview-row">
              <span>01</span>
              <strong>准备环境</strong>
              <em>安装 Git、Node.js，准备 Token 和代理地址。</em>
            </a>
            <a href="#cc-switch" class="overview-row">
              <span>02</span>
              <strong>CC Switch 账号管理</strong>
              <em>桌面端统一管理 Codex 和 Claude Code 账号。</em>
            </a>
            <a href="#cockpit-tools" class="overview-row">
              <span>03</span>
              <strong>Cockpit Tools 账号管理</strong>
              <em>网页面板管理 Codex 账号和会话。</em>
            </a>
            <a href="#codex" class="overview-row">
              <span>04</span>
              <strong>配置 Codex</strong>
              <em>推荐一键配置，失败再手动设置。</em>
            </a>
            <a href="#claude" class="overview-row">
              <span>05</span>
              <strong>配置 Claude Code</strong>
              <em>写入 ANTHROPIC 变量，重启终端后启动。</em>
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
                <li>Windows 电脑；Claude Code 建议 Windows 10 1809+。</li>
                <li>Node.js v18 或更高；教程链接为 v24。</li>
                <li>落叶网络 API Token 和 Base URL。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>02</span>
                <PixelIcon name="folder" size="sm" tone="green" />
              </div>
              <h3>安装 Git / Node.js</h3>
              <p>安装时一路下一步，路径保持默认。</p>
              <ul class="tutorial-link-list">
                <li>
                  <a href="https://github.com/git-for-windows/git/releases/download/v2.53.0.windows.1/Git-2.53.0-64-bit.exe" target="_blank" rel="noopener noreferrer">
                    Git 官方下载
                  </a>
                </li>
                <li>
                  <a href="https://registry.npmmirror.com/-/binary/git-for-windows/v2.51.0.windows.1/Git-2.51.0-64-bit.exe" target="_blank" rel="noopener noreferrer">
                    Git 镜像下载
                  </a>
                </li>
                <li>
                  <a href="https://nodejs.org/dist/v24.13.1/node-v24.13.1-x64.msi" target="_blank" rel="noopener noreferrer">
                    Node 官方下载
                  </a>
                </li>
                <li>
                  <a href="https://npmmirror.com/mirrors/node/v24.13.0/node-v24.13.0-x64.msi" target="_blank" rel="noopener noreferrer">
                    Node 镜像下载
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

        <section id="cc-switch" class="guide-panel">
          <div class="guide-heading">
            <span>CC Switch</span>
            <h2>CC Switch 账号管理</h2>
            <p>用于统一管理 Codex、Claude Code 的账号配置。已经会手动配置的用户可以跳过。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>S1</span>
                <PixelIcon name="panel" size="sm" tone="green" />
              </div>
              <h3>下载 CC Switch</h3>
              <p>Windows 用户下载最新版 <code>.msi</code> 安装包，或下载 Portable 免安装版。</p>
              <ul class="tutorial-link-list">
                <li>
                  <a href="https://ccswitch.ai/" target="_blank" rel="noopener noreferrer">
                    官方介绍
                  </a>
                </li>
                <li>
                  <a href="https://github.com/farion1231/cc-switch/releases" target="_blank" rel="noopener noreferrer">
                    Windows 下载
                  </a>
                </li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>S2</span>
                <PixelIcon name="key" size="sm" tone="green" />
              </div>
              <h3>添加账号配置</h3>
              <ul>
                <li>打开 CC Switch，点击 <code>Add Provider</code>。</li>
                <li>选择要管理的工具：<code>Codex</code> 或 <code>Claude Code</code>。</li>
                <li>填入 API Token 和代理地址。</li>
                <li>保存后点击 <code>Enable</code> 启用。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>S3</span>
                <PixelIcon name="cursor" size="sm" tone="green" />
              </div>
              <h3>验证账号是否生效</h3>
              <ul>
                <li>启用账号配置后，重新打开对应终端或工具。</li>
                <li>启动 <code>codex</code> 或 <code>claude</code>。</li>
                <li>能正常对话，就说明账号配置成功。</li>
              </ul>
            </article>
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
            <p>先用镜像源一键安装，再写入 Token 和代理地址。</p>
          </div>

          <div class="guide-grid">
            <article class="guide-card guide-card--wide">
              <div class="guide-card-head">
                <span>C1</span>
                <PixelIcon name="spark" size="sm" tone="green" />
              </div>
              <h3>推荐：一键配置</h3>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex 镜像安装命令"
                  @click="copyCommand('codex-install-mirror', commands.codexInstallMirror)"
                >
                  {{ copiedCommand === 'codex-install-mirror' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexInstallMirror }}</code></pre>
              </div>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex 一键配置命令"
                  @click="copyCommand('codex-config', commands.codexConfig)"
                >
                  {{ copiedCommand === 'codex-config' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexConfig }}</code></pre>
              </div>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex 检查命令"
                  @click="copyCommand('codex-check', commands.codexCheck)"
                >
                  {{ copiedCommand === 'codex-check' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexCheck }}</code></pre>
              </div>
              <p>看到“配置正常”就可以启动 Codex。</p>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>C2</span>
                <PixelIcon name="settings" size="sm" tone="green" />
              </div>
              <h3>备用：手动配置</h3>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex 常规安装命令"
                  @click="copyCommand('codex-install', commands.codexInstall)"
                >
                  {{ copiedCommand === 'codex-install' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexInstall }}</code></pre>
              </div>
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex 镜像安装命令"
                  @click="copyCommand('codex-install-mirror-manual', commands.codexInstallMirror)"
                >
                  {{ copiedCommand === 'codex-install-mirror-manual' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexInstallMirror }}</code></pre>
              </div>
              <ul>
                <li><code>Win + R</code> 输入 <code>sysdm.cpl</code>。</li>
                <li>打开“高级” → “环境变量”。</li>
                <li>新增 <code>CODEX_TOKEN</code> 和 <code>CODEX_BASE_URL</code>。</li>
                <li>保存后重新打开终端。</li>
              </ul>
            </article>

            <article class="guide-card">
              <div class="guide-card-head">
                <span>C3</span>
                <PixelIcon name="cursor" size="sm" tone="green" />
              </div>
              <h3>启动与常用命令</h3>
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
              <div class="command-block">
                <button
                  type="button"
                  class="copy-command-button"
                  aria-label="复制 Codex 常用命令"
                  @click="copyCommand('codex-common', commands.codexCommon)"
                >
                  {{ copiedCommand === 'codex-common' ? '已复制' : '复制' }}
                </button>
                <pre><code>{{ commands.codexCommon }}</code></pre>
              </div>
              <p>VS Code / Cursor 安装 Codex 插件后，选择本地 Codex 服务即可。</p>
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
              <h3>备用：手动配置</h3>
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
  { id: 'prepare', title: '准备工作', desc: '安装必备软件' },
  { id: 'cc-switch', title: 'CC Switch', desc: '桌面账号管理' },
  { id: 'cockpit-tools', title: 'Cockpit Tools', desc: '网页账号管理' },
  { id: 'codex', title: 'Codex', desc: '命令行使用' },
  { id: 'claude', title: 'Claude', desc: '命令行使用' },
  { id: 'faq', title: '常见问题', desc: '快速排查' },
]

const routeSteps = [
  { id: 'prepare', step: '01', title: '装好基础软件', desc: 'Git、Node.js、Token 和代理地址准备好。' },
  { id: 'cc-switch', step: '02', title: '可选：CC Switch', desc: '桌面端管理多个工具账号。' },
  { id: 'cockpit-tools', step: '03', title: '可选：Cockpit Tools', desc: '网页面板管理 Codex 账号和会话。' },
  { id: 'codex', step: '04', title: '使用 Codex', desc: '一键配置优先，失败再手动设置。' },
  { id: 'claude', step: '05', title: '使用 Claude', desc: '写入 ANTHROPIC 变量后重启终端。' },
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
  codexInstall: 'npm install -g @openai/codex',
  codexInstallMirror: 'npm install -g @openai/codex --registry=https://registry.npmmirror.com',
  codexConfig: `codex config set token "你的API Token"
codex config set base_url "你的代理地址"`,
  codexCheck: 'codex check',
  cockpitBaseUrl: 'https://ai.3zapi.top/',
  codexStart: 'codex',
  codexCommon: `codex chat
codex run
codex clear
codex update`,
  claudeInstall: 'npm install -g @anthropic-ai/claude-code',
  claudeInstallMirror: 'npm install -g @anthropic-ai/claude-code --registry=https://registry.npmmirror.com',
  claudeVersion: 'claude --version',
  claudeEnv: `$env:ANTHROPIC_AUTH_TOKEN="你的令牌"
$env:ANTHROPIC_BASE_URL="你的接口地址"

[Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN","你的令牌","User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL","你的接口地址","User")`,
  claudeSettings: `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "你的令牌",
    "ANTHROPIC_BASE_URL": "你的接口地址",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}`,
  claudeStart: 'claude',
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
    radial-gradient(circle at 50% 18%, rgba(32, 170, 92, 0.22) 0, transparent 34%),
    radial-gradient(circle at 18% 24%, rgba(87, 86, 210, 0.16) 0, transparent 30%),
    radial-gradient(circle at 82% 36%, rgba(45, 178, 105, 0.14) 0, transparent 32%),
    linear-gradient(180deg, #050914 0%, #08110f 48%, #03060a 100%);
  color: rgba(255, 255, 255, 0.94);
}

.tutorial-page::before {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0 0 auto;
  z-index: 0;
  height: 18rem;
  background: linear-gradient(180deg, rgba(5, 7, 18, 0.76), rgba(5, 7, 18, 0));
}

.tutorial-page :deep(.public-matrix-rain) {
  display: block;
  opacity: 0.54;
  mix-blend-mode: screen;
}

.tutorial-page :deep(.public-blur-field) {
  display: block;
  opacity: 0.64;
  filter: blur(68px);
  mix-blend-mode: screen;
}

.tutorial-page :deep(.public-noise) {
  display: block;
  opacity: 0.2;
  mix-blend-mode: screen;
}

.tutorial-page :deep(.public-top-shell) {
  box-shadow: 0 12px 34px rgba(35, 43, 52, 0.14);
}

.tutorial-page .public-kicker {
  border-color: rgba(221, 230, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
  color: rgba(209, 255, 224, 0.9);
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.16);
  backdrop-filter: blur(18px);
}

.tutorial-overview {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(20rem, 23rem);
  gap: 1.6rem;
  align-items: stretch;
  overflow: hidden;
  border: 1px solid rgba(221, 230, 255, 0.14);
  border-radius: 8px;
  background:
    radial-gradient(circle at 76% 24%, rgba(119, 255, 173, 0.12), transparent 34%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.105), rgba(255, 255, 255, 0.058)),
    rgba(6, 13, 18, 0.56);
  padding: clamp(1.2rem, 3vw, 2rem);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 18px 44px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(22px);
}

.tutorial-overview::after {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(119, 255, 173, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(167, 139, 250, 0.04) 1px, transparent 1px);
  background-size: 3rem 3rem;
  opacity: 0.55;
}

.tutorial-intro,
.route-map {
  position: relative;
  z-index: 1;
}

.tutorial-intro h1 {
  margin-top: 1rem;
  max-width: 48rem;
  color: rgba(255, 255, 255, 0.98);
  font-size: clamp(2.6rem, 6.2vw, 4.75rem);
  font-weight: 950;
  line-height: 0.98;
  letter-spacing: 0;
}

.tutorial-intro > p {
  margin-top: 1rem;
  max-width: 48rem;
  color: rgba(222, 232, 255, 0.68);
  font-size: 1.02rem;
  line-height: 1.8;
}

.overview-checklist {
  margin-top: 1.3rem;
  border-top: 1px solid rgba(221, 230, 255, 0.12);
}

.overview-row {
  display: grid;
  grid-template-columns: 2.3rem minmax(10rem, 13rem) minmax(0, 1fr);
  gap: 1rem;
  align-items: center;
  min-height: 4.2rem;
  border-bottom: 1px solid rgba(221, 230, 255, 0.12);
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
  color: rgba(255, 255, 255, 0.94);
  font-size: 0.94rem;
  font-weight: 950;
}

.overview-row em {
  color: rgba(222, 232, 255, 0.6);
  font-size: 0.86rem;
  font-style: normal;
  line-height: 1.6;
}

.route-step em {
  color: rgba(220, 232, 244, 0.76);
  font-size: 0.86rem;
  font-style: normal;
  line-height: 1.6;
}

.route-map {
  display: grid;
  align-content: center;
  gap: 0.75rem;
  overflow: hidden;
  border: 1px solid rgba(45, 212, 191, 0.16);
  border-radius: 8px;
  background:
    linear-gradient(90deg, rgba(45, 212, 191, 0.08) 1px, transparent 1px),
    linear-gradient(180deg, rgba(45, 212, 191, 0.06) 1px, transparent 1px),
    linear-gradient(135deg, #112434, #111831 72%);
  background-size: 3rem 3rem, 3rem 3rem, auto;
  padding: 1.25rem;
  box-shadow: 0 24px 48px rgba(20, 27, 40, 0.22);
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

.tutorial-tabs {
  --tutorial-tabs-gutter: clamp(1rem, 2vw, 1.75rem);
  --tutorial-tabs-sticky-top: 4.45rem;

  position: sticky;
  top: var(--tutorial-tabs-sticky-top);
  z-index: 30;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 0.25rem;
  margin: 1.15rem var(--tutorial-tabs-gutter) 0;
  border: 1px solid rgba(221, 230, 255, 0.14);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.055)),
    rgba(6, 13, 18, 0.58);
  padding: 0.35rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 14px 32px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(20px);
}

.tutorial-tabs a {
  display: grid;
  gap: 0.18rem;
  min-height: 3.6rem;
  align-content: center;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 0.58rem 0.75rem;
  color: rgba(222, 232, 255, 0.62);
  transition: background 160ms ease, border-color 160ms ease, color 160ms ease;
}

.tutorial-tabs a:hover,
.tutorial-tabs a.is-active {
  border-color: rgba(15, 118, 110, 0.18);
  background: rgba(119, 255, 173, 0.12);
  color: white;
}

.tutorial-tabs strong {
  font-size: 0.9rem;
  font-weight: 950;
}

.tutorial-tabs span {
  color: rgba(222, 232, 255, 0.5);
  font-size: 0.78rem;
  line-height: 1.35;
}

.tutorial-tabs a.is-active span,
.tutorial-tabs a:hover span {
  color: rgba(229, 245, 241, 0.78);
}

.tutorial-content {
  display: grid;
  gap: 1.4rem;
  margin-top: 1.4rem;
}

.guide-panel {
  scroll-margin-top: calc(var(--tutorial-tabs-sticky-top) + 5rem);
  border-top: 3px solid rgba(20, 184, 166, 0.72);
  border-radius: 8px;
  border-right: 1px solid rgba(221, 230, 255, 0.14);
  border-bottom: 1px solid rgba(221, 230, 255, 0.14);
  border-left: 1px solid rgba(221, 230, 255, 0.14);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.105), rgba(255, 255, 255, 0.058)),
    rgba(6, 13, 18, 0.54);
  padding: clamp(1rem, 2.5vw, 1.5rem);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 16px 34px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(20px);
}

.guide-heading {
  max-width: 46rem;
}

.guide-heading span {
  color: #77ffad;
  font-size: 0.76rem;
  font-weight: 950;
  letter-spacing: 0.08em;
}

.guide-heading h2 {
  margin-top: 0.28rem;
  color: rgba(255, 255, 255, 0.96);
  font-size: clamp(1.55rem, 3vw, 2.15rem);
  font-weight: 950;
}

.guide-heading p {
  margin-top: 0.45rem;
  color: rgba(222, 232, 255, 0.66);
  line-height: 1.75;
}

.guide-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
  margin-top: 1rem;
}

.guide-card {
  position: relative;
  display: grid;
  align-content: start;
  gap: 0.65rem;
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgba(221, 230, 255, 0.14);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.095), rgba(255, 255, 255, 0.052)),
    rgba(6, 13, 18, 0.5);
  padding: 1rem;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.18);
}

.guide-card::before {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: linear-gradient(180deg, #14b8a6, #5638d7);
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
  color: rgba(255, 255, 255, 0.95);
  font-size: 1.05rem;
  font-weight: 950;
}

.guide-card p,
.guide-card li,
.faq-card p {
  color: rgba(222, 232, 255, 0.66);
  line-height: 1.72;
}

.guide-card ul {
  margin: 0;
  padding-left: 1.15rem;
}

.guide-card code,
.faq-card code {
  color: #77ffad;
  font-size: 0.9em;
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
  border: 1px solid rgba(15, 118, 110, 0.14);
  border-radius: 6px;
  background: rgba(20, 184, 166, 0.08);
  padding: 0.55rem 0.65rem;
  color: #baf7cb;
  font-size: 0.86rem;
  font-weight: 900;
}

.tutorial-link-list a:hover {
  border-color: rgba(119, 255, 173, 0.42);
  color: white;
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
    linear-gradient(180deg, rgba(119, 255, 173, 0.2), rgba(20, 184, 166, 0.08)),
    rgba(5, 15, 18, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    0 10px 22px rgba(0, 0, 0, 0.18);
  padding: 0.58rem 0.75rem;
  color: #d9ffe6;
  font-size: 0.84rem;
  font-weight: 950;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.guide-action-link:hover {
  border-color: rgba(119, 255, 173, 0.62);
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.28), rgba(20, 184, 166, 0.14)),
    rgba(6, 28, 24, 0.86);
  color: white;
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
  border: 1px solid rgba(221, 230, 255, 0.14);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.095), rgba(255, 255, 255, 0.052)),
    rgba(6, 13, 18, 0.5);
  padding: 1rem;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.18);
}

.faq-card::before {
  content: '';
  pointer-events: none;
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: linear-gradient(180deg, #14b8a6, #5638d7);
  opacity: 0.5;
}

.faq-card strong {
  color: rgba(255, 255, 255, 0.95);
  font-size: 0.98rem;
  font-weight: 950;
}

.faq-card p {
  margin-top: 0.45rem;
}

@media (max-width: 980px) {
  .tutorial-overview {
    grid-template-columns: 1fr;
  }

  .route-map {
    align-content: start;
  }

  .tutorial-tabs {
    --tutorial-tabs-gutter: 0.5rem;
    --tutorial-tabs-sticky-top: 6.55rem;

    overflow-x: auto;
    grid-template-columns: repeat(6, minmax(8.5rem, 1fr));
  }

  .faq-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
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

  .tutorial-tabs {
    --tutorial-tabs-gutter: 0.25rem;
    --tutorial-tabs-sticky-top: 6.35rem;

    margin-inline: var(--tutorial-tabs-gutter);
    grid-template-columns: repeat(6, minmax(8rem, 1fr));
    padding: 0.28rem;
  }

  .tutorial-tabs a {
    min-height: 3rem;
    padding: 0.44rem 0.58rem;
  }

  .guide-grid,
  .faq-grid {
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
