<template>
  <div class="legal-document-page public-page-shell min-h-screen text-white">
    <PublicMatrixBackdrop />

    <PublicTopNav />

    <main class="legal-document-main relative z-10 mx-auto w-full px-4 py-8 sm:px-6 lg:py-10">
      <section v-if="loading" class="legal-document-card legal-document-state" aria-live="polite">
        <span class="legal-document-spinner" aria-hidden="true"></span>
        <div>
          <h1>正在加载文档</h1>
          <p>正在读取登录条款配置，请稍候。</p>
        </div>
      </section>

      <section v-else-if="loadError" class="legal-document-card legal-document-state legal-document-state--danger">
        <span class="legal-document-state-icon">
          <Icon name="exclamationTriangle" size="md" />
        </span>
        <div>
          <h1>文档加载失败</h1>
          <p>请稍后刷新页面重试。</p>
        </div>
      </section>

      <section v-else-if="!currentDocument" class="legal-document-card legal-document-state">
        <span class="legal-document-state-icon">
          <Icon name="document" size="md" />
        </span>
        <div>
          <h1>文档不存在</h1>
          <p>当前条款文档不存在或已被管理员移除。</p>
        </div>
      </section>

      <article v-else class="legal-document-card">
        <header class="legal-document-header">
          <span class="legal-document-icon">
            <Icon :name="documentIcon" size="lg" />
          </span>

          <div class="min-w-0">
            <p class="legal-document-eyebrow">登录条款</p>
            <h1>{{ currentDocument.title }}</h1>
            <div class="legal-document-meta">
              <span v-if="updatedAt">
                <Icon name="calendar" size="xs" />
                更新日期：{{ updatedAt }}
              </span>
              <span>
                <Icon name="shield" size="xs" />
                阅读后可返回登录继续
              </span>
            </div>
          </div>
        </header>

        <div v-if="hasContent" class="legal-document-content" v-html="renderedHtml"></div>

        <div v-else class="legal-document-empty">
          暂无正文内容
        </div>

        <footer class="legal-document-actions" aria-label="文档操作">
          <RouterLink to="/login" class="legal-document-action legal-document-action--primary">
            返回登录
          </RouterLink>
          <RouterLink to="/home" class="legal-document-action">
            返回首页
          </RouterLink>
        </footer>
      </article>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import PublicMatrixBackdrop from './components/PublicMatrixBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'
import type { LoginAgreementDocument } from '@/types'

type LegalDocumentIcon = 'document' | 'shield' | 'globe' | 'cog'

const route = useRoute()
const appStore = useAppStore()
const loading = ref(!appStore.cachedPublicSettings && !appStore.publicSettingsLoaded)
const loadError = ref(false)

marked.setOptions({
  breaks: true,
  gfm: true,
})

const settings = computed(() => appStore.cachedPublicSettings ?? null)
const documentId = computed(() => String(route.params.documentId || ''))
const documents = computed(() => settings.value?.login_agreement_documents ?? [])
const updatedAt = computed(() => settings.value?.login_agreement_updated_at || '')

const currentDocument = computed<LoginAgreementDocument | null>(() => {
  const id = documentId.value
  if (!id) {
    return null
  }
  return documents.value.find((doc) => (doc.id || doc.title) === id) ?? null
})

