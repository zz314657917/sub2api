<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="min-w-56 flex-1">
            <input
              v-model="searchQuery"
              type="text"
              class="input"
              placeholder="搜索教程标题、slug 或正文"
              @input="handleSearch"
            />
          </div>
          <Select v-model="filters.status" :options="statusFilterOptions" class="w-40" @change="loadTutorials" />
          <input
            v-model="filters.category"
            type="text"
            class="input w-40"
            placeholder="分组"
            @input="handleSearch"
          />
          <div class="flex flex-1 justify-end gap-2">
            <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadTutorials">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button type="button" class="btn btn-primary" @click="openCreateDialog">
              <Icon name="plus" size="md" class="mr-1" />
              新建教程
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="tutorials"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="sort_order"
          default-sort-order="asc"
          @sort="handleSort"
        >
          <template #cell-title="{ row }">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="truncate font-medium text-gray-900 dark:text-white">{{ row.title }}</span>
                <span class="badge badge-gray">{{ row.slug }}</span>
              </div>
              <p class="mt-1 line-clamp-2 text-xs text-gray-500 dark:text-dark-400">
                {{ row.description || '暂无描述' }}
              </p>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', value === 'published' ? 'badge-success' : 'badge-gray']">
              {{ value === 'published' ? '已发布' : '草稿' }}
            </span>
          </template>

          <template #cell-category="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-300">{{ value || '未分组' }}</span>
          </template>

          <template #cell-updated_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <router-link
                v-if="row.status === 'published'"
                :to="`/tutorial/${row.slug}`"
                target="_blank"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20"
                title="打开前台"
              >
                <Icon name="externalLink" size="sm" />
              </router-link>
              <button
                type="button"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600"
                title="编辑"
                :disabled="isRowBusy(row.id)"
                :class="isRowBusy(row.id) && 'opacity-50'"
                @click="openEditDialog(row)"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                type="button"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20"
                :title="row.status === 'published' ? '下线' : '发布'"
                :disabled="isRowBusy(row.id)"
                :class="isRowBusy(row.id) && 'opacity-50'"
                @click="requestStatusChange(row)"
              >
                <Icon :name="isRowBusy(row.id) ? 'refresh' : row.status === 'published' ? 'x' : 'check'" size="sm" :class="isRowBusy(row.id) ? 'animate-spin' : ''" />
              </button>
              <button
                type="button"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                title="删除"
                :disabled="isRowBusy(row.id)"
                :class="isRowBusy(row.id) && 'opacity-50'"
                @click="handleDelete(row)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <div v-if="loadError" class="tutorial-load-error">
              <Icon name="exclamationCircle" size="xl" />
              <div>
                <h3>教程加载失败</h3>
                <p>{{ loadError }}</p>
              </div>
              <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadTutorials">
                重试
              </button>
            </div>
            <EmptyState
              v-else
              title="暂无教程"
              description="创建第一篇教程后，前台 /tutorial 会自动读取已发布内容。"
              action-text="新建教程"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog :show="showEditDialog" :title="isEditing ? '编辑教程' : '新建教程'" width="full" :close-on-escape="!saving" @close="requestCloseEdit">
      <form id="tutorial-page-form" class="space-y-4" @submit.prevent="handleSave">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">标题</label>
            <input v-model="form.title" type="text" class="input" required />
          </div>
          <div>
            <label class="input-label">slug</label>
            <input v-model="form.slug" type="text" class="input" required pattern="[a-z0-9-]+" />
            <p class="input-hint">仅支持小写字母、数字和连字符，前台路径为 /tutorial/slug。</p>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div>
            <label class="input-label">分组</label>
            <input v-model="form.category" type="text" class="input" placeholder="工具配置" />
          </div>
          <div>
            <label class="input-label">排序</label>
            <input v-model.number="form.sort_order" type="number" class="input" />
          </div>
          <div>
            <label class="input-label">状态</label>
            <Select v-model="form.status" :options="statusOptions" />
          </div>
        </div>

        <div>
          <label class="input-label">描述</label>
          <textarea v-model="form.description" rows="2" class="input" placeholder="显示在目录和文章顶部的摘要"></textarea>
        </div>

        <div class="tutorial-editor-grid">
          <div>
            <label class="input-label">Markdown + 短代码</label>
            <textarea
              v-model="form.content_md"
              rows="24"
              class="input tutorial-editor"
              required
              spellcheck="false"
            ></textarea>
            <p class="input-hint">
              支持 [[command]]、[[screenshot]]、[[callout]]、[[link-button]]。图片先填外链或现有静态资源路径。
            </p>
          </div>
          <div>
            <label class="input-label">预览</label>
            <div class="tutorial-preview" v-html="previewHtml"></div>
          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-between gap-3">
          <button type="button" class="btn btn-secondary" @click="insertShortcodeSample">
            插入短代码示例
          </button>
          <div class="flex gap-3">
            <button type="button" class="btn btn-secondary" :disabled="saving" @click="requestCloseEdit">取消</button>
            <button type="submit" form="tutorial-page-form" :disabled="saving" class="btn btn-primary">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteDialog"
      title="删除教程"
      message="确认删除这篇教程？已发布教程删除后前台将无法访问。"
      confirm-text="删除"
      cancel-text="取消"
      danger
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />

    <ConfirmDialog
      :show="showStatusDialog"
      :title="statusDialogTitle"
      :message="statusDialogMessage"
      confirm-text="确认"
      cancel-text="取消"
      :danger="statusTarget?.status === 'published'"
      @confirm="confirmStatusChange"
      @cancel="cancelStatusChange"
    />

    <ConfirmDialog
      :show="showDiscardDialog"
      title="放弃未保存修改"
      message="当前教程还有未保存修改，关闭后这些内容会丢失。"
      confirm-text="放弃修改"
      cancel-text="继续编辑"
      danger
      @confirm="discardAndClose"
      @cancel="showDiscardDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import { renderTutorialMarkdown } from '@/utils/tutorialMarkdown'
