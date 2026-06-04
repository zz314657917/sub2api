<template>
  <div class="model-plaza-page public-page-shell relative min-h-screen overflow-hidden">
    <PublicMatrixBackdrop />
    <PublicTopNav />

    <main class="model-plaza-main relative z-10 mx-auto">
      <section class="model-plaza-hero">
        <div class="min-w-0">
          <span class="model-title-kicker">Pricing Center</span>
          <h1>模型定价</h1>
          <p>公开展示推理、图像和视频模型价格。实际扣费以控制台使用记录为准。</p>
        </div>
        <button class="model-refresh-button" type="button" :disabled="loading" @click="loadCatalog">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          刷新
        </button>
      </section>

      <section class="model-toolbar">
        <div class="model-search-box">
          <Icon name="search" size="sm" />
          <input v-model="searchQuery" type="search" placeholder="搜索模型、规格或分组..." />
        </div>

        <div class="model-category-tabs" aria-label="模型类型筛选">
          <button
            v-for="item in categoryTabs"
            :key="item.value"
            type="button"
            :class="{ 'is-active': selectedCategory === item.value }"
            @click="selectedCategory = item.value"
          >
            <span>{{ item.label }}</span>
            <small>{{ item.count }}</small>
          </button>
        </div>
      </section>

      <section v-if="loadError" class="model-message-card">
        <Icon name="exclamationCircle" size="lg" />
        <div>
          <h2>模型目录加载失败</h2>
          <p>{{ loadError }}</p>
        </div>
        <button type="button" @click="loadCatalog">重试</button>
      </section>

      <section v-else-if="loading && visibleGroups.length === 0" class="model-message-card">
        <Icon name="refresh" size="lg" class="animate-spin" />
        <div>
          <h2>正在加载模型目录</h2>
          <p>请稍候。</p>
        </div>
      </section>

      <section v-else-if="visibleGroups.length > 0" class="model-card-grid">
        <article
          v-for="group in visibleGroups"
          :key="group.id"
          class="model-market-card"
          :data-category="group.category"
          data-testid="model-market-card"
        >
          <div class="model-card-head">
            <div class="model-card-title">
              <span class="model-card-icon" :class="`is-${group.category}`">
                <ModelIcon v-if="sampleModel(group)" :model="sampleModel(group)" size="24px" />
                <Icon v-else :name="categoryIcon(group.category)" size="md" />
              </span>
              <div class="min-w-0">
                <h2>{{ group.title }}</h2>
                <p>{{ group.description || categoryDescription(group.category) }}</p>
              </div>
            </div>
            <label v-if="groupRateOptions(group).length > 0" class="model-group-rate-select">
              <span>账号分组</span>
              <select
                :value="selectedRateGroupId(group)"
                @change="updateSelectedRateGroup(group.id, Number(($event.target as HTMLSelectElement).value))"
              >
                <option
                  v-for="option in groupRateOptions(group)"
                  :key="option.id"
                  :value="option.id"
                >
                  {{ option.name }} · {{ formatCompactRate(option.effective_rate_multiplier) }}x
                </option>
              </select>
            </label>
          </div>

          <div class="model-table-shell" :class="{ 'is-scrollable': hasScrollableRows(group) }">
            <div
              class="model-table-wrap"
              :class="{ 'is-scrollable': hasScrollableRows(group) }"
              :data-scroll-group-id="group.id"
              @scroll="updateScrollIndicator(group.id, $event)"
            >
              <table class="model-pricing-table" :class="pricingTableClass(group)">
                <thead>
                  <tr v-if="group.category === 'chat'">
                    <th>模型名称</th>
                    <th>输入</th>
                    <th>输出</th>
                    <th>我们的价格</th>
                  </tr>
                  <tr v-else>
                    <th>规格</th>
                    <th>我们的价格</th>
                    <th v-if="showOfficialPriceColumn(group)">官方价格</th>
                    <th v-if="showSavingColumn(group)">节省</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in enabledRows(group)" :key="row.id">
                    <template v-if="group.category === 'chat'">
                      <td data-label="模型名称">
                        <div class="model-name-cell">
                          <ModelIcon :model="row.model || group.title" size="22px" />
                          <div class="min-w-0">
                            <strong>{{ row.model }}</strong>
                            <small v-if="row.note">{{ row.note }}</small>
                          </div>
                        </div>
                      </td>
                      <td data-label="输入">{{ row.input_price || '-' }}</td>
                      <td data-label="输出">{{ row.output_price || '-' }}</td>
                      <td data-label="我们的价格">
                        <span class="model-price-value">{{ displayOurPrice(group, row.our_price) }}</span>
                      </td>
                    </template>
                    <template v-else>
                      <td data-label="规格">
                        <div class="model-spec-cell">
                          <strong>{{ row.spec }}</strong>
                          <small v-if="row.note">{{ row.note }}</small>
                        </div>
                      </td>
                      <td data-label="我们的价格">
                        <span class="model-price-value">{{ displayOurPrice(group, row.our_price) }}</span>
                      </td>
                      <td v-if="showOfficialPriceColumn(group)" data-label="官方价格">{{ row.official_price || '-' }}</td>
                      <td v-if="showSavingColumn(group)" data-label="节省">
                        <span class="model-saving" :class="{ muted: !row.saving }">
                          {{ row.saving || '-' }}
                        </span>
                      </td>
                    </template>
                  </tr>
                </tbody>
              </table>
            </div>
            <span v-if="hasScrollableRows(group)" class="model-scrollbar-rail" aria-hidden="true">
              <span class="model-scrollbar-thumb" :style="scrollThumbStyle(group.id)" />
            </span>
          </div>
        </article>
      </section>

      <section v-else class="model-message-card">
        <Icon name="search" size="lg" />
        <div>
          <h2>没有找到匹配的模型</h2>
          <p>可以调整搜索词或切换分类。</p>
        </div>
        <button type="button" @click="resetFilters">清空筛选</button>
      </section>

      <p class="model-plaza-note">
        官方价格按公开报价或上游口径展示；我们的价格为当前站点公示价。实际扣费会受到渠道、规格、质量和任务参数影响。
      </p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import { modelMarketAPI, type ModelMarketAccountGroup, type ModelMarketCatalog, type ModelMarketCategory, type ModelMarketGroup } from '@/api/modelMarket'