const documentBodyMarkdown = computed(() => {
  const content = currentDocument.value?.content_md?.trim() || ''
  const title = currentDocument.value?.title?.trim() || ''
  if (!content || !title) {
    return content
  }

  const lines = content.split(/\r?\n/)
  const firstContentLineIndex = lines.findIndex((line) => line.trim().length > 0)
  if (firstContentLineIndex === -1) {
    return ''
  }

  const firstLine = lines[firstContentLineIndex].trim().replace(/^#{1,6}\s+/, '').trim()
  if (firstLine === title) {
    lines.splice(firstContentLineIndex, 1)
    return lines.join('\n').trim()
  }

  return content
})

const hasContent = computed(() => Boolean(documentBodyMarkdown.value))

const renderedHtml = computed(() => {
  const content = documentBodyMarkdown.value
  if (!content) {
    return ''
  }
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

const documentIcon = computed<LegalDocumentIcon>(() => {
  const title = currentDocument.value?.title || ''
  if (title.includes('政策') || title.includes('隐私')) {
    return 'shield'
  }
  if (title.includes('国家') || title.includes('地区')) {
    return 'globe'
  }
  if (title.includes('特定')) {
    return 'cog'
  }
  return 'document'
})

onMounted(async () => {
  loading.value = !appStore.cachedPublicSettings && !appStore.publicSettingsLoaded
  loadError.value = false
  try {
    await appStore.fetchPublicSettings()
    if (!appStore.cachedPublicSettings) {
      loadError.value = true
    }
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
@import './public-page.css';

.legal-document-page {
  position: relative;
  min-height: 100vh;
}

.legal-document-main {
  max-width: min(64rem, calc(100vw - 1rem));
}

.legal-document-card {
  border: 1px solid var(--public-border-strong);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.11), rgba(255, 255, 255, 0.055)),
    rgba(6, 13, 18, 0.72);
  box-shadow: var(--public-shadow);
  padding: clamp(1.35rem, 3vw, 2.25rem);
  backdrop-filter: blur(20px);
}

.legal-document-header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 1rem;
  align-items: start;
  border-bottom: 1px solid var(--public-border);
  padding-bottom: 1.35rem;
}

.legal-document-icon,
.legal-document-state-icon {
  display: inline-flex;
  height: 2.75rem;
  width: 2.75rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(119, 255, 173, 0.24);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.18), rgba(99, 102, 241, 0.1)),
    rgba(255, 255, 255, 0.07);
  color: #b7ffd0;
  box-shadow: var(--public-shadow-soft);
}

.legal-document-eyebrow {
  margin: 0;
  color: var(--public-accent);
  font-size: 0.78rem;
  font-weight: 900;
  letter-spacing: 0.04em;
}

.legal-document-header h1,
.legal-document-state h1 {
  margin: 0.4rem 0 0;
  color: #f8fafc;
  font-size: clamp(1.85rem, 4vw, 3rem);
  font-weight: 950;
  line-height: 1.08;
  letter-spacing: 0;
}

.legal-document-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  margin-top: 0.95rem;
  color: var(--public-muted);
  font-size: 0.84rem;
}

.legal-document-meta span {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  min-height: 1.85rem;
  border: 1px solid var(--public-border);
  border-radius: 8px;
  background: var(--public-surface-soft);
  padding: 0.25rem 0.55rem;
}

.legal-document-content {
  margin-top: 1.65rem;
  line-height: 1.85;
  overflow-wrap: anywhere;
  color: rgba(226, 232, 240, 0.86);
}

.legal-document-content :deep(h1) {
  margin: 2rem 0 1rem;
  border-bottom: 1px solid var(--public-border);
  padding-bottom: 0.9rem;
  color: #ffffff;
  font-size: clamp(1.6rem, 3vw, 2.15rem);
  font-weight: 900;
  line-height: 1.2;
}

.legal-document-content :deep(h2) {
  margin: 1.85rem 0 0.75rem;
  color: #f8fafc;
  font-size: clamp(1.25rem, 2vw, 1.55rem);
  font-weight: 850;
  line-height: 1.28;
}

.legal-document-content :deep(h3) {
  margin: 1.5rem 0 0.55rem;
  color: #eef2ff;
  font-size: 1.1rem;
  font-weight: 800;
}

.legal-document-content :deep(h4) {
  margin: 1.25rem 0 0.45rem;
  color: #eef2ff;
  font-size: 1rem;
  font-weight: 800;
}

.legal-document-content :deep(p) {
  margin: 0 0 1rem;
  color: rgba(226, 232, 240, 0.84);
  font-size: 0.96rem;
}

.legal-document-content :deep(a) {
  color: #98ffbd;
  text-decoration: underline;
  text-underline-offset: 4px;
  transition: color 150ms ease;
}

.legal-document-content :deep(a:hover) {
  color: #ffffff;
}

.legal-document-content :deep(ul),
.legal-document-content :deep(ol) {
  margin: 0 0 1rem;
  padding-left: 1.35rem;
  color: rgba(226, 232, 240, 0.84);
}

.legal-document-content :deep(ul) {
  list-style: disc;
}

.legal-document-content :deep(ol) {
  list-style: decimal;
}

.legal-document-content :deep(li) {
  margin: 0.35rem 0;
  padding-left: 0.1rem;
}

