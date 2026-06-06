<template>
  <AppLayout>
    <div class="grid min-h-[calc(100vh-7rem)] gap-4 lg:grid-cols-[360px_minmax(0,1fr)]">
      <section class="card flex min-h-0 flex-col overflow-hidden">
        <div class="border-b border-gray-100 px-5 py-4 dark:border-dark-700">
          <div class="flex items-center gap-2">
            <h1 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('imageCreator.title') }}</h1>
            <span class="rounded bg-primary-100 px-1.5 py-0.5 text-[11px] font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
              Beta
            </span>
          </div>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('imageCreator.subtitle') }}</p>
        </div>

        <div class="custom-scrollbar flex-1 space-y-4 overflow-y-auto p-5">
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
            <div>
              <div class="mb-1.5 flex items-center justify-between gap-2">
                <label class="input-label mb-0">{{ t('imageCreator.apiKey') }}</label>
                <button
                  type="button"
                  class="text-xs font-medium text-primary-600 hover:text-primary-700 disabled:cursor-not-allowed disabled:text-gray-400 dark:text-primary-400"
                  :disabled="loadingKeys"
                  @click="loadApiKeys"
                >
                  {{ t('common.refresh') }}
                </button>
              </div>
              <Select
                v-model="selectedKeyId"
                :options="apiKeyOptions"
                :placeholder="loadingKeys ? t('imageCreator.loadingKeys') : t('imageCreator.selectKey')"
                :disabled="loadingKeys || apiKeyOptions.length === 0"
                searchable="auto"
              />
            </div>

            <div>
              <label class="input-label">{{ t('imageCreator.model') }}</label>
              <Select v-model="model" :options="modelOptions" />
            </div>
          </div>

          <div>
            <div class="mb-1.5 flex items-center justify-between gap-2">
              <label class="input-label mb-0">{{ t('imageCreator.prompt') }}</label>
              <div class="flex items-center gap-2">
                <button type="button" class="text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400" @click="applyExamplePrompt">
                  {{ t('imageCreator.example') }}
                </button>
                <button type="button" class="text-xs text-gray-500 hover:text-gray-700 dark:text-dark-300 dark:hover:text-dark-100" @click="clearPrompt">
                  {{ t('common.clear') }}
                </button>
              </div>
            </div>
            <textarea
              v-model="prompt"
              rows="7"
              class="input min-h-[172px] resize-y text-sm leading-6"
              :placeholder="t('imageCreator.promptPlaceholder')"
            ></textarea>
          </div>

          <div>
            <label class="input-label">{{ t('imageCreator.size') }}</label>
            <Select v-model="size" :options="sizeOptions" />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="input-label">{{ t('imageCreator.count') }}</label>
              <input
                v-model.number="count"
                type="number"
                min="1"
                :max="maxImageCount"
                class="input"
                @change="count = clampCount()"
                @blur="count = clampCount()"
              />
            </div>
            <div>
              <label class="input-label">{{ t('imageCreator.quality') }}</label>
              <Select v-model="quality" :options="qualityOptions" />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="input-label">{{ t('imageCreator.outputFormat') }}</label>
              <Select v-model="outputFormat" :options="outputFormatOptions" />
            </div>
            <div>
              <label class="input-label">{{ t('imageCreator.background') }}</label>
              <Select v-model="background" :options="backgroundOptions" />
            </div>
          </div>

          <div>
            <div class="mb-2 flex items-center justify-between gap-2">
              <label class="input-label mb-0">{{ t('imageCreator.referenceImage') }}</label>
              <button
                v-if="referenceImage"
                type="button"
                class="text-xs text-gray-500 hover:text-red-600 dark:text-dark-300"
                @click="clearReferenceImage"
              >
                {{ t('common.remove') }}
              </button>
            </div>
            <label
              class="flex min-h-[96px] cursor-pointer flex-col items-center justify-center rounded-xl border border-dashed border-gray-300 bg-gray-50 px-4 py-4 text-center transition-colors hover:border-primary-400 hover:bg-primary-50/40 dark:border-dark-600 dark:bg-dark-900/60 dark:hover:border-primary-500 dark:hover:bg-primary-900/10"
            >
              <input type="file" class="hidden" accept="image/png,image/jpeg,image/webp" @change="onReferenceImageChange" />
              <img
                v-if="referencePreviewUrl"
                :src="referencePreviewUrl"
                alt=""
                class="h-20 max-w-full rounded-lg object-contain"
              />
              <template v-else>
                <Icon name="upload" size="lg" class="mb-2 text-gray-400" />
                <span class="text-sm text-gray-600 dark:text-dark-200">{{ t('imageCreator.uploadReference') }}</span>
                <span class="mt-1 text-xs text-gray-400 dark:text-dark-400">{{ t('imageCreator.referenceHint') }}</span>
              </template>
            </label>
          </div>
        </div>

        <div class="border-t border-gray-100 p-4 dark:border-dark-700">
          <button
            type="button"
            class="btn btn-primary w-full justify-center"
            :disabled="!canGenerate"
            @click="handleGenerate"
          >
            <Icon name="sparkles" size="md" class="mr-2" :class="{ 'animate-pulse': generating }" />
            {{ generating ? t('imageCreator.generating') : t('imageCreator.generate') }}
          </button>
        </div>
      </section>

      <section class="card flex min-h-0 flex-col overflow-hidden">
        <div class="border-b border-gray-100 px-5 py-4 dark:border-dark-700">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('imageCreator.results') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('imageCreator.resultsHint') }}</p>
            </div>
            <button
              v-if="results.length > 0"
              type="button"
              class="btn btn-secondary btn-sm"
              @click="clearResults"
            >
              {{ t('common.clear') }}
            </button>
          </div>
        </div>

        <div class="border-b border-amber-100 bg-amber-50 px-5 py-3 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-200">
          <div class="flex gap-2">
            <Icon name="exclamationTriangle" size="sm" class="mt-0.5 flex-shrink-0" />
            <span>{{ t('imageCreator.safetyNotice') }}</span>
          </div>
        </div>

        <div class="custom-scrollbar flex-1 overflow-y-auto p-5">
          <div v-if="generating" class="flex min-h-[420px] flex-col items-center justify-center text-center">
            <div class="relative mb-5 flex h-24 w-24 items-center justify-center">
              <div class="image-wait-ring"></div>
              <div class="image-wait-ring image-wait-ring-delay"></div>
              <div class="relative z-10 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-50 text-primary-600 shadow-sm dark:bg-primary-900/20 dark:text-primary-300">
                <Icon name="sparkles" size="xl" class="animate-pulse" />
              </div>
            </div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('imageCreator.generatingTitle') }}</h3>
            <p class="mt-2 max-w-sm text-sm text-gray-500 dark:text-dark-300">{{ t('imageCreator.generatingHint') }}</p>
            <p class="mt-3 text-sm font-medium text-primary-600 dark:text-primary-300">{{ waitingStepText }}</p>
            <div class="mt-5 w-full max-w-sm">
              <div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-800">
                <div class="image-wait-bar h-full rounded-full bg-primary-500"></div>
              </div>
              <div class="mt-2 flex items-center justify-between gap-3 text-xs text-gray-500 dark:text-dark-300">
                <span>{{ elapsedText }}</span>
                <span>{{ t('imageCreator.storageLocalOnly') }}</span>
              </div>
            </div>
            <div class="mt-5 flex items-center justify-center gap-1.5" aria-hidden="true">
              <span class="image-wait-dot"></span>
              <span class="image-wait-dot image-wait-dot-delay-1"></span>
              <span class="image-wait-dot image-wait-dot-delay-2"></span>
            </div>
          </div>

          <div v-else-if="results.length === 0" class="flex min-h-[420px] items-center justify-center">
            <EmptyState :title="t('imageCreator.emptyTitle')" :description="t('imageCreator.emptyDescription')">
              <template #icon>
                <Icon name="sparkles" size="xl" class="text-primary-500" />
              </template>
            </EmptyState>
          </div>

          <div v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            <article
              v-for="(item, index) in results"
              :key="item.id"
              class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800"
            >
              <div class="flex aspect-square items-center justify-center bg-gray-100 dark:bg-dark-900">
                <button
                  type="button"
                  class="group relative flex h-full w-full cursor-zoom-in items-center justify-center overflow-hidden"
                  :aria-label="t('imageCreator.previewImage')"
                  :title="t('imageCreator.previewImage')"
                  data-testid="image-result-preview"
                  @click="openPreview(item)"
                  @dblclick="openPreview(item)"
                >
                  <img :src="item.url" alt="" class="h-full w-full object-contain transition-transform duration-200 group-hover:scale-[1.02] group-focus-visible:scale-[1.02]" />
                  <span class="pointer-events-none absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-full bg-black/55 text-white opacity-0 shadow-lg backdrop-blur-sm transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100">
                    <Icon name="search" size="sm" />
                  </span>
                </button>
              </div>
              <div class="space-y-3 p-3">
                <p v-if="item.revisedPrompt" class="line-clamp-2 text-xs leading-5 text-gray-500 dark:text-dark-300">
                  {{ item.revisedPrompt }}
                </p>
                <div class="flex items-center justify-between gap-2">
                  <span class="text-xs text-gray-400 dark:text-dark-400">
                    {{ item.outputFormat.toUpperCase() }}
                  </span>
                  <button type="button" class="btn btn-secondary btn-sm" @click="downloadImage(item, index)">
                    <Icon name="download" size="sm" class="mr-1.5" />
                    {{ t('imageCreator.download') }}
                  </button>
                </div>
              </div>
            </article>
          </div>
        </div>
      </section>
    </div>

    <div
      v-if="previewImage"
      class="fixed inset-0 z-[70] flex flex-col bg-black/85 p-4 backdrop-blur-sm sm:p-6"
      data-testid="image-preview-overlay"
      @click.self="closePreview"
    >
      <div class="mb-3 flex items-center justify-end gap-2">
        <button type="button" class="btn btn-secondary btn-sm bg-white/95 text-gray-800 hover:bg-white dark:bg-dark-800/95 dark:text-dark-100" @click.stop="downloadPreviewImage">
          <Icon name="download" size="sm" class="sm:mr-1.5" />
          <span class="sr-only sm:not-sr-only">{{ t('imageCreator.download') }}</span>
        </button>
        <button
          type="button"
          class="flex h-9 w-9 items-center justify-center rounded-lg bg-white/95 text-gray-700 shadow-sm transition-colors hover:bg-white hover:text-gray-900 dark:bg-dark-800/95 dark:text-dark-100 dark:hover:bg-dark-700"
          :aria-label="t('imageCreator.closePreview')"
          data-testid="image-preview-close"
          @click.stop="closePreview"
        >
          <Icon name="x" size="md" />
        </button>
      </div>

      <div class="flex min-h-0 flex-1 items-center justify-center">
        <img :src="previewImage.url" alt="" class="max-h-full max-w-full rounded-lg object-contain shadow-2xl" />
      </div>

      <p v-if="previewImage.revisedPrompt" class="mx-auto mt-3 max-w-4xl text-center text-xs leading-5 text-white/75 sm:text-sm">
        {{ previewImage.revisedPrompt }}
      </p>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI } from '@/api'
