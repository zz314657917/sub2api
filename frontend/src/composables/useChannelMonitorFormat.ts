/**
 * Shared formatting helpers for channel monitor views (admin + user).
 *
 * Centralises:
 *  - status / provider label + badge class lookups
 *  - latency / availability / percent number formatting
 *  - dashboard-style helpers (warm availability colour, provider gradient, relative time)
 *
 * i18n keys live under `monitorCommon.*` so admin and user views share the
 * same translation source.
 */

import { useI18n } from 'vue-i18n'
import type { CheckMode, MonitorStatus, Provider } from '@/api/admin/channelMonitor'
import {
  PROVIDER_OPENAI,
  PROVIDER_ANTHROPIC,
  PROVIDER_GEMINI,
  PROVIDER_GROK,
  PROVIDER_KIMI,
  PROVIDER_ZHIPU,
  PROVIDER_DEEPSEEK,
  PROVIDER_ANTIGRAVITY,
  PROVIDER_MINIMAX,
  PROVIDERS,
  CHECK_MODE_PROBE,
  CHECK_MODE_QUOTA,
  CHECK_MODE_QUOTA_PROBE,
  STATUS_OPERATIONAL,
  STATUS_DEGRADED,
  STATUS_FAILED,
  STATUS_ERROR,
} from '@/constants/channelMonitor'

const NEUTRAL_BADGE = 'bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-300'

export interface AvailabilityRow {
  primary_status: MonitorStatus | ''
  availability_7d: number | null | undefined
}

export function useChannelMonitorFormat() {
  const { t } = useI18n()

  function statusLabel(s: MonitorStatus | ''): string {
    if (!s) return t('monitorCommon.status.unknown')
    return t(`monitorCommon.status.${s}`)
  }

  function statusBadgeClass(s: MonitorStatus | ''): string {
    switch (s) {
      case STATUS_OPERATIONAL:
        return 'bg-[#f3e7df] text-[#a9583e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
      case STATUS_DEGRADED:
        return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
      case STATUS_FAILED:
        return 'bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300'
      case STATUS_ERROR:
      default:
        return NEUTRAL_BADGE
    }
  }

  function providerLabel(p: Provider | string): string {
    if (PROVIDERS.includes(p as Provider)) {
      return t(`monitorCommon.providers.${p}`)
    }
    return p || '-'
  }

  function checkModeLabel(m: CheckMode | string): string {
    if (m === 'probe' || m === 'quota' || m === 'quota_probe') {
      return t(`monitorCommon.checkMode.${m}`)
    }
    return m || '-'
  }

  /**
   * Display label for a monitor's primary model. Pure-quota monitors carry the
   * literal placeholder "quota" (the probe target is an account, not a model),
   * which must not leak into the UI as a fake model name — render the
   * localized mode label instead. quota_probe keeps a real model name.
   */
  const QUOTA_MODEL_PLACEHOLDER = 'quota'

  function formatMonitorModel(model: string): string {
    if (model === QUOTA_MODEL_PLACEHOLDER) {
      return t('monitorCommon.checkMode.quota')
    }
    return model
  }

  function providerBadgeClass(p: Provider | string): string {
    switch (p) {
      case PROVIDER_OPENAI:
        return 'bg-[#f3e7df] text-[#a9583e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
      case PROVIDER_ANTHROPIC:
        return 'bg-orange-100 text-orange-700 dark:bg-orange-500/15 dark:text-orange-300'
      case PROVIDER_GEMINI:
        return 'bg-[#f5f0e8] text-[#6c6a64] ring-1 ring-[#d8cec2] dark:bg-[#8e8b82]/15 dark:text-[#d8cec2] dark:ring-[#8e8b82]/35'
      case PROVIDER_GROK:
        return 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300'
      case PROVIDER_KIMI:
        return 'bg-pink-100 text-pink-700 dark:bg-pink-500/15 dark:text-pink-300'
      case PROVIDER_ZHIPU:
        return 'bg-indigo-100 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300'
      case PROVIDER_ANTIGRAVITY:
        return 'bg-purple-100 text-purple-700 dark:bg-purple-500/15 dark:text-purple-300'
      case PROVIDER_MINIMAX:
        return 'bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
      case PROVIDER_DEEPSEEK:
        return 'bg-teal-100 text-teal-700 dark:bg-teal-500/15 dark:text-teal-300'
      default:
        return NEUTRAL_BADGE
    }
  }

  function checkModeBadgeClass(m: CheckMode | string): string {
    switch (m) {
      case CHECK_MODE_QUOTA:
      case CHECK_MODE_QUOTA_PROBE:
        return 'bg-blue-100 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300'
      case CHECK_MODE_PROBE:
      default:
        return NEUTRAL_BADGE
    }
  }

  /**
   * Tailwind class for a provider radio-button-style picker (active/inactive state).
   * Reuses the same warm/orange/neutral palette as providerBadgeClass to keep
   * visual semantics consistent across badges and pickers.
   */
  function providerPickerClass(p: Provider | string, active: boolean): string {
    switch (p) {
      case PROVIDER_OPENAI:
        return active
          ? 'border-[#cc785c] bg-[#fffaf5] text-[#a9583e] dark:border-[#f0b89e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
          : 'border-gray-200 bg-white text-gray-600 hover:border-[#d9957b] hover:text-[#a9583e] dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400 dark:hover:border-[#cc785c]/50'
      case PROVIDER_ANTHROPIC:
        return active
          ? 'border-orange-500 bg-orange-50 text-orange-700 dark:bg-orange-500/15 dark:text-orange-300 dark:border-orange-400'
          : 'border-gray-200 bg-white text-gray-600 hover:border-orange-300 hover:text-orange-700 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400 dark:hover:border-orange-500/50'
      case PROVIDER_GEMINI:
        return active
          ? 'border-[#8e8b82] bg-[#f5f0e8] text-[#504f49] dark:border-[#d8cec2]/60 dark:bg-[#8e8b82]/15 dark:text-[#d8cec2]'
          : 'border-gray-200 bg-white text-gray-600 hover:border-[#b8afa4] hover:text-[#504f49] dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400 dark:hover:border-[#8e8b82]/50'
      case PROVIDER_GROK:
        return active
          ? 'border-slate-500 bg-slate-50 text-slate-700 dark:border-slate-400 dark:bg-slate-800 dark:text-slate-200'
          : 'border-gray-200 bg-white text-gray-600 hover:border-slate-300 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400'
      case PROVIDER_KIMI:
        return active
          ? 'border-pink-500 bg-pink-50 text-pink-700 dark:border-pink-400 dark:bg-pink-500/15 dark:text-pink-300'
          : 'border-gray-200 bg-white text-gray-600 hover:border-pink-300 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400'
      case PROVIDER_ZHIPU:
        return active
          ? 'border-indigo-500 bg-indigo-50 text-indigo-700 dark:border-indigo-400 dark:bg-indigo-500/15 dark:text-indigo-300'
          : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400'
      case PROVIDER_DEEPSEEK:
        return active
          ? 'border-teal-500 bg-teal-50 text-teal-700 dark:border-teal-400 dark:bg-teal-500/15 dark:text-teal-300'
          : 'border-gray-200 bg-white text-gray-600 hover:border-teal-300 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400'
      default:
        return active
          ? 'border-gray-400 bg-gray-50 text-gray-700 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-200'
          : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400'
    }
  }

  function formatLatency(ms: number | null | undefined): string {
    if (ms == null) return t('monitorCommon.latencyEmpty')
    return String(Math.round(ms))
  }

  function formatPercent(v: number | null | undefined): string {
    if (v == null || Number.isNaN(v)) return '-'
    return `${v.toFixed(2)}%`
  }

  function formatAvailability(row: AvailabilityRow): string {
    if (!row.primary_status) return '-'
    return formatPercent(row.availability_7d)
  }

  function formatRelativeTime(iso: string | null | undefined): string {
    if (!iso) return t('monitorCommon.latencyEmpty')
    const ts = Date.parse(iso)
    if (Number.isNaN(ts)) return t('monitorCommon.latencyEmpty')
    const diffSec = Math.max(0, Math.floor((Date.now() - ts) / 1000))
    if (diffSec < 60) return t('monitorCommon.relativeSecondsAgo', { n: diffSec })
    const diffMin = Math.floor(diffSec / 60)
    if (diffMin < 60) return t('monitorCommon.relativeMinutesAgo', { n: diffMin })
    const diffHour = Math.floor(diffMin / 60)
    if (diffHour < 24) return t('monitorCommon.relativeHoursAgo', { n: diffHour })
    const diffDay = Math.floor(diffHour / 24)
    return t('monitorCommon.relativeDaysAgo', { n: diffDay })
  }

  return {
    statusLabel,
    statusBadgeClass,
    providerLabel,
    checkModeLabel,
    formatMonitorModel,
    checkModeBadgeClass,
    providerBadgeClass,
    providerPickerClass,
    formatLatency,
    formatPercent,
    formatAvailability,
    formatRelativeTime,
  }
}

