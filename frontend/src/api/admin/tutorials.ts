/**
 * Admin tutorial CMS API endpoints
 */

import { apiClient } from '../client'
import type {
  BasePaginationResponse,
  CreateTutorialPageRequest,
  TutorialPage,
  TutorialPageSummary,
  TutorialPageStatus,
  UpdateTutorialPageRequest
} from '@/types'

export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: TutorialPageStatus | ''
    category?: string
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<BasePaginationResponse<TutorialPageSummary>> {
  const { data } = await apiClient.get<BasePaginationResponse<TutorialPageSummary>>('/admin/tutorials', {
    params: { page, page_size: pageSize, ...filters },
    signal: options?.signal
  })
  return data
}

export async function getById(id: number): Promise<TutorialPage> {
  const { data } = await apiClient.get<TutorialPage>(`/admin/tutorials/${id}`)
  return data
}

export async function create(request: CreateTutorialPageRequest): Promise<TutorialPage> {
  const { data } = await apiClient.post<TutorialPage>('/admin/tutorials', request)
  return data
}

export async function update(id: number, request: UpdateTutorialPageRequest): Promise<TutorialPage> {
  const { data } = await apiClient.put<TutorialPage>(`/admin/tutorials/${id}`, request)
  return data
}

export async function updateStatus(id: number, status: TutorialPageStatus): Promise<TutorialPage> {
  const { data } = await apiClient.put<TutorialPage>(`/admin/tutorials/${id}/status`, { status })
  return data
}

export async function deleteTutorial(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/tutorials/${id}`)
  return data
}

const tutorialsAPI = {
  list,
  getById,
  create,
  update,
  updateStatus,
  delete: deleteTutorial
}

export default tutorialsAPI
