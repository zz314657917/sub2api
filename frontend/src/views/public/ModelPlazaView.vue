<template>
  <div class="model-plaza-page public-page-shell relative min-h-screen overflow-hidden">
    <PublicRevealBackdrop variant="page" />
    <PublicTopNav />

    <main class="model-plaza-main relative z-10 mx-auto">
      <section class="model-plaza-hero">
        <div class="min-w-0">
          <span class="model-title-kicker">Pricing Center</span>
          <h1>模型定价</h1>
          <p>公开展示推理、图像和视频模型价格。实际扣费以控制台使用记录为准。</p>
        </div>
        <div class="model-hero-actions">
          <RouterLink class="model-tutorial-link" to="/tutorial/getting-started">
            <Icon name="book" size="sm" />
            接入教程
            <Icon name="arrowRight" size="sm" />
          </RouterLink>
          <button class="model-refresh-button" type="button" :disabled="loading" @click="loadCatalog">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            刷新目录
          </button>
        </div>
      </section>

      <section class="model-toolbar">
        <div class="model-search-box">
          <Icon name="search" size="sm" />
          <input
            v-model="searchQuery"
            data-testid="model-search-input"
            type="search"
            aria-label="搜索模型、规格或分组"
            placeholder="搜索模型、规格或分组..."
          />
          <button
            v-if="searchQuery"
            class="model-search-clear"
            type="button"
            aria-label="清空搜索"
            title="清空搜索"
            @click="searchQuery = ''"
          >
            <Icon name="x" size="sm" />
          </button>
        </div>

        <div class="model-filter-row">
          <div class="model-category-tabs" aria-label="模型类型筛选">
            <button
              v-for="item in categoryTabs"
              :key="item.value"
              type="button"
              :class="{ 'is-active': selectedCategory === item.value }"
              :aria-pressed="selectedCategory === item.value"
              @click="selectedCategory = item.value"
            >
              <span>{{ item.label }}</span>
              <small>{{ item.count }}</small>
            </button>
          </div>
          <p class="model-result-count" data-testid="model-result-count" aria-live="polite">
            共 <strong>{{ visibleRowCount }}</strong> 个型号/规格，来自
            <strong>{{ visibleGroups.length }}</strong> 个分组
          </p>
        </div>
      </section>

      <section class="model-price-context" aria-label="价格口径说明">
        <Icon name="infoCircle" size="md" />
        <div>
          <p>
            <strong>分组价格：</strong>当前选择的账号分组仅用于价格预览，不代表匿名访问者或登录账号的实际分组；最终以控制台显示为准。
          </p>
          <p><strong>✪ 单位：</strong>✪ 是本站额度单位，不代表人民币或美元；实际扣费以使用记录为准。</p>
        </div>
      </section>

      <section v-if="loadError && hasCatalogData" class="model-refresh-warning" role="status">
        <Icon name="exclamationCircle" size="md" />
        <div>
          <strong>目录刷新失败，正在显示上次成功加载的内容。</strong>
          <span>{{ loadError }}</span>
        </div>
        <button type="button" :disabled="loading" @click="loadCatalog">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          再次刷新
        </button>
      </section>

      <section v-if="loadError && !hasCatalogData" class="model-message-card">
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
                <h2>{{ displayModelLabel(group.title) }}</h2>
                <p>{{ displayModelLabel(group.description || categoryDescription(group.category)) }}</p>
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
              :id="`model-table-${group.id}`"
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
                  <tr
                    v-for="(row, rowIndex) in enabledRows(group)"
                    :key="row.id"
                    :class="{ 'is-mobile-preview-hidden': isMobilePreviewRowHidden(group, rowIndex) }"
                  >
                    <template v-if="group.category === 'chat'">
                      <td data-label="模型名称">
                        <div class="model-name-cell">
                          <ModelIcon :model="row.model || group.title" size="22px" />
                          <div class="min-w-0">
                            <strong>{{ displayModelLabel(row.model) }}</strong>
                            <small v-if="row.note">{{ displayModelLabel(row.note) }}</small>
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
                          <strong>{{ displayModelLabel(row.spec) }}</strong>
                          <small v-if="row.note">{{ displayModelLabel(row.note) }}</small>
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
            <span
              v-if="hasScrollableRows(group)"
              class="model-scrollbar-rail"
              :data-scroll-rail-group-id="group.id"
              aria-hidden="true"
              @pointerdown="handleScrollbarRailPointerDown(group.id, $event)"
            >
              <span
                class="model-scrollbar-thumb"
                :style="scrollThumbStyle(group.id)"
                @pointerdown.stop="handleScrollbarThumbPointerDown(group.id, $event)"
              />
            </span>
          </div>
          <button
            v-if="hasMobilePreviewRows(group)"
            class="model-mobile-preview-toggle"
            type="button"
            :aria-expanded="isMobileGroupExpanded(group.id)"
            :aria-controls="`model-table-${group.id}`"
            data-testid="model-mobile-preview-toggle"
            @click="toggleMobileGroup(group.id)"
          >
            <Icon :name="isMobileGroupExpanded(group.id) ? 'chevronUp' : 'chevronDown'" size="sm" />
            <span v-if="isMobileGroupExpanded(group.id)">收起完整列表</span>
            <span v-else>展开其余 {{ mobilePreviewRemainingRows(group) }} 个型号/规格</span>
          </button>
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
import { CREDIT_SYMBOL } from '@/utils/credits'
import { cleanModelDisplayName, displayModelLabel } from '@/utils/modelDisplay'
import PublicRevealBackdrop from './components/PublicRevealBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'

