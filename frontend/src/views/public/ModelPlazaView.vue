<template>
  <div class="model-plaza-page relative min-h-screen overflow-hidden text-white">
    <PublicMatrixBackdrop />

    <PublicTopNav />

    <main class="model-plaza-main relative z-10 mx-auto">
      <section class="model-plaza-title">
        <div>
          <h1>模型广场</h1>
          <p>浏览可用的 AI 模型及其定价</p>
        </div>
        <button class="model-refresh-button" type="button" :disabled="loading" @click="loadChannels(true)">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          刷新
        </button>
      </section>

      <section class="model-stat-grid">
        <article class="model-stat-card">
          <span class="model-stat-icon model-stat-icon-total">
            <Icon name="cube" size="lg" />
          </span>
          <div>
            <p>模型总数</p>
            <strong>{{ modelCards.length }}</strong>
          </div>
        </article>

        <article
          v-for="provider in providerStats"
          :key="provider.platform"
          class="model-stat-card"
        >
          <span class="model-stat-icon" :class="providerToneClass(provider.platform)">
            <ModelIcon :model="provider.sampleModel" size="22px" />
          </span>
          <div>
            <p>{{ provider.label }}</p>
            <strong>{{ provider.count }}</strong>
          </div>
        </article>
      </section>

      <section class="model-filter-panel">
        <div class="model-search-box">
          <Icon name="search" size="sm" />
          <input v-model="searchQuery" type="search" placeholder="搜索模型..." />
        </div>

        <div class="model-filter-actions">
          <Icon name="cube" size="sm" class="text-slate-400" />
          <span class="model-rate-group-label">倍率分组</span>
          <Select
            v-model="selectedGroupId"
            :options="groupOptions"
            class="model-group-select"
            :searchable="groupOptions.length > 8"
          />
          <label class="model-rate-toggle" :class="{ 'is-disabled': !canUseRatePrices }">
            <Toggle
              :model-value="showRatePrices && canUseRatePrices"
              @update:model-value="updateShowRatePrices"
            />
            <span>{{ canUseRatePrices ? '显示倍率价格' : '暂无倍率分组' }}</span>
          </label>
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

      <section v-if="filteredModels.length > 0" class="model-card-grid">
        <article v-for="model in filteredModels" :key="model.key" class="model-price-card">
          <div class="model-card-head">
            <ModelIcon :model="model.name" size="28px" />
            <div class="min-w-0">
              <h2>{{ model.name }}</h2>
              <p>{{ providerLabel(model.platform) }}</p>
            </div>
          </div>

          <div v-if="hasPromptCaching(model)" class="model-cache-badge">
            <Icon name="sparkles" size="xs" />
            Prompt Caching
          </div>

          <dl class="model-price-list">
            <div>
              <dt><Icon name="upload" size="xs" />输入</dt>
              <dd :class="{ 'is-rate-price': isRatePriceActive(model) }">
                <span>{{ formatModelPrice(model.pricing?.input_price, model) }}</span>
                <small v-if="isRatePriceActive(model)">
                  基础 {{ formatBaseModelPrice(model.pricing?.input_price) }} · {{ effectiveRateLabel(model) }}
                </small>
              </dd>
            </div>
            <div>
              <dt><Icon name="download" size="xs" />输出</dt>
              <dd :class="{ 'is-rate-price': isRatePriceActive(model) }">
                <span>{{ formatModelPrice(model.pricing?.output_price, model) }}</span>
                <small v-if="isRatePriceActive(model)">
                  基础 {{ formatBaseModelPrice(model.pricing?.output_price) }} · {{ effectiveRateLabel(model) }}
                </small>
              </dd>
            </div>
            <div class="model-price-muted">
              <dt><Icon name="document" size="xs" />缓存写入</dt>
              <dd>
                <span>{{ formatModelPrice(model.pricing?.cache_write_price, model) }}</span>
              </dd>
            </div>
            <div class="model-price-muted">
              <dt><Icon name="sync" size="xs" />缓存读取</dt>
              <dd>
                <span>{{ formatModelPrice(model.pricing?.cache_read_price, model) }}</span>
              </dd>
            </div>
          </dl>
        </article>
      </section>

      <section v-else class="model-empty-card">
        <Icon name="search" size="lg" />
        <p>{{ emptyStateMessage }}</p>
        <button type="button" @click="resetFilters">清空筛选</button>
      </section>

      <p class="model-plaza-note">
        价格以每百万 token 展示；开启倍率后按当前分组倍率折算。实际扣费以控制台记录为准。
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
  type UserSupportedModelPricing
} from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { BILLING_MODE_TOKEN } from '@/constants/channel'
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
  groups: ModelGroupMeta[]
}

