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
          <PelicanReferenceSettings v-model="radarDraft" :catalog="referenceCatalog" :loading="referenceCatalogLoading" :disabled="!settingsLoaded || settingsLoading || settingsSaving" @validity-change="referenceSettingsValid = $event" />
          <p v-if="settingsError" role="alert">{{ settingsError }}</p>
          <div><button v-if="!settingsLoaded" type="button" :disabled="settingsLoading" @click="loadSettings">重新加载</button><button type="button" :disabled="!settingsLoaded || settingsLoading || settingsSaving" @click="restoreDefaultSettings">恢复默认</button><button type="button" :disabled="!settingsLoaded || settingsLoading || settingsSaving || !referenceSettingsValid" @click="saveSettings">{{ settingsSaving ? '保存中…' : '保存设置' }}</button></div>
        </section>
        <button class="admin-toggle" type="button" @click="showPlans = true">管理测试计划</button>
        <BaseDialog :show="showPlans" :title="`${pelicanMetadata.displayName} · 测试计划`" width="wide" :close-on-escape="!cleanupConfirmation && !planEditorOpen && planBusyId == null" @close="closePlans"><p v-if="cleanupNotice" class="cleanup-notice" role="status">{{ cleanupNotice }}</p><PelicanPlansPanel v-if="showPlans" :plans="plans" :groups="adminGroups" :loading="plansLoading" :error="plansError" :save-error="planSaveError" :busy-id="planBusyId" :save-version="saveVersion" :pelican-enabled="pelicanMetadata.enabled" @editor-change="onPlanEditorChange" @run="runPlan" @resume="resumePlan" @save="savePlan" @delete="removePlan" @cleanup-request="requestCleanup" /></BaseDialog>
        <BaseDialog :show="Boolean(cleanupConfirmation)" title="确认清理测试记录" width="narrow" :z-index="60" :close-on-escape="!cleanupBusy" @close="cancelCleanup"><div v-if="cleanupConfirmation" class="cleanup-confirmation"><p>{{ cleanupConfirmation.scope === 'failed' ? '将永久删除该计划的失败和已跳过记录，成功记录不会删除。' : '将按当前保留天数和结果数量永久删除过期记录，并始终保留最新成功记录。' }}</p><p>此操作不可撤销。删除后会刷新画廊和已打开的作品预览。</p><div><button type="button" :disabled="cleanupBusy" @click="cancelCleanup">取消</button><button type="button" :disabled="cleanupBusy" @click="confirmCleanup">{{ cleanupBusy ? '正在清理…' : '确认清理' }}</button></div></div></BaseDialog>
      </template>
      <PelicanReferenceScores v-if="authStore.isAdmin || pelicanMetadata.enabled" :data="referenceScores" :loading="referenceLoading" :failed="referenceFailed" />
      <section class="toolbar" :aria-label="`筛选${pelicanMetadata.displayName}`">
        <div class="browse-left">
          <label class="search-box"><Icon name="search" size="sm" /><input v-model.trim="searchQuery" type="search" placeholder="搜索分组或模型" aria-label="搜索分组或模型" /></label>
          <label class="select-field"><select v-model.number="groupFilter" aria-label="筛选分组">
            <option value="">全部分组</option>
            <option v-for="item in groups" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select><Icon name="chevronDown" size="sm" /></label>
          <label class="select-field"><select v-model="statusFilter" aria-label="筛选状态">
            <option value="all">全部状态</option>
            <option value="success">生成成功</option>
            <option value="skipped">本轮已跳过</option>
            <option value="running">生成中</option>
            <option value="failed">生成失败</option>
          </select><Icon name="chevronDown" size="sm" /></label>
        </div>
        <div class="refresh-controls">
          <span class="updated-at" role="status">{{ refreshedAt }} 更新</span>
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
      <div class="list-summary"><span><strong>{{ visibleGroupCount }}</strong> 个分组 <i>/</i> <strong>{{ visibleResultCount }}</strong> 条测试结果</span><span class="scroll-tip"><span aria-hidden="true">↔</span>每组独立横向浏览</span></div>
      <p v-if="error && items.length" class="refresh-error" role="alert">{{ error }}</p>
      </div>

      <div ref="resultsRegion" class="results-region">
      <div v-if="loading && !items.length" class="empty-state" aria-live="polite">正在加载测试结果…</div>
      <div v-else-if="error && !items.length" class="empty-state" role="alert"><h2>无法加载测试结果</h2><p>{{ error }}</p><button type="button" @click="refresh()">重试</button></div>
      <div v-else-if="visibleGalleryGroups.length" class="gallery">
        <section v-for="group in visibleGalleryGroups" :key="group.id" class="pelican-card group-row" :data-group="group.id" :aria-label="`${group.name}${group.platform ? ` · ${group.platform}` : ''}`">
          <aside class="group-heading group-meta">
            <div class="platform" :style="{ '--brand-color': platformColor(group.platform) }">{{ group.platform || 'MODEL GROUP' }}</div>
            <h2 class="group-title">{{ group.name }}<span>{{ groupSubtitle(group) }}</span></h2>
            <p class="group-count"><strong>{{ group.entries.length }}</strong> 条测试记录</p>
            <nav class="model-filters" :aria-label="`${group.name}模型筛选`">
              <button type="button" class="model-filter" :class="{ active: selectedModels[group.id] === 'all' }" :aria-pressed="selectedModels[group.id] === 'all'" @click="selectGroupModel(group.id, 'all')">全部模型<Icon name="check" size="xs" /></button>
              <button v-for="model in group.models" :key="model" type="button" class="model-filter" :class="{ active: selectedModels[group.id] === model }" :aria-pressed="selectedModels[group.id] === model" @click="selectGroupModel(group.id, model)">{{ model }}<Icon name="check" size="xs" /></button>
            </nav>
            <div class="row-navigation">
              <button class="icon-button" type="button" :disabled="!railNavigation[group.id]?.canScrollLeft" :aria-label="`向前浏览${group.name}`" @click="scrollGroupRail($event, -1)"><Icon name="chevronLeft" size="sm" /></button>
              <button class="icon-button" type="button" :disabled="!railNavigation[group.id]?.canScrollRight" :aria-label="`向后浏览${group.name}`" @click="scrollGroupRail($event, 1)"><Icon name="chevronRight" size="sm" /></button>
              <span class="position-label">{{ railNavigation[group.id]?.position || '' }}</span>
            </div>
          </aside>
          <div class="row-content">
            <div class="rail" tabindex="0" role="region" :aria-label="`${group.name}作品列表`" @scroll="updateRailNavigation(group.id, $event)">
              <template v-if="group.entries.length">
                <article v-for="item in group.entries" :key="`${item.plan_id}-${item.group_id}-${item.model_id}`" class="model-artwork result-card" :aria-label="`${item.model_id} · 计划 ${item.plan_id}`">
                  <div class="artwork-shell">
                    <PelicanLazyArtwork v-if="item.artwork_result_id" :result-id="item.artwork_result_id" :root="resultsRegion" :min-height="154" :replay-key="replayVersions[`card-${item.result_id}`] ?? 0" @enlarge="openArtwork(item, $event)" />
                    <div v-else class="artwork-placeholder art-placeholder" :class="item.status">
                      <span v-if="item.status === 'running'" class="running-orbit" aria-hidden="true"></span>
                      <Icon v-else :name="item.status === 'failed' ? 'exclamationCircle' : 'clock'" size="md" />
                      <strong>{{ item.status === 'running' ? '正在生成本轮作品' : item.status === 'failed' ? '本轮未完成' : '本轮暂无作品' }}</strong>
                      <small>{{ item.status === 'running' ? '完成后会自动出现在这里' : item.status === 'failed' ? '生成失败 · 可查看其他轮次' : '等待作品进入可视区域' }}</small>
                    </div>
                    <div class="card-actions">
                      <button :disabled="!item.artwork_result_id" type="button" aria-label="放大播放" title="放大播放" @click="openArtwork(item, $event)"><Icon name="eye" size="sm" /></button>
                      <button :disabled="!item.artwork_result_id" type="button" aria-label="重新播放" title="重新播放" @click="replay(`card-${item.result_id}`)"><Icon name="refresh" size="sm" /></button>
                    </div>
                  </div>
                  <div class="card-copy">
                    <h3 class="model-title model-heading"><ModelIcon :model="item.model_id" size="17px" /><strong class="model-name" :title="item.model_id">{{ item.model_id }}</strong><span class="status" :class="statusClass(item.status)"><i></i>{{ previewStatusText(item.status) }}</span></h3>
                    <span v-if="item.round_number || item.history_count" class="round-label">第 {{ item.round_number || item.history_count }} 轮</span>
                    <p v-if="item.artwork_result_id && item.artwork_result_id !== item.result_id" class="last-success">上次成功作品</p>
                    <div class="card-footer">
                      <time class="generation-time" :title="formatTimestamp(item.artwork_finished_at ?? (item.status === 'success' ? item.finished_at : null))">{{ formatCompactDate(item.artwork_finished_at ?? (item.status === 'success' ? item.finished_at : null)) }}</time>
                      <span class="duration"><Icon name="clock" size="xs" />{{ formatCompactDuration(item.latency_ms) }}</span>
                    </div>
                  </div>
                </article>
              </template>
              <div v-else class="rail-empty"><span>该模型暂无符合条件的结果</span><button class="text-button" type="button" @click="selectGroupModel(group.id, 'all')">查看本组全部结果</button></div>
            </div>
          </div>
        </section>
      </div>
      <div v-else class="empty-state"><Icon name="search" size="lg" /><h2>暂无测试结果</h2><p>当前有权限的分组还没有可展示的测试结果。</p><button v-if="hasFilters" type="button" @click="clearFilters">清除筛选</button></div>
      </div>

      <footer class="pagination">
        <span>显示 {{ total ? (page - 1) * pageSize + 1 : 0 }} 至 {{ Math.min(page * pageSize, total) }}，共 {{ total }} 个分组</span>
        <div class="page-controls">
          <label>每页 <select v-model.number="pageSize" aria-label="每页条数"><option :value="6">6</option><option :value="12">12</option><option :value="24">24</option></select></label>
          <button type="button" :disabled="page === 1" @click="page--">上一页</button>
          <span>{{ page }} / {{ totalPages }}</span>
          <button type="button" :disabled="page >= totalPages" @click="page++">下一页</button>
        </div>
      </footer>
      <BaseDialog v-if="dialogMode && selectedEntry" :show="true" :title="dialogTitle" width="wide" :close-on-click-outside="true" @close="closeDialog">
        <div ref="dialogContent" class="dialog-content">
          <template v-if="selectedRevision">
            <div class="enlarged-art"><PelicanLazyArtwork interactive :result-id="selectedRevision.id" :replay-key="replayVersions[`large-${selectedRevision.id}`] ?? 0" /></div>
            <div class="artwork-actions"><span class="success">● 成功作品 · {{ formatDuration(selectedRevision.latency_ms) }}</span><button type="button" @click="replay(`large-${selectedRevision.id}`)">重新播放</button></div>
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
import ModelIcon from '@/components/common/ModelIcon.vue'
import PelicanLazyArtwork from '@/components/pelican/PelicanLazyArtwork.vue'
import { createPelicanArtwork, providePelicanArtwork } from '@/components/pelican/usePelicanArtwork'
import PelicanPlansPanel from '@/components/pelican/PelicanPlansPanel.vue'
import PelicanReferenceScores from '@/components/pelican/PelicanReferenceScores.vue'
import PelicanReferenceSettings from '@/components/pelican/PelicanReferenceSettings.vue'
import { buildPelicanListParams, cleanupPelicanPlan, createPelicanPlan, deletePelicanPlan, getPelicanTestSettings, getPelicanReferenceScores, getPelicanReferenceCatalog, listPelicanPlans, listPelicanTests, resumePelicanPlan, runPelicanPlan, updatePelicanPlan, updatePelicanTestSettings, type PelicanEntry, type PelicanPlan, type PelicanPlanInput, type PelicanResult, type PelicanTestSettings, type PelicanReferenceScores as ReferenceScores, type PelicanReferenceCatalog, type PelicanReferenceSettings as ReferenceSettings } from '@/api/pelicanTests'
import { getAll as getAllAdminGroups } from '@/api/admin/groups'
import { useAuthStore } from '@/stores/auth'
import { usePelicanMetadataStore } from '@/stores/pelicanMetadata'
import { useAppStore } from '@/stores/app'

