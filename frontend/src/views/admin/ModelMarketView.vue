<template>
  <AppLayout>
    <div class="model-market-admin space-y-5">
      <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">模型市场</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              这里维护前台 /models 展示目录。价格只是公开展示数据，不影响真实渠道扣费。
            </p>
          </div>
          <div class="flex flex-wrap justify-end gap-2">
            <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadCatalog">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              刷新
            </button>
            <router-link to="/models" target="_blank" class="btn btn-secondary">
              <Icon name="externalLink" size="md" />
              打开前台
            </router-link>
            <button type="button" class="btn btn-secondary" @click="openJsonPanel">
              <Icon name="document" size="md" />
              JSON 工具
            </button>
            <button type="button" class="btn btn-primary" :disabled="saving || loading" @click="saveCatalog">
              <Icon name="check" size="md" />
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>

      <div class="grid items-start gap-5 xl:grid-cols-[minmax(18rem,22rem)_minmax(0,1fr)]">
        <section class="model-market-group-panel rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <h3 class="font-semibold text-gray-900 dark:text-white">分组</h3>
            <button type="button" class="btn btn-secondary btn-sm" @click="addGroup">
              <Icon name="plus" size="sm" />
              新增
            </button>
          </div>

          <VueDraggable
            v-if="catalog.groups.length > 0"
            v-model="catalog.groups"
            class="model-market-group-list mt-4 space-y-2"
            handle=".model-market-group-drag-handle"
            :animation="180"
            @start="rememberSelectedGroupBeforeSort"
            @end="handleGroupsReordered"
          >
            <div
              v-for="(group, index) in catalog.groups"
              :key="`${index}:${group.id}`"
              class="model-market-group-card"
              :class="{ 'is-active': selectedGroupIndex === index }"
            >
              <button
                type="button"
                class="model-market-group-drag-handle"
                title="拖动调整顺序"
                aria-label="拖动调整顺序"
                @click.stop
              >
                <Icon name="menu" size="sm" />
              </button>
              <button
                type="button"
                class="model-market-group-select"
                @click="selectedGroupIndex = index"
              >
                <span class="min-w-0">
                  <strong>{{ group.title || '未命名分组' }}</strong>
                  <small>{{ categoryLabel(group.category) }} · {{ group.rows.length }} 行</small>
                </span>
                <span :class="['badge', group.enabled ? 'badge-success' : 'badge-gray']">
                  {{ group.enabled ? '显示' : '隐藏' }}
                </span>
              </button>
              <div class="model-market-group-order-actions">
                <button
                  type="button"
                  title="上移分组"
                  aria-label="上移分组"
                  :disabled="index === 0"
                  @click="moveGroup(index, -1)"
                >
                  <Icon name="chevronUp" size="xs" />
                </button>
                <button
                  type="button"
                  title="下移分组"
                  aria-label="下移分组"
                  :disabled="index === catalog.groups.length - 1"
                  @click="moveGroup(index, 1)"
                >
                  <Icon name="chevronDown" size="xs" />
                </button>
              </div>
            </div>
          </VueDraggable>

          <EmptyState
            v-else
            title="暂无分组"
            description="新增一个分组后，前台模型广场会按分组展示价格。"
            action-text="新增分组"
            @action="addGroup"
          />
        </section>

        <section class="space-y-5">
          <div
            v-if="selectedGroup"
            class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <h3 class="font-semibold text-gray-900 dark:text-white">分组设置</h3>
              <button type="button" class="btn btn-danger btn-sm" @click="removeSelectedGroup">
                <Icon name="trash" size="sm" />
                删除分组
              </button>
            </div>

            <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
              <div>
                <label class="input-label">标题</label>
                <input v-model="selectedGroup.title" type="text" class="input" placeholder="ChatGPT" />
              </div>
              <div>
                <label class="input-label">ID</label>
                <input v-model="selectedGroup.id" type="text" class="input" placeholder="chatgpt" />
              </div>
              <div>
                <label class="input-label">类型</label>
                <Select v-model="selectedGroup.category" :options="categoryOptions" />
              </div>
              <div>
                <label class="input-label">平台</label>
                <input v-model="selectedGroup.platform" type="text" class="input" placeholder="openai / gemini / video" />
              </div>
              <div>
                <label class="input-label">价格倍率</label>
                <input
                  v-model.number="selectedGroup.price_multiplier"
                  type="number"
                  min="0.0001"
                  step="0.0001"
                  class="input"
                  placeholder="1"
                />
              </div>
              <label class="flex items-end gap-2 pb-2 text-sm font-medium text-gray-700 dark:text-dark-300">
                <input v-model="selectedGroup.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300" />
                前台显示
              </label>
              <label
                v-if="selectedGroup.category !== 'chat'"
                class="flex items-end gap-2 pb-2 text-sm font-medium text-gray-700 dark:text-dark-300"
              >
                <input v-model="selectedGroup.hide_official_price" type="checkbox" class="h-4 w-4 rounded border-gray-300" />
                隐藏官方价格列
              </label>
              <label
                v-if="selectedGroup.category !== 'chat'"
                class="flex items-end gap-2 pb-2 text-sm font-medium text-gray-700 dark:text-dark-300"
              >
                <input v-model="selectedGroup.hide_saving" type="checkbox" class="h-4 w-4 rounded border-gray-300" />
                隐藏节省列
              </label>
            </div>

            <div class="mt-4">
              <label class="input-label">描述</label>
              <textarea v-model="selectedGroup.description" rows="2" class="input" placeholder="显示在分组标题下方"></textarea>
            </div>

            <div class="mt-4">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <label class="input-label mb-0">支持的账号分组</label>
                <button type="button" class="btn btn-secondary btn-sm" :disabled="groupsLoading" @click="loadAccountGroups">
                  <Icon name="refresh" size="sm" :class="groupsLoading ? 'animate-spin' : ''" />
                  刷新分组
                </button>
              </div>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                前台“我们的价格”会先乘当前模型分组的价格倍率，再按这里选择的账号分组倍率换算。
              </p>
              <div v-if="accountGroups.length > 0" class="model-market-supported-groups mt-3">
                <label
                  v-for="group in accountGroups"
                  :key="group.id"
                  class="model-market-supported-group"
                >
                  <input
                    type="checkbox"
                    class="h-4 w-4 rounded border-gray-300"
                    :checked="selectedGroup.supported_group_ids?.includes(group.id) ?? false"
                    @change="toggleSupportedGroup(group.id, ($event.target as HTMLInputElement).checked)"
                  />
                  <span>
                    <strong>{{ group.name }}</strong>
                    <small>{{ group.platform }} · {{ formatRateLabel(group, selectedGroup.category) }}</small>
                  </span>
                </label>
              </div>
              <p v-else class="mt-3 text-sm text-gray-500 dark:text-dark-400">
                暂无可选账号分组。
              </p>
            </div>
          </div>

          <div
            v-if="selectedGroup"
            class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">价格行</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                  推理模型填写模型名称、输入、输出；图像和视频填写规格和我们的价格，官方价格与节省列可在分组设置里隐藏。
                </p>
              </div>
              <button type="button" class="btn btn-secondary btn-sm" @click="addRow">
                <Icon name="plus" size="sm" />
                新增行
              </button>
            </div>

            <div class="model-market-row-table-wrap mt-4">
              <table class="model-market-admin-table">
                <thead>
                  <tr>
                    <th class="w-36">ID</th>
                    <th v-if="selectedGroup.category === 'chat'" class="w-52">模型名称</th>
                    <th v-else class="w-64">规格</th>
                    <th v-if="selectedGroup.category === 'chat'" class="w-40">输入</th>
                    <th v-if="selectedGroup.category === 'chat'" class="w-40">输出</th>
                    <th class="w-48">我们的价格</th>
                    <th v-if="showOfficialPriceEditor" class="w-44">官方价格</th>
                    <th v-if="showSavingEditor" class="w-32">节省</th>
                    <th class="w-52">备注</th>
                    <th class="w-32">排序</th>
                    <th class="w-20">显示</th>
                    <th class="w-16">操作</th>
                  </tr>
                </thead>
                <VueDraggable
                  v-model="selectedGroup.rows"
                  tag="tbody"
                  handle=".model-market-row-drag-handle"
                  :animation="180"
                  @end="handleRowsReordered"
                >
                  <tr v-for="(row, index) in selectedGroup.rows" :key="row.id || index">
                    <td><input v-model="row.id" type="text" class="input input-sm" /></td>
                    <td v-if="selectedGroup.category === 'chat'">
                      <input v-model="row.model" type="text" class="input input-sm" placeholder="gpt-5.4" />
                    </td>
                    <td v-else>
                      <input v-model="row.spec" type="text" class="input input-sm" placeholder="2576x3216 · 中" />
                    </td>
                    <td v-if="selectedGroup.category === 'chat'">
                      <input v-model="row.input_price" type="text" class="input input-sm" placeholder="$2.5/M tokens" />
                    </td>
                    <td v-if="selectedGroup.category === 'chat'">
                      <input v-model="row.output_price" type="text" class="input input-sm" placeholder="$15/M tokens" />
                    </td>
                    <td>
                      <input v-model="row.our_price" type="text" class="input input-sm" placeholder="¥2.5/M 输入 · ¥15/M 输出" />
                    </td>
                    <td v-if="showOfficialPriceEditor">
                      <input v-model="row.official_price" type="text" class="input input-sm" placeholder="$0.1408" />
                    </td>
                    <td v-if="showSavingEditor">
                      <input v-model="row.saving" type="text" class="input input-sm" placeholder="20% ↓" />
                    </td>
                    <td><input v-model="row.note" type="text" class="input input-sm" /></td>
                    <td>
                      <div class="model-market-row-sort-cell">
                        <button
                          type="button"
                          class="model-market-row-drag-handle"
                          title="拖动排序"
                          aria-label="拖动排序"
                        >
                          <Icon name="menu" size="sm" />
                        </button>
                        <input v-model.number="row.sort_order" type="number" class="input input-sm" />
                      </div>
                    </td>
                    <td class="text-center">
                      <input v-model="row.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300" />
                    </td>
                    <td>
                      <button type="button" class="rounded-lg p-1.5 text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20" @click="removeRow(index)">
                        <Icon name="trash" size="sm" />
                      </button>
                    </td>
                  </tr>
                </VueDraggable>
              </table>
            </div>

            <EmptyState
              v-if="selectedGroup.rows.length === 0"
              title="暂无价格行"
              description="新增价格行后，前台会展示对应表格。"
              action-text="新增行"
              @action="addRow"
            />
          </div>
        </section>
      </div>

      <div
        v-if="showJsonPanel"
        class="model-market-json-overlay"
        @click.self="showJsonPanel = false"
      >
        <section class="model-market-json-panel">
          <div class="flex items-center justify-between gap-3">
            <h3 class="font-semibold text-gray-900 dark:text-white">JSON 导入 / 导出</h3>
            <button
              type="button"
              class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
              title="关闭"
              aria-label="关闭 JSON 工具"
              @click="showJsonPanel = false"
            >
              <Icon name="x" size="md" />
            </button>
          </div>

          <div class="mt-4 flex flex-wrap gap-2">
            <button type="button" class="btn btn-secondary" @click="syncJsonFromCatalog">
              <Icon name="copy" size="md" />
              同步当前
            </button>
            <button type="button" class="btn btn-secondary" @click="exportToJson">
              <Icon name="download" size="md" />
              复制 JSON
            </button>
            <button type="button" class="btn btn-secondary" @click="applyJsonToCatalog">
              <Icon name="upload" size="md" />
              应用 JSON
            </button>
            <button type="button" class="btn btn-danger" :disabled="saving || loading" @click="showResetDialog = true">
              重置默认
            </button>
          </div>

          <textarea v-model="jsonDraft" rows="18" class="input model-market-json mt-4" spellcheck="false"></textarea>
        </section>
      </div>
    </div>

    <ConfirmDialog
      :show="showResetDialog"
      title="重置模型市场"
      message="确认恢复默认模型市场目录？当前后台维护的展示数据会被默认数据覆盖。"
      confirm-text="重置"
      cancel-text="取消"
      danger
      @confirm="resetCatalog"
      @cancel="showResetDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { ModelMarketCatalog, ModelMarketCategory, ModelMarketGroup, ModelMarketPriceRow } from '@/api/modelMarket'
