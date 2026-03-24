import { http } from './request'
import type { Settings, UpdateSettingsRequest } from '@/types/settings'

export const getSettings = () => {
  return http.get<Settings>('/profile/settings')
}

export const updateSettings = (data: UpdateSettingsRequest) => {
  return http.put<Settings>('/profile/settings', data)
}
