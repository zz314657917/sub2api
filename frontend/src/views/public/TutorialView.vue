<template>
  <div class="tutorial-page relative min-h-screen overflow-hidden text-white">
    <PublicMatrixBackdrop />
    <PublicTopNav />

    <main
      class="tutorial-main relative z-10 mx-auto"
      :class="{ 'is-article-route': !isIndexRoute, 'is-index-route': isIndexRoute }"
    >
      <section class="tutorial-overview">
        <div class="tutorial-intro">
          <span class="tutorial-kicker">AI 接入教程</span>
          <h1>从快速开始到工具配置</h1>
          <p>
            教程内容现在由后台发布，前台自动读取已发布文章；后台没有内容时使用内置默认教程兜底。
          </p>
          <div class="tutorial-actions">
            <router-link v-if="firstPage" :to="`/tutorial/${firstPage.slug}`" class="guide-action-link">
              新手最快路线
            </router-link>
            <router-link to="/models" class="guide-action-link guide-action-link--ghost">
              查看模型广场
            </router-link>
          </div>
        </div>

        <div class="beginner-path" aria-label="接入路线图">
          <article
            v-for="(page, index) in orderedPages.slice(0, 4)"
            :key="page.slug"
            class="beginner-step"
          >
            <span>{{ String(index + 1).padStart(2, '0') }}</span>
            <strong>{{ page.title }}</strong>
            <p>{{ page.description }}</p>
          </article>
        </div>
      </section>

      <div v-if="loading" class="tutorial-loading">
        <div class="tutorial-spinner"></div>
        <span>加载教程中...</span>
      </div>

      <template v-else>
        <div v-if="sourceNotice" class="tutorial-source-notice" :class="`tutorial-source-notice--${sourceNotice.type}`">
          <div>
            <strong>{{ sourceNotice.title }}</strong>
            <p>{{ sourceNotice.description }}</p>
          </div>
          <button type="button" @click="refreshPages">重试</button>
        </div>

        <section class="tutorial-reader">
          <aside class="tutorial-sidebar" aria-label="接入教程目录">
            <p class="tutorial-sidebar-title">教程目录</p>
            <nav class="tutorial-tabs">
              <router-link
                v-for="page in orderedPages"
                :key="page.slug"
                :to="`/tutorial/${page.slug}`"
                :class="{ 'is-active': activeSlug === page.slug }"
              >
                <strong>{{ page.title }}</strong>
                <span>{{ page.category || '教程' }}</span>
              </router-link>
            </nav>
          </aside>

          <article v-if="isIndexRoute" class="tutorial-main-column">
            <header class="tutorial-article-head">
              <div>
                <span>目录</span>
                <h2>选择一篇教程开始</h2>
                <p>先按快速开始创建密钥，再根据你使用的工具进入对应配置页。</p>
              </div>
              <button type="button" class="tutorial-refresh" :disabled="loading" @click="refreshPages">
                刷新
              </button>
            </header>

            <div class="tutorial-directory-grid">
              <router-link
                v-for="page in orderedPages"
                :key="page.slug"
                :to="`/tutorial/${page.slug}`"
                class="tutorial-directory-card"
              >
                <span>{{ page.category || '教程' }}</span>
                <strong>{{ page.title }}</strong>
                <p>{{ page.description }}</p>
              </router-link>
            </div>
          </article>

          <article v-else-if="activePage" class="tutorial-main-column">
            <header class="tutorial-article-head">
              <div>
                <span>{{ activePage.category || '教程' }}</span>
                <h2>{{ activePage.title }}</h2>
                <p>{{ activePage.description }}</p>
              </div>
              <button type="button" class="tutorial-refresh" :disabled="loading" @click="refreshPages">
                刷新
              </button>
            </header>

            <div class="tutorial-content-shell">
              <div
                ref="contentRef"
                class="tutorial-content"
                v-html="renderedHtml"
                @click="handleContentClick"
              ></div>

              <aside v-if="tocItems.length" class="tutorial-toc" aria-label="当前文章目录">
                <p>本页目录</p>
                <button
                  v-for="item in tocItems"
                  :key="item.id"
                  type="button"
                  :class="[`toc-level-${item.level}`, { 'is-active': activeHeadingId === item.id }]"
                  @click="scrollToHeading(item.id)"
                >
                  {{ item.text }}
                </button>
              </aside>
            </div>
          </article>

          <div v-else class="tutorial-empty">
            <h2>教程不存在</h2>
            <p>该页面未发布或已下线，请从左侧选择其他教程。</p>
            <router-link v-if="firstPage" :to="`/tutorial/${firstPage.slug}`" class="guide-action-link">
              返回快速开始
            </router-link>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import tutorialsAPI from '@/api/tutorials'
