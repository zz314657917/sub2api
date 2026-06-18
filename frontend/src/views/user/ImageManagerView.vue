<template>
  <AppLayout>
    <div class="image-manager-page" data-testid="image-manager-view">
      <header class="image-manager-header">
        <div class="min-w-0">
          <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('imageManager.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('imageManager.subtitle') }}</p>
        </div>
        <div class="image-manager-actions">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadImages">
            <Icon name="refresh" size="sm" />
            <span>{{ t('common.refresh') }}</span>
          </button>
          <a href="/chat-images" target="_blank" rel="noopener noreferrer" class="btn btn-primary btn-sm">
            <Icon name="sparkles" size="sm" />
            <span>{{ t('imageManager.openStudio') }}</span>
          </a>
        </div>
      </header>

      <section class="image-manager-toolbar">
        <div class="text-sm text-gray-500 dark:text-dark-300">
          {{ t('imageManager.totalImages', { count: total }) }}
        </div>
        <div v-if="selectedIds.length > 0" class="image-manager-selection">
          <span>{{ t('imageManager.selectedImages', { count: selectedIds.length }) }}</span>
          <button type="button" class="image-manager-text-button" @click="downloadSelected">
            <Icon name="download" size="xs" />
            <span>{{ t('imageManager.downloadSelected') }}</span>
          </button>
          <button type="button" class="image-manager-text-button danger" :disabled="deleting" @click="deleteSelected">
            <Icon name="trash" size="xs" />
            <span>{{ t('imageManager.deleteSelected') }}</span>
          </button>
          <button type="button" class="image-manager-text-button" @click="clearSelection">
            {{ t('imageManager.clearSelection') }}
          </button>
        </div>
      </section>

      <section class="image-manager-filters" data-testid="image-manager-filters">
        <label class="image-manager-field image-manager-field-wide">
          <span>{{ t('imageManager.search') }}</span>
          <input v-model.trim="filters.q" type="search" class="input" :placeholder="t('imageManager.searchPlaceholder')" @keyup.enter="applyFilters" />
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.startDate') }}</span>
          <input v-model="filters.start_date" type="date" class="input" />
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.endDate') }}</span>
          <input v-model="filters.end_date" type="date" class="input" />
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.format') }}</span>
          <select v-model="filters.format" class="input">
            <option value="">{{ t('imageManager.allFormats') }}</option>
            <option value="png">PNG</option>
            <option value="jpeg">JPG</option>
            <option value="webp">WEBP</option>
            <option value="other">{{ t('imageManager.other') }}</option>
          </select>
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.orientation') }}</span>
          <select v-model="filters.orientation" class="input">
            <option value="">{{ t('imageManager.allOrientations') }}</option>
            <option value="landscape">{{ t('imageManager.landscape') }}</option>
            <option value="portrait">{{ t('imageManager.portrait') }}</option>
            <option value="square">{{ t('imageManager.square') }}</option>
            <option value="unknown">{{ t('imageManager.unknownSize') }}</option>
          </select>
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.resolution') }}</span>
          <select v-model="filters.resolution" class="input">
            <option value="">{{ t('imageManager.allResolutions') }}</option>
            <option value="1080p">1080P</option>
            <option value="2k">2K</option>
            <option value="4k">4K</option>
            <option value="unknown">{{ t('imageManager.unknownSize') }}</option>
          </select>
        </label>
        <label class="image-manager-field">
          <span>{{ t('imageManager.aspectRatio') }}</span>
          <select v-model="filters.aspect_ratio" class="input">
            <option value="">{{ t('imageManager.allAspectRatios') }}</option>
            <option value="1:1">1:1</option>
            <option value="4:3">4:3</option>
            <option value="3:4">3:4</option>
            <option value="16:9">16:9</option>
            <option value="9:16">9:16</option>
            <option value="other">{{ t('imageManager.other') }}</option>
            <option value="unknown">{{ t('imageManager.unknownSize') }}</option>
          </select>
        </label>
        <div class="image-manager-filter-actions">
          <button type="button" class="btn btn-primary btn-sm" :disabled="loading" @click="applyFilters">
            <Icon name="search" size="sm" />
            <span>{{ t('imageManager.applyFilters') }}</span>
          </button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loading || !hasActiveFilters" @click="resetFilters">
            {{ t('imageManager.resetFilters') }}
          </button>
        </div>
      </section>

      <section v-if="loading && images.length === 0" class="image-manager-state">
        <Icon name="sync" size="xl" class="animate-spin text-primary-500" />
        <span>{{ t('imageManager.loading') }}</span>
      </section>

      <section v-else-if="images.length === 0" class="image-manager-state">
        <Icon name="image" size="xl" class="text-primary-500" />
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('imageManager.emptyTitle') }}</h2>
        <p class="max-w-md text-center text-sm text-gray-500 dark:text-dark-300">{{ t('imageManager.emptyDescription') }}</p>
        <a href="/chat-images" target="_blank" rel="noopener noreferrer" class="btn btn-primary btn-sm">
          <Icon name="sparkles" size="sm" />
          <span>{{ t('imageManager.openStudio') }}</span>
        </a>
      </section>

      <section v-else class="image-manager-grid">
        <article
          v-for="(item, index) in images"
          :key="item.id"
          class="image-manager-card"
          :class="{ selected: isSelected(item.id) }"
          data-testid="image-manager-card"
        >
          <button
            type="button"
            class="image-manager-select"
            :class="{ active: isSelected(item.id) }"
            :aria-label="isSelected(item.id) ? t('imageManager.deselectImage') : t('imageManager.selectImage')"
            @click.stop="toggleImage(item.id)"
          >
            <Icon v-if="isSelected(item.id)" name="check" size="xs" />
          </button>

          <button type="button" class="image-manager-preview" :aria-label="t('imageManager.previewImage')" @click="openPreview(item)">
            <img :src="displayUrl(item)" alt="" loading="lazy" />
          </button>

          <div class="image-manager-card-body">
            <div class="flex items-center justify-between gap-2">
              <span class="image-manager-format">{{ String(item.output_format || 'image').toUpperCase() }}</span>
              <span class="text-xs text-gray-400 dark:text-dark-400">{{ formatFileSize(item.byte_size) }}</span>
            </div>
            <p class="image-manager-prompt">{{ item.task_prompt || item.revised_prompt || t('imageManager.noPrompt') }}</p>
            <div class="image-manager-meta">
              <span>{{ displayModelLabel(item.task_model || 'gpt-image-2') }}</span>
              <span v-if="dimensionSummary(item)">{{ dimensionSummary(item) }}</span>
              <span>{{ formatTime(item.created_at) }}</span>
            </div>
            <div class="image-manager-card-actions">
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.download')" @click="downloadImage(item, index)">
                <Icon name="download" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.copyPrompt')" @click="copyPrompt(item)">
                <Icon name="copy" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.reusePrompt')" @click="reusePrompt(item)">
                <Icon name="sparkles" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button" :title="t('imageManager.useAsReference')" @click="useAsReference(item)">
                <Icon name="image" size="sm" />
              </button>
              <button type="button" class="image-manager-icon-button danger" :title="t('imageManager.delete')" @click="deleteOne(item)">
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
        </article>
      </section>

      <footer v-if="hasMore" class="image-manager-footer">
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadMore">
          {{ loading ? t('imageManager.loading') : t('imageManager.loadMore') }}
        </button>
      </footer>

      <div
        v-if="previewImage"
        class="image-manager-lightbox"
        data-testid="image-manager-lightbox"
        @click.self="closePreview"
      >
        <div class="mb-3 flex items-center justify-end gap-2">
          <button type="button" class="btn btn-secondary btn-sm bg-white/95 text-gray-800 hover:bg-white dark:bg-dark-800/95 dark:text-dark-100" @click.stop="downloadPreview">
            <Icon name="download" size="sm" />
            <span class="sr-only sm:not-sr-only">{{ t('imageManager.download') }}</span>
          </button>
          <button type="button" class="image-manager-lightbox-close" :aria-label="t('imageManager.closePreview')" @click.stop="closePreview">
            <Icon name="x" size="md" />
          </button>
        </div>
        <div class="flex min-h-0 flex-1 items-center justify-center">
          <img :src="displayUrl(previewImage)" alt="" class="max-h-full max-w-full rounded-lg object-contain shadow-2xl" />
        </div>
        <p v-if="previewPrompt" class="mx-auto mt-3 max-w-4xl text-center text-xs leading-5 text-white/75 sm:text-sm">
          {{ previewPrompt }}
        </p>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  deleteManagedImages,
  downloadImageFile,
  listManagedImages,
  type ImageCreatorImageListParams,
  type ImageCreatorManagedImage,
} from '@/api/imageCreator'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { displayModelLabel } from '@/utils/modelDisplay'

