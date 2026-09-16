import { apiClient } from './client'

export interface PelicanGroup { id: number; name: string }
export interface PelicanEntry {
  plan_id: number; group_id: number; group_name: string; account_id: number; model_id: string
  status: 'success' | 'failed' | 'skipped'; latency_ms: number; char_count: number; min_chars: number
  finished_at: string; history_count: number; result_id: number; artwork_result_id: number | null; error_message?: string
}
export interface PelicanResult {
  id: number; plan_id: number; group_id: number; account_id: number; model_id: string; prompt_version: string
  status: 'success' | 'failed' | 'skipped'; error_message?: string; latency_ms: number; char_count: number; min_chars: number
  started_at: string; finished_at: string; html?: string
}
export interface PelicanPlan {
  id: number; group_id: number; group_name: string; model_id: string; interval_minutes: number; enabled: boolean
  max_results: number; min_chars: number; last_run_at?: string | null; next_run_at?: string | null; running_until?: string | null
  created_at: string; updated_at: string
}
export type PelicanPlanInput = Pick<PelicanPlan, 'group_id' | 'model_id' | 'interval_minutes' | 'enabled' | 'max_results' | 'min_chars'>
export interface PelicanListResponse { items: PelicanEntry[]; total: number; page: number; page_size: number; groups: PelicanGroup[] }

export function buildPelicanListParams(groupID: number | '', accountQuery: string, page: number, pageSize: number): { group_id?: number; account_id?: number; page: number; page_size: number } {
  const accountID = Number(accountQuery.replace(/^#/, ''))
  return { group_id: groupID || undefined, account_id: Number.isFinite(accountID) && accountID > 0 ? accountID : undefined, page, page_size: pageSize }
}

export async function listPelicanTests(params: { group_id?: number; account_id?: number; page: number; page_size: number }, signal?: AbortSignal): Promise<PelicanListResponse> {
  const { data } = await apiClient.get<PelicanListResponse>('/pelican-tests', { params, signal })
  return data
}
export async function listPelicanHistory(params: { plan_id: number; account_id?: number }, signal?: AbortSignal): Promise<PelicanResult[]> {
  const { data } = await apiClient.get<PelicanResult[]>('/pelican-tests/history', { params, signal })
  return data ?? []
}
export async function getPelicanResult(id: number, signal?: AbortSignal): Promise<PelicanResult> {
  const { data } = await apiClient.get<PelicanResult>(`/pelican-tests/results/${id}`, { signal })
  return data
}
export async function listPelicanPlans(): Promise<PelicanPlan[]> { const { data } = await apiClient.get<PelicanPlan[]>('/admin/pelican-test-plans'); return data ?? [] }
export async function createPelicanPlan(input: PelicanPlanInput): Promise<PelicanPlan> { const { data } = await apiClient.post<PelicanPlan>('/admin/pelican-test-plans', input); return data }
export async function updatePelicanPlan(id: number, input: PelicanPlanInput): Promise<PelicanPlan> { const { data } = await apiClient.put<PelicanPlan>(`/admin/pelican-test-plans/${id}`, input); return data }
export async function deletePelicanPlan(id: number): Promise<void> { await apiClient.delete(`/admin/pelican-test-plans/${id}`) }
export async function runPelicanPlan(id: number): Promise<{ queued: boolean }> { const { data } = await apiClient.post<{ queued: boolean }>(`/admin/pelican-test-plans/${id}/run`); return data }
