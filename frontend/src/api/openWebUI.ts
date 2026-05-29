import { apiClient } from './client'

export interface OpenWebUILaunchResponse {
  launch_url: string
  expires_at: string
}

export const openWebUIAPI = {
  async launch(apiKeyId?: number): Promise<OpenWebUILaunchResponse> {
    const body = typeof apiKeyId === 'number' && apiKeyId > 0 ? { api_key_id: apiKeyId } : {}
    const response = await apiClient.post<OpenWebUILaunchResponse>('/user/open-webui/launch', body)
    return response.data
  },
}
