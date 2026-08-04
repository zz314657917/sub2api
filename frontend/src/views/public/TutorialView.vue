<template>
  <div class="tutorial-page relative min-h-screen">
    <PublicRevealBackdrop variant="page" />
    <PublicTopNav />

    <main
      class="tutorial-main relative z-10 mx-auto"
      :class="{ 'is-article-route': !isIndexRoute, 'is-index-route': isIndexRoute }"
    >
      <section v-if="showQuickstart" class="tutorial-quickstart" aria-labelledby="tutorial-quickstart-title">
        <div class="tutorial-quickstart-head">
          <div>
            <span class="tutorial-kicker">{{ quickstartConfig.header.kicker }}</span>
            <h1 id="tutorial-quickstart-title">{{ quickstartConfig.header.title }}</h1>
            <p>{{ quickstartConfig.header.description }}</p>
          </div>
          <div class="tutorial-actions">
            <router-link :to="{ path: '/tutorial', query: { view: 'library' } }" class="guide-action-link">
              {{ quickstartConfig.header.library_action_label }}
            </router-link>
            <router-link to="/keys" class="guide-action-link guide-action-link--ghost">
              {{ quickstartConfig.header.keys_action_label }}
            </router-link>
          </div>
        </div>

        <div class="tutorial-quickstart-controls">
          <div class="tutorial-quickstart-control">
            <span>{{ quickstartConfig.header.platform_control_label }}</span>
            <div class="tutorial-segmented-control" role="group" aria-label="选择模型平台">
              <button
                v-for="option in quickstartPlatforms"
                :key="option.id"
                type="button"
                :class="{ 'is-active': quickstartPlatform === option.id }"
                :aria-pressed="quickstartPlatform === option.id"
                @click="quickstartPlatform = option.id"
              >
                {{ option.label }}
              </button>
            </div>
          </div>

          <div class="tutorial-quickstart-control">
            <span>{{ quickstartConfig.header.terminal_control_label }}</span>
            <div class="tutorial-segmented-control" role="group" aria-label="选择系统和终端">
              <button
                v-for="option in quickstartTerminals"
                :key="option.value"
                type="button"
                :class="{ 'is-active': quickstartTerminal === option.value }"
                :aria-pressed="quickstartTerminal === option.value"
                @click="quickstartTerminal = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>

        <div class="tutorial-quickstart-facts">
          <article v-for="fact in quickstartFacts" :key="fact.label" class="tutorial-quickstart-fact">
            <span>{{ fact.label }}</span>
            <strong>{{ fact.value }}</strong>
            <p>{{ fact.description }}</p>
          </article>
        </div>

        <div class="tutorial-quickstart-steps">
          <article v-for="step in quickstartSteps" :key="step.number" class="tutorial-quickstart-step">
            <header>
              <span class="tutorial-quickstart-step-number">{{ step.number }}</span>
              <div>
                <span v-if="step.kicker" class="tutorial-quickstart-step-kicker">{{ step.kicker }}</span>
                <h3>{{ step.title }}</h3>
                <p>{{ step.description }}</p>
              </div>
              <span v-if="step.required" class="tutorial-quickstart-required">首次必做</span>
            </header>

            <div v-if="step.notice" class="tutorial-quickstart-notice">{{ step.notice }}</div>

            <div v-if="step.command" class="tutorial-quickstart-code">
              <div class="tutorial-quickstart-code-head">
                <span>{{ step.commandLabel }}</span>
                <button type="button" @click="copyQuickstartCommand(step.command, $event)">复制</button>
              </div>
              <pre><code>{{ step.command }}</code></pre>
            </div>

            <router-link v-if="step.link" :to="step.link.to" class="tutorial-quickstart-link">
              {{ step.link.label }}
            </router-link>
            <a
              v-if="step.externalLink"
              :href="step.externalLink.href"
              class="tutorial-quickstart-link"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ step.externalLink.label }}
            </a>
          </article>
        </div>

        <div class="tutorial-quickstart-followup">
          <section class="tutorial-quickstart-section">
            <header>
              <span class="tutorial-kicker">{{ quickstartConfig.desktop.kicker }}</span>
              <h3>{{ quickstartConfig.desktop.title }}</h3>
              <p>{{ quickstartConfig.desktop.description }}</p>
            </header>
            <div class="tutorial-quickstart-tile-grid">
              <article v-for="tile in quickstartDesktopTiles" :key="tile.number" class="tutorial-quickstart-tile">
                <span>{{ tile.number }}</span>
                <strong>{{ tile.title }}</strong>
                <p>{{ tile.description }}</p>
              </article>
            </div>
          </section>

          <section class="tutorial-quickstart-section">
            <header>
              <span class="tutorial-kicker">{{ quickstartConfig.api.kicker }}</span>
              <h3>{{ quickstartConfig.api.title }}</h3>
              <p>{{ quickstartConfig.api.description }}</p>
            </header>
            <div class="tutorial-quickstart-code tutorial-quickstart-code--large">
              <div class="tutorial-quickstart-code-head">
                <span>{{ quickstartApiLabel }}</span>
                <button type="button" @click="copyQuickstartCommand(quickstartApiExample, $event)">复制</button>
              </div>
              <pre><code>{{ quickstartApiExample }}</code></pre>
            </div>
            <p class="tutorial-quickstart-hint">{{ quickstartConfig.api_hint }}</p>
          </section>

          <section class="tutorial-quickstart-section">
            <header>
              <span class="tutorial-kicker">{{ quickstartConfig.troubleshooting.kicker }}</span>
              <h3>{{ quickstartConfig.troubleshooting.title }}</h3>
              <p>{{ quickstartConfig.troubleshooting.description }}</p>
            </header>
            <div class="tutorial-quickstart-error-grid">
              <article v-for="error in quickstartErrors" :key="error.code" class="tutorial-quickstart-error">
                <span>{{ error.code }}</span>
                <strong>{{ error.title }}</strong>
                <p>{{ error.description }}</p>
              </article>
            </div>
          </section>
        </div>
      </section>

      <div v-if="isTutorialLibraryView && loading && orderedPages.length === 0" class="tutorial-loading" role="status" aria-live="polite">
        <div class="tutorial-spinner"></div>
        <span>加载教程中...</span>
      </div>

      <template v-else>
        <div v-if="isTutorialLibraryView && sourceNotice" class="tutorial-source-notice" :class="`tutorial-source-notice--${sourceNotice.type}`">
          <div>
            <strong>{{ sourceNotice.title }}</strong>
            <p>{{ sourceNotice.description }}</p>
          </div>
          <button type="button" @click="refreshPages">重试</button>
        </div>

        <section v-if="isTutorialLibraryView" class="tutorial-reader tutorial-reader--index">
          <article class="tutorial-main-column">
            <header class="tutorial-article-head">
              <div>
                <span>目录</span>
                <h1>完整教程目录</h1>
                <p>按工具、使用场景或排查主题查找需要的教程。</p>
              </div>
              <button type="button" class="tutorial-refresh" :disabled="loading" @click="refreshPages">
                刷新
              </button>
            </header>

            <div class="tutorial-index-controls">
              <label class="tutorial-search">
                <span>搜索教程</span>
                <input
                  v-model.trim="searchQuery"
                  type="search"
                  placeholder="搜索标题、分类或内容简介"
                />
              </label>

              <div class="tutorial-category-filter" role="group" aria-label="教程分类">
                <button
                  type="button"
                  :class="{ 'is-active': selectedCategory === 'all' }"
                  :aria-pressed="selectedCategory === 'all'"
                  @click="selectedCategory = 'all'"
                >
                  全部
                </button>
                <button
                  v-for="category in tutorialCategories"
                  :key="category"
                  type="button"
                  :class="{ 'is-active': selectedCategory === category }"
                  :aria-pressed="selectedCategory === category"
                  @click="selectedCategory = category"
                >
                  {{ category }}
                </button>
              </div>
            </div>

            <div v-if="categoryGroups.length === 0" class="tutorial-directory-empty">
              <strong>没有匹配的教程</strong>
              <p>换一个关键词或分类继续查找。</p>
              <button type="button" @click="resetDirectoryFilters">清除筛选</button>
            </div>

            <div v-else class="tutorial-category-groups">
              <section v-for="group in categoryGroups" :key="group.category" class="tutorial-category-group">
                <header>
                  <h3>{{ group.category }}</h3>
                  <span>{{ group.pages.length }} 篇</span>
                </header>
                <div class="tutorial-directory-grid">
                  <router-link
                    v-for="page in group.pages"
                    :key="page.slug"
                    :to="`/tutorial/${page.slug}`"
                    class="tutorial-directory-card"
                  >
                    <span>{{ page.category || '教程' }}</span>
                    <strong>{{ page.title }}</strong>
                    <p>{{ page.description }}</p>
                  </router-link>
                </div>
              </section>
            </div>
          </article>
        </section>

        <section v-else-if="!isIndexRoute" class="tutorial-reader tutorial-reader--detail">
          <aside class="tutorial-sidebar" aria-label="接入教程目录">
            <p class="tutorial-sidebar-title">教程目录</p>
            <nav class="tutorial-tabs">
              <router-link
                v-for="page in orderedPages"
                :key="page.slug"
                :to="`/tutorial/${page.slug}`"
                class="tutorial-tab-link"
                :class="{ 'is-active': activeSlug === page.slug }"
                :aria-current="activeSlug === page.slug ? 'page' : undefined"
              >
                <strong>{{ page.title }}</strong>
                <span>{{ page.category || '教程' }}</span>
              </router-link>
            </nav>
          </aside>

          <div class="tutorial-detail-column">
            <div class="tutorial-mobile-directory">
              <button
                type="button"
                class="tutorial-mobile-directory-toggle"
                aria-controls="tutorial-mobile-directory-list"
                :aria-expanded="mobileDirectoryOpen"
                @click="mobileDirectoryOpen = !mobileDirectoryOpen"
              >
                <span>当前教程</span>
                <strong>{{ activePageTitle }}</strong>
                <span>{{ mobileDirectoryOpen ? '收起目录' : '展开目录' }}</span>
              </button>
              <nav
                v-show="mobileDirectoryOpen"
                id="tutorial-mobile-directory-list"
                class="tutorial-mobile-directory-list"
                aria-label="移动端教程目录"
              >
                <router-link
                  v-for="page in orderedPages"
                  :key="page.slug"
                  :to="`/tutorial/${page.slug}`"
                  :class="{ 'is-active': activeSlug === page.slug }"
                  :aria-current="activeSlug === page.slug ? 'page' : undefined"
                  @click="mobileDirectoryOpen = false"
                >
                  <strong>{{ page.title }}</strong>
                  <span>{{ page.category || '教程' }}</span>
                </router-link>
              </nav>
            </div>

            <div
              v-if="detailState === 'loading' || (loading && detailState === 'idle')"
              class="tutorial-detail-state tutorial-detail-state--loading"
              role="status"
              aria-live="polite"
            >
              <div class="tutorial-spinner"></div>
              <strong>正在加载当前教程</strong>
              <p>正文准备好后会直接显示在这里。</p>
            </div>

            <div v-else-if="detailState === 'error'" class="tutorial-detail-state tutorial-detail-state--error" role="alert">
              <span>加载失败</span>
              <h2>暂时无法打开这篇教程</h2>
              <p>{{ detailError }}</p>
              <div class="tutorial-state-actions">
                <button type="button" @click="retryActivePage">重试</button>
                <router-link to="/tutorial">返回教程目录</router-link>
              </div>
            </div>

            <div v-else-if="detailState === 'notFound'" class="tutorial-empty tutorial-detail-state--not-found">
              <span>404</span>
              <h2>教程不存在</h2>
              <p>该页面未发布或已下线，请从教程目录选择其他内容。</p>
              <router-link to="/tutorial" class="guide-action-link">返回教程目录</router-link>
            </div>

            <article v-else-if="activePage" class="tutorial-main-column tutorial-article">
              <header class="tutorial-article-head tutorial-article-head--compact">
                <div>
                  <div class="tutorial-article-meta">
                    <span>{{ activePage.category || '教程' }}</span>
                    <span>{{ articleProgressLabel }}</span>
                  </div>
                  <h1>{{ activePage.title }}</h1>
                  <p>{{ activePage.description }}</p>
                </div>
                <button type="button" class="tutorial-refresh" :disabled="loading" @click="refreshPages">
                  刷新
                </button>
              </header>

              <div class="tutorial-content-shell">
                <details v-if="tocItems.length" class="tutorial-mobile-toc">
                  <summary>
                    <span>本页目录</span>
                    <small>{{ tocItems.length }} 个章节</small>
                  </summary>
                  <nav aria-label="移动端当前文章目录">
                    <button
                      v-for="item in tocItems"
                      :key="item.id"
                      type="button"
                      :class="[`toc-level-${item.level}`, { 'is-active': activeHeadingId === item.id }]"
                      :aria-current="activeHeadingId === item.id ? 'location' : undefined"
                      @click="navigateToHeading(item.id)"
                    >
                      {{ item.text }}
                    </button>
                  </nav>
                </details>

                <div
                  ref="contentRef"
                  class="tutorial-content"
                  v-html="renderedHtml"
                  @click="handleContentClick"
                  @keydown="handleContentKeydown"
                ></div>

              </div>

              <nav class="tutorial-article-pagination" aria-label="教程篇间导航">
                <router-link
                  v-if="previousPage"
                  :to="`/tutorial/${previousPage.slug}`"
                  class="tutorial-page-link tutorial-page-link--previous"
                >
                  <span>上一篇</span>
                  <strong>{{ previousPage.title }}</strong>
                </router-link>
                <span v-else class="tutorial-page-link tutorial-page-link--disabled">
                  <span>上一篇</span>
                  <strong>已经是第一篇</strong>
                </span>

                <router-link
                  v-if="nextPage"
                  :to="`/tutorial/${nextPage.slug}`"
                  class="tutorial-page-link tutorial-page-link--next"
                >
                  <span>下一篇</span>
                  <strong>{{ nextPage.title }}</strong>
                </router-link>
                <span v-else class="tutorial-page-link tutorial-page-link--next tutorial-page-link--disabled">
                  <span>下一篇</span>
                  <strong>已经是最后一篇</strong>
                </span>
              </nav>
            </article>

            <div v-else class="tutorial-detail-state tutorial-detail-state--loading" role="status">
              <div class="tutorial-spinner"></div>
              <strong>正在准备教程内容</strong>
            </div>
          </div>

          <aside v-if="tocItems.length" class="tutorial-toc" aria-label="当前文章目录">
            <p>本页目录</p>
            <button
              v-for="item in tocItems"
              :key="item.id"
              type="button"
              :class="[`toc-level-${item.level}`, { 'is-active': activeHeadingId === item.id }]"
              :aria-current="activeHeadingId === item.id ? 'location' : undefined"
              @click="navigateToHeading(item.id)"
            >
              {{ item.text }}
            </button>
          </aside>
        </section>
      </template>
    </main>

    <div
      v-if="imagePreview"
      class="tutorial-image-lightbox"
      role="dialog"
      aria-modal="true"
      :aria-label="imagePreview.alt || '教程截图预览'"
      @click.self="closeImagePreview"
    >
      <button
        ref="lightboxCloseRef"
        type="button"
        class="tutorial-image-lightbox__close"
        aria-label="关闭图片预览"
        @click="closeImagePreview"
      >
        关闭
      </button>
      <figure>
        <img :src="imagePreview.src" :alt="imagePreview.alt" />
        <figcaption v-if="imagePreview.caption">{{ imagePreview.caption }}</figcaption>
      </figure>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import tutorialsAPI from '@/api/tutorials'
