export type ProviderType = 'custom'

export type Theme = 'light' | 'dark' | 'system'

export type Language = 'zh-CN' | 'en-US'

export interface Settings {
  id: number
  user_id: number
  api_base_url?: string
  api_key?: string
  model?: string
  embedding_api_base_url?: string
  embedding_api_key?: string
  embedding_model?: string
  stt_api_base_url?: string
  stt_api_key?: string
  stt_model?: string
  tts_api_base_url?: string
  tts_api_key?: string
  tts_model?: string
  tts_voice?: string
  native_language?: string
  target_language?: string
  timezone?: string
  created_at: string
  updated_at: string
}

export interface UpdateSettingsRequest {
  api_base_url?: string
  api_key?: string
  model?: string
  embedding_api_base_url?: string
  embedding_api_key?: string
  embedding_model?: string
  stt_api_base_url?: string
  stt_api_key?: string
  stt_model?: string
  tts_api_base_url?: string
  tts_api_key?: string
  tts_model?: string
  tts_voice?: string
  theme?: Theme
  language?: Language
  native_language?: string
  target_language?: string
  timezone?: string
}