import type { BasePaginationResponse, TutorialPageSummary, TutorialPageStatus } from '@/types'
import type { Column } from '@/components/common/types'

import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)
const tutorials = ref<TutorialPageSummary[]>([])
const loadError = ref('')
const searchQuery = ref('')
const showEditDialog = ref(false)
const showDeleteDialog = ref(false)
const showStatusDialog = ref(false)
const showDiscardDialog = ref(false)
const editingId = ref<number | null>(null)
const deleteTarget = ref<TutorialPageSummary | null>(null)
const statusTarget = ref<TutorialPageSummary | null>(null)
const rowBusyId = ref<number | null>(null)
const formSnapshot = ref('')

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const filters = reactive<{
  status: TutorialPageStatus | ''
  category: string
  sort_by: string
  sort_order: 'asc' | 'desc'
}>({
  status: '',
  category: '',
  sort_by: 'sort_order',
  sort_order: 'asc'
})

const form = reactive({
  slug: '',
  title: '',
  description: '',
  category: '',
  sort_order: 0,
  status: 'draft' as TutorialPageStatus,
  content_md: ''
})

const isEditing = computed(() => editingId.value !== null)
const previewHtml = computed(() => renderTutorialMarkdown(form.content_md).html)
const isFormDirty = computed(() => formSnapshot.value !== serializeForm())
const statusDialogTitle = computed(() => {
  if (!statusTarget.value) return '更新教程状态'
  return statusTarget.value.status === 'published' ? '下线教程' : '发布教程'
})
const statusDialogMessage = computed(() => {
  if (!statusTarget.value) return ''
  if (statusTarget.value.status === 'published') {
    return `确认下线「${statusTarget.value.title}」？下线后前台将无法访问这篇教程。`
  }
  return `确认发布「${statusTarget.value.title}」？发布后所有访客都可以访问这篇教程。`
})

const statusOptions = [
  { value: 'draft', label: '草稿' },
  { value: 'published', label: '已发布' }
]

const statusFilterOptions = [
  { value: '', label: '全部状态' },
  ...statusOptions
]

const columns = computed<Column[]>(() => [
  { key: 'title', label: '教程', sortable: true },
  { key: 'category', label: '分组', sortable: true },
  { key: 'sort_order', label: '排序', sortable: true },
  { key: 'status', label: '状态', sortable: true },
  { key: 'updated_at', label: '更新时间', sortable: true },
  { key: 'actions', label: '操作' }
])