import PublicMatrixBackdrop from './components/PublicMatrixBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'

type CategoryFilter = 'all' | ModelMarketCategory

const loading = ref(false)
const loadError = ref('')
const searchQuery = ref('')
const selectedCategory = ref<CategoryFilter>('all')
const scrollIndicators = reactive<Record<string, { height: number; top: number }>>({})
const selectedRateGroupIds = reactive<Record<string, number>>({})
const catalog = ref<ModelMarketCatalog>({
  version: 1,
  groups: []
})

const enabledGroups = computed(() => catalog.value.groups.filter((group) => group.enabled))

const visibleGroups = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return enabledGroups.value.filter((group) => {
    if (selectedCategory.value !== 'all' && group.category !== selectedCategory.value) return false
    if (!query) return enabledRows(group).length > 0

    const groupMatches = [
      group.title,
      group.description,
      group.platform
    ].some((value) => (value || '').toLowerCase().includes(query))
    const rowMatches = enabledRows(group).some((row) =>
      [
        row.model,
        row.spec,
        row.input_price,
        row.output_price,
        row.our_price,
        group.hide_official_price ? '' : row.official_price,
        group.hide_saving ? '' : row.saving,
        row.note
      ].some((value) => (value || '').toLowerCase().includes(query))
    )
    return groupMatches || rowMatches
  })
})

const categoryTabs = computed(() => {
  const count = (category: CategoryFilter) => {
    if (category === 'all') return enabledGroups.value.length
    return enabledGroups.value.filter((group) => group.category === category).length
  }
  return [
    { value: 'all' as const, label: '全部', count: count('all') },
    { value: 'chat' as const, label: '推理', count: count('chat') },
    { value: 'image' as const, label: '图像', count: count('image') },
    { value: 'video' as const, label: '视频', count: count('video') }
  ]
})

async function loadCatalog(): Promise<void> {
  loading.value = true
  loadError.value = ''
  try {
    catalog.value = await modelMarketAPI.getCatalog()
    initializeSelectedRateGroups()
    await nextTick()
    refreshScrollIndicators()
  } catch (error: any) {
    loadError.value = error?.message || '请稍后再试'
  } finally {
    loading.value = false
  }
}