/**
 * Map availability percent to the warm console palette.
 * Returns undefined for null/NaN so callers can fall back to a neutral colour.
 */
export function hslForPct(pct: number | null | undefined): string | undefined {
  if (pct === null || pct === undefined || Number.isNaN(pct)) return undefined
  const clamped = Math.max(0, Math.min(100, pct))
  if (clamped >= 95) return '#a9583e'
  if (clamped >= 75) return '#cc785c'
  if (clamped >= 50) return '#b45309'
  return '#dc2626'
}

/**
 * Tailwind gradient class for the provider icon tile background.
 */
export function providerGradient(provider: string): string {
  switch (provider) {
    case PROVIDER_OPENAI:
      return 'bg-gradient-to-br from-[#fffaf5] to-[#f3e7df] dark:from-[#cc785c]/10 dark:to-[#cc785c]/20'
    case PROVIDER_ANTHROPIC:
      return 'bg-gradient-to-br from-orange-50 to-amber-100 dark:from-orange-500/10 dark:to-amber-500/20'
    case PROVIDER_GEMINI:
      return 'bg-gradient-to-br from-[#f5f0e8] to-[#e7ded2] dark:from-[#8e8b82]/10 dark:to-[#8e8b82]/20'
    case PROVIDER_GROK:
      return 'bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-800 dark:to-slate-700'
    case PROVIDER_KIMI:
      return 'bg-gradient-to-br from-pink-50 to-pink-100 dark:from-pink-500/10 dark:to-pink-500/20'
    case PROVIDER_ZHIPU:
      return 'bg-gradient-to-br from-indigo-50 to-indigo-100 dark:from-indigo-500/10 dark:to-indigo-500/20'
    case PROVIDER_DEEPSEEK:
      return 'bg-gradient-to-br from-teal-50 to-teal-100 dark:from-teal-500/10 dark:to-teal-500/20'
    default:
      return 'bg-gradient-to-br from-gray-100 to-gray-200 dark:from-dark-700 dark:to-dark-600'
  }
}
