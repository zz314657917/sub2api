/**
 * User API endpoints
 * Handles user profile management and password changes
 */

import { apiClient } from './client'
import {
  resolveWeChatOAuthStartStrict,
  prepareOAuthBindAccessTokenCookie,
  type WeChatOAuthPublicSettings,
} from './auth'
import type {
  User,
  ChangePasswordRequest,
  NotifyEmailEntry,
  UserAuthProvider,
  UserAffiliateDetail,
  AffiliateTransferResponse,
  Account,
  AccountShareMode,
  AccountUsageInfo,
  CreateUserAccountRequest,
  ImportUserAccountRequest,
  PaginatedResponse,
  UpdateUserAccountRequest,
  UserAccountAuthURLRequest,
  UserAccountAuthURLResponse,
  UserAccountCapacityPools,
  UserAccountExchangeCodeRequest,
  UserAccountSessionImportRequest,
  UserAccountShareSummary,
  UserAccountTransferResponse
} from '@/types'

/**
 * Get current user profile
 * @returns User profile data
 */
export async function getProfile(): Promise<User> {
  const { data } = await apiClient.get<User>('/user/profile')
  return data
}

/**
 * Update current user profile
 * @param profile - Profile data to update
 * @returns Updated user profile data
 */
export async function updateProfile(profile: {
  username?: string
  avatar_url?: string | null
  balance_notify_enabled?: boolean
  balance_notify_threshold?: number | null
  balance_notify_extra_emails?: NotifyEmailEntry[]
}): Promise<User> {
  const { data } = await apiClient.put<User>('/user', profile)
  return data
}

/**
 * Change current user password
 * @param passwords - Old and new password
 * @returns Success message
 */
export async function changePassword(
  oldPassword: string,
  newPassword: string
): Promise<{ message: string }> {
  const payload: ChangePasswordRequest = {
    old_password: oldPassword,
    new_password: newPassword
  }

  const { data } = await apiClient.put<{ message: string }>('/user/password', payload)
  return data
}

/**
 * Send verification code for adding a notify email
 * @param email - Email address to verify
 */
export async function sendNotifyEmailCode(email: string): Promise<void> {
  await apiClient.post('/user/notify-email/send-code', { email })
}

/**
 * Verify and add a notify email
 * @param email - Email address to add
 * @param code - Verification code
 */
export async function verifyNotifyEmail(email: string, code: string): Promise<void> {
  await apiClient.post('/user/notify-email/verify', { email, code })
}

/**
 * Remove a notify email
 * @param email - Email address to remove
 */
export async function removeNotifyEmail(email: string): Promise<void> {
  await apiClient.delete('/user/notify-email', { data: { email } })
}

/**
 * Toggle a notify email's disabled state
 * @param email - Email address (empty string for primary email placeholder)
 * @param disabled - Whether to disable the email
 */
export async function toggleNotifyEmail(email: string, disabled: boolean): Promise<User> {
  const { data } = await apiClient.put<User>('/user/notify-email/toggle', { email, disabled })
  return data
}

export async function sendEmailBindingCode(email: string): Promise<void> {
  await apiClient.post('/user/account-bindings/email/send-code', { email })
}

export async function bindEmailIdentity(payload: {
  email: string
  verify_code: string
  password: string
}): Promise<User> {
  const { data } = await apiClient.post<User>('/user/account-bindings/email', payload)
  return data
}

export async function unbindAuthIdentity(provider: BindableOAuthProvider): Promise<User> {
  const { data } = await apiClient.delete<User>(`/user/account-bindings/${provider}`)
  return data
}

export type BindableOAuthProvider = Exclude<UserAuthProvider, 'email'>

interface BuildOAuthBindingStartURLOptions {
  redirectTo?: string
  wechatOAuthSettings?: WeChatOAuthPublicSettings | null
}

export function resolveWeChatOAuthMode(): 'open' | 'mp' {
  if (typeof navigator === 'undefined') {
    return 'open'
  }
  return /MicroMessenger/i.test(navigator.userAgent) ? 'mp' : 'open'
}

function resolveWeChatOAuthBindingMode(
  settings?: WeChatOAuthPublicSettings | null
): 'open' | 'mp' | null {
  if (settings) {
    return resolveWeChatOAuthStartStrict(settings).mode
  }
  return resolveWeChatOAuthMode()
}

export function buildOAuthBindingStartURL(
  provider: BindableOAuthProvider,
  options: BuildOAuthBindingStartURLOptions = {}
): string | null {
  const redirectTo = options.redirectTo?.trim() || '/profile'
  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalized = apiBase.replace(/\/$/, '')
  const params = new URLSearchParams({
    redirect: redirectTo,
    intent: 'bind_current_user'
  })

  if (provider === 'wechat') {
    const mode = resolveWeChatOAuthBindingMode(options.wechatOAuthSettings)
    if (!mode) {
      return null
    }
    params.set('mode', mode)
  }

  return `${normalized}/auth/oauth/${provider}/bind/start?${params.toString()}`
}