type CategoryFilter = 'all' | ModelMarketCategory
type ScrollbarDragState = {
  groupId: string
  pointerId: number
  startY: number
  startScrollTop: number
  maxScroll: number
  maxTop: number
  target: HTMLElement
}

const loading = ref(false)
const loadError = ref('')
const searchQuery = ref('')
const selectedCategory = ref<CategoryFilter>('all')
const scrollIndicators = reactive<Record<string, { height: number; top: number }>>({})
const selectedRateGroupIds = reactive<Record<string, number>>({})
const expandedMobileGroupIds = reactive<Record<string, boolean>>({})
const MOBILE_ROW_PREVIEW_LIMIT = 6
let activeScrollbarDrag: ScrollbarDragState | null = null
const catalog = ref<ModelMarketCatalog>({
  version: 1,
  groups: []
})

const enabledGroups = computed(() => catalog.value.groups.filter((group) => group.enabled))
const hasCatalogData = computed(() => enabledGroups.value.some((group) => enabledRows(group).length > 0))
const normalizedSearchQuery = computed(() => searchQuery.value.trim().toLowerCase())

const searchFilteredGroups = computed<ModelMarketGroup[]>(() => enabledGroups.value.flatMap((group) => {
  const rows = enabledRows(group)
  if (rows.length === 0) return []

  const query = normalizedSearchQuery.value
  const matchingRows = !query || groupMatchesSearch(group, query)
    ? rows
    : rows.filter((row) => rowMatchesSearch(group, row, query))

  return matchingRows.length > 0 ? [{ ...group, rows: matchingRows }] : []
}))

const visibleGroups = computed(() => searchFilteredGroups.value.filter((group) => (
  selectedCategory.value === 'all' || group.category === selectedCategory.value
)))

const visibleRowCount = computed(() => countGroupRows(visibleGroups.value))