const PAGE_SIZE = 40

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const images = ref<ImageCreatorManagedImage[]>([])
const total = ref(0)
const loading = ref(false)
const deleting = ref(false)
const selectedIds = ref<number[]>([])
const previewImage = ref<ImageCreatorManagedImage | null>(null)
const imageDisplayUrls = ref<Record<number, string>>({})
const objectUrls = new Set<string>()
const filters = reactive({
  q: '',
  start_date: '',
  end_date: '',
  format: '',
  orientation: '',
  resolution: '',
  aspect_ratio: '',
})

const hasMore = computed(() => images.value.length < total.value)
const previewPrompt = computed(() => previewImage.value?.task_prompt || previewImage.value?.revised_prompt || '')
const hasActiveFilters = computed(() => Object.values(filters).some((value) => String(value || '').trim() !== ''))

onMounted(() => {
  void loadImages()
})

onUnmounted(() => {
  revokeObjectUrls()
})

async function loadImages(): Promise<void> {
  loading.value = true
  try {
    const response = await listManagedImages(buildListParams(0))
    images.value = response.items || []
    total.value = response.total || images.value.length
    selectedIds.value = selectedIds.value.filter((id) => images.value.some((image) => image.id === id))
    void hydrateImages(images.value)
  } catch (error: any) {
    appStore.showError(error?.message || t('imageManager.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function loadMore(): Promise<void> {
  if (loading.value || !hasMore.value) return
  loading.value = true
  try {
    const response = await listManagedImages(buildListParams(images.value.length))
    const nextImages = response.items || []
    images.value = [...images.value, ...nextImages]
    total.value = response.total || images.value.length
    void hydrateImages(nextImages)
  } catch (error: any) {
    appStore.showError(error?.message || t('imageManager.loadFailed'))
  } finally {
    loading.value = false
  }
}

function buildListParams(offset: number): ImageCreatorImageListParams {
  const params: ImageCreatorImageListParams = {
    limit: PAGE_SIZE,
    offset,
  }
  for (const [key, value] of Object.entries(filters)) {
    const text = String(value || '').trim()
    if (text) {
      const writableParams = params as Record<string, string | number>
      writableParams[key] = text
    }
  }
  return params
}

function applyFilters(): void {
  selectedIds.value = []
  void loadImages()
}

function resetFilters(): void {
  filters.q = ''
  filters.start_date = ''
  filters.end_date = ''
  filters.format = ''
  filters.orientation = ''
  filters.resolution = ''
  filters.aspect_ratio = ''
  applyFilters()
}

function isSelected(id: number): boolean {
  return selectedIds.value.includes(id)
}

function toggleImage(id: number): void {
  selectedIds.value = isSelected(id)
    ? selectedIds.value.filter((item) => item !== id)
    : [...selectedIds.value, id]
}

function clearSelection(): void {
  selectedIds.value = []
}

async function deleteOne(item: ImageCreatorManagedImage): Promise<void> {
  await deleteImages([item.id])
}

async function deleteSelected(): Promise<void> {
  await deleteImages(selectedIds.value)
}

async function deleteImages(ids: number[]): Promise<void> {
  const normalized = Array.from(new Set(ids.filter((id) => Number.isFinite(id) && id > 0)))
  if (normalized.length === 0 || deleting.value) return
  deleting.value = true
  try {
    const result = await deleteManagedImages(normalized)
    const deleted = result.deleted || normalized.length
    images.value = images.value.filter((image) => !normalized.includes(image.id))
    total.value = Math.max(0, total.value - deleted)
    selectedIds.value = selectedIds.value.filter((id) => !normalized.includes(id))
    if (previewImage.value && normalized.includes(previewImage.value.id)) {
      previewImage.value = null
    }
    appStore.showSuccess(t('imageManager.deleteSuccess', { count: deleted }))
  } catch (error: any) {
    appStore.showError(error?.message || t('imageManager.deleteFailed'))
  } finally {
    deleting.value = false
  }
}

async function hydrateImages(items: ImageCreatorManagedImage[]): Promise<void> {
  await Promise.all(items.map(async (item) => {
    if (!shouldFetchImageUrl(item.url) || imageDisplayUrls.value[item.id]) return
    try {
      const blob = await downloadImageFile(item.url)
      const objectUrl = URL.createObjectURL(blob)
      objectUrls.add(objectUrl)
      imageDisplayUrls.value = { ...imageDisplayUrls.value, [item.id]: objectUrl }
    } catch {
      imageDisplayUrls.value = { ...imageDisplayUrls.value, [item.id]: item.url }
    }
  }))
}

function shouldFetchImageUrl(url: string): boolean {
  return typeof url === 'string' && url.startsWith('/api/')
}

function displayUrl(item: ImageCreatorManagedImage): string {
  return imageDisplayUrls.value[item.id] || item.url
}

async function ensureDisplayUrl(item: ImageCreatorManagedImage): Promise<string> {
  if (!imageDisplayUrls.value[item.id]) {
    await hydrateImages([item])
  }
  return displayUrl(item)
}

async function downloadImage(item: ImageCreatorManagedImage, index: number): Promise<void> {
  const href = await ensureDisplayUrl(item)
  const link = document.createElement('a')
  link.href = href
  link.download = buildDownloadName(item, index)
  document.body.appendChild(link)
  link.click()
  link.remove()
}

async function downloadSelected(): Promise<void> {
  const selected = images.value
    .map((image, index) => ({ image, index }))
    .filter(({ image }) => selectedIds.value.includes(image.id))
  for (const { image, index } of selected) {
    await downloadImage(image, index)
  }
}

function downloadPreview(): void {
  if (!previewImage.value) return
  const index = images.value.findIndex((item) => item.id === previewImage.value?.id)
  void downloadImage(previewImage.value, index >= 0 ? index : 0)
}

function buildDownloadName(item: ImageCreatorManagedImage, index: number): string {
  const ext = String(item.output_format || 'png').toLowerCase() === 'jpeg' ? 'jpg' : String(item.output_format || 'png').toLowerCase()
  return `sub2api-image-${String(index + 1).padStart(2, '0')}.${ext}`
}

function copyPrompt(item: ImageCreatorManagedImage): void {
  void copyToClipboard(item.task_prompt || item.revised_prompt || '')
}

function openChatImagesInNewTab(query: Record<string, string>): void {
  const resolved = router.resolve({
    path: '/chat-images',
    query,
  })
  window.open(resolved.href, '_blank', 'noopener,noreferrer')
}

function reusePrompt(item: ImageCreatorManagedImage): void {
  const prompt = item.task_prompt || item.revised_prompt || ''
  openChatImagesInNewTab(prompt ? { prompt, mode: 'image' } : { mode: 'image' })
}

function useAsReference(item: ImageCreatorManagedImage): void {
  const prompt = item.task_prompt || item.revised_prompt || ''
  openChatImagesInNewTab({
    mode: 'image',
    reference_image_id: String(item.id),
    ...(prompt ? { prompt } : {}),
  })
}

function openPreview(item: ImageCreatorManagedImage): void {
  previewImage.value = item
  void hydrateImages([item])
}

function closePreview(): void {
  previewImage.value = null
}

function formatTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatFileSize(value: number): string {
  const size = Number(value)
  if (!Number.isFinite(size) || size <= 0) return ''
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function dimensionSummary(item: ImageCreatorManagedImage): string {
  const width = Number(item.width)
  const height = Number(item.height)
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) {
    return ''
  }
  return [item.resolution || `${width}x${height}`, item.aspect_ratio || '', formatMegapixels(width, height)]
    .filter(Boolean)
    .join(' · ')
}

function formatMegapixels(width: number, height: number): string {
  const mp = width * height / 1_000_000
  if (!Number.isFinite(mp) || mp <= 0) return ''
  return `${mp >= 10 ? mp.toFixed(1) : mp.toFixed(2)}MP`
}

function revokeObjectUrls(): void {
  for (const url of objectUrls) {
    URL.revokeObjectURL(url)
  }
  objectUrls.clear()
}
</script>

<style scoped>
.image-manager-page {
  display: flex;
  min-height: calc(100vh - 7rem);
  flex-direction: column;
  gap: 1rem;
}

.image-manager-header,
.image-manager-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.image-manager-filters {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 0.75rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: rgb(255 255 255 / 0.84);
  padding: 0.85rem;
}

.dark .image-manager-filters {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.72);
}

.image-manager-field {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.72rem;
  font-weight: 600;
  color: rgb(100 116 139);
}

.dark .image-manager-field {
  color: rgb(148 163 184);
}

.image-manager-field-wide {
  grid-column: span 2;
}

.image-manager-filter-actions {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.image-manager-header {
  flex-wrap: wrap;
}

.image-manager-actions,
.image-manager-selection,
.image-manager-card-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.image-manager-toolbar {
  min-height: 3rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: rgb(255 255 255 / 0.84);
  padding: 0.65rem 0.85rem;
}

.dark .image-manager-toolbar {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.72);
}

.image-manager-selection {
  border-radius: 9999px;
  background: rgb(239 246 255);
  padding: 0.35rem 0.55rem 0.35rem 0.75rem;
  color: rgb(30 64 175);
  font-size: 0.75rem;
  font-weight: 600;
}

.dark .image-manager-selection {
  background: rgb(30 64 175 / 0.18);
  color: rgb(191 219 254);
}

.image-manager-text-button,
.image-manager-icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  border-radius: 0.375rem;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.image-manager-text-button {
  padding: 0.2rem 0.45rem;
}

.image-manager-text-button:hover {
  background: rgb(219 234 254);
}

.dark .image-manager-text-button:hover {
  background: rgb(30 64 175 / 0.26);
}

.image-manager-text-button.danger,
.image-manager-icon-button.danger {
  color: rgb(220 38 38);
}

.image-manager-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1rem;
}

.image-manager-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: white;
  box-shadow: 0 12px 30px rgb(15 23 42 / 0.08);
}

