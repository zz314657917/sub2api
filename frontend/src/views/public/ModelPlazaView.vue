<template>
  <div class="model-plaza-page relative min-h-screen overflow-hidden text-white">
    <PublicMatrixBackdrop />

    <PublicTopNav />

    <main class="model-plaza-main relative z-10 mx-auto">
      <section class="model-plaza-title">
        <div>
          <span class="model-title-kicker">Pricing Center</span>
          <h1>模型定价</h1>
          <p>按供应商查看 Claude、GPT、Gemini 等模型价格，支持按分组倍率折算到你的实际成本。</p>
        </div>
        <button class="model-refresh-button" type="button" :disabled="loading" @click="loadChannels(true)">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          刷新
        </button>
      </section>

      <section class="model-filter-panel">
        <div class="model-search-box">
          <Icon name="search" size="sm" />
          <input v-model="searchQuery" type="search" placeholder="搜索模型..." />
        </div>
      </section>

      <section class="model-provider-tabs" aria-label="模型供应商筛选">
        <button
          v-for="provider in providerTabs"
          :key="provider.value"
          type="button"
          :class="{ 'is-active': selectedProvider === provider.value }"
          @click="selectedProvider = provider.value"
        >
          <ModelIcon v-if="provider.sampleModel" :model="provider.sampleModel" size="16px" />
          <Icon v-else name="cube" size="sm" />
          <span>{{ provider.label }}</span>
          <small>{{ provider.count }}</small>
        </button>
      </section>

      <section v-if="filteredModels.length > 0" class="model-pricing-stack model-card-grid">
        <article
          v-for="section in providerPricingSections"
          :key="section.platform"
          class="model-provider-section"
        >
          <div class="model-provider-head">
            <div class="model-provider-title">
              <span class="model-provider-icon" :class="providerToneClass(section.platform)">
                <ModelIcon :model="section.sampleModel" size="24px" />
              </span>
              <div>
                <h2>{{ section.label }}</h2>
                <p>{{ providerDescription(section.platform) }}</p>
              </div>
            </div>
            <div class="model-provider-rate-actions">
              <Icon name="cube" size="sm" class="text-slate-400" />
              <span class="model-rate-group-label">倍率分组</span>
              <Select
                :model-value="selectedGroupKey(section.platform)"
                :options="groupOptionsForPlatform(section.platform)"
                class="model-group-select model-provider-group-select"
                :searchable="groupOptionsForPlatform(section.platform).length > 8"
                @update:model-value="updateProviderGroup(section.platform, $event)"
              />
              <label
                class="model-rate-toggle"
                :class="{ 'is-disabled': !canUseRatePrices(section.platform) }"
              >
                <Toggle
                  :model-value="showRatePricesForPlatform(section.platform) && canUseRatePrices(section.platform)"
                  @update:model-value="updateShowRatePrices(section.platform, $event)"
                />
                <span>{{ canUseRatePrices(section.platform) ? '显示倍率价格' : '暂无倍率分组' }}</span>
              </label>
            </div>
            <div class="model-provider-meta">
              <strong>{{ section.models.length }}</strong>
              <span>Models</span>
            </div>
          </div>

          <div class="model-pricing-table-wrap">
            <table class="model-pricing-table">
              <thead>
                <tr>
                  <th>模型</th>
                  <th v-if="shouldShowReferenceColumn(section.platform)">官方/参考价</th>
                  <th>我们的价格</th>
                  <th>计费说明</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="model in section.models" :key="model.key">
                  <td data-label="模型">
                    <div class="model-name-cell">
                      <ModelIcon :model="model.name" size="22px" />
                      <div class="min-w-0">
                        <strong>{{ model.name }}</strong>
                        <span>{{ modelAvailabilityLabel(model) }}</span>
                        <small v-if="modelVariantDescription(model)">
                          {{ modelVariantDescription(model) }}
                        </small>
                      </div>
                    </div>
                  </td>
                  <td v-if="shouldShowReferenceColumn(section.platform)" data-label="官方/参考价">
                    <span v-if="officialReferenceItems(model).length === 0" class="model-price-value muted">
                      {{ referencePrice(model) }}
                    </span>
                    <div v-if="officialReferenceItems(model).length > 0" class="model-tier-price-list model-reference-price-list">
                      <span
                        v-for="item in officialReferenceItems(model)"
                        :key="item.label"
                        class="model-tier-price-row model-reference-price-row"
                      >
                        <strong>{{ item.label }}</strong>
                        <em>{{ item.price }}</em>
                      </span>
                    </div>
                  </td>
                  <td data-label="我们的价格">
                    <span
                      v-if="tierPriceItems(model).length === 0"
                      class="model-price-value"
                      :class="{ 'is-rate-price': isRatePriceActive(model) }"
                    >
                      {{ formatPrimaryModelPrice(model) }}
                    </span>
                    <div v-if="tierPriceItems(model).length > 0" class="model-tier-price-list">
                      <span
                        v-for="item in tierPriceItems(model)"
                        :key="item.label"
                        class="model-tier-price-row"
                      >
                        <strong>{{ item.label }}</strong>
                        <em>{{ item.price }}</em>
                      </span>
                    </div>
                    <small v-if="isRatePriceActive(model) && model.pricing?.input_price != null">
                      基础 {{ formatBaseModelPrice(model.pricing?.input_price) }}
                    </small>
                  </td>
                  <td data-label="计费说明">
                    <span class="model-price-value" :class="{ 'is-rate-price': isRatePriceActive(model) }">
                      {{ formatSecondaryModelPrice(model) }}
                    </span>
                    <small v-if="shouldShowPricingNote(model)" class="model-price-note">
                      {{ pricingNote(model) }}
                    </small>
                    <small v-if="isRatePriceActive(model) && model.pricing?.output_price != null">
                      基础 {{ formatBaseModelPrice(model.pricing?.output_price) }}
                    </small>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </section>

      <section v-else class="model-empty-card">
        <Icon name="search" size="lg" />
        <p>{{ emptyStateMessage }}</p>
        <button type="button" @click="resetFilters">清空筛选</button>
      </section>

      <p class="model-plaza-note">
        官方价按厂商公开口径展示，单位可能不同；我们的价格为当前接入价按美元折算人民币，开启倍率后按当前分组倍率折算。实际扣费以控制台记录为准。
      </p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import PublicMatrixBackdrop from './components/PublicMatrixBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'