import type { AdminGroup } from '@/types'

import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { VueDraggable } from 'vue-draggable-plus'

const appStore = useAppStore()
const loading = ref(false)
const groupsLoading = ref(false)
const saving = ref(false)
const selectedGroupIndex = ref(0)
const jsonDraft = ref('')
const showResetDialog = ref(false)
const showJsonPanel = ref(false)
const accountGroups = ref<AdminGroup[]>([])
const selectedGroupBeforeSort = ref<ModelMarketGroup | null>(null)

const catalog = reactive<ModelMarketCatalog>({
  version: 1,
  groups: []
})

const categoryOptions = [
  { value: 'chat', label: '推理模型' },
  { value: 'image', label: '图像模型' },
  { value: 'video', label: '视频模型' }
]

const selectedGroup = computed<ModelMarketGroup | null>(() => {
  return catalog.groups[selectedGroupIndex.value] ?? catalog.groups[0] ?? null
})

const showOfficialPriceEditor = computed(() => {
  return selectedGroup.value?.category !== 'chat' && !selectedGroup.value?.hide_official_price
})

const showSavingEditor = computed(() => {
  return selectedGroup.value?.category !== 'chat' && !selectedGroup.value?.hide_saving
})

function categoryLabel(category: ModelMarketCategory | string): string {
  if (category === 'chat') return '推理'
  if (category === 'image') return '图像'
  if (category === 'video') return '视频'
  return '未知'
}

