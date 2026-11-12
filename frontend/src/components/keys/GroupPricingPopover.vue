<template>
  <div
    ref="triggerEl"
    class="min-w-0"
    @mouseenter="onEnter"
    @mouseleave="scheduleHide"
    @focusin="onEnter"
    @focusout="scheduleHide"
  >
    <slot />
    <Teleport to="body">
      <div
        v-if="show"
        ref="popoverEl"
        role="tooltip"
        class="group-pricing-popover fixed flex max-h-[min(32rem,calc(100vh-1rem))] w-[36rem] max-w-[calc(100vw-1rem)] flex-col overflow-hidden rounded-xl border bg-white text-xs shadow-xl dark:bg-dark-800"
        :class="platformBorderClass(platform)"
        :style="popoverStyle"
        data-testid="group-pricing-popover"
        @mouseenter="cancelHide"
        @mouseleave="scheduleHide"
      >
        <div
          class="flex items-center justify-between gap-3 border-b px-4 py-2.5"
          :class="[platformBadgeLightClass(platform), platformBorderClass(platform)]"
        >
          <span class="min-w-0 truncate text-sm font-semibold">{{ groupName }}</span>
          <span class="flex-shrink-0 text-[11px] opacity-80">{{ t('keys.groupPricing.title') }}</span>
        </div>

        <div v-if="loading && !models.length" class="px-4 py-6 text-center text-gray-500 dark:text-gray-400">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="!models.length" class="px-4 py-6 text-center text-gray-500 dark:text-gray-400">
          {{ t('keys.groupPricing.empty') }}
        </div>
        <div v-else class="min-h-0 overflow-y-auto px-4 py-2">
          <table class="w-full table-fixed border-collapse tabular-nums">
            <colgroup>
              <col class="w-[34%]" />
              <col />
              <col />
              <col />
              <col />
            </colgroup>
            <thead>
              <tr class="text-left text-[11px] text-gray-500 dark:text-gray-400">
                <th class="py-1.5 pr-2 font-medium">{{ t('keys.groupPricing.model') }}</th>
                <th class="py-1.5 pr-2 font-medium">{{ t('keys.groupPricing.input') }}</th>
                <th class="py-1.5 pr-2 font-medium">{{ t('keys.groupPricing.output') }}</th>
                <th class="py-1.5 pr-2 font-medium">{{ t('keys.groupPricing.cacheRead') }}</th>
                <th class="py-1.5 font-medium">{{ t('keys.groupPricing.cacheWrite') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in rows"
                :key="row.key"
                class="border-t border-gray-100 text-gray-700 dark:border-dark-700/60 dark:text-gray-300"
              >
                <td class="truncate py-1.5 pr-2 font-mono font-medium text-gray-900 dark:text-white" :title="row.name">{{ row.name }}</td>
                <template v-if="row.perRequest">
                  <td colspan="4" class="py-1.5 font-mono">{{ row.perRequest }}</td>
                </template>
                <template v-else>
                  <td class="py-1.5 pr-2 font-mono">{{ row.input }}</td>
                  <td class="py-1.5 pr-2 font-mono">{{ row.output }}</td>
                  <td class="py-1.5 pr-2 font-mono">{{ row.cacheRead }}</td>
                  <td class="py-1.5 font-mono">{{ row.cacheWrite }}</td>
                </template>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          v-if="models.length && rateMultiplier !== undefined"
          class="border-t border-gray-100 px-4 py-2 text-[11px] text-gray-500 dark:border-dark-700/60 dark:text-gray-400"
        >
          {{ t('keys.groupPricing.rateNote', { rate: rateMultiplier }) }}
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserSupportedModel } from '@/api/channels'
import { BILLING_MODE_IMAGE, BILLING_MODE_PER_REQUEST } from '@/constants/channel'
import { formatScaled } from '@/utils/pricing'
import { displayModelLabel } from '@/utils/modelDisplay'
import { platformBadgeLightClass, platformBorderClass } from '@/utils/platformColors'

const props = withDefaults(defineProps<{
  groupName: string
  platform: string
  models: UserSupportedModel[]
  loading?: boolean
  /** Effective multiplier applied on top of the listed prices. */
  rateMultiplier?: number
  /** `side` opens next to the trigger (dropdown options); `bottom` opens below it. */
  placement?: 'bottom' | 'side'
  disabled?: boolean
}>(), {
  loading: false,
  rateMultiplier: undefined,
  placement: 'bottom',
  disabled: false
})

const emit = defineEmits<{ (e: 'open'): void }>()

const { t } = useI18n()

const PER_MILLION = 1_000_000

const rows = computed(() => props.models.map((model) => {
  const pricing = model.pricing
  // Tiered models without flat prices: show the first tier as the entry price.
  const tier = pricing?.intervals?.[0]
  const pick = (flat: number | null | undefined, tiered: number | null | undefined) => flat ?? tiered ?? null
  const perToken = (value: number | null) => (value == null ? '-' : `${formatScaled(value, PER_MILLION)} /1M`)
  const isPerUnit = pricing?.billing_mode === BILLING_MODE_PER_REQUEST || pricing?.billing_mode === BILLING_MODE_IMAGE
  const unitPrice = pick(pricing?.per_request_price, tier?.per_request_price) ?? (pricing?.billing_mode === BILLING_MODE_IMAGE ? pricing.image_output_price : null)
  return {
    key: `${model.platform}:${model.name}`,
    name: displayModelLabel(model.name),
    perRequest: isPerUnit ? (unitPrice == null ? '-' : `${formatScaled(unitPrice, 1)} ${t('keys.groupPricing.perRequest')}`) : '',
    input: perToken(pick(pricing?.input_price, tier?.input_price)),
    output: perToken(pick(pricing?.output_price, tier?.output_price)),
    cacheRead: perToken(pick(pricing?.cache_read_price, tier?.cache_read_price)),
    cacheWrite: perToken(pick(pricing?.cache_write_price, tier?.cache_write_price))
  }
}))

const show = ref(false)
const triggerEl = ref<HTMLElement | null>(null)
const popoverEl = ref<HTMLElement | null>(null)
const popoverStyle = ref<Record<string, string>>({})
let showTimer: ReturnType<typeof setTimeout> | null = null
let hideTimer: ReturnType<typeof setTimeout> | null = null

const SHOW_DELAY_MS = 250
const HIDE_DELAY_MS = 200
// Above Select dropdowns (100000020) so option hovers are not covered.
const Z_INDEX = '100000030'

function updatePosition() {
  const trigger = triggerEl.value
  const popover = popoverEl.value
  if (!trigger || !popover) return
  const rect = trigger.getBoundingClientRect()
  const margin = 8
  const width = popover.offsetWidth
  const height = popover.offsetHeight
  const vw = window.innerWidth
  const vh = window.innerHeight
  let top: number
  let left: number

  if (props.placement === 'side') {
    // The trigger sits inside a dropdown; anchor to the dropdown edge when possible.
    const anchor = (trigger.closest('.select-dropdown-portal, [data-pricing-anchor]') as HTMLElement | null)?.getBoundingClientRect() ?? rect
    left = anchor.right + margin
    if (left + width > vw - margin) left = anchor.left - width - margin
    if (left < margin) left = Math.max(margin, vw - margin - width)
    top = Math.min(Math.max(rect.top, margin), vh - margin - height)
  } else {
    top = rect.bottom + margin
    if (top + height > vh - margin) top = Math.max(margin, rect.top - height - margin)
    left = Math.min(Math.max(rect.left, margin), vw - margin - width)
  }

  popoverStyle.value = { top: `${Math.round(top)}px`, left: `${Math.round(left)}px`, zIndex: Z_INDEX }
}

function cancelHide() {
  if (hideTimer) clearTimeout(hideTimer)
  hideTimer = null
}

function onEnter() {
  if (props.disabled) return
  cancelHide()
  if (show.value || showTimer) return
  showTimer = setTimeout(() => {
    showTimer = null
    emit('open')
    show.value = true
    nextTick(() => {
      updatePosition()
      window.addEventListener('scroll', onScroll, true)
      window.addEventListener('resize', updatePosition)
    })
  }, SHOW_DELAY_MS)
}

function onScroll(event: Event) {
  // Scrolling inside the popover itself must not move it.
  if (popoverEl.value && event.target instanceof Node && popoverEl.value.contains(event.target)) return
  updatePosition()
}

function hide() {
  show.value = false
  window.removeEventListener('scroll', onScroll, true)
  window.removeEventListener('resize', updatePosition)
}

function scheduleHide() {
  if (showTimer) {
    clearTimeout(showTimer)
    showTimer = null
  }
  cancelHide()
  hideTimer = setTimeout(hide, HIDE_DELAY_MS)
}

// Re-measure once data arrives and the popover grows.
watch(() => props.models.length, () => {
  if (show.value) nextTick(updatePosition)
})

watch(() => props.disabled, (disabled) => {
  if (!disabled) return
  if (showTimer) clearTimeout(showTimer)
  showTimer = null
  hide()
})

onBeforeUnmount(() => {
  if (showTimer) clearTimeout(showTimer)
  cancelHide()
  hide()
})
</script>