import userChannelsAPI, {
  type UserAvailableChannel,
  type UserAvailableGroup,
  type UserChannelPlatformSection,
  type UserSupportedModelPricing
} from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { BILLING_MODE_IMAGE, BILLING_MODE_PER_REQUEST, BILLING_MODE_TOKEN } from '@/constants/channel'
import type { Group } from '@/types'

interface ModelGroupMeta extends UserAvailableGroup {
  channelName: string
  effectiveRate: number
}

interface PlazaModel {
  key: string
  name: string
  platform: string
  pricing: UserSupportedModelPricing | null
  referencePricing: UserSupportedModelPricing | null
  groups: ModelGroupMeta[]
}

interface ProviderStat {
  platform: string
  label: string
  count: number
  sampleModel: string
}

interface PriceChipItem {
  label: string
  price: string
}

const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)

const loading = ref(false)
const searchQuery = ref('')
const selectedProvider = ref('all')
const selectedGroupByPlatform = ref<Record<string, string>>({})
const showRatePricesByPlatform = ref<Record<string, boolean>>({})
const channels = ref<UserAvailableChannel[]>([])
const availableGroups = ref<UserAvailableGroup[]>([])
const userGroupRates = ref<Record<number, number>>({})

const usingFallbackCatalog = computed(() => !isAuthenticated.value || channels.value.length === 0)
const sourceChannels = computed(() => (usingFallbackCatalog.value ? fallbackChannels : channels.value))

const modelCards = computed<PlazaModel[]>(() => {
  const map = new Map<string, PlazaModel>()

  for (const channel of sourceChannels.value) {
    for (const section of channel.platforms) {
      for (const supportedModel of section.supported_models) {
        if (shouldHideModelFromPlaza(supportedModel.name, supportedModel.platform || section.platform)) continue

        const key = `${section.platform}:${supportedModel.name}`.toLowerCase()
        const existing = map.get(key) ?? {
          key,
          name: supportedModel.name,
          platform: supportedModel.platform || section.platform,
          pricing: supportedModel.pricing,
          referencePricing: supportedModel.reference_pricing ?? null,
          groups: []
        }

        for (const group of groupsForCatalogSection(section)) {
          if (existing.groups.some((item) => item.id === group.id)) continue
          existing.groups.push({
            ...group,
            channelName: channel.name,
            effectiveRate: userGroupRates.value[group.id] ?? group.rate_multiplier ?? 1
          })
        }

        map.set(key, existing)
      }
    }
  }

  return Array.from(map.values()).sort(compareModelCardsByCost)
})

const providerStats = computed<ProviderStat[]>(() => {
  const stats = new Map<string, ProviderStat>()
  for (const model of modelCards.value) {
    const platform = model.platform || 'unknown'
    const current = stats.get(platform) ?? {
      platform,
      label: providerLabel(platform),
      count: 0,
      sampleModel: model.name
    }
    current.count += 1
    stats.set(platform, current)
  }
  return Array.from(stats.values()).sort((a, b) => a.label.localeCompare(b.label))
})

const providerTabs = computed(() => [
  { value: 'all', label: '全部', count: modelCards.value.length, sampleModel: '' },
  ...providerStats.value.map((item) => ({
    value: item.platform,
    label: item.label,
    count: item.count,
    sampleModel: item.sampleModel
  }))
])

const rateGroups = computed<UserAvailableGroup[]>(() => {
  const groups = new Map<number, UserAvailableGroup>()

  if (isAuthenticated.value) {
    for (const group of availableGroups.value) {
      groups.set(group.id, group)
    }

    for (const channel of channels.value) {
      for (const section of channel.platforms) {
        for (const group of section.groups) {
          if (!groups.has(group.id)) groups.set(group.id, group)
        }
      }
    }

    return Array.from(groups.values())
  }

  for (const channel of fallbackChannels) {
    for (const section of channel.platforms) {
      for (const group of section.groups) {
        if (!groups.has(group.id)) groups.set(group.id, group)
      }
    }
  }
  return Array.from(groups.values())
})

const sortedRateGroups = computed<UserAvailableGroup[]>(() =>
  [...rateGroups.value].sort(compareRateGroupsByCost)
)

const filteredModels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()

  return modelCards.value.filter((model) => {
    if (selectedProvider.value !== 'all' && model.platform !== selectedProvider.value) return false
    if (!query) return true

    return (
      model.name.toLowerCase().includes(query) ||
      providerLabel(model.platform).toLowerCase().includes(query)
    )
  })
})

const providerPricingSections = computed(() => {
  const sections = new Map<string, { platform: string; label: string; sampleModel: string; models: PlazaModel[] }>()
  for (const model of filteredModels.value) {
    const platform = model.platform || 'unknown'
    const section = sections.get(platform) ?? {
      platform,
      label: providerLabel(platform),
      sampleModel: model.name,
      models: []
    }
    section.models.push(model)
    sections.set(platform, section)
  }

  return Array.from(sections.values()).sort((a, b) => a.label.localeCompare(b.label))
})

const emptyStateMessage = computed(() => modelCards.value.length === 0 ? '暂无可用模型' : '没有找到匹配的模型')

async function loadChannels(force = false): Promise<void> {
  if (!isAuthenticated.value) {
    if (force) {
      channels.value = fallbackChannels
    }
    selectCheapestRateMode()
    return
  }

  loading.value = true
  try {
    const [list, groups, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getAvailable().catch(() => [] as Group[]),
      userGroupsAPI.getUserGroupRates().catch(() => ({} as Record<number, number>))
    ])
    channels.value = list
    availableGroups.value = groups.map(toAvailableGroup)
    userGroupRates.value = rates
  } catch (error) {
    console.warn('Failed to load model plaza data.', error)
    channels.value = []
    availableGroups.value = []
    userGroupRates.value = {}
  } finally {
    selectCheapestRateMode()
    loading.value = false
  }
}

function resetFilters(): void {
  searchQuery.value = ''
  selectedProvider.value = 'all'
  selectCheapestRateMode()
}

function platformKey(platform: string): string {
  return platform.trim().toLowerCase()
}

function selectedGroupKey(platform: string): string {
  return selectedGroupByPlatform.value[platformKey(platform)] ?? 'cheapest'
}

