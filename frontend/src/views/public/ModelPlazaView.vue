<template>
  <div class="model-plaza-page min-h-screen text-slate-900">
    <header class="model-plaza-topbar">
      <nav class="model-plaza-nav mx-auto">
        <router-link to="/home" class="model-plaza-brand">
          <span class="model-plaza-logo">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </span>
          <span>
            <span class="block text-sm font-black leading-tight sm:text-base">{{ siteName }}</span>
            <span class="hidden text-[11px] font-semibold text-slate-500 sm:block">模型与价格</span>
          </span>
        </router-link>

        <div class="model-plaza-links">
          <router-link to="/home">首页</router-link>
          <router-link to="/tutorial">教程</router-link>
          <router-link to="/models" class="router-link-active">模型广场</router-link>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="model-plaza-login">
            {{ isAuthenticated ? '控制台' : '登录' }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="model-plaza-main mx-auto">
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
          <Select
            v-model="selectedGroupId"
            :options="groupOptions"
            class="model-group-select"
            :searchable="groupOptions.length > 8"
          />
          <label class="model-rate-toggle">
            <Toggle v-model="showRatePrices" />
            <span>显示倍率价格</span>
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
            <span class="model-rate-badge">
              充值 ¥1 = $1
            </span>
          </div>

          <div v-if="hasPromptCaching(model)" class="model-cache-badge">
            <Icon name="sparkles" size="xs" />
            Prompt Caching
          </div>

          <dl class="model-price-list">
            <div>
              <dt><Icon name="upload" size="xs" />输入</dt>
              <dd>
                {{ formatModelPrice(model.pricing?.input_price, model) }}
                <small v-if="showRatePrices">{{ rateSuffix(model) }}</small>
              </dd>
            </div>
            <div>
              <dt><Icon name="download" size="xs" />输出</dt>
              <dd>
                {{ formatModelPrice(model.pricing?.output_price, model) }}
                <small v-if="showRatePrices">{{ rateSuffix(model) }}</small>
              </dd>
            </div>
            <div class="model-price-muted">
              <dt><Icon name="document" size="xs" />缓存写入</dt>
              <dd>
                {{ formatModelPrice(model.pricing?.cache_write_price, model) }}
                <small v-if="showRatePrices && model.pricing?.cache_write_price != null">{{ rateSuffix(model) }}</small>
              </dd>
            </div>
            <div class="model-price-muted">
              <dt><Icon name="sync" size="xs" />缓存读取</dt>
              <dd>
                {{ formatModelPrice(model.pricing?.cache_read_price, model) }}
                <small v-if="showRatePrices && model.pricing?.cache_read_price != null">{{ rateSuffix(model) }}</small>
              </dd>
            </div>
          </dl>
        </article>
      </section>

      <section v-else class="model-empty-card">
        <Icon name="search" size="lg" />
        <p>没有找到匹配的模型</p>
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
import { useAuthStore, useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import userChannelsAPI, {
  type UserAvailableChannel,
  type UserAvailableGroup,
  type UserSupportedModelPricing
} from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { BILLING_MODE_TOKEN } from '@/constants/channel'

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

const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '落叶网络')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))

const loading = ref(false)
const searchQuery = ref('')
const selectedProvider = ref('all')
const selectedGroupId = ref<string | number | boolean | null>('all')
const showRatePrices = ref(false)
const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})

const sourceChannels = computed(() => channels.value.length > 0 ? channels.value : fallbackChannels)

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

  return Array.from(map.values()).sort((a, b) => {
    const providerCompare = providerLabel(a.platform).localeCompare(providerLabel(b.platform))
    return providerCompare || a.name.localeCompare(b.name)
  })
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
  const groups = new Map<number, ModelGroupMeta>()
  for (const model of modelCards.value) {
    for (const group of model.groups) {
      if (!groups.has(group.id)) groups.set(group.id, group)
    }
  }

  return [
    { value: 'all', label: '全部分组' },
    ...Array.from(groups.values())
      .sort((a, b) => a.name.localeCompare(b.name))
      .map((group) => ({
        value: String(group.id),
        label: `${group.name} (x${formatRate(group.effectiveRate)})`
      }))
  ]
})

const filteredModels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const groupId = selectedGroupKey()

  return modelCards.value.filter((model) => {
    if (selectedProvider.value !== 'all' && model.platform !== selectedProvider.value) return false
    if (groupId !== 'all' && !model.groups.some((group) => String(group.id) === groupId)) return false
    if (!query) return true

    return (
      model.name.toLowerCase().includes(query) ||
      providerLabel(model.platform).toLowerCase().includes(query) ||
      model.groups.some((group) => group.name.toLowerCase().includes(query))
    )
  })
})

