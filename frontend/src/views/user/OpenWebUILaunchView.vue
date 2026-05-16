<template>
  <AppLayout>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-4">
      <section class="card overflow-hidden">
        <div class="border-b border-gray-100 px-6 py-5 dark:border-dark-700">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('openWebUI.title') }}</h1>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('openWebUI.subtitle') }}</p>
            </div>
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="loadingKeys"
              :title="t('common.refresh')"
              @click="loadApiKeys"
            >
              <Icon name="refresh" size="md" :class="loadingKeys ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>

        <div class="p-6">
          <div v-if="loadingKeys" class="flex min-h-[220px] items-center justify-center text-gray-500 dark:text-dark-300">
            <Icon name="refresh" size="md" class="mr-2 animate-spin" />
            {{ t('openWebUI.loadingKeys') }}
          </div>

          <div v-else-if="usableKeys.length === 0" class="flex min-h-[220px] flex-col items-center justify-center text-center">
            <Icon name="key" size="xl" class="mb-4 text-gray-400" />
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('openWebUI.noKeysTitle') }}</h2>
            <p class="mt-2 max-w-md text-sm text-gray-500 dark:text-dark-300">{{ t('openWebUI.noKeysHint') }}</p>
            <RouterLink to="/keys" class="btn btn-primary mt-5">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('openWebUI.createKey') }}
            </RouterLink>
          </div>

          <div v-else class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
            <div class="space-y-4">
              <div>
                <label class="input-label">{{ t('openWebUI.apiKey') }}</label>
                <Select
                  v-model="selectedKeyId"
                  :options="apiKeyOptions"
                  :placeholder="t('openWebUI.selectKey')"
                  searchable="auto"
                />
              </div>

              <div class="rounded-lg border border-gray-100 bg-gray-50 px-4 py-3 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-900/50 dark:text-dark-200">
                <div class="flex gap-2">
                  <Icon name="shield" size="sm" class="mt-0.5 flex-shrink-0 text-primary-500" />
                  <span>{{ t('openWebUI.keyPrivacyHint') }}</span>
                </div>
              </div>

              <div v-if="selectedKey" class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ selectedKey.name }}</span>
                  <span class="rounded bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
                    {{ t('openWebUI.active') }}
                  </span>
                  <span
                    v-if="selectedKey.group?.platform === 'openai' && selectedKey.group?.allow_image_generation"
                    class="rounded bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                  >
                    {{ t('openWebUI.imageReady') }}
                  </span>
                </div>
                <dl class="mt-4 grid gap-3 text-sm sm:grid-cols-2">
                  <div>
                    <dt class="text-gray-500 dark:text-dark-300">{{ t('openWebUI.group') }}</dt>
                    <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ selectedKey.group?.name || '-' }}</dd>
                  </div>
                  <div>
                    <dt class="text-gray-500 dark:text-dark-300">{{ t('openWebUI.platform') }}</dt>
                    <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ selectedKey.group?.platform || '-' }}</dd>
                  </div>
                  <div>
                    <dt class="text-gray-500 dark:text-dark-300">{{ t('openWebUI.quota') }}</dt>
                    <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ quotaText(selectedKey) }}</dd>
                  </div>
                  <div>
                    <dt class="text-gray-500 dark:text-dark-300">{{ t('openWebUI.expiresAt') }}</dt>
                    <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ selectedKey.expires_at ? formatDateTime(selectedKey.expires_at) : t('openWebUI.neverExpires') }}</dd>
                  </div>
                </dl>
              </div>
            </div>

            <aside class="flex flex-col justify-between rounded-lg border border-gray-100 bg-gray-50 p-5 dark:border-dark-700 dark:bg-dark-900/50">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('openWebUI.launchPanelTitle') }}</h2>
                <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-dark-300">{{ t('openWebUI.launchPanelHint') }}</p>
              </div>

              <button
                type="button"
                class="btn btn-primary mt-6 w-full justify-center"
                :disabled="!selectedKeyId || launching"
                @click="handleLaunch"
              >
                <Icon name="externalLink" size="md" class="mr-2" />
                {{ launching ? t('openWebUI.opening') : t('openWebUI.launch') }}
              </button>
            </aside>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import { keysAPI, openWebUIAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { ApiKey } from '@/types'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const loadingKeys = ref(false)
const launching = ref(false)
const selectedKeyId = ref<number | null>(null)
const apiKeys = ref<ApiKey[]>([])

const usableKeys = computed(() =>
  apiKeys.value.filter((key) => key.status === 'active' && Boolean(key.key) && key.group_id !== null && Boolean(key.group))
)

const selectedKey = computed(() => usableKeys.value.find((key) => key.id === selectedKeyId.value) ?? null)

const apiKeyOptions = computed(() =>
  usableKeys.value.map((key) => ({
    label: `${key.name} · ${key.group?.name ?? t('keys.noGroup')}`,
    value: key.id,
  }))
)

async function loadApiKeys() {
  loadingKeys.value = true
  try {
    const result = await keysAPI.list(1, 100, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc',
    })
    apiKeys.value = result.items
    if (!selectedKeyId.value || !usableKeys.value.some((key) => key.id === selectedKeyId.value)) {
      selectedKeyId.value = preferredKeyId()
    }
  } catch (error) {
    console.error('Failed to load image workspace API keys:', error)
    appStore.showError(t('openWebUI.loadKeysFailed'))
  } finally {
    loadingKeys.value = false
  }
}

function preferredKeyId(): number | null {
  const imageReady = usableKeys.value.find((key) => key.group?.platform === 'openai' && key.group?.allow_image_generation)
  return imageReady?.id ?? usableKeys.value[0]?.id ?? null
}

async function handleLaunch() {
  if (!selectedKeyId.value || launching.value) return
  const popup = window.open('about:blank', '_blank')
  if (!popup) {
    appStore.showError(t('openWebUI.popupBlocked'))
    return
  }
  popup.opener = null
  launching.value = true
  try {
    const result = await openWebUIAPI.launch(selectedKeyId.value)
    popup.location.replace(result.launch_url)
  } catch (error) {
    console.error('Failed to launch image workspace:', error)
    popup.close()
    appStore.showError(t('openWebUI.launchFailed'))
  } finally {
    launching.value = false
  }
}

function quotaText(key: ApiKey): string {
  if (key.quota <= 0) return t('openWebUI.unlimited')
  const remaining = Math.max(0, key.quota - key.quota_used)
  return `$${remaining.toFixed(2)} / $${key.quota.toFixed(2)}`
}

onMounted(loadApiKeys)
</script>