function createGroup(): ModelMarketGroup {
  const next = catalog.groups.length + 1
  return {
    id: `group-${Date.now()}`,
    title: '新模型分组',
    category: 'chat',
    platform: '',
    description: '',
    hide_official_price: false,
    hide_saving: false,
    price_multiplier: 1,
    supported_group_ids: [],
    sort_order: next * 100,
    enabled: true,
    rows: []
  }
}

function createRow(category: ModelMarketCategory): ModelMarketPriceRow {
  const group = selectedGroup.value
  const next = (group?.rows.length ?? 0) + 1
  return {
    id: `row-${Date.now()}`,
    model: category === 'chat' ? 'new-model' : '',
    spec: category === 'chat' ? '' : '默认',
    input_price: category === 'chat' ? '$0/M tokens' : '',
    output_price: category === 'chat' ? '$0/M tokens' : '',
    our_price: category === 'chat' ? '¥0/M 输入 · ¥0/M 输出' : '¥0',
    official_price: '',
    saving: '',
    note: '',
    sort_order: next * 100,
    enabled: true
  }
}

function replaceCatalog(next: ModelMarketCatalog): void {
  catalog.version = next.version || 1
  catalog.updated_at = next.updated_at
  catalog.groups.splice(0, catalog.groups.length, ...(next.groups ?? []))
  selectedGroupIndex.value = 0
  syncJsonFromCatalog()
}