import type { TutorialPage, TutorialPageSummary } from '@/types'
import { renderTutorialMarkdown, type TutorialTocItem } from '@/utils/tutorialMarkdown'
import PublicRevealBackdrop from './components/PublicRevealBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'
import { tutorialFallbackPages } from './tutorialFallback'
import {
  defaultQuickstartTutorialConfig,
  type QuickstartPlatformID,
  type QuickstartTutorialConfig
} from './tutorialQuickstart'

type TutorialScrollBehavior = 'auto' | 'smooth'

type DetailState = 'idle' | 'loading' | 'ready' | 'error' | 'notFound'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const legacyHashRedirects: Record<string, string> = {
  quickstart: 'getting-started',
  'quick-start': 'getting-started',
  platforms: 'getting-started',
  codex: 'codex',
  claude: 'claude-code',
  'claude-code': 'claude-code',
  ccswitch: 'cc-switch',
  'cc-switch': 'cc-switch',
  cockpit: 'cockpit-tools',
  'cockpit-tools': 'cockpit-tools',
  'recover-session': 'recover-codex-session',
  'restore-session': 'recover-codex-session',
  'recover-codex-session': 'recover-codex-session',
  minepilotqa: 'minepilotqa',
  openclaw: 'openclaw',
  hermes: 'hermes-agent',
  'hermes-agent': 'hermes-agent',
  faq: 'faq'
}

const loading = ref(false)
const summaries = ref<TutorialPageSummary[]>([])
const loadedPages = ref<Record<string, TutorialPage>>({})
const usingFallback = ref(false)
const sourceState = ref<'cms' | 'fallback-empty' | 'fallback-error'>('cms')
const sourceError = ref('')
const renderedHtml = ref('')
const tocItems = ref<TutorialTocItem[]>([])
const activeHeadingId = ref('')
const contentRef = ref<HTMLElement | null>(null)
const detailState = ref<DetailState>('idle')
const detailError = ref('')
const searchQuery = ref('')
const selectedCategory = ref('all')
const mobileDirectoryOpen = ref(false)
const imagePreview = ref<{ src: string; alt: string; caption: string } | null>(null)
const lightboxCloseRef = ref<HTMLButtonElement | null>(null)
let observer: IntersectionObserver | null = null
let detailRequestId = 0
let imagePreviewTrigger: HTMLElement | null = null
const copyFeedbackTimers = new Map<HTMLButtonElement, number>()

type QuickstartTerminal = 'cmd' | 'powershell' | 'unix'

const quickstartConfig = ref<QuickstartTutorialConfig>(defaultQuickstartTutorialConfig)
const quickstartPlatforms = computed(() => quickstartConfig.value.platforms)
const quickstartTerminals = [
  { value: 'cmd', label: 'Windows CMD' },
  { value: 'powershell', label: 'PowerShell' },
  { value: 'unix', label: 'macOS / Linux' }
] as const
const quickstartPlatform = ref<QuickstartPlatformID>('codex')
const quickstartTerminal = ref<QuickstartTerminal>('cmd')
const quickstartErrors = computed(() => quickstartConfig.value.errors)

const orderedPages = computed(() => {
  return [...summaries.value].sort((a, b) => {
    if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
    if (a.category !== b.category) return a.category.localeCompare(b.category, 'zh-Hans-CN')
    return a.id - b.id
  })
})

