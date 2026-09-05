<template>
  <section aria-labelledby="prompt-policy-title" class="py-6">
    <div>
      <h2 id="prompt-policy-title" class="text-base font-semibold text-gray-950 dark:text-white">{{ t('admin.promptAudit.policy.title') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('admin.promptAudit.policy.description') }}</p>
    </div>

    <div class="mt-5 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(260px,0.45fr)]">
      <div class="rounded-xl border border-gray-200 p-4 dark:border-dark-700/60 dark:bg-dark-900/20 sm:p-5">
        <fieldset>
          <legend class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.promptAudit.policy.scope') }}</legend>
          <div class="mt-3 flex flex-wrap gap-5 text-sm text-gray-700 dark:text-dark-200">
            <label class="flex items-center gap-2">
              <input type="radio" name="prompt-audit-scope" :checked="draft.all_groups" @change="patch({ all_groups: true, group_ids: [] })" />
              {{ t('admin.promptAudit.policy.allGroups') }}
            </label>
            <label class="flex items-center gap-2">
              <input type="radio" name="prompt-audit-scope" :checked="!draft.all_groups" @change="patch({ all_groups: false })" />
              {{ t('admin.promptAudit.policy.selectedGroups') }}
            </label>
          </div>
        </fieldset>

        <div v-if="!draft.all_groups" class="mt-4">
          <label class="block text-sm text-gray-700 dark:text-dark-200">
            <span>{{ t('admin.promptAudit.policy.searchGroups') }}</span>
            <input v-model="groupSearch" type="search" class="input mt-1.5 w-full" :aria-label="t('admin.promptAudit.policy.searchGroups')" />
          </label>
          <div class="mt-3 max-h-52 overflow-y-auto rounded-lg border border-gray-200 p-2 dark:border-dark-700">
            <label v-for="group in filteredGroups" :key="group.id" class="flex cursor-pointer items-center justify-between gap-3 rounded-md px-2 py-2 text-sm hover:bg-gray-50 dark:hover:bg-dark-800">
              <span class="flex items-center gap-2 text-gray-800 dark:text-dark-100">
                <input type="checkbox" :checked="draft.group_ids.includes(group.id)" @change="toggleGroup(group.id)" />
                {{ group.name }}
              </span>
              <span class="text-xs text-gray-500 dark:text-dark-400">{{ group.platform }} · {{ group.status }}</span>
            </label>
            <p v-if="filteredGroups.length === 0" class="px-2 py-4 text-center text-sm text-gray-500">{{ t('admin.promptAudit.policy.noGroups') }}</p>
          </div>
          <div v-if="missingGroupIds.length" class="mt-3 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:bg-amber-950/30 dark:text-amber-200">
            {{ t('admin.promptAudit.policy.missingGroups') }}: {{ missingGroupIds.join(', ') }}
          </div>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">{{ t('admin.promptAudit.policy.selectedCount', { count: draft.group_ids.length }) }}</p>
        </div>

        <fieldset class="mt-5 border-t border-gray-100 pt-5 dark:border-dark-800">
          <legend class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.promptAudit.policy.scanners') }}</legend>
          <div class="mt-3 grid gap-2 sm:grid-cols-2">
            <label v-for="scanner in SCANNER_CATALOG" :key="scanner.id" class="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-gray-700 hover:bg-gray-50 dark:text-dark-200 dark:hover:bg-dark-800">
              <input type="checkbox" :checked="draft.scanners.includes(scanner.id)" :aria-label="scannerLabel(scanner.id)" @change="toggleScanner(scanner.id)" />
              <span>{{ scannerLabel(scanner.id) }}</span>
            </label>
          </div>
        </fieldset>

        <fieldset class="mt-5 border-t border-gray-100 pt-5 dark:border-dark-800">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <legend class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.promptAudit.policy.rules') }}</legend>
            <button type="button" class="btn btn-secondary btn-sm" @click="addRule">{{ t('admin.promptAudit.policy.addRule') }}</button>
          </div>
          <div v-if="rulesList.length" class="mt-3 space-y-3">
            <div v-for="(rule, index) in rulesList" :key="rule.id || index" class="grid gap-2 rounded-lg border border-gray-200 p-3 dark:border-dark-700 sm:grid-cols-[minmax(120px,1fr)_90px_minmax(150px,1fr)_100px_auto]">
              <input :value="rule.id" class="input" :aria-label="`${t('admin.promptAudit.policy.ruleId')} ${index + 1}`" @input="updateRule(index, { id: ($event.target as HTMLInputElement).value })" />
              <input :value="rule.priority" type="number" class="input" :aria-label="`${t('admin.promptAudit.policy.priority')} ${index + 1}`" @input="updateRule(index, { priority: Number(($event.target as HTMLInputElement).value) })" />
              <input :value="(rule.safety || []).join(', ')" class="input" :placeholder="t('admin.promptAudit.policy.ruleSafetyHint')" :aria-label="`${t('admin.promptAudit.policy.ruleId')} safety ${index + 1}`" @input="updateRule(index, { safety: parseList(($event.target as HTMLInputElement).value) })" />
              <select :value="rule.categories || []" multiple class="input min-h-20" :aria-label="`${t('admin.promptAudit.policy.ruleCategory')} ${index + 1}`" @change="updateRule(index, { categories: selectedValues($event.target as HTMLSelectElement) })">
                <option v-for="scanner in SCANNER_CATALOG" :key="scanner.id" :value="scanner.id">{{ scannerLabel(scanner.id) }}</option>
              </select>
              <select :value="rule.action" class="input" :aria-label="`${t('admin.promptAudit.policy.ruleAction')} ${index + 1}`" @change="updateRule(index, { action: ($event.target as HTMLSelectElement).value, risk_level: ($event.target as HTMLSelectElement).value === 'Block' ? 'critical' : 'medium' })">
                <option value="Warn">Warn</option><option value="Block">Block</option>
              </select>
              <button type="button" class="btn btn-secondary btn-sm" :aria-label="`${t('admin.promptAudit.policy.removeRule')} ${index + 1}`" @click="removeRule(index)">{{ t('common.delete') }}</button>
              <input :value="(rule.owasp_tags || []).join(', ')" class="input sm:col-span-4" :placeholder="t('admin.promptAudit.policy.owaspHint')" :aria-label="`${t('admin.promptAudit.policy.owaspTags')} ${index + 1}`" @input="updateRule(index, { owasp_tags: ($event.target as HTMLInputElement).value.split(',').map((value) => value.trim()).filter(Boolean) })" />
              <div class="sm:col-span-2">
                <input :value="(rule.groups || []).join(', ')" class="input w-full" :class="{ 'border-red-500': invalidNumberFields.includes(`groups-${index}`) }" :placeholder="t('admin.promptAudit.policy.ruleGroupsHint')" :aria-label="`${t('admin.promptAudit.policy.ruleId')} groups ${index + 1}`" @input="updateRuleGroups(index, ($event.target as HTMLInputElement).value)" />
                <p v-if="invalidNumberFields.includes(`groups-${index}`)" class="mt-1 text-xs text-red-600 dark:text-red-300">{{ t('admin.promptAudit.policy.invalidGroupIds') }}</p>
              </div>
              <input :value="(rule.models || []).join(', ')" class="input sm:col-span-2" :placeholder="t('admin.promptAudit.policy.ruleModelsHint')" :aria-label="`${t('admin.promptAudit.policy.ruleId')} models ${index + 1}`" @input="updateRule(index, { models: parseList(($event.target as HTMLInputElement).value) })" />
              <input :value="(rule.providers || []).join(', ')" class="input sm:col-span-2" :placeholder="t('admin.promptAudit.policy.ruleProvidersHint')" :aria-label="`${t('admin.promptAudit.policy.ruleId')} providers ${index + 1}`" @input="updateRule(index, { providers: parseList(($event.target as HTMLInputElement).value) })" />
            </div>
          </div>
          <p v-else class="mt-3 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.promptAudit.policy.noRules') }}</p>
        </fieldset>

        <fieldset class="mt-5 border-t border-gray-100 pt-5 dark:border-dark-800">
          <legend class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.promptAudit.policy.decisions') }}</legend>
          <div class="mt-3 space-y-3">
            <div v-for="row in safetyRows" :key="row.id" class="flex flex-wrap items-center justify-between gap-3 rounded-md bg-gray-50 px-3 py-2 text-sm dark:bg-dark-900/50">
              <div>
                <p class="font-medium text-gray-800 dark:text-dark-100">{{ t(`admin.promptAudit.policy.safety.${row.id}`) }}</p>
                <p v-if="row.locked" class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.promptAudit.policy.locked') }}</p>
              </div>
              <select :value="safetyAction(row.id)" class="input min-w-28" :disabled="row.locked" :aria-label="t(`admin.promptAudit.policy.safety.${row.id}`)" @change="updateSafety(row.id, ($event.target as HTMLSelectElement).value)">
                <option v-for="action in row.actions" :key="action" :value="action">{{ action }}</option>
              </select>
            </div>
          </div>
        </fieldset>

        <fieldset class="mt-5 border-t border-gray-100 pt-5 dark:border-dark-800">
          <legend class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.promptAudit.policy.categoryActions') }}</legend>
          <div class="mt-3 grid gap-2 sm:grid-cols-2">
            <label v-for="scanner in SCANNER_CATALOG" :key="scanner.id" class="flex items-center justify-between gap-2 rounded-md border border-gray-100 px-2 py-1.5 text-sm dark:border-dark-800">
              <span class="truncate text-gray-700 dark:text-dark-200">{{ scannerLabel(scanner.id) }}</span>
              <select :value="categoryAction(scanner.id)" class="input w-auto min-w-24 py-1 text-xs" :aria-label="scannerLabel(scanner.id)" @change="updateCategory(scanner.id, ($event.target as HTMLSelectElement).value)">
                <option value="">{{ t('admin.promptAudit.policy.inherit') }}</option>
                <option value="Warn">Warn</option>
                <option value="Block">Block</option>
              </select>
            </label>
          </div>
        </fieldset>
      </div>

      <div class="space-y-4 rounded-xl border border-gray-200 p-4 dark:border-dark-700/60 dark:bg-dark-900/20 sm:p-5">
        <label class="block text-sm text-gray-700 dark:text-dark-200">
          <span>{{ t('admin.promptAudit.policy.workerCount') }}</span>
          <input :value="draft.worker_count" type="number" min="1" max="32" class="input mt-1.5 w-full" :aria-label="t('admin.promptAudit.policy.workerCount')" @input="patch({ worker_count: Number(($event.target as HTMLInputElement).value) })" />
        </label>
        <label class="block text-sm text-gray-700 dark:text-dark-200">
          <span>{{ t('admin.promptAudit.policy.queueCapacity') }}</span>
          <input :value="draft.queue_capacity" type="number" min="1" max="100000" class="input mt-1.5 w-full" :aria-label="t('admin.promptAudit.policy.queueCapacity')" @input="patch({ queue_capacity: Number(($event.target as HTMLInputElement).value) })" />
        </label>
        <div class="rounded-lg bg-gray-50 px-4 py-3 text-sm text-gray-600 dark:bg-dark-900/50 dark:text-dark-300">
          <p class="font-medium text-gray-800 dark:text-dark-100">{{ t('admin.promptAudit.policy.strategy') }}</p>
          <p class="mt-1">priority · {{ t('admin.promptAudit.policy.strategyHint') }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PromptAuditDraft, PromptAuditGroup, PromptRiskActionRules, PromptRiskPolicyRule } from '../types'