import type { TutorialPage, TutorialPageSummary } from '@/types'
import { renderTutorialMarkdown, type TutorialTocItem } from '@/utils/tutorialMarkdown'
import PublicMatrixBackdrop from './components/PublicMatrixBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'
import { tutorialFallbackPages } from './tutorialFallback'

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
let observer: IntersectionObserver | null = null

const orderedPages = computed(() => {
  return [...summaries.value].sort((a, b) => {
    if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
    if (a.category !== b.category) return a.category.localeCompare(b.category, 'zh-Hans-CN')
    return a.id - b.id
  })
})

const firstPage = computed(() => orderedPages.value[0] ?? null)
const routeSlug = computed(() => String(route.params.slug || ''))
const isIndexRoute = computed(() => !routeSlug.value)
const activeSlug = computed(() => routeSlug.value)
const activePage = computed(() => loadedPages.value[activeSlug.value] ?? null)
const sourceNotice = computed(() => {
  if (sourceState.value === 'fallback-empty') {
    return {
      type: 'empty',
      title: '后台暂无已发布教程',
      description: '当前显示内置默认教程，管理员发布教程后前台会自动切换到后台内容。'
    }
  }
  if (sourceState.value === 'fallback-error') {
    return {
      type: 'error',
      title: '教程 CMS 暂不可用',
      description: sourceError.value || '当前显示内置默认教程，请稍后重试或检查后端教程接口。'
    }
  }
  return null
})

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
  return hash.replace(/^#/, '').trim().toLowerCase()
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
  } catch (error: any) {
    setFallbackPages('error', error?.message || '公开教程接口请求失败，当前显示内置默认教程。')
  } finally {
    loading.value = false
  }
}

async function ensurePage(slug: string) {
  if (!slug || loadedPages.value[slug]) return
  const fallback = tutorialFallbackPages.find((page) => page.slug === slug)
  if (usingFallback.value) {
    if (fallback) loadedPages.value[slug] = fallback
    return
  }
  try {
    loadedPages.value[slug] = await tutorialsAPI.getBySlug(slug)
  } catch (error: any) {
    if (fallback) {
      sourceState.value = 'fallback-error'
      sourceError.value = error?.message || '教程详情接口请求失败，当前显示内置默认教程。'
      loadedPages.value[slug] = fallback
    }
  }
}

async function refreshPages() {
  await loadPages(true)
  await ensurePage(activeSlug.value)
  await renderActivePage()
}

async function renderActivePage() {
  const page = activePage.value
  if (!page) {
    renderedHtml.value = ''
    tocItems.value = []
    activeHeadingId.value = ''
    return
  }
  const result = renderTutorialMarkdown(page.content_md)
  renderedHtml.value = result.html
  tocItems.value = result.toc.filter((item) => item.level <= 3)
  activeHeadingId.value = tocItems.value[0]?.id ?? ''
  await nextTick()
  setupHeadingObserver()
}

function setupHeadingObserver() {
  observer?.disconnect()
  observer = null
  const root = contentRef.value
  if (!root) return
  const headings = Array.from(root.querySelectorAll<HTMLElement>('h1[id], h2[id], h3[id]'))
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

function scrollToHeading(id: string) {
  const heading = contentRef.value?.querySelector<HTMLElement>(`#${CSS.escape(id)}`)
  if (!heading) return
  heading.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function handleContentClick(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  const copyButton = target?.closest<HTMLButtonElement>('[data-copy-code]')
  if (!copyButton) return
  const encoded = copyButton.dataset.copyCode || ''
  const text = decodeURIComponent(encoded)
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess('已复制命令')
  } catch {
    appStore.showError('复制失败，请手动选择命令')
  }
}

function handleLegacyHashRedirect() {
  const hash = normalizeHash(route.hash)
  if (!hash || routeSlug.value) return
  const target = legacyHashRedirects[hash]
  if (target) {
    router.replace({ path: `/tutorial/${target}` })
  }
}

watch(
  () => route.hash,
  () => handleLegacyHashRedirect()
)

watch(
  activeSlug,
  async (slug) => {
    await ensurePage(slug)
    await renderActivePage()
  }
)

watch(activePage, () => {
  renderActivePage()
})

onMounted(async () => {
  handleLegacyHashRedirect()
  await loadPages()
  await ensurePage(activeSlug.value)
  await renderActivePage()
})

onUnmounted(() => {
  observer?.disconnect()
})
</script>

<style scoped>
.tutorial-page {
  --tutorial-panel: rgba(8, 13, 26, 0.78);
  --tutorial-panel-strong: rgba(11, 18, 32, 0.92);
  --tutorial-border: rgba(185, 209, 255, 0.16);
  --tutorial-border-strong: rgba(112, 255, 179, 0.3);
  --tutorial-text: rgba(236, 244, 255, 0.92);
  --tutorial-muted: rgba(207, 220, 240, 0.66);
  background:
    radial-gradient(circle at 18% 12%, rgba(22, 163, 74, 0.18), transparent 26rem),
    radial-gradient(circle at 90% 6%, rgba(14, 165, 233, 0.14), transparent 24rem),
    #050712;
}

.tutorial-main {
  width: min(100%, 80rem);
  padding: 1.25rem 1rem 4rem;
}

.tutorial-overview {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(20rem, 0.8fr);
  gap: 1rem;
  margin-bottom: 1rem;
}

.tutorial-intro,
.beginner-path,
.tutorial-sidebar,
.tutorial-main-column,
.tutorial-empty,
.tutorial-loading {
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.035)),
    var(--tutorial-panel);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.24);
  backdrop-filter: blur(18px);
}

