<template>
  <AppLayout>
    <div class="canvas-studio" data-testid="canvas-view">
      <aside class="canvas-panel canvas-list-panel">
        <div class="canvas-panel-header">
          <div class="min-w-0">
            <h1 class="truncate text-lg font-semibold text-gray-900 dark:text-white">{{ t('canvas.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('canvas.subtitle') }}</p>
          </div>
          <button
            type="button"
            class="btn btn-primary btn-sm"
            data-testid="canvas-new-button"
            @click="beginNewCanvas"
          >
            <Icon name="plus" size="sm" />
            <span>{{ t('canvas.newCanvas') }}</span>
          </button>
        </div>

        <div class="canvas-section">
          <div class="canvas-section-title">
            <span>{{ t('canvas.myCanvases') }}</span>
            <button
              type="button"
              class="canvas-icon-button"
              :title="t('common.refresh')"
              :disabled="loadingCanvases"
              @click="loadCanvases"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loadingCanvases }" />
            </button>
          </div>

          <div v-if="canvasLoadError" class="canvas-alert">
            <Icon name="exclamationTriangle" size="sm" />
            <span>{{ canvasLoadError }}</span>
          </div>

          <div class="canvas-list custom-scrollbar" data-testid="canvas-list">
            <button
              v-for="item in canvases"
              :key="item.id"
              type="button"
              class="canvas-list-item"
              :class="{ 'canvas-list-item-active': item.id === selectedCanvasId }"
              data-testid="canvas-open-button"
              @click="openCanvas(item.id)"
            >
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-semibold">{{ item.name }}</span>
                <span class="mt-1 block truncate text-xs text-gray-500 dark:text-dark-300">
                  {{ canvasMeta(item) }}
                </span>
              </span>
              <Icon name="chevronRight" size="sm" />
            </button>

            <div v-if="!loadingCanvases && canvases.length === 0" class="canvas-empty-list">
              <Icon name="inbox" size="lg" />
              <span>{{ t('canvas.emptyList') }}</span>
            </div>
          </div>
        </div>

        <div class="canvas-section">
          <div class="canvas-section-title">
            <span>{{ t('canvas.models') }}</span>
            <button
              type="button"
              class="canvas-icon-button"
              :title="t('common.refresh')"
              :disabled="loadingModels"
              @click="loadModels"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loadingModels }" />
            </button>
          </div>
          <select v-model="selectedModel" class="input text-sm" data-testid="canvas-model-select">
            <option value="">{{ t('canvas.defaultModel') }}</option>
            <option v-for="modelItem in models" :key="modelItem.id" :value="modelItem.id">
              {{ modelLabel(modelItem) }}
            </option>
          </select>
          <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-300">{{ t('canvas.modelHint') }}</p>
        </div>
      </aside>

      <main class="canvas-workspace">
        <header class="canvas-toolbar">
          <div class="min-w-0 flex-1">
            <input
              v-model="draftName"
              type="text"
              class="canvas-title-input"
              :placeholder="t('canvas.namePlaceholder')"
              data-testid="canvas-name-input"
            />
            <input
              v-model="draftDescription"
              type="text"
              class="canvas-description-input"
              :placeholder="t('canvas.descriptionPlaceholder')"
            />
          </div>
          <div class="canvas-toolbar-actions">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="!canQueueRun"
              data-testid="canvas-run-button"
              @click="queueCanvasRun"
            >
              <Icon name="play" size="sm" :class="{ 'animate-pulse': queuingRun }" />
              <span>{{ queuingRun ? t('canvas.queuing') : t('canvas.queueRun') }}</span>
            </button>
            <button
              type="button"
              class="btn btn-primary btn-sm"
              :disabled="!canSave"
              data-testid="canvas-save-button"
              @click="saveCanvas"
            >
              <Icon name="check" size="sm" />
              <span>{{ saving ? t('canvas.saving') : t('canvas.saveCanvas') }}</span>
            </button>
          </div>
        </header>

        <section class="canvas-stage-shell">
          <div class="canvas-stage-header">
            <span>{{ t('canvas.stage') }}</span>
            <span>{{ t('canvas.nodeCount', { count: canvasDocument.nodes.length }) }}</span>
          </div>
          <div class="canvas-stage custom-scrollbar" data-testid="canvas-stage">
            <svg class="canvas-edges" viewBox="0 0 980 620" preserveAspectRatio="none" aria-hidden="true">
              <path
                v-for="edge in edgeLines"
                :key="edge.id"
                :d="edge.path"
                class="canvas-edge"
              />
            </svg>
            <button
              v-for="node in canvasDocument.nodes"
              :key="node.id"
              type="button"
              class="canvas-node"
              :class="[nodeKindClass(node.type), { 'canvas-node-selected': node.id === selectedNodeId }]"
              :style="nodeStyle(node)"
              data-testid="canvas-node"
              @click="selectedNodeId = node.id"
            >
              <span class="canvas-node-kind">{{ nodeTypeLabel(node.type) }}</span>
              <span class="canvas-node-title">{{ node.title }}</span>
              <span class="canvas-node-status">{{ t(`canvas.nodeStatus.${node.status || 'idle'}`) }}</span>
            </button>

            <div v-if="canvasDocument.nodes.length === 0" class="canvas-stage-empty">
              <Icon name="cube" size="xl" />
              <span>{{ t('canvas.emptyStage') }}</span>
            </div>
          </div>
        </section>
      </main>

      <aside class="canvas-panel canvas-inspector-panel">
        <div class="canvas-section">
          <div class="canvas-section-title">
            <span>{{ t('canvas.nodeTypes') }}</span>
          </div>
          <div class="canvas-node-type-grid">
            <button
              v-for="item in nodeTypeItems"
              :key="item.type"
              type="button"
              class="canvas-node-type-button"
              :class="nodeKindClass(item.type)"
              :data-testid="`canvas-node-type-${item.type}`"
              @click="addNode(item.type)"
            >
              <Icon :name="item.icon" size="sm" />
              <span>{{ item.label }}</span>
            </button>
          </div>
        </div>

        <div class="canvas-section">
          <div class="canvas-section-title">
            <span>{{ t('canvas.nodeList') }}</span>
            <button
              type="button"
              class="canvas-icon-button"
              :title="t('canvas.removeNode')"
              :disabled="!selectedNode"
              data-testid="canvas-remove-node-button"
              @click="removeSelectedNode"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
          <div class="canvas-node-list custom-scrollbar" data-testid="canvas-node-list">
            <button
              v-for="node in canvasDocument.nodes"
              :key="node.id"
              type="button"
              class="canvas-node-list-item"
              :class="{ 'canvas-node-list-item-active': node.id === selectedNodeId }"
              @click="selectedNodeId = node.id"
            >
              <span class="canvas-node-list-dot" :class="nodeKindClass(node.type)"></span>
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-medium">{{ node.title }}</span>
                <span class="block truncate text-xs text-gray-500 dark:text-dark-300">{{ nodeTypeLabel(node.type) }}</span>
              </span>
            </button>
          </div>
        </div>

        <div class="canvas-section">
          <div class="canvas-section-title">
            <span>{{ t('canvas.runHistory') }}</span>
          </div>
          <div class="canvas-run-list custom-scrollbar" data-testid="canvas-run-list">
            <div v-for="run in runs" :key="run.id" class="canvas-run-item">
              <span class="canvas-run-status" :class="`canvas-run-status-${run.status}`"></span>
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-medium">{{ runStatusLabel(run.status) }}</span>
                <span class="block truncate text-xs text-gray-500 dark:text-dark-300">{{ formatDate(run.created_at) }}</span>
              </span>
            </div>
            <div v-if="runs.length === 0" class="canvas-placeholder">
              <Icon name="clock" size="md" />
              <span>{{ t('canvas.runPlaceholder') }}</span>
            </div>
          </div>
        </div>

        <div class="canvas-section">
          <div class="canvas-section-title">
            <span>{{ t('canvas.templates') }}</span>
          </div>
          <div class="canvas-template-entry">
            <Icon name="book" size="md" />
            <div class="min-w-0">
              <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ t('canvas.templateEntry') }}</p>
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-300">{{ t('canvas.templatePlaceholder') }}</p>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import {
  createCanvas,
  createCanvasRun,
  getCanvas,
  listCanvasModels,
  listCanvasRuns,
  listCanvases,
  updateCanvas,
  type CanvasDocument,
  type CanvasModel,
  type CanvasNode,
  type CanvasNodeType,
  type CanvasRun,
  type CanvasRunStatus,
  type UserCanvas,
  type UserCanvasSummary,
} from '@/api/canvas'

