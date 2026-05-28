import { apiClient } from './client'

export type ImageCreatorOutputFormat = 'png' | 'jpeg' | 'webp'
export type ImageCreatorTaskStatus = 'pending' | 'running' | 'succeeded' | 'failed'

export interface ImageCreatorCreateTaskInput {
  apiKeyId: number
  model: string
  prompt: string
  size: string
  quality: string
  count: number
  outputFormat: ImageCreatorOutputFormat
  background: string
  referenceImage?: File | null
}

export interface ImageCreatorStoredImage {
  id: number
  task_id: number
  user_id: number
  url: string
  output_format: ImageCreatorOutputFormat | string
  mime_type: string
  byte_size: number
  width?: number
  height?: number
  resolution?: string
  aspect_ratio?: string
  orientation?: string
  megapixels?: number
  sha256: string
  revised_prompt?: string
  expires_at: string
  created_at: string
}

export interface ImageCreatorManagedImage extends ImageCreatorStoredImage {
  task_prompt?: string
  task_model?: string
  task_size?: string
  task_quality?: string
}

export interface ImageCreatorTask {
  id: number
  user_id: number
  api_key_id: number
  status: ImageCreatorTaskStatus
  model: string
  prompt: string
  size: string
  quality: string
  output_format: ImageCreatorOutputFormat | string
  background: string
  count: number
  reference_image_mime_type?: string
  reference_image_filename?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  expires_at: string
  created_at: string
  updated_at: string
  images?: ImageCreatorStoredImage[]
}

export interface ImageCreatorTaskListResponse {
  tasks: ImageCreatorTask[]
  images: ImageCreatorStoredImage[]
}

export interface ImageCreatorImageListResponse {
  items: ImageCreatorManagedImage[]
  total: number
  limit: number
  offset: number
}

export interface ImageCreatorImageListParams {
  limit?: number
  offset?: number
  q?: string
  start_date?: string
  end_date?: string
  format?: string
  orientation?: string
  resolution?: string
  aspect_ratio?: string
  min_width?: number
  min_height?: number
}

function appendIfPresent(form: FormData, key: string, value: string | number | undefined): void {
  if (value === undefined) return
  const text = String(value).trim()
  if (text === '') return
  form.append(key, text)
}

function buildJSONPayload(input: ImageCreatorCreateTaskInput): Record<string, unknown> {
  return {
    api_key_id: input.apiKeyId,
    model: input.model,
    prompt: input.prompt,
    size: input.size,
    quality: input.quality,
    count: input.count,
    output_format: input.outputFormat,
    background: input.background,
  }
}

function buildMultipartPayload(input: ImageCreatorCreateTaskInput): FormData {
  const form = new FormData()
  appendIfPresent(form, 'api_key_id', input.apiKeyId)
  appendIfPresent(form, 'model', input.model)
  appendIfPresent(form, 'prompt', input.prompt)
  appendIfPresent(form, 'size', input.size)
  appendIfPresent(form, 'quality', input.quality)
  appendIfPresent(form, 'count', input.count)
  appendIfPresent(form, 'output_format', input.outputFormat)
  appendIfPresent(form, 'background', input.background)
  if (input.referenceImage) {
    form.append('reference_image', input.referenceImage, input.referenceImage.name || 'reference.png')
  }
  return form
}

export async function createImageTask(input: ImageCreatorCreateTaskInput): Promise<ImageCreatorTask> {
  const payload = input.referenceImage ? buildMultipartPayload(input) : buildJSONPayload(input)
  const { data } = await apiClient.post<ImageCreatorTask>('/user/image-creator/tasks', payload)
  return data
}

export async function listImageTasks(): Promise<ImageCreatorTaskListResponse> {
  const { data } = await apiClient.get<ImageCreatorTaskListResponse>('/user/image-creator/tasks')
  return data
}

export async function listManagedImages(params: ImageCreatorImageListParams = {}): Promise<ImageCreatorImageListResponse> {
  const { data } = await apiClient.get<ImageCreatorImageListResponse>('/user/image-creator/images', {
    params,
  })
  return data
}

export async function deleteManagedImages(ids: number[]): Promise<{ deleted: number }> {
  const { data } = await apiClient.delete<{ deleted: number }>('/user/image-creator/images', {
    data: { ids },
  })
  return data
}

export async function getImageTask(id: number): Promise<ImageCreatorTask> {
  const { data } = await apiClient.get<ImageCreatorTask>(`/user/image-creator/tasks/${id}`)
  return data
}

export async function downloadImageFile(url: string): Promise<Blob> {
  const { data } = await apiClient.get<Blob>(url, {
    baseURL: '',
    responseType: 'blob',
  })
  return data
}
