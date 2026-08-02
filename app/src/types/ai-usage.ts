export type AIUsageOperation =
  | "chat"
  | "tts"
  | "stt"
  | "embedding"
  | "translation"
export type AIUsageUnit = "tokens" | "seconds" | "characters"

export interface AIUsageEvent {
  operation: AIUsageOperation
  model: string
  usage: number
  unit: AIUsageUnit
  status: "succeeded" | "failed"
  created_at: string
}

export interface AIUsagePage {
  items: AIUsageEvent[]
  total: number
  page: number
  page_size: number
}

export interface AIUsageSummary {
  operation: AIUsageOperation
  unit: AIUsageUnit
  usage: number
  request_count: number
}
