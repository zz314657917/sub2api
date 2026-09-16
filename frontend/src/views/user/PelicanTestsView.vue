<template>
  <AppLayout>
    <div class="pelican-page">
      <p class="demo-note"><span>实时结果</span> 仅展示当前有权限访问的分组与账号测试结果。</p>
      <template v-if="authStore.isAdmin"><button class="admin-toggle" type="button" @click="showPlans = !showPlans">{{ showPlans ? '收起' : '管理' }}测试计划</button><PelicanPlansPanel v-if="showPlans" :plans="plans" :groups="adminGroups" :loading="plansLoading" :error="plansError" :busy-id="planBusyId" :save-version="saveVersion" @run="runPlan" @save="savePlan" @delete="removePlan" /></template>
      <section class="toolbar" aria-label="筛选鹈鹕测试">
        <label class="search-box">
          <Icon name="search" size="sm" />
          <input v-model.trim="search" type="search" placeholder="搜索账号 ID" aria-label="搜索账号 ID" />
        </label>
        <select v-model.number="groupFilter" aria-label="筛选分组">
          <option value="">全部分组</option>
          <option v-for="item in groups" :key="item.id" :value="item.id">{{ item.name }}</option>
        </select>
        <div class="refresh-controls">
          <select v-model.number="refreshSeconds" aria-label="自动刷新间隔">
            <option :value="0">自动刷新：关闭</option>
            <option :value="15">每 15 秒刷新</option>
            <option :value="30">每 30 秒刷新</option>
            <option :value="60">每 60 秒刷新</option>
          </select>
          <button type="button" :disabled="!hasFilters" @click="clearFilters">清除筛选</button>
          <button class="refresh-button" type="button" aria-label="刷新动画列表" title="刷新测试结果" @click="refresh"><Icon name="refresh" size="sm" /></button>
        </div>
      </section>
      <div class="list-summary"><span>{{ total }} 个结果 <i>/</i> 查看当前动画与历史表现</span><span role="status">{{ refreshedAt }} 更新</span></div>

      <div v-if="loading" class="empty-state" aria-live="polite">正在加载测试结果…</div>
      <div v-else-if="error" class="empty-state" role="alert"><h2>无法加载测试结果</h2><p>{{ error }}</p><button type="button" @click="refresh">重试</button></div>
      <div v-else-if="items.length" class="gallery">
        <article v-for="item in items" :key="item.result_id" class="pelican-card" :aria-label="`账号 ${item.account_id}`">
          <header class="card-heading">
            <div><span class="team-badge">{{ item.group_name }}</span><strong>#{{ item.account_id }}</strong></div>
            <button class="history-count" type="button" :aria-label="`账号 ${item.account_id} 历史 ${item.history_count} 次`" @click="openHistory(item, $event)"><Icon name="clock" size="xs" /> {{ item.history_count }}</button>
          </header>
          <PelicanSandboxFrame v-if="cardHtml[item.result_id]" :html="cardHtml[item.result_id]" :replay-key="replayVersions[`card-${item.result_id}`] ?? 0" />
          <div v-else class="artwork-placeholder">{{ item.artwork_result_id ? (artworkLoading[item.result_id] ? '正在加载作品…' : '作品当前不可访问') : '本次暂无可展示的成功作品' }}</div>
          <div class="card-actions">
            <button :disabled="!item.artwork_result_id" type="button" aria-label="放大播放" title="放大播放" @click="openArtwork(item, $event)"><Icon name="eye" size="sm" /></button>
            <button :disabled="!item.artwork_result_id" type="button" aria-label="重新播放" title="重新播放" @click="replay(`card-${item.result_id}`)"><Icon name="refresh" size="sm" /></button>
          </div>
          <div class="result-row"><span :class="item.status === 'success' ? 'success' : 'failed'">● {{ statusText(item.status) }}</span><span>{{ formatDuration(item.latency_ms) }}</span></div>
          <p class="html-meta">{{ item.model_id }} · HTML {{ formatChars(item.char_count) }} 字符 · 阈值 {{ formatChars(item.min_chars) }}</p>
          <p v-if="item.error_message" class="html-meta">{{ item.error_message }}</p><p class="last-success">完成于 {{ formatDate(item.finished_at) }}</p>
        </article>
      </div>
      <div v-else class="empty-state"><Icon name="search" size="lg" /><h2>暂无测试结果</h2><p>当前有权限的分组还没有可展示的测试结果。</p><button v-if="hasFilters" type="button" @click="clearFilters">清除筛选</button></div>

      <footer class="pagination">
        <span>显示 {{ total ? (page - 1) * pageSize + 1 : 0 }} 至 {{ Math.min(page * pageSize, total) }}，共 {{ total }} 条结果</span>
        <div class="page-controls">
          <label>每页 <select v-model.number="pageSize" aria-label="每页条数"><option :value="6">6</option><option :value="12">12</option><option :value="24">24</option></select></label>
          <button type="button" :disabled="page === 1" @click="page--">上一页</button>
          <span>{{ page }} / {{ totalPages }}</span>
          <button type="button" :disabled="page >= totalPages" @click="page++">下一页</button>
        </div>
      </footer>
      <BaseDialog v-if="dialogMode && selectedEntry" :show="true" :title="dialogTitle" width="wide" :close-on-click-outside="true" @close="closeDialog">
        <div ref="dialogContent" class="dialog-content">
          <template v-if="dialogMode === 'history'">
            <p class="dialog-note">历史记录</p><p v-if="historyLoading" class="dialog-note">加载中…</p><p v-else-if="historyError" class="dialog-note">{{ historyError }}</p>
            <div class="history-grid">
              <article v-for="revision in history" :key="revision.id" class="history-record">
                <div class="result-row"><span :class="revision.status === 'success' ? 'success' : 'failed'">● {{ statusText(revision.status) }}</span><span>{{ formatDate(revision.finished_at) }}</span></div>
                <p class="dialog-note">{{ revision.model_id }} · {{ formatDuration(revision.latency_ms) }}</p>
                <div class="history-actions">
                  <button :disabled="revision.status !== 'success'" type="button" @click="openRevision(revision, $event)">放大播放</button>
                </div>
                <p class="html-meta">HTML {{ formatChars(revision.char_count) }} 字符 · 阈值 {{ formatChars(revision.min_chars) }}</p>
              </article>
            </div>
          </template>
          <template v-else-if="selectedRevision">
            <button v-if="fromHistory" class="back-button" type="button" @click="backToHistory">← 返回历史记录</button>
            <div class="enlarged-art"><PelicanSandboxFrame v-if="selectedRevision.html" :html="selectedRevision.html" :replay-key="replayVersions[`large-${selectedRevision.id}`] ?? 0" /></div>
            <div class="history-actions"><span class="success">● 本轮成功 · {{ formatDuration(selectedRevision.latency_ms) }}</span><button type="button" @click="replay(`large-${selectedRevision.id}`)">重新播放</button></div>
          </template>
        </div>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import PelicanSandboxFrame from '@/components/pelican/PelicanSandboxFrame.vue'