const categoryTabs = computed(() => {
  const count = (category: CategoryFilter) => {
    const groups = category === 'all'
      ? searchFilteredGroups.value
      : searchFilteredGroups.value.filter((group) => group.category === category)
    return countGroupRows(groups)
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
    catalog.value = normalizeCatalog(await modelMarketAPI.getCatalog())
    initializeSelectedRateGroups()
    pruneExpandedMobileGroups()
    await nextTick()
    refreshScrollIndicators()
  } catch (error: any) {
    loadError.value = error?.message || '请稍后再试'
  } finally {
    loading.value = false
  }
}

function normalizeCatalog(value: ModelMarketCatalog): ModelMarketCatalog {
  if (!value || !Array.isArray(value.groups)) {
    throw new Error('模型目录响应格式无效')
  }

  return {
    ...value,
    groups: value.groups.map((group) => ({
      ...group,
      rows: Array.isArray(group.rows) ? group.rows : [],
      supported_groups: Array.isArray(group.supported_groups) ? group.supported_groups : []
    }))
  }
}

function enabledRows(group: ModelMarketGroup) {
  return group.rows.filter((row) => row.enabled)
}

function countGroupRows(groups: ModelMarketGroup[]): number {
  return groups.reduce((total, group) => total + enabledRows(group).length, 0)
}

function groupMatchesSearch(group: ModelMarketGroup, query: string): boolean {
  return searchValuesMatch([
    group.title,
    group.description,
    group.platform
  ], query)
}

function rowMatchesSearch(
  group: ModelMarketGroup,
  row: ModelMarketGroup['rows'][number],
  query: string
): boolean {
  return searchValuesMatch([
    row.model,
    row.spec,
    row.input_price,
    row.output_price,
    row.our_price,
    group.hide_official_price ? '' : row.official_price,
    group.hide_saving ? '' : row.saving,
    row.note
  ], query)
}

function searchValuesMatch(values: unknown[], query: string): boolean {
  return values.some((value) => searchableText(value).includes(query))
}

function searchableText(value: unknown): string {
  if (typeof value !== 'string' || !value.trim()) return ''
  return cleanModelDisplayName(value, '').toLowerCase()
}

function groupRateOptions(group: ModelMarketGroup): ModelMarketAccountGroup[] {
  return (group.supported_groups ?? [])
    .filter((item) => item.id > 0 && Number.isFinite(item.effective_rate_multiplier))
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
      selectedRateGroupIds[group.id] = defaultRateGroupId(options)
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
  return options.some((option) => option.id === current) ? current : defaultRateGroupId(options)
}

function defaultRateGroupId(options: ModelMarketAccountGroup[]): number {
  return options.reduce((closest, option) => {
    const closestDistance = Math.abs(closest.effective_rate_multiplier - 1)
    const optionDistance = Math.abs(option.effective_rate_multiplier - 1)
    return optionDistance < closestDistance ? option : closest
  }).id
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
  return price.replace(/[¥￥$]/g, CREDIT_SYMBOL)
}

function multiplyPriceText(price: string, rate: number): string {
  return price.replace(/([✪¥￥$])\s*(-?\d+(?:\.\d+)?)(?:\s*-\s*([✪¥￥$])?\s*(\d+(?:\.\d+)?))?/g, (_match, symbol: string, value: string, rangeSymbol: string | undefined, rangeValue: string | undefined) => {
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

function hasMobilePreviewRows(group: ModelMarketGroup): boolean {
  return enabledRows(group).length > MOBILE_ROW_PREVIEW_LIMIT
}

function isMobileGroupExpanded(groupId: string): boolean {
  return expandedMobileGroupIds[groupId] === true
}

function isMobilePreviewRowHidden(group: ModelMarketGroup, rowIndex: number): boolean {
  return hasMobilePreviewRows(group)
    && !isMobileGroupExpanded(group.id)
    && rowIndex >= MOBILE_ROW_PREVIEW_LIMIT
}

function mobilePreviewRemainingRows(group: ModelMarketGroup): number {
  return Math.max(0, enabledRows(group).length - MOBILE_ROW_PREVIEW_LIMIT)
}

function toggleMobileGroup(groupId: string): void {
  expandedMobileGroupIds[groupId] = !isMobileGroupExpanded(groupId)
}

function pruneExpandedMobileGroups(): void {
  const activeGroupIds = new Set(catalog.value.groups.map((group) => group.id))
  Object.keys(expandedMobileGroupIds).forEach((groupId) => {
    if (!activeGroupIds.has(groupId)) {
      delete expandedMobileGroupIds[groupId]
    }
  })
}

function collapseMobileGroups(): void {
  Object.keys(expandedMobileGroupIds).forEach((groupId) => {
    delete expandedMobileGroupIds[groupId]
  })
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
  const railHeight = scrollRailHeightForGroup(groupId, clientHeight)

  if (scrollHeight <= clientHeight) {
    scrollIndicators[groupId] = { height: railHeight, top: 0 }
    return
  }

  const height = Math.max(44, Math.round((clientHeight / scrollHeight) * railHeight))
  const maxTop = Math.max(0, railHeight - height)
  const maxScroll = Math.max(1, scrollHeight - clientHeight)
  const top = Math.round((element.scrollTop / maxScroll) * maxTop)
  scrollIndicators[groupId] = { height, top }
}

function handleScrollbarRailPointerDown(groupId: string, event: PointerEvent): void {
  if (event.button !== 0) return
  const rail = event.currentTarget as HTMLElement
  const element = scrollElementForGroup(groupId)
  if (!element) return

  event.preventDefault()
  const indicator = scrollIndicators[groupId] || { height: 44, top: 0 }
  const railRect = rail.getBoundingClientRect()
  const thumbTop = event.clientY - railRect.top - indicator.height / 2
  setScrollFromThumbTop(groupId, element, railRect.height, indicator.height, thumbTop)
}

function handleScrollbarThumbPointerDown(groupId: string, event: PointerEvent): void {
  if (event.button !== 0) return
  const target = event.currentTarget as HTMLElement
  const rail = target.parentElement
  const element = scrollElementForGroup(groupId)
  if (!rail || !element) return

  event.preventDefault()
  stopActiveScrollbarDrag()
  target.setPointerCapture?.(event.pointerId)

  const indicator = scrollIndicators[groupId] || { height: target.offsetHeight || 44, top: 0 }
  const railHeight = rail.getBoundingClientRect().height
  activeScrollbarDrag = {
    groupId,
    pointerId: event.pointerId,
    startY: event.clientY,
    startScrollTop: element.scrollTop,
    maxScroll: Math.max(0, element.scrollHeight - element.clientHeight),
    maxTop: Math.max(0, railHeight - indicator.height),
    target
  }

  if (typeof window !== 'undefined') {
    window.addEventListener('pointermove', handleScrollbarPointerMove)
    window.addEventListener('pointerup', handleScrollbarPointerEnd)
    window.addEventListener('pointercancel', handleScrollbarPointerEnd)
  }
}

function handleScrollbarPointerMove(event: PointerEvent): void {
  const drag = activeScrollbarDrag
  if (!drag || event.pointerId !== drag.pointerId || drag.maxScroll <= 0 || drag.maxTop <= 0) return
  const element = scrollElementForGroup(drag.groupId)
  if (!element) return

  event.preventDefault()
  const scrollDelta = ((event.clientY - drag.startY) / drag.maxTop) * drag.maxScroll
  element.scrollTop = clampNumber(drag.startScrollTop + scrollDelta, 0, drag.maxScroll)
  setScrollIndicator(drag.groupId, element)
}

function handleScrollbarPointerEnd(event: PointerEvent): void {
  if (activeScrollbarDrag && event.pointerId !== activeScrollbarDrag.pointerId) return
  stopActiveScrollbarDrag()
}

function stopActiveScrollbarDrag(): void {
  if (typeof window !== 'undefined') {
    window.removeEventListener('pointermove', handleScrollbarPointerMove)
    window.removeEventListener('pointerup', handleScrollbarPointerEnd)
    window.removeEventListener('pointercancel', handleScrollbarPointerEnd)
  }
  if (activeScrollbarDrag?.target.hasPointerCapture?.(activeScrollbarDrag.pointerId)) {
    activeScrollbarDrag.target.releasePointerCapture?.(activeScrollbarDrag.pointerId)
  }
  activeScrollbarDrag = null
}

function setScrollFromThumbTop(
  groupId: string,
  element: HTMLElement,
  railHeight: number,
  thumbHeight: number,
  thumbTop: number
): void {
  const maxScroll = Math.max(0, element.scrollHeight - element.clientHeight)
  const maxTop = Math.max(0, railHeight - thumbHeight)
  if (maxScroll <= 0 || maxTop <= 0) {
    element.scrollTop = 0
    setScrollIndicator(groupId, element)
    return
  }

  const nextTop = clampNumber(thumbTop, 0, maxTop)
  element.scrollTop = (nextTop / maxTop) * maxScroll
  setScrollIndicator(groupId, element)
}

function scrollElementForGroup(groupId: string): HTMLElement | null {
  if (typeof document === 'undefined') return null
  const elements = document.querySelectorAll<HTMLElement>('.model-table-wrap.is-scrollable[data-scroll-group-id]')
  return Array.from(elements).find((element) => element.dataset.scrollGroupId === groupId) || null
}

function scrollRailForGroup(groupId: string): HTMLElement | null {
  if (typeof document === 'undefined') return null
  const elements = document.querySelectorAll<HTMLElement>('.model-scrollbar-rail[data-scroll-rail-group-id]')
  return Array.from(elements).find((element) => element.dataset.scrollRailGroupId === groupId) || null
}

function scrollRailHeightForGroup(groupId: string, fallbackClientHeight: number): number {
  const rail = scrollRailForGroup(groupId)
  if (rail) {
    return Math.max(0, rail.getBoundingClientRect().height)
  }

  return Math.max(0, fallbackClientHeight - 66)
}

function clampNumber(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
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

watch([searchQuery, selectedCategory], collapseMobileGroups)

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
  stopActiveScrollbarDrag()
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', refreshScrollIndicators)
  }
})
</script>

<style scoped>
@import './public-page.css';

.model-plaza-page {
  background: #faf9f5;
  color: var(--public-text);
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
  color: var(--public-accent);
  font-size: 0.74rem;
  font-weight: 500;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.model-plaza-hero h1 {
  color: var(--public-text);
  font-family: var(--public-font-display);
  font-size: clamp(1.65rem, 2.6vw, 2.45rem);
  font-weight: 400;
  line-height: 1;
}

.model-plaza-hero p {
  margin-top: 0.55rem;
  max-width: 42rem;
  color: var(--public-muted);
  font-size: 0.86rem;
  font-weight: 350;
}

.model-hero-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.55rem;
}

.model-tutorial-link,
.model-refresh-button,
.model-refresh-warning button,
.model-message-card button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.55rem;
  border-radius: 8px;
  padding: 0.58rem 0.78rem;
  font-size: 0.8rem;
  font-weight: 500;
  text-decoration: none;
}

.model-tutorial-link,
.model-message-card button {
  border: 1px solid #cc785c;
  background: #cc785c;
  color: #ffffff;
}

.model-refresh-button,
.model-refresh-warning button {
  border: 1px solid var(--public-border-strong);
  background: rgba(250, 249, 245, 0.78);
  color: var(--public-muted-strong);
}

.model-refresh-button:disabled,
.model-refresh-warning button:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.model-tutorial-link:focus-visible,
.model-refresh-button:focus-visible,
.model-refresh-warning button:focus-visible,
.model-message-card button:focus-visible,
.model-search-clear:focus-visible,
.model-category-tabs button:focus-visible,
.model-group-rate-select select:focus-visible,
.model-mobile-preview-toggle:focus-visible {
  outline: 2px solid var(--public-accent);
  outline-offset: 2px;
}

.model-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.62rem;
}

