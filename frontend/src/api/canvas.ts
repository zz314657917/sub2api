import { apiClient } from './client'

export type CanvasNodeType =
  | 'text'
  | 'image'
  | 'prompt'
  | 'loop'
  | 'group'
  | 'text_to_image'
  | 'image_to_image'
  | 'result'

export type CanvasRunStatus = 'pending' | 'queued' | 'running' | 'succeeded' | 'failed' | 'canceled'

export interface CanvasNode {
  id: string
  type: CanvasNodeType
  title: string
  x: number
  y: number
  width?: number
  height?: number
  status?: 'idle' | 'queued' | 'running' | 'done' | 'failed'
  config?: Record<string, unknown>
  result?: unknown
  error?: unknown
}

export interface CanvasEdge {
  id: string
  source_node_id: string
  target_node_id: string
  source_handle?: string
  target_handle?: string
}

export interface CanvasDocument {
  nodes: CanvasNode[]
  edges: CanvasEdge[]
  viewport?: {
    x: number
    y: number
    zoom: number
  }
  metadata?: Record<string, unknown>
}

export interface UserCanvasSummary {
  id: string
  name: string
  description?: string
  node_count?: number
  run_count?: number
  thumbnail_url?: string
  created_at: string
  updated_at: string
}

export interface UserCanvas extends UserCanvasSummary {
  model?: string
  document: CanvasDocument
  settings?: Record<string, unknown>
}

export interface CanvasListParams {
  limit?: number
  offset?: number
}

export interface CanvasListResponse {
  items: UserCanvasSummary[]
  total: number
  limit?: number
  offset?: number
}

export interface CanvasWritePayload {
  name: string
  description?: string
  model?: string
  document: CanvasDocument
  settings?: Record<string, unknown>
}