const routeSlug = computed(() => String(route.params.slug || ''))
const isIndexRoute = computed(() => !routeSlug.value)
const isTutorialLibraryView = computed(() => isIndexRoute.value && route.query.view === 'library')
const showQuickstart = computed(() => isIndexRoute.value && !isTutorialLibraryView.value)
const activeSlug = computed(() => routeSlug.value)
const activePage = computed(() => loadedPages.value[activeSlug.value] ?? null)
const activeSummary = computed(() => orderedPages.value.find((page) => page.slug === activeSlug.value) ?? null)
const activePageTitle = computed(() => activePage.value?.title || activeSummary.value?.title || '教程详情')
const activePageIndex = computed(() => orderedPages.value.findIndex((page) => page.slug === activeSlug.value))
const previousPage = computed(() => {
  const index = activePageIndex.value
  return index > 0 ? orderedPages.value[index - 1] : null
})
const nextPage = computed(() => {
  const index = activePageIndex.value
  return index >= 0 && index < orderedPages.value.length - 1 ? orderedPages.value[index + 1] : null
})
const articleProgressLabel = computed(() => {
  if (activePageIndex.value < 0) return `共 ${orderedPages.value.length} 篇`
  return `第 ${activePageIndex.value + 1} 篇，共 ${orderedPages.value.length} 篇`
})
const tutorialCategories = computed(() => {
  return Array.from(new Set(orderedPages.value.map((page) => page.category || '教程')))
})
const filteredPages = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return orderedPages.value.filter((page) => {
    const category = page.category || '教程'
    if (selectedCategory.value !== 'all' && category !== selectedCategory.value) return false
    if (!query) return true
    return [page.title, page.description, category, page.slug].some((value) => value.toLowerCase().includes(query))
  })
})
const categoryGroups = computed(() => {
  const groups = new Map<string, TutorialPageSummary[]>()
  filteredPages.value.forEach((page) => {
    const category = page.category || '教程'
    const pages = groups.get(category) ?? []
    pages.push(page)
    groups.set(category, pages)
  })
  return Array.from(groups, ([category, pages]) => ({ category, pages }))
})
const sourceNotice = computed(() => {
  if (sourceState.value === 'fallback-empty') {
    return {
      type: 'empty',
      title: '当前暂时使用备用教程',
      description: '最新教程正在准备中，你可以先继续阅读当前内容。'
    }
  }
  if (sourceState.value === 'fallback-error') {
    return {
      type: 'error',
      title: '当前暂时使用备用教程',
      description: '教程内容暂时无法更新，你可以先继续阅读当前内容，稍后再试。'
    }
  }
  return null
})

const quickstartTerminalLabel = computed(() => {
  return quickstartTerminals.find((option) => option.value === quickstartTerminal.value)?.label || 'Windows CMD'
})

const quickstartClient = computed(() => {
  const platform = quickstartPlatforms.value.find((option) => option.id === quickstartPlatform.value)
    ?? quickstartPlatforms.value[0]
    ?? defaultQuickstartTutorialConfig.platforms[0]
  return {
    name: platform.client_name,
    baseUrl: platform.base_url,
    baseUrlDescription: platform.base_url_description,
    auth: platform.auth_hint,
    protocol: platform.protocol,
    model: platform.model_hint
  }
})

const quickstartFacts = computed(() => [
  {
    label: quickstartConfig.value.facts.base_url_label,
    value: quickstartClient.value.baseUrl,
    description: quickstartClient.value.baseUrlDescription
  },
  {
    label: quickstartConfig.value.facts.auth_label,
    value: quickstartClient.value.auth,
    description: quickstartConfig.value.facts.auth_description
  },
  {
    label: quickstartConfig.value.facts.protocol_label,
    value: quickstartClient.value.protocol,
    description: quickstartConfig.value.facts.protocol_description
  },
  {
    label: quickstartConfig.value.facts.model_label,
    value: quickstartClient.value.model,
    description: quickstartConfig.value.facts.model_description
  }
])

const quickstartInstallCommand = computed(() => {
  if (quickstartPlatform.value === 'claude') {
    return 'npm install -g @anthropic-ai/claude-code --registry=https://registry.npmmirror.com\nclaude --version'
  }
  if (quickstartTerminal.value === 'unix') {
    return 'npm install -g @openai/codex --registry=https://registry.npmmirror.com\ncodex --version'
  }
  return 'winget install --id OpenAI.Codex -e\ncodex --version'
})

const quickstartConfigDirectoryCommand = computed(() => {
  if (quickstartPlatform.value === 'claude') return ''
  if (quickstartTerminal.value === 'unix') return 'mkdir -p ~/.codex\ncd ~/.codex'
  if (quickstartTerminal.value === 'powershell') {
    return 'New-Item -ItemType Directory -Force "$env:USERPROFILE\\.codex" | Out-Null\nexplorer "$env:USERPROFILE\\.codex"'
  }
  return 'if not exist "%USERPROFILE%\\.codex" mkdir "%USERPROFILE%\\.codex"\nexplorer "%USERPROFILE%\\.codex"'
})

const quickstartConfigDirectoryDescription = computed(() => {
  if (quickstartPlatform.value === 'claude') {
    return 'Claude Code 不使用 config.toml。下一步直接在当前终端设置环境变量；需要长期生效时，再将同样的变量写入终端的配置文件。'
  }
  if (quickstartTerminal.value === 'unix') {
    return '默认配置文件是 ~/.codex/config.toml。macOS 可在 Finder 按 Cmd+Shift+G 后输入 ~/.codex；Linux 请在文件管理器按 Ctrl+H 显示隐藏目录。'
  }
  return '默认配置文件是 C:\\Users\\你的用户名\\.codex\\config.toml。也可在文件资源管理器地址栏输入 %USERPROFILE%\\.codex；找不到目录时执行下方命令。'
})

const quickstartConfigDirectoryCommandLabel = computed(() => {
  if (quickstartPlatform.value === 'claude') return ''
  if (quickstartTerminal.value === 'unix') return '创建并进入目录'
  return `${quickstartTerminalLabel.value} - 打开配置目录`
})

const quickstartConfigSnippet = computed(() => {
  if (quickstartPlatform.value === 'claude') {
    if (quickstartTerminal.value === 'unix') {
      return `export ANTHROPIC_AUTH_TOKEN="你的 API Key"\nexport ANTHROPIC_BASE_URL="${quickstartClient.value.baseUrl}"`
    }
    if (quickstartTerminal.value === 'powershell') {
      return `$env:ANTHROPIC_AUTH_TOKEN="你的 API Key"\n$env:ANTHROPIC_BASE_URL="${quickstartClient.value.baseUrl}"`
    }
    return `set ANTHROPIC_AUTH_TOKEN=你的 API Key\nset ANTHROPIC_BASE_URL=${quickstartClient.value.baseUrl}`
  }
  return `model = "gpt-5.5"\nmodel_provider = "luoye"\n\n[model_providers.luoye]\nname = "luoye"\nbase_url = "${quickstartClient.value.baseUrl}"\nenv_key = "OPENAI_API_KEY"\nwire_api = "responses"`
})

const quickstartAuthAndStartCommand = computed(() => {
  if (quickstartPlatform.value === 'claude') {
    return `${quickstartConfigSnippet.value}\nclaude`
  }
  return '{\n  "OPENAI_API_KEY": "替换成你的 API Key"\n}\n\ncodex'
})

const quickstartApiLabel = computed(() => {
  return quickstartPlatform.value === 'claude' ? 'cURL / Anthropic Messages' : 'cURL / OpenAI Responses'
})

const quickstartApiExample = computed(() => {
  const continuation = '\\'
  if (quickstartPlatform.value === 'claude') {
    return [
      'curl ' + quickstartClient.value.baseUrl + '/v1/messages ' + continuation,
      '  -H "x-api-key: 你的 API Key" ' + continuation,
      '  -H "anthropic-version: 2023-06-01" ' + continuation,
      '  -H "Content-Type: application/json" ' + continuation,
      "  -d '{\"model\":\"claude-sonnet-4-5\",\"max_tokens\":128,\"messages\":[{\"role\":\"user\",\"content\":\"你好\"}]}'",
    ].join('\n')
  }
  return [
    'curl ' + quickstartClient.value.baseUrl + '/responses ' + continuation,
    '  -H "Authorization: Bearer 你的 API Key" ' + continuation,
    '  -H "Content-Type: application/json" ' + continuation,
    "  -d '{\"model\":\"gpt-5.5\",\"input\":\"你好\"}'",
  ].join('\n')
})

