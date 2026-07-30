/**
 * Public tutorial CMS API endpoints
 */

import { apiClient } from './client'
import type { TutorialPage, TutorialPageSummary } from '@/types'
import type { QuickstartTutorialConfig } from '@/views/public/tutorialQuickstart'

export async function list(): Promise<TutorialPageSummary[]> {
  const { data } = await apiClient.get<{ items: TutorialPageSummary[] }>('/tutorials')
  return data.items ?? []
}

export async function getBySlug(slug: string): Promise<TutorialPage> {
  const { data } = await apiClient.get<TutorialPage>(`/tutorials/${encodeURIComponent(slug)}`)
  return data
}

export async function getQuickstartConfig(): Promise<QuickstartTutorialConfig> {
  const { data } = await apiClient.get<QuickstartTutorialConfig>('/tutorials/quickstart-config')
  return data
}

const tutorialsAPI = {
  list,
  getBySlug,
  getQuickstartConfig
}

export default tutorialsAPI