const groupFilter = ref<number | ''>('')
const searchQuery = ref('')
const statusFilter = ref<'all' | PelicanEntry['status']>('all')
const selectedModels = ref<Record<number, string>>({})
const railNavigation = ref<Record<number, { canScrollLeft: boolean; canScrollRight: boolean; position: string }>>({})
const page = ref(1)
const pageSize = ref(24)
const refreshSeconds = ref(0)
const refreshedAt = ref(formatNow())
const replayVersions = ref<Record<string, number>>({})
const dialogMode = ref<'artwork' | null>(null)
const selectedEntry = ref<PelicanEntry | null>(null)
const selectedRevision = ref<PelicanResult | null>(null)
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
const artwork = createPelicanArtwork(() => {
  listRequestVersion++; dialogRequestVersion++
  listAbort?.abort(); dialogAbort?.abort()
  items.value = []; total.value = 0; groups.value = []; closeDialog()
  error.value = '作品访问权限已变更，请刷新列表。'
})
providePelicanArtwork(artwork)
const showPlans = ref(false)
const plans = ref<PelicanPlan[]>([])
const plansLoading = ref(false)
const plansError = ref('')
const planSaveError = ref('')
const planEditorOpen = ref(false)
const planBusyId = ref<number | null>(null)
let plansSession = 0
let plansRequestVersion = 0
let planEditorVersion = 0
let planActionVersion = 0
const cleanupConfirmation = ref<{ plan: PelicanPlan; scope: 'failed' | 'expired' } | null>(null)
const cleanupBusy = ref(false)
const cleanupNotice = ref('')
const adminGroups = ref<Array<{ id: number; name: string; platform?: string }>>([])
let opener: HTMLElement | null = null
let cleanupOpener: HTMLElement | null = null
let interval: ReturnType<typeof setInterval> | undefined
let listRequestVersion = 0
let dialogRequestVersion = 0
let listAbort: AbortController | undefined
let dialogAbort: AbortController | undefined
let refreshing = false
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
const radarDraft = computed<ReferenceSettings>({
  get: () => settingsDraft.value.radar ?? { enabled: false, selections: [] },
  set: radar => { settingsDraft.value = { ...settingsDraft.value, radar } }
})
const referenceSettingsValid = ref(true)
const referenceScores = ref<ReferenceScores | null>(null)
const referenceLoading = ref(true)
const referenceFailed = ref(false)
const referenceCatalog = ref<PelicanReferenceCatalog | null>(null)
const referenceCatalogLoading = ref(false)
let referenceRequestVersion = 0
let referenceCatalogVersion = 0
let referenceAbort: AbortController | undefined
let referenceCatalogAbort: AbortController | undefined