import {
  createImageTask,
  downloadImageFile,
  getImageTask,
  listImageTasks,
  type ImageCreatorOutputFormat,
  type ImageCreatorStoredImage,
  type ImageCreatorTask,
} from '@/api/imageCreator'
import { useAppStore } from '@/stores'
import type { ApiKey } from '@/types'
import { apiKeySupportsOpenAI, apiKeySupportsOpenAIImageGeneration, primaryAPIKeyImageGroupName } from '@/utils/apiKeyCapabilities'

interface GeneratedImage {
  id: string
  url: string
  sourceUrl: string
  revisedPrompt: string
  outputFormat: ImageCreatorOutputFormat | string
  mimeType: string
}

const { t } = useI18n()
const appStore = useAppStore()

const loadingKeys = ref(false)
const generating = ref(false)
const apiKeys = ref<ApiKey[]>([])
const selectedKeyId = ref<number | null>(null)
const model = ref('gpt-image-2')
const prompt = ref('')
const size = ref('auto')
const count = ref(1)
const maxImageCount = 8
const quality = ref('auto')
const outputFormat = ref<ImageCreatorOutputFormat>('webp')
const background = ref('auto')
const referenceImage = ref<File | null>(null)
const referencePreviewUrl = ref('')
const results = ref<GeneratedImage[]>([])
const previewImage = ref<GeneratedImage | null>(null)
const elapsedSeconds = ref(0)
const waitingStepIndex = ref(0)
const activeTaskId = ref<number | null>(null)
let generationTimerId: ReturnType<typeof setInterval> | null = null
let taskPollTimerId: ReturnType<typeof setInterval> | null = null
let imagePreviewLoadToken = 0
const generatedImageObjectUrls = new Set<string>()