type IconName = InstanceType<typeof Icon>['$props']['name']

interface NodeTypeItem {
  type: CanvasNodeType
  label: string
  icon: IconName
}

interface EdgeLine {
  id: string
  path: string
}

const { t } = useI18n()
const appStore = useAppStore()

const canvases = ref<UserCanvasSummary[]>([])
const models = ref<CanvasModel[]>([])
const runs = ref<CanvasRun[]>([])
const selectedCanvasId = ref<string | null>(null)
const selectedNodeId = ref<string | null>(null)
const draftName = ref('')
const draftDescription = ref('')
const selectedModel = ref('')
const canvasDocument = ref<CanvasDocument>(createDefaultDocument())
const loadingCanvases = ref(false)
const loadingCanvas = ref(false)
const loadingModels = ref(false)
const saving = ref(false)
const queuingRun = ref(false)
const canvasLoadError = ref('')

const nodeTypes: Array<{ type: CanvasNodeType, icon: IconName }> = [
  { type: 'text', icon: 'document' },
  { type: 'image', icon: 'image' },
  { type: 'prompt', icon: 'chatBubble' },
  { type: 'loop', icon: 'sync' },
  { type: 'group', icon: 'folder' },
  { type: 'text_to_image', icon: 'sparkles' },
  { type: 'image_to_image', icon: 'swap' },
  { type: 'result', icon: 'checkCircle' },
]