const quickstartSteps = computed(() => {
  return [
    {
      number: '01',
      kicker: '准备信息',
      title: '确认 API Key 和 Base URL',
      description: `先从控制台复制 ${quickstartClient.value.name} 对应的 API Key。`,
      commandLabel: '接入信息',
      command: `${quickstartClient.value.baseUrl}\n${quickstartClient.value.auth}`,
      required: true,
      link: { to: '/keys', label: '打开 API 密钥页面' }
    },
    {
      number: '02',
      kicker: quickstartTerminalLabel.value,
      title: quickstartPlatform.value === 'claude' ? '安装 Claude Code' : '安装 Codex CLI 或桌面端',
      description: quickstartPlatform.value === 'claude' ? '需要 Node.js 18 或更高版本；桌面端用户可以跳过 CLI 安装。' : 'Codex App 不要求先安装 Node.js；只有使用 CLI 时才执行下面的安装命令。',
      commandLabel: '安装命令',
      command: quickstartInstallCommand.value,
      required: true,
      externalLink: quickstartPlatform.value === 'claude'
        ? undefined
        : {
            href: 'https://developers.openai.com/codex/app#getting-started',
            label: '下载 ChatGPT Desktop（Windows / macOS）'
          }
    },
    {
      number: '03',
      kicker: '配置文件位置',
      title: quickstartPlatform.value === 'claude' ? '确认 Claude 配置方式' : '找到或创建 config.toml',
      description: quickstartConfigDirectoryDescription.value,
      commandLabel: quickstartConfigDirectoryCommandLabel.value,
      command: quickstartConfigDirectoryCommand.value,
      required: true,
      notice: quickstartPlatform.value === 'claude'
        ? undefined
        : '在打开的目录中新建或编辑 config.toml；确认 Windows 没有把文件保存成 config.toml.txt。'
    },
    {
      number: '04',
      kicker: quickstartPlatform.value === 'claude' ? '环境变量' : 'config.toml',
      title: '填写接口配置',
      description: quickstartPlatform.value === 'claude' ? '将环境变量写入当前终端或 settings.json。' : '将下面内容保存到 config.toml，并把模型替换为账号可用的模型 ID。',
      commandLabel: quickstartPlatform.value === 'claude' ? quickstartTerminalLabel.value : 'config.toml',
      command: quickstartConfigSnippet.value,
      required: true
    },
    {
      number: '05',
      kicker: '最后一步',
      title: '填写密钥并启动验证',
      description: quickstartPlatform.value === 'claude' ? '保存 Token 后运行 claude，看到可交互提示即完成接入。' : '保存 auth.json 后启动 Codex，发送一句简单问题确认模型能正常响应。',
      commandLabel: quickstartPlatform.value === 'claude' ? quickstartTerminalLabel.value : 'auth.json + 启动命令',
      command: quickstartAuthAndStartCommand.value,
      required: true,
      notice: '真实 API Key 只粘贴到本地配置，不要提交到代码仓库。'
    }
  ]
})

const quickstartDesktopTiles = computed(() => quickstartConfig.value.desktop.tiles)

function fallbackSummary(page: TutorialPage): TutorialPageSummary {
  const { content_md: _content, ...summary } = page
  return summary
}

function setFallbackPages(reason: 'empty' | 'error', message = '') {
  usingFallback.value = true
  sourceState.value = reason === 'empty' ? 'fallback-empty' : 'fallback-error'
  sourceError.value = message
  summaries.value = tutorialFallbackPages.map(fallbackSummary)
  loadedPages.value = Object.fromEntries(tutorialFallbackPages.map((page) => [page.slug, page]))
}

function normalizeHash(hash: string): string {
  const raw = hash.replace(/^#/, '').trim()
  try {
    return decodeURIComponent(raw).toLowerCase()
  } catch {
    return raw.toLowerCase()
  }
}

function resolveLegacyHashTarget(): string {
  return legacyHashRedirects[normalizeHash(route.hash)] ?? ''
}

function getErrorStatus(error: unknown): number | undefined {
  return (error as { response?: { status?: number } })?.response?.status
}

function getErrorMessage(error: unknown, fallback: string): string {
  const responseMessage = (error as { response?: { data?: { message?: string } } })?.response?.data?.message
  if (responseMessage) return responseMessage
  return error instanceof Error && error.message ? error.message : fallback
}

function cachePage(page: TutorialPage) {
  loadedPages.value = { ...loadedPages.value, [page.slug]: page }
  if (!summaries.value.some((summary) => summary.slug === page.slug)) {
    summaries.value = [...summaries.value, fallbackSummary(page)]
  }
}

function removeCachedPage(slug: string) {
  const nextPages = { ...loadedPages.value }
  delete nextPages[slug]
  loadedPages.value = nextPages
}

async function loadQuickstartConfig() {
  try {
    const config = await tutorialsAPI.getQuickstartConfig()
    if (!Array.isArray(config.platforms) || config.platforms.length === 0) return
    quickstartConfig.value = config
    if (!config.platforms.some((platform) => platform.id === quickstartPlatform.value)) {
      quickstartPlatform.value = config.platforms[0].id
    }
  } catch (error) {
    console.warn('Failed to load quickstart tutorial config; using built-in defaults.', error)
  }
}

async function loadPages(force = false) {
  if (loading.value) return
  if (!force && summaries.value.length > 0) return
  loading.value = true
  try {
    const items = await tutorialsAPI.list()
    if (items.length === 0) {
      setFallbackPages('empty')
      return
    }
    usingFallback.value = false
    sourceState.value = 'cms'
    sourceError.value = ''
    summaries.value = items
    loadedPages.value = {}
  } catch (error: unknown) {
    setFallbackPages('error', getErrorMessage(error, '公开教程接口请求失败，当前显示内置默认教程。'))
  } finally {
    loading.value = false
  }
}

async function ensurePage(slug: string, force = false) {
  if (!slug) {
    detailState.value = 'idle'
    detailError.value = ''
    return
  }

  if (!force && loadedPages.value[slug]) {
    detailState.value = 'ready'
    detailError.value = ''
    return
  }

  const requestId = ++detailRequestId
  const fallback = tutorialFallbackPages.find((page) => page.slug === slug)
  detailState.value = 'loading'
  detailError.value = ''

  if (usingFallback.value && fallback) {
    cachePage(fallback)
    detailState.value = 'ready'
    return
  }

  if (sourceState.value === 'fallback-error' && !fallback) {
    detailState.value = 'error'
    detailError.value = sourceError.value || '教程服务暂不可用，请稍后重试。'
    return
  }

  try {
    const page = await tutorialsAPI.getBySlug(slug)
    cachePage(page)
    if (requestId === detailRequestId && activeSlug.value === slug) {
      detailState.value = 'ready'
    }
  } catch (error: unknown) {
    if (requestId !== detailRequestId || activeSlug.value !== slug) return
    removeCachedPage(slug)
    if (getErrorStatus(error) === 404) {
      detailState.value = 'notFound'
      detailError.value = ''
      return
    }
    detailState.value = 'error'
    detailError.value = getErrorMessage(error, '教程详情加载失败，请稍后重试。')
  }
}

async function refreshPages() {
  await loadPages(true)
  if (isIndexRoute.value) return
  await ensurePage(activeSlug.value, true)
  await renderActivePage()
}

async function retryActivePage() {
  const fallback = tutorialFallbackPages.find((page) => page.slug === activeSlug.value)
  if (sourceState.value === 'fallback-error' && !fallback) {
    await refreshPages()
    return
  }
  await ensurePage(activeSlug.value, true)
  await renderActivePage()
}

function resetDirectoryFilters() {
  searchQuery.value = ''
  selectedCategory.value = 'all'
}

async function renderActivePage() {
  const page = activePage.value
  if (!page || detailState.value !== 'ready') {
    renderedHtml.value = ''
    tocItems.value = []
    activeHeadingId.value = ''
    return
  }
  const result = renderTutorialMarkdown(page.content_md, { skipTitle: page.title })
  renderedHtml.value = result.html
  tocItems.value = result.toc.filter((item) => item.level >= 2 && item.level <= 3)
  activeHeadingId.value = tocItems.value[0]?.id ?? ''
  await nextTick()
  enhanceRenderedContent()
  setupHeadingObserver()
  scrollToRouteHash('auto')
}

function setupHeadingObserver() {
  observer?.disconnect()
  observer = null
  const root = contentRef.value
  if (!root) return
  const headings = Array.from(root.querySelectorAll<HTMLElement>('h2[id], h3[id]'))
  if (!headings.length) return
  if (!('IntersectionObserver' in window)) return
  observer = new IntersectionObserver(
    (entries) => {
      const visible = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
      if (visible[0]?.target?.id) {
        activeHeadingId.value = visible[0].target.id
      }
    },
    { rootMargin: '-22% 0px -62% 0px', threshold: 0.01 }
  )
  headings.forEach((heading) => observer?.observe(heading))
}

function findHeading(id: string): HTMLElement | null {
  const root = contentRef.value
  if (!root) return null
  return Array.from(root.querySelectorAll<HTMLElement>('[id]')).find((element) => element.id === id) ?? null
}

function routeHeadingId(): string {
  const raw = route.hash.replace(/^#/, '')
  try {
    return decodeURIComponent(raw)
  } catch {
    return raw
  }
}

function scrollToHeading(id: string, behavior: TutorialScrollBehavior = 'smooth'): boolean {
  const heading = findHeading(id)
  if (!heading) return false
  activeHeadingId.value = id
  heading.scrollIntoView({ behavior, block: 'start' })
  return true
}

function scrollToRouteHash(behavior: TutorialScrollBehavior = 'auto') {
  const id = routeHeadingId()
  if (id) scrollToHeading(id, behavior)
}

async function navigateToHeading(id: string) {
  if (routeHeadingId() !== id) {
    await router.push({ path: route.path, query: route.query, hash: `#${id}` })
  }
  await nextTick()
  scrollToHeading(id, 'smooth')
}

function configureCopyButton(button: HTMLButtonElement) {
  button.dataset.copyLabel = button.dataset.copyLabel || button.textContent?.trim() || '复制'
  button.setAttribute('aria-label', button.getAttribute('aria-label') || '复制代码')
}

function enhanceRenderedContent() {
  const root = contentRef.value
  if (!root) return

  root.querySelectorAll<HTMLButtonElement>('[data-copy-code]').forEach(configureCopyButton)
  root.querySelectorAll<HTMLPreElement>('pre').forEach((pre) => {
    if (pre.closest('.tutorial-command-block')) return
    const code = pre.querySelector('code')
    if (!code) return

    const block = document.createElement('div')
    block.className = 'tutorial-command-block tutorial-code-block'
    const header = document.createElement('div')
    header.className = 'command-block-header'
    const title = document.createElement('span')
    const languageClass = Array.from(code.classList).find((className) => className.startsWith('language-'))
    title.textContent = languageClass ? languageClass.replace('language-', '') : '代码'
    const button = document.createElement('button')
    button.type = 'button'
    button.className = 'copy-command-button'
    button.dataset.copyCode = encodeURIComponent(code.textContent || '')
    button.textContent = '复制'
    configureCopyButton(button)
    header.append(title, button)
    pre.replaceWith(block)
    block.append(header, pre)
  })

  root.querySelectorAll<HTMLElement>('.tutorial-screenshot-card').forEach((card) => {
    const image = card.querySelector<HTMLImageElement>('img')
    card.tabIndex = 0
    card.setAttribute('role', 'button')
    card.setAttribute('aria-label', `预览图片：${image?.alt || '教程截图'}`)
  })
}

function showCopyFeedback(button: HTMLButtonElement, state: 'success' | 'error') {
  const previousTimer = copyFeedbackTimers.get(button)
  if (previousTimer) window.clearTimeout(previousTimer)

  button.dataset.copyState = state
  button.textContent = state === 'success' ? '已复制' : '复制失败'
  const timer = window.setTimeout(() => {
    button.textContent = button.dataset.copyLabel || '复制'
    delete button.dataset.copyState
    copyFeedbackTimers.delete(button)
  }, 2200)
  copyFeedbackTimers.set(button, timer)
}

async function copyQuickstartCommand(command: string, event: MouseEvent) {
  const button = event.currentTarget
  if (!(button instanceof HTMLButtonElement)) return
  try {
    if (!navigator.clipboard?.writeText) throw new Error('Clipboard API unavailable')
    await navigator.clipboard.writeText(command)
    showCopyFeedback(button, 'success')
    appStore.showSuccess('已复制命令')
  } catch {
    showCopyFeedback(button, 'error')
    appStore.showError('复制失败，请手动选择命令')
  }
}

async function handleContentClick(event: MouseEvent) {
  const target = event.target instanceof Element ? event.target : null
  const copyButton = target?.closest<HTMLButtonElement>('[data-copy-code]')
  if (copyButton) {
    try {
      const encoded = copyButton.dataset.copyCode || ''
      const text = decodeURIComponent(encoded)
      if (!navigator.clipboard?.writeText) throw new Error('Clipboard API unavailable')
      await navigator.clipboard.writeText(text)
      showCopyFeedback(copyButton, 'success')
      appStore.showSuccess('已复制命令')
    } catch {
      showCopyFeedback(copyButton, 'error')
      appStore.showError('复制失败，请手动选择命令')
    }
    return
  }

  const screenshotCard = target?.closest<HTMLElement>('.tutorial-screenshot-card')
  if (screenshotCard) openImagePreview(screenshotCard)
}

function handleContentKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' && event.key !== ' ') return
  const target = event.target instanceof Element ? event.target : null
  const screenshotCard = target?.closest<HTMLElement>('.tutorial-screenshot-card')
  if (!screenshotCard) return
  event.preventDefault()
  openImagePreview(screenshotCard)
}

function openImagePreview(screenshotCard: HTMLElement) {
  const image = screenshotCard.querySelector<HTMLImageElement>('img')
  if (!image) return
  imagePreviewTrigger = screenshotCard
  imagePreview.value = {
    src: image.currentSrc || image.src,
    alt: image.alt || '教程截图',
    caption: screenshotCard.querySelector('figcaption')?.textContent?.trim() || image.alt || ''
  }
  nextTick(() => lightboxCloseRef.value?.focus())
}

function closeImagePreview() {
  if (!imagePreview.value) return
  const trigger = imagePreviewTrigger
  imagePreview.value = null
  imagePreviewTrigger = null
  nextTick(() => {
    if (trigger?.isConnected) trigger.focus()
  })
}

function clearImagePreview() {
  imagePreview.value = null
  imagePreviewTrigger = null
}

function handleGlobalKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && imagePreview.value) {
    event.preventDefault()
    closeImagePreview()
  }
}

