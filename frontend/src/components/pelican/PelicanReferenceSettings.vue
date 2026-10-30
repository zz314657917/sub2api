<template>
  <fieldset class="reference-settings" :disabled="disabled" aria-label="第三方参考评分设置">
    <legend>降智雷达 · 第三方参考评分</legend>
    <label class="reference-enabled"><input type="checkbox" :checked="modelValue.enabled" aria-label="启用第三方参考评分" @change="setEnabled" />在页面顶部展示参考评分</label>
    <p>按模型首次出现的顺序展示，各模型内保留档位顺序，最多 64 个组合。目录只列出来源支持的模型与档位，也可输入自定义模型。</p>
    <p v-if="loading" role="status">正在加载来源目录…</p>
    <p v-else-if="!catalog || catalog.status === 'unavailable'" role="status">来源目录暂不可用，已保存的组合会保留，仍可手动编辑。</p>
    <p v-else-if="catalog.status === 'stale'" role="status">来源目录更新延迟，当前显示上次成功同步的目录。</p>
    <datalist id="pelican-reference-models"><option v-for="model in catalog?.models || []" :key="model.model" :value="model.model">{{ model.display_name }}</option></datalist>
    <ol class="reference-selection-list">
      <li v-for="(selection, index) in modelValue.selections" :key="index">
        <span class="reference-order" aria-hidden="true">{{ index + 1 }}</span>
        <label class="reference-model-input">模型<input :value="selection.model" list="pelican-reference-models" maxlength="200" :aria-label="`参考模型 ${index + 1}`" placeholder="输入或选择模型" @input="setModel(index, $event)" /></label>
        <label class="reference-effort-input">思考档位<select :value="selection.effort" :aria-label="`参考档位 ${index + 1}`" @change="setEffort(index, $event)">
          <option v-for="effort in effortOptions(selection)" :key="effort" :value="effort">{{ effort }}{{ unsupported(selection, effort) ? '（来源未支持）' : '' }}</option>
        </select></label>
        <div class="reference-row-actions">
          <button type="button" :disabled="Boolean(batchReason(selection))" :title="batchReason(selection) || '按来源顺序添加此模型缺少的全部档位'" :aria-label="`添加参考模型 ${index + 1} 的全部档位`" @click="addAll(selection)">添加全部档位</button>
          <button type="button" :disabled="disabled || index === 0" :aria-label="`上移参考组合 ${index + 1}`" @click="move(index, -1)">↑</button>
          <button type="button" :disabled="disabled || index === modelValue.selections.length - 1" :aria-label="`下移参考组合 ${index + 1}`" @click="move(index, 1)">↓</button>
          <button type="button" :aria-label="`移除参考组合 ${index + 1}`" @click="remove(index)">移除</button>
        </div>
        <span v-if="batchReason(selection) === capacityReason" class="reference-capacity" role="status">{{ capacityReason }}</span>
      </li>
    </ol>
    <p v-if="!modelValue.selections.length">尚未选择参考组合。</p>
    <p v-if="validationError" class="reference-validation" role="alert">{{ validationError }}</p>
    <button class="reference-add" type="button" :disabled="disabled || modelValue.selections.length >= selectionLimit" @click="add">添加参考组合</button>
  </fieldset>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import type { PelicanReferenceCatalog, PelicanReferenceSelection, PelicanReferenceSettings } from '@/api/pelicanTests'