function updateProviderGroup(platform: string, value: string | number | boolean | null): void {
  const key = platformKey(platform)
  selectedGroupByPlatform.value = {
    ...selectedGroupByPlatform.value,
    [key]: String(value ?? 'cheapest')
  }

  if (!canUseRatePrices(platform)) {
    showRatePricesByPlatform.value = {
      ...showRatePricesByPlatform.value,
      [key]: false
    }
  }
}

function updateShowRatePrices(platform: string, value: boolean): void {
  const key = platformKey(platform)
  showRatePricesByPlatform.value = {
    ...showRatePricesByPlatform.value,
    [key]: canUseRatePrices(platform) ? value : false
  }
}

function showRatePricesForPlatform(platform: string): boolean {
  return showRatePricesByPlatform.value[platformKey(platform)] ?? true
}

function canUseRatePrices(platform: string): boolean {
  return sortedRateGroupsForPlatform(platform).length > 0 && selectedGroupKey(platform) !== 'all'
}

function shouldShowReferenceColumn(platform: string): boolean {
  return platformKey(platform) !== 'video'
}

function shouldShowPricingNote(model: PlazaModel): boolean {
  return shouldShowReferenceColumn(model.platform) && pricingNote(model) !== ''
}

function effectiveRate(model: PlazaModel): number | null {
  const groupKey = selectedGroupKey(model.platform)
  if (groupKey === 'cheapest') {
    const group = cheapestRateGroupForPlatform(model.platform)
    if (!group) return null

    const rate = effectiveGroupRate(group)
    return Number.isFinite(rate) && rate > 0 ? rate : null
  }

  const group = selectedRateGroupForPlatform(model.platform)
  if (!group) return null
  if (!samePlatform(group.platform, model.platform)) return null

  const rate = effectiveGroupRate(group)
  return Number.isFinite(rate) && rate > 0 ? rate : null
}

function formatRate(rate: number): string {
  return Number.isInteger(rate) ? String(rate) : rate.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}

function formatModelPrice(value: number | null | undefined, model: PlazaModel): string {
  if (value == null) return '-'
  const rate = showRatePricesForPlatform(model.platform) ? effectiveRate(model) ?? 1 : 1
  return formatPricePerMillion(value * rate)
}

function formatPrimaryModelPrice(model: PlazaModel): string {
  const perRequestPrice = model.pricing?.per_request_price
  if (model.pricing?.billing_mode === BILLING_MODE_PER_REQUEST && perRequestPrice != null) {
    const suffix = tierPriceItems(model).length > 0 ? '起' : ''
    return `${formatPerRequestPrice(perRequestPrice, model)}${suffix}`
  }

  return formatModelPrice(model.pricing?.input_price, model)
}

function formatSecondaryModelPrice(model: PlazaModel): string {
  if (model.pricing?.billing_mode === BILLING_MODE_PER_REQUEST) {
    if (tierPriceItems(model).length > 0) return `按规格${videoTierUnit()}计费`
    return model.pricing.intervals[0]?.tier_label || '单次计费'
  }

  return formatModelPrice(model.pricing?.output_price, model)
}

function referencePrice(model: PlazaModel): string {
  const officialVideoSummary = officialVideoReferenceSummary(model.name)
  if (officialVideoSummary) return officialVideoSummary

  if (model.referencePricing) {
    return formatReferencePricing(model.referencePricing)
  }

  return '-'
}

function officialVideoReferenceSummary(modelName: string): string {
  switch (modelName) {
    case 'kling-v3-omni':
      return formatKlingCreditRangePerSecond(6, 16)
    case 'kling-v2-6':
      return formatKlingCreditRangePerSecond(3, 10)
    case 'wan2.7':
      return '¥0.6-1/秒'
    case 'veo3.1-fast':
      return '$0.08-0.30/秒'
    case 'doubao-seedance-2.0':
      return '¥28-51/M tokens'
    default:
      return ''
  }
}

function officialReferenceItems(model: PlazaModel): PriceChipItem[] {
  switch (model.name) {
    case 'kling-v3-omni':
      return [
        { label: '720P 无音频', price: formatKlingCreditPerSecond(6) },
        { label: '1080P 无音频', price: formatKlingCreditPerSecond(8) },
        { label: '720P 有音频', price: formatKlingCreditPerSecond(9) },
        { label: '1080P 有音频', price: formatKlingCreditPerSecond(12) },
        { label: '720P 视频输入', price: formatKlingCreditPerSecond(12) },
        { label: '1080P 视频输入', price: formatKlingCreditPerSecond(16) }
      ]
    case 'kling-v2-6':
      return [
        { label: '标准无音频', price: formatKlingCreditPerSecond(3) },
        { label: '专业无音频', price: formatKlingCreditPerSecond(5) },
        { label: '专业有音频', price: formatKlingCreditPerSecond(10) },
        { label: '声音控制', price: formatKlingCreditPerSecond(2, '+') }
      ]
    case 'wan2.7':
      return [
        { label: '720P', price: '¥0.6/秒' },
        { label: '1080P', price: '¥1/秒' }
      ]
    case 'veo3.1-fast':
      return [
        { label: '720P 视频', price: '$0.08/秒' },
        { label: '1080P 视频', price: '$0.10/秒' },
        { label: '4K 视频', price: '$0.25/秒' },
        { label: '720P 视频+音频', price: '$0.10/秒' },
        { label: '1080P 视频+音频', price: '$0.12/秒' },
        { label: '4K 视频+音频', price: '$0.30/秒' }
      ]
    case 'doubao-seedance-2.0':
      return [
        { label: '含视频 480/720P', price: '¥28/M tokens' },
        { label: '无视频 480/720P', price: '¥46/M tokens' },
        { label: '含视频 1080P', price: '¥31/M tokens' },
        { label: '无视频 1080P', price: '¥51/M tokens' }
      ]
    default:
      return []
  }
}

const KLING_REFERENCE_RMB_PER_CREDIT = 0.098

function formatKlingCreditRangePerSecond(minCredits: number, maxCredits: number): string {
  return `${formatRmbAmount(minCredits * KLING_REFERENCE_RMB_PER_CREDIT)}-${formatRmbAmount(maxCredits * KLING_REFERENCE_RMB_PER_CREDIT)}/秒`
}

function formatKlingCreditPerSecond(credits: number, prefix = ''): string {
  return `${prefix}${formatRmbAmount(credits * KLING_REFERENCE_RMB_PER_CREDIT)}/秒`
}

function formatRmbAmount(value: number): string {
  return `¥${value.toFixed(3).replace(/0+$/, '').replace(/\.$/, '')}`
}

