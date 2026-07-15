import { http } from "./request"

export interface TranslationRequest {
  text: string
}

export interface TranslationResponse {
  translation: string
}

export const translateText = (data: TranslationRequest) => {
  return http.post<TranslationResponse>("/chat/translate", data)
}