interface ProviderStat {
  platform: string
  label: string
  count: number
  sampleModel: string
}

const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)

const loading = ref(false)
const searchQuery = ref('')
const selectedProvider = ref('all')
const selectedGroupId = ref<string | number | boolean | null>(null)
const showRatePrices = ref(true)
const channels = ref<UserAvailableChannel[]>([])
const availableGroups = ref<UserAvailableGroup[]>([])
const userGroupRates = ref<Record<number, number>>({})

const sourceChannels = computed(() => isAuthenticated.value ? channels.value : fallbackChannels)

const modelCards = computed<PlazaModel[]>(() => {
  const map = new Map<string, PlazaModel>()

  for (const channel of sourceChannels.value) {
    for (const section of channel.platforms) {
      for (const supportedModel of section.supported_models) {
        const key = `${section.platform}:${supportedModel.name}`.toLowerCase()
        const existing = map.get(key) ?? {
          key,
          name: supportedModel.name,
          platform: supportedModel.platform || section.platform,
          pricing: supportedModel.pricing,
          groups: []
        }

        for (const group of section.groups) {
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

const groupOptions = computed(() => {
  return [
    { value: 'cheapest', label: '最低倍率优先' },
    { value: 'all', label: '不使用倍率' },
    ...sortedRateGroups.value
      .map((group, index) => ({
        value: String(group.id),
        label: `${providerLabel(group.platform)} · ${group.name} · x${formatRate(effectiveGroupRate(group))}${index === 0 ? ' · 最低倍率' : ''}`
      }))
  ]
})

const rateGroups = computed<UserAvailableGroup[]>(() => {
  const groups = new Map<number, UserAvailableGroup>()

  if (isAuthenticated.value) {
    for (const group of availableGroups.value) {
      groups.set(group.id, group)
    }

    for (const channel of sourceChannels.value) {
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

const selectedRateGroup = computed(() => {
  const groupId = selectedGroupKey()
  if (groupId === 'all' || groupId === 'cheapest') return null
  return rateGroups.value.find((group) => String(group.id) === groupId) ?? null
})

const canUseRatePrices = computed(() => sortedRateGroups.value.length > 0 && selectedGroupKey() !== 'all')

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

function selectedGroupKey(): string {
  return String(selectedGroupId.value ?? 'cheapest')
}

function updateShowRatePrices(value: boolean): void {
  showRatePrices.value = canUseRatePrices.value ? value : false
}

function effectiveRate(model: PlazaModel): number | null {
  if (selectedGroupKey() === 'cheapest') {
    const group = cheapestRateGroupForPlatform(model.platform)
    if (!group) return null

    const rate = effectiveGroupRate(group)
    return Number.isFinite(rate) && rate > 0 ? rate : null
  }

  const group = selectedRateGroup.value
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
  const rate = showRatePrices.value ? effectiveRate(model) ?? 1 : 1
  return formatPricePerMillion(value * rate)
}

function formatBaseModelPrice(value: number | null | undefined): string {
  if (value == null) return '-'
  return formatPricePerMillion(value)
}

function formatPricePerMillion(value: number): string {
  const perMillion = value * 1_000_000
  const digits = perMillion > 0 && perMillion < 1 ? 3 : 2
  return `$${perMillion.toFixed(digits)}/M`
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
  return showRatePrices.value && selectedGroupKey() !== 'all' && effectiveRate(model) != null
}

function effectiveGroupRate(group: UserAvailableGroup): number {
  return userGroupRates.value[group.id] ?? group.rate_multiplier ?? 1
}

function compareRateGroupsByCost(a: UserAvailableGroup, b: UserAvailableGroup): number {
  const rateCompare = effectiveGroupRate(a) - effectiveGroupRate(b)
  if (rateCompare !== 0) return rateCompare

  const providerCompare = providerLabel(a.platform).localeCompare(providerLabel(b.platform))
  return providerCompare || a.name.localeCompare(b.name)
}

function cheapestRateGroupForPlatform(platform: string): UserAvailableGroup | null {
  return sortedRateGroups.value.find((group) => samePlatform(group.platform, platform)) ?? null
}

function cheapestPlatformRate(platform: string): number {
  const group = cheapestRateGroupForPlatform(platform)
  return group ? effectiveGroupRate(group) : Number.POSITIVE_INFINITY
}

function modelEffectivePriceScore(model: PlazaModel): number {
  const rate = selectedGroupKey() === 'all' ? 1 : cheapestPlatformRate(model.platform)
  return modelBasePriceScore(model) * (Number.isFinite(rate) ? rate : 1)
}

function compareModelCardsByCost(a: PlazaModel, b: PlazaModel): number {
  const rateCompare = cheapestPlatformRate(a.platform) - cheapestPlatformRate(b.platform)
  if (rateCompare !== 0) return rateCompare

  const priceCompare = modelEffectivePriceScore(a) - modelEffectivePriceScore(b)
  if (priceCompare !== 0) return priceCompare

  const providerCompare = providerLabel(a.platform).localeCompare(providerLabel(b.platform))
  return providerCompare || a.name.localeCompare(b.name)
}

function effectiveRateLabel(model: PlazaModel): string {
  const rate = effectiveRate(model) ?? 1
  if (selectedGroupKey() !== 'cheapest') return `x${formatRate(rate)}`

  const group = cheapestRateGroupForPlatform(model.platform)
  return group ? `${group.name} · x${formatRate(rate)}` : `x${formatRate(rate)}`
}

function selectCheapestRateMode(): void {
  if (sortedRateGroups.value.length === 0) {
    selectedGroupId.value = 'all'
    showRatePrices.value = false
    return
  }

  selectedGroupId.value = 'cheapest'
  showRatePrices.value = true
}

function samePlatform(a: string, b: string): boolean {
  return a.trim().toLowerCase() === b.trim().toLowerCase()
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

function hasPromptCaching(model: PlazaModel): boolean {
  return !!model.pricing && (
    (model.pricing.cache_write_price ?? 0) > 0 ||
    (model.pricing.cache_read_price ?? 0) > 0
  )
}

function providerLabel(platform: string): string {
  const normalized = platform.toLowerCase()
  if (normalized.includes('anthropic') || normalized.includes('claude')) return 'Anthropic'
  if (normalized.includes('openai') || normalized.includes('gpt')) return 'Openai'
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

const fallbackChannels: UserAvailableChannel[] = [
  {
    name: 'Openai',
    description: 'ChatGPT 与 Codex 常用模型',
    platforms: [
      {
        platform: 'openai',
        groups: openaiGroups,
        supported_models: [
          { name: 'gpt-5.2', platform: 'openai', pricing: tokenPricing(1.75, 14, 0, 0.175) },
          { name: 'gpt-5.2-codex', platform: 'openai', pricing: tokenPricing(1.75, 14, 0, 0.175) },
          { name: 'gpt-5.3-codex', platform: 'openai', pricing: tokenPricing(1.75, 14, 0, 0.175) },
          { name: 'gpt-5.4', platform: 'openai', pricing: tokenPricing(2.5, 15, 0, 0.25) },
          { name: 'gpt-5.4-mini', platform: 'openai', pricing: tokenPricing(0.75, 4.5, 0, 0.075) },
          { name: 'gpt-5.5', platform: 'openai', pricing: tokenPricing(5, 30, 0, 0.5) }
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
          { name: 'claude-haiku-4', platform: 'anthropic', pricing: tokenPricing(1, 5, 1.25, 0.1) },
          { name: 'claude-opus-4', platform: 'anthropic', pricing: tokenPricing(5, 25, 6.25, 0.5) },
          { name: 'claude-opus-4.1', platform: 'anthropic', pricing: tokenPricing(5, 25, 6.25, 0.5) },
          { name: 'claude-opus-4.5', platform: 'anthropic', pricing: tokenPricing(5, 25, 6.25, 0.5) },
          { name: 'claude-sonnet-4', platform: 'anthropic', pricing: tokenPricing(3, 15, 3.75, 0.3) },
          { name: 'claude-sonnet-4.5', platform: 'anthropic', pricing: tokenPricing(3, 15, 3.75, 0.3) }
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
  font-size: 1.55rem;
  font-weight: 950;
  line-height: 1;
}

.model-plaza-title p {
  margin-top: 0.28rem;
  color: rgba(238, 246, 240, 0.68);
  font-size: 0.82rem;
  font-weight: 700;
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

.model-stat-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
  max-width: none;
}

.model-stat-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  min-height: 4.9rem;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: var(--public-surface-raised);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    var(--public-shadow);
  padding: 1rem;
  backdrop-filter: blur(18px);
}

.model-stat-card p {
  color: rgba(222, 232, 255, 0.58);
  font-size: 0.78rem;
  font-weight: 800;
  text-transform: uppercase;
}

.model-stat-card strong {
  display: block;
  color: rgba(255, 255, 255, 0.96);
  font-size: 1.7rem;
  font-weight: 950;
  line-height: 1.05;
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
  margin-top: 1.45rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 1rem;
  align-items: center;
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

.model-filter-actions {
  display: flex;
  align-items: center;
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

.model-card-grid {
  margin-top: 1.45rem;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(16.5rem, 1fr));
  gap: 1rem;
}

.model-price-card {
  min-height: 14.2rem;
  border-radius: 8px;
  border: 1px solid var(--public-border);
  background: var(--public-surface-raised);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    var(--public-shadow);
  padding: 1rem;
  backdrop-filter: blur(18px);
}

.model-card-head {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.72rem;
  align-items: start;
}

.model-card-head h2 {
  overflow: hidden;
  color: rgba(255, 255, 255, 0.95);
  font-size: 1rem;
  font-weight: 950;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-card-head p {
  margin-top: 0.16rem;
  color: rgba(222, 232, 255, 0.72);
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.model-cache-badge {
  margin-top: 1rem;
  display: inline-flex;
  align-items: center;
  gap: 0.32rem;
  border-radius: 0.42rem;
  background: rgba(251, 191, 36, 0.13);
  padding: 0.28rem 0.5rem;
  color: #f8d36d;
  font-size: 0.72rem;
  font-weight: 850;
}

.model-price-list {
  margin-top: 1.05rem;
  display: grid;
  gap: 0.78rem;
}

.model-price-list div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.model-price-list dt,
.model-price-list dd {
  display: inline-flex;
  align-items: center;
}

.model-price-list dt {
  gap: 0.38rem;
  color: rgba(222, 232, 255, 0.58);
  font-size: 0.78rem;
  font-weight: 800;
}

.model-price-list dd {
  flex-direction: column;
  align-items: flex-end;
  gap: 0.12rem;
  color: rgba(255, 255, 255, 0.94);
  font-size: 0.9rem;
  font-weight: 950;
  text-align: right;
}

.model-price-list dd.is-rate-price {
  color: #7dffaa;
}

.model-price-list dd small {
  color: rgba(222, 232, 255, 0.68);
  font-size: 0.68rem;
  font-weight: 800;
}

.model-price-muted {
  border-top: 1px dashed rgba(221, 230, 255, 0.14);
  padding-top: 0.72rem;
}

.model-price-muted dt,
.model-price-muted dd {
  color: rgba(222, 232, 255, 0.7);
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

  .model-stat-grid,
  .model-filter-panel {
    grid-template-columns: 1fr;
  }

  .model-filter-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .model-group-select {
    width: 100%;
  }

  .model-rate-toggle {
    justify-content: flex-end;
  }
}
</style>