const modelOptions = [
  { value: 'gpt-image-2', label: 'gpt-image-2' },
  { value: 'gpt-image-1.5', label: 'gpt-image-1.5' },
  { value: 'gpt-image-1', label: 'gpt-image-1' },
]

const sizeOptions = [
  { value: 'auto', label: 'Auto' },
  { value: '1024x1024', label: '1:1 1024x1024' },
  { value: '1536x1024', label: '3:2 1536x1024' },
  { value: '1024x1536', label: '2:3 1024x1536' },
  { value: '2048x2048', label: '1:1 2048x2048' },
  { value: '3840x2160', label: '16:9 3840x2160' },
  { value: '2160x3840', label: '9:16 2160x3840' },
]

const qualityOptions = [
  { value: 'auto', label: 'auto' },
  { value: 'low', label: 'low' },
  { value: 'medium', label: 'medium' },
  { value: 'high', label: 'high' },
]

const outputFormatOptions = [
  { value: 'webp', label: 'WEBP' },
  { value: 'jpeg', label: 'JPEG' },
]

const transparentUnsupportedImageModels = new Set(['gpt-image-1.5'])

const allBackgroundOptions = [
  { value: 'auto', label: 'auto' },
  { value: 'transparent', label: 'transparent' },
  { value: 'opaque', label: 'opaque' },
]