.image-manager-card.selected {
  border-color: rgb(37 99 235);
  box-shadow: 0 0 0 3px rgb(59 130 246 / 0.18);
}

.dark .image-manager-card {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39 / 0.72);
}

.image-manager-preview {
  display: flex;
  width: 100%;
  aspect-ratio: 1 / 1;
  align-items: center;
  justify-content: center;
  background: rgb(241 245 249);
}

.dark .image-manager-preview {
  background: rgb(15 23 42 / 0.75);
}

.image-manager-preview img {
  height: 100%;
  width: 100%;
  object-fit: contain;
}

.image-manager-select {
  position: absolute;
  left: 0.55rem;
  top: 0.55rem;
  z-index: 2;
  display: flex;
  height: 1.55rem;
  width: 1.55rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  border: 1px solid rgb(255 255 255 / 0.88);
  background: rgb(15 23 42 / 0.42);
  color: white;
}

.image-manager-select.active {
  background: rgb(37 99 235);
}

.image-manager-card-body {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  padding: 0.8rem;
}

.image-manager-format {
  border-radius: 9999px;
  background: rgb(241 245 249);
  padding: 0.18rem 0.5rem;
  font-size: 0.68rem;
  font-weight: 700;
  color: rgb(71 85 105);
}

.dark .image-manager-format {
  background: rgb(30 41 59);
  color: rgb(203 213 225);
}