function formatReferencePricing(pricing: UserSupportedModelPricing): string {
  if (pricing.billing_mode === BILLING_MODE_PER_REQUEST && pricing.per_request_price != null) {
    return formatReferencePerRequestPrice(pricing)
  }

  if (pricing.billing_mode === BILLING_MODE_IMAGE && pricing.image_output_price != null) {
    return `图片 ${formatPricePerMillion(pricing.image_output_price)}`
  }

  const parts: string[] = []
  if (pricing.input_price != null) parts.push(`输入 ${formatPricePerMillion(pricing.input_price)}`)
  if (pricing.output_price != null) parts.push(`输出 ${formatPricePerMillion(pricing.output_price)}`)
  if (parts.length > 0) return parts.join(' / ')

  if (pricing.per_request_price != null) return formatReferencePerRequestPrice(pricing)
  return '-'
}

function formatReferencePerRequestPrice(pricing: UserSupportedModelPricing): string {
  const unit = pricing.intervals[0]?.tier_label || '/次'
  const value = pricing.per_request_price
  if (value == null) return '-'
  return formatRmbReferencePrice(value, unit)
}

function formatRmbReferencePrice(usdPrice: number, unit: string): string {
  const rmbPrice = usdPrice * 7
  const digits = rmbPrice > 0 && rmbPrice < 1 ? 3 : 2
  return `¥${rmbPrice.toFixed(digits).replace(/0+$/, '').replace(/\.$/, '')}${unit}`
}

function pricingNote(model: PlazaModel): string {
  switch (model.name) {
    case 'kling-v3-omni':
      return '可灵官方 Credit 消耗折算，按 ¥0.098/Credit 估算'
    case 'kling-v2-6':
      return '可灵官方 Credit 消耗折算，声音控制另加约 ¥0.196/秒'
    case 'wan2.7':
      return '阿里百炼中国内地官方价，国际地域另计'
    case 'veo3.1-fast':
      return 'Google Vertex AI 官方美元秒价'
    case 'doubao-seedance-2.0':
      return '火山方舟按 tokens 计费，任务价随时长/分辨率变化'
    default:
      return ''
  }
}

function formatBaseModelPrice(value: number | null | undefined): string {
  if (value == null) return '-'
  return formatPricePerMillion(value)
}

function formatPricePerMillion(value: number): string {
  const perMillion = value * 1_000_000 * 7
  const digits = perMillion > 0 && perMillion < 1 ? 3 : 2
  return `¥${perMillion.toFixed(digits)}/M`
}

function formatPerRequestPrice(value: number, model: PlazaModel): string {
  const rate = showRatePricesForPlatform(model.platform) ? effectiveRate(model) ?? 1 : 1
  const unit = tierPriceItems(model).length > 0 ? videoTierUnit() : model.pricing?.intervals[0]?.tier_label || '/次'
  const rmbPrice = value * rate * 7
  return `¥${rmbPrice.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')}${unit}`
}

function tierPriceItems(model: PlazaModel): Array<{ label: string; price: string }> {
  if (model.pricing?.billing_mode !== BILLING_MODE_PER_REQUEST) return []
  if (model.pricing.intervals.length <= 1) return []

  const rate = showRatePricesForPlatform(model.platform) ? effectiveRate(model) ?? 1 : 1
  return model.pricing.intervals
    .filter((interval) => interval.tier_label && interval.per_request_price != null)
    .map((interval) => ({
      label: interval.tier_label || '默认',
      price: formatRmbPrice(interval.per_request_price ?? 0, rate, videoTierUnit())
    }))
}

function formatRmbPrice(value: number, rate: number, unit: string): string {
  const rmbPrice = value * rate * 7
  return `¥${rmbPrice.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')}${unit}`
}

function videoTierUnit(): string {
  return '/秒'
}

function modelBasePriceScore(model: PlazaModel): number {
  const inputPrice = model.pricing?.input_price
  const outputPrice = model.pricing?.output_price
  const cacheWritePrice = model.pricing?.cache_write_price ?? 0
  const cacheReadPrice = model.pricing?.cache_read_price ?? 0
  const knownPrices = [inputPrice, outputPrice].filter((value): value is number => value != null)
  if (knownPrices.length === 0) return Number.POSITIVE_INFINITY

  return knownPrices.reduce((sum, value) => sum + value, 0) + cacheWritePrice + cacheReadPrice
}

function isRatePriceActive(model: PlazaModel): boolean {
  return showRatePricesForPlatform(model.platform) && selectedGroupKey(model.platform) !== 'all' && effectiveRate(model) != null
}

function effectiveGroupRate(group: UserAvailableGroup): number {
  return userGroupRates.value[group.id] ?? group.rate_multiplier ?? 1
}

function groupOptionsForPlatform(platform: string): Array<{ value: string; label: string }> {
  const groups = sortedRateGroupsForPlatform(platform)
  return [
    { value: 'cheapest', label: '最低倍率优先' },
    { value: 'all', label: '不使用倍率' },
    ...groups.map((group, index) => ({
      value: String(group.id),
      label: `${group.name} · x${formatRate(effectiveGroupRate(group))}${index === 0 ? ' · 最低倍率' : ''}`
    }))
  ]
}

function sortedRateGroupsForPlatform(platform: string): UserAvailableGroup[] {
  return sortedRateGroups.value.filter((group) => samePlatform(group.platform, platform))
}

function selectedRateGroupForPlatform(platform: string): UserAvailableGroup | null {
  const groupId = selectedGroupKey(platform)
  if (groupId === 'all' || groupId === 'cheapest') return null
  return sortedRateGroupsForPlatform(platform).find((group) => String(group.id) === groupId) ?? null
}

function compareRateGroupsByCost(a: UserAvailableGroup, b: UserAvailableGroup): number {
  const rateCompare = effectiveGroupRate(a) - effectiveGroupRate(b)
  if (rateCompare !== 0) return rateCompare

  const providerCompare = providerLabel(a.platform).localeCompare(providerLabel(b.platform))
  return providerCompare || a.name.localeCompare(b.name)
}

function cheapestRateGroupForPlatform(platform: string): UserAvailableGroup | null {
  return sortedRateGroupsForPlatform(platform)[0] ?? null
}

function cheapestPlatformRate(platform: string): number {
  const group = cheapestRateGroupForPlatform(platform)
  return group ? effectiveGroupRate(group) : Number.POSITIVE_INFINITY
}

