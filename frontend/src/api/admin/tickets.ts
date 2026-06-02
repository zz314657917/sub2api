import { apiClient } from '../client'
import type {
  AdminSupportTicketListParams,
  CreateSupportTicketMessageRequest,
  CreateSupportTicketRequest,
  PaginatedResponse,
  SupportTicket,
  SupportTicketDetail,
} from '@/types'

function buildListParams(params: AdminSupportTicketListParams = {}) {
  return {
    page: params.page,
    page_size: params.page_size,
    status: params.status || undefined,
    ticket_type: params.ticket_type || undefined,
    search: params.search || undefined,
    user_id: params.user_id || undefined,
    event_type: params.event_type || undefined,
    event_key: params.event_key || undefined,
    date_from: params.date_from || undefined,
    date_to: params.date_to || undefined,
    unread_only: params.unread_only || undefined,
    sort_by: params.sort_by || undefined,
    sort_order: params.sort_order || undefined,
  }
}

export async function listTickets(
  params: AdminSupportTicketListParams = {}
): Promise<PaginatedResponse<SupportTicket>> {
  const { data } = await apiClient.get<PaginatedResponse<SupportTicket>>('/admin/tickets', {
    params: buildListParams(params),
  })
  return data
}

export async function getTicket(id: number): Promise<SupportTicketDetail> {
  const { data } = await apiClient.get<SupportTicketDetail>(`/admin/tickets/${id}`)
  return data
}

export async function createMessage(
  id: number,
  input: CreateSupportTicketMessageRequest
): Promise<SupportTicketDetail> {
  const { data } = await apiClient.post<SupportTicketDetail>(`/admin/tickets/${id}/messages`, input)
  return data
}

export async function markRead(id: number): Promise<void> {
  await apiClient.post(`/admin/tickets/${id}/read`)
}

export async function closeTicket(id: number): Promise<void> {
  await apiClient.post(`/admin/tickets/${id}/close`)
}

export async function reopenTicket(id: number): Promise<void> {
  await apiClient.post(`/admin/tickets/${id}/reopen`)
}

export async function createForUser(
  userId: number,
  input: CreateSupportTicketRequest
): Promise<SupportTicket> {
  const { data } = await apiClient.post<SupportTicket>(`/admin/users/${userId}/tickets`, input)
  return data
}

export const adminTicketsAPI = {
  list: listTickets,
  get: getTicket,
  createMessage,
  markRead,
  close: closeTicket,
  reopen: reopenTicket,
  createForUser,
}

export default adminTicketsAPI