import { cloneData, SCANNER_CATALOG } from '../viewModel'

const props = withDefaults(defineProps<{ draft: PromptAuditDraft; groups: PromptAuditGroup[]; rules?: PromptRiskActionRules }>(), { rules: () => ({}) })
const emit = defineEmits<{ (event: 'update:draft', value: PromptAuditDraft): void; (event: 'update:rules', value: PromptRiskActionRules): void; (event: 'validation-change', value: boolean): void }>()
const { t } = useI18n()
const groupSearch = ref('')
const invalidNumberFields = ref<string[]>([])
const safetyRows = [
  { id: 'safe', actions: ['Allow'], locked: true },
  { id: 'controversial', actions: ['Warn', 'Block'], locked: false },
  { id: 'unsafe', actions: ['Block'], locked: true },
] as const
const rulesList = computed(() => props.rules.rules ?? [])

const filteredGroups = computed(() => {
  const query = groupSearch.value.trim().toLowerCase()
  if (!query) return props.groups
  return props.groups.filter((group) => `${group.name} ${group.id} ${group.platform}`.toLowerCase().includes(query))
})
const knownGroupIds = computed(() => new Set(props.groups.map((group) => group.id)))
const missingGroupIds = computed(() => props.draft.group_ids.filter((id) => !knownGroupIds.value.has(id)))