import PelicanPlansPanel from '@/components/pelican/PelicanPlansPanel.vue'
import { buildPelicanListParams, createPelicanPlan, deletePelicanPlan, getPelicanResult, listPelicanHistory, listPelicanPlans, listPelicanTests, runPelicanPlan, updatePelicanPlan, type PelicanEntry, type PelicanPlan, type PelicanPlanInput, type PelicanResult } from '@/api/pelicanTests'
import { getAll as getAllAdminGroups } from '@/api/admin/groups'
import { useAuthStore } from '@/stores/auth'

const search = ref('')
const groupFilter = ref<number | ''>('')
const page = ref(1)
const pageSize = ref(24)
const refreshSeconds = ref(0)
const refreshedAt = ref(formatNow())
const replayVersions = ref<Record<string, number>>({})
const dialogMode = ref<'history' | 'artwork' | null>(null)
const selectedEntry = ref<PelicanEntry | null>(null)
const selectedRevision = ref<PelicanResult | null>(null)
const fromHistory = ref(false)
const dialogContent = ref<HTMLElement | null>(null)
const authStore = useAuthStore()
const items = ref<PelicanEntry[]>([])
const groups = ref<{ id: number; name: string }[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref('')
const history = ref<PelicanResult[]>([])
const historyLoading = ref(false)
const historyError = ref('')
const cardHtml = ref<Record<number, string>>({})
const artworkLoading = ref<Record<number, boolean>>({})
const showPlans = ref(false)
const plans = ref<PelicanPlan[]>([])
const plansLoading = ref(false)
const plansError = ref('')
const planBusyId = ref<number | null>(null)
const adminGroups = ref<Array<{ id: number; name: string }>>([])
let opener: HTMLElement | null = null
let interval: ReturnType<typeof setInterval> | undefined
let listRequestVersion = 0
let historyRequestVersion = 0
let dialogRequestVersion = 0
let listAbort: AbortController | undefined
let dialogAbort: AbortController | undefined
let disposed = false
const saveVersion = ref(0)

const hasFilters = computed(() => Boolean(search.value || groupFilter.value !== ''))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const dialogTitle = computed(() => `账号 #${selectedEntry.value?.account_id} · ${dialogMode.value === 'history' ? '测试历史' : '放大播放'}`)

watch([search, groupFilter, pageSize], () => { closeDialog(); if (page.value !== 1) page.value = 1; else void refresh() })
watch(page, () => void refresh())
watch(() => [authStore.user?.id, authStore.isAdmin], () => { closeDialog(); plans.value = []; adminGroups.value = []; showPlans.value = false; void refresh() })
watch(refreshSeconds, seconds => {
  if (interval) clearInterval(interval)
  interval = seconds ? setInterval(refresh, seconds * 1000) : undefined
})
onMounted(() => { document.addEventListener('keydown', trapFocus); void refresh() })
onBeforeUnmount(() => {
  disposed = true
  listRequestVersion++; historyRequestVersion++; dialogRequestVersion++
  listAbort?.abort(); dialogAbort?.abort()
  if (interval) clearInterval(interval)
  document.removeEventListener('keydown', trapFocus)
})

function formatNow(): string {
  return new Intl.DateTimeFormat('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(new Date())
}
function formatChars(value: number): string { return value.toLocaleString('zh-CN') }
function statusText(status: PelicanEntry['status']): string { return status === 'success' ? '成功' : status === 'failed' ? '失败' : '已跳过' }
function formatDate(value?: string | null): string { const date = value ? new Date(value) : null; return date && !Number.isNaN(date.getTime()) ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'short', timeStyle: 'short' }).format(date) : '时间未知' }
function formatDuration(value: number): string { return value > 1000 ? `${(value / 1000).toFixed(1)} 秒` : `${value} 毫秒` }
async function refresh(): Promise<void> {
  if (disposed) return
  const version = ++listRequestVersion
  listAbort?.abort()
  const controller = new AbortController()
  listAbort = controller
  cardHtml.value = {}; artworkLoading.value = {}
  loading.value = true; error.value = ''
  try {
    const data = await listPelicanTests(buildPelicanListParams(groupFilter.value, search.value, page.value, pageSize.value), controller.signal)
    if (version !== listRequestVersion) return
    const lastPage = Math.max(1, Math.ceil(data.total / pageSize.value))
    if (page.value > lastPage) { total.value = data.total; page.value = lastPage; return }
    items.value = data.items; groups.value = data.groups; total.value = data.total; page.value = data.page; refreshedAt.value = formatNow()
    if (selectedEntry.value && !data.items.some(item => item.plan_id === selectedEntry.value?.plan_id && item.account_id === selectedEntry.value?.account_id)) closeDialog()
    void loadVisibleArtwork(data.items, version, controller.signal)
  } catch (reason) { if (version === listRequestVersion) { error.value = messageOf(reason); items.value = []; total.value = 0; closeDialog() } } finally { if (version === listRequestVersion) loading.value = false }
}
async function loadVisibleArtwork(entries: PelicanEntry[], version: number, signal: AbortSignal): Promise<void> {
  const queue = entries.filter(entry => entry.artwork_result_id && !cardHtml.value[entry.result_id])
  queue.forEach(entry => { artworkLoading.value[entry.result_id] = true })
  await Promise.all(Array.from({ length: Math.min(3, queue.length) }, async () => {
    while (queue.length && version === listRequestVersion) {
      const entry = queue.shift()!
      try { const result = await getPelicanResult(entry.artwork_result_id!, signal); if (version === listRequestVersion && result.html) cardHtml.value[entry.result_id] = result.html } catch { /* authorization may change between list and detail */ } finally { if (version === listRequestVersion) artworkLoading.value[entry.result_id] = false }
    }
  }))
}
function clearFilters(): void { search.value = ''; groupFilter.value = '' }
function replay(id: string): void { replayVersions.value[id] = (replayVersions.value[id] ?? 0) + 1 }
async function openHistory(entry: PelicanEntry, event: MouseEvent): Promise<void> {
  dialogRequestVersion++; dialogAbort?.abort(); dialogAbort = new AbortController()
  opener = event.currentTarget as HTMLElement
  selectedEntry.value = entry
  selectedRevision.value = null
  fromHistory.value = false
  dialogMode.value = 'history'
  const version = ++historyRequestVersion
  history.value = []; historyLoading.value = true; historyError.value = ''
  try { const result = await listPelicanHistory({ plan_id: entry.plan_id, account_id: entry.account_id }, dialogAbort.signal); if (version === historyRequestVersion) history.value = result } catch (reason) { if (version === historyRequestVersion) historyError.value = messageOf(reason) } finally { if (version === historyRequestVersion) historyLoading.value = false }
}
async function openArtwork(entry: PelicanEntry, event: MouseEvent): Promise<void> {
  const version = ++dialogRequestVersion
  historyRequestVersion++; dialogAbort?.abort(); dialogAbort = new AbortController()
  opener = event.currentTarget as HTMLElement; selectedEntry.value = entry; fromHistory.value = false
  if (!entry.artwork_result_id) return
  try { const result = await getPelicanResult(entry.artwork_result_id, dialogAbort.signal); if (version !== dialogRequestVersion) return; selectedRevision.value = result; dialogMode.value = 'artwork' } catch (reason) { if (version === dialogRequestVersion) error.value = messageOf(reason) }
  void nextTick(() => dialogContent.value?.querySelector<HTMLButtonElement>('button')?.focus())
}
async function openRevision(revision: PelicanResult, event: MouseEvent): Promise<void> {
  const version = ++dialogRequestVersion
  dialogAbort?.abort(); dialogAbort = new AbortController()
  fromHistory.value = true; opener = event.currentTarget as HTMLElement
  try { const result = await getPelicanResult(revision.id, dialogAbort.signal); if (version !== dialogRequestVersion) return; selectedRevision.value = result; dialogMode.value = 'artwork' } catch (reason) { if (version === dialogRequestVersion) historyError.value = messageOf(reason); return }
  void nextTick(() => dialogContent.value?.querySelector<HTMLButtonElement>('button')?.focus())
}
function backToHistory(): void {
  dialogRequestVersion++; dialogAbort?.abort()
  dialogMode.value = 'history'
  selectedRevision.value = null
  void nextTick(() => dialogContent.value?.querySelector<HTMLButtonElement>('button')?.focus())
}
function messageOf(reason: unknown): string { return typeof (reason as { message?: unknown })?.message === 'string' ? (reason as { message: string }).message : '请求失败，请稍后重试。' }
async function loadPlans(): Promise<void> { if (!authStore.isAdmin) return; plansLoading.value = true; plansError.value = ''; try { const [planData, groupData] = await Promise.all([listPelicanPlans(), getAllAdminGroups()]); plans.value = planData; adminGroups.value = groupData.map(group => ({ id: group.id, name: group.name })) } catch (reason) { plansError.value = messageOf(reason) } finally { plansLoading.value = false } }
watch(showPlans, visible => { if (visible) void loadPlans() })
async function runPlan(id: number): Promise<void> { planBusyId.value = id; try { await runPelicanPlan(id); await loadPlans() } catch (reason) { plansError.value = messageOf(reason) } finally { planBusyId.value = null } }
async function savePlan(id: number | null, input: PelicanPlanInput): Promise<void> { planBusyId.value = id ?? -1; plansError.value = ''; try { if (id) await updatePelicanPlan(id, input); else await createPelicanPlan({ ...input, enabled: false }); saveVersion.value++; await loadPlans() } catch (reason) { plansError.value = messageOf(reason) } finally { planBusyId.value = null } }
async function removePlan(id: number): Promise<void> { planBusyId.value = id; try { await deletePelicanPlan(id); await loadPlans() } catch (reason) { plansError.value = messageOf(reason) } finally { planBusyId.value = null } }
function closeDialog(): void {
  historyRequestVersion++; dialogRequestVersion++; dialogAbort?.abort()
  dialogMode.value = null
  selectedRevision.value = null
  selectedEntry.value = null
  history.value = []
  const target = opener
  opener = null
  void nextTick(() => target?.focus())
}
function trapFocus(event: KeyboardEvent): void {
  if (!dialogMode.value || event.key !== 'Tab') return
  const dialog = dialogContent.value?.closest('[role="dialog"]')
  const buttons = dialog?.querySelectorAll<HTMLElement>('button:not(:disabled), [tabindex="0"]')
  if (!buttons?.length) return
  const first = buttons[0]
  const last = buttons[buttons.length - 1]
  if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}