function handleLegacyHashRedirect() {
  if (!isIndexRoute.value) return
  const target = resolveLegacyHashTarget()
  if (!target || routeSlug.value === target) return
  if (target) {
    router.replace({ path: `/tutorial/${target}`, hash: '' })
  }
}

watch(
  () => route.hash,
  async () => {
    if (isIndexRoute.value) {
      handleLegacyHashRedirect()
      return
    }
    await nextTick()
    scrollToRouteHash('auto')
  }
)

watch(
  activeSlug,
  async (slug) => {
    mobileDirectoryOpen.value = false
    clearImagePreview()
    await ensurePage(slug)
    await renderActivePage()
  }
)

watch(tutorialCategories, (categories) => {
  if (selectedCategory.value !== 'all' && !categories.includes(selectedCategory.value)) {
    selectedCategory.value = 'all'
  }
})

onMounted(async () => {
  window.addEventListener('keydown', handleGlobalKeydown)
  handleLegacyHashRedirect()
  await Promise.all([loadPages(), loadQuickstartConfig()])
  await ensurePage(activeSlug.value)
  await renderActivePage()
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
  observer?.disconnect()
  copyFeedbackTimers.forEach((timer) => window.clearTimeout(timer))
  copyFeedbackTimers.clear()
})
</script>

<style scoped>
@import './public-page.css';

.tutorial-page {
  overflow-x: clip;
  --tutorial-panel: rgba(250, 249, 245, 0.9);
  --tutorial-panel-strong: #faf9f5;
  --tutorial-border: #e6dfd8;
  --tutorial-border-strong: #d8cec2;
  --tutorial-text: #141413;
  --tutorial-muted: #6c6a64;
  background: #faf9f5;
  color: var(--tutorial-text);
}

.tutorial-main {
  width: min(100%, 104rem);
  padding: 1.25rem clamp(1rem, 2vw, 2rem) 4rem;
}

.tutorial-main.is-article-route {
  width: 100%;
  padding-top: 0.75rem;
}

.tutorial-overview {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(26rem, 0.85fr);
  gap: 1rem;
  margin-bottom: 1rem;
}

.tutorial-intro,
.beginner-path,
.tutorial-sidebar,
.tutorial-main-column,
.tutorial-empty,
.tutorial-loading,
.tutorial-detail-state,
.tutorial-mobile-directory {
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: var(--tutorial-panel);
  box-shadow: 0 12px 32px rgba(20, 20, 19, 0.05);
  backdrop-filter: blur(16px);
}

.tutorial-source-notice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(250, 249, 245, 0.9);
  color: var(--tutorial-text);
}

.tutorial-source-notice--empty {
  border-color: #d8cec2;
  background: rgba(245, 240, 232, 0.82);
}

.tutorial-source-notice--error {
  border-color: rgba(217, 119, 6, 0.28);
  background: rgba(251, 191, 36, 0.12);
}

.tutorial-source-notice strong {
  display: block;
  color: var(--tutorial-text);
  font-size: 0.92rem;
}

.tutorial-source-notice p {
  margin: 0.2rem 0 0;
  color: var(--tutorial-muted);
  font-size: 0.88rem;
}

.tutorial-source-notice button {
  flex: 0 0 auto;
  min-height: 2.25rem;
  padding: 0.45rem 0.8rem;
  border: 1px solid var(--tutorial-border-strong);
  border-radius: 999px;
  background: #faf9f5;
  color: var(--public-accent);
  font-weight: 500;
}

.tutorial-intro {
  padding: clamp(1.2rem, 2vw, 2rem);
}

.tutorial-kicker {
  display: inline-flex;
  margin-bottom: 0.75rem;
  color: var(--public-accent);
  font-size: 0.75rem;
  font-weight: 500;
}

.tutorial-intro h1,
.tutorial-intro h2 {
  margin: 0;
  max-width: 12em;
  font-family: var(--public-font-display);
  font-size: 3.8rem;
  font-weight: 400;
  line-height: 0.98;
}

.tutorial-intro p {
  margin: 1rem 0 0;
  max-width: 48rem;
  color: var(--tutorial-muted);
  line-height: 1.8;
}

.tutorial-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.7rem;
  margin-top: 1.2rem;
}

.guide-action-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.45rem;
  padding: 0.65rem 1rem;
  border: 1px solid #cc785c;
  border-radius: 999px;
  background: #cc785c;
  color: #ffffff;
  font-weight: 500;
}

.guide-action-link--ghost {
  border-color: var(--tutorial-border);
  background: #faf9f5;
  color: var(--tutorial-text);
}

.beginner-path {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  padding: 1rem;
}

.beginner-step {
  min-height: 9rem;
  padding: 1rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.82);
  color: var(--tutorial-text);
  text-decoration: none;
  transition: border-color 0.16s ease, background 0.16s ease, transform 0.16s ease;
}

.beginner-step:hover {
  border-color: var(--tutorial-border-strong);
  background: var(--public-accent-soft);
  transform: translateY(-1px);
}

.beginner-step span {
  color: var(--public-accent);
  font-size: 0.75rem;
  font-weight: 500;
}

.beginner-step strong {
  display: block;
  margin-top: 0.35rem;
  font-size: 1.05rem;
}

.beginner-step p {
  margin: 0.5rem 0 0;
  color: var(--tutorial-muted);
  font-size: 0.88rem;
  line-height: 1.55;
}

.tutorial-reader {
  display: grid;
  grid-template-columns: 18rem minmax(0, 1fr);
  gap: 1rem;
}

.tutorial-reader--index {
  grid-template-columns: minmax(0, 1fr);
}

.tutorial-reader--detail {
  grid-template-columns:
    clamp(11.25rem, 14vw, 18rem)
    minmax(0, 1fr)
    minmax(0, 50rem)
    clamp(9.5rem, 12vw, 15rem)
    minmax(0, 1fr);
  align-items: start;
  gap: clamp(0.75rem, 1.5vw, 2rem);
}

.tutorial-reader--detail > .tutorial-sidebar {
  grid-column: 1;
}