function patch(value: Partial<PromptAuditDraft>) {
  emit('update:draft', { ...cloneData(props.draft), ...value })
}
function policyPatch(value: Partial<PromptRiskActionRules>) {
  emit('update:rules', { ...cloneData(props.rules), ...value })
}
function safetyAction(id: string): string {
  const configured = props.rules.defaults?.[id]?.action || props.rules.safety?.[id]
  if (configured) return configured
  return id === 'safe' ? 'Allow' : id === 'unsafe' ? 'Block' : 'Warn'
}
function updateSafety(id: string, action: string) {
  const defaults = { ...(props.rules.defaults ?? {}) }
  defaults[id] = { ...(defaults[id] ?? {}), action, risk_level: id === 'safe' ? 'low' : action === 'Block' ? 'critical' : 'medium' }
  policyPatch({ defaults })
}
function categoryAction(id: string): string {
  return props.rules.categories?.[id] || ''
}
function updateCategory(id: string, action: string) {
  const categories = { ...(props.rules.categories ?? {}) }
  if (action) categories[id] = action
  else delete categories[id]
  policyPatch({ categories })
}
function updateRule(index: number, patchValue: Partial<PromptRiskPolicyRule>) {
  const rules = cloneData(rulesList.value)
  rules[index] = { ...rules[index], ...patchValue }
  policyPatch({ rules })
}
function parseList(value: string): string[] {
  return value.split(',').map((item) => item.trim()).filter(Boolean)
}
function updateRuleGroups(index: number, value: string) {
  const key = `groups-${index}`
  const values = parseList(value)
  const groups = values.map(Number)
  const invalid = groups.some((item) => !Number.isInteger(item) || item <= 0)
  invalidNumberFields.value = invalid
    ? [...new Set([...invalidNumberFields.value, key])]
    : invalidNumberFields.value.filter((item) => item !== key)
  emit('validation-change', invalidNumberFields.value.length > 0)
  if (!invalid) updateRule(index, { groups })
}
function selectedValues(select: HTMLSelectElement): string[] {
  return [...select.selectedOptions].map((option) => option.value)
}
function addRule() {
  policyPatch({ rules: [...rulesList.value, { id: `rule-${Date.now()}`, priority: 10, categories: ['jailbreak'], action: 'Block', risk_level: 'critical', owasp_tags: ['LLM01'] }] })
}
function removeRule(index: number) {
  policyPatch({ rules: rulesList.value.filter((_, current) => current !== index) })
  invalidNumberFields.value = invalidNumberFields.value.flatMap((key) => {
    const match = /^groups-(\d+)$/.exec(key)
    if (!match) return [key]
    const position = Number(match[1])
    if (position === index) return []
    return [`groups-${position > index ? position - 1 : position}`]
  })
  emit('validation-change', invalidNumberFields.value.length > 0)
}
function toggleGroup(id: number) {
  const selected = new Set(props.draft.group_ids)
  if (selected.has(id)) selected.delete(id)
  else selected.add(id)
  patch({ group_ids: [...selected].sort((a, b) => a - b) })
}
function toggleScanner(id: string) {
  const selected = new Set(props.draft.scanners)
  if (selected.has(id)) selected.delete(id)
  else selected.add(id)
  patch({ scanners: SCANNER_CATALOG.map((item) => item.id).filter((item) => selected.has(item)) })
}
function scannerLabel(id: string): string {
  return t(`admin.promptAudit.scanners.${id}`)
}
</script>
