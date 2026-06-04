import { apiClient } from '../client'
import type { ModelMarketCatalog } from '../modelMarket'

export async function getCatalog(): Promise<ModelMarketCatalog> {
  const { data } = await apiClient.get<ModelMarketCatalog>('/admin/model-market/catalog')
  return data
}

export async function updateCatalog(catalog: ModelMarketCatalog): Promise<ModelMarketCatalog> {
  const { data } = await apiClient.put<ModelMarketCatalog>('/admin/model-market/catalog', catalog)
  return data
}

export async function resetCatalog(): Promise<ModelMarketCatalog> {
  const { data } = await apiClient.post<ModelMarketCatalog>('/admin/model-market/catalog/reset')
  return data
}

export const modelMarketAdminAPI = {
  getCatalog,
  updateCatalog,
  resetCatalog
}

export default modelMarketAdminAPI