export async function startOAuthBinding(
  provider: BindableOAuthProvider,
  options: BuildOAuthBindingStartURLOptions = {}
): Promise<void> {
  if (typeof window === 'undefined') {
    return
  }
  const startURL = buildOAuthBindingStartURL(provider, options)
  if (!startURL) {
    return
  }
  await prepareOAuthBindAccessTokenCookie()
  window.location.href = startURL
}

export async function getAffiliateDetail(): Promise<UserAffiliateDetail> {
  const { data } = await apiClient.get<UserAffiliateDetail>('/user/aff')
  return data
}

export async function transferAffiliateQuota(): Promise<AffiliateTransferResponse> {
  const { data } = await apiClient.post<AffiliateTransferResponse>('/user/aff/transfer')
  return data
}

export async function listAccounts(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<Account>> {
  const { data } = await apiClient.get<PaginatedResponse<Account>>('/user/accounts', {
    params: { page, page_size: pageSize, ...filters },
    signal: options?.signal
  })
  return data
}

export async function getAccountById(id: number): Promise<Account> {
  const { data } = await apiClient.get<Account>(`/user/accounts/${id}`)
  return data
}

export async function createAccount(payload: CreateUserAccountRequest): Promise<Account> {
  const { data } = await apiClient.post<Account>('/user/accounts', payload)
  return data
}

export async function importAccount(payload: ImportUserAccountRequest): Promise<Account> {
  const { data } = await apiClient.post<Account>('/user/accounts/import', payload)
  return data
}

export async function updateAccount(id: number, payload: UpdateUserAccountRequest): Promise<Account> {
  const { data } = await apiClient.put<Account>(`/user/accounts/${id}`, payload)
  return data
}

export async function deleteAccount(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/user/accounts/${id}`)
  return data
}

export async function updateAccountShareMode(id: number, shareMode: AccountShareMode): Promise<Account> {
  const { data } = await apiClient.post<Account>(`/user/accounts/${id}/share-mode`, {
    share_mode: shareMode
  })
  return data
}

export async function testAccount(id: number, payload?: { model_id?: string; prompt?: string; mode?: string }): Promise<unknown> {
  const { data } = await apiClient.post<unknown>(`/user/accounts/${id}/test`, payload ?? {})
  return data
}

export async function getAccountUsage(id: number, source: 'active' | 'passive' = 'active'): Promise<AccountUsageInfo> {
  const { data } = await apiClient.get<AccountUsageInfo>(`/user/accounts/${id}/usage`, {
    params: { source }
  })
  return data
}

export async function getAccountShareSummary(): Promise<UserAccountShareSummary> {
  const { data } = await apiClient.get<UserAccountShareSummary>('/user/accounts/share/summary')
  return data
}

export async function getAccountCapacityPools(options?: { signal?: AbortSignal }): Promise<UserAccountCapacityPools> {
  const { data } = await apiClient.get<UserAccountCapacityPools>('/user/accounts/capacity-pools', {
    signal: options?.signal
  })
  return data
}

export async function transferAccountShareToBalance(): Promise<UserAccountTransferResponse> {
  const { data } = await apiClient.post<UserAccountTransferResponse>('/user/accounts/share/transfer')
  return data
}

export async function generateAccountAuthURL(payload: UserAccountAuthURLRequest): Promise<UserAccountAuthURLResponse> {
  const { data } = await apiClient.post<UserAccountAuthURLResponse>('/user/accounts/oauth/auth-url', payload)
  return data
}

export async function exchangeAccountOAuthCode(payload: UserAccountExchangeCodeRequest): Promise<Account> {
  const { data } = await apiClient.post<Account>('/user/accounts/oauth/exchange-code', payload)
  return data
}

export async function importAccountSession(payload: UserAccountSessionImportRequest): Promise<Account> {
  const { data } = await apiClient.post<Account>('/user/accounts/session-import', payload)
  return data
}

export const userAPI = {
  getProfile,
  updateProfile,
  changePassword,
  sendNotifyEmailCode,
  verifyNotifyEmail,
  removeNotifyEmail,
  toggleNotifyEmail,
  sendEmailBindingCode,
  bindEmailIdentity,
  unbindAuthIdentity,
  buildOAuthBindingStartURL,
  startOAuthBinding,
  getAffiliateDetail,
  transferAffiliateQuota,
  listAccounts,
  getAccountById,
  createAccount,
  importAccount,
  updateAccount,
  deleteAccount,
  updateAccountShareMode,
  testAccount,
  getAccountUsage,
  getAccountShareSummary,
  getAccountCapacityPools,
  transferAccountShareToBalance,
  generateAccountAuthURL,
  exchangeAccountOAuthCode,
  importAccountSession
}

export default userAPI