.image-manager-prompt {
  min-height: 2.5rem;
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  font-size: 0.82rem;
  line-height: 1.55;
  color: rgb(51 65 85);
}

.dark .image-manager-prompt {
  color: rgb(226 232 240);
}

.image-manager-meta {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  font-size: 0.72rem;
  color: rgb(100 116 139);
}

.dark .image-manager-meta {
  color: rgb(148 163 184);
}

.image-manager-card-actions {
  justify-content: flex-end;
}

.image-manager-icon-button {
  height: 2rem;
  width: 2rem;
  border: 1px solid rgb(226 232 240);
  background: white;
  color: rgb(71 85 105);
}

.image-manager-icon-button:hover {
  border-color: rgb(147 197 253);
  color: rgb(37 99 235);
}

.dark .image-manager-icon-button {
  border-color: rgb(55 65 81);
  background: rgb(15 23 42 / 0.6);
  color: rgb(203 213 225);
}

.image-manager-state {
  display: flex;
  min-height: 28rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
}

.image-manager-footer {
  display: flex;
  justify-content: center;
  padding: 0.5rem 0 1rem;
}

.image-manager-lightbox {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: flex;
  flex-direction: column;
  background: rgb(0 0 0 / 0.86);
  padding: 1rem;
  backdrop-filter: blur(8px);
}

.image-manager-lightbox-close {
  display: flex;
  height: 2.25rem;
  width: 2.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: rgb(255 255 255 / 0.95);
  color: rgb(55 65 81);
}

.dark .image-manager-lightbox-close {
  background: rgb(31 41 55 / 0.95);
  color: rgb(243 244 246);
}

@media (max-width: 640px) {
  .image-manager-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 0.75rem;
  }

  .image-manager-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .image-manager-field-wide {
    grid-column: span 1;
  }
}
</style>