function enabledRows(group: ModelMarketGroup) {
  return group.rows.filter((row) => row.enabled)
}

function groupRateOptions(group: ModelMarketGroup): ModelMarketAccountGroup[] {
  return (group.supported_groups ?? [])
    .filter((item) => item.id > 0 && Number.isFinite(item.effective_rate_multiplier))
    .sort((a, b) => {
      if (a.effective_rate_multiplier === b.effective_rate_multiplier) {
        return a.name.localeCompare(b.name)
      }
      return a.effective_rate_multiplier - b.effective_rate_multiplier
    })
}

function initializeSelectedRateGroups(): void {
  const activeGroupIds = new Set<string>()
  for (const group of catalog.value.groups) {
    activeGroupIds.add(group.id)
    const options = groupRateOptions(group)
    if (options.length === 0) {
      delete selectedRateGroupIds[group.id]
      continue
    }
    const current = selectedRateGroupIds[group.id]
    if (!options.some((option) => option.id === current)) {
      selectedRateGroupIds[group.id] = options[0].id
    }
  }
  Object.keys(selectedRateGroupIds).forEach((groupId) => {
    if (!activeGroupIds.has(groupId)) {
      delete selectedRateGroupIds[groupId]
    }
  })
}

function selectedRateGroupId(group: ModelMarketGroup): number | '' {
  const options = groupRateOptions(group)
  if (options.length === 0) return ''
  const current = selectedRateGroupIds[group.id]
  return options.some((option) => option.id === current) ? current : options[0].id
}

function updateSelectedRateGroup(groupId: string, value: number): void {
  if (Number.isFinite(value) && value > 0) {
    selectedRateGroupIds[groupId] = value
  }
}

function selectedRateMultiplier(group: ModelMarketGroup): number {
  const id = selectedRateGroupId(group)
  if (id === '') return 1
  const option = groupRateOptions(group).find((item) => item.id === id)
  const rate = Number(option?.effective_rate_multiplier ?? 1)
  return Number.isFinite(rate) && rate >= 0 ? rate : 1
}

function displayOurPrice(group: ModelMarketGroup, price: string): string {
  const normalizedPrice = normalizeOurPriceCurrency(price)
  const rate = groupPriceMultiplier(group) * selectedRateMultiplier(group)
  if (Math.abs(rate - 1) < 0.0000001) {
    return normalizedPrice
  }
  return multiplyPriceText(normalizedPrice, rate)
}

function groupPriceMultiplier(group: ModelMarketGroup): number {
  const rate = Number(group.price_multiplier ?? 1)
  return Number.isFinite(rate) && rate >= 0 ? rate : 1
}

function normalizeOurPriceCurrency(price: string): string {
  return price.replace(/\$/g, '¥')
}

function multiplyPriceText(price: string, rate: number): string {
  return price.replace(/([¥$])\s*(-?\d+(?:\.\d+)?)(?:\s*-\s*([¥$])?\s*(\d+(?:\.\d+)?))?/g, (_match, symbol: string, value: string, rangeSymbol: string | undefined, rangeValue: string | undefined) => {
    const numeric = Number(value)
    if (!Number.isFinite(numeric)) {
      return `${symbol}${value}`
    }
    const scaled = `${symbol}${formatScaledPrice(numeric * rate, decimalPlaces(value))}`
    if (!rangeValue) {
      return scaled
    }
    const rangeNumeric = Number(rangeValue)
    if (!Number.isFinite(rangeNumeric)) {
      return scaled
    }
    return `${scaled}-${rangeSymbol || ''}${formatScaledPrice(rangeNumeric * rate, decimalPlaces(rangeValue))}`
  })
}

function decimalPlaces(value: string): number {
  const index = value.indexOf('.')
  return index >= 0 ? value.length - index - 1 : 0
}

function formatScaledPrice(value: number, sourceDecimals: number): string {
  if (!Number.isFinite(value)) return '0'
  const decimals = Math.min(8, Math.max(sourceDecimals, value < 1 ? 4 : 2))
  return Number.parseFloat(value.toFixed(decimals)).toString()
}