function copySettings(settings: PelicanTestSettings): PelicanTestSettings {
  return { ...settings, ...(settings.radar ? { radar: { enabled: settings.radar.enabled, selections: settings.radar.selections.map(item => ({ ...item })) } } : {}) }
}
function resetReferenceSession(): void {
  referenceRequestVersion++; referenceCatalogVersion++
  referenceAbort?.abort(); referenceCatalogAbort?.abort()
  referenceScores.value = null; referenceCatalog.value = null
  referenceLoading.value = false; referenceFailed.value = false; referenceCatalogLoading.value = false; referenceSettingsValid.value = true
}
async function loadReferenceScores(): Promise<void> {
  if (disposed || (!authStore.isAdmin && !pelicanMetadata.enabled)) return
  const version = ++referenceRequestVersion
  referenceAbort?.abort()
  const controller = new AbortController()
  referenceAbort = controller
  referenceLoading.value = true
  referenceFailed.value = false
  try {
    const data = await getPelicanReferenceScores(controller.signal)
    if (!disposed && version === referenceRequestVersion) referenceScores.value = data
  } catch {
    if (!disposed && version === referenceRequestVersion) referenceFailed.value = true
  } finally { if (!disposed && version === referenceRequestVersion) referenceLoading.value = false }
}
async function loadReferenceCatalog(): Promise<void> {
  if (disposed || !authStore.isAdmin) return
  const version = ++referenceCatalogVersion
  referenceCatalogAbort?.abort()
  const controller = new AbortController()
  referenceCatalogAbort = controller
  referenceCatalogLoading.value = true
  try {
    const data = await getPelicanReferenceCatalog(controller.signal)
    if (!disposed && authStore.isAdmin && version === referenceCatalogVersion) referenceCatalog.value = data
  } catch {
    if (!disposed && version === referenceCatalogVersion) referenceCatalog.value = null
  } finally { if (!disposed && version === referenceCatalogVersion) referenceCatalogLoading.value = false }
}

const hasFilters = computed(() => groupFilter.value !== '' || searchQuery.value !== '' || statusFilter.value !== 'all' || Object.values(selectedModels.value).some(model => model !== 'all'))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const dialogTitle = computed(() => `${selectedEntry.value?.group_name} · ${selectedEntry.value?.model_id} · 计划 #${selectedEntry.value?.plan_id} · 放大播放`)
const galleryGroups = computed(() => {
  const grouped = new Map<number, { id: number; name: string; platform?: string; entries: PelicanEntry[]; models: string[] }>()
  for (const entry of items.value) {
    let group = grouped.get(entry.group_id)
    if (!group) { group = { id: entry.group_id, name: entry.group_name, platform: entry.platform, entries: [], models: [] }; grouped.set(entry.group_id, group) }
    group.entries.push(entry)
    if (!group.models.includes(entry.model_id)) group.models.push(entry.model_id)
  }
  return [...grouped.values()]
})
const visibleGalleryGroups = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return galleryGroups.value
    .map(group => {
      const groupMatches = !query || group.name.toLowerCase().includes(query) || group.models.some(model => model.toLowerCase().includes(query))
      const selectedModel = selectedModels.value[group.id] || 'all'
      const entries = group.entries.filter(entry => {
        const matchesModel = selectedModel === 'all' || entry.model_id === selectedModel
        const matchesStatus = statusFilter.value === 'all' || entry.status === statusFilter.value
        const matchesQuery = !query || group.name.toLowerCase().includes(query) || entry.model_id.toLowerCase().includes(query)
        return matchesModel && matchesStatus && matchesQuery
      })
      return groupMatches ? { ...group, entries } : null
    })
    .filter((group): group is NonNullable<typeof group> => group !== null)
})
const visibleGroupCount = computed(() => visibleGalleryGroups.value.length)
const visibleResultCount = computed(() => visibleGalleryGroups.value.reduce((count, group) => count + group.entries.length, 0))