function modelEffectivePriceScore(model: PlazaModel): number {
  const rate = selectedGroupKey(model.platform) === 'all' ? 1 : cheapestPlatformRate(model.platform)
  return modelBasePriceScore(model) * (Number.isFinite(rate) ? rate : 1)
}

function modelVersionScore(modelName: string): number {
  const normalized = modelName.toLowerCase()
  const familyMatch = normalized.match(/^(claude|gpt|gemini)-[a-z]+-(\d+(?:[.-]\d+)*)/)
  if (!familyMatch) return 0

  return familyMatch[2]
    .split(/[.-]/)
    .reduce((score, part) => score * 100 + Number(part || 0), 0)
}

function isOpenAIModel(model: PlazaModel): boolean {
  return providerLabel(model.platform) === 'Openai' || model.name.toLowerCase().startsWith('gpt-')
}

function compareModelCardsByCost(a: PlazaModel, b: PlazaModel): number {
  const rateCompare = cheapestPlatformRate(a.platform) - cheapestPlatformRate(b.platform)
  if (rateCompare !== 0) return rateCompare

  const aPrice = modelEffectivePriceScore(a)
  const bPrice = modelEffectivePriceScore(b)
  const priceCompare = isOpenAIModel(a) && isOpenAIModel(b) ? bPrice - aPrice : aPrice - bPrice
  if (priceCompare !== 0) return priceCompare

  const versionCompare = modelVersionScore(b.name) - modelVersionScore(a.name)
  if (versionCompare !== 0) return versionCompare

  const providerCompare = providerLabel(a.platform).localeCompare(providerLabel(b.platform))
  return providerCompare || a.name.localeCompare(b.name)
}

function modelAvailabilityLabel(model: PlazaModel): string {
  const count = model.groups.length
  if (count <= 0) return '基础目录'
  return `${count} 个可用分组`
}

function modelVariantDescription(model: PlazaModel): string {
  switch (model.name) {
    case 'doubao-seedance-2.0':
      return '标准版：完整视频生成能力，质量优先，支持 1080P'
    case 'doubao-seedance-2.0-fast':
      return '快速版：同标准能力，出片更快，适合草稿和批量试错'
    case 'doubao-seedance-2.0-fast-face':
      return '快速真人版：快速版 + 真人上传/人像素材，最高 720P'
    case 'doubao-seedance-2.0-face':
      return '真人版：标准版 + 真人上传/人像素材，支持 1080P'
    default:
      return ''
  }
}

function providerDescription(platform: string): string {
  const label = providerLabel(platform).toLowerCase()
  if (label === 'anthropic') return 'Claude 长上下文、推理与编码模型'
  if (label === 'openai') return 'GPT、Codex 与 OpenAI 兼容模型'
  if (label === 'gemini') return 'Gemini 多模态与长上下文模型'
  return '兼容模型与自定义渠道'
}

function shouldHideModelFromPlaza(modelName: string, platform: string): boolean {
  if (providerLabel(platform).toLowerCase() !== 'openai') return false

  const normalized = modelName.trim().toLowerCase()
  return normalized === 'gpt-5.2' ||
    normalized.startsWith('gpt-5.2-') ||
    normalized === 'gpt-5.3' ||
    normalized.startsWith('gpt-5.3-')
}

function selectCheapestRateMode(): void {
  const selected: Record<string, string> = {}
  const visiblePlatforms = new Set(modelCards.value.map((model) => platformKey(model.platform)))

  for (const platform of visiblePlatforms) {
    selected[platform] = sortedRateGroupsForPlatform(platform).length === 0 ? 'all' : 'cheapest'
  }

  selectedGroupByPlatform.value = selected
  showRatePricesByPlatform.value = Object.fromEntries(
    Array.from(visiblePlatforms).map((platform) => [platform, selected[platform] !== 'all'])
  )
}

function samePlatform(a: string, b: string): boolean {
  return a.trim().toLowerCase() === b.trim().toLowerCase()
}

function groupsForCatalogSection(section: UserChannelPlatformSection): UserAvailableGroup[] {
  if (!usingFallbackCatalog.value || !isAuthenticated.value) return section.groups

  const platformGroups = availableGroups.value.filter((group) => samePlatform(group.platform, section.platform))
  return platformGroups.length > 0 ? platformGroups : section.groups
}

function toAvailableGroup(group: Group): UserAvailableGroup {
  return {
    id: group.id,
    name: group.name,
    platform: group.platform,
    subscription_type: group.subscription_type,
    rate_multiplier: group.rate_multiplier,
    is_exclusive: group.is_exclusive
  }
}

function providerLabel(platform: string): string {
  const normalized = platform.toLowerCase()
  if (normalized.includes('anthropic') || normalized.includes('claude')) return 'Anthropic'
  if (normalized.includes('openai') || normalized.includes('gpt')) return 'OpenAI'
  if (normalized.includes('gemini') || normalized.includes('google')) return 'Gemini'
  return platform ? platform.charAt(0).toUpperCase() + platform.slice(1) : 'Other'
}

function providerToneClass(platform: string): string {
  const label = providerLabel(platform).toLowerCase()
  if (label === 'anthropic') return 'model-stat-icon-anthropic'
  if (label === 'openai') return 'model-stat-icon-openai'
  if (label === 'gemini') return 'model-stat-icon-gemini'
  return 'model-stat-icon-total'
}

function tokenPricing(
  inputPricePerMillion: number,
  outputPricePerMillion: number,
  cacheWritePerMillion = 0,
  cacheReadPerMillion = 0
): UserSupportedModelPricing {
  return {
    billing_mode: BILLING_MODE_TOKEN,
    input_price: inputPricePerMillion / 1_000_000,
    output_price: outputPricePerMillion / 1_000_000,
    cache_write_price: cacheWritePerMillion / 1_000_000,
    cache_read_price: cacheReadPerMillion / 1_000_000,
    image_output_price: null,
    per_request_price: null,
    intervals: []
  }
}

function tokenReferencePricing(
  inputPricePerMillion: number,
  outputPricePerMillion: number,
  cacheWritePerMillion = 0,
  cacheReadPerMillion = 0
): UserSupportedModelPricing {
  return tokenPricing(inputPricePerMillion, outputPricePerMillion, cacheWritePerMillion, cacheReadPerMillion)
}

