import { apiClient } from './client'

export interface OpenWebUILaunchResponse {
  launch_url: string
  expires_at: string
}

export const openWebUIAPI = {
  async launch(apiKeyId: number): Promise<OpenWebUILaunchResponse> {
    const response = await apiClient.post<OpenWebUILaunchResponse>('/user/open-webui/launch', {
      api_key_id: apiKeyId,
    })
    return response.data
  },
}