.model-filter-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-width: 0;
  gap: 1rem;
}

.model-search-box {
  display: flex;
  min-height: 2.7rem;
  align-items: center;
  gap: 0.7rem;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: #faf9f5;
  padding: 0 0.85rem;
  color: var(--public-muted);
}

.model-search-box:focus-within {
  border-color: var(--public-accent);
  box-shadow: 0 0 0 3px rgba(204, 120, 92, 0.2);
}

.model-search-clear {
  display: inline-flex;
  width: 1.85rem;
  height: 1.85rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: var(--public-muted);
}

.model-search-clear:hover {
  background: #eee8df;
  color: var(--public-text);
}

.model-search-box input {
  min-width: 0;
  width: 100%;
  background: transparent;
  color: var(--public-text);
  font-size: 0.86rem;
  font-weight: 400;
  outline: none;
}

.model-search-box input::placeholder {
  color: var(--public-muted-soft);
}

.model-result-count {
  flex: 0 0 auto;
  color: var(--public-muted);
  font-size: 0.74rem;
  font-weight: 350;
  text-align: right;
}

.model-result-count strong {
  color: var(--public-text);
  font-weight: 520;
}

.model-category-tabs {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  gap: 0.18rem;
}

.model-category-tabs button {
  display: inline-flex;
  align-items: center;
  gap: 0.42rem;
  min-height: 2.2rem;
  border-bottom: 2px solid transparent;
  border-radius: 4px 4px 0 0;
  padding: 0.34rem 0.62rem 0.28rem;
  color: var(--public-muted-strong);
  font-size: 0.8rem;
  font-weight: 500;
}