.tutorial-detail-column {
  grid-column: 3;
  width: 100%;
  min-width: 0;
  max-width: 50rem;
  justify-self: center;
}

.tutorial-reader--detail > .tutorial-toc {
  grid-column: 4;
}

.tutorial-mobile-directory {
  display: none;
}

.tutorial-sidebar {
  position: sticky;
  top: var(--tutorial-sticky-top, 5rem);
  z-index: 12;
  align-self: start;
  max-height: calc(100vh - var(--tutorial-sticky-top, 5rem) - 1rem);
  padding: 0.85rem;
  overflow: auto;
  overscroll-behavior: contain;
}

.tutorial-sidebar-title,
.tutorial-toc p {
  margin: 0 0 0.65rem;
  color: var(--public-accent);
  font-size: 0.75rem;
  font-weight: 500;
}

.tutorial-tabs {
  display: grid;
  gap: 0.45rem;
}

.tutorial-tabs a {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
  min-height: 2.35rem;
  padding: 0.45rem 0.55rem;
  border: 1px solid transparent;
  border-radius: 8px;
  color: var(--tutorial-muted);
}

.tutorial-tabs strong {
  min-width: 0;
  overflow: hidden;
  color: inherit;
  font-size: 0.92rem;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tutorial-tabs a:hover,
.tutorial-tabs a.is-active {
  border-color: var(--tutorial-border-strong);
  background: var(--public-accent-soft);
  color: var(--public-accent);
}

.tutorial-tabs span {
  flex: 0 0 auto;
  font-size: 0.75rem;
  color: var(--public-muted-soft);
  line-height: 1;
  white-space: nowrap;
}

.tutorial-main-column {
  min-width: 0;
  padding: 1rem;
}

.tutorial-article {
  width: 100%;
  padding: 0 clamp(0.25rem, 1vw, 0.75rem);
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  backdrop-filter: none;
}

.tutorial-article-head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.4rem 0.3rem 1rem;
  border-bottom: 1px solid var(--tutorial-border);
}

.tutorial-article-head span {
  color: var(--public-accent);
  font-size: 0.75rem;
  font-weight: 500;
}

.tutorial-article-head h2 {
  margin: 0.2rem 0;
  font-family: var(--public-font-display);
  font-size: 2.35rem;
  font-weight: 400;
}

.tutorial-article-head h1 {
  margin: 0.25rem 0 0.45rem;
  font-family: var(--public-font-display);
  font-size: 2.25rem;
  font-weight: 400;
  line-height: 1.08;
}

.tutorial-article-head p {
  margin: 0;
  color: var(--tutorial-muted);
}

.tutorial-article-head--compact {
  align-items: flex-start;
  padding: 0.15rem 0.15rem 0.9rem;
}

.tutorial-article-head--compact > div {
  min-width: 0;
}

.tutorial-article-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem 0.75rem;
}

.tutorial-article-meta span + span {
  color: var(--tutorial-muted);
}

.tutorial-refresh {
  align-self: start;
  min-height: 2.25rem;
  padding: 0.45rem 0.8rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 999px;
  background: #faf9f5;
  color: var(--public-accent);
  font-weight: 500;
}

.tutorial-content-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
  align-items: start;
  margin-top: 1rem;
}

.tutorial-content-shell > .tutorial-content {
  grid-column: 1;
  grid-row: 1;
}

.tutorial-mobile-toc {
  display: none;
}

.tutorial-index-controls {
  display: grid;
  gap: 0.8rem;
  margin-top: 1rem;
  padding: 0.9rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.58);
}

.tutorial-search {
  display: grid;
  gap: 0.4rem;
}

.tutorial-search span {
  color: var(--tutorial-muted);
  font-size: 0.78rem;
  font-weight: 500;
}

.tutorial-search input {
  width: 100%;
  min-height: 2.65rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--tutorial-border-strong);
  border-radius: 8px;
  outline: none;
  background: #faf9f5;
  color: var(--tutorial-text);
}

.tutorial-search input:focus-visible {
  border-color: var(--public-accent);
  box-shadow: 0 0 0 3px rgba(204, 120, 92, 0.14);
}

.tutorial-category-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.tutorial-category-filter button {
  min-height: 2.25rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--tutorial-border-strong);
  border-radius: 999px;
  background: #faf9f5;
  color: var(--tutorial-muted);
  font-size: 0.86rem;
}

.tutorial-category-filter button:hover,
.tutorial-category-filter button.is-active {
  border-color: rgba(204, 120, 92, 0.42);
  background: var(--public-accent-soft);
  color: var(--public-accent-strong);
}

.tutorial-category-groups {
  display: grid;
  gap: 1.25rem;
  margin-top: 1.2rem;
}

.tutorial-category-group > header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  padding: 0 0.15rem;
}

.tutorial-category-group h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.tutorial-category-group > header span {
  color: var(--tutorial-muted);
  font-size: 0.78rem;
}

.tutorial-directory-empty {
  margin-top: 1rem;
  padding: 2rem 1rem;
  border: 1px dashed var(--tutorial-border-strong);
  border-radius: 8px;
  text-align: center;
}

.tutorial-directory-empty p {
  margin: 0.35rem 0 0.9rem;
  color: var(--tutorial-muted);
}

.tutorial-directory-empty button {
  min-height: 2.25rem;
  padding: 0.45rem 0.8rem;
  border: 1px solid var(--tutorial-border-strong);
  border-radius: 999px;
  background: #faf9f5;
  color: var(--public-accent);
}

.tutorial-directory-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));
  gap: 0.8rem;
  margin-top: 0.65rem;
}

.tutorial-directory-card {
  display: grid;
  min-height: 10rem;
  align-content: start;
  gap: 0.45rem;
  padding: 1rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(250, 249, 245, 0.9);
  color: var(--tutorial-text);
}

.tutorial-directory-card:hover {
  border-color: var(--tutorial-border-strong);
  background: var(--public-accent-soft);
}

.tutorial-directory-card span {
  color: var(--public-accent);
  font-size: 0.75rem;
  font-weight: 500;
}

.tutorial-directory-card strong {
  font-size: 1.08rem;
}

.tutorial-directory-card p {
  margin: 0;
  color: var(--tutorial-muted);
  line-height: 1.6;
}

.tutorial-content {
  min-width: 0;
  color: var(--tutorial-text);
}

.tutorial-content :deep(h1),
.tutorial-content :deep(h2),
.tutorial-content :deep(h3) {
  font-family: var(--public-font-display);
  scroll-margin-top: 6rem;
  margin: 1.6rem 0 0.7rem;
  color: var(--tutorial-text);
  font-weight: 400;
}

.tutorial-content :deep(h1) {
  margin-top: 0;
  font-size: 2rem;
}

.tutorial-content :deep(h2) {
  font-size: 1.35rem;
}

.tutorial-content :deep(p),
.tutorial-content :deep(li) {
  color: var(--tutorial-muted);
  line-height: 1.82;
}

.tutorial-content :deep(a) {
  color: var(--public-accent);
  font-weight: 500;
  text-decoration-line: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.18em;
}

.tutorial-content :deep(a:hover) {
  color: var(--public-accent-strong);
  text-decoration-thickness: 2px;
}

.tutorial-content :deep(.guide-action-link) {
  display: inline;
  min-height: 0;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: var(--public-accent);
  font-weight: 500;
  text-decoration-line: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.18em;
}

.tutorial-content :deep(ul),
.tutorial-content :deep(ol) {
  padding-left: 1.25rem;
}

.tutorial-content :deep(code) {
  border-radius: 6px;
  background: #f5f0e8;
  color: #141413;
  padding: 0.1rem 0.3rem;
}

.tutorial-content :deep(.tutorial-command-block) {
  margin: 1rem 0;
  overflow: hidden;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: #181715;
}

.tutorial-content :deep(.command-block-header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.65rem 0.8rem;
  border-bottom: 1px solid rgba(250, 249, 245, 0.14);
  color: #faf9f5;
  font-size: 0.82rem;
  font-weight: 500;
}

.tutorial-content :deep(.copy-command-button) {
  min-height: 1.75rem;
  padding: 0.25rem 0.55rem;
  border: 1px solid rgba(250, 249, 245, 0.24);
  border-radius: 999px;
  background: rgba(250, 249, 245, 0.1);
  color: #faf9f5;
}

.tutorial-content :deep(.copy-command-button[data-copy-state='success']) {
  border-color: rgba(172, 205, 162, 0.7);
  background: rgba(172, 205, 162, 0.18);
  color: #f4f7ed;
}

.tutorial-content :deep(.copy-command-button[data-copy-state='error']) {
  border-color: rgba(239, 154, 134, 0.72);
  background: rgba(239, 154, 134, 0.16);
  color: #fff1ed;
}

.tutorial-content :deep(pre) {
  margin: 0;
  overflow-x: auto;
  padding: 0.9rem;
}

.tutorial-content :deep(pre code) {
  display: block;
  min-width: max-content;
  background: transparent;
  color: #faf9f5;
  padding: 0;
}

.tutorial-content :deep(.tutorial-callout) {
  margin: 1rem 0;
  padding: 0.9rem 1rem;
  border: 1px solid #e6dfd8;
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.82);
}

.tutorial-content :deep(.tutorial-callout strong) {
  display: block;
  margin-bottom: 0.35rem;
  color: var(--public-accent);
}

.tutorial-content :deep(.tutorial-screenshot-card) {
  display: inline-grid;
  width: fit-content;
  max-width: 100%;
  margin: 0.75rem 0.5rem 0.75rem 0;
  overflow: hidden;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: #faf9f5;
  cursor: zoom-in;
  vertical-align: top;
  transition: border-color 0.16s ease, transform 0.16s ease, background 0.16s ease;
}