.legal-document-content :deep(blockquote) {
  margin: 1.35rem 0;
  border-left: 3px solid rgba(119, 255, 173, 0.52);
  border-radius: 0 8px 8px 0;
  background: rgba(119, 255, 173, 0.08);
  padding: 0.85rem 1rem;
  color: rgba(238, 246, 240, 0.82);
}

.legal-document-content :deep(code) {
  border: 1px solid var(--public-border);
  border-radius: 6px;
  background: rgba(2, 8, 10, 0.64);
  padding: 0.12rem 0.35rem;
  color: #d9ffe7;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.86em;
}

.legal-document-content :deep(pre) {
  margin: 1.35rem 0;
  overflow-x: auto;
  border: 1px solid var(--public-border);
  border-radius: 8px;
  background: rgba(2, 8, 10, 0.82);
  padding: 1rem;
  color: #f8fafc;
}

.legal-document-content :deep(pre code) {
  border: 0;
  background: transparent;
  padding: 0;
  color: inherit;
}

.legal-document-content :deep(table) {
  display: block;
  width: 100%;
  margin: 1.35rem 0;
  overflow-x: auto;
  border-collapse: collapse;
}

.legal-document-content :deep(th),
.legal-document-content :deep(td) {
  border: 1px solid var(--public-border);
  padding: 0.65rem 0.75rem;
  text-align: left;
}

.legal-document-content :deep(th) {
  background: rgba(255, 255, 255, 0.08);
  color: #f8fafc;
  font-weight: 800;
}

.legal-document-content :deep(img) {
  height: auto;
  max-width: 100%;
  margin: 1.35rem 0;
  border-radius: 8px;
}

.legal-document-content :deep(hr) {
  margin: 1.75rem 0;
  border: 0;
  border-top: 1px solid var(--public-border);
}

.legal-document-empty {
  margin-top: 1.65rem;
  border: 1px dashed var(--public-border-strong);
  border-radius: 8px;
  background: var(--public-surface-soft);
  padding: 3rem 1rem;
  text-align: center;
  color: var(--public-muted);
  font-size: 0.92rem;
}

.legal-document-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 2rem;
  border-top: 1px solid var(--public-border);
  padding-top: 1.2rem;
}

.legal-document-action {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--public-border-strong);
  border-radius: 8px;
  background: var(--public-surface-soft);
  padding: 0.5rem 0.85rem;
  color: rgba(255, 255, 255, 0.82);
  font-size: 0.86rem;
  font-weight: 850;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease,
    transform 150ms ease;
}

.legal-document-action:hover {
  border-color: var(--public-ring);
  background: var(--public-surface-hover);
  color: #ffffff;
  transform: translateY(-1px);
}

.legal-document-action--primary {
  border-color: rgba(119, 255, 173, 0.38);
  background: linear-gradient(180deg, rgba(119, 255, 173, 0.22), rgba(119, 255, 173, 0.12));
  color: #eafff1;
}

.legal-document-state {
  display: flex;
  min-height: 18rem;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  text-align: left;
}

.legal-document-state p {
  margin: 0.55rem 0 0;
  color: var(--public-muted);
  line-height: 1.7;
}

.legal-document-state--danger .legal-document-state-icon {
  border-color: rgba(248, 113, 113, 0.36);
  background: rgba(248, 113, 113, 0.12);
  color: #fecaca;
}

.legal-document-spinner {
  height: 2.6rem;
  width: 2.6rem;
  flex: 0 0 auto;
  border: 2px solid rgba(255, 255, 255, 0.14);
  border-top-color: var(--public-accent);
  border-radius: 999px;
  animation: legal-document-spin 0.82s linear infinite;
}

@keyframes legal-document-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 640px) {
  .legal-document-main {
    padding-top: 1.25rem;
    padding-bottom: 2rem;
  }

  .legal-document-card {
    padding: 1rem;
  }

  .legal-document-header,
  .legal-document-state {
    grid-template-columns: 1fr;
  }

  .legal-document-header {
    gap: 0.85rem;
  }

  .legal-document-state {
    align-items: flex-start;
    justify-content: flex-start;
    min-height: 14rem;
    text-align: left;
  }

  .legal-document-actions {
    display: grid;
    grid-template-columns: 1fr;
  }

  .legal-document-action {
    width: 100%;
  }
}
</style>
