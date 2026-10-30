<template>
  <section v-if="!data || (data.enabled && data.status !== 'disabled')" class="reference-scores" aria-label="降智雷达 · 第三方参考评分">
    <header class="reference-heading">
      <h2><span class="reference-mark" aria-hidden="true">◎</span>降智雷达 <span class="reference-subtitle">· 第三方参考评分</span><span v-if="data && data.score_max != null" class="reference-max">满分 {{ formatScore(data.score_max) }}</span></h2>
      <a href="https://codexradar.com/#stationmaster-round" target="_blank" rel="noopener noreferrer">查看来源 <span aria-hidden="true">↗</span></a>
    </header>
    <p class="reference-disclaimer">第三方 Codex 订阅测试，不代表本站分组或账号的实测成绩。</p>
    <p v-if="loading && !data" class="reference-message" role="status">正在加载参考评分…</p>
    <template v-else>
      <p v-if="failed || data?.status === 'stale'" class="reference-delayed" role="status">{{ data && data.status !== 'unavailable' ? '来源更新延迟，保留上次成功同步的参考评分。' : '参考评分暂时不可用，请稍后再试。' }}</p>
      <p v-else-if="!data || data.status === 'unavailable'" class="reference-message" role="status">参考评分暂时不可用，请稍后再试。</p>
      <p v-if="data?.enabled && !data.items.length" class="reference-message" role="status">管理员尚未选择要展示的模型与思考档位。</p>
      <div v-if="data?.items.length" class="reference-strip" role="region" aria-label="第三方模型参考评分，可横向滚动" tabindex="0">
        <article v-for="group in modelGroups" :key="group.model" class="reference-card" :aria-label="group.displayName">
          <div class="reference-model"><span class="reference-icon"><ModelIcon :model="group.model" size="21px" /></span><strong :title="group.model">{{ group.displayName }}</strong></div>
          <div class="reference-efforts">
            <div v-for="item in group.items" :key="item.effort" class="reference-effort-row" :class="{ 'reference-effort-best': item.status === 'available' && item.score != null && item.score === group.bestScore && data.score_max != null }" :aria-label="`${group.displayName} · ${item.effort}`">
              <span class="reference-effort">{{ item.effort }}</span>
              <div v-if="item.status === 'available' && item.score != null && data.score_max != null" class="reference-value"><span v-if="item.score === group.bestScore" class="reference-best" role="img" aria-label="本模型已展示档位最高分" title="本模型已展示档位最高分（并列同样标记）"><svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="m2 5 5 5 5-8 5 8 5-5-3 13H5L2 5Zm3 15h14v2H5v-2Z" /></svg></span><strong>{{ formatScore(item.score) }}</strong></div>
              <p v-else class="reference-unrecorded">{{ itemStatus(item.status) }}</p>
            </div>
          </div>
        </article>
      </div>
      <div v-if="data?.synced_at" class="reference-metadata">
        <span>来源测试：<time :datetime="data.test_started_at || undefined" :title="data.test_started_at || undefined">{{ formatTime(data.test_started_at) }}</time><template v-if="data.test_completed_at"> — <time :datetime="data.test_completed_at" :title="data.test_completed_at">{{ formatTime(data.test_completed_at) }}</time></template></span>
        <span>来源批次：{{ data.round_id || '未提供' }}</span>
        <span>最近成功同步：<time :datetime="data.synced_at" :title="data.synced_at">{{ formatTime(data.synced_at) }}</time></span>
        <span v-if="data.source_snapshot_at">来源快照：<time :datetime="data.source_snapshot_at" :title="data.source_snapshot_at">{{ formatTime(data.source_snapshot_at) }}</time></span>
      </div>
    </template>
    <p class="reference-method">四任务加权总分：综合、庞贝、鹈鹕与糖果任务按来源权重汇总，满分以来源为准；未收录的组合不补零。</p>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import type { PelicanReferenceItem, PelicanReferenceScores } from '@/api/pelicanTests'

const props = withDefaults(defineProps<{ data?: PelicanReferenceScores | null; loading?: boolean; failed?: boolean }>(), { data: null, loading: false, failed: false })
const modelGroups = computed(() => {
  const groups = new Map<string, { model: string; displayName: string; items: PelicanReferenceItem[] }>()
  for (const item of props.data?.items || []) {
    const group = groups.get(item.model)
    if (group) group.items.push(item)
    else groups.set(item.model, { model: item.model, displayName: item.display_name || item.model, items: [item] })
  }
  return [...groups.values()].map(group => {
    const available = group.items.filter(item => item.status === 'available' && item.score != null && Number.isFinite(item.score))
    return { ...group, bestScore: available.length ? Math.max(...available.map(item => item.score!)) : null }
  })
})

