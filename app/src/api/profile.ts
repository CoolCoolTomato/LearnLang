import { http } from './request'
import type { User } from '@/types/auth'
import { API_CONFIG } from './config'

export interface UpdateProfileRequest {
  username?: string
  avatar_url?: string
  email?: string
  phone?: string
}

export interface UploadAvatarResponse {
  filename: string
}

export const getProfile = () => {
  return http.get<User>('/profile')
}

export const updateProfile = (data: UpdateProfileRequest) => {
  return http.put<User>('/profile', data)
}

export const uploadProfileAvatar = (file: File) => {
  const formData = new FormData()
  formData.append('avatar', file)
  return http.upload<UploadAvatarResponse>('/profile/avatar/upload', formData)
}

export const updateProfileAvatar = (filename: string) => {
  return http.put('/profile/avatar', { filename })
}

export const resolveAvatarUrl = (avatarUrl?: string | null) => {
  if (!avatarUrl) return ''
  if (avatarUrl.startsWith('http://') || avatarUrl.startsWith('https://') || avatarUrl.startsWith('blob:') || avatarUrl.startsWith('data:')) {
    return avatarUrl
  }
  if (avatarUrl.startsWith('/')) {
    return `${API_CONFIG.BASE_URL}${avatarUrl}`
  }
  return `${API_CONFIG.BASE_URL}${API_CONFIG.API_PREFIX}/profile/avatar/${avatarUrl}`
}
