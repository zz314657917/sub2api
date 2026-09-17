<template>
  <section class="pelican-plans" :aria-busy="loading">
    <header><h2>{{ editingID ? '编辑计划' : '新建计划' }}</h2><p>手动和定时测试均可能产生费用；新计划默认停用，保存后可编辑启用。</p></header>
    <p v-if="error" class="error" role="alert">{{ error }}</p><p v-if="loading">加载中…</p>
    <form class="plan-form" @submit.prevent="save">
      <label>分组<select v-model.number="draft.group_id" required aria-label="分组"><option :value="0" disabled>选择分组</option><option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}</option></select></label>
      <label>模型<select v-model="draft.model_id" required :disabled="!draft.group_id || modelsLoading" aria-label="模型"><option value="" disabled>{{ modelPlaceholder }}</option><option v-for="model in models" :key="model" :value="model">{{ model }}</option></select></label>
      <label><span class="field-title">思考强度 <span class="field-help" tabindex="0" aria-label="思考强度说明：各模型支持的档位不同，默认使用模型自身设置。">?<span class="field-tooltip" role="tooltip">各模型支持的档位不同，默认使用模型自身设置。</span></span></span><select v-model="draft.reasoning_effort" aria-label="思考强度"><option v-for="option in reasoningEffortOptions" :key="option.value" :value="option.value">{{ option.label }}</option></select></label>
      <p v-if="modelError" class="model-status error" role="alert">{{ modelError }} <button type="button" :disabled="modelsLoading" @click="reloadModels">重试</button></p>
      <p v-else-if="modelsLoading" class="model-status" role="status">正在加载该分组的已配置模型…</p>
      <p v-else-if="draft.group_id && !models.length" class="model-status" role="status">该分组没有可用于测试的已配置模型。<button type="button" @click="reloadModels">重新加载</button></p>
      <p v-else-if="modelSelectionNotice" class="model-status error" role="alert">{{ modelSelectionNotice }}</p>
      <label>间隔（分钟）<input v-model.number="draft.interval_minutes" required type="number" min="15" max="1440" aria-label="间隔分钟"></label>
      <label>保留结果数<input v-model.number="draft.max_results" required type="number" min="1" max="50" aria-label="保留结果数"></label>
      <label>最小字符数<input v-model.number="draft.min_chars" required type="number" min="100" max="50000" aria-label="最小字符数"></label>
      <label>每日调用上限（0 为不限）<input v-model.number="draft.daily_call_limit" required type="number" min="0" max="100000" aria-label="每日调用上限"></label>
      <label>连续失败暂停阈值（0 为关闭）<input v-model.number="draft.failure_pause_threshold" required type="number" min="0" max="100" aria-label="连续失败暂停阈值"></label>
      <label>按天保留（0 为关闭）<select v-model="retentionMode" aria-label="按天保留" @change="setRetentionPreset"><option value="0">不按天清理</option><option value="7">保留 7 天</option><option value="30">保留 30 天</option><option value="custom">自定义天数</option></select><input v-if="retentionMode === 'custom'" v-model.number="draft.retention_days" required type="number" min="0" max="3650" aria-label="自定义保留天数"></label>
      <label v-if="editingID" class="enabled"><input v-model="draft.enabled" :disabled="editingPaused" type="checkbox"> {{ editingPaused ? '已自动暂停，请使用下方“恢复计划”' : '启用定时测试' }}</label>
      <div class="form-actions"><button v-if="editingID" type="button" :disabled="busyId != null" @click="reset">取消</button><button class="save-plan" type="submit" :disabled="loading || busyId != null || modelsLoading || !!modelError || !draft.group_id || !isCurrentModel">{{ editingID ? '保存计划' : '新建计划' }}</button></div>
    </form>
    <p class="retention-note">每轮由分组调度选择一个账号，最多测试一次。每日用量按 UTC+8 重置；0 表示不设上限或关闭对应限制。每个计划和分组保留最近 {{ draft.max_results }} 条结果，并保护最新成功作品。</p>
    <div v-if="!loading && !plans.length">尚无计划。</div>
    <article v-for="plan in plans" :key="plan.id" class="plan"><strong>{{ plan.group_name }} · {{ plan.model_id }}</strong><span>{{ plan.interval_minutes }} 分钟 · 保留 {{ plan.max_results }} 条 · {{ plan.enabled ? '已启用' : '已停用' }}</span><small class="plan-timing">思考强度：{{ reasoningEffortLabel(plan.reasoning_effort) }} · 每日调用：{{ plan.daily_calls_used ?? 0 }} / {{ plan.daily_call_limit || '不限' }}（{{ plan.usage_day || 'UTC+8 当日' }}） · 本轮预估 {{ plan.estimated_calls_per_run ?? 0 }} 次 · 上轮实际 {{ plan.last_run_calls ?? 0 }} 次</small><small class="plan-timing">连续失败 {{ plan.consecutive_failed_runs ?? 0 }} 轮{{ plan.failure_pause_threshold ? ` / ${plan.failure_pause_threshold}` : '（自动暂停关闭）' }}{{ plan.pause_reason === 'consecutive_failures' ? ' · 已因连续失败自动暂停' : '' }} · 按天保留 {{ plan.retention_days || '关闭' }}{{ plan.retention_days ? ' 天' : '' }}</small><small v-if="isAutoPaused(plan)" class="plan-timing error">计划已自动暂停，请先恢复计划后再运行。</small><small v-else-if="isDailyLimitReached(plan)" class="plan-timing error">今日调用次数已达到上限（按 UTC+8 日期重置）。</small><small class="plan-timing">上次：{{ planTime(plan.last_run_at) }} · 下次：{{ plan.enabled ? planTime(plan.next_run_at) : '已停用' }}<b v-if="isRunning(plan)"> · 正在测试</b></small><div><button :disabled="busyId != null || !pelicanEnabled || isRunning(plan) || isAutoPaused(plan) || isDailyLimitReached(plan)" @click="$emit('run', plan.id)">立即运行</button><button v-if="plan.pause_reason === 'consecutive_failures'" :disabled="busyId != null || isRunning(plan)" @click="$emit('resume', plan.id)">恢复计划</button><button :disabled="busyId != null" @click="startEdit(plan)">编辑</button><button :disabled="busyId != null || isRunning(plan)" @click="$emit('cleanup-request', plan, 'failed')">清理失败/跳过记录</button><button :disabled="busyId != null || isRunning(plan)" @click="$emit('cleanup-request', plan, 'expired')">按保留规则清理</button><button :disabled="busyId != null" @click="$emit('delete', plan.id)">删除</button></div></article>
  </section>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { listPelicanPlanModels, type PelicanPlan, type PelicanPlanInput } from '@/api/pelicanTests'

