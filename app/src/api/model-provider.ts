import { http } from './request'
import type { GetModelsResponse } from '@/types/model'

export interface CustomProviderModelsRequest {
  api_base_url: string
  api_key: string
}

export const getCustomProviderModels = async (
  data: CustomProviderModelsRequest
): Promise<GetModelsResponse> => {
  return http.post('/user-settings/custom-provider-models', data)
}