const backgroundOptions = computed(() => {
  if (!modelSupportsTransparentBackground(model.value)) {
    return allBackgroundOptions.filter((option) => option.value !== 'transparent')
  }
  return allBackgroundOptions
})

const waitingStepKeys = [
  'imageCreator.waitingSteps.routing',
  'imageCreator.waitingSteps.rendering',
  'imageCreator.waitingSteps.receiving',
  'imageCreator.waitingSteps.finishing',
]

const examplePrompts = [
  'A polished SaaS dashboard hero image, clean interface, soft daylight, realistic product screenshot style',
  '一张赛博朋克风格的雨夜城市街景，霓虹灯反射在路面，电影感构图',
  'A cute orange cat astronaut sticker, crisp vector-like edges, clean pastel background',
]

const apiKeyOptions = computed(() =>
  apiKeys.value.map((key) => ({
    value: key.id,
    label: [
      key.name,
      primaryAPIKeyImageGroupName(key),
      apiKeySupportsOpenAIImageGeneration(key) ? t('imageCreator.imageEnabled') : t('imageCreator.imageDisabled'),
    ].filter(Boolean).join(' · '),
    disabled: !isOpenAIImageKey(key),
  }))
)

const selectedKey = computed(() => apiKeys.value.find((key) => key.id === selectedKeyId.value) ?? null)

const canGenerate = computed(() => {
  return !generating.value && !!selectedKey.value && isOpenAIImageKey(selectedKey.value) && prompt.value.trim().length > 0 && count.value >= 1
})

const waitingStepText = computed(() => t(waitingStepKeys[waitingStepIndex.value]))

const elapsedText = computed(() => t('imageCreator.elapsedSeconds', { seconds: elapsedSeconds.value }))