function resetForm() {
  Object.assign(form, {
    slug: '',
    title: '',
    description: '',
    category: '工具配置',
    sort_order: 0,
    status: 'draft',
    content_md: '# 新教程\n\n这里写教程正文。'
  })
  captureFormSnapshot()
}

function serializeForm(): string {
  return JSON.stringify({
    slug: form.slug,
    title: form.title,
    description: form.description,
    category: form.category,
    sort_order: Number(form.sort_order || 0),
    status: form.status,
    content_md: form.content_md
  })
}

function captureFormSnapshot() {
  formSnapshot.value = serializeForm()
}

function isRowBusy(id: number): boolean {
  return rowBusyId.value === id
}

async function loadTutorials() {
  loading.value = true
  try {
    const result: BasePaginationResponse<TutorialPageSummary> = await adminAPI.tutorials.list(
      pagination.page,
      pagination.page_size,
      {
        status: filters.status,
        category: filters.category.trim(),
        search: searchQuery.value.trim(),
        sort_by: filters.sort_by,
        sort_order: filters.sort_order
      }
    )
    loadError.value = ''
    tutorials.value = result.items
    pagination.total = result.total
    pagination.page = result.page
    pagination.page_size = result.page_size
    pagination.pages = result.pages
  } catch (error: any) {
    loadError.value = error?.message || '加载教程失败'
    tutorials.value = []
    pagination.total = 0
    pagination.pages = 0
    appStore.showError(loadError.value)
  } finally {
    loading.value = false
  }
}

let searchTimer: number | null = null
function handleSearch() {
  if (searchTimer) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => {
    pagination.page = 1
    loadTutorials()
  }, 300)
}

function handleSort(key: string, order: 'asc' | 'desc') {
  filters.sort_by = key
  filters.sort_order = order
  loadTutorials()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadTutorials()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadTutorials()
}

function openCreateDialog() {
  editingId.value = null
  resetForm()
  showEditDialog.value = true
}

async function openEditDialog(row: TutorialPageSummary) {
  try {
    const item = await adminAPI.tutorials.getById(row.id)
    editingId.value = item.id
    Object.assign(form, {
      slug: item.slug,
      title: item.title,
      description: item.description,
      category: item.category,
      sort_order: item.sort_order,
      status: item.status,
      content_md: item.content_md
    })
    captureFormSnapshot()
    showEditDialog.value = true
  } catch (error: any) {
    appStore.showError(error?.message || '加载教程正文失败')
  }
}

function closeEdit() {
  showEditDialog.value = false
  editingId.value = null
  showDiscardDialog.value = false
  formSnapshot.value = ''
}

function requestCloseEdit() {
  if (saving.value) return
  if (isFormDirty.value) {
    showDiscardDialog.value = true
    return
  }
  closeEdit()
}

function discardAndClose() {
  closeEdit()
}

async function handleSave() {
  saving.value = true
  try {
    const payload = {
      slug: form.slug.trim(),
      title: form.title.trim(),
      description: form.description.trim(),
      category: form.category.trim(),
      sort_order: Number(form.sort_order || 0),
      status: form.status,
      content_md: form.content_md
    }
    if (editingId.value) {
      await adminAPI.tutorials.update(editingId.value, payload)
    } else {
      await adminAPI.tutorials.create(payload)
    }
    appStore.showSuccess('教程已保存')
    captureFormSnapshot()
    closeEdit()
    await loadTutorials()
  } catch (error: any) {
    appStore.showError(error?.message || '保存教程失败')
  } finally {
    saving.value = false
  }
}

function requestStatusChange(row: TutorialPageSummary) {
  statusTarget.value = row
  showStatusDialog.value = true
}

function cancelStatusChange() {
  showStatusDialog.value = false
  statusTarget.value = null
}