watch([groupFilter, pageSize], () => { closeDialog(); if (page.value !== 1) page.value = 1; else void refresh({ resetView: true }) })
watch([searchQuery, statusFilter], () => { closeDialog(); if (resultsRegion.value) resultsRegion.value.scrollTop = 0; void nextTick(updateAllRailNavigation) })
watch(page, () => void refresh({ resetView: true }))
watch([() => authStore.user?.id, () => authStore.isAdmin], () => { invalidatePlansSession(); resetReferenceSession(); listRequestVersion++; listAbort?.abort(); clearArtworkCache(); items.value = []; total.value = 0; groups.value = []; selectedModels.value = {}; railNavigation.value = {}; loading.value = false; closeDialog(); plans.value = []; adminGroups.value = []; showPlans.value = false; cleanupConfirmation.value = null; showSettings.value = false; settingsRequestVersion++; settingsLoaded.value = false; settingsLoading.value = false; settingsSaving.value = false; settingsError.value = ''; settingsDraft.value = { ...defaultSettings }; pelicanMetadata.reset?.(); void Promise.resolve(pelicanMetadata.load(true)).then(() => { if (authStore.isAdmin || pelicanMetadata.enabled) { void loadReferenceScores(); return refresh({ resetView: true }) } }).catch(() => { if (authStore.isAdmin) { void loadReferenceScores(); return refresh({ resetView: true }) } }) }, { flush: 'sync' })
watch([() => pelicanMetadata.displayName, () => appStore.siteName], ([name, siteName]) => { document.title = `${name} - ${siteName || 'Sub2API'}` }, { immediate: true })
function canPoll(): boolean { return !document.hidden && !showPlans.value && !showSettings.value && !dialogMode.value && !refreshing && (authStore.isAdmin || pelicanMetadata.enabled) }
function configurePolling(): void {
  if (interval) clearInterval(interval)
  interval = refreshSeconds.value && !document.hidden ? setInterval(() => { if (canPoll()) void refresh() }, refreshSeconds.value * 1000) : undefined
}
function visibilityChanged(): void {
  artwork.setPaused(document.hidden)
  configurePolling()
  if (!document.hidden && refreshSeconds.value && canPoll()) void refresh()
}
watch(refreshSeconds, configurePolling)
watch(() => pelicanMetadata.enabled, enabled => {
  if (!authStore.isAdmin && !enabled) { resetReferenceSession(); listRequestVersion++; listAbort?.abort(); clearArtworkCache(); items.value = []; total.value = 0; groups.value = []; closeDialog() }
})
onMounted(() => {
  document.addEventListener('keydown', trapFocus)
  document.addEventListener('keydown', focusSearch)
  document.addEventListener('visibilitychange', visibilityChanged)
  artwork.setPaused(document.hidden)
  void Promise.resolve(pelicanMetadata.load()).then(() => { if (authStore.isAdmin || pelicanMetadata.enabled) { void loadReferenceScores(); return refresh() } }).catch(() => { if (authStore.isAdmin) { void loadReferenceScores(); return refresh() } })
})
onBeforeUnmount(() => {
  disposed = true
  artwork.dispose()
  resetReferenceSession()
  invalidatePlansSession()
  settingsRequestVersion++
  listRequestVersion++; dialogRequestVersion++
  listAbort?.abort(); dialogAbort?.abort()
  if (interval) clearInterval(interval)
  document.removeEventListener('keydown', trapFocus)
  document.removeEventListener('keydown', focusSearch)
  document.removeEventListener('visibilitychange', visibilityChanged)
})