async function loadApiKeys(): Promise<void> {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc',
    })
    apiKeys.value = response.items
      .filter(isOpenAIKey)
      .sort((a, b) => Number(isOpenAIImageKey(b)) - Number(isOpenAIImageKey(a)))
    selectedKeyId.value = pickDefaultApiKey(apiKeys.value)?.id ?? null
  } catch (error) {
    appStore.showError(t('imageCreator.loadKeysFailed'))
  } finally {
    loadingKeys.value = false
  }
}

function applyExamplePrompt(): void {
  const index = Math.floor(Math.random() * examplePrompts.length)
  prompt.value = examplePrompts[index]
}

function clearPrompt(): void {
  prompt.value = ''
}

function clampCount(): number {
  const n = Number(count.value)
  if (!Number.isFinite(n)) return 1
  return Math.min(Math.max(Math.trunc(n), 1), maxImageCount)
}

function modelSupportsTransparentBackground(value: string): boolean {
  return !transparentUnsupportedImageModels.has(value.trim().toLowerCase())
}

function normalizeBackgroundForModel(value: string, modelValue = model.value): string {
  if (value.trim().toLowerCase() === 'transparent' && !modelSupportsTransparentBackground(modelValue)) {
    return 'auto'
  }
  return value.trim()
}

watch(model, () => {
  background.value = normalizeBackgroundForModel(background.value)
})

function startGenerationTimer(startedAtMs = Date.now()): void {
  stopGenerationTimer()
  const startedAt = Number.isFinite(startedAtMs) ? startedAtMs : Date.now()
  elapsedSeconds.value = Math.max(0, Math.floor((Date.now() - startedAt) / 1000))
  waitingStepIndex.value = 0
  generationTimerId = setInterval(() => {
    elapsedSeconds.value = Math.max(0, Math.floor((Date.now() - startedAt) / 1000))
    waitingStepIndex.value = Math.floor(elapsedSeconds.value / 8) % waitingStepKeys.length
  }, 1000)
}

function stopGenerationTimer(): void {
  if (!generationTimerId) return
  clearInterval(generationTimerId)
  generationTimerId = null
}

function stopTaskPolling(): void {
  if (!taskPollTimerId) return
  clearInterval(taskPollTimerId)
  taskPollTimerId = null
}

function taskIsActive(task: ImageCreatorTask | null | undefined): boolean {
  return task?.status === 'pending' || task?.status === 'running'
}

function storedImageToResult(image: ImageCreatorStoredImage, index: number): GeneratedImage {
  return {
    id: String(image.id || `${Date.now()}-${index}`),
    url: image.url,
    sourceUrl: image.url,
    revisedPrompt: image.revised_prompt || '',
    outputFormat: image.output_format || outputFormat.value,
    mimeType: image.mime_type || '',
  }
}

function imagesFromTasks(tasks: ImageCreatorTask[]): ImageCreatorStoredImage[] {
  return tasks.flatMap((task) => task.images || [])
}

function applyStoredImages(images: ImageCreatorStoredImage[]): void {
  imagePreviewLoadToken += 1
  revokeGeneratedImageObjectUrls()
  results.value = images
    .filter((image) => typeof image.url === 'string' && image.url.trim().length > 0)
    .map(storedImageToResult)
  void hydrateGeneratedImagePreviews(imagePreviewLoadToken)
}

function revokeGeneratedImageObjectUrls(): void {
  for (const url of generatedImageObjectUrls) {
    URL.revokeObjectURL(url)
  }
  generatedImageObjectUrls.clear()
}

function revokeGeneratedImageObjectUrl(url: string): void {
  if (!url.startsWith('blob:')) return
  if (!generatedImageObjectUrls.delete(url)) return
  URL.revokeObjectURL(url)
}

function shouldFetchImageUrl(url: string): boolean {
  const value = url.trim().toLowerCase()
  return value !== '' && !value.startsWith('data:') && !value.startsWith('blob:')
}

async function createObjectUrlForImage(item: GeneratedImage): Promise<string> {
  const sourceUrl = item.sourceUrl || item.url
  if (!shouldFetchImageUrl(sourceUrl)) {
    return sourceUrl
  }
  const blob = await downloadImageFile(sourceUrl)
  const objectUrl = URL.createObjectURL(blob)
  generatedImageObjectUrls.add(objectUrl)
  return objectUrl
}