.tutorial-source-notice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.7);
  color: var(--tutorial-text);
}

.tutorial-source-notice--empty {
  border-color: rgba(59, 130, 246, 0.28);
  background: rgba(37, 99, 235, 0.1);
}

.tutorial-source-notice--error {
  border-color: rgba(251, 191, 36, 0.32);
  background: rgba(180, 83, 9, 0.12);
}

.tutorial-source-notice strong {
  display: block;
  color: #f8fafc;
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
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.08);
  color: white;
  font-weight: 800;
}

.tutorial-intro {
  padding: clamp(1.2rem, 2vw, 2rem);
}

.tutorial-kicker {
  display: inline-flex;
  margin-bottom: 0.75rem;
  color: #86efac;
  font-size: 0.75rem;
  font-weight: 800;
}

.tutorial-intro h1 {
  margin: 0;
  max-width: 12em;
  font-size: clamp(2rem, 5vw, 4.1rem);
  font-weight: 950;
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
  border: 1px solid rgba(134, 239, 172, 0.35);
  border-radius: 8px;
  background: rgba(34, 197, 94, 0.16);
  color: #dcfce7;
  font-weight: 800;
}

.guide-action-link--ghost {
  border-color: rgba(147, 197, 253, 0.26);
  background: rgba(59, 130, 246, 0.12);
  color: #dbeafe;
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
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.055);
}

.beginner-step span {
  color: #67e8f9;
  font-size: 0.75rem;
  font-weight: 900;
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
  grid-template-columns: 17rem minmax(0, 1fr);
  gap: 1rem;
}

.tutorial-sidebar {
  position: sticky;
  top: 5rem;
  align-self: start;
  max-height: calc(100vh - 6rem);
  padding: 0.85rem;
  overflow: auto;
}

.tutorial-sidebar-title,
.tutorial-toc p {
  margin: 0 0 0.65rem;
  color: #93c5fd;
  font-size: 0.75rem;
  font-weight: 900;
}

.tutorial-tabs {
  display: grid;
  gap: 0.45rem;
}

.tutorial-tabs a {
  display: grid;
  gap: 0.2rem;
  padding: 0.75rem;
  border: 1px solid transparent;
  border-radius: 8px;
  color: var(--tutorial-muted);
}

.tutorial-tabs a:hover,
.tutorial-tabs a.is-active {
  border-color: var(--tutorial-border-strong);
  background: rgba(34, 197, 94, 0.11);
  color: white;
}

.tutorial-tabs span {
  font-size: 0.75rem;
  color: rgba(207, 220, 240, 0.58);
}

.tutorial-main-column {
  min-width: 0;
  padding: 1rem;
}

.tutorial-article-head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.4rem 0.3rem 1rem;
  border-bottom: 1px solid var(--tutorial-border);
}

.tutorial-article-head span {
  color: #86efac;
  font-size: 0.75rem;
  font-weight: 900;
}

.tutorial-article-head h2 {
  margin: 0.2rem 0;
  font-size: clamp(1.6rem, 3vw, 2.6rem);
  font-weight: 950;
}

.tutorial-article-head p {
  margin: 0;
  color: var(--tutorial-muted);
}

.tutorial-refresh {
  align-self: start;
  min-height: 2.25rem;
  padding: 0.45rem 0.8rem;
  border: 1px solid rgba(148, 163, 184, 0.25);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.06);
  color: white;
  font-weight: 800;
}

.tutorial-content-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 15rem;
  gap: 1rem;
  align-items: start;
  margin-top: 1rem;
}

.tutorial-directory-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));
  gap: 0.8rem;
  margin-top: 1rem;
}

