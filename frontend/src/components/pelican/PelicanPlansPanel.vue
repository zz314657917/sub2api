<template>
  <section class="pelican-plans" :aria-busy="loading"><header><h2>鹈鹕测试计划</h2><p>手动运行会调用模型并可能产生费用。</p></header>
    <p v-if="error" class="error" role="alert">{{ error }}</p><p v-if="loading">加载中…</p>
    <p>新计划默认停用，保存后可编辑启用；定时运行也会调用模型并可能产生费用。</p>
    <form class="plan-form" @submit.prevent="save">
      <label>分组<select v-model.number="draft.group_id" required aria-label="分组"><option :value="0" disabled>选择分组</option><option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}</option></select></label>
      <label>模型 ID<input v-model.trim="draft.model_id" required maxlength="200" placeholder="例如 gpt-5.4" aria-label="模型 ID"></label>
      <label>间隔（分钟）<input v-model.number="draft.interval_minutes" required type="number" min="15" max="1440" aria-label="间隔分钟"></label>
      <label>保留结果数<input v-model.number="draft.max_results" required type="number" min="1" max="50" aria-label="保留结果数"></label>
      <label>最小字符数<input v-model.number="draft.min_chars" required type="number" min="100" max="50000" aria-label="最小字符数"></label>
      <label v-if="editingID" class="enabled"><input v-model="draft.enabled" type="checkbox"> 启用定时测试</label>
      <button :disabled="loading || busyId != null || !draft.group_id">{{ editingID ? '保存计划' : '新建计划' }}</button><button v-if="editingID" type="button" :disabled="busyId != null" @click="reset">取消</button>
    </form>
    <div v-if="!loading && !plans.length">尚无计划。</div><article v-for="plan in plans" :key="plan.id" class="plan"><strong>{{ plan.group_name }} · {{ plan.model_id }}</strong><span>{{ plan.interval_minutes }} 分钟 · 保留 {{ plan.max_results }} 条 · {{ plan.enabled ? '已启用' : '已停用' }}</span><small class="plan-timing">上次：{{ planTime(plan.last_run_at) }} · 下次：{{ plan.enabled ? planTime(plan.next_run_at) : '已停用' }}<b v-if="isRunning(plan)"> · 正在测试</b></small><div><button :disabled="busyId != null || isRunning(plan)" @click="$emit('run', plan.id)">立即运行</button><button :disabled="busyId != null" @click="startEdit(plan)">编辑</button><button :disabled="busyId != null" @click="$emit('delete', plan.id)">删除</button></div></article>
  </section>
</template>
<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import type { PelicanPlan, PelicanPlanInput } from '@/api/pelicanTests'

const props = defineProps<{ plans: PelicanPlan[]; groups: Array<{ id: number; name: string }>; loading: boolean; error?: string; busyId?: number | null; saveVersion?: number }>()
const emit = defineEmits<{ run: [id: number]; save: [id: number | null, input: PelicanPlanInput]; 'delete': [id: number] }>()
const editingID = ref<number | null>(null)
const defaults = (): PelicanPlanInput => ({ group_id: 0, model_id: '', interval_minutes: 60, enabled: false, max_results: 20, min_chars: 9366 })
const draft = reactive<PelicanPlanInput>(defaults())
function reset(): void { editingID.value = null; Object.assign(draft, defaults()) }
function startEdit(plan: PelicanPlan): void { editingID.value = plan.id; Object.assign(draft, { group_id: plan.group_id, model_id: plan.model_id, interval_minutes: plan.interval_minutes, enabled: plan.enabled, max_results: plan.max_results, min_chars: plan.min_chars }) }
watch(() => props.saveVersion, () => reset())
function save(): void { if (props.busyId != null || props.loading || !draft.group_id) return; emit('save', editingID.value, { ...draft, enabled: editingID.value ? draft.enabled : false }) }
function isRunning(plan: PelicanPlan): boolean { return Boolean(plan.running_until && new Date(plan.running_until).getTime() > Date.now()) }
function planTime(value?: string | null): string { const date = value ? new Date(value) : null; return date && !Number.isNaN(date.getTime()) ? date.toLocaleString('zh-CN') : '暂无' }
</script>
<style scoped>
.pelican-plans { margin-bottom: 18px; background: var(--console-surface); color: var(--console-text); }
.plan-form label { display: flex; flex-direction: column; gap: 5px; flex: 1 1 130px; min-width: 0; color: var(--console-muted); font-size: 11px; }
.plan-form label.enabled { flex-direction: row; align-items: center; }
.plan-form input, .plan-form select { width: 100%; max-width: 100%; color: var(--console-text); }
.plan-form input[type=checkbox] { width: auto; }
button:disabled { opacity: .5; cursor: default; }
button { color: var(--console-text); cursor: pointer; }
.plan > strong { overflow-wrap: anywhere; }
.plan-timing { flex-basis: 100%; color: var(--console-muted); }
</style>
<style scoped>.pelican-plans{padding:16px;border:1px solid var(--console-border,#ddd);border-radius:12px}.pelican-plans header p,.plan span{color:var(--console-muted,#666);font-size:12px}.plan-form,.plan{display:flex;align-items:center;gap:8px;padding:10px 0;border-top:1px solid var(--console-border,#ddd);flex-wrap:wrap}.plan-form input:not([type=checkbox]),.plan-form select{min-width:90px;padding:5px;border:1px solid var(--console-border,#ddd);border-radius:6px;background:var(--console-surface,#fff)}.plan div{margin-left:auto;display:flex;gap:6px}.error{color:#dc2626}button{padding:5px 8px;border:1px solid var(--console-border,#ddd);border-radius:6px;background:var(--console-surface,#fff)}</style>