.model-category-tabs button.is-active {
  border-bottom-color: var(--public-accent);
  background: rgba(204, 120, 92, 0.08);
  color: var(--public-accent);
}

.model-category-tabs small {
  color: var(--public-muted-soft);
  font-size: 0.72rem;
  font-weight: 500;
}

.model-price-context {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  margin-top: 0.75rem;
  border-left: 3px solid #cc785c;
  background: rgba(245, 240, 232, 0.72);
  padding: 0.68rem 0.78rem;
  color: var(--public-muted-strong);
}

.model-price-context > svg {
  margin-top: 0.08rem;
  flex: 0 0 auto;
  color: var(--public-accent);
}

.model-price-context div {
  display: grid;
  gap: 0.22rem;
}

.model-price-context p {
  font-size: 0.75rem;
  font-weight: 350;
  line-height: 1.5;
}

.model-price-context strong {
  color: var(--public-text);
  font-weight: 520;
}

.model-refresh-warning {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  margin-top: 0.75rem;
  border: 1px solid rgba(169, 88, 62, 0.24);
  background: rgba(249, 235, 228, 0.86);
  padding: 0.65rem 0.72rem;
  color: #8f4c36;
}

.model-refresh-warning > svg {
  flex: 0 0 auto;
}

.model-refresh-warning div {
  display: grid;
  min-width: 0;
  gap: 0.12rem;
}

