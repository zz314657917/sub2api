import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'
import type {
  CafeRoom,
  CafeRoomBulkInput,
  CafeRoomBulkResult,
  CafeRoomInput,
  CafeRoomUpdateInput,
  CafeRound,
  CafeWorkstationPosition,
} from '@/types/pixelCafe'

export interface CafeRoomListParams {
  page?: number
  page_size?: number
  status?: string
  zone?: string
  search?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface CafeRoomAccountOption {
  id: number
  name: string
  platform: string
  status: string
  email_masked?: string
  plan_type?: string
}

export interface CafePendingRound {
  id: number
  status: string
  room_id: number
  room_code: string
  room_name: string
  subscription_tier: 'plus' | 'pro'
  paid_shares: number
  total_shares: number
  joined_buyers: number
  max_buyers: number
  paid_full_at?: string | null
  fulfillment_deadline_at?: string | null
}

export interface CafeRoomAccountOptionParams {
  page?: number
  page_size?: number
  search?: string
  plan_id?: number
  exclude_room_id?: number
  ids?: number[]
}

const cafeRoomsAPI = {
  getWorkstationLayout() {
    return apiClient.get<CafeWorkstationPosition[]>('/admin/cafe/layout')
  },

  updateWorkstationLayout(layout: CafeWorkstationPosition[]) {
    return apiClient.put<CafeWorkstationPosition[]>('/admin/cafe/layout', layout)
  },

  list(params?: CafeRoomListParams) {
    return apiClient.get<PaginatedResponse<CafeRoom>>('/admin/cafe/rooms', { params })
  },

  listAccountOptions(params: CafeRoomAccountOptionParams) {
    return apiClient.get<PaginatedResponse<CafeRoomAccountOption>>('/admin/cafe/rooms/account-options', {
      params: { ...params, ids: params.ids?.join(',') || undefined },
    })
  },

  get(id: number) {
    return apiClient.get<CafeRoom>(`/admin/cafe/rooms/${id}`)
  },

  create(data: CafeRoomInput) {
    return apiClient.post<CafeRoom>('/admin/cafe/rooms', data)
  },

  update(id: number, data: CafeRoomUpdateInput) {
    return apiClient.patch<CafeRoom>(`/admin/cafe/rooms/${id}`, data)
  },

  remove(id: number) {
    return apiClient.delete<{ message: string }>(`/admin/cafe/rooms/${id}`)
  },

  bulkCreate(data: CafeRoomBulkInput) {
    return apiClient.post<CafeRoomBulkResult>('/admin/cafe/rooms/bulk', data)
  },

  openRound(id: number) {
    return apiClient.post<CafeRound>(`/admin/cafe/rooms/${id}/open-round`)
  },

  pauseRound(id: number) {
    return apiClient.post<CafeRound>(`/admin/cafe/rooms/${id}/pause-round`)
  },

  listPendingRounds(params?: { page?: number; page_size?: number; search?: string }) {
    return apiClient.get<PaginatedResponse<CafePendingRound>>('/admin/cafe/rounds/pending', { params })
  },

  listRoundAccountOptions(roundID: number, params?: { page?: number; page_size?: number; search?: string }) {
    return apiClient.get<PaginatedResponse<CafeRoomAccountOption>>(`/admin/cafe/rounds/${roundID}/account-options`, { params })
  },

  assignRoundAccount(roundID: number, accountID: number) {
    return apiClient.post<CafePendingRound>(`/admin/cafe/rounds/${roundID}/assign-account`, { account_id: accountID })
  },
}

export default cafeRoomsAPI