const nodeTypeItems = computed<NodeTypeItem[]>(() =>
  nodeTypes.map((item) => ({
    ...item,
    label: nodeTypeLabel(item.type),
  }))
)

const selectedNode = computed(() =>
  canvasDocument.value.nodes.find((node) => node.id === selectedNodeId.value) ?? null
)

const canSave = computed(() =>
  !saving.value &&
  !loadingCanvas.value &&
  draftName.value.trim().length > 0 &&
  canvasDocument.value.nodes.length > 0
)

const canQueueRun = computed(() =>
  !queuingRun.value &&
  !saving.value &&
  !!selectedCanvasId.value &&
  canvasDocument.value.nodes.length > 0
)

const edgeLines = computed<EdgeLine[]>(() => {
  const nodeById = new Map(canvasDocument.value.nodes.map((node) => [node.id, node]))
  return canvasDocument.value.edges
    .map((edge) => {
      const source = nodeById.get(edge.source_node_id)
      const target = nodeById.get(edge.target_node_id)
      if (!source || !target) return null
      const sx = source.x + (source.width || 160)
      const sy = source.y + ((source.height || 82) / 2)
      const tx = target.x
      const ty = target.y + ((target.height || 82) / 2)
      const mid = Math.max(40, Math.abs(tx - sx) / 2)
      return {
        id: edge.id,
        path: `M ${sx} ${sy} C ${sx + mid} ${sy}, ${tx - mid} ${ty}, ${tx} ${ty}`,
      }
    })
    .filter((line): line is EdgeLine => line !== null)
})

onMounted(() => {
  void loadCanvases()
  void loadModels()
})

function createDefaultDocument(): CanvasDocument {
  const nodes: CanvasNode[] = [
    makeNode('prompt', 'canvas.sampleNodes.prompt', 70, 70),
    makeNode('text', 'canvas.sampleNodes.text', 290, 70),
    makeNode('text_to_image', 'canvas.sampleNodes.textToImage', 510, 70),
    makeNode('image', 'canvas.sampleNodes.image', 730, 70),
    makeNode('image_to_image', 'canvas.sampleNodes.imageToImage', 180, 250),
    makeNode('loop', 'canvas.sampleNodes.loop', 400, 250),
    makeNode('group', 'canvas.sampleNodes.group', 620, 250),
    makeNode('result', 'canvas.sampleNodes.result', 400, 430),
  ]
  return {
    nodes,
    edges: [
      makeEdge(nodes[0], nodes[1]),
      makeEdge(nodes[1], nodes[2]),
      makeEdge(nodes[2], nodes[3]),
      makeEdge(nodes[3], nodes[4]),
      makeEdge(nodes[4], nodes[5]),
      makeEdge(nodes[5], nodes[6]),
      makeEdge(nodes[6], nodes[7]),
    ],
    viewport: { x: 0, y: 0, zoom: 1 },
  }
}