async function loadChannels(force = false): Promise<void> {
  if (!isAuthenticated.value) {
    if (force) {
      channels.value = fallbackChannels
    }
    return
  }

  loading.value = true
  try {
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch(() => ({} as Record<number, number>))
    ])
    channels.value = list.length > 0 ? list : fallbackChannels
    userGroupRates.value = rates
  } catch (error) {
    console.warn('Failed to load model plaza data, fallback models will be shown.', error)
    channels.value = fallbackChannels
  } finally {
    loading.value = false
  }
}

function resetFilters(): void {
  searchQuery.value = ''
  selectedProvider.value = 'all'
  selectedGroupId.value = 'all'
}

function selectedGroupKey(): string {
  return String(selectedGroupId.value ?? 'all')
}

function effectiveRate(model: PlazaModel): number {
  const groupId = selectedGroupKey()
  if (groupId !== 'all') {
    const selected = model.groups.find((group) => String(group.id) === groupId)
    return selected?.effectiveRate ?? 1
  }

  const rates = model.groups.map((group) => group.effectiveRate).filter((rate) => Number.isFinite(rate) && rate > 0)
  return rates.length > 0 ? Math.min(...rates) : 1
}

function rateSuffix(model: PlazaModel): string {
  return `x${formatRate(effectiveRate(model))}`
}

function formatRate(rate: number): string {
  return Number.isInteger(rate) ? String(rate) : rate.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}

function formatModelPrice(value: number | null | undefined, model: PlazaModel): string {
  if (value == null) return '-'
  const rate = showRatePrices.value ? effectiveRate(model) : 1
  const perMillion = value * 1_000_000 * rate
  const digits = perMillion > 0 && perMillion < 1 ? 3 : 2
  return `$${perMillion.toFixed(digits)}/M`
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
.model-plaza-page {
  background:
    radial-gradient(circle at 42% 10%, rgba(174, 232, 228, 0.52), transparent 34%),
    radial-gradient(circle at 76% 0%, rgba(222, 224, 232, 0.72), transparent 28%),
    linear-gradient(135deg, #e8f4f3 0%, #eef2f2 48%, #dfe9ea 100%);
}

.model-plaza-topbar {
  position: sticky;
  top: 0;
  z-index: 40;
  border-bottom: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(246, 248, 248, 0.82);
  backdrop-filter: blur(18px);
}

.model-plaza-nav {
  display: flex;
  width: min(100%, 96rem);
  min-height: 3.25rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.38rem 1rem;
}

.model-plaza-brand,
.model-plaza-links {
  display: flex;
  align-items: center;
}

.model-plaza-brand {
  min-width: 0;
  gap: 0.65rem;
  color: #0f172a;
}

.model-plaza-logo {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  flex: 0 0 2rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.55rem;
  background: rgba(255, 255, 255, 0.74);
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.12);
  padding: 0.18rem;
}

.model-plaza-links {
  gap: 0.35rem;
}

.model-plaza-links a {
  border-radius: 0.55rem;
  padding: 0.48rem 0.68rem;
  color: #64748b;
  font-size: 0.78rem;
  font-weight: 800;
}

.model-plaza-links a:hover,
.model-plaza-links a.router-link-active {
  background: rgba(15, 118, 110, 0.08);
  color: #0f766e;
}

.model-plaza-links .model-plaza-login {
  background: #0f766e;
  color: white;
}

.model-plaza-main {
  width: min(100%, 96rem);
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
  font-size: 1.55rem;
  font-weight: 950;
  line-height: 1;
}

.model-plaza-title p {
  margin-top: 0.28rem;
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
}

.model-refresh-button {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 0.72rem;
  border: 1px solid rgba(148, 163, 184, 0.34);
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 10px 22px rgba(15, 23, 42, 0.1);
  padding: 0.65rem 0.9rem;
  color: #334155;
  font-size: 0.82rem;
  font-weight: 800;
}

.model-stat-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
  max-width: 64rem;
}

.model-stat-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  min-height: 4.9rem;
  border-radius: 0.85rem;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.66);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.1);
  padding: 1rem;
}

.model-stat-card p {
  color: #64748b;
  font-size: 0.78rem;
  font-weight: 800;
  text-transform: uppercase;
}

.model-stat-card strong {
  display: block;
  color: #0f172a;
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
  border-radius: 0.8rem;
}