.tutorial-content :deep(.tutorial-screenshot-card:hover) {
  border-color: rgba(204, 120, 92, 0.36);
  background: #f5f0e8;
  transform: translateY(-1px);
}

.tutorial-content :deep(.tutorial-screenshot-card:focus-visible) {
  outline: 3px solid rgba(204, 120, 92, 0.24);
  outline-offset: 3px;
  border-color: var(--public-accent);
}

.tutorial-content :deep(.tutorial-screenshot-card img) {
  display: block;
  width: auto;
  max-width: 100%;
  height: auto;
  object-fit: contain;
  background: #f5f0e8;
}

.tutorial-content :deep(.tutorial-screenshot-card figcaption) {
  padding: 0.55rem 0.7rem;
  color: var(--tutorial-muted);
  font-size: 0.82rem;
}

.tutorial-image-lightbox {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: grid;
  place-items: center;
  padding: clamp(1rem, 3vw, 2rem);
  background: rgba(20, 20, 19, 0.9);
  backdrop-filter: blur(12px);
}

.tutorial-image-lightbox figure {
  display: grid;
  gap: 0.75rem;
  max-width: min(96vw, 76rem);
  max-height: 92vh;
  margin: 0;
}

.tutorial-image-lightbox img {
  display: block;
  max-width: 100%;
  max-height: min(82vh, calc(100vh - 7rem));
  object-fit: contain;
  border: 1px solid rgba(250, 249, 245, 0.24);
  border-radius: 8px;
  background: #181715;
  box-shadow: 0 28px 90px rgba(0, 0, 0, 0.48);
}

.tutorial-image-lightbox figcaption {
  max-width: min(96vw, 76rem);
  color: rgba(250, 249, 245, 0.86);
  text-align: center;
  font-size: 0.92rem;
}

.tutorial-image-lightbox__close {
  position: fixed;
  top: 1rem;
  right: 1rem;
  min-height: 2.35rem;
  padding: 0.45rem 0.8rem;
  border: 1px solid rgba(250, 249, 245, 0.32);
  border-radius: 8px;
  background: rgba(20, 20, 19, 0.88);
  color: #faf9f5;
  font-weight: 500;
}

.tutorial-image-lightbox__close:hover {
  border-color: rgba(204, 120, 92, 0.8);
  background: rgba(204, 120, 92, 0.2);
}

.tutorial-toc {
  position: sticky;
  top: var(--tutorial-sticky-top, 5rem);
  z-index: 11;
  display: grid;
  max-height: calc(100vh - var(--tutorial-sticky-top, 5rem) - 1rem);
  overflow: auto;
  gap: 0.25rem;
  padding: 0.85rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(250, 249, 245, 0.9);
  box-shadow: var(--public-shadow-soft);
}

.tutorial-toc button {
  text-align: left;
  padding: 0.42rem 0.5rem;
  border-radius: 7px;
  color: var(--tutorial-muted);
  font-size: 0.86rem;
}

.tutorial-toc button:hover,
.tutorial-toc button.is-active {
  background: var(--public-accent-soft);
  color: var(--public-accent);
}

.tutorial-toc .toc-level-3 {
  padding-left: 1rem;
  font-size: 0.8rem;
}

.tutorial-detail-state {
  min-height: 16rem;
  padding: 2rem;
}

.tutorial-detail-state--loading {
  display: grid;
  place-content: center;
  justify-items: center;
  text-align: center;
}

.tutorial-detail-state--loading p,
.tutorial-detail-state--error p {
  margin: 0.45rem 0 0;
  color: var(--tutorial-muted);
}

.tutorial-detail-state--error > span,
.tutorial-detail-state--not-found > span {
  color: var(--public-accent);
  font-size: 0.78rem;
  font-weight: 600;
}

.tutorial-detail-state--error h2,
.tutorial-detail-state--not-found h2 {
  margin: 0.35rem 0 0;
  font-family: var(--public-font-display);
  font-size: 2rem;
  font-weight: 400;
}

.tutorial-state-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin-top: 1rem;
}

.tutorial-state-actions button,
.tutorial-state-actions a {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.35rem;
  padding: 0.5rem 0.85rem;
  border: 1px solid var(--tutorial-border-strong);
  border-radius: 999px;
  background: #faf9f5;
  color: var(--public-accent);
  font-weight: 500;
}

.tutorial-article-pagination {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--tutorial-border);
}

.tutorial-page-link {
  display: grid;
  min-width: 0;
  gap: 0.25rem;
  padding: 0.85rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.62);
  color: var(--tutorial-text);
}

.tutorial-page-link:hover {
  border-color: var(--tutorial-border-strong);
  background: var(--public-accent-soft);
}

.tutorial-page-link > span {
  color: var(--tutorial-muted);
  font-size: 0.75rem;
}

.tutorial-page-link strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tutorial-page-link--next {
  text-align: right;
}

.tutorial-page-link--disabled {
  opacity: 0.55;
}

.tutorial-empty,
.tutorial-loading {
  padding: 2rem;
  text-align: center;
}

.tutorial-empty p,
.tutorial-loading span {
  color: var(--tutorial-muted);
}

.tutorial-spinner {
  width: 2rem;
  height: 2rem;
  margin: 0 auto 0.8rem;
  border: 2px solid rgba(38, 37, 30, 0.14);
  border-top-color: var(--public-accent);
  border-radius: 999px;
  animation: tutorial-spin 0.8s linear infinite;
}

@keyframes tutorial-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (min-width: 1101px) and (max-width: 1360px) {
  .tutorial-tabs span {
    display: none;
  }
}

@media (max-width: 1100px) {
  .tutorial-overview,
  .tutorial-reader--detail,
  .tutorial-content-shell {
    grid-template-columns: 1fr;
  }

  .tutorial-reader--detail > .tutorial-sidebar,
  .tutorial-toc {
    display: none;
  }

  .tutorial-detail-column {
    grid-column: 1;
  }

  .tutorial-mobile-directory {
    display: block;
    margin-bottom: 0.75rem;
    padding: 0.55rem;
  }

  .tutorial-mobile-directory-toggle {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    width: 100%;
    min-height: 3.25rem;
    align-items: center;
    gap: 0.2rem 0.8rem;
    padding: 0.55rem 0.65rem;
    border-radius: 7px;
    text-align: left;
    color: var(--tutorial-text);
  }

  .tutorial-mobile-directory-toggle > span:first-child {
    grid-column: 1;
    color: var(--public-accent);
    font-size: 0.72rem;
    font-weight: 600;
  }

  .tutorial-mobile-directory-toggle strong {
    grid-column: 1;
    min-width: 0;
    overflow: hidden;
    font-size: 0.98rem;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tutorial-mobile-directory-toggle > span:last-child {
    grid-column: 2;
    grid-row: 1 / span 2;
    color: var(--tutorial-muted);
    font-size: 0.8rem;
  }

  .tutorial-mobile-directory-toggle:hover,
  .tutorial-mobile-directory-toggle:focus-visible {
    background: var(--public-accent-soft);
  }

  .tutorial-mobile-directory-list {
    display: grid;
    gap: 0.4rem;
    max-height: min(60vh, 28rem);
    margin-top: 0.45rem;
    padding-top: 0.45rem;
    overflow-y: auto;
    border-top: 1px solid var(--tutorial-border);
  }

  .tutorial-mobile-directory-list a {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    min-height: 2.6rem;
    padding: 0.55rem 0.65rem;
    border: 1px solid transparent;
    border-radius: 7px;
    color: var(--tutorial-muted);
  }

  .tutorial-mobile-directory-list a.is-active {
    border-color: var(--tutorial-border-strong);
    background: var(--public-accent-soft);
    color: var(--public-accent-strong);
  }

  .tutorial-mobile-directory-list strong {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tutorial-mobile-directory-list span {
    flex: 0 0 auto;
    font-size: 0.75rem;
  }

  .tutorial-mobile-toc {
    display: block;
    grid-column: 1;
    padding: 0.7rem 0.8rem;
    border: 1px solid var(--tutorial-border);
    border-radius: 8px;
    background: rgba(245, 240, 232, 0.62);
  }

  .tutorial-mobile-toc summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    min-height: 2rem;
    color: var(--tutorial-text);
    cursor: pointer;
    list-style: none;
    font-weight: 600;
  }

  .tutorial-mobile-toc summary::-webkit-details-marker {
    display: none;
  }

  .tutorial-mobile-toc summary small {
    color: var(--tutorial-muted);
    font-size: 0.75rem;
    font-weight: 400;
  }

  .tutorial-mobile-toc nav {
    display: grid;
    gap: 0.25rem;
    margin-top: 0.55rem;
    padding-top: 0.55rem;
    border-top: 1px solid var(--tutorial-border);
  }

  .tutorial-mobile-toc button {
    min-height: 2.35rem;
    padding: 0.45rem 0.5rem;
    border-radius: 7px;
    text-align: left;
    color: var(--tutorial-muted);
  }

  .tutorial-mobile-toc button:hover,
  .tutorial-mobile-toc button.is-active {
    background: var(--public-accent-soft);
    color: var(--public-accent-strong);
  }

  .tutorial-mobile-toc .toc-level-3 {
    padding-left: 1rem;
    font-size: 0.82rem;
  }

  .tutorial-content-shell > .tutorial-content {
    grid-column: 1;
    grid-row: auto;
  }
}

@media (max-width: 640px) {
  .tutorial-main {
    padding: 0.75rem 0.65rem 2.5rem;
  }

  .tutorial-main {
    width: min(100%, 34rem);
  }

  .tutorial-main.is-article-route {
    padding-top: 0.45rem;
  }

  .tutorial-intro h1,
  .tutorial-intro h2 {
    font-size: 2rem;
  }

  .tutorial-source-notice {
    display: grid;
  }

  .tutorial-main.is-index-route .beginner-path {
    display: none;
  }

  .beginner-step {
    min-height: 0;
  }

  .tutorial-article-head {
    display: grid;
  }

  .tutorial-article-head h1,
  .tutorial-article-head h2 {
    font-size: 1.75rem;
  }

  .tutorial-refresh {
    justify-self: start;
  }

  .tutorial-mobile-directory-toggle {
    min-height: 3.5rem;
  }

  .tutorial-directory-grid {
    grid-template-columns: 1fr;
  }

  .tutorial-directory-card {
    min-height: auto;
  }

  .tutorial-detail-state {
    min-height: 13rem;
    padding: 1.4rem 1rem;
  }

  .tutorial-article-pagination {
    grid-template-columns: 1fr;
  }

  .tutorial-page-link--next {
    text-align: left;
  }
}

.tutorial-quickstart {
  display: grid;
  gap: 1rem;
  margin-bottom: 1rem;
  padding: clamp(1rem, 2vw, 1.5rem);
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(250, 249, 245, 0.9);
  box-shadow: 0 16px 36px rgba(20, 20, 19, 0.06);
  backdrop-filter: blur(16px);
}

.tutorial-quickstart-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.25rem 0.25rem 1.1rem;
  border-bottom: 1px solid var(--tutorial-border);
}

