export interface ChatRequest {
  message: string
}
export type MessageRole = 'user' | 'assistant' | 'system'

export type InputType = 'text' | 'audio'

export type ChatResponse = ChatMessage

export interface ChatHistoryRequest {
  before_id?: number
}

export interface VoiceFileInMessage {
  id: number
  user_id: number
  voice_role?: string
  voice_url: string
  duration?: number
  file_size?: number
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  id: number
  user_id: number
  role: MessageRole
  text_content: string
  translation?: string
  voice_file_id?: number
  voice_file?: VoiceFileInMessage
  input_type: InputType
  token_count: number
  created_at: string
}

export interface GetChatHistoryParams extends Record<string, string | number | boolean | undefined> {
  before_id?: number
}

export interface GetChatHistoryResponse {
  data: ChatMessage[]
}
