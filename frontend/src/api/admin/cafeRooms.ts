import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'
import type {
  CafeRoom,
  CafeRoomBulkInput,
  CafeRoomBulkResult,
  CafeRoomInput,
  CafeRoomUpdateInput,
  CafeRound,
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
}

export default cafeRoomsAPI
