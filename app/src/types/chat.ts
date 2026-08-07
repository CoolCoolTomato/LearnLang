export interface ChatRequest {
  message: string
  voice_file_id?: number
}
export type MessageRole = "user" | "assistant" | "system"

export type InputType = "text" | "voice"

export type ChatResponse = ChatMessage

export interface TranscriptionResponse {
  text: string
  voice_file_id: number
}

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
  created_at: string
}

export interface AgentErrorEvent {
  type: "agent_error"
}

export interface AgentTurnCompletedEvent {
  type: "agent_turn_completed"
}

export type ChatWebSocketEvent =
  | ChatMessage
  | AgentErrorEvent
  | AgentTurnCompletedEvent

export interface GetChatHistoryParams extends Record<
  string,
  string | number | boolean | undefined
> {
  before_id?: number
}

export interface GetChatHistoryResponse {
  data: ChatMessage[]
}