const scoreFormatter = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 })
const timeFormatter = new Intl.DateTimeFormat('zh-CN', { timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false })
function formatScore(score: number): string { return scoreFormatter.format(score) }
function formatTime(value: string | null): string {
  if (!value) return '未提供'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '未提供' : `${timeFormatter.format(date)} (UTC+8)`
}
function itemStatus(status: PelicanReferenceItem['status']): string {
  return status === 'unsupported' ? '来源不支持此档位' : status === 'not_recorded' ? '来源未收录此组合' : '参考评分暂不可用'
}
</script>

<style scoped>
.reference-scores { min-width: 0; margin: 0 0 16px; padding: 13px 14px 11px; border: 1px solid var(--pelican-border, #d8cec2); border-radius: 12px; background: var(--pelican-surface, #faf9f5); color: var(--pelican-text, #141413); }
.reference-heading { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 8px; }
.reference-heading h2 { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; margin: 0; font-size: 14px; font-weight: 800; }
.reference-subtitle, .reference-max { font-size: 12px; font-weight: 500; color: var(--pelican-muted, #6c6a64); }
.reference-mark { color: var(--pelican-accent, #a34e34); font-size: 22px; line-height: 1; }
.reference-heading a { flex-shrink: 0; color: var(--pelican-accent, #a34e34); font-size: 12px; text-decoration: none; }
.reference-heading a:hover { text-decoration: underline; }
.reference-disclaimer, .reference-method { margin: 7px 0 0; color: var(--pelican-muted, #6c6a64); font-size: 11px; line-height: 1.6; }
.reference-strip { display: flex; gap: 9px; width: 100%; min-width: 0; overflow-x: auto; overscroll-behavior-x: contain; margin-top: 10px; padding-bottom: 4px; scrollbar-width: thin; }
.reference-card { box-sizing: border-box; flex: 1 0 208px; min-width: 0; padding: 8px 10px; border: 1px solid var(--pelican-border, #d8cec2); border-radius: 9px; background: var(--pelican-input, #fffdf9); }
.reference-model { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; }
.reference-icon { display: inline-flex; align-items: center; justify-content: center; flex: 0 0 27px; height: 27px; border-radius: 6px; background: #fff; color: #141413; }
.reference-model strong { min-width: 0; font-size: 12px; overflow-wrap: anywhere; }
.reference-efforts { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 4px 10px; margin-top: 6px; }
.reference-effort-row { position: relative; display: flex; align-items: center; justify-content: space-between; gap: 2px; min-width: 0; padding-top: 4px; border-top: 1px solid var(--pelican-border, #d8cec2); }
.reference-best { display: inline-flex; flex: 0 0 10px; width: 10px; height: 10px; color: #9a6700; }
.reference-best svg { width: 100%; height: 100%; }
.reference-effort-best .reference-value { color: #000; }
.dark .reference-best { color: #f1c35e; }
.dark .reference-effort-best .reference-value { color: #fff; }
.reference-effort { flex-shrink: 0; color: var(--pelican-muted, #6c6a64); font-size: 10px; line-height: 1.3; }
.reference-value { display: inline-flex; align-items: center; gap: 2px; flex-shrink: 0; margin-left: auto; color: var(--pelican-muted, #6c6a64); text-align: right; font-variant-numeric: tabular-nums; }
.reference-value strong { font-size: 17px; line-height: 1.2; letter-spacing: -.4px; }
.reference-unrecorded { min-width: 0; margin: 0 0 0 auto; color: var(--pelican-muted, #6c6a64); font-size: 10px; line-height: 1.4; text-align: right; overflow-wrap: anywhere; }
.reference-metadata { display: flex; flex-wrap: wrap; column-gap: 14px; row-gap: 3px; margin-top: 7px; color: var(--pelican-muted, #6c6a64); font-size: 10px; line-height: 1.6; overflow-wrap: anywhere; }
.reference-message, .reference-delayed { margin: 10px 0; font-size: 12px; line-height: 1.6; color: var(--pelican-muted, #6c6a64); }
.reference-delayed { color: #94600e; }
:global(.dark) .reference-delayed { color: #f8cb80; }
.reference-strip:focus-visible, .reference-heading a:focus-visible { outline: 2px solid var(--pelican-accent, #a34e34); outline-offset: 2px; }
@media (max-width: 640px) { .reference-scores { padding: 11px; } }
</style>