async function loadAccountGroups(): Promise<void> {
  groupsLoading.value = true
  try {
    accountGroups.value = await adminAPI.groups.getAll()
  } catch (error: any) {
    appStore.showError(error?.message || '加载账号分组失败')
  } finally {
    groupsLoading.value = false
  }
}

async function loadCatalog(): Promise<void> {
  loading.value = true
  try {
    const [result] = await Promise.all([
      adminAPI.modelMarket.getCatalog(),
      accountGroups.value.length === 0 ? loadAccountGroups() : Promise.resolve()
    ])
    replaceCatalog(result)
  } catch (error: any) {
    appStore.showError(error?.message || '加载模型市场失败')
  } finally {
    loading.value = false
  }
}

async function saveCatalog(): Promise<void> {
  saving.value = true
  try {
    const result = await adminAPI.modelMarket.updateCatalog(toPlainCatalog())
    replaceCatalog(result)
    appStore.showSuccess('模型市场已保存')
  } catch (error: any) {
    appStore.showError(error?.message || '保存模型市场失败')
  } finally {
    saving.value = false
  }
}

async function resetCatalog(): Promise<void> {
  saving.value = true
  try {
    const result = await adminAPI.modelMarket.resetCatalog()
    replaceCatalog(result)
    appStore.showSuccess('模型市场已重置')
  } catch (error: any) {
    appStore.showError(error?.message || '重置模型市场失败')
  } finally {
    saving.value = false
    showResetDialog.value = false
  }
}