async function ensureImageDisplayUrl(item: GeneratedImage): Promise<string> {
  if (!shouldFetchImageUrl(item.url)) {
    return item.url
  }
  const objectUrl = await createObjectUrlForImage(item)
  const current = results.value.find((result) => result.id === item.id)
  if (current) {
    revokeGeneratedImageObjectUrl(current.url)
    current.url = objectUrl
  }
  if (previewImage.value?.id === item.id) {
    previewImage.value.url = objectUrl
  }
  return objectUrl
}

async function hydrateGeneratedImagePreviews(token: number): Promise<void> {
  const items = results.value.slice()
  await Promise.all(items.map(async (item) => {
    if (!shouldFetchImageUrl(item.url)) return
    try {
      const objectUrl = await createObjectUrlForImage(item)
      if (token !== imagePreviewLoadToken) {
        revokeGeneratedImageObjectUrl(objectUrl)
        return
      }
      const current = results.value.find((result) => result.id === item.id)
      if (!current) {
        revokeGeneratedImageObjectUrl(objectUrl)
        return
      }
      revokeGeneratedImageObjectUrl(current.url)
      current.url = objectUrl
      if (previewImage.value?.id === current.id) {
        previewImage.value.url = objectUrl
      }
    } catch {
      // Keep the original URL as a fallback for cookie-based sessions.
    }
  }))
}

function latestActiveTask(tasks: ImageCreatorTask[]): ImageCreatorTask | null {
  return tasks.find(taskIsActive) ?? null
}

function taskTimerStartMs(task: ImageCreatorTask | null | undefined): number {
  const raw = task?.started_at || task?.created_at
  const parsed = raw ? Date.parse(raw) : NaN
  return Number.isFinite(parsed) ? parsed : Date.now()
}

async function loadImageCreatorTasks(): Promise<void> {
  try {
    const response = await listImageTasks()
    applyStoredImages(response.images?.length ? response.images : imagesFromTasks(response.tasks || []))
    const active = latestActiveTask(response.tasks || [])
    if (active) {
      startTaskPolling(active.id, taskTimerStartMs(active))
    }
  } catch (error: any) {
    appStore.showError(error?.message || t('imageCreator.loadTasksFailed'))
  }
}

function startTaskPolling(taskId: number, startedAtMs = Date.now()): void {
  activeTaskId.value = taskId
  generating.value = true
  startGenerationTimer(startedAtMs)
  stopTaskPolling()
  void pollImageTask(taskId)
  taskPollTimerId = setInterval(() => {
    void pollImageTask(taskId)
  }, 2500)
}

async function pollImageTask(taskId: number): Promise<void> {
  try {
    const task = await getImageTask(taskId)
    if (activeTaskId.value !== taskId) return
    if (taskIsActive(task)) {
      return
    }
    stopTaskPolling()
    stopGenerationTimer()
    generating.value = false
    activeTaskId.value = null
    if (task.status === 'succeeded') {
      applyStoredImages(task.images || [])
      appStore.showSuccess(t('imageCreator.generateSuccess', { count: task.images?.length || 0 }))
      return
    }
    appStore.showError(task.error_message || t('imageCreator.generateFailed'))
  } catch (error: any) {
    stopTaskPolling()
    stopGenerationTimer()
    generating.value = false
    activeTaskId.value = null
    appStore.showError(error?.message || t('imageCreator.generateFailed'))
  }
}

function isOpenAIImageKey(key: ApiKey): boolean {
  return apiKeySupportsOpenAIImageGeneration(key)
}

function isOpenAIKey(key: ApiKey): boolean {
  return apiKeySupportsOpenAI(key)
}

function pickDefaultApiKey(keys: ApiKey[]): ApiKey | null {
  const current = keys.find((key) => key.id === selectedKeyId.value)
  if (current && isOpenAIImageKey(current)) return current

  return keys.find(isOpenAIImageKey) ?? null
}

