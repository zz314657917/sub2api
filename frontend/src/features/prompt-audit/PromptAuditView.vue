<template>
  <AppLayout>
    <div class="mx-auto max-w-[1600px]" :class="activeTab === 'config' && draft ? 'pb-28' : 'pb-8'">
      <header class="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary-600 dark:text-primary-400">{{ t('nav.securityAudit') }}</p>
          <h1 class="mt-1 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white">{{ t('admin.promptAudit.title') }}</h1>
          <p class="mt-2 max-w-3xl text-sm text-gray-500 dark:text-dark-300">{{ t('admin.promptAudit.description') }}</p>
        </div>
        <div v-if="draft" class="text-right text-xs text-gray-500 dark:text-dark-400">
          <p>{{ t('admin.promptAudit.configVersion', { version: draft.config_version }) }}</p>
          <p v-if="draft.updated_at" class="mt-1">{{ formatDate(draft.updated_at) }}</p>
        </div>
      </header>

      <div v-if="loadErrors.config && !draft" role="alert" class="rounded-xl border border-red-200 bg-red-50 p-5 dark:border-red-900 dark:bg-red-950/30">
        <p class="text-sm text-red-700 dark:text-red-300">{{ loadErrors.config }}</p>
        <button type="button" class="btn btn-secondary btn-sm mt-3" @click="loadConfig">{{ t('admin.promptAudit.actions.retry') }}</button>
      </div>

      <template v-else>
        <div class="mb-4" role="tablist" :aria-label="t('admin.promptAudit.title')">
          <div class="tabs inline-flex">
            <button
              v-for="tab in pageTabs"
              :key="tab.id"
              type="button"
              role="tab"
              class="tab"
              :class="{ 'tab-active': activeTab === tab.id }"
              :aria-selected="activeTab === tab.id"
              :data-test="`tab-${tab.id}`"
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </div>
        </div>

        <main class="card px-4 sm:px-6 lg:px-8">
          <div v-show="activeTab === 'config'" data-test="tab-panel-config">
            <RuntimeOverview :runtime="runtime" :loading="loading.runtime" :error="loadErrors.runtime" @refresh="loadRuntime" />

            <template v-if="draft">
              <EndpointPool
                :endpoints="draft.endpoints"
                :probe-results="probeResults"
                :probing-ids="probingIds"
                @update:endpoints="updateEndpoints"
                @probe="runProbe"
              />
              <div v-if="loadErrors.groups" role="alert" class="mt-5 rounded-lg bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:bg-amber-950/30 dark:text-amber-200">{{ loadErrors.groups }}</div>
              <PolicyPanel :draft="draft" :groups="groups" :rules="policyRules" @update:draft="replaceDraft" @update:rules="updatePolicyRules" @validation-change="policyInvalid = $event" />
              <section class="border-t border-gray-100 py-6 dark:border-dark-800" aria-labelledby="prompt-policy-lifecycle-title">
                <div class="flex flex-wrap items-start justify-between gap-4">
                  <div>
                    <h2 id="prompt-policy-lifecycle-title" class="text-base font-semibold text-gray-950 dark:text-white">{{ t('admin.promptAudit.policy.lifecycleTitle') }}</h2>
                    <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('admin.promptAudit.policy.lifecycleDescription') }}</p>
                  </div>
                  <div class="flex flex-wrap gap-2">
                    <button type="button" class="btn btn-secondary btn-sm" :disabled="loading.policy || policyInvalid" @click="previewPolicyDraft">{{ t('admin.promptAudit.policy.preview') }}</button>
                    <button type="button" class="btn btn-secondary btn-sm" :disabled="loading.policy || policyInvalid" @click="savePolicyDraft">{{ t('admin.promptAudit.policy.saveDraft') }}</button>
                    <button type="button" class="btn btn-primary btn-sm" :disabled="loading.policy || !policyHistory?.draft || policyDirty || policyConflict || policyBaseStale || policyInvalid" @click="publishPolicyDraft">{{ t('admin.promptAudit.policy.publish') }}</button>
                  </div>
                </div>
                <p v-if="policyConflict || policyBaseStale" role="alert" class="mt-3 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:bg-amber-950/30 dark:text-amber-200">{{ t('admin.promptAudit.policy.draftConflict') }}</p>
                <div v-if="policyPreview" class="mt-4 rounded-lg bg-gray-50 px-4 py-3 text-sm text-gray-700 dark:bg-dark-900/50 dark:text-dark-200">
                  {{ t('admin.promptAudit.policy.previewSummary', policyPreview) }}
                  <div v-if="policyPreview.examples?.length" class="mt-3 grid gap-2 sm:grid-cols-3" data-test="policy-preview-examples">
                    <div v-for="example in policyPreview.examples" :key="example.name" class="rounded-md border border-gray-200 bg-white px-3 py-2 dark:border-dark-700 dark:bg-dark-900/30">
                      <p class="font-medium">{{ policyExampleLabel(example.name) }}</p>
                      <p class="mt-1 text-xs">{{ example.current_action }} -> {{ example.candidate_action }} · {{ example.current_risk_level }} -> {{ example.candidate_risk_level }}</p>
                      <p v-if="example.matched_rule_id" class="mt-1 break-words text-xs text-primary-700 dark:text-primary-300">{{ example.matched_rule_id }}</p>
                    </div>
                  </div>
                </div>
                <div class="mt-4 flex flex-wrap items-end gap-3 rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-700/60">
                  <label class="min-w-52 flex-1 text-sm text-gray-700 dark:text-dark-200">
                    <span>{{ t('admin.promptAudit.policy.shadowSample') }}</span>
                    <select v-model="policyShadowSample" class="input mt-1.5 w-full" data-test="policy-shadow-sample" @change="invalidatePolicyResults">
                      <option value="safe">{{ t('admin.promptAudit.policy.shadowSamples.safe') }}</option>
                      <option value="controversial">{{ t('admin.promptAudit.policy.shadowSamples.controversial') }}</option>
                      <option value="violent">{{ t('admin.promptAudit.policy.shadowSamples.violent') }}</option>
                      <option value="unsafe">{{ t('admin.promptAudit.policy.shadowSamples.unsafe') }}</option>
                    </select>
                  </label>
                  <label class="min-w-36 flex-1 text-sm text-gray-700 dark:text-dark-200"><span>{{ t('admin.promptAudit.policy.shadowGroup') }}</span><input v-model="policyShadowContext.group_id" data-test="policy-shadow-group" class="input mt-1.5 w-full" inputmode="numeric" @input="invalidatePolicyResults" /></label>
                  <label class="min-w-36 flex-1 text-sm text-gray-700 dark:text-dark-200"><span>{{ t('admin.promptAudit.policy.shadowModel') }}</span><input v-model="policyShadowContext.model" data-test="policy-shadow-model" class="input mt-1.5 w-full" @input="invalidatePolicyResults" /></label>
                  <label class="min-w-36 flex-1 text-sm text-gray-700 dark:text-dark-200"><span>{{ t('admin.promptAudit.policy.shadowProvider') }}</span><input v-model="policyShadowContext.provider" data-test="policy-shadow-provider" class="input mt-1.5 w-full" @input="invalidatePolicyResults" /></label>
                  <button type="button" class="btn btn-secondary btn-sm" :disabled="loading.policy" data-test="policy-shadow" @click="runPolicyShadow">
                    {{ t('admin.promptAudit.policy.shadowRun') }}
                  </button>
                  <div v-if="policyShadow" class="basis-full text-sm text-gray-700 dark:text-dark-200" data-test="policy-shadow-result">
                    {{ t('admin.promptAudit.policy.shadowResult') }}: {{ policyShadow.current.action }} -> {{ policyShadow.candidate.action }} ·
                    {{ policyShadow.current.risk_level }} -> {{ policyShadow.candidate.risk_level }}<span v-if="policyShadow.candidate.matched_rule_id"> · {{ policyShadow.candidate.matched_rule_id }}</span>
                  </div>
                </div>
                <div v-if="policyHistory" class="mt-5 overflow-x-auto">
                  <table class="min-w-full text-left text-sm">
                    <thead class="text-xs uppercase text-gray-500 dark:text-dark-400"><tr><th class="px-2 py-2">{{ t('admin.promptAudit.policy.version') }}</th><th class="px-2 py-2">{{ t('admin.promptAudit.policy.createdAt') }}</th><th class="px-2 py-2">{{ t('admin.promptAudit.common.actions') }}</th></tr></thead>
                    <tbody><tr v-for="version in policyHistory.versions" :key="version.policy_version" class="border-t border-gray-100 dark:border-dark-800"><td class="px-2 py-2">v{{ version.policy_version }}<span v-if="version.policy_version === policyHistory.active_version" class="ml-2 text-xs text-primary-600">{{ t('admin.promptAudit.policy.active') }}</span></td><td class="px-2 py-2">{{ formatDate(version.created_at) }}</td><td class="px-2 py-2"><button type="button" class="btn btn-secondary btn-sm" :disabled="loading.policy || version.policy_version === policyHistory.active_version" @click="rollbackPolicy(version.policy_version)">{{ t('admin.promptAudit.policy.rollback') }}</button></td></tr></tbody>
                  </table>
                </div>
              </section>
            </template>
          </div>

          <div v-show="activeTab === 'events'" data-test="tab-panel-events">
            <div
              v-if="draft?.enabled && !draft.store_pass_events"
              data-test="pass-events-disabled-notice"
              role="status"
              class="mt-6 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-900/70 dark:bg-amber-950/30 dark:text-amber-200"
            >
              <span>{{ t('admin.promptAudit.events.passEventsDisabled') }}</span>
              <button type="button" class="btn btn-secondary btn-sm" @click="activeTab = 'config'">
                {{ t('admin.promptAudit.events.openConfiguration') }}
              </button>
            </div>
            <EventWorkspace
              :events="events.items"
              :total="events.total"
              :page="events.page"
              :page-size="events.page_size"
              :filters="filters"
              :selected-ids="selectedEventIds"
              :loading="loading.events"
              :error="loadErrors.events"
              @filters-change="handleFiltersChanged"
              @search="applyEventFilters"
              @selection="selectedEventIds = $event"
              @page="changePage"
              @page-size="changePageSize"
              @view="openEvent"
              @delete="requestSingleDelete"
              @batch-delete="requestBatchDelete"
              @preview-delete="requestFilterDeletePreview"
            />
          </div>
        </main>
      </template>
    </div>

    <div v-if="draft && activeTab === 'config'" class="fixed inset-x-0 bottom-0 z-30 border-t border-gray-200 bg-white/95 px-4 py-3 shadow-[0_-12px_35px_rgba(15,23,42,0.08)] backdrop-blur dark:border-dark-700/80 dark:bg-dark-900/95 dark:shadow-[0_-12px_35px_rgba(0,0,0,0.35)] lg:left-64">
      <div class="mx-auto flex max-w-[1600px] flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-x-5 gap-y-2">
          <SaveToggle :label="t('admin.promptAudit.saveBar.enabled')" :model-value="draft.enabled" data-test="enabled-toggle" @update:model-value="setEnabled" />
          <SaveToggle :label="t('admin.promptAudit.saveBar.blocking')" :model-value="draft.blocking_enabled" :disabled="!draft.enabled" data-test="blocking-toggle" @update:model-value="setBlocking" />
          <SaveToggle :label="t('admin.promptAudit.saveBar.blockingLatestTurnOnly')" :model-value="draft.blocking_latest_turn_only" :disabled="!draft.enabled || !draft.blocking_enabled" data-test="blocking-latest-turn-only-toggle" @update:model-value="replaceDraft({ ...draft!, blocking_latest_turn_only: $event })" />
          <SaveToggle :label="t('admin.promptAudit.saveBar.storePass')" :model-value="draft.store_pass_events" data-test="store-pass-toggle" @update:model-value="replaceDraft({ ...draft!, store_pass_events: $event })" />
        </div>
        <div class="flex items-center gap-3">
          <span class="text-sm" :class="dirty ? 'text-amber-700 dark:text-amber-300' : 'text-gray-500 dark:text-dark-400'">
            {{ dirty ? t('admin.promptAudit.saveBar.dirty') : t('admin.promptAudit.saveBar.synced') }}
          </span>
          <button type="button" class="btn btn-secondary" :disabled="!dirty || loading.saving" @click="resetDraft">{{ t('common.reset') }}</button>
          <button type="button" class="btn btn-primary" :disabled="!dirty || loading.saving" data-test="save-config" @click="saveConfig">
            {{ loading.saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>

    <ConfirmDialog
      :show="showBlockingConfirmation"
      :title="t('admin.promptAudit.blockingConfirm.title')"
      :message="t('admin.promptAudit.blockingConfirm.message')"
      :confirm-text="t('admin.promptAudit.blockingConfirm.confirm')"
      danger
      @confirm="confirmBlocking"
      @cancel="showBlockingConfirmation = false"
    />
    <ConfirmDialog
      :show="deleteRequest.mode !== ''"
      :title="t('admin.promptAudit.events.deleteConfirmTitle')"
      :message="t('admin.promptAudit.events.deleteConfirmMessage', { count: deleteRequest.ids.length })"
      :confirm-text="t('common.delete')"
      danger
      @confirm="confirmIDDelete"
      @cancel="clearDeleteRequest"
    />
    <FilterDeleteDialog
      :show="showFilterDelete"
      :initial-filters="filters"
      :preview="deletePreview"
      :previewing="loading.previewing"
      :deleting="loading.deleting"
      @close="closeFilterDelete"
      @preview="runFilterDeletePreview"
      @confirm="confirmFilterDelete"
      @criteria-change="clearDeletePreview"
    />
    <EventDetailDialog :show="showEventDetail" :event="activeEvent" :loading="loading.detail" @close="closeEventDetail" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useAppStore } from '@/stores/app'
import { extractApiErrorCode, extractApiErrorMessage } from '@/utils/apiError'
import RuntimeOverview from './components/RuntimeOverview.vue'
import EndpointPool from './components/EndpointPool.vue'
import PolicyPanel from './components/PolicyPanel.vue'
import EventWorkspace from './components/EventWorkspace.vue'
import EventDetailDialog from './components/EventDetailDialog.vue'
import FilterDeleteDialog from './components/FilterDeleteDialog.vue'
import promptAuditAPI from './api'
import type {
  PromptAuditDraft,
  PromptAuditEndpointDraft,
  PromptAuditEvent,
  PromptAuditGroup,
  PromptAuditRuntime,
  PromptDeletePreview,
  PromptEventFilters,
  PromptEventPage,
  PromptLoadErrors,
  PromptProbeResult,
  PromptPolicyHistory,
  PromptPolicyPreview,
  PromptPolicyShadowResult,
  PromptPolicyShadowSample,
  PromptPolicyMatchContext,
  PromptRiskActionRules,
} from './types'
import { buildUpdateRequest, cloneData, configToDraft, draftFingerprint, emptyEventFilters, policyRulesFingerprint } from './viewModel'

const { t, locale } = useI18n()
const appStore = useAppStore()
type PromptAuditPageTab = 'config' | 'events'
const activeTab = ref<PromptAuditPageTab>('events')
const pageTabs = computed(() => [
  { id: 'events' as const, label: t('admin.promptAudit.tabs.events') },
  { id: 'config' as const, label: t('admin.promptAudit.tabs.config') },
])
const serverConfig = ref<PromptAuditDraft | null>(null)
const draft = ref<PromptAuditDraft | null>(null)
const runtime = ref<PromptAuditRuntime | null>(null)
const groups = ref<PromptAuditGroup[]>([])
const policyRules = ref<PromptRiskActionRules>({})
const policyHistory = ref<PromptPolicyHistory | null>(null)
const policyPreview = ref<PromptPolicyPreview | null>(null)
const policyShadowSample = ref<PromptPolicyShadowSample>('controversial')
const policyShadow = ref<PromptPolicyShadowResult | null>(null)
const savedPolicyFingerprint = ref('')
const policyInvalid = ref(false)
const policyConflict = ref(false)
const policyRequestGeneration = ref(0)
const policyShadowContext = reactive({ group_id: '', model: '', provider: '' })
const events = reactive<PromptEventPage>({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
const filters = ref<PromptEventFilters>(emptyEventFilters())
const appliedFilters = ref<PromptEventFilters>(emptyEventFilters())
const selectedEventIds = ref<number[]>([])
const activeEvent = ref<PromptAuditEvent | null>(null)
const showEventDetail = ref(false)
const probeResults = reactive<Record<string, PromptProbeResult>>({})
const probingIds = ref<string[]>([])
const showFilterDelete = ref(false)
const deletePreview = ref<PromptDeletePreview | null>(null)
const deletePreviewFilters = ref<PromptEventFilters | null>(null)
const showBlockingConfirmation = ref(false)
const deleteRequest = reactive<{ mode: '' | 'single' | 'batch'; ids: number[] }>({ mode: '', ids: [] })
const loading = reactive({ config: false, runtime: false, groups: false, events: false, saving: false, detail: false, deleting: false, previewing: false, policy: false })
const loadErrors = reactive<PromptLoadErrors>({ config: '', runtime: '', groups: '', events: '' })
const dirty = computed(() => draftFingerprint(draft.value) !== draftFingerprint(serverConfig.value))
const policyDirty = computed(() => Boolean(policyHistory.value?.draft) && policyRulesFingerprint(policyRules.value) !== savedPolicyFingerprint.value)
const policyBaseStale = computed(() => Boolean(policyHistory.value?.draft && policyHistory.value.draft.base_config_version !== draft.value?.config_version))

const SaveToggle = defineComponent({
  inheritAttrs: false,
  props: { label: { type: String, required: true }, modelValue: { type: Boolean, required: true }, disabled: { type: Boolean, default: false } },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    return () => h('label', { class: ['flex items-center gap-2.5 text-sm', props.disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'] }, [
      h('button', {
        ...attrs,
        type: 'button',
        role: 'switch',
        'aria-checked': props.modelValue,
        'aria-label': props.label,
        disabled: props.disabled,
        class: [
          'relative inline-flex h-6 w-11 shrink-0 items-center rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2',
          props.modelValue ? 'bg-primary-600' : 'bg-gray-300 dark:bg-dark-600',
          props.disabled ? 'cursor-not-allowed' : 'cursor-pointer',
        ],
        onClick: (event: MouseEvent) => {
          event.preventDefault()
          if (!props.disabled) emit('update:modelValue', !props.modelValue)
        },
      }, [
        h('span', {
          class: [
            'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transition-transform duration-200 ease-in-out',
            props.modelValue ? 'translate-x-5' : 'translate-x-0',
          ],
        }),
      ]),
      h('span', { class: 'select-none text-gray-700 dark:text-dark-200' }, props.label),
    ])
  },
})

function errorMessage(error: unknown, fallbackKey: string): string {
  const code = extractApiErrorCode(error)
  if (code) {
    const key = `admin.promptAudit.errors.${code}`
    const translated = t(key)
    if (translated !== key) return translated
  }
  return extractApiErrorMessage(error, t(fallbackKey))
}

async function loadConfig() {
  loading.config = true
  loadErrors.config = ''
  try {
    const config = await promptAuditAPI.getConfig()
    serverConfig.value = configToDraft(config)
    draft.value = configToDraft(config)
    policyRules.value = cloneData(config.rules ?? {})
    invalidatePolicyResults()
  } catch (error) {
    loadErrors.config = errorMessage(error, 'admin.promptAudit.errors.loadConfig')
  } finally {
    loading.config = false
  }
}
async function loadPolicyHistory() {
  try { policyHistory.value = await promptAuditAPI.listPolicyVersions() } catch (error) { appStore.showError(errorMessage(error, 'admin.promptAudit.errors.loadPolicyHistory')) }
}
async function loadRuntime() {
  loading.runtime = true
  loadErrors.runtime = ''
  try { runtime.value = await promptAuditAPI.getRuntime() }
  catch (error) { loadErrors.runtime = errorMessage(error, 'admin.promptAudit.errors.loadRuntime') }
  finally { loading.runtime = false }
}
async function loadGroups() {
  loading.groups = true
  loadErrors.groups = ''
  try { groups.value = await promptAuditAPI.listGroups() }
  catch (error) { loadErrors.groups = errorMessage(error, 'admin.promptAudit.errors.loadGroups') }
  finally { loading.groups = false }
}
async function loadEvents() {
  loading.events = true
  loadErrors.events = ''
  try {
    const result = await promptAuditAPI.listEvents(appliedFilters.value, events.page, events.page_size)
    Object.assign(events, result)
    selectedEventIds.value = []
  } catch (error) {
    loadErrors.events = errorMessage(error, 'admin.promptAudit.errors.loadEvents')
  } finally {
    loading.events = false
  }
}
async function loadInitial() {
  await Promise.allSettled([loadConfig(), loadRuntime(), loadGroups(), loadEvents(), loadPolicyHistory()])
  reconcilePolicyState()
}

function reconcilePolicyState() {
  if (!draft.value) return
  const storedDraft = policyHistory.value?.draft
  if (storedDraft) {
    policyRules.value = cloneData(storedDraft.rules)
  } else {
    policyRules.value = cloneData(draft.value.rules ?? {})
  }
  savedPolicyFingerprint.value = policyRulesFingerprint(policyRules.value)
  policyConflict.value = Boolean(storedDraft && storedDraft.base_config_version !== draft.value.config_version)
  invalidatePolicyResults()
}

function updatePolicyRules(value: PromptRiskActionRules) {
  policyRules.value = cloneData(value)
  invalidatePolicyResults()
}
function invalidatePolicyResults() {
  policyRequestGeneration.value++
  policyPreview.value = null
  policyShadow.value = null
}

function policyShadowMatchContext(): PromptPolicyMatchContext {
  const group = policyShadowContext.group_id.trim()
  const groupID = group ? Number(group) : undefined
  return {
    ...(Number.isInteger(groupID) && groupID! > 0 ? { group_id: groupID } : {}),
    ...(policyShadowContext.model.trim() ? { model: policyShadowContext.model.trim() } : {}),
    ...(policyShadowContext.provider.trim() ? { provider: policyShadowContext.provider.trim() } : {}),
  }
}
function policyRequestIdentity(): string {
  return JSON.stringify({ rules: policyRules.value, sample: policyShadowSample.value, context: policyShadowMatchContext(), active: draft.value?.config_version ?? 0, generation: policyRequestGeneration.value })
}

async function previewPolicyDraft() {
  if (loading.policy || policyInvalid.value) return
  const identity = policyRequestIdentity()
  loading.policy = true
  try {
    const result = await promptAuditAPI.previewPolicy(cloneData(policyRules.value))
    if (identity === policyRequestIdentity()) policyPreview.value = result
  } catch (error) {
    if (identity === policyRequestIdentity()) appStore.showError(errorMessage(error, 'admin.promptAudit.errors.policy'))
  } finally { loading.policy = false }
}
function shadowSampleGuardOutput(sample: PromptPolicyShadowSample): string {
  if (sample === 'safe') return 'Safety: Safe\nCategories: None'
  if (sample === 'unsafe') return 'Safety: Unsafe\nCategories: PII'
  if (sample === 'violent') return 'Safety: Controversial\nCategories: Violent'
  return 'Safety: Controversial\nCategories: Jailbreak'
}
function policyExampleLabel(name: string): string {
  if (name === 'safe') return t('admin.promptAudit.policy.shadowSamples.safe')
  if (name === 'unsafe_pii') return t('admin.promptAudit.policy.shadowSamples.unsafe')
  if (name === 'controversial_violent') return t('admin.promptAudit.policy.shadowSamples.violent')
  return t('admin.promptAudit.policy.shadowSamples.controversial')
}
async function runPolicyShadow() {
  if (loading.policy || policyInvalid.value) return
  const identity = policyRequestIdentity()
  loading.policy = true
  try {
    const result = await promptAuditAPI.shadowPolicyGuardOutput(shadowSampleGuardOutput(policyShadowSample.value), cloneData(policyRules.value), policyShadowMatchContext())
    if (identity === policyRequestIdentity()) policyShadow.value = result
  } catch (error) {
    if (identity === policyRequestIdentity()) appStore.showError(errorMessage(error, 'admin.promptAudit.errors.policy'))
  } finally { loading.policy = false }
}
async function savePolicyDraft() {
  if (!draft.value || loading.policy || policyInvalid.value) return
  const submittedRules = cloneData(policyRules.value)
  const expectedConfigVersion = draft.value.config_version
  const expectedDraftVersion = policyHistory.value?.draft?.draft_version ?? 0
  loading.policy = true
  try {
    const history = await promptAuditAPI.savePolicyDraft(expectedConfigVersion, expectedDraftVersion, submittedRules)
    policyHistory.value = history
    savedPolicyFingerprint.value = policyRulesFingerprint(history.draft?.rules ?? submittedRules)
    policyConflict.value = false
    appStore.showSuccess(t('admin.promptAudit.messages.policyDraftSaved'))
  } catch (error) {
    policyConflict.value = extractApiErrorCode(error) === 'prompt_audit_config_conflict'
    appStore.showError(errorMessage(error, 'admin.promptAudit.errors.policy'))
  } finally { loading.policy = false }
}
async function publishPolicyDraft() {
  if (!draft.value || !policyHistory.value?.draft || policyDirty.value || policyConflict.value || policyBaseStale.value || policyInvalid.value || loading.policy) return
  const submittedRulesFingerprint = policyRulesFingerprint(policyRules.value)
  const expectedConfigVersion = draft.value.config_version
  const expectedDraftVersion = policyHistory.value.draft.draft_version
  loading.policy = true
  try {
    const saved = await promptAuditAPI.publishPolicyDraft(expectedConfigVersion, expectedDraftVersion)
    serverConfig.value = configToDraft(saved)
    if (draft.value) {
      draft.value = {
        ...draft.value,
        rules: cloneData(saved.rules ?? {}),
        config_version: saved.config_version,
        updated_at: saved.updated_at,
        updated_by: saved.updated_by,
        change_summary: saved.change_summary,
      }
    }
    if (policyRulesFingerprint(policyRules.value) === submittedRulesFingerprint) policyRules.value = cloneData(saved.rules ?? {})
    policyHistory.value = await promptAuditAPI.listPolicyVersions()
    savedPolicyFingerprint.value = policyRulesFingerprint(saved.rules ?? {})
    policyConflict.value = false
    invalidatePolicyResults()
    appStore.showSuccess(t('admin.promptAudit.messages.policyPublished'))
    await loadRuntime()
  } catch (error) { policyConflict.value = extractApiErrorCode(error) === 'prompt_audit_config_conflict'; appStore.showError(errorMessage(error, 'admin.promptAudit.errors.policy')) } finally { loading.policy = false }
}
async function rollbackPolicy(version: number) {
  if (!draft.value || loading.policy || !window.confirm(t('admin.promptAudit.policy.rollbackConfirm', { version }))) return
  const submittedRulesFingerprint = policyRulesFingerprint(policyRules.value)
  const expectedConfigVersion = draft.value.config_version
  loading.policy = true
  try {
    const saved = await promptAuditAPI.rollbackPolicy(version, expectedConfigVersion)
    serverConfig.value = configToDraft(saved)
    if (draft.value) {
      draft.value = {
        ...draft.value,
        rules: cloneData(saved.rules ?? {}),
        config_version: saved.config_version,
        updated_at: saved.updated_at,
        updated_by: saved.updated_by,
        change_summary: saved.change_summary,
      }
    }
    if (policyRulesFingerprint(policyRules.value) === submittedRulesFingerprint) policyRules.value = cloneData(saved.rules ?? {})
    policyHistory.value = await promptAuditAPI.listPolicyVersions()
    savedPolicyFingerprint.value = policyRulesFingerprint(saved.rules ?? {})
    policyConflict.value = false
    invalidatePolicyResults()
    appStore.showSuccess(t('admin.promptAudit.messages.policyRolledBack'))
    await loadRuntime()
  } catch (error) { policyConflict.value = extractApiErrorCode(error) === 'prompt_audit_config_conflict'; appStore.showError(errorMessage(error, 'admin.promptAudit.errors.policy')) } finally { loading.policy = false }
}

function replaceDraft(value: PromptAuditDraft) { draft.value = cloneData(value) }
function updateEndpoints(value: PromptAuditEndpointDraft[]) {
  if (!draft.value) return
  replaceDraft({ ...draft.value, endpoints: value })
}
function setEnabled(value: boolean) {
  if (!draft.value) return
  replaceDraft({ ...draft.value, enabled: value, blocking_enabled: value ? draft.value.blocking_enabled : false })
}
function setBlocking(value: boolean) {
  if (!draft.value || !draft.value.enabled) return
  if (value && !draft.value.blocking_enabled) { showBlockingConfirmation.value = true; return }
  replaceDraft({ ...draft.value, blocking_enabled: value })
}
function confirmBlocking() {
  showBlockingConfirmation.value = false
  if (draft.value) replaceDraft({ ...draft.value, blocking_enabled: true })
}
function resetDraft() {
  if (serverConfig.value) draft.value = cloneData(serverConfig.value)
}
async function saveConfig() {
  if (!draft.value || !dirty.value) return
  const submittedDraft = cloneData(draft.value)
  const submittedFingerprint = draftFingerprint(submittedDraft)
  loading.saving = true
  try {
    const saved = await promptAuditAPI.updateConfig(buildUpdateRequest(submittedDraft))
    serverConfig.value = configToDraft(saved)
    if (draftFingerprint(draft.value) === submittedFingerprint) draft.value = configToDraft(saved)
    else if (draft.value) draft.value = { ...draft.value, config_version: saved.config_version, updated_at: saved.updated_at, updated_by: saved.updated_by, change_summary: saved.change_summary }
    invalidatePolicyResults()
    appStore.showSuccess(t('admin.promptAudit.messages.saved'))
    await loadRuntime()
  } catch (error) {
    const code = extractApiErrorCode(error)
    appStore.showError(errorMessage(error, code === 'prompt_audit_config_conflict' ? 'admin.promptAudit.errors.prompt_audit_config_conflict' : 'admin.promptAudit.errors.saveConfig'))
  } finally {
    loading.saving = false
  }
}
async function runProbe(endpoint: PromptAuditEndpointDraft) {
  if (probingIds.value.includes(endpoint.id)) return
  probingIds.value = [...probingIds.value, endpoint.id]
  try {
    const result = await promptAuditAPI.probeEndpoint(endpoint)
    probeResults[endpoint.id] = result
    if (result.ok) appStore.showSuccess(t('admin.promptAudit.messages.probeSucceeded'))
    else appStore.showError(`${result.error_code || result.status}: ${result.message}`)
  } catch (error) {
    appStore.showError(errorMessage(error, 'admin.promptAudit.errors.probe'))
  } finally {
    probingIds.value = probingIds.value.filter((id) => id !== endpoint.id)
  }
}

function handleFiltersChanged(value: PromptEventFilters) {
  filters.value = cloneData(value)
  clearDeletePreview()
}
function applyEventFilters(value: PromptEventFilters) {
  filters.value = cloneData(value)
  appliedFilters.value = cloneData(value)
  events.page = 1
  clearDeletePreview()
  void loadEvents()
}
function changePage(value: number) { events.page = value; void loadEvents() }
function changePageSize(value: number) { events.page_size = value; events.page = 1; void loadEvents() }
async function openEvent(id: number) {
  showEventDetail.value = true
  loading.detail = true
  activeEvent.value = null
  try { activeEvent.value = await promptAuditAPI.getEvent(id) }
  catch (error) { appStore.showError(errorMessage(error, 'admin.promptAudit.errors.loadDetail')); showEventDetail.value = false }
  finally { loading.detail = false }
}
function closeEventDetail() { showEventDetail.value = false; activeEvent.value = null }
function requestSingleDelete(id: number) { deleteRequest.mode = 'single'; deleteRequest.ids = [id] }
function requestBatchDelete() { if (selectedEventIds.value.length) { deleteRequest.mode = 'batch'; deleteRequest.ids = [...selectedEventIds.value] } }
function clearDeleteRequest() { deleteRequest.mode = ''; deleteRequest.ids = [] }
async function confirmIDDelete() {
  const mode = deleteRequest.mode
  const ids = [...deleteRequest.ids]
  clearDeleteRequest()
  if (!mode || ids.length === 0) return
  loading.deleting = true
  try {
    const result = mode === 'single' ? await promptAuditAPI.deleteEvent(ids[0]) : await promptAuditAPI.batchDeleteEvents(ids)
    appStore.showSuccess(t('admin.promptAudit.messages.deleted', { count: result.deleted_events }))
    await Promise.allSettled([loadEvents(), loadRuntime()])
  } catch (error) { appStore.showError(errorMessage(error, 'admin.promptAudit.errors.delete')) }
  finally { loading.deleting = false }
}
function clearDeletePreview() {
  deletePreview.value = null
  deletePreviewFilters.value = null
}
function requestFilterDeletePreview() {
  clearDeletePreview()
  showFilterDelete.value = true
}
function closeFilterDelete() {
  showFilterDelete.value = false
  clearDeletePreview()
}
async function runFilterDeletePreview(value: PromptEventFilters) {
  loading.previewing = true
  try {
    deletePreview.value = await promptAuditAPI.previewDelete(value)
    deletePreviewFilters.value = cloneData(value)
  } catch (error) {
    clearDeletePreview()
    appStore.showError(errorMessage(error, 'admin.promptAudit.errors.previewDelete'))
  } finally { loading.previewing = false }
}
async function confirmFilterDelete(filters?: PromptEventFilters) {
  if (loading.deleting) return
  loading.deleting = true
  try {
    let preview = deletePreview.value
    let previewFilters = deletePreviewFilters.value ? cloneData(deletePreviewFilters.value) : null
    // One-click path: no fresh preview (never requested, or cleared by a
    // criteria change) — mint the confirmation token on the fly from the
    // criteria the dialog just emitted, then delete in the same action.
    if ((!preview || !previewFilters) && filters) {
      preview = await promptAuditAPI.previewDelete(filters)
      previewFilters = cloneData(filters)
    }
    if (!preview || !previewFilters) return
    const result = await promptAuditAPI.deleteEventsByFilter(previewFilters, preview)
    closeFilterDelete()
    appStore.showSuccess(t('admin.promptAudit.messages.deleted', { count: result.deleted_events }))
    await Promise.allSettled([loadEvents(), loadRuntime()])
  } catch (error) {
    clearDeletePreview()
    appStore.showError(errorMessage(error, 'admin.promptAudit.errors.deleteConfirmation'))
  } finally { loading.deleting = false }
}
function formatDate(value: string): string {
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value))
}

onMounted(loadInitial)
</script>
