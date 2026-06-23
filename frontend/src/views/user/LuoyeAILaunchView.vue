<template>
  <AppLayout>
    <div class="luoye-launch-layout">
      <div class="luoye-launch-panel">
        <Icon
          :name="launchError ? 'exclamationTriangle' : 'refresh'"
          size="xl"
          :class="launchError ? 'text-amber-500' : 'animate-spin text-primary-500'"
        />
        <div class="space-y-2">
          <h1>{{ launchError ? t('chatImageStudio.launchFailedTitle') : t('chatImageStudio.launchingTitle') }}</h1>
          <p>{{ launchError || t('chatImageStudio.launchingDescription') }}</p>
        </div>
        <button
          v-if="launchError"
          type="button"
          class="btn btn-primary"
          @click="openLuoyeAI"
        >
          <Icon name="refresh" size="sm" />
          <span>{{ t('chatImageStudio.retryLaunch') }}</span>
        </button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI, studioBridgeAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { ApiKey } from '@/types'
import { apiKeySupportsOpenAIImageGeneration } from '@/utils/apiKeyCapabilities'

const { t } = useI18n()
const appStore = useAppStore()

const launchError = ref('')
const launching = ref(false)

onMounted(() => {
  void openLuoyeAI()
})

function pickDefaultAPIKey(keys: ApiKey[]): ApiKey | null {
  return keys.find((key) => key.is_default) ?? keys[0] ?? null
}

async function ensureDefaultAPIKeySupportsImages(): Promise<boolean> {
  const response = await keysAPI.list(1, 100, {
    status: 'active',
    sort_by: 'created_at',
    sort_order: 'desc',
  })
  const defaultKey = pickDefaultAPIKey(response.items || [])
  if (!defaultKey || !apiKeySupportsOpenAIImageGeneration(defaultKey)) {
    launchError.value = t('chatImageStudio.defaultImageGroupMissing')
    appStore.showError(t('chatImageStudio.defaultImageGroupMissing'))
    return false
  }
  return true
}

async function openLuoyeAI(): Promise<void> {
  if (launching.value) return
  launching.value = true
  launchError.value = ''
  try {
    if (!await ensureDefaultAPIKeySupportsImages()) return
    const result = await studioBridgeAPI.launch()
    window.location.assign(result.launch_url)
  } catch (error) {
    console.error('Failed to launch Luoye Creative studio:', error)
    launchError.value = t('chatImageStudio.launchFailedDescription')
    appStore.showError(t('chatImageStudio.launchFailedDescription'))
  } finally {
    launching.value = false
  }
}
</script>

<style scoped>
.luoye-launch-layout {
  display: grid;
  min-height: calc(100vh - 64px - 1.5rem);
  place-items: center;
  padding: 2rem;
  background: rgb(248 250 252);
}

.dark .luoye-launch-layout {
  background: rgb(15 23 42);
}

.luoye-launch-panel {
  display: flex;
  width: min(100%, 28rem);
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  text-align: center;
}

.luoye-launch-panel h1 {
  font-size: 1.125rem;
  font-weight: 800;
  color: rgb(15 23 42);
}

.luoye-launch-panel p {
  font-size: 0.875rem;
  line-height: 1.6;
  color: rgb(100 116 139);
}

.dark .luoye-launch-panel h1 {
  color: rgb(248 250 252);
}

.dark .luoye-launch-panel p {
  color: rgb(148 163 184);
}
</style>