const props = withDefaults(defineProps<{ plans: PelicanPlan[]; groups: Array<{ id: number; name: string }>; loading: boolean; error?: string; busyId?: number | null; saveVersion?: number; pelicanEnabled?: boolean }>(), { pelicanEnabled: true })
const emit = defineEmits<{ run: [id: number]; resume: [id: number]; save: [id: number | null, input: PelicanPlanInput]; 'delete': [id: number]; 'cleanup-request': [plan: PelicanPlan, scope: 'failed' | 'expired'] }>()
const editingID = ref<number | null>(null)
const models = ref<string[]>([])
const modelsLoading = ref(false)
const modelError = ref('')
const modelSelectionNotice = ref('')
let modelsRequestVersion = 0
let modelsAbort: AbortController | undefined
let retainedModelForGroup: { groupID: number; modelID: string } | undefined
const reasoningEffortOptions = [{ value: '', label: '默认' }, { value: 'none', label: '无' }, { value: 'minimal', label: '最低' }, { value: 'low', label: '低' }, { value: 'medium', label: '中' }, { value: 'high', label: '高' }, { value: 'xhigh', label: '极高' }]
const defaults = (): PelicanPlanInput => ({ group_id: 0, model_id: '', interval_minutes: 60, enabled: false, max_results: 20, min_chars: 9366, daily_call_limit: 0, failure_pause_threshold: 0, retention_days: 0, reasoning_effort: '' })
const draft = reactive<PelicanPlanInput>(defaults())
const isCurrentModel = computed(() => models.value.includes(draft.model_id))
const editingPaused = computed(() => props.plans.find(plan => plan.id === editingID.value)?.pause_reason === 'consecutive_failures')
const retentionMode = ref('0')
const modelPlaceholder = computed(() => !draft.group_id ? '请先选择分组' : modelsLoading.value ? '加载中…' : modelError.value ? '模型加载失败' : models.value.length ? '选择已配置模型' : '没有可选模型')
function reset(): void { editingID.value = null; retainedModelForGroup = undefined; retentionMode.value = '0'; Object.assign(draft, defaults()); clearModels() }
function startEdit(plan: PelicanPlan): void {
  const sameGroup = draft.group_id === plan.group_id
  editingID.value = plan.id
  retainedModelForGroup = { groupID: plan.group_id, modelID: plan.model_id }
  Object.assign(draft, { group_id: plan.group_id, model_id: plan.model_id, interval_minutes: plan.interval_minutes, enabled: plan.pause_reason === 'consecutive_failures' ? false : plan.enabled, max_results: plan.max_results, min_chars: plan.min_chars, daily_call_limit: plan.daily_call_limit ?? 0, failure_pause_threshold: plan.failure_pause_threshold ?? 0, retention_days: plan.retention_days ?? 0, reasoning_effort: plan.reasoning_effort ?? '' })
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
function planTime(value?: string | null): string { const date = value ? new Date(value) : null; return date && !Number.isNaN(date.getTime()) ? date.toLocaleString('zh-CN') : '暂无' }
onBeforeUnmount(() => { modelsRequestVersion++; modelsAbort?.abort() })
</script>
<style scoped>
.pelican-plans {
  /* BaseDialog teleports to body, outside the console theme variable scope. */
  --console-border: #d8cec2;
  --console-surface: #faf9f5;
  --console-text: #141413;
  --console-muted: #6c6a64;
  color-scheme: light;
}
.dark .pelican-plans {
  --console-border: #334155;
  --console-surface: #111c2e;
  --console-text: #e2e8f0;
  --console-muted: #a5b4c8;
  color-scheme: dark;
}
.pelican-plans input:disabled, .pelican-plans select:disabled, .pelican-plans button:disabled { background: var(--console-surface); color: var(--console-muted); opacity: .65; cursor: not-allowed; }
.pelican-plans input::placeholder { color: var(--console-muted); }
.pelican-plans button:not(:disabled):hover { border-color: #818cf8; }
.pelican-plans :is(input, select, button):focus-visible { outline: 2px solid #818cf8; outline-offset: 2px; }
.dark .pelican-plans .error { color: #fca5a5; }
.pelican-plans { margin-bottom: 18px; padding: 16px; border: 1px solid var(--console-border,#ddd); border-radius: 12px; background: var(--console-surface); color: var(--console-text); }
.pelican-plans header p, .plan span, .model-status, .retention-note { color: var(--console-muted,#666); font-size: 12px; }
.plan-form, .plan { display: flex; align-items: center; gap: 8px; padding: 10px 0; border-top: 1px solid var(--console-border,#ddd); flex-wrap: wrap; }
.pelican-plans header { margin-bottom: 14px; }.pelican-plans header h2 { margin: 0 0 4px; font-size: 14px; font-weight: 600; }
.plan-form { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); align-items: start; gap: 14px 16px; padding: 16px 0; }
.plan-form label { display: flex; flex-direction: column; gap: 6px; min-width: 0; color: var(--console-muted); font-size: 12px; line-height: 18px; }
.field-title { display: flex; align-items: center; gap: 5px; min-height: 18px; }.field-help { position: relative; display: inline-grid; place-items: center; width: 14px; height: 14px; border: 1px solid var(--console-border,#ddd); border-radius: 50%; font-size: 10px; cursor: help; }.field-tooltip { display: none; position: absolute; bottom: calc(100% + 7px); right: 0; width: 190px; padding: 8px 10px; border: 1px solid var(--console-border,#ddd); border-radius: 6px; background: var(--console-surface,#fff); color: var(--console-text); box-shadow: 0 4px 12px #0002; font-size: 12px; line-height: 1.5; z-index: 2; }.field-help:hover .field-tooltip,.field-help:focus .field-tooltip { display: block; }
.form-actions { grid-column: 1 / -1; display: flex; justify-content: flex-end; gap: 8px; }.form-actions button { padding: 7px 16px; }.model-status,.plan-form .enabled { grid-column: 1 / -1; }
.plan-form label.enabled { flex-direction: row; align-items: center; }
.plan-form input, .plan-form select { width: 100%; max-width: 100%; min-width: 0; min-height: 36px; box-sizing: border-box; padding: 7px 9px; border: 1px solid var(--console-border,#ddd); border-radius: 6px; background: var(--console-surface,#fff); color: var(--console-text); }
.plan-form input[type=checkbox] { width: auto; min-height: auto; }
.plan div { margin-left: auto; display: flex; flex-wrap: wrap; gap: 6px; }
.plan > strong { overflow-wrap: anywhere; }.plan-timing, .model-status, .retention-note { flex-basis: 100%; }.retention-note { margin: 0; padding: 8px 0; border-top: 1px solid var(--console-border,#ddd); }.model-status { margin: 0; }.model-status button { margin-left: 6px; }.error { color: #dc2626; }button { padding: 5px 8px; border: 1px solid var(--console-border,#ddd); border-radius: 6px; background: var(--console-surface,#fff); color: var(--console-text); cursor: pointer; }button:disabled { opacity: .5; cursor: default; }
@media(max-width:640px) { .pelican-plans { padding: 12px; }.plan-form { grid-template-columns: repeat(2,minmax(0,1fr)); gap: 12px; }.plan div { width: 100%; margin-left: 0; justify-content: flex-end; } }
@media(max-width:420px) { .plan-form { grid-template-columns: 1fr; } }
</style>