function toPlainCatalog(): ModelMarketCatalog {
  const plain = JSON.parse(JSON.stringify(catalog)) as ModelMarketCatalog
  plain.groups = plain.groups.map((group) => ({
    ...group,
    price_multiplier: normalizePriceMultiplier(group.price_multiplier)
  }))
  return plain
}

function syncJsonFromCatalog(): void {
  jsonDraft.value = JSON.stringify(toPlainCatalog(), null, 2)
}

function exportToJson(): void {
  syncJsonFromCatalog()
  void navigator.clipboard?.writeText(jsonDraft.value)
  appStore.showSuccess('JSON 已复制到剪贴板')
}

function openJsonPanel(): void {
  syncJsonFromCatalog()
  showJsonPanel.value = true
}

function applyJsonToCatalog(): void {
  try {
    const parsed = JSON.parse(jsonDraft.value) as ModelMarketCatalog
    if (!Array.isArray(parsed.groups)) {
      throw new Error('groups must be an array')
    }
    replaceCatalog(parsed)
    appStore.showSuccess('JSON 已应用，请保存后生效')
  } catch (error: any) {
    appStore.showError(error?.message || 'JSON 格式错误')
  }
}

function addGroup(): void {
  const group = createGroup()
  catalog.groups.push(group)
  selectedGroupIndex.value = catalog.groups.length - 1
  syncJsonFromCatalog()
}

function removeSelectedGroup(): void {
  if (!selectedGroup.value) return
  const index = selectedGroupIndex.value
  if (index >= 0) {
    catalog.groups.splice(index, 1)
    selectedGroupIndex.value = Math.min(index, Math.max(0, catalog.groups.length - 1))
    syncJsonFromCatalog()
  }
}

function rememberSelectedGroupBeforeSort(): void {
  selectedGroupBeforeSort.value = selectedGroup.value
}

function handleGroupsReordered(): void {
  normalizeGroupOrder()
  restoreSelectedGroup(selectedGroupBeforeSort.value)
  selectedGroupBeforeSort.value = null
  syncJsonFromCatalog()
}

function moveGroup(index: number, direction: -1 | 1): void {
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= catalog.groups.length) return
  const group = catalog.groups[index]
  const currentSelection = selectedGroup.value
  catalog.groups.splice(index, 1)
  catalog.groups.splice(targetIndex, 0, group)
  normalizeGroupOrder()
  restoreSelectedGroup(currentSelection)
  syncJsonFromCatalog()
}

function normalizeGroupOrder(): void {
  catalog.groups.forEach((group, index) => {
    group.sort_order = (index + 1) * 100
  })
}

function restoreSelectedGroup(group: ModelMarketGroup | null): void {
  if (!catalog.groups.length) {
    selectedGroupIndex.value = 0
    return
  }
  const index = group ? catalog.groups.indexOf(group) : -1
  selectedGroupIndex.value = index >= 0
    ? index
    : Math.min(selectedGroupIndex.value, catalog.groups.length - 1)
}

function addRow(): void {
  if (!selectedGroup.value) return
  selectedGroup.value.rows.unshift(createRow(selectedGroup.value.category))
  normalizeSelectedGroupRowOrder()
  syncJsonFromCatalog()
}

function removeRow(index: number): void {
  selectedGroup.value?.rows.splice(index, 1)
  normalizeSelectedGroupRowOrder()
  syncJsonFromCatalog()
}

