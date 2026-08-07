import { http } from './request'
import type { ChatRequest, ChatResponse, GetChatHistoryParams, GetChatHistoryResponse, TranscriptionResponse } from '@/types/chat'

export const sendChatMessage = async (data: ChatRequest): Promise<ChatResponse> => {
  return http.post('/chat', data)
}

export const getChatHistory = async (params?: GetChatHistoryParams): Promise<GetChatHistoryResponse> => {
  return http.get('/chat/history', { params })
}

export const transcribeAudio = async (audioFile: File): Promise<TranscriptionResponse> => {
  const formData = new FormData()
  formData.append('audio', audioFile)
  return http.upload('/chat/transcribe', formData)
}