function tierPricing(tiers: Array<{ label: string; price: number }>): UserSupportedModelPricing {
  const [firstTier] = tiers
  return {
    billing_mode: BILLING_MODE_PER_REQUEST,
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    image_output_price: null,
    per_request_price: firstTier?.price ?? null,
    intervals: tiers.map((tier) => ({
      min_tokens: 0,
      max_tokens: null,
      tier_label: tier.label,
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      per_request_price: tier.price
    }))
  }
}

function withReferencePricing(
  name: string,
  platform: string,
  pricing: UserSupportedModelPricing,
  referencePricing: UserSupportedModelPricing
) {
  return {
    name,
    platform,
    pricing,
    reference_pricing: referencePricing
  }
}

const openaiGroups: UserAvailableGroup[] = [
  { id: 101, name: 'CC 原生直转渠道', platform: 'openai', subscription_type: 'standard', rate_multiplier: 1.5, is_exclusive: false },
  { id: 102, name: 'CodeX 官转 pro号池', platform: 'openai', subscription_type: 'standard', rate_multiplier: 0.35, is_exclusive: false },
  { id: 103, name: 'CC AWS 逆向', platform: 'openai', subscription_type: 'standard', rate_multiplier: 0.5, is_exclusive: false },
  { id: 104, name: 'CodeX team/plus号池', platform: 'openai', subscription_type: 'standard', rate_multiplier: 0.15, is_exclusive: false }
]

const anthropicGroups: UserAvailableGroup[] = [
  { id: 201, name: 'Anthropic 原生官转渠道', platform: 'anthropic', subscription_type: 'standard', rate_multiplier: 1.5, is_exclusive: false },
  { id: 202, name: 'Claude Code 专线', platform: 'anthropic', subscription_type: 'standard', rate_multiplier: 0.5, is_exclusive: false }
]

const videoGroups: UserAvailableGroup[] = [
  { id: 301, name: '视频模型标准渠道', platform: 'video', subscription_type: 'standard', rate_multiplier: 1, is_exclusive: false }
]

const fallbackChannels: UserAvailableChannel[] = [
  {
    name: 'Openai',
    description: 'ChatGPT 与 Codex 常用模型',
    platforms: [
      {
        platform: 'openai',
        groups: openaiGroups,
        supported_models: [
          withReferencePricing('gpt-5.5', 'openai', tokenPricing(5, 30, 0, 0.5), tokenReferencePricing(2.5, 15, 0, 0.25)),
          withReferencePricing('gpt-5.4', 'openai', tokenPricing(2.5, 15, 0, 0.25), tokenReferencePricing(2.5, 15, 0, 0.25)),
          withReferencePricing('gpt-5.4-mini', 'openai', tokenPricing(0.75, 4.5, 0, 0.075), tokenReferencePricing(0.75, 4.5, 0, 0.075))
        ]
      }
    ]
  },
  {
    name: 'Anthropic',
    description: 'Claude 长上下文与编码模型',
    platforms: [
      {
        platform: 'anthropic',
        groups: anthropicGroups,
        supported_models: [
          withReferencePricing('claude-opus-4.8', 'anthropic', tokenPricing(5, 25, 6.25, 0.5), tokenReferencePricing(5, 25, 6.25, 0.5)),
          withReferencePricing('claude-opus-4.7', 'anthropic', tokenPricing(5, 25, 6.25, 0.5), tokenReferencePricing(5, 25, 6.25, 0.5)),
          withReferencePricing('claude-opus-4.6', 'anthropic', tokenPricing(5, 25, 6.25, 0.5), tokenReferencePricing(5, 25, 6.25, 0.5)),
          withReferencePricing('claude-opus-4.5', 'anthropic', tokenPricing(5, 25, 6.25, 0.5), tokenReferencePricing(5, 25, 6.25, 0.5)),
          withReferencePricing('claude-opus-4.1', 'anthropic', tokenPricing(5, 25, 6.25, 0.5), tokenReferencePricing(5, 25, 6.25, 0.5)),
          withReferencePricing('claude-opus-4', 'anthropic', tokenPricing(5, 25, 6.25, 0.5), tokenReferencePricing(5, 25, 6.25, 0.5)),
          withReferencePricing('claude-sonnet-4.6', 'anthropic', tokenPricing(3, 15, 3.75, 0.3), tokenReferencePricing(3, 15, 3.75, 0.3)),
          withReferencePricing('claude-sonnet-4.5', 'anthropic', tokenPricing(3, 15, 3.75, 0.3), tokenReferencePricing(3, 15, 3.75, 0.3)),
          withReferencePricing('claude-sonnet-4', 'anthropic', tokenPricing(3, 15, 3.75, 0.3), tokenReferencePricing(3, 15, 3.75, 0.3)),
          withReferencePricing('claude-haiku-4.5', 'anthropic', tokenPricing(1, 5, 1.25, 0.1), tokenReferencePricing(1, 5, 1.25, 0.1)),
          withReferencePricing('claude-haiku-4', 'anthropic', tokenPricing(1, 5, 1.25, 0.1), tokenReferencePricing(1, 5, 1.25, 0.1))
        ]
      }
    ]
  },
  {
    name: 'Video',
    description: '图片与视频生成模型',
    platforms: [
      {
        platform: 'video',
        groups: videoGroups,
        supported_models: [
          {
            name: 'kling-v3-omni',
            platform: 'video',
            pricing: tierPricing([
              { label: 'default', price: 0.0672 },
              { label: 'pro', price: 0.0896 },
              { label: 'sound', price: 0.0896 },
              { label: 'video', price: 0.1008 },
              { label: 'pro-sound', price: 0.112 },
              { label: 'pro-video', price: 0.1344 },
              { label: '4k', price: 0.42856 },
              { label: '4k-sound', price: 0.42856 }
            ])
          },
          {
            name: 'kling-v2-6',
            platform: 'video',
            pricing: tierPricing([
              { label: 'default', price: 0.0368 },
              { label: 'pro', price: 0.0625 },
              { label: 'pro-sound', price: 0.125 },
              { label: 'pro-sound-voice', price: 0.15 }
            ])
          },
          {
            name: 'wan2.7',
            platform: 'video',
            pricing: tierPricing([
              { label: 'default', price: 0.0664 },
              { label: '1080P', price: 0.1096 }
            ])
          },
          {
            name: 'veo3.1-fast',
            platform: 'video',
            pricing: tierPricing([
              { label: 'default', price: 0.18 },
              { label: 'extend', price: 0.08 },
              { label: '4K', price: 0.24 },
              { label: 'EXTEND-4K', price: 0.24 }
            ])
          },
          {
            name: 'doubao-seedance-2.0',
            platform: 'video',
            pricing: tierPricing([
              { label: '480P-input', price: 0.044 },
              { label: '480P', price: 0.07256 },
              { label: '720P-input', price: 0.0944 },
              { label: '720P', price: 0.15616 },
              { label: '1080P-input', price: 0.2136 },
              { label: '1080P', price: 0.352 }
            ])
          },
          {
            name: 'doubao-seedance-2.0-fast',
            platform: 'video',
            pricing: tierPricing([
              { label: '480P-input', price: 0.0348 },
              { label: '480P', price: 0.0584 },
              { label: '720P-input', price: 0.0752 },
              { label: '720P', price: 0.1256 }
            ])
          },
          {
            name: 'doubao-seedance-2.0-fast-face',
            platform: 'video',
            pricing: tierPricing([
              { label: '480P-input', price: 0.048 },
              { label: '480P', price: 0.08 },
              { label: '720P-input', price: 0.1032 },
              { label: '720P', price: 0.172 }
            ])
          },
          {
            name: 'doubao-seedance-2.0-face',
            platform: 'video',
            pricing: tierPricing([
              { label: '480P-input', price: 0.06 },
              { label: '480P', price: 0.0992 },
              { label: '720P-input', price: 0.1288 },
              { label: '720P', price: 0.2136 },
              { label: '1080P-input', price: 0.3 },
              { label: '1080P', price: 0.5 }
            ])
          }
        ]
      }
    ]
  }
]

