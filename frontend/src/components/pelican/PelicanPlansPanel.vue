<template>
  <section class="pelican-plans" :aria-busy="loading">
    <header class="panel-header">
      <div>
        <p class="eyebrow">计划管理</p>
        <h2>运行概览</h2>
        <p>集中查看运行状态、调用额度和结果留存策略。</p>
      </div>
      <button v-if="!showEditor" class="primary-button" type="button" @click="startCreate">新建计划</button>
    </header>
    <div class="overview-grid" aria-label="计划概览">
      <div class="overview-item"><span>启用中</span><strong>{{ activeCount }}</strong><small>按计划自动运行</small></div>
      <div class="overview-item"><span>自动暂停</span><strong>{{ pausedCount }}</strong><small>需要人工恢复</small></div>
      <div class="overview-item"><span>今日调用</span><strong>{{ dailyCalls }}</strong><small>按 UTC+8 统计</small></div>
      <div class="overview-item"><span>正在测试</span><strong>{{ runningCount }}</strong><small>当前执行中的轮次</small></div>
    </div>
    <p v-if="error" class="error form-message" role="alert">{{ error }}</p><p v-if="loading" class="form-message">加载中…</p>
    <section v-if="showEditor" class="editor-panel">
      <div class="section-heading">
        <div><span class="eyebrow">{{ editingID ? '计划设置' : '新建计划' }}</span><h3>{{ editingID ? '编辑计划' : '配置一个新的测试计划' }}</h3><p>手动和定时测试均可能产生费用；新计划默认停用，保存后可编辑启用。</p></div>
        <button class="ghost-button" type="button" :disabled="busyId != null" @click="reset">关闭</button>
      </div>
      <form class="plan-form" @submit.prevent="save">
        <div class="form-section"><span class="form-section-title">测试目标</span><div class="form-grid">
          <label>分组<select v-model.number="draft.group_id" required aria-label="分组"><option :value="0" disabled>选择分组</option><option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}{{ group.platform ? ` · ${group.platform}` : '' }}</option></select></label>
          <label>模型<select v-model="draft.model_id" required :disabled="!draft.group_id || modelsLoading" aria-label="模型"><option value="" disabled>{{ modelPlaceholder }}</option><option v-for="model in models" :key="model" :value="model">{{ model }}</option></select></label>
          <label><span class="field-title">思考强度 <span class="field-help" tabindex="0" aria-label="思考强度说明：各模型支持的档位不同，默认使用模型自身设置。">?<span class="field-tooltip" role="tooltip">各模型支持的档位不同，默认使用模型自身设置。</span></span></span><select v-model="draft.reasoning_effort" aria-label="思考强度"><option v-for="option in reasoningEffortOptions" :key="option.value" :value="option.value">{{ option.label }}</option></select></label>
        </div></div>
        <p v-if="modelError" class="model-status error" role="alert">{{ modelError }} <button type="button" :disabled="modelsLoading" @click="reloadModels">重试</button></p>
        <p v-else-if="modelsLoading" class="model-status" role="status">正在加载该分组的已配置模型…</p>
        <p v-else-if="draft.group_id && !models.length" class="model-status" role="status">该分组没有可用于测试的已配置模型。<button type="button" @click="reloadModels">重新加载</button></p>
        <p v-else-if="modelSelectionNotice" class="model-status error" role="alert">{{ modelSelectionNotice }}</p>
        <div class="form-section"><span class="form-section-title">执行策略</span><div class="form-grid">
          <label>间隔（分钟）<input v-model.number="draft.interval_minutes" required type="number" min="15" max="1440" aria-label="间隔分钟"></label>
          <label>测试超时（秒）<input v-model.number="draft.timeout_seconds" required type="number" min="30" max="3600" aria-label="测试超时秒" title="30–3600 秒，默认 180 秒"></label>
          <label>最小字符数<input v-model.number="draft.min_chars" required type="number" min="100" max="50000" aria-label="最小字符数"></label>
          <label v-if="editingID" class="enabled"><input v-model="draft.enabled" :disabled="editingPaused" type="checkbox"> {{ editingPaused ? '已自动暂停，请使用下方“恢复计划”' : '启用定时测试' }}</label>
        </div></div>
        <div class="form-section"><span class="form-section-title">费用与留存</span><div class="form-grid">
          <label>保留结果数<input v-model.number="draft.max_results" required type="number" min="1" max="50" aria-label="保留结果数"></label>
          <label>每日调用上限（0 为不限）<input v-model.number="draft.daily_call_limit" required type="number" min="0" max="100000" aria-label="每日调用上限"></label>
          <label>连续失败暂停阈值（0 为关闭）<input v-model.number="draft.failure_pause_threshold" required type="number" min="0" max="100" aria-label="连续失败暂停阈值"></label>
          <label>按天保留（0 为关闭）<select v-model="retentionMode" aria-label="按天保留" @change="setRetentionPreset"><option value="0">不按天清理</option><option value="7">保留 7 天</option><option value="30">保留 30 天</option><option value="custom">自定义天数</option></select><input v-if="retentionMode === 'custom'" v-model.number="draft.retention_days" required type="number" min="0" max="3650" aria-label="自定义保留天数"></label>
        </div></div>
        <div class="form-actions"><button v-if="editingID" class="ghost-button" type="button" :disabled="busyId != null" @click="reset">取消</button><button class="primary-button save-plan" type="submit" :disabled="loading || busyId != null || modelsLoading || !!modelError || !draft.group_id || !isCurrentModel">{{ editingID ? '保存计划' : '新建计划' }}</button></div>
      </form>
      <p class="retention-note">每轮由分组调度选择一个账号，最多测试一次。每日用量按 UTC+8 重置；0 表示不设上限或关闭对应限制。每个计划和分组保留最近 {{ draft.max_results }} 条结果，并保护最新成功作品。</p>
    </section>
    <div class="list-toolbar">
      <div><span class="eyebrow">计划列表</span><strong>{{ filteredPlans.length }} 个计划</strong></div>
      <label class="filter-control">筛选<select v-model="planFilter" aria-label="筛选计划"><option value="all">全部状态</option><option value="enabled">已启用</option><option value="paused">自动暂停</option><option value="disabled">已停用</option></select></label>
    </div>
    <div v-if="!loading && !plans.length" class="empty-state"><strong>还没有测试计划</strong><span>创建计划后，系统会按设定的分组、模型和时间间隔自动运行测试。</span><button class="primary-button" type="button" @click="startCreate">创建第一个计划</button></div>
    <div v-else-if="!loading && !filteredPlans.length" class="empty-state compact"><strong>没有符合条件的计划</strong><span>换一个状态筛选试试。</span></div>
    <div v-else class="plans-list">
      <article v-for="plan in filteredPlans" :key="plan.id" class="plan-card">
        <header class="plan-card-header">
          <div class="plan-title"><div class="status-line"><span class="status-badge" :class="`status-${planStatus(plan)}`">{{ planStatusLabel(plan) }}</span><strong>{{ plan.group_name }}</strong></div><span class="plan-model">{{ plan.model_id }}</span></div>
          <div class="plan-schedule"><strong>{{ plan.interval_minutes }} 分钟</strong><span>·</span><span>{{ plan.enabled ? `下次 ${planTime(plan.next_run_at)}` : isAutoPaused(plan) ? '等待恢复' : '定时已停用' }}</span></div>
        </header>
        <div class="plan-metrics">
          <div><span>今日调用</span><strong>{{ plan.daily_calls_used ?? 0 }}<small>{{ plan.daily_call_limit ? ` / ${plan.daily_call_limit}` : ' / 不限' }}</small></strong><em>{{ plan.usage_day || 'UTC+8 当日' }}</em></div>
          <div><span>上次运行</span><strong>{{ plan.last_run_at ? planTime(plan.last_run_at) : '尚未运行' }}</strong><em>{{ plan.last_run_calls ?? 0 }} 次调用</em></div>
          <div><span>连续失败</span><strong>{{ plan.consecutive_failed_runs ?? 0 }}<small>{{ plan.failure_pause_threshold ? ` / ${plan.failure_pause_threshold}` : ' 次' }}</small></strong><em>{{ plan.pause_reason === 'consecutive_failures' ? '已触发自动暂停' : '自动暂停未触发' }}</em></div>
          <div><span>结果留存</span><strong>{{ plan.max_results }}<small> 条</small></strong><em>{{ plan.retention_days ? `按天保留 ${plan.retention_days} 天` : '不按天清理' }}</em></div>
        </div>
        <p class="plan-footnote">思考强度：{{ reasoningEffortLabel(plan.reasoning_effort) }} · 超时 {{ plan.timeout_seconds ?? 180 }} 秒 · 本轮预估 {{ plan.estimated_calls_per_run ?? 0 }} 次调用</p>
        <div v-if="isAutoPaused(plan) || isDailyLimitReached(plan) || isRunning(plan)" class="plan-alert" :class="{ warning: isRunning(plan), error: !isRunning(plan) }">
          <span v-if="isAutoPaused(plan)">计划已因连续失败自动暂停，请恢复后再运行。</span>
          <span v-else-if="isDailyLimitReached(plan)">今日调用次数已达到上限，按 UTC+8 日期重置。</span>
          <span v-else>本轮正在测试，请等待结果。</span>
        </div>
        <footer class="plan-actions">
          <button class="primary-button" :disabled="busyId != null || !pelicanEnabled || isRunning(plan) || isAutoPaused(plan) || isDailyLimitReached(plan)" @click="$emit('run', plan.id)">立即运行</button>
          <button v-if="plan.pause_reason === 'consecutive_failures'" class="secondary-button" :disabled="busyId != null || isRunning(plan)" @click="$emit('resume', plan.id)">恢复计划</button>
          <button class="secondary-button" :disabled="busyId != null" @click="startEdit(plan)">编辑</button>
          <details class="plan-more"><summary>更多</summary><div class="more-menu"><button :disabled="busyId != null || isRunning(plan)" @click="$emit('cleanup-request', plan, 'failed')">清理失败/跳过记录</button><button :disabled="busyId != null || isRunning(plan)" @click="$emit('cleanup-request', plan, 'expired')">按保留规则清理</button><button class="danger-button" :disabled="busyId != null" @click="$emit('delete', plan.id)">删除</button></div></details>
        </footer>
      </article>
    </div>
  </section>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { listPelicanPlanModels, type PelicanPlan, type PelicanPlanInput } from '@/api/pelicanTests'