function makeNode(type: CanvasNodeType, titleKey: string, x: number, y: number): CanvasNode {
  return {
    id: createId(type),
    type,
    title: t(titleKey),
    x,
    y,
    width: 170,
    height: 86,
    status: 'idle',
    config: {},
  }
}

function makeEdge(source: CanvasNode, target: CanvasNode) {
  return {
    id: createId('edge'),
    source_node_id: source.id,
    target_node_id: target.id,
  }
}

function createId(prefix: string): string {
  return `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
}

function beginNewCanvas(): void {
  selectedCanvasId.value = null
  draftName.value = t('canvas.untitledCanvas')
  draftDescription.value = ''
  selectedModel.value = models.value[0]?.id ?? ''
  canvasDocument.value = createDefaultDocument()
  selectedNodeId.value = canvasDocument.value.nodes[0]?.id ?? null
  runs.value = []
}

async function loadCanvases(): Promise<void> {
  loadingCanvases.value = true
  canvasLoadError.value = ''
  try {
    const response = await listCanvases({ limit: 30, offset: 0 })
    canvases.value = response.items
    if (response.items.length > 0 && !selectedCanvasId.value) {
      await openCanvas(response.items[0].id)
    } else if (!selectedCanvasId.value && !draftName.value) {
      beginNewCanvas()
    }
  } catch (error: unknown) {
    canvases.value = []
    canvasLoadError.value = errorMessage(error, t('canvas.loadCanvasesFailed'))
    if (!draftName.value) {
      beginNewCanvas()
    }
  } finally {
    loadingCanvases.value = false
  }
}

async function loadModels(): Promise<void> {
  loadingModels.value = true
  try {
    const response = await listCanvasModels()
    models.value = response.items
    if (!selectedModel.value) {
      selectedModel.value = response.items[0]?.id ?? ''
    }
  } catch {
    models.value = []
  } finally {
    loadingModels.value = false
  }
}

async function openCanvas(id: string): Promise<void> {
  loadingCanvas.value = true
  selectedCanvasId.value = id
  try {
    const item = await getCanvas(id)
    applyCanvas(item)
    await loadRuns(id)
  } catch (error: unknown) {
    appStore.showError(errorMessage(error, t('canvas.openFailed')))
  } finally {
    loadingCanvas.value = false
  }
}

async function loadRuns(canvasId: string): Promise<void> {
  try {
    const response = await listCanvasRuns({ canvas_id: canvasId, limit: 8, offset: 0 })
    runs.value = response.items
  } catch {
    runs.value = []
  }
}

async function saveCanvas(): Promise<void> {
  if (!canSave.value) return
  saving.value = true
  try {
    const payload = {
      name: draftName.value.trim(),
      description: draftDescription.value.trim() || undefined,
      model: selectedModel.value || undefined,
      document: canvasDocument.value,
    }
    const saved = selectedCanvasId.value
      ? await updateCanvas(selectedCanvasId.value, payload)
      : await createCanvas(payload)
    applyCanvas(saved)
    upsertCanvasSummary(saved)
    appStore.showSuccess(t('canvas.saveSuccess'))
  } catch (error: unknown) {
    appStore.showError(errorMessage(error, t('canvas.saveFailed')))
  } finally {
    saving.value = false
  }
}

async function queueCanvasRun(): Promise<void> {
  if (!selectedCanvasId.value || !canQueueRun.value) return
  queuingRun.value = true
  try {
    const run = await createCanvasRun({
      canvas_id: selectedCanvasId.value,
      model: selectedModel.value || undefined,
    })
    runs.value = [run, ...runs.value].slice(0, 8)
    appStore.showSuccess(t('canvas.runQueued'))
  } catch (error: unknown) {
    appStore.showError(errorMessage(error, t('canvas.queueFailed')))
  } finally {
    queuingRun.value = false
  }
}

function applyCanvas(item: UserCanvas): void {
  selectedCanvasId.value = item.id
  draftName.value = item.name
  draftDescription.value = item.description || ''
  selectedModel.value = item.model || selectedModel.value
  canvasDocument.value = normalizeDocument(item.document)
  selectedNodeId.value = canvasDocument.value.nodes[0]?.id ?? null
}

function normalizeDocument(document: CanvasDocument | null | undefined): CanvasDocument {
  if (!document || !Array.isArray(document.nodes) || !Array.isArray(document.edges)) {
    return createDefaultDocument()
  }
  return {
    ...document,
    nodes: document.nodes,
    edges: document.edges,
  }
}

function upsertCanvasSummary(item: UserCanvas): void {
  const summary: UserCanvasSummary = {
    id: item.id,
    name: item.name,
    description: item.description,
    node_count: item.node_count ?? item.document.nodes.length,
    run_count: item.run_count,
    thumbnail_url: item.thumbnail_url,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
  const index = canvases.value.findIndex((canvas) => canvas.id === item.id)
  if (index >= 0) {
    canvases.value.splice(index, 1, summary)
  } else {
    canvases.value.unshift(summary)
  }
}

function addNode(type: CanvasNodeType): void {
  const index = canvasDocument.value.nodes.length
  const node = makeNode(type, `canvas.nodeTypes.${type}`, 80 + (index % 4) * 210, 90 + Math.floor(index / 4) * 140)
  canvasDocument.value.nodes.push(node)
  selectedNodeId.value = node.id
}

function removeSelectedNode(): void {
  const id = selectedNodeId.value
  if (!id) return
  canvasDocument.value.nodes = canvasDocument.value.nodes.filter((node) => node.id !== id)
  canvasDocument.value.edges = canvasDocument.value.edges.filter((edge) =>
    edge.source_node_id !== id && edge.target_node_id !== id
  )
  selectedNodeId.value = canvasDocument.value.nodes[0]?.id ?? null
}

function nodeTypeLabel(type: CanvasNodeType): string {
  return t(`canvas.nodeTypes.${type}`)
}

function nodeKindClass(type: CanvasNodeType): string {
  return `canvas-kind-${type.replace(/_/g, '-')}`
}

function nodeStyle(node: CanvasNode): Record<string, string> {
  return {
    left: `${node.x}px`,
    top: `${node.y}px`,
    width: `${node.width || 170}px`,
    minHeight: `${node.height || 86}px`,
  }
}

function canvasMeta(item: UserCanvasSummary): string {
  const parts = [
    t('canvas.nodeCount', { count: item.node_count ?? 0 }),
    formatDate(item.updated_at),
  ].filter(Boolean)
  return parts.join(' · ')
}

function modelLabel(modelItem: CanvasModel): string {
  return [modelItem.name || modelItem.id, modelItem.provider].filter(Boolean).join(' · ')
}

function runStatusLabel(status: CanvasRunStatus): string {
  return t(`canvas.runStatus.${status}`)
}

function formatDate(value?: string): string {
  if (!value) return t('common.notAvailable')
  const time = Date.parse(value)
  if (!Number.isFinite(time)) return value
  return new Intl.DateTimeFormat(undefined, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(time))
}

function errorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'object' && error !== null && 'message' in error) {
    const value = (error as { message?: unknown }).message
    if (typeof value === 'string' && value.trim()) {
      return value
    }
  }
  return fallback
}
</script>

<style scoped>
.canvas-studio {
  display: grid;
  min-height: calc(100vh - 7rem);
  min-height: calc(100dvh - 7rem);
  grid-template-columns: 300px minmax(0, 1fr) 330px;
  gap: 1rem;
}

.canvas-panel,
.canvas-workspace {
  min-height: 0;
  overflow: hidden;
  border: 1px solid rgb(229 231 235);
  border-radius: 0.5rem;
  background: rgb(255 255 255);
}

.dark .canvas-panel,
.dark .canvas-workspace {
  border-color: rgb(55 65 81);
  background: rgb(31 41 55 / 0.78);
}

.canvas-panel {
  display: flex;
  flex-direction: column;
}

.canvas-panel-header,
.canvas-toolbar,
.canvas-stage-header {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 1rem;
}

.dark .canvas-panel-header,
.dark .canvas-toolbar,
.dark .canvas-stage-header {
  border-color: rgb(55 65 81);
}

.canvas-section {
  border-bottom: 1px solid rgb(243 244 246);
  padding: 1rem;
}

.dark .canvas-section {
  border-color: rgb(55 65 81);
}

.canvas-section:last-child {
  border-bottom: 0;
}

.canvas-section-title {
  margin-bottom: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  color: rgb(100 116 139);
}

.dark .canvas-section-title {
  color: rgb(148 163 184);
}

.canvas-icon-button {
  display: inline-flex;
  height: 1.875rem;
  width: 1.875rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  color: rgb(100 116 139);
  transition: background-color 0.15s ease, color 0.15s ease;
}

.canvas-icon-button:hover:not(:disabled) {
  background: rgb(241 245 249);
  color: rgb(15 23 42);
}

.canvas-icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.dark .canvas-icon-button {
  color: rgb(148 163 184);
}

.dark .canvas-icon-button:hover:not(:disabled) {
  background: rgb(55 65 81 / 0.72);
  color: rgb(243 244 246);
}

.canvas-list,
.canvas-node-list,
.canvas-run-list {
  max-height: 16rem;
  overflow-y: auto;
}

.canvas-list-item,
.canvas-node-list-item,
.canvas-run-item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.75rem;
  border-radius: 0.5rem;
  padding: 0.625rem;
  text-align: left;
  color: rgb(51 65 85);
}

.canvas-list-item:hover,
.canvas-list-item-active,
.canvas-node-list-item:hover,
.canvas-node-list-item-active {
  background: rgb(236 253 245);
  color: rgb(15 118 110);
}

.dark .canvas-list-item,
.dark .canvas-node-list-item,
.dark .canvas-run-item {
  color: rgb(209 213 219);
}

.dark .canvas-list-item:hover,
.dark .canvas-list-item-active,
.dark .canvas-node-list-item:hover,
.dark .canvas-node-list-item-active {
  background: rgb(20 83 45 / 0.28);
  color: rgb(167 243 208);
}

.canvas-alert {
  margin-bottom: 0.75rem;
  display: flex;
  gap: 0.5rem;
  border-radius: 0.5rem;
  border: 1px solid rgb(254 215 170);
  background: rgb(255 247 237);
  padding: 0.625rem;
  font-size: 0.75rem;
  line-height: 1.25rem;
  color: rgb(154 52 18);
}

.dark .canvas-alert {
  border-color: rgb(154 52 18 / 0.5);
  background: rgb(124 45 18 / 0.22);
  color: rgb(253 186 116);
}

.canvas-empty-list,
.canvas-placeholder,
.canvas-stage-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  color: rgb(100 116 139);
  text-align: center;
}

.canvas-empty-list,
.canvas-placeholder {
  min-height: 7rem;
  border: 1px dashed rgb(203 213 225);
  padding: 1rem;
  font-size: 0.8125rem;
}

.dark .canvas-empty-list,
.dark .canvas-placeholder {
  border-color: rgb(75 85 99);
  color: rgb(156 163 175);
}

.canvas-workspace {
  display: flex;
  flex-direction: column;
}

.canvas-title-input {
  width: 100%;
  border: 0;
  background: transparent;
  font-size: 1.125rem;
  font-weight: 700;
  color: rgb(17 24 39);
  outline: none;
}

.canvas-description-input {
  margin-top: 0.25rem;
  width: 100%;
  border: 0;
  background: transparent;
  font-size: 0.875rem;
  color: rgb(100 116 139);
  outline: none;
}

.dark .canvas-title-input {
  color: rgb(255 255 255);
}

.dark .canvas-description-input {
  color: rgb(156 163 175);
}

.canvas-toolbar-actions {
  display: flex;
  flex-shrink: 0;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.5rem;
}

.canvas-stage-shell {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
}

.canvas-stage-header {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: rgb(100 116 139);
}

.dark .canvas-stage-header {
  color: rgb(148 163 184);
}

.canvas-stage {
  position: relative;
  min-height: 620px;
  flex: 1;
  overflow: auto;
  background-color: rgb(248 250 252);
  background-image:
    linear-gradient(rgb(226 232 240 / 0.72) 1px, transparent 1px),
    linear-gradient(90deg, rgb(226 232 240 / 0.72) 1px, transparent 1px);
  background-size: 28px 28px;
}

.dark .canvas-stage {
  background-color: rgb(15 23 42);
  background-image:
    linear-gradient(rgb(51 65 85 / 0.58) 1px, transparent 1px),
    linear-gradient(90deg, rgb(51 65 85 / 0.58) 1px, transparent 1px);
}

.canvas-edges {
  pointer-events: none;
  position: absolute;
  left: 0;
  top: 0;
  height: 620px;
  width: 980px;
}

.canvas-edge {
  fill: none;
  stroke: rgb(20 184 166);
  stroke-linecap: round;
  stroke-width: 2;
}

.canvas-node {
  position: absolute;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.35rem;
  border: 1px solid currentColor;
  border-radius: 0.5rem;
  background: rgb(255 255 255);
  padding: 0.75rem;
  text-align: left;
  box-shadow: 0 12px 28px rgb(15 23 42 / 0.1);
  transition: box-shadow 0.15s ease, transform 0.15s ease;
}

.canvas-node:hover,
.canvas-node-selected {
  box-shadow: 0 18px 38px rgb(15 23 42 / 0.16);
  transform: translateY(-1px);
}

.dark .canvas-node {
  background: rgb(17 24 39);
  box-shadow: 0 16px 32px rgb(0 0 0 / 0.24);
}

.canvas-node-kind {
  font-size: 0.6875rem;
  font-weight: 800;
  text-transform: uppercase;
}

.canvas-node-title {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.875rem;
  font-weight: 700;
  color: rgb(17 24 39);
}

.dark .canvas-node-title {
  color: rgb(243 244 246);
}

.canvas-node-status {
  font-size: 0.75rem;
  color: rgb(100 116 139);
}

.dark .canvas-node-status {
  color: rgb(148 163 184);
}

.canvas-stage-empty {
  position: absolute;
  inset: 0;
  min-height: 18rem;
}

.canvas-node-type-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.canvas-node-type-button {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid currentColor;
  border-radius: 0.5rem;
  padding: 0.5rem 0.625rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-align: left;
}

.canvas-node-list-dot,
.canvas-run-status {
  height: 0.625rem;
  width: 0.625rem;
  flex-shrink: 0;
  border-radius: 9999px;
}

.canvas-template-entry {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  border-radius: 0.5rem;
  border: 1px dashed rgb(203 213 225);
  padding: 0.75rem;
  color: rgb(20 184 166);
}

.dark .canvas-template-entry {
  border-color: rgb(75 85 99);
  color: rgb(94 234 212);
}

.canvas-kind-text {
  color: rgb(37 99 235);
  background: rgb(239 246 255);
}

.canvas-kind-image {
  color: rgb(5 150 105);
  background: rgb(236 253 245);
}

.canvas-kind-prompt {
  color: rgb(124 58 237);
  background: rgb(245 243 255);
}

.canvas-kind-loop {
  color: rgb(217 119 6);
  background: rgb(255 251 235);
}

.canvas-kind-group {
  color: rgb(71 85 105);
  background: rgb(248 250 252);
}

.canvas-kind-text-to-image {
  color: rgb(219 39 119);
  background: rgb(253 242 248);
}

.canvas-kind-image-to-image {
  color: rgb(14 116 144);
  background: rgb(236 254 255);
}

.canvas-kind-result {
  color: rgb(22 163 74);
  background: rgb(240 253 244);
}

.canvas-run-status-queued {
  background: rgb(59 130 246);
}

.canvas-run-status-running {
  background: rgb(245 158 11);
}

.canvas-run-status-succeeded {
  background: rgb(34 197 94);
}

.canvas-run-status-failed,
.canvas-run-status-canceled {
  background: rgb(239 68 68);
}

@media (max-width: 1279px) {
  .canvas-studio {
    grid-template-columns: 280px minmax(0, 1fr);
  }

  .canvas-inspector-panel {
    grid-column: 1 / -1;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .canvas-inspector-panel .canvas-section {
    border-right: 1px solid rgb(243 244 246);
  }

  .dark .canvas-inspector-panel .canvas-section {
    border-right-color: rgb(55 65 81);
  }
}

@media (max-width: 900px) {
  .canvas-studio {
    grid-template-columns: 1fr;
  }

  .canvas-inspector-panel {
    display: flex;
  }

  .canvas-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .canvas-toolbar-actions {
    justify-content: flex-start;
  }
}
</style>