.model-refresh-warning strong {
  color: #783f2d;
  font-size: 0.78rem;
  font-weight: 520;
}

.model-refresh-warning span {
  overflow-wrap: anywhere;
  font-size: 0.72rem;
}

.model-refresh-warning button {
  min-height: 2.25rem;
  margin-left: auto;
  background: #faf9f5;
  color: #783f2d;
  white-space: nowrap;
}

.model-card-grid {
  margin-top: 1rem;
  display: grid;
  gap: 1rem;
}

.model-market-card {
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: rgba(250, 249, 245, 0.92);
  box-shadow: var(--public-shadow);
  backdrop-filter: blur(16px);
}

.model-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid var(--public-border);
  background: rgba(245, 240, 232, 0.8);
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
  border: 1px solid var(--public-border);
}

.model-card-icon.is-chat {
  background: rgba(20, 20, 19, 0.06);
  color: #141413;
}

.model-card-icon.is-image {
  background: var(--public-accent-soft);
  color: var(--public-accent);
}

.model-card-icon.is-video {
  background: rgba(204, 120, 92, 0.1);
  color: #a9583e;
}

.model-card-title h2 {
  overflow: hidden;
  color: var(--public-text);
  font-size: 1.05rem;
  font-weight: 560;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-card-title p {
  margin-top: 0.15rem;
  color: var(--public-muted);
  font-size: 0.78rem;
  font-weight: 350;
}

.model-group-rate-select {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.45rem;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: #faf9f5;
  padding: 0.42rem 0.5rem 0.42rem 0.62rem;
}

.model-group-rate-select span {
  color: var(--public-muted);
  font-size: 0.72rem;
  font-weight: 500;
  white-space: nowrap;
}

.model-group-rate-select select {
  max-width: 14rem;
  min-height: 2rem;
  border: 0;
  background: transparent;
  color: var(--public-accent);
  font-size: 0.78rem;
  font-weight: 500;
  outline: none;
}

.model-group-rate-select option {
  background: #faf9f5;
  color: #141413;
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
  top: 3rem;
  right: 0.24rem;
  bottom: 1.1rem;
  z-index: 6;
  width: 0.72rem;
  border-radius: 999px;
  background: #e6dfd8;
  box-shadow:
    inset 0 0 0 1px rgba(20, 20, 19, 0.05);
  cursor: pointer;
  pointer-events: auto;
  touch-action: none;
}

.model-scrollbar-thumb {
  display: block;
  min-height: 44px;
  width: 100%;
  border: 2px solid #e6dfd8;
  border-radius: 999px;
  background: #cc785c;
  box-shadow: 0 4px 12px rgba(204, 120, 92, 0.2);
  cursor: grab;
  transition: background 120ms ease, box-shadow 120ms ease;
}

.model-scrollbar-thumb:active {
  cursor: grabbing;
}

.model-mobile-preview-toggle {
  display: none;
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
  border-bottom: 1px solid var(--public-border);
  padding: 0.82rem 1rem;
  text-align: left;
  vertical-align: middle;
}

.model-pricing-table th {
  background: #f5f0e8;
  color: #6c6a64;
  font-size: 0.72rem;
  font-weight: 500;
}

.model-pricing-table tbody tr:hover {
  background: #f5f0e8;
}

.model-pricing-table tbody tr:last-child td {
  border-bottom: 0;
}

.model-pricing-table td {
  color: #3d3d3a;
  font-size: 0.86rem;
  font-weight: 350;
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
  color: var(--public-text);
  font-size: 0.9rem;
  font-weight: 520;
  line-height: 1.25;
}

.model-name-cell small,
.model-spec-cell small {
  display: block;
  margin-top: 0.22rem;
  color: var(--public-muted);
  font-size: 0.72rem;
  font-weight: 350;
  line-height: 1.35;
}

.model-price-value {
  color: var(--public-success);
  font-weight: 520;
  white-space: nowrap;
}

.model-saving {
  display: inline-flex;
  border-radius: 999px;
  border: 1px solid rgba(47, 122, 82, 0.2);
  background: rgba(47, 122, 82, 0.08);
  padding: 0.18rem 0.52rem;
  color: var(--public-success);
  font-size: 0.76rem;
  font-weight: 500;
  white-space: nowrap;
}

.model-saving.muted {
  border-color: var(--public-border);
  background: #f5f0e8;
  color: #6c6a64;
}

.model-message-card {
  margin-top: 1rem;
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-height: 10rem;
  border-radius: 8px;
  border: 1px dashed var(--public-border-strong);
  background: rgba(250, 249, 245, 0.9);
  padding: 1.2rem;
  color: var(--public-muted);
  box-shadow: var(--public-shadow-soft);
  backdrop-filter: blur(16px);
}

.model-message-card h2 {
  color: var(--public-text);
  font-size: 1rem;
  font-weight: 560;
}

.model-message-card p {
  margin-top: 0.2rem;
  font-size: 0.82rem;
  font-weight: 350;
}

.model-message-card button {
  margin-left: auto;
}

.model-plaza-note {
  padding: 1.25rem 0 0.25rem;
  color: var(--public-muted);
  text-align: center;
  font-size: 0.78rem;
  font-weight: 350;
}

@media (max-width: 900px) {
  .model-plaza-hero,
  .model-card-head,
  .model-message-card {
    align-items: stretch;
    flex-direction: column;
  }

  .model-group-rate-select {
    width: 100%;
  }

  .model-group-rate-select select {
    min-width: 0;
    max-width: none;
    flex: 1 1 auto;
  }

  .model-category-tabs {
    overflow-x: auto;
  }

  .model-refresh-warning {
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .model-refresh-warning button {
    margin-left: 1.9rem;
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

  .model-pricing-table.is-chat-table td,
  .model-pricing-table.is-media-table td {
    width: 100%;
  }

  .model-pricing-table.is-chat-table td:nth-child(n),
  .model-pricing-table.is-market-two-column td,
  .model-pricing-table.is-market-three-column td:nth-child(n),
  .model-pricing-table.is-market-four-column td:nth-child(n) {
    width: 100%;
  }

  .model-pricing-table tr {
    border-bottom: 1px solid var(--public-border);
    padding: 0.72rem 0;
  }

  .model-pricing-table tbody tr:last-child {
    border-bottom: 0;
  }

  .model-pricing-table td {
    display: grid;
    grid-template-columns: minmax(3.5rem, auto) minmax(0, 1fr);
    align-items: center;
    gap: 0.75rem;
    border-bottom: 0;
    padding: 0.38rem 1rem;
    text-align: right;
    white-space: nowrap;
  }

  .model-pricing-table td::before {
    content: attr(data-label);
    flex: 0 0 auto;
    color: var(--public-muted);
    font-size: 0.72rem;
    font-weight: 500;
    text-align: left;
  }

  .model-pricing-table td[data-label='模型名称'],
  .model-pricing-table td[data-label='规格'] {
    display: block;
    text-align: left;
    white-space: normal;
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

@media (max-width: 640px) {
  .model-plaza-page {
    overflow-x: hidden;
  }

  .model-plaza-main {
    padding: 0.9rem 0.72rem 4rem;
  }

  .model-plaza-hero {
    gap: 0.75rem;
    padding-bottom: 0.85rem;
  }

  .model-title-kicker {
    margin-bottom: 0.32rem;
    font-size: 0.7rem;
  }

  .model-plaza-hero h1 {
    font-size: 1.55rem;
  }

  .model-plaza-hero p {
    margin-top: 0.42rem;
    font-size: 0.82rem;
    line-height: 1.55;
  }

  .model-hero-actions {
    width: 100%;
  }

  .model-tutorial-link {
    flex: 1 1 auto;
  }

  .model-refresh-button,
  .model-tutorial-link,
  .model-refresh-warning button,
  .model-message-card button {
    min-height: 2.55rem;
    justify-content: center;
    padding-inline: 0.78rem;
  }

  .model-refresh-button {
    flex: 0 0 auto;
  }

  .model-toolbar {
    gap: 0.65rem;
  }

  .model-search-box {
    min-height: 2.5rem;
    padding-inline: 0.72rem;
  }

  .model-category-tabs {
    display: grid;
    width: 100%;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 0.22rem;
    overflow: visible;
  }

  .model-filter-row {
    display: grid;
    gap: 0.35rem;
  }

  .model-result-count {
    justify-self: start;
    text-align: left;
  }

  .model-category-tabs button {
    justify-content: center;
    gap: 0.3rem;
    min-height: 2.12rem;
    padding: 0.3rem 0.38rem;
    font-size: 0.76rem;
  }

  .model-card-grid {
    margin-top: 0.72rem;
    gap: 0.78rem;
  }

  .model-card-head {
    gap: 0.72rem;
    padding: 0.82rem;
  }

  .model-card-title {
    gap: 0.62rem;
  }

  .model-card-icon {
    width: 2.25rem;
    height: 2.25rem;
  }

  .model-card-title h2 {
    font-size: 0.98rem;
  }

  .model-card-title p {
    font-size: 0.74rem;
    line-height: 1.35;
  }

  .model-group-rate-select {
    gap: 0.42rem;
    padding: 0.42rem 0.52rem;
  }

  .model-group-rate-select select {
    font-size: 0.74rem;
    text-overflow: ellipsis;
  }

  .model-table-wrap.is-scrollable {
    max-height: none;
    overflow-y: visible;
  }

  .model-scrollbar-rail {
    display: none;
  }

  .model-pricing-table tr.is-mobile-preview-hidden {
    display: none;
  }

  .model-mobile-preview-toggle {
    display: flex;
    width: 100%;
    min-height: 2.75rem;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    border-top: 1px solid var(--public-border);
    background: rgba(245, 240, 232, 0.72);
    color: var(--public-accent);
    font-size: 0.76rem;
    font-weight: 520;
  }

  .model-pricing-table tr {
    margin: 0;
    padding: 0.62rem 0.62rem;
  }

  .model-pricing-table td {
    margin: 0.28rem 0;
    border: 1px solid var(--public-border);
    border-radius: 8px;
    background: #f5f0e8;
    padding: 0.44rem max(0.75rem, env(safe-area-inset-right)) 0.44rem 0.65rem;
    color: #3d3d3a;
    font-size: 0.8rem;
    line-height: 1.2;
  }

  .model-pricing-table td:not([data-label='模型名称']):not([data-label='规格']) {
    padding-right: calc(max(0.75rem, env(safe-area-inset-right)) + 2.55rem);
  }

  .model-pricing-table td[data-label='模型名称'],
  .model-pricing-table td[data-label='规格'] {
    margin-bottom: 0.42rem;
    border: 0;
    background: transparent;
    padding: 0 0.1rem;
  }

  .model-name-cell {
    gap: 0.58rem;
  }

  .model-name-cell strong,
  .model-spec-cell strong {
    font-size: 0.9rem;
  }

  .model-name-cell small,
  .model-spec-cell small {
    font-size: 0.7rem;
  }

  .model-price-value,
  .model-saving {
    justify-self: end;
  }

  .model-plaza-note {
    padding-top: 0.9rem;
    font-size: 0.72rem;
    line-height: 1.6;
  }
}
</style>
