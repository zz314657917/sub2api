import { apiClient } from './client'

export type ModelMarketCategory = 'chat' | 'image' | 'video'

export interface ModelMarketPriceRow {
  id: string
  model?: string
  spec?: string
  input_price?: string
  output_price?: string
  our_price: string
  official_price?: string
  saving?: string
  note?: string
  sort_order: number
  enabled: boolean
}

export interface ModelMarketAccountGroup {
  id: number
  name: string
  platform: string
  rate_multiplier: number
  image_rate_independent: boolean
  image_rate_multiplier: number
  effective_rate_multiplier: number
}

export interface ModelMarketGroup {
  id: string
  title: string
  category: ModelMarketCategory
  platform?: string
  description?: string
  hide_official_price?: boolean
  hide_saving?: boolean
  price_multiplier?: number
  supported_group_ids?: number[]
  supported_groups?: ModelMarketAccountGroup[]
  sort_order: number
  enabled: boolean
  rows: ModelMarketPriceRow[]
}

export interface ModelMarketCatalog {
  version: number
  updated_at?: string
  groups: ModelMarketGroup[]
}

export async function getCatalog(options?: { signal?: AbortSignal }): Promise<ModelMarketCatalog> {
  const { data } = await apiClient.get<ModelMarketCatalog>('/model-market/catalog', {
    signal: options?.signal
  })
  return data
}

export const modelMarketAPI = {
  getCatalog
}

export default modelMarketAPI