.model-stat-icon-total {
  background: #eee5ff;
  color: #8b5cf6;
}

.model-stat-icon-openai {
  background: #dcf8ed;
}

.model-stat-icon-anthropic {
  background: #f9e8db;
}

.model-stat-icon-gemini {
  background: #dbeafe;
}

.model-filter-panel {
  margin-top: 1.45rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 1rem;
  align-items: center;
  border-radius: 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.68);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.1);
  padding: 1rem;
}

.model-search-box {
  display: flex;
  min-height: 2.7rem;
  align-items: center;
  gap: 0.7rem;
  border-radius: 0.72rem;
  border: 1px solid rgba(148, 163, 184, 0.34);
  background: rgba(248, 250, 252, 0.72);
  padding: 0 0.85rem;
  color: #94a3b8;
}

.model-search-box input {
  min-width: 0;
  width: 100%;
  background: transparent;
  color: #0f172a;
  font-size: 0.86rem;
  font-weight: 700;
  outline: none;
}

.model-filter-actions {
  display: flex;
  align-items: center;
  gap: 0.7rem;
}

.model-group-select {
  width: min(18rem, 42vw);
}

.model-rate-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.48rem;
  color: #475569;
  font-size: 0.82rem;
  font-weight: 800;
  white-space: nowrap;
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
  border-radius: 0.72rem;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
  padding: 0.55rem 0.85rem;
  color: #475569;
  font-weight: 850;
}

.model-provider-tabs button.is-active {
  border-color: rgba(20, 184, 166, 0.42);
  background: rgba(240, 253, 250, 0.92);
  color: #0f766e;
}

.model-provider-tabs small {
  color: #64748b;
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
  border-radius: 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.7);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.1);
  padding: 1rem;
}

.model-card-head {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 0.72rem;
  align-items: start;
}

.model-card-head h2 {
  overflow: hidden;
  color: #111827;
  font-size: 1rem;
  font-weight: 950;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-card-head p {
  margin-top: 0.16rem;
  color: #94a3b8;
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.model-rate-badge {
  border-radius: 999px;
  border: 1px solid rgba(139, 92, 246, 0.24);
  background: rgba(237, 233, 254, 0.8);
  padding: 0.18rem 0.45rem;
  color: #7c3aed;
  font-size: 0.68rem;
  font-weight: 900;
  white-space: nowrap;
}

.model-cache-badge {
  margin-top: 1rem;
  display: inline-flex;
  align-items: center;
  gap: 0.32rem;
  border-radius: 0.42rem;
  background: rgba(254, 243, 199, 0.72);
  padding: 0.28rem 0.5rem;
  color: #d97706;
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
  color: #64748b;
  font-size: 0.78rem;
  font-weight: 800;
}

.model-price-list dd {
  gap: 0.25rem;
  color: #111827;
  font-size: 0.9rem;
  font-weight: 950;
}

.model-price-list dd small {
  color: #ea580c;
  font-size: 0.64rem;
  font-weight: 950;
}

.model-price-muted {
  border-top: 1px dashed rgba(148, 163, 184, 0.2);
  padding-top: 0.72rem;
}

.model-price-muted dt,
.model-price-muted dd {
  color: #64748b;
}

.model-empty-card {
  margin-top: 1.5rem;
  display: grid;
  place-items: center;
  gap: 0.7rem;
  min-height: 16rem;
  border-radius: 0.9rem;
  border: 1px dashed rgba(148, 163, 184, 0.5);
  background: rgba(255, 255, 255, 0.48);
  color: #64748b;
  text-align: center;
}

.model-empty-card button {
  border-radius: 0.55rem;
  background: #0f766e;
  padding: 0.52rem 0.8rem;
  color: white;
  font-size: 0.8rem;
  font-weight: 850;
}

.model-plaza-note {
  padding: 1.35rem 0 0.25rem;
  color: #94a3b8;
  text-align: center;
  font-size: 0.78rem;
  font-weight: 700;
}

:deep(.select-trigger) {
  min-height: 2.7rem;
  border-color: rgba(148, 163, 184, 0.34);
  background: rgba(248, 250, 252, 0.72);
  color: #0f172a;
}

@media (max-width: 900px) {
  .model-plaza-nav {
    align-items: flex-start;
    flex-direction: column;
    padding: 0.7rem 1rem;
  }

  .model-plaza-links {
    width: 100%;
    overflow-x: auto;
    padding-bottom: 0.1rem;
  }

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