onMounted(() => {
  void loadChannels()
})
</script>

<style scoped>
@import './public-page.css';

.model-plaza-page {
  background:
    radial-gradient(circle at 50% 18%, rgba(32, 170, 92, 0.22) 0, transparent 34%),
    radial-gradient(circle at 18% 24%, rgba(87, 86, 210, 0.16) 0, transparent 30%),
    radial-gradient(circle at 82% 36%, rgba(45, 178, 105, 0.14) 0, transparent 32%),
    linear-gradient(180deg, #050914 0%, #08110f 48%, #03060a 100%);
}

.model-plaza-main {
  width: min(100%, 92rem);
  padding: 1.2rem 1rem 3rem;
}

.model-plaza-title {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.25rem 0 2.2rem;
}

.model-plaza-title h1 {
  color: rgba(255, 255, 255, 0.96);
  font-size: clamp(1.65rem, 2.6vw, 2.45rem);
  font-weight: 950;
  line-height: 1;
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

.model-plaza-title p {
  margin-top: 0.5rem;
  color: rgba(238, 246, 240, 0.68);
  font-size: 0.82rem;
  font-weight: 700;
  max-width: 44rem;
}

.model-refresh-button {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 8px;
  border: 1px solid var(--public-border-strong);
  background: var(--public-surface-soft);
  box-shadow: var(--public-shadow-soft);
  padding: 0.65rem 0.9rem;
  color: rgba(238, 246, 240, 0.84);
  font-size: 0.82rem;
  font-weight: 800;
  backdrop-filter: blur(18px);
}

.model-stat-icon {
  display: inline-flex;
  height: 2.65rem;
  width: 2.65rem;
  flex: 0 0 2.65rem;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.model-stat-icon-total {
  background: rgba(139, 92, 246, 0.18);
  color: #b79cff;
}

.model-stat-icon-openai {
  background: rgba(39, 220, 132, 0.16);
}

.model-stat-icon-anthropic {
  background: rgba(249, 115, 22, 0.16);
}

.model-stat-icon-gemini {
  background: rgba(56, 189, 248, 0.16);
}

.model-filter-panel {
  margin-top: 0;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: var(--public-surface-raised);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    var(--public-shadow);
  padding: 1rem;
  backdrop-filter: blur(18px);
}

.model-search-box {
  display: flex;
  min-height: 2.7rem;
  align-items: center;
  gap: 0.7rem;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: var(--public-input-surface);
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

.model-provider-rate-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  min-width: 0;
  gap: 0.7rem;
}

.model-rate-group-label {
  color: rgba(238, 246, 240, 0.7);
  font-size: 0.8rem;
  font-weight: 850;
  white-space: nowrap;
}

.model-group-select {
  width: min(18rem, 42vw);
}

.model-provider-group-select {
  width: min(18rem, 28vw);
}

.model-rate-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.48rem;
  color: rgba(238, 246, 240, 0.76);
  font-size: 0.82rem;
  font-weight: 800;
  white-space: nowrap;
}

.model-rate-toggle.is-disabled {
  cursor: not-allowed;
  color: rgba(222, 232, 255, 0.42);
}

.model-rate-toggle.is-disabled :deep(button) {
  pointer-events: none;
  opacity: 0.48;
}

.model-provider-tabs {
  margin-top: 1.2rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.model-provider-tabs button {
  display: inline-flex;
  align-items: center;
  gap: 0.46rem;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: var(--public-surface-soft);
  box-shadow: var(--public-shadow-soft);
  padding: 0.55rem 0.85rem;
  color: rgba(238, 246, 240, 0.72);
  font-weight: 850;
  backdrop-filter: blur(14px);
}

.model-provider-tabs button.is-active {
  border-color: rgba(119, 255, 173, 0.34);
  background: var(--public-accent-soft);
  color: rgba(220, 255, 230, 0.94);
}

.model-provider-tabs small {
  color: rgba(222, 232, 255, 0.68);
  font-size: 0.75rem;
  font-weight: 800;
}

.model-pricing-stack,
.model-card-grid {
  margin-top: 1.45rem;
  display: grid;
  gap: 1.1rem;
}

.model-provider-section {
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: var(--public-surface-raised);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    var(--public-shadow);
  backdrop-filter: blur(18px);
}

.model-provider-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(221, 230, 255, 0.12);
  padding: 1rem 1.1rem;
}

.model-provider-title {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.8rem;
  flex: 1 1 18rem;
}

.model-provider-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.65rem;
  height: 2.65rem;
  flex: 0 0 auto;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.model-provider-title h2 {
  overflow: hidden;
  color: rgba(255, 255, 255, 0.95);
  font-size: 1.05rem;
  font-weight: 950;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-provider-title p {
  margin-top: 0.16rem;
  color: rgba(222, 232, 255, 0.72);
  font-size: 0.78rem;
  font-weight: 750;
}

.model-provider-meta {
  display: grid;
  min-width: 4.8rem;
  flex: 0 0 auto;
  justify-items: end;
  color: rgba(222, 232, 255, 0.62);
  font-size: 0.68rem;
  font-weight: 900;
  text-transform: uppercase;
}

.model-provider-meta strong {
  color: rgba(255, 255, 255, 0.94);
  font-size: 1.45rem;
  font-weight: 950;
  line-height: 1;
}

.model-pricing-table-wrap {
  overflow-x: auto;
}

.model-pricing-table {
  width: 100%;
  min-width: 58rem;
  border-collapse: collapse;
}

.model-pricing-table th,
.model-pricing-table td {
  border-bottom: 1px solid rgba(221, 230, 255, 0.1);
  padding: 0.86rem 1rem;
  text-align: left;
  vertical-align: middle;
}

.model-pricing-table th {
  background: rgba(2, 8, 12, 0.32);
  color: rgba(222, 232, 255, 0.58);
  font-size: 0.72rem;
  font-weight: 900;
}

.model-pricing-table tbody tr:hover {
  background: rgba(119, 255, 173, 0.055);
}

.model-pricing-table tbody tr:last-child td {
  border-bottom: 0;
}

.model-name-cell {
  display: flex;
  align-items: flex-start;
  gap: 0.72rem;
  min-width: 17rem;
}

.model-name-cell strong {
  display: block;
  overflow: hidden;
  color: rgba(255, 255, 255, 0.94);
  font-size: 0.9rem;
  font-weight: 950;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-name-cell span {
  display: block;
  margin-top: 0.16rem;
  color: rgba(222, 232, 255, 0.58);
  font-size: 0.72rem;
  font-weight: 800;
}

.model-name-cell small {
  display: block;
  max-width: 18rem;
  margin-top: 0.24rem;
  color: rgba(125, 255, 170, 0.72);
  font-size: 0.72rem;
  font-weight: 800;
  line-height: 1.4;
}

.model-price-value {
  display: block;
  color: rgba(255, 255, 255, 0.92);
  font-size: 0.9rem;
  font-weight: 950;
  white-space: nowrap;
}

.model-price-value.is-rate-price {
  color: #7dffaa;
}

.model-price-value.muted {
  color: rgba(222, 232, 255, 0.7);
}

.model-tier-price-list {
  display: grid;
  width: min(100%, 24rem);
  gap: 0.36rem;
  margin-top: 0.58rem;
}

.model-tier-price-row {
  display: grid;
  grid-template-columns: minmax(7.5rem, 1fr) auto;
  align-items: center;
  gap: 0.72rem;
  border-radius: 8px;
  border: 1px solid rgba(221, 230, 255, 0.1);
  background: rgba(2, 8, 12, 0.34);
  padding: 0.34rem 0.5rem;
}

.model-tier-price-row strong,
.model-tier-price-row em {
  font-size: 0.72rem;
  line-height: 1.15;
}

.model-tier-price-row strong {
  overflow: hidden;
  color: rgba(222, 232, 255, 0.72);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-tier-price-row em {
  color: #7dffaa;
  font-style: normal;
  font-weight: 950;
  white-space: nowrap;
}

.model-pricing-table small {
  display: block;
  margin-top: 0.16rem;
  color: rgba(222, 232, 255, 0.68);
  font-size: 0.68rem;
  font-weight: 800;
  white-space: nowrap;
}

.model-price-note {
  max-width: 13rem;
  color: rgba(251, 191, 36, 0.82) !important;
  line-height: 1.35;
  white-space: normal !important;
}

.model-empty-card {
  margin-top: 1.5rem;
  display: grid;
  place-items: center;
  gap: 0.7rem;
  min-height: 16rem;
  border-radius: 8px;
  border: 1px dashed var(--public-border-strong);
  background: var(--public-surface-soft);
  color: rgba(238, 246, 240, 0.72);
  text-align: center;
  backdrop-filter: blur(18px);
}

.model-empty-card button {
  border: 1px solid rgba(119, 255, 173, 0.34);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.22), rgba(20, 184, 166, 0.1)),
    rgba(5, 15, 18, 0.82);
  padding: 0.52rem 0.8rem;
  color: #eafff0;
  font-size: 0.8rem;
  font-weight: 850;
}

.model-plaza-note {
  padding: 1.35rem 0 0.25rem;
  color: rgba(222, 232, 255, 0.68);
  text-align: center;
  font-size: 0.78rem;
  font-weight: 700;
}

:deep(.select-trigger) {
  min-height: 2.7rem;
  border-color: var(--public-border);
  background: var(--public-input-surface);
  color: rgba(255, 255, 255, 0.9);
}

:deep(.select-trigger:hover),
:deep(.select-trigger-open) {
  border-color: rgba(119, 255, 173, 0.36);
  background: rgba(5, 15, 18, 0.74);
}

@media (max-width: 900px) {
  .model-plaza-title {
    align-items: stretch;
    flex-direction: column;
    padding-top: 0.5rem;
  }

  .model-group-select {
    width: 100%;
  }

  .model-rate-toggle {
    justify-content: flex-end;
  }

  .model-provider-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .model-provider-title {
    flex-basis: auto;
  }

  .model-provider-rate-actions {
    width: 100%;
    align-items: center;
    flex-wrap: wrap;
    justify-content: flex-start;
    gap: 0.55rem;
  }

  .model-provider-group-select {
    width: min(12rem, 100%);
    flex: 1 1 10rem;
  }

  .model-provider-meta {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    width: 100%;
  }

  .model-pricing-table {
    min-width: 0;
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
    border-bottom: 1px solid rgba(221, 230, 255, 0.12);
    padding: 0.75rem 0;
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
    color: rgba(222, 232, 255, 0.54);
    font-size: 0.72rem;
    font-weight: 900;
  }

  .model-pricing-table td[data-label="模型"] {
    display: block;
    text-align: left;
  }

  .model-pricing-table td[data-label="模型"]::before {
    display: none;
  }

  .model-name-cell {
    min-width: 0;
  }

  .model-name-cell strong {
    white-space: normal;
  }

  .model-tier-price-list {
    width: min(100%, 22rem);
    margin-left: auto;
  }

  .model-tier-price-row {
    grid-template-columns: minmax(0, 1fr) auto;
  }
}
</style>