function handleRowsReordered(): void {
  normalizeSelectedGroupRowOrder()
  syncJsonFromCatalog()
}

function normalizeSelectedGroupRowOrder(): void {
  selectedGroup.value?.rows.forEach((row, index) => {
    row.sort_order = (index + 1) * 100
  })
}

function toggleSupportedGroup(groupID: number, checked: boolean): void {
  if (!selectedGroup.value) return
  const ids = selectedGroup.value.supported_group_ids ?? []
  if (checked) {
    if (!ids.includes(groupID)) {
      selectedGroup.value.supported_group_ids = [...ids, groupID]
    }
  } else {
    selectedGroup.value.supported_group_ids = ids.filter((id) => id !== groupID)
  }
  syncJsonFromCatalog()
}

function formatRateLabel(group: AdminGroup, category: ModelMarketCategory): string {
  const rate = category === 'image' && group.image_rate_independent
    ? group.image_rate_multiplier
    : group.rate_multiplier
  const suffix = category === 'image' && group.image_rate_independent ? '图片倍率' : '倍率'
  return `${suffix} ${formatCompactRate(rate)}x`
}

function formatCompactRate(value: number | null | undefined): string {
  const rate = Number(value ?? 1)
  if (!Number.isFinite(rate)) return '1'
  return Number.parseFloat(rate.toFixed(6)).toString()
}

function normalizePriceMultiplier(value: unknown): number {
  if (value === '' || value === null || typeof value === 'undefined') return 1
  const rate = Number(value)
  return Number.isFinite(rate) && rate > 0 ? rate : 1
}

onMounted(loadCatalog)
</script>

<style scoped>
.model-market-group-panel {
  align-self: start;
  display: flex;
  height: calc(100vh - 8rem);
  max-height: calc(100vh - 8rem);
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
}

.model-market-group-list {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  border-radius: 8px;
  border: 1px solid rgb(229 231 235);
  padding: 0.5rem;
  scrollbar-gutter: stable;
}

.model-market-group-card {
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;
  gap: 0.5rem;
  border-radius: 8px;
  border: 1px solid rgb(229 231 235);
  padding: 0.45rem;
  text-align: left;
  transition:
    border-color 0.16s ease,
    background 0.16s ease;
}

.model-market-group-drag-handle,
.model-market-group-order-actions button {
  display: inline-flex;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  color: rgb(107 114 128);
  transition:
    background 0.16s ease,
    color 0.16s ease,
    opacity 0.16s ease;
}

.model-market-group-drag-handle {
  height: 2rem;
  width: 2rem;
  cursor: grab;
}

.model-market-group-drag-handle:hover,
.model-market-group-order-actions button:hover:not(:disabled) {
  background: rgb(243 244 246);
  color: rgb(55 65 81);
}

.model-market-group-drag-handle:active {
  cursor: grabbing;
}

.model-market-group-select {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
  padding: 0.35rem 0;
  text-align: left;
}

.model-market-group-order-actions {
  display: flex;
  flex: none;
  flex-direction: column;
  gap: 0.15rem;
}

.model-market-group-order-actions button {
  height: 1.45rem;
  width: 1.45rem;
}

.model-market-group-order-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.32;
}

.model-market-group-card strong,
.model-market-group-card small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-market-group-card strong {
  color: rgb(17 24 39);
  font-size: 0.9rem;
  line-height: 1.35;
}

.model-market-group-card small {
  margin-top: 0.15rem;
  color: rgb(107 114 128);
  font-size: 0.75rem;
}

.model-market-group-card.is-active {
  border-color: rgb(59 130 246);
  background: rgb(239 246 255);
}

.dark .model-market-group-card {
  border-color: rgb(55 65 81);
}

.dark .model-market-group-list {
  border-color: rgb(55 65 81);
}

.dark .model-market-group-card strong {
  color: rgb(243 244 246);
}

.dark .model-market-group-card small {
  color: rgb(156 163 175);
}

.dark .model-market-group-card.is-active {
  border-color: rgb(96 165 250);
  background: rgba(37, 99, 235, 0.16);
}

