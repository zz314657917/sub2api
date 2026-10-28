<template>
  <BaseDialog :show="show" :title="t('dashboard.onboarding.setupTitle')" width="wide" @close="emit('skip')">
    <div class="setup-intro">
      <span class="setup-eyebrow">{{ t('dashboard.onboarding.setupBadge') }}</span>
      <p>{{ t('dashboard.onboarding.setupDescription') }}</p>
    </div>

    <div class="setup-grid">
      <article class="setup-card setup-download">
        <div class="setup-step"><span>01</span>{{ t('dashboard.onboarding.downloadStep') }}</div>
        <div class="setup-product">
          <img src="/onboarding/config-helper-logo.png" alt="" aria-hidden="true" width="56" height="56">
          <div>
            <h4>{{ t('dashboard.onboarding.helperName') }}</h4>
            <p>{{ t('dashboard.onboarding.helperMeta') }}</p>
          </div>
        </div>
        <p class="setup-description">{{ t('dashboard.onboarding.helperDescription') }}</p>
        <a
          :href="downloadUrl"
          download="ChatGPT配置助手.exe"
          class="btn btn-primary setup-action"
          data-testid="setup-download"
        >
          <Icon name="download" size="sm" />
          {{ t('dashboard.onboarding.downloadHelper') }}
        </a>
        <p class="setup-footnote">
          {{ t('dashboard.onboarding.helperNote') }}
          <span class="block" data-testid="setup-download-warning">{{ t('dashboard.onboarding.helperWarning') }}</span>
        </p>
      </article>

      <article class="setup-card">
        <div class="setup-step"><span>02</span>{{ t('dashboard.onboarding.copyStep') }}</div>
        <div class="setup-key-heading">
          <span class="setup-key-icon"><Icon name="key" size="md" /></span>
          <div>
            <h4>{{ t('dashboard.onboarding.yourDefaultKey') }}</h4>
            <p>{{ t(`dashboard.onboarding.keyState.${keyState}`) }}</p>
          </div>
        </div>
        <div class="setup-key-value" aria-live="polite" data-testid="setup-key-state">
          <code v-if="keyState === 'ready'">{{ maskedKey }}</code>
          <span v-else>{{ t(`dashboard.onboarding.keyMessage.${keyState}`) }}</span>
          <Icon v-if="keyState === 'ready'" name="lock" size="sm" />
        </div>
        <button
          v-if="keyState === 'ready'"
          type="button"
          class="btn setup-copy setup-action"
          data-testid="setup-copy"
          :disabled="copying"
          @click="copyDefaultKey"
        >
          <Icon :name="copied ? 'check' : 'clipboard'" size="sm" />
          {{ t(copied ? 'dashboard.onboarding.keyCopied' : 'dashboard.onboarding.copyDefaultKey') }}
        </button>
        <button
          v-else-if="keyState === 'loading' || keyState === 'error'"
          type="button"
          class="btn btn-secondary setup-action"
          :disabled="keyState === 'loading'"
          @click="loadDefaultKey"
        >
          {{ t(keyState === 'loading' ? 'dashboard.onboarding.loadingKey' : 'dashboard.onboarding.retryKey') }}
        </button>
        <button v-else type="button" class="btn btn-secondary setup-action" @click="emit('create')">
          {{ t(hasApiKey || defaultKey ? 'dashboard.onboarding.manageKey' : 'dashboard.onboarding.createKey') }}
        </button>
        <p class="setup-footnote">{{ t(copied ? 'dashboard.onboarding.pasteHint' : 'dashboard.onboarding.keyPrivacy') }}</p>
      </article>
    </div>

    <div class="setup-node-note">
      <Icon name="bolt" size="sm" />
      <p>{{ t('dashboard.onboarding.nodeHint') }}</p>
    </div>
    <p v-if="hasBenefit" class="setup-benefit">
      <Icon name="gift" size="sm" />
      {{ benefitText }}
    </p>

    <template #footer>
      <div class="setup-footer">
        <div>
          <button type="button" @click="emit('tutorial')">{{ t('dashboard.onboarding.manualSetup') }} <Icon name="externalLink" size="xs" /></button>
          <button type="button" @click="openSupportPopup">{{ t('dashboard.onboarding.joinGroup') }} <Icon name="chat" size="xs" /></button>
        </div>
        <button type="button" @click="emit('skip')">{{ t('dashboard.onboarding.skip') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI } from '@/api/keys'
import { useClipboard } from '@/composables/useClipboard'
import { maskApiKey } from '@/utils/maskApiKey'
import { openSupportPopup } from '@/utils/supportPopup'
import type { ApiKey } from '@/types'

const props = defineProps<{
  show: boolean
  hasApiKey?: boolean
  hasBenefit?: boolean
  benefitLabel?: string
  benefitRewardLabel?: string
  benefitKind?: 'reward' | 'trial' | 'wallet'
}>()
const emit = defineEmits<{
  (e: 'create'): void
  (e: 'tutorial'): void
  (e: 'skip'): void
}>()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const downloadUrl = `${import.meta.env.BASE_URL}downloads/chatgpt-config-windows.exe`
const defaultKey = ref<ApiKey | null>(null)
const loadingKey = ref(false)
const loadFailed = ref(false)
const copied = ref(false)
const copying = ref(false)
const checkedAt = ref(Date.now())
let request: AbortController | null = null

const keyState = computed(() => {
  if (loadingKey.value) return 'loading'
  if (loadFailed.value) return 'error'
  const key = defaultKey.value
  if (!key) return 'missing'
  if (key.status === 'inactive') return 'inactive'
  const expiresAt = key.expires_at ? Date.parse(key.expires_at) : null
  if (key.status === 'expired' || expiresAt !== null && (!Number.isFinite(expiresAt) || expiresAt <= checkedAt.value)) return 'expired'
  if (key.status === 'quota_exhausted' || key.quota > 0 && key.quota_used >= key.quota) return 'quota'
  return key.status === 'active' && key.key?.trim() ? 'ready' : 'error'
})
const maskedKey = computed(() => maskApiKey(defaultKey.value?.key || ''))
const benefitText = computed(() => {
  if (props.benefitKind === 'reward') {
    return t('dashboard.onboarding.setupReward', { amount: props.benefitLabel || t('dashboard.onboarding.rewardFallback') })
  }
  if (props.benefitKind === 'wallet') return t('dashboard.onboarding.setupWallet')
  return t('dashboard.onboarding.setupTrial')
})

async function loadDefaultKey() {
  if (!props.show) return
  request?.abort()
  const current = new AbortController()
  request = current
  defaultKey.value = null
  copied.value = false
  loadFailed.value = false
  loadingKey.value = true
  try {
    // Backend identifies the oldest ID as default; never guess from a key's name.
    const result = await keysAPI.list(1, 1, { sort_by: 'id', sort_order: 'asc' }, { signal: current.signal })
    if (current.signal.aborted || request !== current) return
    checkedAt.value = Date.now()
    defaultKey.value = result.items.find(key => key.is_default) || null
  } catch {
    if (!current.signal.aborted && request === current) loadFailed.value = true
  } finally {
    if (request === current) loadingKey.value = false
  }
}
async function copyDefaultKey() {
  checkedAt.value = Date.now()
  if (keyState.value !== 'ready' || !defaultKey.value || copying.value) return
  const current = defaultKey.value
  copying.value = true
  try {
    const success = await copyToClipboard(current.key, t('dashboard.onboarding.keyCopied'))
    if (props.show && defaultKey.value === current) copied.value = success
  } finally {
    copying.value = false
  }
}
watch(() => props.show, show => {
  if (show) void loadDefaultKey()
  else {
    request?.abort()
    request = null
    defaultKey.value = null
    copied.value = false
  }
}, { immediate: true })
onUnmounted(() => { request?.abort(); defaultKey.value = null })
</script>

<style scoped>
.setup-intro { margin-bottom: 22px; }
.setup-eyebrow { font-size: 11px; font-weight: 700; letter-spacing: .1em; color: #a9583e; }
.setup-intro > p { margin-top: 8px; color: #6c6a64; font-size: 14px; line-height: 1.7; }
.setup-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 16px; }
.setup-card { display: flex; flex-direction: column; min-width: 0; padding: 22px; border: 1px solid #d8cec2; border-radius: 12px; background: #fffdf9; }
.setup-download { background: linear-gradient(145deg, #f8eee3, #fffaf5); }
.setup-step { display: flex; align-items: center; gap: 10px; color: #504f49; font-size: 12px; font-weight: 600; }
.setup-step > span { color: #a9583e; font-variant-numeric: tabular-nums; font-size: 11px; border: 1px solid #d8cec2; border-radius: 6px; padding: 3px 6px; }
.setup-product, .setup-key-heading { display: flex; align-items: center; gap: 14px; margin: 24px 0 16px; min-height: 56px; }
.setup-product img { image-rendering: pixelated; flex-shrink: 0; }
.setup-card h4 { font-size: 17px; font-weight: 650; line-height: 1.5; }
.setup-product p, .setup-key-heading p { margin-top: 4px; color: #8e8b82; font-size: 12px; }
.setup-key-icon { display: flex; align-items: center; justify-content: center; width: 48px; height: 48px; flex-shrink: 0; border-radius: 12px; background: #f0efe7; color: #5f7f68; }
.setup-description, .setup-key-value { min-height: 60px; margin-bottom: 20px; color: #6c6a64; font-size: 13px; line-height: 1.7; }
.setup-key-value { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 12px; border: 1px dashed #d8cec2; border-radius: 8px; background: #faf9f5; overflow-wrap: anywhere; }
.setup-key-value code { font-size: 14px; color: #504f49; }
.setup-action { display: flex; align-items: center; justify-content: center; gap: 8px; width: 100%; margin-top: auto; min-height: 44px; font-size: 13px; }
.setup-copy { background: #526f5c; color: white; }
.setup-copy:hover { background: #435d4c; }
.setup-footnote { min-height: 18px; margin-top: 10px; color: #8e8b82; font-size: 11px; text-align: center; line-height: 1.7; }
.setup-node-note, .setup-benefit { display: flex; align-items: flex-start; gap: 8px; color: #6c6a64; font-size: 12px; line-height: 1.7; }
.setup-node-note { margin-top: 18px; }
.setup-node-note > svg, .setup-benefit > svg { flex-shrink: 0; margin-top: 2px; }
.setup-benefit { margin-top: 12px; padding: 12px 14px; border-radius: 8px; background: #f0efe7; color: #526f5c; }
.setup-footer, .setup-footer > div { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; }
.setup-footer { width: 100%; }
.setup-footer button { display: inline-flex; align-items: center; gap: 5px; color: #6c6a64; font-size: 12px; padding: 4px 0; }
.setup-footer button:hover { color: #a9583e; }
:global(.dark .setup-card) { background: #172235; border-color: #334155; }
:global(.dark .setup-download) { background: linear-gradient(145deg, #322923, #172235); }
:global(.dark .setup-eyebrow), :global(.dark .setup-step > span) { color: #d9957b; }
:global(.dark .setup-step), :global(.dark .setup-key-value code) { color: #e2e8f0; }
:global(.dark .setup-intro > p), :global(.dark .setup-description), :global(.dark .setup-node-note), :global(.dark .setup-key-value), :global(.dark .setup-footer button) { color: #aeb7c4; }
:global(.dark .setup-step > span), :global(.dark .setup-key-value) { border-color: #475569; }
:global(.dark .setup-key-value) { background: #101a2a; }
:global(.dark .setup-key-icon), :global(.dark .setup-benefit) { background: #21372f; color: #a3c8ad; }
@media (max-width: 639px) {
  .setup-intro { margin-bottom: 16px; }
  .setup-grid { grid-template-columns: minmax(0, 1fr); gap: 12px; }
  .setup-card { padding: 16px; }
  .setup-product, .setup-key-heading { margin: 16px 0 12px; }
  .setup-description, .setup-key-value { min-height: 44px; margin-bottom: 14px; }
}
</style>
