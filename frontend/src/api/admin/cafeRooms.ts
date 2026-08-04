import { apiClient } from '../client'
import type { PaginatedResponse, Account } from '@/types'
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

const cafeRoomsAPI = {
  list(params?: CafeRoomListParams) {
    return apiClient.get<PaginatedResponse<CafeRoom>>('/admin/cafe/rooms', { params })
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

export type CafeRoomAccountOption = Pick<Account, 'id' | 'name' | 'platform' | 'status' | 'current_concurrency' | 'concurrency'>

export default cafeRoomsAPI