const props = withDefaults(defineProps<{ plans: PelicanPlan[]; groups: Array<{ id: number; name: string; platform?: string }>; loading: boolean; error?: string; busyId?: number | null; saveVersion?: number; pelicanEnabled?: boolean }>(), { pelicanEnabled: true })
const emit = defineEmits<{ run: [id: number]; resume: [id: number]; save: [id: number | null, input: PelicanPlanInput]; 'delete': [id: number]; 'cleanup-request': [plan: PelicanPlan, scope: 'failed' | 'expired'] }>()
const editingID = ref<number | null>(null)
const showEditor = ref(false)
const planFilter = ref<'all' | 'enabled' | 'paused' | 'disabled'>('all')
const models = ref<string[]>([])
const modelsLoading = ref(false)
const modelError = ref('')
const modelSelectionNotice = ref('')
let modelsRequestVersion = 0
let modelsAbort: AbortController | undefined
let retainedModelForGroup: { groupID: number; modelID: string } | undefined
const reasoningEffortOptions = [{ value: '', label: '默认' }, { value: 'none', label: '无' }, { value: 'minimal', label: '最低' }, { value: 'low', label: '低' }, { value: 'medium', label: '中' }, { value: 'high', label: '高' }, { value: 'xhigh', label: '极高' }]
const defaults = (): PelicanPlanInput => ({ group_id: 0, model_id: '', interval_minutes: 60, timeout_seconds: 180, enabled: false, max_results: 20, min_chars: 9366, daily_call_limit: 0, failure_pause_threshold: 0, retention_days: 0, reasoning_effort: '' })
const draft = reactive<PelicanPlanInput>(defaults())
const isCurrentModel = computed(() => models.value.includes(draft.model_id))
const editingPaused = computed(() => props.plans.find(plan => plan.id === editingID.value)?.pause_reason === 'consecutive_failures')
const retentionMode = ref('0')
const modelPlaceholder = computed(() => !draft.group_id ? '请先选择分组' : modelsLoading.value ? '加载中…' : modelError.value ? '模型加载失败' : models.value.length ? '选择已配置模型' : '没有可选模型')
const activeCount = computed(() => props.plans.filter(plan => plan.enabled && !isAutoPaused(plan)).length)
const pausedCount = computed(() => props.plans.filter(plan => isAutoPaused(plan)).length)
const dailyCalls = computed(() => {
  const today = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10)
  return props.plans.reduce((total, plan) => total + (plan.usage_day === today ? plan.daily_calls_used ?? 0 : 0), 0)
})
const runningCount = computed(() => props.plans.filter(plan => isRunning(plan)).length)
const filteredPlans = computed(() => props.plans.filter(plan => planFilter.value === 'all' || (planFilter.value === 'enabled' && plan.enabled && !isAutoPaused(plan)) || (planFilter.value === 'paused' && isAutoPaused(plan)) || (planFilter.value === 'disabled' && !plan.enabled && !isAutoPaused(plan))))
function reset(): void { editingID.value = null; showEditor.value = false; retainedModelForGroup = undefined; retentionMode.value = '0'; Object.assign(draft, defaults()); clearModels() }
function startCreate(): void { reset(); showEditor.value = true }
function startEdit(plan: PelicanPlan): void {
  const sameGroup = draft.group_id === plan.group_id
  editingID.value = plan.id
  showEditor.value = true
  retainedModelForGroup = { groupID: plan.group_id, modelID: plan.model_id }
  Object.assign(draft, { group_id: plan.group_id, model_id: plan.model_id, interval_minutes: plan.interval_minutes, timeout_seconds: plan.timeout_seconds ?? 180, enabled: plan.pause_reason === 'consecutive_failures' ? false : plan.enabled, max_results: plan.max_results, min_chars: plan.min_chars, daily_call_limit: plan.daily_call_limit ?? 0, failure_pause_threshold: plan.failure_pause_threshold ?? 0, retention_days: plan.retention_days ?? 0, reasoning_effort: plan.reasoning_effort ?? '' })
  retentionMode.value = [0, 7, 30].includes(draft.retention_days ?? 0) ? String(draft.retention_days ?? 0) : 'custom'
  if (sameGroup) { retainedModelForGroup = undefined; void loadModels(plan.group_id) }
}
watch(() => props.saveVersion, () => reset())
watch(() => draft.group_id, (groupID, previousGroupID) => {
  if (!groupID) { clearModels(); return }
  const retain = retainedModelForGroup?.groupID === groupID ? retainedModelForGroup.modelID : ''
  retainedModelForGroup = undefined
  if (groupID !== previousGroupID && !retain) { draft.model_id = ''; modelSelectionNotice.value = '' }
  void loadModels(groupID)
})
function clearModels(): void { modelsRequestVersion++; modelsAbort?.abort(); models.value = []; modelsLoading.value = false; modelError.value = ''; modelSelectionNotice.value = '' }
async function loadModels(groupID: number): Promise<void> {
  const version = ++modelsRequestVersion
  modelsAbort?.abort()
  const controller = new AbortController()
  modelsAbort = controller
  models.value = []; modelsLoading.value = true; modelError.value = ''; modelSelectionNotice.value = ''
  try {
    const loaded = await listPelicanPlanModels(groupID, controller.signal)
    if (version !== modelsRequestVersion || draft.group_id !== groupID) return
    models.value = loaded
    if (draft.model_id && !loaded.includes(draft.model_id)) { draft.model_id = ''; modelSelectionNotice.value = '原模型已不在该分组的可选模型中，请重新选择。' }
  } catch (reason) {
    if (version !== modelsRequestVersion || draft.group_id !== groupID) return
    modelError.value = reason instanceof Error && reason.message ? `无法加载模型：${reason.message}` : '无法加载模型，请稍后重试。'
    draft.model_id = ''
  } finally { if (version === modelsRequestVersion) modelsLoading.value = false }
}
function reloadModels(): void { if (draft.group_id) void loadModels(draft.group_id) }
function setRetentionPreset(): void { if (retentionMode.value !== 'custom') draft.retention_days = Number(retentionMode.value); else if ([0, 7, 30].includes(draft.retention_days ?? 0)) draft.retention_days = 1 }
function reasoningEffortLabel(value?: string): string { return reasoningEffortOptions.find(option => option.value === (value ?? ''))?.label ?? '默认' }
function save(): void { if (props.busyId != null || props.loading || modelsLoading.value || modelError.value || !draft.group_id || !isCurrentModel.value) return; emit('save', editingID.value, { ...draft, enabled: editingID.value ? draft.enabled : false }) }
function isRunning(plan: PelicanPlan): boolean { return Boolean(plan.running_until && new Date(plan.running_until).getTime() > Date.now()) }
function isAutoPaused(plan: PelicanPlan): boolean { return plan.pause_reason === 'consecutive_failures' }
function isDailyLimitReached(plan: PelicanPlan): boolean { return Boolean(plan.daily_call_limit && (plan.daily_calls_used ?? 0) >= plan.daily_call_limit) }
function planStatus(plan: PelicanPlan): 'running' | 'paused' | 'enabled' | 'disabled' { if (isRunning(plan)) return 'running'; if (isAutoPaused(plan)) return 'paused'; return plan.enabled ? 'enabled' : 'disabled' }
function planStatusLabel(plan: PelicanPlan): string { return { running: '正在测试', paused: '自动暂停', enabled: '已启用', disabled: '已停用' }[planStatus(plan)] }
function planTime(value?: string | null): string { const date = value ? new Date(value) : null; return date && !Number.isNaN(date.getTime()) ? date.toLocaleString('zh-CN') : '暂无' }
onBeforeUnmount(() => { modelsRequestVersion++; modelsAbort?.abort() })
</script>
<style scoped>
.pelican-plans {
  /* BaseDialog teleports outside the console theme variable scope. */
  --plan-border: #d8cec2;
  --plan-surface: #faf9f5;
  --plan-card: #fff;
  --plan-text: #252525;
  --plan-muted: #686761;
  --plan-accent: #4338ca;
  --plan-soft: #f1f0fb;
  color-scheme: light;
}
.dark .pelican-plans {
  --plan-border: #334155;
  --plan-surface: #111c2e;
  --plan-card: #172337;
  --plan-text: #e2e8f0;
  --plan-muted: #a5b4c8;
  --plan-accent: #a5b4fc;
  --plan-soft: #253150;
  color-scheme: dark;
}
.pelican-plans {
  min-width: 0;
  padding: 4px 0 8px;
  color: var(--plan-text);
  font-size: 13px;
}
.pelican-plans * { box-sizing: border-box; }
.pelican-plans :is(button, input, select, summary):focus-visible { outline: 2px solid #818cf8; outline-offset: 2px; }
.pelican-plans :is(input, select, button):disabled { opacity: .58; cursor: not-allowed; }
.pelican-plans button { min-height: 34px; padding: 6px 12px; border: 1px solid var(--plan-border); border-radius: 8px; background: var(--plan-card); color: var(--plan-text); font-size: 12px; font-weight: 600; cursor: pointer; }
.pelican-plans button:not(:disabled):hover { border-color: #818cf8; }
.pelican-plans .primary-button { border-color: #5046b6; background: #5046b6; color: #fff; }
.pelican-plans .primary-button:not(:disabled):hover { border-color: #4338ca; background: #4338ca; }
.pelican-plans .ghost-button { border-color: transparent; background: transparent; color: var(--plan-muted); }
.pelican-plans .secondary-button { background: transparent; }
.pelican-plans .danger-button { color: #b91c1c; }
.dark .pelican-plans .danger-button { color: #fca5a5; }
.panel-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin: 0 0 18px; }
.panel-header h2 { margin: 3px 0 4px; font-size: 20px; line-height: 1.3; }
.panel-header p, .section-heading p { margin: 0; color: var(--plan-muted); font-size: 12px; }
.pelican-plans .eyebrow { color: var(--plan-accent); font-size: 11px; font-weight: 700; letter-spacing: .08em; }
.overview-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; margin-bottom: 22px; }
.overview-item { display: flex; flex-direction: column; gap: 4px; min-width: 0; padding: 12px 14px; border: 1px solid var(--plan-border); border-radius: 10px; background: var(--plan-card); }
.overview-item span, .overview-item small { color: var(--plan-muted); font-size: 11px; }
.overview-item strong { font-size: 22px; line-height: 1.1; font-variant-numeric: tabular-nums; }
.form-message { margin: 8px 0; }
.editor-panel { margin-bottom: 22px; padding: 18px; border: 1px solid var(--plan-border); border-radius: 12px; background: var(--plan-card); }
.section-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 18px; }
.section-heading > div { min-width: 0; }
.section-heading > button { flex: none; white-space: nowrap; }
.section-heading h3 { margin: 4px 0 5px; font-size: 16px; }
.plan-form { display: grid; gap: 18px; }
.form-section { display: grid; gap: 10px; }
.form-section + .form-section { padding-top: 16px; border-top: 1px solid var(--plan-border); }
.form-section-title { font-weight: 700; }
.form-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; align-items: start; }
.plan-form label { display: flex; flex-direction: column; gap: 6px; min-width: 0; color: var(--plan-muted); font-size: 12px; line-height: 18px; }
.field-title { display: flex; align-items: center; gap: 5px; min-height: 18px; }
.field-help { position: relative; display: inline-grid; place-items: center; width: 14px; height: 14px; border: 1px solid var(--plan-border); border-radius: 50%; font-size: 10px; cursor: help; }
.field-tooltip { display: none; position: absolute; bottom: calc(100% + 7px); right: 0; width: 190px; padding: 8px 10px; border: 1px solid var(--plan-border); border-radius: 6px; background: var(--plan-card); color: var(--plan-text); box-shadow: 0 4px 12px #0002; font-size: 12px; line-height: 1.5; z-index: 2; }
.field-help:hover .field-tooltip, .field-help:focus .field-tooltip { display: block; }
.plan-form input, .plan-form select, .filter-control select { width: 100%; min-width: 0; min-height: 36px; padding: 7px 9px; border: 1px solid var(--plan-border); border-radius: 7px; background: var(--plan-surface); color: var(--plan-text); }
.plan-form label.enabled { flex-direction: row; align-items: center; align-self: end; min-height: 36px; color: var(--plan-text); }
.plan-form input[type=checkbox] { width: auto; min-height: auto; accent-color: #5046b6; }
.model-status { margin: -8px 0 0; color: var(--plan-muted); }
.model-status button { margin-left: 6px; }
.form-actions { display: flex; justify-content: flex-end; gap: 8px; }
.retention-note { margin: 14px 0 0; padding-top: 12px; border-top: 1px solid var(--plan-border); color: var(--plan-muted); font-size: 12px; line-height: 1.6; }
.list-toolbar { display: flex; justify-content: space-between; align-items: end; gap: 14px; margin-bottom: 12px; }
.list-toolbar > div { display: grid; gap: 4px; }
.list-toolbar strong { font-size: 16px; }
.filter-control { display: flex; align-items: center; gap: 8px; color: var(--plan-muted); font-size: 12px; }
.filter-control select { width: 138px; }
.plans-list { display: grid; gap: 12px; }
.plan-card { min-width: 0; padding: 16px; border: 1px solid var(--plan-border); border-radius: 12px; background: var(--plan-card); }
.plan-card-header { display: flex; justify-content: space-between; gap: 12px; margin-bottom: 16px; }
.plan-title { display: grid; gap: 6px; min-width: 0; }
.status-line { display: flex; align-items: center; gap: 8px; min-width: 0; }
.status-line strong { overflow-wrap: anywhere; font-size: 15px; }
.status-badge { flex: none; padding: 3px 7px; border-radius: 5px; font-size: 11px; font-weight: 700; }
.status-enabled { background: #dcfce7; color: #166534; }
.status-disabled { background: #e5e7eb; color: #374151; }
.status-paused { background: #fef3c7; color: #92400e; }
.status-running { background: #dbeafe; color: #1d4ed8; }
.plan-model { color: var(--plan-muted); font-size: 12px; overflow-wrap: anywhere; }
.plan-schedule { display: flex; align-items: baseline; gap: 5px; color: var(--plan-muted); font-size: 12px; white-space: nowrap; }
.plan-schedule strong { color: var(--plan-text); }
.plan-metrics { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; }
.plan-metrics > div { display: grid; gap: 5px; min-width: 0; padding: 10px; border-radius: 8px; background: var(--plan-surface); }
.plan-metrics span, .plan-metrics em { color: var(--plan-muted); font-size: 11px; font-style: normal; }
.plan-metrics strong { overflow-wrap: anywhere; font-size: 13px; font-weight: 700; font-variant-numeric: tabular-nums; }
.plan-metrics strong small { color: var(--plan-muted); font-size: 11px; font-weight: 400; }
.plan-footnote { margin: 10px 0 0; color: var(--plan-muted); font-size: 11px; }
.plan-alert { margin-top: 12px; padding: 9px 10px; border-radius: 7px; background: #fef3c7; color: #78350f; font-size: 12px; }
.plan-alert.error { background: #fee2e2; color: #991b1b; }
.dark .pelican-plans .plan-alert.warning { background: #3d2f1b; color: #fde68a; }
.dark .pelican-plans .plan-alert.error { background: #4b1d29; color: #fecaca; }
.plan-actions { display: flex; align-items: center; justify-content: flex-end; gap: 8px; margin-top: 14px; padding-top: 12px; border-top: 1px solid var(--plan-border); }
.plan-more { position: relative; }
.plan-more summary { display: inline-flex; align-items: center; justify-content: center; min-height: 34px; padding: 6px 12px; border: 1px solid var(--plan-border); border-radius: 8px; font-size: 12px; font-weight: 600; cursor: pointer; list-style: none; }
.plan-more summary::-webkit-details-marker { display: none; }
.plan-more summary::after { content: '⌄'; margin-left: 7px; }
.plan-more[open] summary { border-color: #818cf8; }
.more-menu { position: absolute; right: 0; bottom: calc(100% + 6px); z-index: 4; display: grid; gap: 2px; min-width: 190px; padding: 5px; border: 1px solid var(--plan-border); border-radius: 9px; background: var(--plan-card); box-shadow: 0 12px 24px #0003; }
.more-menu button { width: 100%; border: 0; text-align: left; white-space: nowrap; }
.empty-state { display: grid; justify-items: center; gap: 8px; padding: 36px 16px; border: 1px dashed var(--plan-border); border-radius: 12px; color: var(--plan-muted); text-align: center; }
.empty-state strong { color: var(--plan-text); font-size: 15px; }
.empty-state .primary-button { margin-top: 6px; }
.empty-state.compact { padding: 24px 16px; }
.error { color: #dc2626; }
.dark .pelican-plans .error { color: #fca5a5; }
@media (max-width: 700px) {
  .overview-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .form-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .plan-metrics { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .plan-card-header { flex-direction: column; }
  .plan-schedule { white-space: normal; }
}
@media (max-width: 430px) {
  .panel-header { align-items: flex-start; }
  .panel-header h2 { font-size: 18px; }
  .overview-item { padding: 10px; }
  .overview-item small { display: none; }
  .editor-panel, .plan-card { padding: 12px; }
  .form-grid { grid-template-columns: 1fr; }
  .plan-actions { flex-wrap: wrap; }
  .plan-actions > button:first-child { margin-right: auto; }
}
</style>