export interface CanvasRun {
  id: string
  canvas_id: string
  status: CanvasRunStatus
  api_key_id?: number
  model?: string
  queued_at?: string
  started_at?: string
  completed_at?: string
  canceled_at?: string
  error_message?: string
  result_node_ids?: string[]
  input?: unknown
  output?: unknown
  outputs?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CanvasRunListParams {
  canvas_id?: string
  limit?: number
  offset?: number
}

export interface CanvasRunListResponse {
  items: CanvasRun[]
  total: number
  limit?: number
  offset?: number
}

export interface CanvasRunCreatePayload {
  canvas_id: string
  api_key_id: number
  model?: string
}

export interface CanvasModel {
  id: string
  name: string
  provider?: string
  capabilities?: CanvasNodeType[]
  supports_image_input?: boolean
  supports_image_output?: boolean
}

export interface CanvasModelListResponse {
  items: CanvasModel[]
}

interface BackendCanvasNode {
  id: string
  type: CanvasNodeType
  position?: Record<string, unknown>
  data?: Record<string, unknown>
  metadata?: Record<string, unknown>
}

interface BackendCanvasEdge {
  id: string
  source: string
  target: string
  source_handle?: string
  target_handle?: string
  data?: Record<string, unknown>
  metadata?: Record<string, unknown>
}

interface BackendCanvasDocument {
  id: number
  title: string
  description?: string
  nodes?: BackendCanvasNode[]
  edges?: BackendCanvasEdge[]
  viewport?: Record<string, unknown>
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

interface BackendCanvasSummary {
  id: number
  title: string
  description?: string
  viewport?: Record<string, unknown>
  metadata?: Record<string, unknown>
  node_count?: number
  edge_count?: number
  created_at: string
  updated_at: string
}

interface BackendCanvasRun {
  id: number
  canvas_id?: number
  status: CanvasRunStatus
  api_key_id?: number
  model?: string
  error_message?: string
  result_node_ids?: string[]
  input?: unknown
  output?: unknown
  outputs?: unknown
  started_at?: string
  completed_at?: string
  canceled_at?: string
  created_at: string
  updated_at: string
}

interface BackendCanvasModelsResponse {
  items: Array<{
    id: string
    name?: string
    display_name?: string
    capabilities?: string[]
  }>
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function numberFromRecord(value: Record<string, unknown> | undefined, key: string, fallback: number): number {
  const raw = value?.[key]
  return typeof raw === 'number' && Number.isFinite(raw) ? raw : fallback
}

function stringFromRecord(value: Record<string, unknown> | undefined, key: string, fallback: string): string {
  const raw = value?.[key]
  return typeof raw === 'string' && raw.trim() ? raw.trim() : fallback
}

function recordFromUnknown(value: unknown): Record<string, unknown> {
  return isRecord(value) ? value : {}
}

function nodeStatusFromRecord(value: Record<string, unknown> | undefined): CanvasNode['status'] {
  const raw = value?.status
  return raw === 'queued' || raw === 'running' || raw === 'done' || raw === 'failed' ? raw : 'idle'
}

function normalizeRunStatus(status: CanvasRunStatus): CanvasRunStatus {
  return status === 'pending' ? 'queued' : status
}

function toBackendNode(node: CanvasNode): BackendCanvasNode {
  return {
    id: node.id,
    type: node.type,
    position: {
      x: node.x,
      y: node.y,
      width: node.width,
      height: node.height,
    },
    data: {
      title: node.title,
      status: node.status,
      config: node.config ?? {},
      result: node.result,
      error: node.error,
    },
  }
}

function fromBackendNode(node: BackendCanvasNode): CanvasNode {
  const position = recordFromUnknown(node.position)
  const data = recordFromUnknown(node.data)
  return {
    id: node.id,
    type: node.type,
    title: stringFromRecord(data, 'title', node.id),
    x: numberFromRecord(position, 'x', 80),
    y: numberFromRecord(position, 'y', 80),
    width: numberFromRecord(position, 'width', 170),
    height: numberFromRecord(position, 'height', 86),
    status: nodeStatusFromRecord(data),
    config: recordFromUnknown(data.config),
    result: data.result,
    error: data.error,
  }
}

function toBackendEdge(edge: CanvasEdge): BackendCanvasEdge {
  return {
    id: edge.id,
    source: edge.source_node_id,
    target: edge.target_node_id,
    source_handle: edge.source_handle,
    target_handle: edge.target_handle,
  }
}

function fromBackendEdge(edge: BackendCanvasEdge): CanvasEdge {
  return {
    id: edge.id,
    source_node_id: edge.source,
    target_node_id: edge.target,
    source_handle: edge.source_handle,
    target_handle: edge.target_handle,
  }
}

function toBackendPayload(payload: CanvasWritePayload): Record<string, unknown> {
  return {
    title: payload.name,
    description: payload.description || '',
    nodes: payload.document.nodes.map(toBackendNode),
    edges: payload.document.edges.map(toBackendEdge),
    viewport: payload.document.viewport ?? {},
    metadata: {
      ...(payload.document.metadata ?? {}),
      ...(payload.settings ?? {}),
      model: payload.model || undefined,
    },
  }
}

function fromBackendSummary(item: BackendCanvasSummary): UserCanvasSummary {
  return {
    id: String(item.id),
    name: item.title,
    description: item.description,
    node_count: item.node_count,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function fromBackendCanvas(item: BackendCanvasDocument): UserCanvas {
  const metadata = recordFromUnknown(item.metadata)
  return {
    ...fromBackendSummary({
      id: item.id,
      title: item.title,
      description: item.description,
      node_count: item.nodes?.length ?? 0,
      created_at: item.created_at,
      updated_at: item.updated_at,
    }),
    model: typeof metadata.model === 'string' ? metadata.model : undefined,
    document: {
      nodes: (item.nodes ?? []).map(fromBackendNode),
      edges: (item.edges ?? []).map(fromBackendEdge),
      viewport: item.viewport as CanvasDocument['viewport'],
      metadata,
    },
    settings: metadata,
  }
}

function fromBackendRun(run: BackendCanvasRun): CanvasRun {
  return {
    id: String(run.id),
    canvas_id: run.canvas_id ? String(run.canvas_id) : '',
    status: normalizeRunStatus(run.status),
    api_key_id: run.api_key_id,
    model: run.model,
    started_at: run.started_at,
    completed_at: run.completed_at,
    canceled_at: run.canceled_at,
    error_message: run.error_message,
    result_node_ids: run.result_node_ids,
    input: run.input,
    output: run.output,
    outputs: recordFromUnknown(run.outputs),
    created_at: run.created_at,
    updated_at: run.updated_at,
  }
}

export async function listCanvases(params: CanvasListParams = {}): Promise<CanvasListResponse> {
  const { data } = await apiClient.get<{
    items: BackendCanvasSummary[]
    total: number
    limit?: number
    offset?: number
  }>('/user/canvases', { params })
  return {
    ...data,
    items: (data.items ?? []).map(fromBackendSummary),
  }
}

export async function getCanvas(id: string): Promise<UserCanvas> {
  const { data } = await apiClient.get<{ item: BackendCanvasDocument }>(`/user/canvases/${id}`)
  return fromBackendCanvas(data.item)
}

export async function createCanvas(payload: CanvasWritePayload): Promise<UserCanvas> {
  const { data } = await apiClient.post<{ item: BackendCanvasDocument }>('/user/canvases', toBackendPayload(payload))
  return fromBackendCanvas(data.item)
}

export async function updateCanvas(id: string, payload: CanvasWritePayload): Promise<UserCanvas> {
  const { data } = await apiClient.put<{ item: BackendCanvasDocument }>(`/user/canvases/${id}`, toBackendPayload(payload))
  return fromBackendCanvas(data.item)
}

export async function deleteCanvas(id: string): Promise<{ deleted: boolean }> {
  const { data } = await apiClient.delete<{ deleted: boolean }>(`/user/canvases/${id}`)
  return data
}

export async function listCanvasRuns(params: CanvasRunListParams = {}): Promise<CanvasRunListResponse> {
  const { data } = await apiClient.get<{
    items: BackendCanvasRun[]
    total: number
    limit?: number
    offset?: number
  }>('/user/canvas-runs', { params })
  return {
    ...data,
    items: (data.items ?? []).map(fromBackendRun),
  }
}

export async function createCanvasRun(payload: CanvasRunCreatePayload): Promise<CanvasRun> {
  const { data } = await apiClient.post<{ item: BackendCanvasRun }>('/user/canvas-runs', {
    canvas_id: Number(payload.canvas_id),
    api_key_id: payload.api_key_id,
    model: payload.model,
  })
  return fromBackendRun(data.item)
}

export async function getCanvasRun(id: string): Promise<CanvasRun> {
  const { data } = await apiClient.get<{ item: BackendCanvasRun }>(`/user/canvas-runs/${id}`)
  return fromBackendRun(data.item)
}

export async function cancelCanvasRun(id: string): Promise<CanvasRun> {
  const { data } = await apiClient.post<{ item: BackendCanvasRun }>(`/user/canvas-runs/${id}/cancel`)
  return fromBackendRun(data.item)
}

export async function listCanvasModels(): Promise<CanvasModelListResponse> {
  const { data } = await apiClient.get<BackendCanvasModelsResponse>('/user/canvas/models')
  return {
    items: (data.items ?? []).map((item) => ({
      id: item.id,
      name: item.display_name || item.name || item.id,
      provider: 'openai',
      capabilities: (item.capabilities ?? []).flatMap((capability): CanvasNodeType[] => {
        if (capability === 'image') return ['text_to_image', 'image_to_image']
        if (capability === 'chat') return ['text', 'prompt']
        return []
      }),
      supports_image_input: item.capabilities?.includes('image') ?? false,
      supports_image_output: item.capabilities?.includes('image') ?? false,
    })),
  }
}
