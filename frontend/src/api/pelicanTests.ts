import { apiClient } from './client'

export interface PelicanGroup { id: number; name: string; platform?: string }
export interface PelicanEntry {
  reasoning_effort?: string | null
  plan_id: number; group_id: number; group_name: string; platform?: string; account_id: number; model_id: string
  status: 'success' | 'failed' | 'skipped'; latency_ms: number; char_count: number; min_chars: number
  finished_at: string; history_count: number; result_id: number; artwork_result_id: number | null; error_message?: string; error_code?: string; error_message_safe?: string
}
export interface PelicanResult {
  reasoning_effort?: string | null
  id: number; plan_id: number; group_id: number; account_id: number; model_id: string; prompt_version: string
  status: 'success' | 'failed' | 'skipped'; error_message?: string; error_code?: string; error_message_safe?: string; latency_ms: number; char_count: number; min_chars: number
  started_at: string; finished_at: string; html?: string
}
export interface PelicanPlan {
  id: number; group_id: number; group_name: string; model_id: string; interval_minutes: number; enabled: boolean
  max_results: number; min_chars: number; last_run_at?: string | null; next_run_at?: string | null; running_until?: string | null
  daily_call_limit: number | null; failure_pause_threshold: number | null; retention_days: number | null
  timeout_seconds?: number
  reasoning_effort: string
  daily_calls_used: number | null; usage_day: string | null; consecutive_failed_runs: number | null; pause_reason: '' | 'consecutive_failures' | null
  estimated_calls_per_run: number | null; last_run_calls: number | null
  created_at: string; updated_at: string
}
export type PelicanPlanInput = Pick<PelicanPlan, 'group_id' | 'model_id' | 'interval_minutes' | 'enabled' | 'max_results' | 'min_chars' | 'daily_call_limit' | 'failure_pause_threshold' | 'retention_days' | 'reasoning_effort' | 'timeout_seconds'>
export interface PelicanListResponse { items: PelicanEntry[]; total: number; page: number; page_size: number; groups: PelicanGroup[] }
export interface PelicanTestMetadata { display_name: string; enabled: boolean }
export interface PelicanTestSettings extends PelicanTestMetadata { prompt: string }
export interface PelicanTestSettingsUpdate { display_name: string; prompt: string; enabled?: boolean }
export interface PelicanPlanModelsResponse { models: string[] }

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
export async function getPelicanTestMetadata(signal?: AbortSignal): Promise<PelicanTestMetadata> { const { data } = await apiClient.get<PelicanTestMetadata>('/pelican-tests/metadata', { signal }); return data }
export async function getPelicanTestSettings(): Promise<PelicanTestSettings> { const { data } = await apiClient.get<PelicanTestSettings>('/admin/pelican-test-settings'); return data }
export async function listPelicanPlanModels(groupID: number, signal?: AbortSignal): Promise<string[]> { const { data } = await apiClient.get<PelicanPlanModelsResponse>('/admin/pelican-test-plans/models', { params: { group_id: groupID }, signal }); return data?.models ?? [] }
export async function updatePelicanTestSettings(input: PelicanTestSettingsUpdate): Promise<PelicanTestSettings> { const { data } = await apiClient.put<PelicanTestSettings>('/admin/pelican-test-settings', input); return data }
export async function listPelicanPlans(): Promise<PelicanPlan[]> { const { data } = await apiClient.get<PelicanPlan[]>('/admin/pelican-test-plans'); return data ?? [] }
export async function createPelicanPlan(input: PelicanPlanInput): Promise<PelicanPlan> { const { data } = await apiClient.post<PelicanPlan>('/admin/pelican-test-plans', input); return data }
export async function updatePelicanPlan(id: number, input: PelicanPlanInput): Promise<PelicanPlan> { const { data } = await apiClient.put<PelicanPlan>(`/admin/pelican-test-plans/${id}`, input); return data }
export async function deletePelicanPlan(id: number): Promise<void> { await apiClient.delete(`/admin/pelican-test-plans/${id}`) }
export async function runPelicanPlan(id: number): Promise<{ queued: boolean }> { const { data } = await apiClient.post<{ queued: boolean }>(`/admin/pelican-test-plans/${id}/run`); return data }
export async function resumePelicanPlan(id: number): Promise<PelicanPlan> { const { data } = await apiClient.post<PelicanPlan>(`/admin/pelican-test-plans/${id}/resume`); return data }
export async function cleanupPelicanPlan(id: number, scope: 'failed' | 'expired'): Promise<{ deleted_count: number }> { const { data } = await apiClient.post<{ deleted_count: number }>(`/admin/pelican-test-plans/${id}/cleanup`, { scope }); return data }