function formatCompactRate(value: number): string {
  const rate = Number(value)
  if (!Number.isFinite(rate)) return '1'
  return Number.parseFloat(rate.toFixed(6)).toString()
}

function hasScrollableRows(group: ModelMarketGroup): boolean {
  return enabledRows(group).length > 10
}

function showOfficialPriceColumn(group: ModelMarketGroup): boolean {
  return group.category !== 'chat' && !group.hide_official_price && enabledRows(group).some((row) => !!row.official_price)
}

function showSavingColumn(group: ModelMarketGroup): boolean {
  return group.category !== 'chat' && !group.hide_saving && enabledRows(group).some((row) => !!row.saving)
}

function pricingTableClass(group: ModelMarketGroup): Record<string, boolean> {
  if (group.category === 'chat') {
    return { 'is-chat-table': true }
  }
  const columnCount = 2 + (showOfficialPriceColumn(group) ? 1 : 0) + (showSavingColumn(group) ? 1 : 0)
  return {
    'is-media-table': true,
    'is-market-two-column': columnCount === 2,
    'is-market-three-column': columnCount === 3,
    'is-market-four-column': columnCount === 4
  }
}

function updateScrollIndicator(groupId: string, event: Event): void {
  setScrollIndicator(groupId, event.currentTarget as HTMLElement)
}

function scrollThumbStyle(groupId: string): Record<string, string> {
  const indicator = scrollIndicators[groupId] || { height: 72, top: 0 }
  return {
    height: `${indicator.height}px`,
    transform: `translateY(${indicator.top}px)`
  }
}

function refreshScrollIndicators(): void {
  if (typeof document === 'undefined') return
  document.querySelectorAll<HTMLElement>('.model-table-wrap.is-scrollable[data-scroll-group-id]').forEach((element) => {
    const groupId = element.dataset.scrollGroupId
    if (groupId) {
      setScrollIndicator(groupId, element)
    }
  })
}

function setScrollIndicator(groupId: string, element: HTMLElement): void {
  const scrollHeight = element.scrollHeight
  const clientHeight = element.clientHeight
  if (scrollHeight <= clientHeight) {
    scrollIndicators[groupId] = { height: clientHeight, top: 0 }
    return
  }

  const railPadding = 18
  const railHeight = Math.max(0, clientHeight - railPadding * 2)
  const height = Math.max(44, Math.round((clientHeight / scrollHeight) * railHeight))
  const maxTop = Math.max(0, railHeight - height)
  const maxScroll = Math.max(1, scrollHeight - clientHeight)
  const top = Math.round((element.scrollTop / maxScroll) * maxTop)
  scrollIndicators[groupId] = { height, top }
}

function sampleModel(group: ModelMarketGroup): string {
  return enabledRows(group).find((row) => row.model)?.model || group.title
}

function categoryIcon(category: ModelMarketCategory) {
  if (category === 'chat') return 'cube'
  if (category === 'image') return 'image'
  return 'play'
}

function categoryDescription(category: ModelMarketCategory): string {
  if (category === 'chat') return '按百万 tokens 展示输入、输出和本站价格。'
  if (category === 'image') return '按输出尺寸、质量或固定规格展示单次价格。'
  return '按秒、任务或分辨率规格展示视频模型价格。'
}

function resetFilters(): void {
  searchQuery.value = ''
  selectedCategory.value = 'all'
}

watch(visibleGroups, async () => {
  await nextTick()
  refreshScrollIndicators()
}, { flush: 'post' })

onMounted(() => {
  void loadCatalog()
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', refreshScrollIndicators)
  }
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', refreshScrollIndicators)
  }
})
</script>

<style scoped>
@import './public-page.css';

.model-plaza-page {
  background:
    radial-gradient(circle at 50% 18%, rgba(32, 170, 92, 0.18) 0, transparent 34%),
    radial-gradient(circle at 18% 24%, rgba(87, 86, 210, 0.14) 0, transparent 30%),
    linear-gradient(180deg, #050914 0%, #08110f 48%, #03060a 100%);
}

.model-plaza-main {
  width: min(100%, 88rem);
  padding: 1.2rem 1rem 3rem;
}

.model-plaza-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.25rem 0 1.2rem;
}

