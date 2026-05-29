<template>
  <AppLayout>
    <div class="chat-image-embed-layout">
      <div v-if="loadingWorkspace" class="chat-image-embed-state">
        <Icon name="refresh" size="lg" class="animate-spin text-primary-500" />
        <p>{{ t('chatImageStudio.embeddedLoading') }}</p>
      </div>

      <div v-else-if="workspaceError" class="chat-image-embed-state">
        <Icon name="exclamationTriangle" size="xl" class="text-amber-500" />
        <h2>{{ t('chatImageStudio.embeddedOpenFailedTitle') }}</h2>
        <p>{{ workspaceError }}</p>
        <button type="button" class="btn btn-primary" @click="openWorkspace">
          <Icon name="refresh" size="sm" />
          <span>{{ t('chatImageStudio.reloadWorkspace') }}</span>
        </button>
      </div>

      <iframe
        v-else-if="embeddedUrl"
        :key="iframeKey"
        :src="embeddedUrl"
        :title="t('chatImageStudio.embeddedFrameTitle')"
        class="chat-image-embed-frame"
        data-testid="chat-image-embedded-frame"
        allow="clipboard-read; clipboard-write; fullscreen; camera; microphone"
        allowfullscreen
      ></iframe>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { openWebUIAPI } from '@/api'
import { useAppStore } from '@/stores'
import { buildEmbeddedUrl, detectTheme } from '@/utils/embedded-url'

const { t, locale } = useI18n()
const route = useRoute()
const appStore = useAppStore()

const previousSidebarCollapsed = appStore.sidebarCollapsed
let autoCollapsedSidebar = false

if (!previousSidebarCollapsed) {
  appStore.setSidebarCollapsed(true)
  autoCollapsedSidebar = true
}

const EMBEDDED_INTENT_QUERY_KEYS = new Set(['prompt', 'mode', 'reference_image_id'])

const loadingWorkspace = ref(false)
const embeddedUrl = ref('')
const workspaceError = ref('')
const iframeKey = ref(0)
const pageTheme = ref<'light' | 'dark'>(detectTheme())

let themeObserver: MutationObserver | null = null
let launchRequestId = 0

watch(
  () => route.fullPath,
  () => {
    void openWorkspace()
  },
  { immediate: true }
)

onMounted(() => {
  pageTheme.value = detectTheme()
  if (typeof document !== 'undefined') {
    themeObserver = new MutationObserver(() => {
      pageTheme.value = detectTheme()
    })
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    })
  }
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
  if (autoCollapsedSidebar && appStore.sidebarCollapsed) {
    appStore.setSidebarCollapsed(previousSidebarCollapsed)
  }
})

async function openWorkspace(): Promise<void> {
  const requestId = ++launchRequestId
  loadingWorkspace.value = true
  workspaceError.value = ''
  embeddedUrl.value = ''
  try {
    const result = await openWebUIAPI.launch()
    if (requestId !== launchRequestId) return
    embeddedUrl.value = buildLaunchFrameUrl(result.launch_url)
    iframeKey.value += 1
  } catch (error) {
    if (requestId !== launchRequestId) return
    console.error('Failed to open embedded chat image workspace:', error)
    workspaceError.value = t('chatImageStudio.embeddedOpenFailedDescription')
    appStore.showError(t('openWebUI.launchFailed'))
  } finally {
    if (requestId === launchRequestId) {
      loadingWorkspace.value = false
    }
  }
}

function buildLaunchFrameUrl(launchUrl: string): string {
  return appendRouteIntent(
    buildEmbeddedUrl(launchUrl, undefined, null, pageTheme.value, String(locale.value))
  )
}

function appendRouteIntent(inputUrl: string): string {
  try {
    const url = new URL(inputUrl)
    for (const [key, value] of Object.entries(route.query)) {
      if (!EMBEDDED_INTENT_QUERY_KEYS.has(key)) continue
      if (value === undefined || value === null) continue
      url.searchParams.delete(key)
      const values = Array.isArray(value) ? value : [value]
      for (const item of values) {
        if (item === undefined || item === null) continue
        url.searchParams.append(key, String(item))
      }
    }
    return url.toString()
  } catch {
    return inputUrl
  }
}
</script>

<style scoped>
.chat-image-embed-layout {
  display: flex;
  height: calc(100vh - 64px - 1.5rem);
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  background: rgb(248 250 252);
}

.dark .chat-image-embed-layout {
  background: rgb(15 23 42);
}

.chat-image-embed-frame {
  display: block;
  width: 100%;
  height: 100%;
  border: 0;
  background: rgb(255 255 255);
}

.chat-image-embed-state {
  display: flex;
  height: 100%;
  min-height: 320px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.875rem;
  padding: 2rem;
  text-align: center;
}

.chat-image-embed-state h2 {
  font-size: 1rem;
  font-weight: 800;
  color: rgb(15 23 42);
}

.chat-image-embed-state p {
  max-width: 30rem;
  font-size: 0.875rem;
  line-height: 1.6;
  color: rgb(100 116 139);
}

.dark .chat-image-embed-state h2 {
  color: rgb(248 250 252);
}

.dark .chat-image-embed-state p {
  color: rgb(148 163 184);
}
</style>