.tutorial-directory-card {
  display: grid;
  min-height: 10rem;
  align-content: start;
  gap: 0.45rem;
  padding: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.055);
  color: white;
}

.tutorial-directory-card:hover {
  border-color: var(--tutorial-border-strong);
  background: rgba(34, 197, 94, 0.1);
}

.tutorial-directory-card span {
  color: #86efac;
  font-size: 0.75rem;
  font-weight: 900;
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
  scroll-margin-top: 6rem;
  margin: 1.6rem 0 0.7rem;
  color: #f8fafc;
  font-weight: 900;
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

.tutorial-content :deep(ul),
.tutorial-content :deep(ol) {
  padding-left: 1.25rem;
}

.tutorial-content :deep(code) {
  border-radius: 6px;
  background: rgba(148, 163, 184, 0.13);
  color: #bae6fd;
  padding: 0.1rem 0.3rem;
}

.tutorial-content :deep(.tutorial-command-block) {
  margin: 1rem 0;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 8px;
  background: rgba(2, 6, 23, 0.76);
}

.tutorial-content :deep(.command-block-header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.65rem 0.8rem;
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
  color: #e2e8f0;
  font-size: 0.82rem;
  font-weight: 900;
}

.tutorial-content :deep(.copy-command-button) {
  min-height: 1.75rem;
  padding: 0.25rem 0.55rem;
  border: 1px solid rgba(134, 239, 172, 0.28);
  border-radius: 7px;
  background: rgba(34, 197, 94, 0.12);
  color: #bbf7d0;
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
  color: #f8fafc;
  padding: 0;
}

.tutorial-content :deep(.tutorial-callout) {
  margin: 1rem 0;
  padding: 0.9rem 1rem;
  border: 1px solid rgba(134, 239, 172, 0.22);
  border-radius: 8px;
  background: rgba(22, 163, 74, 0.1);
}

.tutorial-content :deep(.tutorial-callout strong) {
  display: block;
  margin-bottom: 0.35rem;
  color: #bbf7d0;
}

.tutorial-content :deep(.tutorial-screenshot-card) {
  display: inline-grid;
  width: min(100%, 28rem);
  margin: 0.75rem 0.5rem 0.75rem 0;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.055);
  vertical-align: top;
}

.tutorial-content :deep(.tutorial-screenshot-card img) {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
  background: rgba(15, 23, 42, 0.8);
}

.tutorial-content :deep(.tutorial-screenshot-card figcaption) {
  padding: 0.55rem 0.7rem;
  color: var(--tutorial-muted);
  font-size: 0.82rem;
}

.tutorial-toc {
  position: sticky;
  top: 5rem;
  display: grid;
  gap: 0.25rem;
  padding: 0.85rem;
  border: 1px solid var(--tutorial-border);
  border-radius: 8px;
  background: rgba(4, 8, 18, 0.62);
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
  background: rgba(59, 130, 246, 0.14);
  color: white;
}

.tutorial-toc .toc-level-3 {
  padding-left: 1rem;
  font-size: 0.8rem;
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
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-top-color: #86efac;
  border-radius: 999px;
  animation: tutorial-spin 0.8s linear infinite;
}

@keyframes tutorial-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1100px) {
  .tutorial-overview,
  .tutorial-reader,
  .tutorial-content-shell {
    grid-template-columns: 1fr;
  }

  .tutorial-sidebar,
  .tutorial-toc {
    position: static;
    max-height: none;
  }

  .tutorial-tabs {
    display: flex;
    overflow-x: auto;
    padding-bottom: 0.2rem;
  }

  .tutorial-tabs a {
    min-width: 12rem;
    padding: 0.58rem 0.7rem;
  }
}

@media (max-width: 640px) {
  .tutorial-main {
    padding: 0.75rem 0.65rem 2.5rem;
  }

  .tutorial-main {
    width: min(100%, 34rem);
  }

  .tutorial-main.is-article-route .tutorial-overview {
    display: none;
  }

  .tutorial-main.is-article-route .tutorial-sidebar-title {
    display: none;
  }

  .tutorial-main.is-article-route .tutorial-sidebar {
    padding: 0.55rem;
  }

  .tutorial-main.is-index-route .beginner-path {
    display: none;
  }

  .tutorial-intro h1 {
    font-size: clamp(1.65rem, 8vw, 2.05rem);
  }

  .tutorial-source-notice {
    display: grid;
  }

  .beginner-path {
    grid-template-columns: 1fr;
  }

  .tutorial-article-head {
    display: grid;
  }

  .tutorial-directory-grid {
    grid-template-columns: 1fr;
  }

  .tutorial-directory-card {
    min-height: auto;
  }
}
</style>
