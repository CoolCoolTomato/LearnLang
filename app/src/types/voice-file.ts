export interface VoiceFile {
  id: number
  user_id: number
  voice_role?: string
  voice_url: string
  duration?: number
  file_size?: number
  created_at: string
  updated_at: string
}

export interface CreateVoiceFileRequest {
  user_id: number
  voice_role?: string
  voice_url: string
  duration?: number
  file_size?: number
}

export interface UpdateVoiceFileRequest {
  user_id?: number
  voice_role?: string
  voice_url?: string
  duration?: number
  file_size?: number
}

export interface GetVoiceFilesParams {
  user_id?: number
  page?: number
  size?: number
}

export interface GetVoiceFilesResponse {
  data: VoiceFile[]
}

export interface DeleteVoiceFileResponse {
  message: string
}

export interface GetVoiceFileContentResponse {
  url?: string
}