async function confirmStatusChange() {
  const row = statusTarget.value
  if (!row) return
  const nextStatus: TutorialPageStatus = row.status === 'published' ? 'draft' : 'published'
  rowBusyId.value = row.id
  try {
    await adminAPI.tutorials.updateStatus(row.id, nextStatus)
    appStore.showSuccess(nextStatus === 'published' ? '教程已发布' : '教程已下线')
    cancelStatusChange()
    await loadTutorials()
  } catch (error: any) {
    appStore.showError(error?.message || '状态更新失败')
  } finally {
    rowBusyId.value = null
  }
}

function handleDelete(row: TutorialPageSummary) {
  deleteTarget.value = row
  showDeleteDialog.value = true
}

function cancelDelete() {
  showDeleteDialog.value = false
  deleteTarget.value = null
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  try {
    await adminAPI.tutorials.delete(deleteTarget.value.id)
    appStore.showSuccess('教程已删除')
    cancelDelete()
    await loadTutorials()
  } catch (error: any) {
    appStore.showError(error?.message || '删除教程失败')
  }
}

function insertShortcodeSample() {
  form.content_md += `

[[callout type="tip" title="提示"]]
这里写提示内容。
[[/callout]]

[[command title="命令示例" lang="bash"]]
npm install
[[/command]]

[[screenshot src="/tutorial/example.png" alt="截图说明" caption="截图说明"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]
`
}

onMounted(loadTutorials)
</script>

<style scoped>
.tutorial-editor-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 1rem;
}

.tutorial-editor {
  min-height: 34rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  line-height: 1.6;
}

.tutorial-preview {
  min-height: 34rem;
  max-height: 40rem;
  overflow: auto;
  padding: 1rem;
  border: 1px solid rgb(229 231 235);
  border-radius: 8px;
  background: white;
  color: rgb(17 24 39);
}

.dark .tutorial-preview {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39);
  color: rgb(243 244 246);
}

.tutorial-preview :deep(h1),
.tutorial-preview :deep(h2),
.tutorial-preview :deep(h3) {
  margin: 1rem 0 0.5rem;
  font-weight: 800;
}

.tutorial-preview :deep(h1) {
  font-size: 1.7rem;
}

.tutorial-preview :deep(h2) {
  font-size: 1.25rem;
}

.tutorial-preview :deep(p),
.tutorial-preview :deep(li) {
  line-height: 1.75;
}

.tutorial-preview :deep(.tutorial-command-block) {
  margin: 1rem 0;
  overflow: hidden;
  border: 1px solid rgb(51 65 85);
  border-radius: 8px;
  background: rgb(15 23 42);
  color: rgb(248 250 252);
}

.tutorial-preview :deep(.command-block-header) {
  display: flex;
  justify-content: space-between;
  padding: 0.65rem 0.8rem;
  border-bottom: 1px solid rgb(51 65 85);
}

.tutorial-preview :deep(pre) {
  margin: 0;
  overflow-x: auto;
  padding: 0.9rem;
}

.tutorial-preview :deep(pre code) {
  color: rgb(248 250 252);
}

.tutorial-preview :deep(.tutorial-callout) {
  padding: 0.8rem;
  border: 1px solid rgb(187 247 208);
  border-radius: 8px;
  background: rgb(240 253 244);
}

.dark .tutorial-preview :deep(.tutorial-callout) {
  border-color: rgba(134, 239, 172, 0.28);
  background: rgba(22, 163, 74, 0.12);
}

.tutorial-preview :deep(.tutorial-screenshot-card) {
  overflow: hidden;
  border: 1px solid rgb(229 231 235);
  border-radius: 8px;
}

.tutorial-preview :deep(.tutorial-screenshot-card img) {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
}

.tutorial-preview :deep(.tutorial-screenshot-card figcaption) {
  padding: 0.5rem;
  color: rgb(107 114 128);
}

.tutorial-load-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2.5rem 1rem;
  text-align: center;
  color: rgb(107 114 128);
}

.tutorial-load-error h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  color: rgb(17 24 39);
}

.tutorial-load-error p {
  margin: 0.25rem 0 0;
  font-size: 0.875rem;
}

.dark .tutorial-load-error h3 {
  color: rgb(243 244 246);
}

@media (max-width: 960px) {
  .tutorial-editor-grid {
    grid-template-columns: 1fr;
  }
}
</style>