function formatNow(): string {
  return new Intl.DateTimeFormat('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(new Date())
}
function statusClass(status: PelicanEntry['status']): string { return status === 'success' ? 'success' : status === 'failed' ? 'failed' : status === 'running' ? 'running' : 'skipped' }
function formatTimestamp(value?: string | null): string { const date = value ? new Date(value) : null; return date && !Number.isNaN(date.getTime()) ? `${date.toISOString()} (UTC)` : '时间未知' }
function formatDuration(value: number): string { return value > 1000 ? `${(value / 1000).toFixed(1)} 秒` : `${value} 毫秒` }
function formatCompactDate(value?: string | null): string {
  const date = value ? new Date(value) : null
  return date && !Number.isNaN(date.getTime()) ? new Intl.DateTimeFormat('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false }).format(date).replace(/\//g, ' 月 ').replace('月 ', '月 ') : '时间未知'
}
function formatCompactDuration(value?: number | null): string {
  if (value == null || !Number.isFinite(value)) return '—'
  return value > 1000 ? `${(value / 1000).toFixed(2)} 秒` : `${value} 毫秒`
}
function previewStatusText(status: PelicanEntry['status']): string {
  return status === 'success' ? '生成成功' : status === 'failed' ? '生成失败' : status === 'running' ? '生成中' : '本轮已跳过'
}
function groupSubtitle(group: { platform?: string }): string {
  return group.platform ? `${group.platform} · 多模型对照` : '多模型并行对照'
}
function platformColor(platform?: string): string {
  const colors: Record<string, string> = { openai: '#557c70', anthropic: '#c28d73', google: '#668bbd', gemini: '#668bbd', deepseek: '#7d8fbc', xai: '#6b7280' }
  return colors[(platform || '').toLowerCase()] || 'var(--pelican-accent)'
}
function selectGroupModel(groupID: number, model: string): void {
  selectedModels.value = { ...selectedModels.value, [groupID]: model }
  closeDialog()
  void nextTick(() => updateAllRailNavigation())
}
function updateRailNavigation(groupID: number, event: Event | HTMLElement): void {
  const rail = event instanceof HTMLElement ? event : event.currentTarget instanceof HTMLElement ? event.currentTarget : null
  if (!rail) return
  const cards = rail.querySelectorAll<HTMLElement>('.result-card')
  const max = Math.max(0, rail.scrollWidth - rail.clientWidth)
  const step = cards[0] ? cards[0].offsetWidth + 14 : 1
  const first = cards.length ? Math.min(cards.length, Math.round(rail.scrollLeft / step) + 1) : 0
  railNavigation.value = {
    ...railNavigation.value,
    [groupID]: { canScrollLeft: rail.scrollLeft > 2, canScrollRight: max > 2 && rail.scrollLeft < max - 2, position: cards.length ? `${first} / ${cards.length}` : '' }
  }
}
function updateAllRailNavigation(): void {
  resultsRegion.value?.querySelectorAll<HTMLElement>('.group-row').forEach(row => {
    const rail = row.querySelector<HTMLElement>('.rail')
    const groupID = Number(row.getAttribute('data-group'))
    if (rail && Number.isFinite(groupID)) updateRailNavigation(groupID, rail)
  })
}
function scrollGroupRail(event: MouseEvent, direction: -1 | 1): void {
  const rail = (event.currentTarget as HTMLElement).closest('.group-row')?.querySelector<HTMLElement>('.rail')
  if (!rail) return
  const left = direction * Math.max(rail.clientWidth * .85, 220)
  if (typeof rail.scrollBy === 'function') rail.scrollBy({ left, behavior: window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' })
  else rail.scrollLeft += left
  window.setTimeout(() => updateRailNavigation(Number((event.currentTarget as HTMLElement).closest('.group-row')?.getAttribute('data-group')), rail), 180)
}
function clearArtworkCache(): void { artwork.invalidate() }
async function refresh({ resetView = false }: { resetView?: boolean } = {}): Promise<void> {
  if (disposed) return
  const version = ++listRequestVersion
  refreshing = true
  const previousScrollTop = resetView ? 0 : resultsRegion.value?.scrollTop
  listAbort?.abort()
  const controller = new AbortController()
  listAbort = controller
  if (resetView) {
    artwork.allowed.value = false; artwork.cancelPending(); items.value = []; total.value = 0; groups.value = []
    if (resultsRegion.value) resultsRegion.value.scrollTop = 0
  }
  loading.value = !items.value.length; error.value = ''
  try {
    const data = await listPelicanTests(buildPelicanListParams(groupFilter.value, '', page.value, pageSize.value), controller.signal)
    if (version !== listRequestVersion) return
    const lastPage = Math.max(1, Math.ceil(data.total / pageSize.value))
    if (page.value > lastPage) { total.value = data.total; page.value = lastPage; return }
    const incomingIDs = new Set(data.items.map(item => item.artwork_result_id))
    if (items.value.some(item => item.artwork_result_id && !incomingIDs.has(item.artwork_result_id))) clearArtworkCache()
    artwork.resume()
    items.value = data.items; groups.value = data.groups; total.value = data.total; page.value = data.page; refreshedAt.value = formatNow()
    void nextTick(updateAllRailNavigation)
    if (!resetView && previousScrollTop != null) void nextTick(() => { if (version === listRequestVersion && resultsRegion.value) resultsRegion.value.scrollTop = previousScrollTop })
    if (selectedEntry.value && !data.items.some(item => item.plan_id === selectedEntry.value?.plan_id && item.group_id === selectedEntry.value?.group_id && item.model_id === selectedEntry.value?.model_id)) closeDialog()
  } catch (reason) { if (version === listRequestVersion) { const status = (reason as { status?: number; response?: { status?: number } })?.status ?? (reason as { response?: { status?: number } })?.response?.status; if (status === 403) pelicanMetadata.reset?.(); error.value = messageOf(reason); if (status === 401 || status === 403) { clearArtworkCache(); items.value = []; total.value = 0; groups.value = []; closeDialog() } } } finally { if (version === listRequestVersion) { loading.value = false; refreshing = false } }
}
function clearFilters(): void { groupFilter.value = ''; searchQuery.value = ''; statusFilter.value = 'all'; selectedModels.value = {} }
function replay(id: string): void { replayVersions.value[id] = (replayVersions.value[id] ?? 0) + 1 }
async function openArtwork(entry: PelicanEntry, event: MouseEvent | KeyboardEvent): Promise<void> {
  const version = ++dialogRequestVersion
  dialogAbort?.abort(); dialogAbort = new AbortController()
  opener = event.currentTarget as HTMLElement; selectedEntry.value = entry
  if (!entry.artwork_result_id) return
  try { const result = await artwork.request(entry.artwork_result_id, dialogAbort.signal, true); if (version !== dialogRequestVersion) return; selectedRevision.value = { ...result, html: undefined }; dialogMode.value = 'artwork' } catch (reason) { if (version === dialogRequestVersion) error.value = messageOf(reason) }
  void nextTick(() => dialogContent.value?.querySelector<HTMLButtonElement>('button')?.focus())
}
function messageOf(reason: unknown): string { const message = typeof (reason as { message?: unknown })?.message === 'string' ? (reason as { message: string }).message : ''; const known: Record<string, string> = { plan_paused: '计划因连续失败已自动暂停，请先恢复计划。', daily_call_limit_exceeded: '今日调用次数已达到上限（按 UTC+8 日期重置）。', daily_limit_exceeded: '今日调用次数已达到上限（按 UTC+8 日期重置）。', plan_running: '计划正在运行，暂时不能清理或恢复。', model_unsupported: '本轮模型配置检查未通过，已跳过；请检查账号的模型配置。' }; return (known[message] ?? message) || '请求失败，请稍后重试。' }
function invalidatePlansSession(): void {
  plansSession++; plansRequestVersion++; planActionVersion++; planEditorVersion++
  planEditorOpen.value = false; planBusyId.value = null; plansLoading.value = false
  cleanupBusy.value = false; cleanupNotice.value = ''
  planSaveError.value = ''; plansError.value = ''
}
function isPlansSession(session: number): boolean { return !disposed && authStore.isAdmin && showPlans.value && session === plansSession }
function onPlanEditorChange(open: boolean): void {
  planEditorVersion++; planEditorOpen.value = open; planSaveError.value = ''
}
async function loadPlans(session = plansSession, afterSave = false): Promise<void> {
  if (!isPlansSession(session)) return
  const version = ++plansRequestVersion
  const isCurrent = () => isPlansSession(session) && version === plansRequestVersion
  plansLoading.value = true; plansError.value = ''
  try {
    const [planData, groupData] = await Promise.all([listPelicanPlans(), getAllAdminGroups()])
    if (!isCurrent()) return
    plans.value = planData
    adminGroups.value = groupData.filter(group => group.status === 'active').map(group => ({ id: group.id, name: group.name, platform: group.platform }))
  } catch (reason) {
    if (isCurrent()) plansError.value = `${afterSave ? '计划已保存，但列表刷新失败：' : ''}${messageOf(reason)}`
  } finally { if (isCurrent()) plansLoading.value = false }
}
watch(showPlans, visible => { invalidatePlansSession(); if (visible) void loadPlans() }, { flush: 'sync' })
watch(showSettings, visible => { if (visible) void loadSettings() })
async function loadSettings(): Promise<void> {
  if (!authStore.isAdmin) return
  const version = ++settingsRequestVersion
  settingsLoading.value = true; settingsError.value = ''
  void loadReferenceCatalog()
  try { const loaded = await getPelicanTestSettings(); if (!disposed && version === settingsRequestVersion && authStore.isAdmin) { settingsDraft.value = copySettings(loaded); settingsLoaded.value = true } } catch (reason) { if (version === settingsRequestVersion) { settingsLoaded.value = false; settingsError.value = messageOf(reason) } } finally { if (version === settingsRequestVersion) settingsLoading.value = false }
}
function restoreDefaultSettings(): void { settingsDraft.value = { ...defaultSettings, ...(settingsDraft.value.radar ? { radar: copySettings(settingsDraft.value).radar } : {}) }; settingsError.value = '' }
async function saveSettings(): Promise<void> {
  if (!settingsLoaded.value || settingsLoading.value || settingsSaving.value || !referenceSettingsValid.value || !authStore.isAdmin) return
	const version = ++settingsRequestVersion
  settingsSaving.value = true; settingsError.value = ''
	try {
		const saved = await updatePelicanTestSettings(copySettings(settingsDraft.value))
		if (version === settingsRequestVersion && authStore.isAdmin && !disposed) { settingsDraft.value = copySettings(saved); if (pelicanMetadata.applySavedSettings) pelicanMetadata.applySavedSettings(saved.display_name, saved.enabled); else pelicanMetadata.applySavedName?.(saved.display_name); void loadReferenceScores() }
	} catch (reason) {
		if (version === settingsRequestVersion && !disposed) settingsError.value = messageOf(reason)
	} finally {
		if (version === settingsRequestVersion) settingsSaving.value = false
	}
}
async function mutatePlan(id: number, mutate: () => Promise<unknown>): Promise<void> {
  if (!isPlansSession(plansSession) || planBusyId.value != null || planEditorOpen.value) return
  const session = plansSession
  const action = ++planActionVersion
  const isCurrent = () => isPlansSession(session) && action === planActionVersion
  planBusyId.value = id; plansError.value = ''
  try { await mutate(); if (isCurrent()) await loadPlans(session) }
  catch (reason) { if (isCurrent()) plansError.value = messageOf(reason) }
  finally { if (isCurrent()) planBusyId.value = null }
}
async function runPlan(id: number): Promise<void> { if (!pelicanMetadata.enabled) { plansError.value = '鹈鹕广场已关闭，无法开始新的测试轮次。'; return } await mutatePlan(id, () => runPelicanPlan(id)) }
async function resumePlan(id: number): Promise<void> { await mutatePlan(id, () => resumePelicanPlan(id)) }
async function savePlan(id: number | null, input: PelicanPlanInput): Promise<void> {
  if (!isPlansSession(plansSession) || !planEditorOpen.value || planBusyId.value != null) return
  const session = plansSession
  const editor = planEditorVersion
  const action = ++planActionVersion
  const isCurrent = () => isPlansSession(session) && action === planActionVersion
  planBusyId.value = id ?? -1; planSaveError.value = ''
  try {
    if (id != null) await updatePelicanPlan(id, input)
    else await createPelicanPlan({ ...input, enabled: false })
    if (!isCurrent()) return
    if (editor === planEditorVersion) saveVersion.value++
    await loadPlans(session, true)
  } catch (reason) {
    if (isCurrent() && editor === planEditorVersion) planSaveError.value = messageOf(reason)
  } finally { if (isCurrent()) planBusyId.value = null }
}
async function removePlan(id: number): Promise<void> { await mutatePlan(id, async () => { await deletePelicanPlan(id); clearArtworkCache(); closeDialog(); await refresh() }) }
function requestCleanup(plan: PelicanPlan, scope: 'failed' | 'expired'): void { if (planBusyId.value != null || planEditorOpen.value || isPlanRunning(plan)) return; cleanupOpener = document.activeElement instanceof HTMLElement ? document.activeElement : null; cleanupConfirmation.value = { plan, scope } }
function cancelCleanup(): void { if (cleanupBusy.value) return; cleanupConfirmation.value = null; const target = cleanupOpener; cleanupOpener = null; void nextTick(() => target?.focus()) }
function closePlans(): void { if (!cleanupConfirmation.value && !planEditorOpen.value && planBusyId.value == null) showPlans.value = false }
async function confirmCleanup(): Promise<void> {
  const request = cleanupConfirmation.value
  if (!request || cleanupBusy.value || planBusyId.value != null || !isPlansSession(plansSession)) return
  const session = plansSession
  const action = ++planActionVersion
  const isCurrent = () => isPlansSession(session) && action === planActionVersion
  cleanupBusy.value = true; planBusyId.value = request.plan.id; plansError.value = ''; cleanupNotice.value = ''
  try {
    const { deleted_count } = await cleanupPelicanPlan(request.plan.id, request.scope)
    if (!isCurrent()) return
    cleanupNotice.value = `已删除 ${deleted_count} 条${request.scope === 'failed' ? '失败或跳过' : '过期'}记录。`
    cleanupConfirmation.value = null; cleanupOpener = null
    clearArtworkCache()
    await Promise.all([loadPlans(session), refresh()])
  } catch (reason) { if (isCurrent()) plansError.value = messageOf(reason) }
  finally { if (isCurrent()) { cleanupBusy.value = false; planBusyId.value = null } }
}
function isPlanRunning(plan: PelicanPlan): boolean { return Boolean(plan.running_until && new Date(plan.running_until).getTime() > Date.now()) }
function closeDialog(): void {
  dialogRequestVersion++; dialogAbort?.abort()
  dialogMode.value = null
  selectedRevision.value = null
  selectedEntry.value = null
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
function focusSearch(event: KeyboardEvent): void {
  if (event.key !== '/' || event.ctrlKey || event.metaKey || event.altKey || dialogMode.value || showPlans.value || showSettings.value) return
  const target = event.target
  if (target instanceof Element && target.matches('input, textarea, select, [contenteditable="true"]')) return
  event.preventDefault()
  resultsRegion.value?.closest('.pelican-page')?.querySelector<HTMLInputElement>('.search-box input')?.focus()
}
</script>

<style scoped>
.group-heading { min-width: 0; }
.model-artwork { width: 242px; flex: 0 0 242px; min-width: 0; }
.model-heading { display: flex !important; flex-direction: row !important; flex-wrap: wrap; align-items: center; gap: 8px; min-width: 0; }
.plan-reference { font-size: 11px; opacity: .75; }
.generation-time { color: inherit; font-size: inherit; font-weight: inherit; overflow-wrap: anywhere; }
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
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding-bottom: 12px; border-bottom: 1px solid var(--pelican-border); }
.browse-left, .refresh-controls { display: flex; align-items: center; gap: 8px; min-width: 0; }
.browse-left { flex-wrap: wrap; }
.refresh-controls { margin-left: auto; color: var(--pelican-muted); font-size: 11px; white-space: nowrap; }
.updated-at { white-space: nowrap; }
.search-box { display: flex; align-items: center; gap: 9px; width: 215px; height: 35px; padding: 0 11px; color: var(--pelican-muted); }
.search-box, .select-field select, .refresh-controls select, .toolbar button { border: 1px solid var(--pelican-border); border-radius: 7px; background: var(--pelican-input); }
.search-box input { width: 100%; min-width: 0; outline: none; border: 0; background: transparent; color: var(--pelican-text); font: inherit; font-size: 12px; }
.search-box input::placeholder { color: var(--pelican-muted); }
.search-box:focus-within { border-color: var(--pelican-accent); }
.select-field { position: relative; display: flex; align-items: center; }
.select-field select { appearance: none; height: 35px; min-width: 0; max-width: 180px; padding: 0 32px 0 12px; color: inherit; font: inherit; font-size: 12px; cursor: pointer; }
.select-field > svg { position: absolute; right: 9px; width: 13px; height: 13px; color: var(--pelican-muted); pointer-events: none; }
.refresh-controls select { height: 35px; min-width: 140px; padding: 0 10px; color: inherit; font: inherit; font-size: 11px; cursor: pointer; }
.toolbar button { min-height: 35px; color: inherit; padding: 0 11px; cursor: pointer; font: inherit; font-size: 11px; }
.toolbar button:disabled { opacity: .4; cursor: default; }
.refresh-button { display: inline-flex; align-items: center; justify-content: center; width: 35px; padding: 0 !important; }
.list-summary { display: flex; justify-content: space-between; align-items: center; gap: 10px; color: var(--pelican-muted); margin: 13px 0 0; font-size: 11px; }
.list-summary strong { color: var(--pelican-text); font-weight: 600; font-variant-numeric: tabular-nums; }
.list-summary i { margin: 0 9px; color: var(--pelican-muted); font-style: normal; }
.scroll-tip { display: inline-flex; align-items: center; gap: 6px; white-space: nowrap; }
.scroll-tip > span { font-size: 14px; line-height: 1; }
.refresh-error { margin: -8px 0 10px; color: #dc2626; font-size: 11px; }
.results-region { min-height: 0; flex: 1 1 auto; overflow: auto; padding: 0 3px 10px 0; container-type: inline-size; }
.gallery { display: flex; width: 100%; max-width: 100%; flex-direction: column; }
.pelican-card { min-width: 0; box-sizing: border-box; }
.group-row { display: grid; grid-template-columns: 158px minmax(0, 1fr); gap: 20px; width: 100%; padding: 20px 0 14px; border: 0; border-bottom: 1px solid var(--pelican-border); border-radius: 0; background: transparent; box-shadow: none; overflow: visible; }
.group-meta { padding-top: 7px; color: var(--pelican-muted); }
.platform { display: inline-flex; align-items: center; gap: 5px; color: var(--pelican-muted); font-size: 9px; font-weight: 600; letter-spacing: 1.2px; margin-bottom: 8px; }
.platform::before { content: ""; width: 4px; height: 4px; border-radius: 1px; background: var(--brand-color, var(--pelican-accent)); }
.group-title { margin: 0 0 4px; color: var(--pelican-text); font-size: 13px; line-height: 1.6; font-weight: 700; }
.group-title span { display: block; color: var(--pelican-muted); font-size: 11px; font-weight: 400; margin-top: 1px; }
.group-count { display: flex; align-items: center; gap: 5px; margin: 6px 0 12px; color: var(--pelican-muted); font-size: 10px; }
.group-count strong { color: var(--pelican-text); font-size: 11px; font-weight: 500; font-variant-numeric: tabular-nums; }
.model-filters { display: flex; flex-direction: column; gap: 3px; margin-left: -8px; }
.model-filter { display: flex; align-items: center; justify-content: space-between; gap: 5px; width: 100%; padding: 5px 8px; border: 0; border-radius: 5px; background: transparent; color: var(--pelican-muted); text-align: left; font-size: 11px; white-space: nowrap; cursor: pointer; }
.model-filter:hover { background: var(--pelican-surface); }
.model-filter.active { background: var(--pelican-accent-soft); color: var(--pelican-accent); font-weight: 600; }
.model-filter > svg { visibility: hidden; width: 12px; height: 12px; }
.model-filter.active > svg { visibility: visible; }
.row-navigation { display: flex; align-items: center; gap: 6px; margin-top: 17px; }
.icon-button { display: inline-flex; align-items: center; justify-content: center; width: 27px; height: 27px; padding: 0; border: 1px solid var(--pelican-border); border-radius: 5px; background: transparent; color: var(--pelican-muted); cursor: pointer; }
.icon-button:not(:disabled):hover { background: var(--pelican-surface); color: var(--pelican-accent); }
.icon-button:disabled { cursor: default; opacity: .35; }
.position-label { margin-left: 4px; color: var(--pelican-muted); font-size: 10px; font-variant-numeric: tabular-nums; }
.row-content { min-width: 0; position: relative; }
.rail { display: flex; gap: 14px; min-width: 0; overflow-x: auto; overflow-y: hidden; padding: 1px 1px 12px; scroll-snap-type: x proximity; scrollbar-width: thin; scrollbar-color: var(--pelican-border) transparent; overscroll-behavior-x: contain; }
.rail::-webkit-scrollbar { height: 5px; }
.rail::-webkit-scrollbar-thumb { background: var(--pelican-border); border-radius: 6px; }
.rail::-webkit-scrollbar-track { background: transparent; }
.model-artwork.result-card { overflow: hidden; border: 1px solid var(--pelican-border); border-radius: 9px; background: var(--pelican-surface); scroll-snap-align: start; box-shadow: var(--console-shadow, none); transition: border-color .18s, box-shadow .18s; }
.model-artwork.result-card:hover { border-color: color-mix(in srgb, var(--pelican-accent) 35%, var(--pelican-border)); box-shadow: 0 4px 18px color-mix(in srgb, var(--pelican-text) 7%, transparent); }
.artwork-shell { position: relative; width: 100%; }
.model-artwork .lazy-artwork { min-height: 0 !important; }
:global(.model-artwork .artwork-preview) { position: relative; display: block; width: 100%; aspect-ratio: 16 / 10; overflow: hidden; border: 0; border-radius: 0; padding: 0; cursor: zoom-in; background: var(--pelican-accent-soft); }
:global(.model-artwork .artwork-preview .pelican-frame) { position: absolute; inset: 0; width: 100%; height: 100% !important; max-height: none; border-radius: 0; pointer-events: none; }
.preview-overlay { position: absolute; inset: 0; z-index: 1; }
.art-placeholder { display: flex; min-height: 0; aspect-ratio: 16 / 10; flex-direction: column; align-items: center; justify-content: center; gap: 7px; border-radius: 0; background: var(--pelican-input); color: var(--pelican-muted); font-size: 11px; text-align: center; }
.art-placeholder.failed { color: #b56d59; }
.art-placeholder.running { color: var(--pelican-accent); }
.running-orbit { width: 24px; height: 24px; border: 2px solid color-mix(in srgb, var(--pelican-accent) 22%, transparent); border-top-color: var(--pelican-accent); border-radius: 50%; animation: pelican-spin 1s linear infinite; }
@keyframes pelican-spin { to { transform: rotate(360deg); } }
.art-placeholder strong { font-size: 11px; font-weight: 500; }
.art-placeholder small { color: var(--pelican-muted); font-size: 9px; }
.card-copy { padding: 9px 10px 8px; }
.model-title { display: flex; align-items: center; gap: 6px; min-width: 0; margin: 0; color: var(--pelican-text); font-size: 11px; font-weight: 600; letter-spacing: -.1px; }
.model-title :deep(svg) { width: 17px; height: 17px; flex: 0 0 17px; }
.model-name { min-width: 0; flex: 1 1 auto; overflow: hidden; color: inherit; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.status { display: inline-flex; align-items: center; gap: 4px; flex: 0 0 auto; margin-left: auto; color: var(--pelican-muted); font-size: 9px; font-weight: 400; white-space: nowrap; }
.status i { width: 4px; height: 4px; border-radius: 50%; background: currentColor; }
.status.success { color: #338069; }
.status.failed { color: #b56d59; }
.status.running { color: var(--pelican-accent); }
.status.skipped { color: var(--pelican-muted); }
.round-label { display: inline-block; margin-top: 6px; color: var(--pelican-muted); font-size: 9px; font-variant-numeric: tabular-nums; }
.card-footer { display: flex; align-items: center; justify-content: space-between; gap: 6px; margin-top: 7px; border-top: 1px solid var(--pelican-border); padding-top: 6px; color: var(--pelican-muted); font-size: 9px; font-variant-numeric: tabular-nums; }
.card-footer time { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.duration { display: inline-flex; align-items: center; gap: 4px; flex: 0 0 auto; white-space: nowrap; }
.duration svg { width: 12px; height: 12px; }
.card-actions { position: absolute; right: 9px; bottom: 9px; display: flex; gap: 6px; margin: 0; opacity: 0; transform: translateY(3px); transition: opacity .15s, transform .15s; }
.artwork-shell:hover .card-actions, .artwork-shell:focus-within .card-actions { opacity: 1; transform: none; }
.card-actions button { display: inline-flex; padding: 5px; border: 0; border-radius: 5px; background: #ffffffdc; color: #4a635c; cursor: pointer; }
.card-actions button:hover { background: #fff; }
.card-actions button:disabled { display: none; }
.success { @apply text-emerald-700 dark:text-emerald-400; }
.skipped { color: var(--pelican-muted); }
.stars { color: var(--pelican-accent); letter-spacing: 1px; white-space: nowrap; }
.last-success { margin: 6px 0 0; color: var(--pelican-muted); font-size: 9px; }
.rail-empty { display: grid; width: 100%; min-height: 154px; align-content: center; justify-items: center; gap: 7px; border: 1px dashed var(--pelican-border); border-radius: 8px; color: var(--pelican-muted); font-size: 11px; }
.text-button { border: 0; background: transparent; color: var(--pelican-accent); padding: 3px 0; cursor: pointer; font: inherit; font-size: 11px; }
.pagination { display: flex; flex: 0 0 auto; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px; margin-top: 12px; color: var(--pelican-muted); }
.page-controls { display: flex; align-items: center; gap: 10px; }
.page-controls select { min-width: 64px; margin-left: 6px; }
.page-controls button, .empty-state button, .artwork-actions button { border: 1px solid var(--pelican-border); background: var(--pelican-input); color: var(--pelican-text); border-radius: 7px; padding: 7px 10px; cursor: pointer; font: inherit; }
.toolbar button:not(:disabled):hover, .page-controls button:not(:disabled):hover, .empty-state button:hover, .artwork-actions button:hover { background: var(--pelican-hover); border-color: var(--pelican-accent); }
.page-controls button:disabled { opacity: .4; cursor: default; }
.empty-state { text-align: center; padding: 60px 15px; border: 1px dashed var(--pelican-border); border-radius: 12px; color: var(--pelican-muted); }
.empty-state h2 { font-size: 16px; color: var(--pelican-text); margin-top: 15px; }
.empty-state p { margin: 8px 0 18px; }
.dialog-content { container-type: inline-size; }
.artwork-actions { display: flex; justify-content: space-between; align-items: center; gap: 10px; margin-top: 12px; font-size: 12px; }
.enlarged-art { max-width: 560px; margin: auto; }
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
@media (max-width: 760px) {
  .toolbar { align-items: stretch; flex-direction: column; gap: 10px; }
  .browse-left { width: 100%; }
  .search-box { width: 100%; }
  .select-field { flex: 1 1 0; min-width: 0; }
  .select-field select { width: 100%; max-width: none; }
  .refresh-controls { width: 100%; margin: 0; }
  .refresh-controls select { min-width: 0; flex: 1; }
  .list-summary { align-items: flex-start; flex-direction: column; gap: 5px; }
  .page-controls { flex-wrap: wrap; gap: 7px; }
  .group-row { display: block; padding: 18px 0 12px; }
  .group-meta { padding-top: 0; position: relative; }
  .group-title { padding-right: 82px; font-size: 14px; }
  .group-title span { display: none; }
  .group-count { position: absolute; top: 20px; right: 0; margin: 0; }
  .model-filters { flex-direction: row; gap: 5px; margin: 11px 0 13px; overflow-x: auto; padding: 1px; }
  .model-filter { width: auto; flex: 0 0 auto; padding: 4px 10px; border: 1px solid var(--pelican-border); font-size: 10px; }
  .model-filter.active { border-color: transparent; }
  .model-filter > svg { display: none; }
  .row-navigation { display: none; }
  .rail { gap: 12px; }
}
@media (max-width: 540px) {
  .refresh-controls > select { max-width: 100%; }
  .model-artwork { width: 242px; flex-basis: 242px; }
}
</style>