const props = withDefaults(defineProps<{ modelValue: PelicanReferenceSettings; catalog?: PelicanReferenceCatalog | null; loading?: boolean; disabled?: boolean }>(), { catalog: null, loading: false, disabled: false })
const emit = defineEmits<{ 'update:modelValue': [value: PelicanReferenceSettings]; 'validity-change': [valid: boolean] }>()
const knownEfforts = ['low', 'medium', 'high', 'xhigh', 'max', 'ultra']
const selectionLimit = 64
const capacityReason = '添加全部档位后将超过 64 个参考组合，请先移除部分组合。'
function sourceModel(model: string) { return props.catalog?.models.find(item => item.model === model || item.display_name === model) }
function missingEfforts(selection: PelicanReferenceSelection): PelicanReferenceSelection[] {
  const model = sourceModel(selection.model)
  if (!model) return []
  const selected = new Set(props.modelValue.selections.filter(item => (sourceModel(item.model)?.model || item.model) === model.model).map(item => item.effort))
  return model.efforts.filter(effort => {
    if (selected.has(effort)) return false
    selected.add(effort)
    return true
  }).map(effort => ({ model: model.model, effort }))
}
function batchReason(selection: PelicanReferenceSelection): string {
  if (props.disabled) return '设置已禁用。'
  if (props.loading || !props.catalog || props.catalog.status === 'unavailable') return '来源目录暂不可用。'
  if (!sourceModel(selection.model)) return '来源目录未收录此模型。'
  const missing = missingEfforts(selection)
  if (!missing.length) return '已包含来源支持的全部档位。'
  return props.modelValue.selections.length + missing.length > selectionLimit ? capacityReason : ''
}
function addAll(selection: PelicanReferenceSelection): void {
  if (batchReason(selection)) return
  update([...props.modelValue.selections.map(item => ({ ...item })), ...missingEfforts(selection)])
}
const validationError = computed(() => {
  if (props.modelValue.selections.length > selectionLimit) return '最多选择 64 个参考组合。'
  const identities = new Set<string>()
  for (const selection of props.modelValue.selections) {
    if (!selection.model || !knownEfforts.includes(selection.effort)) return '请填写模型名称，并选择有效的思考档位。'
    if (selection.model.trim() !== selection.model || new TextEncoder().encode(selection.model).length > 200 || Array.from(selection.model).some(character => { const code = character.codePointAt(0)!; return code <= 31 || (code >= 127 && code <= 159) })) return '模型名称不能有首尾空格或控制字符，且不能超过 200 字节。'
    const key = JSON.stringify([sourceModel(selection.model)?.model || selection.model, selection.effort])
    if (identities.has(key)) return '模型与档位组合不能重复。'
    identities.add(key)
  }
  return ''
})
watch(validationError, error => emit('validity-change', !error), { immediate: true })
function update(selections: PelicanReferenceSelection[]): void { emit('update:modelValue', { ...props.modelValue, selections }) }
function setEnabled(event: Event): void { emit('update:modelValue', { ...props.modelValue, enabled: (event.target as HTMLInputElement).checked }) }
function setModel(index: number, event: Event): void { update(props.modelValue.selections.map((item, i) => i === index ? { ...item, model: (event.target as HTMLInputElement).value } : { ...item })) }
function setEffort(index: number, event: Event): void { update(props.modelValue.selections.map((item, i) => i === index ? { ...item, effort: (event.target as HTMLSelectElement).value } : { ...item })) }
function effortOptions(selection: PelicanReferenceSelection): string[] {
  const efforts = sourceModel(selection.model)?.efforts || knownEfforts
  return efforts.includes(selection.effort) ? efforts : [selection.effort, ...efforts]
}
function unsupported(selection: PelicanReferenceSelection, effort: string): boolean { const model = sourceModel(selection.model); return Boolean(model && !model.efforts.includes(effort)) }
function add(): void {
  if (props.disabled || props.modelValue.selections.length >= selectionLimit) return
  for (const model of props.catalog?.models || []) {
    const effort = model.efforts.find(effort => !props.modelValue.selections.some(item => (sourceModel(item.model)?.model || item.model) === model.model && item.effort === effort))
    if (effort) { update([...props.modelValue.selections, { model: model.model, effort }]); return }
  }
  update([...props.modelValue.selections, { model: '', effort: 'high' }])
}
function remove(index: number): void { update(props.modelValue.selections.filter((_, i) => i !== index)) }
function move(index: number, direction: number): void {
  const target = index + direction
  if (target < 0 || target >= props.modelValue.selections.length) return
  const selections = props.modelValue.selections.map(item => ({ ...item }))
  const current = selections[index]!
  selections[index] = selections[target]!
  selections[target] = current
  update(selections)
}
</script>

<style scoped>
.reference-settings { display: grid; gap: 9px; min-width: 0; margin: 3px 0; padding: 12px; border: 1px solid var(--pelican-border, #d8cec2); border-radius: 8px; color: var(--pelican-text, #141413); }
.reference-settings legend { padding: 0 5px; font-weight: 700; font-size: 12px; }
.reference-settings p { margin: 0; color: var(--pelican-muted, #6c6a64); font-size: 11px; line-height: 1.6; }
label.reference-enabled { display: flex; align-items: center; gap: 7px; font-weight: 500; }
.reference-enabled input { width: auto; accent-color: var(--pelican-accent, #a34e34); }
.reference-selection-list { display: grid; gap: 8px; margin: 0; padding: 0; list-style: none; min-width: 0; }
.reference-selection-list li { display: flex; flex-wrap: wrap; align-items: flex-end; gap: 7px; min-width: 0; }
.reference-capacity { flex-basis: 100%; color: var(--pelican-muted, #6c6a64); font-size: 11px; line-height: 1.6; }
.reference-order { align-self: center; flex-shrink: 0; width: 16px; color: var(--pelican-muted, #6c6a64); font-variant-numeric: tabular-nums; }
.reference-model-input { flex: 1 1 160px; min-width: 0; }
.reference-effort-input { flex: 0 1 160px; min-width: 90px; }
.reference-selection-list label { display: grid; gap: 3px; font-size: 11px; }
.reference-selection-list input, .reference-selection-list select { box-sizing: border-box; width: 100%; min-width: 0; min-height: 33px; padding: 6px 7px; border: 1px solid var(--pelican-border, #d8cec2); border-radius: 6px; background: var(--pelican-input, #fffdf9); color: var(--pelican-text, #141413); font: inherit; }
.reference-row-actions { display: flex; gap: 4px; flex-shrink: 0; }
.reference-settings button { min-height: 33px; border: 1px solid var(--pelican-border, #d8cec2); border-radius: 6px; background: var(--pelican-input, #fffdf9); color: var(--pelican-text, #141413); padding: 6px 8px; font: inherit; cursor: pointer; }
.reference-add { justify-self: start; }
.reference-settings button:disabled { opacity: .45; cursor: default; }
.reference-settings .reference-validation { color: #b42318; }
:global(.dark) .reference-settings .reference-validation { color: #fca5a5; }
.reference-settings button:focus-visible, .reference-settings input:focus-visible, .reference-settings select:focus-visible { outline: 2px solid var(--pelican-accent, #a34e34); outline-offset: 2px; }
@media (max-width: 640px) { .reference-selection-list li { flex-wrap: wrap; align-items: center; padding: 8px; border: 1px solid var(--pelican-border, #d8cec2); border-radius: 7px; } .reference-model-input { flex-basis: calc(100% - 24px); } .reference-effort-input { flex: 1 1 110px; margin-left: 23px; } }
</style>