.model-title-kicker {
  display: block;
  margin-bottom: 0.42rem;
  color: #7dffaa;
  font-size: 0.74rem;
  font-weight: 950;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.model-plaza-hero h1 {
  color: rgba(255, 255, 255, 0.96);
  font-size: clamp(1.65rem, 2.6vw, 2.45rem);
  font-weight: 950;
  line-height: 1;
}

.model-plaza-hero p {
  margin-top: 0.55rem;
  max-width: 42rem;
  color: rgba(238, 246, 240, 0.72);
  font-size: 0.86rem;
  font-weight: 700;
}

.model-refresh-button,
.model-message-card button {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 8px;
  border: 1px solid rgba(119, 255, 173, 0.34);
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.18), rgba(20, 184, 166, 0.08)),
    rgba(5, 15, 18, 0.72);
  padding: 0.62rem 0.86rem;
  color: #eafff0;
  font-size: 0.82rem;
  font-weight: 850;
}

.model-refresh-button:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.model-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.9rem;
  border-radius: 8px;
  border: 1px solid rgba(221, 230, 255, 0.12);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.055)),
    rgba(6, 13, 18, 0.66);
  padding: 0.85rem;
  box-shadow: 0 18px 36px rgba(0, 0, 0, 0.22);
  backdrop-filter: blur(18px);
}

.model-search-box {
  display: flex;
  min-height: 2.7rem;
  align-items: center;
  gap: 0.7rem;
  border-radius: 8px;
  border: 1px solid rgba(221, 230, 255, 0.12);
  background: rgba(2, 8, 12, 0.54);
  padding: 0 0.85rem;
  color: rgba(222, 232, 255, 0.62);
}

.model-search-box input {
  min-width: 0;
  width: 100%;
  background: transparent;
  color: rgba(255, 255, 255, 0.92);
  font-size: 0.86rem;
  font-weight: 700;
  outline: none;
}

.model-search-box input::placeholder {
  color: rgba(222, 232, 255, 0.46);
}

.model-category-tabs {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  border-radius: 8px;
  border: 1px solid rgba(221, 230, 255, 0.12);
  background: rgba(2, 8, 12, 0.38);
  padding: 0.24rem;
}

.model-category-tabs button {
  display: inline-flex;
  align-items: center;
  gap: 0.42rem;
  min-height: 2.2rem;
  border-radius: 6px;
  padding: 0.34rem 0.65rem;
  color: rgba(238, 246, 240, 0.74);
  font-size: 0.8rem;
  font-weight: 850;
}

.model-category-tabs button.is-active {
  background: rgba(119, 255, 173, 0.16);
  color: rgba(236, 255, 241, 0.96);
}

.model-category-tabs small {
  color: rgba(222, 232, 255, 0.62);
  font-size: 0.72rem;
  font-weight: 900;
}

.model-card-grid {
  margin-top: 1rem;
  display: grid;
  gap: 1rem;
}

.model-market-card {
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid rgba(221, 230, 255, 0.14);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.085), rgba(255, 255, 255, 0.035)),
    rgba(5, 13, 18, 0.76);
  box-shadow: 0 22px 46px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(18px);
}

.model-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(221, 230, 255, 0.12);
  background: rgba(4, 11, 16, 0.58);
  padding: 1rem 1.1rem;
}

.model-card-title {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.75rem;
}

.model-card-icon {
  display: inline-flex;
  width: 2.55rem;
  height: 2.55rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid rgba(221, 230, 255, 0.14);
}

.model-card-icon.is-chat {
  background: rgba(96, 165, 250, 0.14);
  color: rgb(147, 197, 253);
}

.model-card-icon.is-image {
  background: rgba(74, 222, 128, 0.14);
  color: rgb(134, 239, 172);
}

.model-card-icon.is-video {
  background: rgba(196, 181, 253, 0.14);
  color: rgb(196, 181, 253);
}