.tutorial-quickstart-head h1,
.tutorial-quickstart-head h2 {
  margin: 0.15rem 0 0.35rem;
  font-family: var(--public-font-display);
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 400;
  line-height: 1.05;
}

.tutorial-quickstart-head p,
.tutorial-quickstart-section header p {
  margin: 0;
  color: var(--tutorial-muted);
  line-height: 1.7;
}

.tutorial-quickstart-head .tutorial-actions {
  flex: 0 0 auto;
  margin-top: 0;
}

.tutorial-quickstart-controls {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.8rem;
  padding: 0.15rem 0.25rem 0;
}

.tutorial-quickstart-control {
  display: grid;
  gap: 0.45rem;
  min-width: 0;
}

.tutorial-quickstart-control > span {
  color: var(--tutorial-muted);
  font-size: 0.78rem;
  font-weight: 600;
}

.tutorial-segmented-control {
  display: flex;
  flex-wrap: wrap;
  gap: 0.2rem;
  padding: 0.2rem;
  border: 1px solid var(--tutorial-border-strong);
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.74);
}

.tutorial-segmented-control button {
  flex: 1 1 auto;
  min-height: 2.2rem;
  padding: 0.45rem 0.65rem;
  border: 1px solid transparent;
  border-radius: 6px;
  color: var(--tutorial-muted);
  font-size: 0.82rem;
  font-weight: 500;
  white-space: nowrap;
}

.tutorial-segmented-control button:hover,
.tutorial-segmented-control button.is-active {
  border-color: rgba(204, 120, 92, 0.38);
  background: #faf9f5;
  color: var(--public-accent-strong);
  box-shadow: 0 3px 10px rgba(20, 20, 19, 0.05);
}

.tutorial-quickstart-facts {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.tutorial-quickstart-fact {
  display: grid;
  align-content: start;
  min-width: 0;
  min-height: 7.6rem;
  padding: 0.9rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.72);
}

.tutorial-quickstart-fact > span,
.tutorial-quickstart-error > span {
  color: var(--public-accent-strong);
  font-size: 0.76rem;
  font-weight: 600;
}

.tutorial-quickstart-fact strong {
  min-width: 0;
  margin-top: 0.3rem;
  overflow-wrap: anywhere;
  color: var(--tutorial-text);
  font-family: var(--public-font-mono);
  font-size: 0.9rem;
  line-height: 1.35;
}

.tutorial-quickstart-fact p,
.tutorial-quickstart-error p,
.tutorial-quickstart-tile p {
  margin: 0.45rem 0 0;
  color: var(--tutorial-muted);
  font-size: 0.82rem;
  line-height: 1.55;
}

.tutorial-quickstart-steps {
  display: grid;
  gap: 0.8rem;
}

.tutorial-quickstart-step {
  display: grid;
  gap: 0.8rem;
  padding: 1rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(250, 249, 245, 0.9);
}

.tutorial-quickstart-step > header {
  display: grid;
  grid-template-columns: 2.15rem minmax(0, 1fr) auto;
  align-items: start;
  gap: 0.75rem;
}

.tutorial-quickstart-step-number {
  display: grid;
  width: 2.15rem;
  height: 2.15rem;
  place-items: center;
  border: 1px solid rgba(204, 120, 92, 0.48);
  border-radius: 50%;
  color: var(--public-accent-strong);
  font-size: 0.82rem;
  font-weight: 600;
}

.tutorial-quickstart-step-kicker {
  color: var(--public-accent-strong);
  font-size: 0.74rem;
  font-weight: 600;
  text-transform: uppercase;
}

.tutorial-quickstart-step h3 {
  margin: 0.15rem 0 0.25rem;
  color: var(--tutorial-text);
  font-size: 1.15rem;
  font-weight: 600;
}

.tutorial-quickstart-step header p {
  margin: 0;
  color: var(--tutorial-muted);
  font-size: 0.88rem;
  line-height: 1.6;
}

.tutorial-quickstart-required {
  align-self: start;
  padding: 0.32rem 0.55rem;
  border: 1px solid rgba(204, 120, 92, 0.3);
  border-radius: 999px;
  background: rgba(204, 120, 92, 0.08);
  color: var(--public-accent-strong);
  font-size: 0.72rem;
  white-space: nowrap;
}

.tutorial-quickstart-notice,
.tutorial-quickstart-hint {
  margin: 0;
  padding: 0.75rem 0.9rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.72);
  color: var(--tutorial-muted);
  font-size: 0.84rem;
  line-height: 1.6;
}

.tutorial-quickstart-code {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #2e2c28;
  border-radius: 8px;
  background: #181715;
  color: #faf9f5;
}

.tutorial-quickstart-code-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-height: 2.2rem;
  padding: 0.45rem 0.75rem;
  border-bottom: 1px solid #35322e;
  color: #a09d96;
  font-size: 0.76rem;
}

.tutorial-quickstart-code-head span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tutorial-quickstart-code-head button {
  flex: 0 0 auto;
  min-height: 1.8rem;
  padding: 0.25rem 0.45rem;
  border: 1px solid transparent;
  border-radius: 5px;
  color: #e7a084;
  font-size: 0.76rem;
}

.tutorial-quickstart-code-head button:hover,
.tutorial-quickstart-code-head button[data-copy-state='success'] {
  border-color: rgba(231, 160, 132, 0.42);
  background: rgba(204, 120, 92, 0.16);
}

.tutorial-quickstart-code-head button[data-copy-state='error'] {
  border-color: rgba(248, 113, 113, 0.46);
  color: #fca5a5;
}

.tutorial-quickstart-code pre {
  max-width: 100%;
  margin: 0;
  padding: 0.9rem 0.95rem 1rem;
  overflow-x: auto;
  color: #faf9f5;
  font-family: var(--public-font-mono);
  font-size: 0.82rem;
  line-height: 1.7;
  white-space: pre;
}

.tutorial-quickstart-code code {
  font-family: inherit;
}

.tutorial-quickstart-link {
  justify-self: start;
  color: var(--public-accent-strong);
  font-size: 0.86rem;
  font-weight: 600;
}

.tutorial-quickstart-link:hover {
  text-decoration: underline;
  text-underline-offset: 0.2em;
}

.tutorial-quickstart-followup {
  display: grid;
  gap: 1rem;
}

.tutorial-quickstart-section {
  display: grid;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(250, 249, 245, 0.86);
}

.tutorial-quickstart-section header h3 {
  margin: 0.15rem 0 0.25rem;
  font-size: 1.35rem;
  font-weight: 600;
}

.tutorial-quickstart-tile-grid,
.tutorial-quickstart-error-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.tutorial-quickstart-tile,
.tutorial-quickstart-error {
  min-width: 0;
  padding: 0.9rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(245, 240, 232, 0.7);
}

.tutorial-quickstart-tile > span {
  color: var(--public-accent-strong);
  font-size: 0.78rem;
  font-weight: 600;
}

.tutorial-quickstart-tile strong,
.tutorial-quickstart-error strong {
  display: block;
  margin-top: 0.35rem;
  color: var(--tutorial-text);
  font-size: 0.98rem;
}

.tutorial-quickstart-error-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.tutorial-quickstart-code--large pre {
  min-height: 7.4rem;
}

@media (max-width: 900px) {
  .tutorial-quickstart-controls,
  .tutorial-quickstart-facts {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .tutorial-quickstart-tile-grid,
  .tutorial-quickstart-error-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .tutorial-quickstart {
    padding: 0.8rem;
  }

  .tutorial-quickstart-head {
    display: grid;
  }

  .tutorial-quickstart-head .tutorial-actions {
    width: 100%;
  }

  .tutorial-quickstart-head .guide-action-link {
    flex: 1 1 auto;
  }

  .tutorial-quickstart-controls,
  .tutorial-quickstart-facts,
  .tutorial-quickstart-tile-grid,
  .tutorial-quickstart-error-grid {
    grid-template-columns: 1fr;
  }

  .tutorial-quickstart-step > header {
    grid-template-columns: 2rem minmax(0, 1fr);
  }

  .tutorial-quickstart-step-number {
    width: 2rem;
    height: 2rem;
  }

  .tutorial-quickstart-required {
    grid-column: 2;
    justify-self: start;
  }

  .tutorial-quickstart-code pre {
    font-size: 0.76rem;
  }
}
</style>
