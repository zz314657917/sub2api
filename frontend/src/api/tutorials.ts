/**
 * Public tutorial CMS API endpoints
 */

import { apiClient } from './client'
import type { TutorialPage, TutorialPageSummary } from '@/types'

export async function list(): Promise<TutorialPageSummary[]> {
  const { data } = await apiClient.get<{ items: TutorialPageSummary[] }>('/tutorials')
  return data.items ?? []
}

export async function getBySlug(slug: string): Promise<TutorialPage> {
  const { data } = await apiClient.get<TutorialPage>(`/tutorials/${encodeURIComponent(slug)}`)
  return data
}

const tutorialsAPI = {
  list,
  getBySlug
}

export default tutorialsAPI