.model-card-title h2 {
  overflow: hidden;
  color: rgba(255, 255, 255, 0.94);
  font-size: 1.05rem;
  font-weight: 950;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-card-title p {
  margin-top: 0.15rem;
  color: rgba(222, 232, 255, 0.62);
  font-size: 0.78rem;
  font-weight: 700;
}

.model-group-rate-select {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.45rem;
  border-radius: 8px;
  border: 1px solid rgba(119, 255, 173, 0.22);
  background: rgba(2, 8, 12, 0.42);
  padding: 0.42rem 0.5rem 0.42rem 0.62rem;
}

.model-group-rate-select span {
  color: rgba(191, 209, 230, 0.72);
  font-size: 0.72rem;
  font-weight: 850;
  white-space: nowrap;
}

.model-group-rate-select select {
  max-width: 14rem;
  min-height: 2rem;
  border: 0;
  background: transparent;
  color: rgba(236, 255, 241, 0.96);
  font-size: 0.78rem;
  font-weight: 900;
  outline: none;
}

.model-group-rate-select option {
  background: #07110f;
  color: #eafff0;
}

.model-table-shell {
  position: relative;
}

.model-table-wrap {
  overflow-x: auto;
}

.model-table-wrap.is-scrollable {
  max-height: 34rem;
  overflow-x: auto;
  overflow-y: scroll;
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.model-table-wrap.is-scrollable::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.model-scrollbar-rail {
  position: absolute;
  top: 18px;
  right: 0.24rem;
  bottom: 18px;
  z-index: 6;
  width: 0.72rem;
  border-radius: 999px;
  background: rgba(7, 17, 24, 0.9);
  box-shadow:
    inset 0 0 0 1px rgba(221, 230, 255, 0.08),
    0 0 0 1px rgba(0, 0, 0, 0.22);
  pointer-events: none;
}

.model-scrollbar-thumb {
  display: block;
  min-height: 44px;
  width: 100%;
  border: 2px solid rgba(7, 17, 24, 0.9);
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(164, 255, 196, 0.95), rgba(74, 222, 128, 0.82));
  box-shadow: 0 0 12px rgba(74, 222, 128, 0.32);
  transition: transform 120ms ease;
}

.model-table-wrap.is-scrollable .model-pricing-table th {
  position: sticky;
  top: 0;
  z-index: 2;
}

.model-pricing-table {
  width: 100%;
  min-width: 48rem;
  border-collapse: collapse;
  table-layout: fixed;
}

.model-pricing-table.is-chat-table th:nth-child(1),
.model-pricing-table.is-chat-table td:nth-child(1) {
  width: 36%;
}

.model-pricing-table.is-chat-table th:nth-child(2),
.model-pricing-table.is-chat-table td:nth-child(2),
.model-pricing-table.is-chat-table th:nth-child(3),
.model-pricing-table.is-chat-table td:nth-child(3) {
  width: 20%;
}

.model-pricing-table.is-chat-table th:nth-child(4),
.model-pricing-table.is-chat-table td:nth-child(4) {
  width: 24%;
}

.model-pricing-table.is-market-two-column th,
.model-pricing-table.is-market-two-column td {
  width: 50%;
}

.model-pricing-table.is-market-three-column th:nth-child(1),
.model-pricing-table.is-market-three-column td:nth-child(1) {
  width: 46%;
}

.model-pricing-table.is-market-three-column th:nth-child(2),
.model-pricing-table.is-market-three-column td:nth-child(2),
.model-pricing-table.is-market-three-column th:nth-child(3),
.model-pricing-table.is-market-three-column td:nth-child(3) {
  width: 27%;
}

.model-pricing-table.is-market-four-column th:nth-child(1),
.model-pricing-table.is-market-four-column td:nth-child(1) {
  width: 40%;
}

.model-pricing-table.is-market-four-column th:nth-child(2),
.model-pricing-table.is-market-four-column td:nth-child(2),
.model-pricing-table.is-market-four-column th:nth-child(3),
.model-pricing-table.is-market-four-column td:nth-child(3) {
  width: 23%;
}

.model-pricing-table.is-market-four-column th:nth-child(4),
.model-pricing-table.is-market-four-column td:nth-child(4) {
  width: 14%;
}

.model-pricing-table th,
.model-pricing-table td {
  border-bottom: 1px solid rgba(221, 230, 255, 0.1);
  padding: 0.82rem 1rem;
  text-align: left;
  vertical-align: middle;
}

.model-pricing-table th {
  background: rgba(9, 20, 28, 0.82);
  color: rgba(191, 209, 230, 0.68);
  font-size: 0.72rem;
  font-weight: 900;
}

.model-pricing-table tbody tr:hover {
  background: rgba(74, 222, 128, 0.07);
}

.model-pricing-table tbody tr:last-child td {
  border-bottom: 0;
}

.model-pricing-table td {
  color: rgba(224, 235, 244, 0.78);
  font-size: 0.86rem;
  font-weight: 750;
}

.model-name-cell {
  display: flex;
  align-items: flex-start;
  min-width: 14rem;
  gap: 0.68rem;
}

.model-name-cell strong,
.model-spec-cell strong {
  display: block;
  color: rgba(255, 255, 255, 0.92);
  font-size: 0.9rem;
  font-weight: 950;
  line-height: 1.25;
}

.model-name-cell small,
.model-spec-cell small {
  display: block;
  margin-top: 0.22rem;
  color: rgba(191, 209, 230, 0.6);
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.35;
}

.model-price-value {
  color: rgb(125, 255, 170);
  font-weight: 950;
  white-space: nowrap;
}

.model-saving {
  display: inline-flex;
  border-radius: 999px;
  border: 1px solid rgba(125, 255, 170, 0.24);
  background: rgba(34, 197, 94, 0.12);
  padding: 0.18rem 0.52rem;
  color: rgb(178, 255, 207);
  font-size: 0.76rem;
  font-weight: 950;
  white-space: nowrap;
}

.model-saving.muted {
  border-color: rgba(191, 209, 230, 0.14);
  background: rgba(148, 163, 184, 0.1);
  color: rgba(191, 209, 230, 0.62);
}

.model-message-card {
  margin-top: 1rem;
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-height: 10rem;
  border-radius: 8px;
  border: 1px dashed rgba(221, 230, 255, 0.24);
  background: rgba(6, 13, 18, 0.72);
  padding: 1.2rem;
  color: rgba(238, 246, 240, 0.76);
  backdrop-filter: blur(18px);
}

.model-message-card h2 {
  color: rgba(255, 255, 255, 0.94);
  font-size: 1rem;
  font-weight: 900;
}

.model-message-card p {
  margin-top: 0.2rem;
  font-size: 0.82rem;
  font-weight: 700;
}

.model-message-card button {
  margin-left: auto;
}

.model-plaza-note {
  padding: 1.25rem 0 0.25rem;
  color: rgba(222, 232, 255, 0.68);
  text-align: center;
  font-size: 0.78rem;
  font-weight: 700;
}

@media (max-width: 900px) {
  .model-plaza-hero,
  .model-card-head,
  .model-message-card {
    align-items: stretch;
    flex-direction: column;
  }

  .model-toolbar {
    grid-template-columns: 1fr;
  }

  .model-category-tabs {
    overflow-x: auto;
  }

  .model-pricing-table {
    min-width: 0;
    table-layout: auto;
  }

  .model-pricing-table thead {
    display: none;
  }

  .model-pricing-table,
  .model-pricing-table tbody,
  .model-pricing-table tr,
  .model-pricing-table td {
    display: block;
    width: 100%;
  }

  .model-pricing-table tr {
    border-bottom: 1px solid rgba(221, 230, 255, 0.1);
    padding: 0.72rem 0;
  }

  .model-pricing-table tbody tr:last-child {
    border-bottom: 0;
  }

  .model-pricing-table td {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    border-bottom: 0;
    padding: 0.38rem 1rem;
    text-align: right;
  }

  .model-pricing-table td::before {
    content: attr(data-label);
    flex: 0 0 auto;
    color: rgba(191, 209, 230, 0.62);
    font-size: 0.72rem;
    font-weight: 900;
  }

  .model-pricing-table td[data-label='模型名称'],
  .model-pricing-table td[data-label='规格'] {
    display: block;
    text-align: left;
  }

  .model-pricing-table td[data-label='模型名称']::before,
  .model-pricing-table td[data-label='规格']::before {
    display: none;
  }

  .model-name-cell {
    min-width: 0;
  }

  .model-message-card button {
    margin-left: 0;
  }
}
</style>
