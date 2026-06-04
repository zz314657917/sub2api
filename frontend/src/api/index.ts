/**
 * API Client for Sub2API Backend
 * Central export point for all API modules
 */

// Re-export the HTTP client
export { apiClient } from './client'

// Auth API
export { authAPI, isTotp2FARequired, type LoginResponse } from './auth'

// User APIs
export { keysAPI } from './keys'
export { usageAPI } from './usage'
export { userAPI } from './user'
export { redeemAPI, type RedeemHistoryItem } from './redeem'
export { paymentAPI } from './payment'
export { userGroupsAPI } from './groups'
export { userChannelsAPI } from './channels'
export { openWebUIAPI, type OpenWebUILaunchResponse } from './openWebUI'
export { ticketsAPI } from './tickets'
export {
  welfareAPI,
  type WelfareDailyCheckin,
  type WelfareDailyCheckinClaimResponse,
  type WelfareDailyCheckinMilestone,
  type WelfareMilestoneClaimResponse,
  type WelfareNewUserTrial,
  type WelfareNewUserTrialRewardClaimResponse,
  type WelfareOverview,
} from './welfare'
export {
  CHAT_STUDIO_DEFAULT_MODEL,
  CHAT_STUDIO_STORAGE_KEY,
  ChatStudioError,
  createChatCompletionStream,
  extractChatStudioDelta,
  isAbortError,
  listChatModels,
  type ChatStudioCompletionInput,
  type ChatStudioCompletionResult,
  type ChatStudioModel,
  type ChatStudioMessage,
  type ChatStudioRole,
} from './chatStudio'
export { totpAPI } from './totp'
export { default as announcementsAPI } from './announcements'
export { channelMonitorUserAPI } from './channelMonitor'
export { default as tutorialsAPI } from './tutorials'
export { modelMarketAPI, type ModelMarketCatalog, type ModelMarketGroup, type ModelMarketPriceRow } from './modelMarket'

// Admin APIs
export { adminAPI } from './admin'

// Default export
export { default } from './client'