.dark .model-market-group-drag-handle,
.dark .model-market-group-order-actions button {
  color: rgb(156 163 175);
}

.dark .model-market-group-drag-handle:hover,
.dark .model-market-group-order-actions button:hover:not(:disabled) {
  background: rgb(55 65 81);
  color: rgb(229 231 235);
}

.model-market-supported-groups {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(13rem, 1fr));
  gap: 0.5rem;
  max-height: 14rem;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.model-market-supported-group {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  border-radius: 8px;
  border: 1px solid rgb(229 231 235);
  padding: 0.65rem;
  color: rgb(55 65 81);
  font-size: 0.82rem;
}

.model-market-supported-group strong,
.model-market-supported-group small {
  display: block;
}

.model-market-supported-group small {
  margin-top: 0.15rem;
  color: rgb(107 114 128);
}

.dark .model-market-supported-group {
  border-color: rgb(55 65 81);
  color: rgb(229 231 235);
}

.dark .model-market-supported-group small {
  color: rgb(156 163 175);
}

.model-market-admin-table {
  width: 100%;
  min-width: 76rem;
  border-collapse: collapse;
}

.model-market-row-table-wrap {
  max-height: min(32rem, calc(100vh - 22rem));
  overflow: auto;
  border-radius: 8px;
  border: 1px solid rgb(229 231 235);
}

.model-market-admin-table th,
.model-market-admin-table td {
  border-bottom: 1px solid rgb(229 231 235);
  padding: 0.5rem;
  text-align: left;
  vertical-align: middle;
}

.model-market-row-sort-cell {
  display: flex;
  min-width: 7rem;
  align-items: center;
  gap: 0.4rem;
}

.model-market-row-sort-cell .input {
  min-width: 0;
}

.model-market-row-drag-handle {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  flex: none;
  cursor: grab;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  color: rgb(107 114 128);
  transition:
    background 0.16s ease,
    color 0.16s ease;
}

.model-market-row-drag-handle:hover {
  background: rgb(243 244 246);
  color: rgb(55 65 81);
}

.model-market-row-drag-handle:active {
  cursor: grabbing;
}

.model-market-admin-table th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: rgb(249 250 251);
  color: rgb(107 114 128);
  font-size: 0.75rem;
  font-weight: 700;
}

.dark .model-market-row-table-wrap {
  border-color: rgb(55 65 81);
}

.dark .model-market-admin-table th {
  background: rgb(31 41 55);
}

.dark .model-market-admin-table th,
.dark .model-market-admin-table td {
  border-color: rgb(55 65 81);
}

.dark .model-market-row-drag-handle {
  color: rgb(156 163 175);
}

.dark .model-market-row-drag-handle:hover {
  background: rgb(55 65 81);
  color: rgb(229 231 235);
}

.model-market-json-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  justify-content: flex-end;
  background: rgb(15 23 42 / 0.28);
}

.model-market-json-panel {
  display: flex;
  height: 100%;
  width: min(44rem, 100vw);
  flex-direction: column;
  border-left: 1px solid rgb(229 231 235);
  background: white;
  padding: 1rem;
  box-shadow: -18px 0 40px rgb(15 23 42 / 0.18);
}

.dark .model-market-json-overlay {
  background: rgb(2 6 23 / 0.48);
}

.dark .model-market-json-panel {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39);
}

.model-market-group-list::-webkit-scrollbar,
.model-market-row-table-wrap::-webkit-scrollbar {
  width: 0.45rem;
  height: 0.45rem;
}

.model-market-group-list::-webkit-scrollbar-thumb,
.model-market-row-table-wrap::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgb(156 163 175 / 0.45);
}

.dark .model-market-group-list::-webkit-scrollbar-thumb,
.dark .model-market-row-table-wrap::-webkit-scrollbar-thumb {
  background: rgb(75 85 99 / 0.7);
}

.input-sm {
  min-height: 2rem;
  padding: 0.35rem 0.5rem;
  font-size: 0.82rem;
}

.model-market-json {
  min-height: 0;
  flex: 1;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  line-height: 1.55;
}
</style>
