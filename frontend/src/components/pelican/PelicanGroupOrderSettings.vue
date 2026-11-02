<template>
  <fieldset class="group-order-settings" :disabled="disabled">
    <legend>分组展示优先级</legend>
    <p>越靠前越优先，保存后对所有用户生效；未添加的分组按最近测试时间排列。</p>
    <div class="order-toolbar">
      <select v-model.number="selected" aria-label="添加优先分组" :disabled="loading || available.length === 0 || modelValue.length >= 1000">
        <option value="">{{ loading ? '加载分组…' : '选择分组' }}</option>
        <option v-for="group in available" :key="group.id" :value="group.id">{{ group.name }}</option>
      </select>
      <button type="button" :disabled="selected === '' || modelValue.length >= 1000" @click="add">添加</button>
      <button type="button" :disabled="!modelValue.length" @click="emit('update:modelValue', [])">恢复默认排序</button>
    </div>
    <p v-if="error" role="alert">分组列表加载失败，已保存的排序仍保留。<button type="button" @click="emit('retry')">重试</button></p>
    <ol v-if="modelValue.length" class="group-order-list">
      <li v-for="(id, index) in modelValue" :key="id">
        <span class="order-name"><small>{{ index + 1 }}</small>{{ nameOf(id) }}</span>
        <div class="order-actions">
          <button type="button" :aria-label="`上移 ${nameOf(id)}`" :disabled="index === 0" @click="move(index, -1)">↑</button>
          <button type="button" :aria-label="`下移 ${nameOf(id)}`" :disabled="index === modelValue.length - 1" @click="move(index, 1)">↓</button>
          <button type="button" :aria-label="`移除 ${nameOf(id)}`" @click="emit('update:modelValue', modelValue.filter(value => value !== id))">移除</button>
        </div>
      </li>
    </ol>
    <p v-else>当前使用默认排序。</p>
  </fieldset>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{
  modelValue: number[]
  groups: { id: number; name: string }[]
  disabled?: boolean
  loading?: boolean
  error?: boolean
}>()
const emit = defineEmits<{ 'update:modelValue': [value: number[]]; retry: [] }>()
const selected = ref<number | ''>('')
const available = computed(() => props.groups.filter(group => !props.modelValue.includes(group.id)))
const names = computed(() => new Map(props.groups.map(group => [group.id, group.name])))
function nameOf(id: number): string { return names.value.get(id) ?? `分组 #${id}（暂不可用）` }
function add(): void {
  if (props.disabled || selected.value === '' || props.modelValue.length >= 1000 || !available.value.some(group => group.id === selected.value)) return
  emit('update:modelValue', [...props.modelValue, selected.value])
  selected.value = ''
}
function move(index: number, delta: number): void {
  const next = [...props.modelValue]
  const target = index + delta
  if (target < 0 || target >= next.length) return
  ;[next[index], next[target]] = [next[target]!, next[index]!]
  emit('update:modelValue', next)
}
</script>

<style scoped>
.group-order-settings { min-width: 0; border: 1px solid var(--pelican-border); border-radius: 8px; padding: 10px 12px; }
legend { font-size: 13px; font-weight: 600; }
p { margin: 4px 0 8px; color: var(--pelican-muted); font-size: 12px; }
.order-toolbar, .order-actions { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
select, button { border: 1px solid var(--pelican-border); border-radius: 5px; background: var(--pelican-input); color: var(--pelican-text); padding: 5px 8px; font-size: 12px; }
select { min-width: 0; max-width: 100%; flex: 1 1 180px; }
button { cursor: pointer; }
button:disabled { opacity: .4; cursor: default; }
button:focus-visible, select:focus-visible { outline: 2px solid var(--pelican-accent); outline-offset: 2px; }
.group-order-list { list-style: none; padding: 0; margin: 8px 0 0; max-height: 220px; overflow: auto; }
li { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 5px 0; border-bottom: 1px solid var(--pelican-border); }
li:last-child { border-bottom: 0; }
.order-name { min-width: 0; overflow-wrap: anywhere; font-size: 12px; }
small { margin-right: 8px; color: var(--pelican-muted); font-variant-numeric: tabular-nums; }
.order-actions { flex-wrap: nowrap; flex-shrink: 0; }
</style>