async function handleGenerate(): Promise<void> {
  if (!selectedKey.value || !isOpenAIImageKey(selectedKey.value)) {
    appStore.showError(t('imageCreator.selectKeyFirst'))
    return
  }
  if (!prompt.value.trim()) {
    appStore.showError(t('imageCreator.promptRequired'))
    return
  }

  generating.value = true
  startGenerationTimer()
  try {
    count.value = clampCount()
    background.value = normalizeBackgroundForModel(background.value)
    const task = await createImageTask({
      apiKeyId: selectedKey.value.id,
      model: model.value,
      prompt: prompt.value.trim(),
      size: size.value,
      quality: quality.value,
      count: count.value,
      outputFormat: outputFormat.value,
      background: background.value,
      referenceImage: referenceImage.value,
    })
    startTaskPolling(task.id)
  } catch (error: any) {
    stopGenerationTimer()
    generating.value = false
    appStore.showError(error?.message || t('imageCreator.generateFailed'))
  }
}

function onReferenceImageChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  if (!file) return
  clearReferenceImage()
  referenceImage.value = file
  referencePreviewUrl.value = URL.createObjectURL(file)
  input.value = ''
}

function clearReferenceImage(): void {
  referenceImage.value = null
  if (referencePreviewUrl.value) {
    URL.revokeObjectURL(referencePreviewUrl.value)
    referencePreviewUrl.value = ''
  }
}

function clearResults(): void {
  imagePreviewLoadToken += 1
  revokeGeneratedImageObjectUrls()
  results.value = []
  closePreview()
}

async function downloadImage(item: GeneratedImage, index: number): Promise<void> {
  const href = await ensureImageDisplayUrl(item)
  const link = document.createElement('a')
  link.href = href
  link.download = `image-${Date.now()}-${index + 1}.${String(item.outputFormat || outputFormat.value).toLowerCase()}`
  if (!href.startsWith('data:') && !href.startsWith('blob:')) {
    link.target = '_blank'
    link.rel = 'noopener'
  }
  document.body.appendChild(link)
  link.click()
  link.remove()
}

function openPreview(item: GeneratedImage): void {
  previewImage.value = item
}

function closePreview(): void {
  previewImage.value = null
}

function downloadPreviewImage(): void {
  if (!previewImage.value) return
  const index = results.value.findIndex((item) => item.id === previewImage.value?.id)
  void downloadImage(previewImage.value, index >= 0 ? index : 0)
}

function onPreviewKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    closePreview()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onPreviewKeydown)
  loadApiKeys()
  loadImageCreatorTasks()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onPreviewKeydown)
  stopTaskPolling()
  stopGenerationTimer()
  imagePreviewLoadToken += 1
  revokeGeneratedImageObjectUrls()
  clearReferenceImage()
})
</script>

<style scoped>
.image-wait-ring {
  position: absolute;
  inset: 0;
  border-radius: 9999px;
  border: 1px solid rgb(20 184 166 / 0.2);
  background:
    radial-gradient(circle at center, rgb(20 184 166 / 0.12), transparent 58%),
    conic-gradient(from 90deg, rgb(20 184 166 / 0), rgb(20 184 166 / 0.55), rgb(20 184 166 / 0));
  animation: image-wait-spin 2.4s linear infinite;
}

.image-wait-ring-delay {
  inset: 10px;
  animation-direction: reverse;
  animation-duration: 3.2s;
  opacity: 0.65;
}

.image-wait-bar {
  width: 45%;
  animation: image-wait-bar 1.6s ease-in-out infinite;
}

.image-wait-dot {
  height: 0.45rem;
  width: 0.45rem;
  border-radius: 9999px;
  background: rgb(20 184 166 / 0.75);
  animation: image-wait-dot 1.2s ease-in-out infinite;
}

.image-wait-dot-delay-1 {
  animation-delay: 0.16s;
}

.image-wait-dot-delay-2 {
  animation-delay: 0.32s;
}

@keyframes image-wait-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes image-wait-bar {
  0% {
    transform: translateX(-110%);
  }
  55% {
    transform: translateX(85%);
  }
  100% {
    transform: translateX(230%);
  }
}

@keyframes image-wait-dot {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.45;
  }
  40% {
    transform: translateY(-0.35rem);
    opacity: 1;
  }
}
</style>
