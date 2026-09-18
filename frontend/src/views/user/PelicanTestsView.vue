<template>
  <AppLayout viewport>
    <div class="pelican-page">
      <div class="pelican-controls">
      <template v-if="authStore.isAdmin">
        <button class="admin-toggle" type="button" @click="showSettings = !showSettings">{{ showSettings ? '收起' : '管理' }}页面设置</button>
        <p v-if="!pelicanMetadata.enabled" class="settings-disabled-banner" role="status">鹈鹕广场当前已关闭。普通用户入口、结果访问和新的手动/定时运行均已停用；进行中的轮次会完成，现有数据和配置会保留。</p>
        <section v-if="showSettings" class="settings-editor" :aria-label="`${pelicanMetadata.displayName}页面设置`">
          <p>所有测试计划共用此提示词。当前只渲染 HTML，仍会在受限沙箱中展示。</p>
          <label class="settings-enabled"><input v-model="settingsDraft.enabled" :disabled="settingsLoading || settingsSaving" type="checkbox" aria-label="启用鹈鹕广场" /> 启用鹈鹕广场</label><small v-if="!settingsDraft.enabled">保存关闭后会隐藏普通用户入口，阻止普通用户访问结果与新的手动/定时运行。</small>
          <label>页面名称<input v-model="settingsDraft.display_name" :disabled="settingsLoading || settingsSaving" maxlength="40" aria-label="页面名称" /></label>
          <label>共享测试提示词<textarea v-model="settingsDraft.prompt" :disabled="settingsLoading || settingsSaving" maxlength="20000" rows="8" aria-label="共享测试提示词" /></label>
          <p v-if="settingsError" role="alert">{{ settingsError }}</p>
          <div><button v-if="!settingsLoaded" type="button" :disabled="settingsLoading" @click="loadSettings">重新加载</button><button type="button" :disabled="!settingsLoaded || settingsLoading || settingsSaving" @click="restoreDefaultSettings">恢复默认</button><button type="button" :disabled="!settingsLoaded || settingsLoading || settingsSaving" @click="saveSettings">{{ settingsSaving ? '保存中…' : '保存设置' }}</button></div>
        </section>
        <button class="admin-toggle" type="button" @click="showPlans = true">管理测试计划</button>
        <BaseDialog :show="showPlans" :title="`${pelicanMetadata.displayName} · 测试计划`" width="wide" :close-on-escape="!cleanupConfirmation" @close="closePlans"><p v-if="cleanupNotice" class="cleanup-notice" role="status">{{ cleanupNotice }}</p><PelicanPlansPanel v-if="showPlans" :plans="plans" :groups="adminGroups" :loading="plansLoading" :error="plansError" :busy-id="planBusyId" :save-version="saveVersion" :pelican-enabled="pelicanMetadata.enabled" @run="runPlan" @resume="resumePlan" @save="savePlan" @delete="removePlan" @cleanup-request="requestCleanup" /></BaseDialog>
        <BaseDialog :show="Boolean(cleanupConfirmation)" title="确认清理测试记录" width="narrow" :z-index="60" :close-on-escape="!cleanupBusy" @close="cancelCleanup"><div v-if="cleanupConfirmation" class="cleanup-confirmation"><p>{{ cleanupConfirmation.scope === 'failed' ? '将永久删除该计划的失败和已跳过记录，成功记录不会删除。' : '将按当前保留天数和结果数量永久删除过期记录，并始终保留最新成功记录。' }}</p><p>此操作不可撤销。删除后会刷新画廊和已打开的历史记录。</p><div><button type="button" :disabled="cleanupBusy" @click="cancelCleanup">取消</button><button type="button" :disabled="cleanupBusy" @click="confirmCleanup">{{ cleanupBusy ? '正在清理…' : '确认清理' }}</button></div></div></BaseDialog>
      </template>
      <section class="toolbar" :aria-label="`筛选${pelicanMetadata.displayName}`">
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
          <button class="refresh-button" type="button" :aria-label="`刷新${pelicanMetadata.displayName}列表`" :title="`刷新${pelicanMetadata.displayName}结果`" @click="refresh()"><Icon name="refresh" size="sm" /></button>
        </div>
      </section>
      <div class="list-summary"><span>{{ total }} 个结果 <i>/</i> 查看当前动画与历史表现</span><span role="status">{{ refreshedAt }} 更新</span></div>
      <p v-if="error && items.length" class="refresh-error" role="alert">{{ error }}</p>
      </div>

      <div ref="resultsRegion" class="results-region">
      <div v-if="loading && !items.length" class="empty-state" aria-live="polite">正在加载测试结果…</div>
      <div v-else-if="error && !items.length" class="empty-state" role="alert"><h2>无法加载测试结果</h2><p>{{ error }}</p><button type="button" @click="refresh()">重试</button></div>
      <div v-else-if="items.length" class="gallery">
        <article v-for="item in items" :key="`${item.plan_id}-${item.group_id}`" class="pelican-card" :aria-label="`${item.group_name} · 计划 ${item.plan_id}`">
          <header class="card-heading">
            <div><span class="team-badge">{{ item.group_name }}</span><strong>计划 #{{ item.plan_id }}</strong></div>
            <button class="history-count" type="button" :aria-label="`${item.group_name} 计划 ${item.plan_id} 历史 ${item.history_count} 次`" @click="openHistory(item, $event)"><Icon name="clock" size="xs" /> {{ item.history_count }}</button>
          </header>
          <button v-if="cardHtml[item.result_id]" class="artwork-preview" type="button" :aria-label="`放大查看${item.group_name} 计划 ${item.plan_id} 的作品`" @click="openArtwork(item, $event)" @keydown.enter.prevent="openArtwork(item, $event)" @keydown.space.prevent="openArtwork(item, $event)">
            <PelicanSandboxFrame :html="cardHtml[item.result_id]" :replay-key="replayVersions[`card-${item.result_id}`] ?? 0" />
            <span class="preview-overlay" aria-hidden="true" />
          </button>
          <div v-else class="artwork-placeholder">{{ item.artwork_result_id ? (artworkLoading[item.result_id] ? '正在加载作品…' : '作品当前不可访问') : '本次暂无可展示的成功作品' }}</div>
          <p v-if="item.artwork_result_id && item.artwork_result_id !== item.result_id" class="last-success">上次成功作品</p>
          <div class="card-actions">
            <button :disabled="!cardHtml[item.result_id]" type="button" aria-label="放大播放" title="放大播放" @click="openArtwork(item, $event)"><Icon name="eye" size="sm" /></button>
            <button :disabled="!item.artwork_result_id" type="button" aria-label="重新播放" title="重新播放" @click="replay(`card-${item.result_id}`)"><Icon name="refresh" size="sm" /></button>
          </div>
          <p class="html-meta">{{ item.model_id }} · {{ effortText(item.reasoning_effort) }}</p>
          <p class="last-success">{{ item.status === 'success' ? '生成于' : '测试于' }} {{ formatDate(item.finished_at) }}</p>
        </article>
      </div>
      <div v-else class="empty-state"><Icon name="search" size="lg" /><h2>暂无测试结果</h2><p>当前有权限的分组还没有可展示的测试结果。</p><button v-if="hasFilters" type="button" @click="clearFilters">清除筛选</button></div>
      </div>

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
            <p v-if="historyLoading" class="dialog-note">加载中…</p><p v-else-if="historyError" class="dialog-note" role="alert">{{ historyError }}</p>
            <p v-else-if="!history.length" class="dialog-note">暂无历史记录</p>
            <div class="history-grid">
              <article v-for="revision in visibleHistory" :key="revision.id" class="history-record">
                <div class="result-row"><span :class="statusClass(revision.status)">● {{ statusText(revision.status) }}</span><span>{{ formatDate(revision.finished_at) }}</span></div>
                <p class="dialog-note">{{ revision.model_id }} · {{ effortText(revision.reasoning_effort) }} · {{ formatDuration(revision.latency_ms) }}</p>
                <PelicanHistoryPreview v-if="revision.status === 'success'" :result-id="revision.id" @enlarge="openRevision(revision, $event)" />
                <div v-else class="history-unavailable">{{ revision.status === 'failed' ? '本轮生成失败，没有可播放的作品' : '本轮未执行，没有生成作品' }}<p v-if="revision.error_message">{{ planReasonText(revision.error_message) }}</p></div>
              </article>
            </div>
            <footer v-if="history.length" class="history-pagination">
              <span>显示 {{ (historyPage - 1) * 6 + 1 }}–{{ Math.min(historyPage * 6, history.length) }}，共 {{ history.length }} 条</span>
              <div class="page-controls"><button type="button" :disabled="historyPage === 1" @click="historyPage--">上一页</button><span>{{ historyPage }} / {{ historyPages }}</span><button type="button" :disabled="historyPage >= historyPages" @click="historyPage++">下一页</button></div>
            </footer>
          </template>
          <template v-else-if="selectedRevision">
            <button v-if="fromHistory" class="back-button" type="button" @click="backToHistory">← 返回历史记录</button>
            <div class="enlarged-art"><PelicanSandboxFrame v-if="selectedRevision.html" interactive :html="selectedRevision.html" :replay-key="replayVersions[`large-${selectedRevision.id}`] ?? 0" /></div>
            <div class="history-actions"><span class="success">● 成功作品 · {{ formatDuration(selectedRevision.latency_ms) }}</span><button type="button" @click="replay(`large-${selectedRevision.id}`)">重新播放</button></div>
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
import PelicanHistoryPreview from '@/components/pelican/PelicanHistoryPreview.vue'
import { buildPelicanListParams, cleanupPelicanPlan, createPelicanPlan, deletePelicanPlan, getPelicanResult, getPelicanTestSettings, listPelicanHistory, listPelicanPlans, listPelicanTests, resumePelicanPlan, runPelicanPlan, updatePelicanPlan, updatePelicanTestSettings, type PelicanEntry, type PelicanPlan, type PelicanPlanInput, type PelicanResult, type PelicanTestSettings } from '@/api/pelicanTests'
import { getAll as getAllAdminGroups } from '@/api/admin/groups'
import { useAuthStore } from '@/stores/auth'
import { usePelicanMetadataStore } from '@/stores/pelicanMetadata'
import { useAppStore } from '@/stores/app'

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
const resultsRegion = ref<HTMLElement | null>(null)
const authStore = useAuthStore()
const appStore = useAppStore()
const pelicanMetadata = usePelicanMetadataStore()
const items = ref<PelicanEntry[]>([])
const groups = ref<{ id: number; name: string }[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref('')
const history = ref<PelicanResult[]>([])
const historyPage = ref(1)
const historyPages = computed(() => Math.max(1, Math.ceil(history.value.length / 6)))
const visibleHistory = computed(() => history.value.slice((historyPage.value - 1) * 6, historyPage.value * 6))
watch(historyPages, pages => { historyPage.value = Math.min(historyPage.value, pages) })
const historyLoading = ref(false)
const historyError = ref('')
const cardHtml = ref<Record<number, string>>({})
const cardArtworkIDs = ref<Record<number, number | null>>({})
const artworkLoading = ref<Record<number, boolean>>({})
const showPlans = ref(false)
const plans = ref<PelicanPlan[]>([])
const plansLoading = ref(false)
const plansError = ref('')
const planBusyId = ref<number | null>(null)
const cleanupConfirmation = ref<{ plan: PelicanPlan; scope: 'failed' | 'expired' } | null>(null)
const cleanupBusy = ref(false)
const cleanupNotice = ref('')
const adminGroups = ref<Array<{ id: number; name: string }>>([])
let opener: HTMLElement | null = null
let cleanupOpener: HTMLElement | null = null
let interval: ReturnType<typeof setInterval> | undefined
let listRequestVersion = 0
let historyRequestVersion = 0
let dialogRequestVersion = 0
let listAbort: AbortController | undefined
let dialogAbort: AbortController | undefined
let disposed = false
let settingsRequestVersion = 0
const saveVersion = ref(0)
const showSettings = ref(false)
const settingsLoading = ref(false)
const settingsSaving = ref(false)
const settingsLoaded = ref(false)
const settingsError = ref('')
const defaultSettings: PelicanTestSettings = { display_name: '鹈鹕测试', prompt: '请生成一个精致、完整、独立的中文海边鹈鹕骑自行车动画 HTML：鹈鹕双脚连续踩踏，车轮持续转动，海浪和云层轻轻移动，循环流畅。使用内联 CSS keyframes 和 SVG，支持不同屏幕宽度。不得包含 JavaScript、外部资源、链接或网络请求。仅输出完整 HTML 文档，不加 Markdown 代码围栏或解释。', enabled: true }
const settingsDraft = ref<PelicanTestSettings>({ ...defaultSettings })

const hasFilters = computed(() => groupFilter.value !== '')
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const dialogTitle = computed(() => `${selectedEntry.value?.group_name} · 计划 #${selectedEntry.value?.plan_id} · ${dialogMode.value === 'history' ? '测试历史' : '放大播放'}`)

watch([groupFilter, pageSize], () => { closeDialog(); if (page.value !== 1) page.value = 1; else void refresh({ resetView: true }) })
watch(page, () => void refresh({ resetView: true }))
watch(() => [authStore.user?.id, authStore.isAdmin], () => { listRequestVersion++; listAbort?.abort(); clearArtworkCache(); items.value = []; total.value = 0; groups.value = []; loading.value = false; closeDialog(); plans.value = []; adminGroups.value = []; showPlans.value = false; cleanupConfirmation.value = null; showSettings.value = false; settingsRequestVersion++; settingsLoaded.value = false; settingsLoading.value = false; settingsSaving.value = false; settingsError.value = ''; settingsDraft.value = { ...defaultSettings }; pelicanMetadata.reset?.(); void Promise.resolve(pelicanMetadata.load(true)).then(() => { if (authStore.isAdmin || pelicanMetadata.enabled) return refresh({ resetView: true }) }).catch(() => { if (authStore.isAdmin) return refresh({ resetView: true }) }) })
watch([() => pelicanMetadata.displayName, () => appStore.siteName], ([name, siteName]) => { document.title = `${name} - ${siteName || 'Sub2API'}` }, { immediate: true })
watch(refreshSeconds, seconds => {
  if (interval) clearInterval(interval)
  interval = seconds ? setInterval(() => void refresh(), seconds * 1000) : undefined
})
watch(() => pelicanMetadata.enabled, enabled => {
  if (!authStore.isAdmin && !enabled) { listRequestVersion++; listAbort?.abort(); clearArtworkCache(); items.value = []; total.value = 0; groups.value = []; closeDialog() }
})
onMounted(() => { document.addEventListener('keydown', trapFocus); void Promise.resolve(pelicanMetadata.load(true)).then(() => { if (authStore.isAdmin || pelicanMetadata.enabled) return refresh() }).catch(() => { if (authStore.isAdmin) return refresh() }) })
onBeforeUnmount(() => {
  disposed = true
  settingsRequestVersion++
  listRequestVersion++; historyRequestVersion++; dialogRequestVersion++
  listAbort?.abort(); dialogAbort?.abort()
  if (interval) clearInterval(interval)
  document.removeEventListener('keydown', trapFocus)
})

function formatNow(): string {
  return new Intl.DateTimeFormat('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(new Date())
}
function effortText(effort?: string | null): string {
  if (effort == null) return '强度未记录'
  const labels: Record<string, string> = { '': '默认', none: '无', minimal: '最低', low: '低', medium: '中', high: '高', xhigh: '极高' }
  return labels[effort] ?? effort
}
function statusText(status: PelicanEntry['status']): string { return status === 'success' ? '成功' : status === 'failed' ? '失败' : '本轮已跳过（未执行）' }
function statusClass(status: PelicanEntry['status']): string { return status === 'success' ? 'success' : status === 'failed' ? 'failed' : 'skipped' }
function formatDate(value?: string | null): string { const date = value ? new Date(value) : null; return date && !Number.isNaN(date.getTime()) ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'short', timeStyle: 'short' }).format(date) : '时间未知' }
function formatDuration(value: number): string { return value > 1000 ? `${(value / 1000).toFixed(1)} 秒` : `${value} 毫秒` }
function clearArtworkCache(): void { cardHtml.value = {}; cardArtworkIDs.value = {}; artworkLoading.value = {} }
async function refresh({ resetView = false }: { resetView?: boolean } = {}): Promise<void> {
  if (disposed) return
  const version = ++listRequestVersion
  const previousScrollTop = resetView ? 0 : resultsRegion.value?.scrollTop
  listAbort?.abort()
  const controller = new AbortController()
  listAbort = controller
  if (resetView) {
    clearArtworkCache(); items.value = []; total.value = 0; groups.value = []
    if (resultsRegion.value) resultsRegion.value.scrollTop = 0
  }
  loading.value = !items.value.length; error.value = ''
  try {
    const data = await listPelicanTests(buildPelicanListParams(groupFilter.value, '', page.value, pageSize.value), controller.signal)
    if (version !== listRequestVersion) return
    const lastPage = Math.max(1, Math.ceil(data.total / pageSize.value))
    if (page.value > lastPage) { total.value = data.total; page.value = lastPage; return }
    const currentArtworkIDs = new Map(data.items.map(entry => [entry.result_id, entry.artwork_result_id]))
    cardHtml.value = Object.fromEntries(Object.entries(cardHtml.value).filter(([id]) => cardArtworkIDs.value[Number(id)] === currentArtworkIDs.get(Number(id))))
    cardArtworkIDs.value = Object.fromEntries(Object.entries(cardArtworkIDs.value).filter(([id, artworkID]) => artworkID === currentArtworkIDs.get(Number(id))))
    artworkLoading.value = Object.fromEntries(Object.entries(artworkLoading.value).filter(([id]) => currentArtworkIDs.has(Number(id))))
    items.value = data.items; groups.value = data.groups; total.value = data.total; page.value = data.page; refreshedAt.value = formatNow()
    if (!resetView && previousScrollTop != null) void nextTick(() => { if (version === listRequestVersion && resultsRegion.value) resultsRegion.value.scrollTop = previousScrollTop })
    if (selectedEntry.value && !data.items.some(item => item.plan_id === selectedEntry.value?.plan_id && item.group_id === selectedEntry.value?.group_id)) closeDialog()
    void loadVisibleArtwork(data.items, version, controller.signal)
  } catch (reason) { if (version === listRequestVersion) { const status = (reason as { status?: number; response?: { status?: number } })?.status ?? (reason as { response?: { status?: number } })?.response?.status; if (status === 403) pelicanMetadata.reset?.(); error.value = messageOf(reason); clearArtworkCache(); items.value = []; total.value = 0; groups.value = []; closeDialog() } } finally { if (version === listRequestVersion) loading.value = false }
}
async function loadVisibleArtwork(entries: PelicanEntry[], version: number, signal: AbortSignal): Promise<void> {
  const queue = entries.filter(entry => entry.artwork_result_id && !cardHtml.value[entry.result_id])
  queue.forEach(entry => { artworkLoading.value[entry.result_id] = true })
  await Promise.all(Array.from({ length: Math.min(3, queue.length) }, async () => {
    while (queue.length && version === listRequestVersion) {
      const entry = queue.shift()!
      try { const result = await getPelicanResult(entry.artwork_result_id!, signal); if (version === listRequestVersion && result.html) { cardHtml.value[entry.result_id] = result.html; cardArtworkIDs.value[entry.result_id] = entry.artwork_result_id } } catch { /* authorization may change between list and detail */ } finally { if (version === listRequestVersion) artworkLoading.value[entry.result_id] = false }
    }
  }))
}
function clearFilters(): void { groupFilter.value = '' }
function replay(id: string): void { replayVersions.value[id] = (replayVersions.value[id] ?? 0) + 1 }
async function openHistory(entry: PelicanEntry, event: MouseEvent): Promise<void> {
  dialogRequestVersion++; dialogAbort?.abort(); dialogAbort = new AbortController()
  opener = event.currentTarget as HTMLElement
  selectedEntry.value = entry
  selectedRevision.value = null
  fromHistory.value = false
  dialogMode.value = 'history'
  const version = ++historyRequestVersion
  historyPage.value = 1; history.value = []; historyLoading.value = true; historyError.value = ''
  try { const result = await listPelicanHistory({ plan_id: entry.plan_id }, dialogAbort.signal); if (version === historyRequestVersion) history.value = result } catch (reason) { if (version === historyRequestVersion) historyError.value = messageOf(reason) } finally { if (version === historyRequestVersion) historyLoading.value = false }
}
async function openArtwork(entry: PelicanEntry, event: MouseEvent | KeyboardEvent): Promise<void> {
  const version = ++dialogRequestVersion
  historyRequestVersion++; dialogAbort?.abort(); dialogAbort = new AbortController()
  opener = event.currentTarget as HTMLElement; selectedEntry.value = entry; fromHistory.value = false
  if (!entry.artwork_result_id || !cardHtml.value[entry.result_id]) return
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
function messageOf(reason: unknown): string { const message = typeof (reason as { message?: unknown })?.message === 'string' ? (reason as { message: string }).message : ''; const known: Record<string, string> = { plan_paused: '计划因连续失败已自动暂停，请先恢复计划。', daily_call_limit_exceeded: '今日调用次数已达到上限（按 UTC+8 日期重置）。', daily_limit_exceeded: '今日调用次数已达到上限（按 UTC+8 日期重置）。', plan_running: '计划正在运行，暂时不能清理或恢复。', model_unsupported: '本轮模型配置检查未通过，已跳过；请检查账号的模型配置。' }; return (known[message] ?? message) || '请求失败，请稍后重试。' }
function planReasonText(message: string): string {
  if (message === 'group_scheduler_unavailable') return '分组当前没有可调度的账号，已跳过本轮测试'
  if (message === 'group_busy') return '分组当前繁忙，已跳过本轮测试'
  const reasons: Record<string, string> = { account_model_not_allowed: '该账号未配置此模型，已跳过', model_not_html_capable: '该账号映射后的模型不适合生成HTML，已跳过', account_scheduling_disabled: '账号已关闭调度，已跳过测试', account_inactive: '账号未启用', account_rate_limited: '账号当前受限流影响', account_cooldown: '账号仍在冷却中', account_outside_schedule: '账号当前不在可用时间窗内', account_quota_exceeded: '账号配额已耗尽', account_unavailable: '账号当前不可用', account_expired: '账号已过期', account_not_in_group: '账号不属于该分组', account_unsupported: '该账号类型暂不支持此测试', account_lookup_failed: '账号信息读取失败', group_unavailable: '分组当前不可用', model_unsupported: '本轮模型配置检查未通过，已跳过；请检查账号的模型配置', daily_call_limit_reached: '今日调用次数已达到上限（按 UTC+8 日期重置）', quota_reservation_failed: '调用额度预留失败，已安全跳过本次测试', plan_changed: '计划已变更，请重新运行', plan_cancelled: '计划已取消' }
  if (reasons[message]) return reasons[message]
  const fixedLegacy: Record<string, string> = { 'generation failed or timed out': '生成失败或已超时', 'generated content is too short': '生成内容未达到最小字符数', 'plan or membership changed': '计划或分组成员关系已变更，请重新运行', 'plan cancelled or time limit reached': '计划已取消或达到时间限制', 'pelican provider unavailable': '测试服务当前不可用', 'pelican account unavailable': '账号当前不可用（历史记录未提供具体原因）' }
  if (fixedLegacy[message]) return fixedLegacy[message]
  const legacy = message.toLowerCase()
  if (legacy.includes('account unavailable') || legacy.includes('no available account')) return '账号当前不可用（历史记录未提供具体原因）'
  if (legacy.includes('not schedulable')) return '账号当前不可用于定时测试（历史记录未提供具体原因）'
  if (legacy.includes('rate limit')) return '账号当前受限流影响（历史记录未提供具体原因）'
  if (legacy.includes('cooldown')) return '账号仍在冷却中（历史记录未提供具体原因）'
  if (legacy.includes('expired')) return '账号已过期（历史记录未提供具体原因）'
  return message
}
async function loadPlans(): Promise<void> { if (!authStore.isAdmin) return; plansLoading.value = true; plansError.value = ''; try { const [planData, groupData] = await Promise.all([listPelicanPlans(), getAllAdminGroups('openai')]); plans.value = planData; adminGroups.value = groupData.filter(group => group.status === 'active').map(group => ({ id: group.id, name: group.name })) } catch (reason) { plansError.value = messageOf(reason) } finally { plansLoading.value = false } }
watch(showPlans, visible => { if (visible) void loadPlans() })
watch(showSettings, visible => { if (visible) void loadSettings() })
async function loadSettings(): Promise<void> {
  if (!authStore.isAdmin) return
  const version = ++settingsRequestVersion
  settingsLoading.value = true; settingsError.value = ''
  try { const loaded = await getPelicanTestSettings(); if (version === settingsRequestVersion && authStore.isAdmin) { settingsDraft.value = { ...loaded }; settingsLoaded.value = true } } catch (reason) { if (version === settingsRequestVersion) { settingsLoaded.value = false; settingsError.value = messageOf(reason) } } finally { if (version === settingsRequestVersion) settingsLoading.value = false }
}
function restoreDefaultSettings(): void { settingsDraft.value = { ...defaultSettings }; settingsError.value = '' }
async function saveSettings(): Promise<void> {
	const version = ++settingsRequestVersion
  settingsSaving.value = true; settingsError.value = ''
	try {
		const saved = await updatePelicanTestSettings({ ...settingsDraft.value })
		if (version === settingsRequestVersion && authStore.isAdmin && !disposed) { settingsDraft.value = { ...saved }; if (pelicanMetadata.applySavedSettings) pelicanMetadata.applySavedSettings(saved.display_name, saved.enabled); else pelicanMetadata.applySavedName?.(saved.display_name) }
	} catch (reason) {
		if (version === settingsRequestVersion && !disposed) settingsError.value = messageOf(reason)
	} finally {
		if (version === settingsRequestVersion) settingsSaving.value = false
	}
}
async function runPlan(id: number): Promise<void> { if (!pelicanMetadata.enabled) { plansError.value = '鹈鹕广场已关闭，无法开始新的测试轮次。'; return } planBusyId.value = id; try { await runPelicanPlan(id); await loadPlans() } catch (reason) { plansError.value = messageOf(reason) } finally { planBusyId.value = null } }
async function resumePlan(id: number): Promise<void> { planBusyId.value = id; plansError.value = ''; try { await resumePelicanPlan(id); await loadPlans() } catch (reason) { plansError.value = messageOf(reason) } finally { planBusyId.value = null } }
async function savePlan(id: number | null, input: PelicanPlanInput): Promise<void> { planBusyId.value = id ?? -1; plansError.value = ''; try { if (id) await updatePelicanPlan(id, input); else await createPelicanPlan({ ...input, enabled: false }); saveVersion.value++; await loadPlans() } catch (reason) { plansError.value = messageOf(reason) } finally { planBusyId.value = null } }
async function removePlan(id: number): Promise<void> { planBusyId.value = id; try { await deletePelicanPlan(id); await loadPlans() } catch (reason) { plansError.value = messageOf(reason) } finally { planBusyId.value = null } }
function requestCleanup(plan: PelicanPlan, scope: 'failed' | 'expired'): void { if (planBusyId.value != null || isPlanRunning(plan)) return; cleanupOpener = document.activeElement instanceof HTMLElement ? document.activeElement : null; cleanupConfirmation.value = { plan, scope } }
function cancelCleanup(): void { if (cleanupBusy.value) return; cleanupConfirmation.value = null; const target = cleanupOpener; cleanupOpener = null; void nextTick(() => target?.focus()) }
function closePlans(): void { if (!cleanupConfirmation.value) showPlans.value = false }
async function confirmCleanup(): Promise<void> { const request = cleanupConfirmation.value; if (!request || cleanupBusy.value) return; cleanupBusy.value = true; planBusyId.value = request.plan.id; plansError.value = ''; cleanupNotice.value = ''; try { const { deleted_count } = await cleanupPelicanPlan(request.plan.id, request.scope); cleanupNotice.value = `已删除 ${deleted_count} 条${request.scope === 'failed' ? '失败或跳过' : '过期'}记录。`; cleanupConfirmation.value = null; cleanupOpener = null; await Promise.all([loadPlans(), refresh(), refreshOpenHistory()]) } catch (reason) { plansError.value = messageOf(reason) } finally { cleanupBusy.value = false; planBusyId.value = null } }
async function refreshOpenHistory(): Promise<void> { if (dialogMode.value !== 'history' || !selectedEntry.value) return; const entry = selectedEntry.value; const version = ++historyRequestVersion; historyLoading.value = true; historyError.value = ''; try { const data = await listPelicanHistory({ plan_id: entry.plan_id }); if (version === historyRequestVersion) history.value = data } catch (reason) { if (version === historyRequestVersion) historyError.value = messageOf(reason) } finally { if (version === historyRequestVersion) historyLoading.value = false } }
function isPlanRunning(plan: PelicanPlan): boolean { return Boolean(plan.running_until && new Date(plan.running_until).getTime() > Date.now()) }
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
.pelican-page { display: flex; min-width: 0; min-height: 0; height: 100%; flex-direction: column; font-size: 12px; }
.pelican-controls { flex: 0 0 auto; max-height: 50%; min-height: 0; overflow: auto; padding-right: 2px; }
.admin-toggle { margin: 0 0 12px; border: 1px solid var(--pelican-border); border-radius: 7px; background: var(--pelican-input); color: var(--pelican-text); padding: 7px 10px; cursor: pointer; }
.settings-editor { display: grid; gap: 9px; margin: 0 0 12px; padding: 12px; border: 1px solid var(--pelican-border); border-radius: 8px; background: var(--pelican-surface); }
.settings-editor p { margin: 0; color: var(--pelican-muted); }
.settings-editor label { display: grid; gap: 4px; font-weight: 600; }
.settings-editor input, .settings-editor textarea { width: 100%; box-sizing: border-box; border: 1px solid var(--pelican-border); border-radius: 6px; background: var(--pelican-input); color: var(--pelican-text); padding: 7px; font: inherit; }
.settings-editor .settings-enabled { display: flex; align-items: center; gap: 8px; }.settings-editor .settings-enabled input { width: auto; }
.settings-disabled-banner { margin: 0 0 12px; border: 1px solid color-mix(in srgb, var(--pelican-accent) 45%, var(--pelican-border)); border-radius: 8px; background: var(--pelican-accent-soft); color: var(--pelican-text); padding: 9px 12px; font-size: 12px; }
.settings-editor textarea { resize: vertical; white-space: pre-wrap; }
.settings-editor div { display: flex; gap: 8px; }
.cleanup-confirmation { --pelican-text: var(--console-text, #141413); --pelican-muted: var(--console-muted, #6c6a64); --pelican-border: var(--console-border, rgba(216, 206, 194, .68)); --pelican-input: var(--console-input, rgba(250, 249, 245, .94)); --pelican-accent: var(--console-accent, theme('colors.primary.500')); display: grid; gap: 10px; color: var(--pelican-text); }.cleanup-confirmation p { margin: 0; color: var(--pelican-muted); }.cleanup-confirmation div { display: flex; justify-content: flex-end; flex-wrap: wrap; gap: 8px; }.cleanup-confirmation button { border: 1px solid var(--pelican-border); border-radius: 7px; background: var(--pelican-input); color: var(--pelican-text); padding: 7px 10px; cursor: pointer; }.cleanup-confirmation button:last-child { background: var(--pelican-accent); color: #fff; border-color: var(--pelican-accent); }.cleanup-notice { margin: 0 0 8px; color: var(--pelican-accent); }
.dark .cleanup-confirmation { --pelican-text: theme('colors.dark.50'); --pelican-muted: theme('colors.dark.400'); --pelican-border: rgba(51, 65, 85, .82); --pelican-input: rgba(10, 18, 32, .86); --pelican-accent: theme('colors.indigo.400'); }
.settings-editor button { border: 1px solid var(--pelican-border); border-radius: 6px; background: var(--pelican-input); color: var(--pelican-text); padding: 7px 10px; cursor: pointer; }
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
.refresh-error { margin: -8px 0 10px; color: #dc2626; font-size: 11px; }
.results-region { min-height: 0; flex: 1 1 auto; overflow: auto; padding: 0 2px 10px 0; }
.gallery { display: grid; grid-template-columns: repeat(auto-fill, minmax(255px, 1fr)); gap: 12px; }
.pelican-card { min-width: 0; border: 1px solid var(--pelican-border); background: var(--pelican-surface); box-shadow: var(--console-shadow, none); border-radius: 12px; padding: 9px; }
.artwork-placeholder { display: grid; place-items: center; min-height: 180px; border-radius: 8px; background: var(--pelican-accent-soft); color: var(--pelican-muted); font-size: 11px; }
.artwork-preview { position: relative; display: block; width: 100%; overflow: hidden; border: 0; border-radius: 8px; padding: 0; cursor: zoom-in; background: var(--pelican-accent-soft); }
.artwork-preview :deep(.pelican-frame) { pointer-events: none; }
.preview-overlay { position: absolute; inset: 0; z-index: 1; }
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
.skipped { color: var(--pelican-muted); }
.html-meta { color: var(--pelican-muted); font-size: 10px; line-height: 1.6; margin: 7px 0 11px; overflow-wrap: anywhere; }
.group-meta { margin: 0 0 4px; color: var(--pelican-muted); font-size: 10px; }
.stars { color: var(--pelican-accent); letter-spacing: 1px; white-space: nowrap; }
.last-success { color: var(--pelican-muted); font-size: 10px; margin: 0 0 2px; }
.pagination { display: flex; flex: 0 0 auto; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px; margin-top: 12px; color: var(--pelican-muted); }
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
.history-unavailable { min-height: 120px; padding: 20px; border-radius: 10px; background: var(--pelican-input); color: var(--pelican-muted); font-size: 13px; }
.history-unavailable p { margin-top: 8px; }
.history-pagination { position: sticky; bottom: 0; display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; padding: 16px 0 4px; background: var(--pelican-input); color: var(--pelican-text); font-size: 12px; }
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