</script>

<style scoped>
.pelican-page, .dialog-content {
  --pelican-text: var(--console-text, #141413);
  --pelican-muted: var(--console-muted, #6c6a64);
  --pelican-border: var(--console-border, rgba(216, 206, 194, .68));
  --pelican-surface: var(--console-surface, rgba(250, 249, 245, .88));
  --pelican-input: var(--console-input, rgba(250, 249, 245, .94));
  --pelican-hover: var(--console-surface-hover, #fffaf5);
  --pelican-accent: var(--console-accent, theme('colors.primary.500'));
  --pelican-accent-soft: var(--console-accent-soft, rgba(204, 120, 92, .1));
  color: var(--pelican-text);
}
.pelican-page { font-size: 12px; min-width: 0; }
.admin-toggle { margin: 0 0 12px; border: 1px solid var(--pelican-border); border-radius: 7px; background: var(--pelican-input); color: var(--pelican-text); padding: 7px 10px; cursor: pointer; }
.demo-note { color: var(--pelican-muted); margin: 0 0 18px; display: flex; align-items: center; gap: 8px; font-size: 11px; }
.demo-note > span { color: var(--pelican-accent); background: var(--pelican-accent-soft); border-radius: 5px; padding: 3px 7px; white-space: nowrap; }
.toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 12px; }
.search-box { display: flex; align-items: center; gap: 10px; width: 216px; padding: 0 12px; color: var(--pelican-muted); }
.search-box, select, .toolbar button { border: 1px solid var(--pelican-border); border-radius: 10px; background: var(--pelican-input); min-height: 38px; }
.search-box input { width: 100%; min-width: 0; outline: none; border: 0; background: transparent; color: var(--pelican-text); font: inherit; }
.search-box input::placeholder { color: var(--pelican-muted); }
select { font: inherit; color: inherit; padding: 8px 30px 8px 12px; min-width: 190px; cursor: pointer; }
.refresh-controls { display: flex; align-items: center; gap: 8px; margin-left: auto; }
.refresh-controls select { min-width: 140px; }
.toolbar button { color: inherit; padding: 8px 12px; cursor: pointer; }
.toolbar button:disabled { opacity: .4; cursor: default; }
.refresh-button { display: inline-flex; align-items: center; justify-content: center; }
.list-summary { display: flex; justify-content: space-between; gap: 10px; color: var(--pelican-muted); margin: 12px 0 15px; font-size: 11px; }
.list-summary i { margin: 0 8px; color: var(--pelican-muted); font-style: normal; }
.gallery { display: grid; grid-template-columns: repeat(auto-fill, minmax(255px, 1fr)); gap: 12px; }
.pelican-card { min-width: 0; border: 1px solid var(--pelican-border); background: var(--pelican-surface); box-shadow: var(--console-shadow, none); border-radius: 12px; padding: 9px; }
.artwork-placeholder { display: grid; place-items: center; min-height: 180px; border-radius: 8px; background: var(--pelican-accent-soft); color: var(--pelican-muted); font-size: 11px; }
.card-heading { display: flex; align-items: center; justify-content: space-between; gap: 5px; margin-bottom: 9px; }
.card-heading > div { display: flex; align-items: center; gap: 7px; }
.card-heading strong { font-size: 12px; }
.team-badge { background: var(--pelican-accent-soft); color: var(--pelican-accent); border-radius: 5px; font-size: 11px; font-weight: 700; padding: 3px 6px; }
.history-count { display: inline-flex; align-items: center; gap: 4px; border: 0; background: transparent; color: var(--pelican-muted); cursor: pointer; padding: 3px; }
.card-actions { display: flex; gap: 6px; margin: 6px 0; }
.card-actions button { border: 0; background: transparent; color: var(--pelican-accent); cursor: pointer; display: inline-flex; padding: 5px; border-radius: 5px; }
.card-actions button:hover { background: var(--pelican-accent-soft); }
.result-row { display: flex; justify-content: space-between; align-items: center; gap: 8px; color: var(--pelican-muted); font-size: 11px; }
.success { @apply text-emerald-700 dark:text-emerald-400; }
.html-meta { color: var(--pelican-muted); font-size: 10px; line-height: 1.6; margin: 7px 0 11px; overflow-wrap: anywhere; }
.group-meta { margin: 0 0 4px; color: var(--pelican-muted); font-size: 10px; }
.stars { @apply text-accent-500 dark:text-accent-300; letter-spacing: 1px; white-space: nowrap; }
.last-success { color: var(--pelican-muted); font-size: 10px; margin: 0 0 2px; }
.pagination { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px; margin-top: 22px; color: var(--pelican-muted); }
.page-controls { display: flex; align-items: center; gap: 10px; }
.page-controls select { min-width: 64px; margin-left: 6px; }
.page-controls button, .empty-state button, .history-actions button, .back-button { border: 1px solid var(--pelican-border); background: var(--pelican-input); color: var(--pelican-text); border-radius: 7px; padding: 7px 10px; cursor: pointer; font: inherit; }
.toolbar button:not(:disabled):hover, .page-controls button:not(:disabled):hover, .empty-state button:hover, .history-actions button:hover, .back-button:hover { background: var(--pelican-hover); border-color: var(--pelican-accent); }
.page-controls button:disabled { opacity: .4; cursor: default; }
.empty-state { text-align: center; padding: 60px 15px; border: 1px dashed var(--pelican-border); border-radius: 12px; color: var(--pelican-muted); }
.empty-state h2 { font-size: 16px; color: var(--pelican-text); margin-top: 15px; }
.empty-state p { margin: 8px 0 18px; }
.dialog-note { font-size: 11px; color: var(--pelican-muted); margin: 0 0 12px; }
.history-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 18px; }
.history-record { min-width: 0; }
.history-record .result-row { margin-bottom: 7px; }
.history-actions { display: flex; justify-content: space-between; align-items: center; gap: 10px; margin-top: 12px; font-size: 12px; }
.enlarged-art { max-width: 560px; margin: auto; }
.back-button { margin-bottom: 12px; }
button:focus-visible, select:focus-visible, .search-box:focus-within { outline: 2px solid var(--pelican-accent); outline-offset: 2px; }
/* BaseDialog teleports outside .console-shell, so its dark tokens need fallbacks. */
.dark .dialog-content {
  --pelican-text: theme('colors.dark.50');
  --pelican-muted: theme('colors.dark.400');
  --pelican-border: rgba(51, 65, 85, .82);
  --pelican-input: rgba(10, 18, 32, .86);
  --pelican-hover: rgba(30, 41, 59, .82);
  --pelican-accent: theme('colors.indigo.400');
  --pelican-accent-soft: rgba(129, 140, 248, .12);
}
@media (min-width: 1800px) { .gallery { grid-template-columns: repeat(6, minmax(0, 1fr)); } }
@media (max-width: 760px) {
  .search-box { width: 100%; min-height: 42px; }
  .toolbar > select { flex: 1; min-width: 120px; max-width: 100%; }
  .refresh-controls { width: 100%; margin: 0; }
  .refresh-controls select { min-width: 0; flex: 1; }
  .list-summary { align-items: flex-start; flex-direction: column; }
  .gallery { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .history-grid { grid-template-columns: 1fr; }
  .page-controls { flex-wrap: wrap; gap: 7px; }
}
@media (max-width: 540px) { .gallery { grid-template-columns: 1fr; } }
</style>
